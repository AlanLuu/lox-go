package ast

import (
	"encoding/json"
	"fmt"

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

	i.globals.Define(className, jsonClass)
}
