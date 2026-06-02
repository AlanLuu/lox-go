package ast

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

func (i *Interpreter) defineJSONFuncs() {
	className := "JSON"
	jsonClass := NewLoxClass(className, nil, false)
	jsonFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native JSON fn %v at %p>", name, &s)
		}
		jsonClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'JSON.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	jsonFunc("marshal", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		selfReferentialErr := func(source any) (any, error) {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf(
					"JSON.marshal: cannot marshal self-referential %v.",
					getType(source),
				),
			)
		}
		var jsonObj any
		var processArg func(any, any, map[any]struct{}) (any, error)
		processArg = func(
			arg any,
			originalArg any,
			visited map[any]struct{},
		) (any, error) {
			switch arg := arg.(type) {
			case nil, bool, int64, float64:
				return arg, nil
			case *LoxString:
				return arg.str, nil
			case LoxStringStr:
				return arg.str, nil
			case *LoxDict:
				visited[arg] = struct{}{}
				jsonMap := map[string]any{}
				for key, value := range arg.entries {
					_, keyVisited := visited[key]
					if key == originalArg || keyVisited {
						jsonMap = nil
						return selfReferentialErr(key)
					}
					_, valueVisited := visited[value]
					if value == originalArg || valueVisited {
						jsonMap = nil
						return selfReferentialErr(value)
					}
					var resultKey string
					resultKeyAny, err := processArg(key, originalArg, visited)
					if err != nil {
						jsonMap = nil
						return nil, err
					}
					switch result := resultKeyAny.(type) {
					case nil:
						resultKey = "null"
					case string:
						resultKey = result
					case fmt.Stringer:
						resultKey = result.String()
					default:
						resultKey = fmt.Sprint(result)
					}
					var resultValue any
					resultValue, err = processArg(value, originalArg, visited)
					if err != nil {
						jsonMap = nil
						return nil, err
					}
					jsonMap[resultKey] = resultValue
				}
				return jsonMap, nil
			case *LoxList:
				visited[arg] = struct{}{}
				jsonList := make([]any, 0, len(arg.elements))
				for _, element := range arg.elements {
					_, elementVisited := visited[element]
					if element == originalArg || elementVisited {
						jsonList = nil
						return selfReferentialErr(element)
					}
					result, err := processArg(element, originalArg, visited)
					if err != nil {
						jsonList = nil
						return nil, err
					}
					jsonList = append(jsonList, result)
				}
				return jsonList, nil
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("JSON.marshal: type '%v' cannot be serialized as JSON.",
						getType(arg)))
			}
		}
		arg := args[0]
		var processErr error
		jsonObj, processErr = processArg(arg, arg, map[any]struct{}{})
		if processErr != nil {
			return nil, processErr
		}
		jsonBytes, jsonErr := json.Marshal(jsonObj)
		if jsonErr != nil {
			return nil, loxerror.RuntimeError(
				in.callToken,
				"JSON.marshal: "+jsonErr.Error(),
			)
		}
		return NewLoxStringQuote(string(jsonBytes)), nil
	})
	jsonFunc("parse", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if jsonLoxStr, ok := args[0].(*LoxString); ok {
			jsonStr := strings.TrimSpace(jsonLoxStr.str)
			if len(jsonStr) == 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"unexpected end of JSON input")
			}
			if jsonStr == "null" {
				return nil, nil
			}

			jsonStrByteArr := []byte(jsonStr)
			var jsonArr []any
			var jsonBool bool
			var jsonMap map[string]any
			var jsonNum float64
			var finalJsonString string
			var jsonErr error

			setJsonBool := false
			setJsonNum := false
			switch jsonStr[0] {
			case '{':
				jsonMap = make(map[string]any)
				jsonErr = json.Unmarshal(jsonStrByteArr, &jsonMap)
			case '[':
				jsonArr = make([]any, 0)
				jsonErr = json.Unmarshal(jsonStrByteArr, &jsonArr)
			default:
				_, numErr := strconv.ParseFloat(jsonStr, 64)
				if numErr == nil {
					setJsonNum = true
					jsonErr = json.Unmarshal(jsonStrByteArr, &jsonNum)
					break
				}
				if jsonStr == "true" || jsonStr == "false" {
					setJsonBool = true
					jsonErr = json.Unmarshal(jsonStrByteArr, &jsonBool)
					break
				}
				jsonErr = json.Unmarshal(jsonStrByteArr, &finalJsonString)
			}
			if jsonErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, jsonErr.Error())
			}

			escapeChars := map[rune]string{
				'\a': "\\a",
				'\n': "\\n",
				'\r': "\\r",
				'\t': "\\t",
				'\b': "\\b",
				'\f': "\\f",
				'\v': "\\v",
			}
			processString := func(str string) *LoxString {
				useDoubleQuote := false
				var finalStrBuilder strings.Builder
				for _, c := range str {
					if escapeChar, ok := escapeChars[c]; ok {
						finalStrBuilder.WriteString(escapeChar)
					} else {
						switch c {
						case '\'':
							useDoubleQuote = true
						case '"', '\\':
							finalStrBuilder.WriteRune('\\')
						}
						finalStrBuilder.WriteRune(c)
					}
				}
				finalStr := finalStrBuilder.String()
				if useDoubleQuote {
					return NewLoxString(finalStr, '"')
				}
				return NewLoxString(finalStr, '\'')
			}
			parseValue := func(value any) any {
				switch value := value.(type) {
				case float64:
					return util.IntOrFloat(value)
				case string:
					return processString(value)
				}
				return value
			}
			var parseList func(*LoxList, *[]any)
			var parseMap func(*LoxDict, *map[string]any)
			parseList = func(jsonLoxList *LoxList, jsonList *[]any) {
				for _, value := range *jsonList {
					switch value := value.(type) {
					case []any:
						innerLoxList := EmptyLoxList()
						parseList(innerLoxList, &value)
						jsonLoxList.elements.Add(innerLoxList)
					case map[string]any:
						innerLoxDict := EmptyLoxDict()
						parseMap(innerLoxDict, &value)
						jsonLoxList.elements.Add(innerLoxDict)
					default:
						jsonLoxList.elements.Add(parseValue(value))
					}
				}
			}
			parseMap = func(jsonLoxDict *LoxDict, jsonMap *map[string]any) {
				for key, value := range *jsonMap {
					switch value := value.(type) {
					case []any:
						innerLoxList := EmptyLoxList()
						parseList(innerLoxList, &value)
						jsonLoxDict.setKeyValue(processString(key), innerLoxList)
					case map[string]any:
						innerLoxDict := EmptyLoxDict()
						parseMap(innerLoxDict, &value)
						jsonLoxDict.setKeyValue(processString(key), innerLoxDict)
					default:
						jsonLoxDict.setKeyValue(processString(key), parseValue(value))
					}
				}
			}

			switch {
			case jsonArr != nil:
				finalLoxList := EmptyLoxList()
				parseList(finalLoxList, &jsonArr)
				return finalLoxList, nil
			case jsonMap != nil:
				finalLoxDict := EmptyLoxDict()
				parseMap(finalLoxDict, &jsonMap)
				return finalLoxDict, nil
			case setJsonBool:
				return jsonBool, nil
			case setJsonNum:
				return util.IntOrFloat(jsonNum), nil
			default:
				return NewLoxStringQuote(finalJsonString), nil
			}
		}
		return argMustBeType(in.callToken, "parse", "string")
	})
	jsonFunc("stringify", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		escapeChars := map[rune]string{
			'\a': "\\\\a",
			'\n': "\\\\n",
			'\r': "\\\\r",
			'\t': "\\\\t",
			'\b': "\\\\b",
			'\f': "\\\\f",
			'\v': "\\\\v",
		}
		selfReferentialErr := func(source any) (string, error) {
			return "", loxerror.RuntimeError(in.callToken,
				fmt.Sprintf(
					"JSON.stringify: cannot stringify self-referential %v.",
					getType(source),
				),
			)
		}
		processString := func(str string, doubleQuotes bool) string {
			var finalStrBuilder strings.Builder
			for _, c := range str {
				if escapeChar, ok := escapeChars[c]; ok {
					finalStrBuilder.WriteString(escapeChar)
				} else {
					switch c {
					case '"', '\\':
						finalStrBuilder.WriteRune('\\')
					}
					finalStrBuilder.WriteRune(c)
				}
			}
			finalStr := finalStrBuilder.String()
			if doubleQuotes {
				return fmt.Sprintf("\"%v\"", finalStr)
			}
			return finalStr
		}
		var getJSONString func(any, any, map[any]struct{}, bool) (string, error)
		getJSONString = func(
			source any,
			originalSource any,
			visited map[any]struct{},
			doubleQuotes bool,
		) (string, error) {
			switch source := source.(type) {
			case nil:
				return processString("null", doubleQuotes), nil
			case bool:
				if source {
					return processString("true", doubleQuotes), nil
				}
				return processString("false", doubleQuotes), nil
			case int64:
				return processString(fmt.Sprint(source), doubleQuotes), nil
			case float64:
				switch {
				case math.IsInf(source, 1), math.IsInf(source, -1):
					return processString("null", doubleQuotes), nil
				case util.FloatIsInt(source):
					return processString(fmt.Sprintf("%.1f", source), doubleQuotes), nil
				default:
					return processString(util.FormatFloat(source), doubleQuotes), nil
				}
			case *LoxString:
				return processString(source.str, true), nil
			case LoxStringStr:
				return processString(source.str, true), nil
			case *LoxDict:
				visited[source] = struct{}{}
				sourceLen := len(source.entries)
				var dictStr strings.Builder
				dictStr.WriteByte('{')
				i := 0
				for key, value := range source.entries {
					_, keyVisited := visited[key]
					if key == originalSource || keyVisited {
						return selfReferentialErr(key)
					}
					result, err := getJSONString(key, originalSource, visited, true)
					if err != nil {
						return "", err
					}
					dictStr.WriteString(result)
					dictStr.WriteString(": ")
					_, valueVisited := visited[value]
					if value == originalSource || valueVisited {
						return selfReferentialErr(value)
					}
					result, err = getJSONString(value, originalSource, visited, false)
					if err != nil {
						return "", err
					}
					dictStr.WriteString(result)
					if i < sourceLen-1 {
						dictStr.WriteString(", ")
					}
					i++
				}
				dictStr.WriteByte('}')
				return dictStr.String(), nil
			case *LoxList:
				visited[source] = struct{}{}
				sourceLen := len(source.elements)
				var listStr strings.Builder
				listStr.WriteByte('[')
				for i, element := range source.elements {
					_, elementVisited := visited[element]
					if element == originalSource || elementVisited {
						return selfReferentialErr(element)
					}
					result, err := getJSONString(element, originalSource, visited, doubleQuotes)
					if err != nil {
						return "", err
					}
					listStr.WriteString(result)
					if i < sourceLen-1 {
						listStr.WriteString(", ")
					}
				}
				listStr.WriteByte(']')
				return listStr.String(), nil
			default:
				return "", loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Type '%v' cannot be serialized as JSON.",
						getType(source)))
			}
		}

		arg := args[0]
		jsonString, jsonStringErr := getJSONString(arg, arg, map[any]struct{}{}, false)
		if jsonStringErr != nil {
			return nil, jsonStringErr
		}
		return NewLoxString(jsonString, '\''), nil
	})
	jsonFunc("valid", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return json.Valid([]byte(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "valid", "string")
	})

	i.globals.Define(className, jsonClass)
}
