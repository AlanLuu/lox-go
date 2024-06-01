package ast

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
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
	argMustBeType := func(name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'JSON.%v' must be a %v.", name, theType)
		return nil, loxerror.Error(errStr)
	}

	jsonFunc("parse", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if jsonLoxStr, ok := args[0].(*LoxString); ok {
			jsonStr := jsonLoxStr.str
			jsonMap := make(map[string]any)
			jsonErr := json.Unmarshal([]byte(jsonStr), &jsonMap)
			if jsonErr != nil {
				return nil, jsonErr
			}

			parseValue := func(value any) any {
				switch value := value.(type) {
				case float64:
					return util.IntOrFloat(value)
				case string:
					return NewLoxStringQuote(value)
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
						jsonLoxDict.setKeyValue(NewLoxStringQuote(key), innerLoxList)
					case map[string]any:
						innerLoxDict := EmptyLoxDict()
						parseMap(innerLoxDict, &value)
						jsonLoxDict.setKeyValue(NewLoxStringQuote(key), innerLoxDict)
					default:
						jsonLoxDict.setKeyValue(NewLoxStringQuote(key), parseValue(value))
					}
				}
			}

			finalLoxDict := EmptyLoxDict()
			parseMap(finalLoxDict, &jsonMap)
			return finalLoxDict, nil
		}
		return argMustBeType("parse", "string")
	})
	jsonFunc("stringify", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		escapeChars := map[rune]string{
			'\a': "\\\\a",
			'\n': "\\\\n",
			'\r': "\\\\r",
			'\t': "\\\\t",
			'\b': "\\\\b",
			'\f': "\\\\f",
			'\v': "\\\\v",
		}
		selfReferentialErr := func(originalSource any) (string, error) {
			return "", loxerror.Error(
				fmt.Sprintf(
					"Cannot stringify self-referential %v.",
					getType(originalSource),
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
					case '"', '\'', '\\':
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
		var getJSONString func(any, any, bool) (string, error)
		getJSONString = func(
			source any,
			originalSource any,
			doubleQuotes bool,
		) (string, error) {
			switch source := source.(type) {
			case nil:
				return processString("null", doubleQuotes), nil
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
				sourceLen := len(source.entries)
				var dictStr strings.Builder
				dictStr.WriteByte('{')
				i := 0
				for key, value := range source.entries {
					if key == originalSource {
						return selfReferentialErr(originalSource)
					} else {
						result, err := getJSONString(key, originalSource, true)
						if err != nil {
							return "", err
						}
						dictStr.WriteString(result)
					}
					dictStr.WriteString(": ")
					if value == originalSource {
						return selfReferentialErr(originalSource)
					} else {
						result, err := getJSONString(value, originalSource, false)
						if err != nil {
							return "", err
						}
						dictStr.WriteString(result)
					}
					if i < sourceLen-1 {
						dictStr.WriteString(", ")
					}
					i++
				}
				dictStr.WriteByte('}')
				return dictStr.String(), nil
			case *LoxList:
				sourceLen := len(source.elements)
				var listStr strings.Builder
				listStr.WriteByte('[')
				for i, element := range source.elements {
					if element == originalSource {
						return selfReferentialErr(originalSource)
					} else {
						result, err := getJSONString(element, originalSource, doubleQuotes)
						if err != nil {
							return "", err
						}
						listStr.WriteString(result)
					}
					if i < sourceLen-1 {
						listStr.WriteString(", ")
					}
				}
				listStr.WriteByte(']')
				return listStr.String(), nil
			default:
				return "", loxerror.Error(
					fmt.Sprintf("Type '%v' cannot be serialized as JSON.",
						getType(source)))
			}
		}

		arg := args[0]
		jsonString, jsonStringErr := getJSONString(arg, arg, arg == nil)
		if jsonStringErr != nil {
			return nil, jsonStringErr
		}
		return NewLoxString(jsonString, '\''), nil
	})

	i.globals.Define(className, jsonClass)
}
