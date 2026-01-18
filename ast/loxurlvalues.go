package ast

import (
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strings"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxURLValues struct {
	values  url.Values
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxURLValues(values url.Values) *LoxURLValues {
	return &LoxURLValues{
		values:  values,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func EmptyLoxURLValues() *LoxURLValues {
	return NewLoxURLValues(url.Values{})
}

func (l *LoxURLValues) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxURLValues:
		if l == obj {
			return true
		}
		return reflect.DeepEqual(l.values, obj.values)
	default:
		return false
	}
}

func (l *LoxURLValues) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if field, ok := l.methods[methodName]; ok {
		return field, nil
	}
	urlFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native URL values object fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'URL values object.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "add":
		return urlFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'URL values object.add' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'URL values object.add' must be a string.")
			}
			l.values.Add(
				args[0].(*LoxString).str,
				args[1].(*LoxString).str,
			)
			return l, nil
		})
	case "addArgs":
		return urlFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen < 2 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"URL values object.addArgs: "+
							"expected at least 2 arguments but got %v.",
						argsLen,
					),
				)
			}
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'URL values object.addArgs' must be a string.")
			}
			var strArgs []string
			for index, arg := range args[1:] {
				switch arg := arg.(type) {
				case *LoxString:
					if index == 0 {
						strArgs = make([]string, 0, argsLen)
					}
					strArgs = append(strArgs, arg.str)
				default:
					strArgs = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument #%v in 'URL values object.addArgs' must be a string.",
							index+2,
						),
					)
				}
			}
			key := args[0].(*LoxString).str
			if arr, ok := l.values[key]; ok {
				if len(strArgs) > 0 {
					l.values[key] = append(arr, strArgs...)
				}
			} else {
				l.values[key] = strArgs
			}
			return l, nil
		})
	case "addList":
		return urlFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'URL values object.addList' must be a string.")
			}
			if _, ok := args[1].(*LoxList); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'URL values object.addList' must be a list.")
			}
			elements := args[1].(*LoxList).elements
			var strArgs []string
			for index, element := range elements {
				switch element := element.(type) {
				case *LoxString:
					if index == 0 {
						strArgs = make([]string, 0, len(elements))
					}
					strArgs = append(strArgs, element.str)
				default:
					strArgs = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"URL values object.addList: "+
								"list element at index %v must be a string.",
							index,
						),
					)
				}
			}
			key := args[0].(*LoxString).str
			if arr, ok := l.values[key]; ok {
				if len(strArgs) > 0 {
					l.values[key] = append(arr, strArgs...)
				}
			} else {
				l.values[key] = strArgs
			}
			return l, nil
		})
	case "clear":
		return urlFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.values = url.Values{}
			return l, nil
		})
	case "del":
		return urlFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.values.Del(loxStr.str)
				return l, nil
			}
			return argMustBeType("string")
		})
	case "delFunc":
		return urlFunc(2, func(i *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'URL values object.delFunc' must be a string.")
			}
			if _, ok := args[1].(LoxCallable); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'URL values object.delFunc' must be a function.")
			}
			key := args[0].(*LoxString).str
			values, ok := l.values[key]
			valuesLen := len(values)
			if ok && values != nil && valuesLen > 0 {
				callback := args[1].(LoxCallable)
				argList := getArgList(callback, 2)
				defer argList.Clear()
				if valuesLen == 1 {
					argList[0] = NewLoxStringQuote(values[0])
					argList[1] = int64(0)
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return nil, resultErr
					}
					if i.isTruthy(result) {
						delete(l.values, key)
					}
					return l, nil
				}
				deleted := map[string]struct{}{}
				var index int64 = 0
				for _, value := range slices.Clone(values) {
					if _, ok := deleted[value]; ok {
						continue
					}
					argList[0] = NewLoxStringQuote(value)
					argList[1] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return nil, resultErr
					}
					if i.isTruthy(result) {
						newValues := slices.DeleteFunc(l.values[key], func(s string) bool {
							return s == value
						})
						if len(newValues) == 0 {
							delete(l.values, key)
							break
						}
						l.values[key] = newValues
						deleted[value] = struct{}{}
					}
					index++
				}
			}
			return l, nil
		})
	case "delKeepSet":
		return urlFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				tempMap := map[string]any{}
				var index int64 = 0
				for key, values := range l.values {
					argList[0] = NewLoxStringQuote(key)
					valuesList := list.NewListCap[any](int64(len(values)))
					for _, value := range values {
						valuesList.Add(NewLoxStringQuote(value))
					}
					argList[1] = NewLoxList(valuesList)
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						tempMap = nil
						return nil, resultErr
					}
					const startErrStr = "URL values object.delKeepSet: on iteration #%v: "
					switch result := result.(type) {
					case int64:
						switch result {
						case 0:
							tempMap[key] = true
						case 1:
						default:
							tempMap = nil
							return nil, loxerror.RuntimeError(
								name,
								fmt.Sprintf(
									startErrStr+"integer return value must be equal "+
										"to 0 or 1, not %v.",
									index+1,
									result,
								),
							)
						}
					case *LoxString:
						tempMap[key] = []string{result.str}
					case *LoxList:
						var strValues []string
						for index2, element := range result.elements {
							switch element := element.(type) {
							case *LoxString:
								if index2 == 0 {
									strValues = make([]string, 0, len(result.elements))
								}
								strValues = append(strValues, element.str)
							default:
								strValues = nil
								tempMap = nil
								return nil, loxerror.RuntimeError(
									name,
									fmt.Sprintf(
										startErrStr+"list element at index %v "+
											"must be a string, not type '%v'.",
										index+1,
										index2,
										getType(element),
									),
								)
							}
						}
						tempMap[key] = strValues
					default:
						tempMap = nil
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								startErrStr+"return value of callback must be a "+
									"string, list, or an integer equal to 0 or 1, "+
									"not type '%v'.",
								index+1,
								getType(result),
							),
						)
					}
					index++
				}
				for key, value := range tempMap {
					switch value := value.(type) {
					case bool:
						if value {
							delete(l.values, key)
						}
					case []string:
						l.values[key] = value
					}
				}
				tempMap = nil
				return l, nil
			}
			return argMustBeType("function")
		})
	case "delKeepSetMap":
		return urlFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				newURLValues := EmptyLoxURLValues()
				var index int64 = 0
				for key, values := range l.values {
					argList[0] = NewLoxStringQuote(key)
					valuesList := list.NewListCap[any](int64(len(values)))
					for _, value := range values {
						valuesList.Add(NewLoxStringQuote(value))
					}
					argList[1] = NewLoxList(valuesList)
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						newURLValues = nil
						return nil, resultErr
					}
					const startErrStr = "URL values object.delKeepSetMap: on iteration #%v: "
					switch result := result.(type) {
					case int64:
						switch result {
						case 0:
						case 1:
							newURLValues.values[key] = values
						default:
							newURLValues = nil
							return nil, loxerror.RuntimeError(
								name,
								fmt.Sprintf(
									startErrStr+"integer return value must be equal "+
										"to 0 or 1, not %v.",
									index+1,
									result,
								),
							)
						}
					case *LoxString:
						newURLValues.values[key] = []string{result.str}
					case *LoxList:
						var strValues []string
						for index2, element := range result.elements {
							switch element := element.(type) {
							case *LoxString:
								if index2 == 0 {
									strValues = make([]string, 0, len(result.elements))
								}
								strValues = append(strValues, element.str)
							default:
								strValues = nil
								newURLValues = nil
								return nil, loxerror.RuntimeError(
									name,
									fmt.Sprintf(
										startErrStr+"list element at index %v "+
											"must be a string, not type '%v'.",
										index+1,
										index2,
										getType(element),
									),
								)
							}
						}
						newURLValues.values[key] = strValues
					default:
						newURLValues = nil
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								startErrStr+"return value of callback must be a "+
									"string, list, or an integer equal to 0 or 1, "+
									"not type '%v'.",
								index+1,
								getType(result),
							),
						)
					}
					index++
				}
				return newURLValues, nil
			}
			return argMustBeType("function")
		})
	case "encoded", "str":
		return urlFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.values.Encode()), nil
		})
	case "filter":
		return urlFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				newURLValues := EmptyLoxURLValues()
				var index int64 = 0
				for key, values := range l.values {
					argList[0] = NewLoxStringQuote(key)
					valuesList := list.NewListCap[any](int64(len(values)))
					for _, value := range values {
						valuesList.Add(NewLoxStringQuote(value))
					}
					argList[1] = NewLoxList(valuesList)
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						newURLValues = nil
						return nil, resultErr
					}
					if i.isTruthy(result) {
						newURLValues.values[key] = slices.Clone(values)
					}
					index++
				}
				return newURLValues, nil
			}
			return argMustBeType("function")
		})
	case "forEach":
		return urlFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				var index int64 = 0
				for key, values := range l.values {
					if len(values) == 0 {
						continue
					}
					argList[0] = NewLoxStringQuote(key)
					argList[1] = NewLoxStringQuote(values[0])
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultErr != nil && result == nil {
						return nil, resultErr
					}
					index++
				}
				return nil, nil
			}
			return argMustBeType("function")
		})
	case "forEachList":
		return urlFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				var index int64 = 0
				for key, values := range l.values {
					argList[0] = NewLoxStringQuote(key)
					valuesList := list.NewListCap[any](int64(len(values)))
					for _, value := range values {
						valuesList.Add(NewLoxStringQuote(value))
					}
					argList[1] = NewLoxList(valuesList)
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultErr != nil && result == nil {
						return nil, resultErr
					}
					index++
				}
				return nil, nil
			}
			return argMustBeType("function")
		})
	case "getList":
		return urlFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				values, ok := l.values[loxStr.str]
				if !ok || values == nil {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"URL values object.getList: "+
								"unknown URL values object key '%v'.",
							loxStr.str,
						),
					)
				}
				if len(values) == 0 {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"URL values object.getList: "+
								"no values associated with URL values object key '%v'.",
							loxStr.str,
						),
					)
				}
				valuesList := list.NewListCap[any](int64(len(values)))
				for _, value := range values {
					valuesList.Add(NewLoxStringQuote(value))
				}
				return NewLoxList(valuesList), nil
			}
			return argMustBeType("string")
		})
	case "has", "contains":
		return urlFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return l.values.Has(loxStr.str), nil
			}
			return argMustBeType("string")
		})
	case "isEmpty":
		return urlFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.Length() == 0, nil
		})
	case "iterList":
		return urlFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pairs := list.NewListCap[*LoxList](int64(len(l.values)))
			for key, values := range l.values {
				pair := list.NewListCap[any](2)
				pair.Add(NewLoxStringQuote(key))
				valuesList := list.NewListCap[any](int64(len(values)))
				for _, value := range values {
					valuesList.Add(NewLoxStringQuote(value))
				}
				pair.Add(NewLoxList(valuesList))
				pairs.Add(NewLoxList(pair))
			}
			return NewLoxIterator(&LoxDictIterator{pairs, 0}), nil
		})
	case "map":
		return urlFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				newURLValues := EmptyLoxURLValues()
				var index int64 = 0
				for key, values := range l.values {
					argList[0] = NewLoxStringQuote(key)
					valuesList := list.NewListCap[any](int64(len(values)))
					for _, value := range values {
						valuesList.Add(NewLoxStringQuote(value))
					}
					argList[1] = NewLoxList(valuesList)
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						newURLValues = nil
						return nil, resultErr
					}
					const startErrStr = "URL values object.map: on iteration #%v: "
					switch result := result.(type) {
					case *LoxString:
						newURLValues.values[key] = []string{result.str}
					case *LoxList:
						var strValues []string
						for index2, element := range result.elements {
							switch element := element.(type) {
							case *LoxString:
								if index2 == 0 {
									strValues = make([]string, 0, len(result.elements))
								}
								strValues = append(strValues, element.str)
							default:
								strValues = nil
								newURLValues = nil
								return nil, loxerror.RuntimeError(
									name,
									fmt.Sprintf(
										startErrStr+"list element at index %v "+
											"must be a string, not type '%v'.",
										index+1,
										index2,
										getType(element),
									),
								)
							}
						}
						newURLValues.values[key] = strValues
					default:
						newURLValues = nil
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								startErrStr+"return value of callback must be a "+
									"string or list, not type '%v'.",
								index+1,
								getType(result),
							),
						)
					}
					index++
				}
				return newURLValues, nil
			}
			return argMustBeType("function")
		})
	case "numValues":
		return urlFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var count int64 = 0
			for _, values := range l.values {
				for range values {
					if count < 0 {
						//Integer overflow, return max 64-bit signed value
						return int64((1 << 63) - 1), nil
					}
					count++
				}
			}
			return count, nil
		})
	case "set":
		return urlFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'URL values object.set' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'URL values object.set' must be a string.")
			}
			l.values.Set(
				args[0].(*LoxString).str,
				args[1].(*LoxString).str,
			)
			return l, nil
		})
	case "setArgs":
		return urlFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen < 2 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"URL values object.setArgs: "+
							"expected at least 2 arguments but got %v.",
						argsLen,
					),
				)
			}
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'URL values object.setArgs' must be a string.")
			}
			var strArgs []string
			for index, arg := range args[1:] {
				switch arg := arg.(type) {
				case *LoxString:
					if index == 0 {
						strArgs = make([]string, 0, argsLen)
					}
					strArgs = append(strArgs, arg.str)
				default:
					strArgs = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument #%v in 'URL values object.setArgs' must be a string.",
							index+2,
						),
					)
				}
			}
			key := args[0].(*LoxString).str
			l.values[key] = strArgs
			return l, nil
		})
	case "setFunc":
		return urlFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				tempMap := map[string][]string{}
				var index int64 = 0
				for key, values := range l.values {
					argList[0] = NewLoxStringQuote(key)
					valuesList := list.NewListCap[any](int64(len(values)))
					for _, value := range values {
						valuesList.Add(NewLoxStringQuote(value))
					}
					argList[1] = NewLoxList(valuesList)
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						tempMap = nil
						return nil, resultErr
					}
					const startErrStr = "URL values object.setFunc: on iteration #%v: "
					switch result := result.(type) {
					case *LoxString:
						tempMap[key] = []string{result.str}
					case *LoxList:
						var strValues []string
						for index2, element := range result.elements {
							switch element := element.(type) {
							case *LoxString:
								if index2 == 0 {
									strValues = make([]string, 0, len(result.elements))
								}
								strValues = append(strValues, element.str)
							default:
								strValues = nil
								tempMap = nil
								return nil, loxerror.RuntimeError(
									name,
									fmt.Sprintf(
										startErrStr+"list element at index %v "+
											"must be a string, not type '%v'.",
										index+1,
										index2,
										getType(element),
									),
								)
							}
						}
						tempMap[key] = strValues
					default:
						tempMap = nil
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								startErrStr+"return value of callback must be a "+
									"string or list, not type '%v'.",
								index+1,
								getType(result),
							),
						)
					}
					index++
				}
				for key, value := range tempMap {
					l.values[key] = value
				}
				tempMap = nil
				return l, nil
			}
			return argMustBeType("function")
		})
	case "setKeyToEmptyList":
		return urlFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.values[loxStr.str] = []string{}
				return l, nil
			}
			return argMustBeType("string")
		})
	case "setList":
		return urlFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'URL values object.setList' must be a string.")
			}
			if _, ok := args[1].(*LoxList); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'URL values object.setList' must be a list.")
			}
			key := args[0].(*LoxString).str
			elements := args[1].(*LoxList).elements
			if len(elements) == 0 {
				l.values[key] = []string{}
				return l, nil
			}
			var strArgs []string
			for index, element := range elements {
				switch element := element.(type) {
				case *LoxString:
					if index == 0 {
						strArgs = make([]string, 0, len(elements))
					}
					strArgs = append(strArgs, element.str)
				default:
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"URL values object.setList: "+
								"list element at index %v must be a string.",
							index,
						),
					)
				}
			}
			l.values[key] = strArgs
			return l, nil
		})
	case "toDict":
		return urlFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			dict := EmptyLoxDict()
			for key, values := range l.values {
				if len(values) == 0 {
					continue
				}
				dict.setKeyValue(NewLoxStringQuote(key), NewLoxStringQuote(values[0]))
			}
			return dict, nil
		})
	case "toDictList":
		return urlFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			dict := EmptyLoxDict()
			for key, values := range l.values {
				inner := list.NewListCap[any](int64(len(values)))
				for _, value := range values {
					inner.Add(NewLoxStringQuote(value))
				}
				dict.setKeyValue(NewLoxStringQuote(key), NewLoxList(inner))
			}
			return dict, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "URL values objects have no property called '"+methodName+"'.")
}

func (l *LoxURLValues) Index(element any) (any, error) {
	switch element := element.(type) {
	case *LoxString:
		values, ok := l.values[element.str]
		if !ok || values == nil {
			return nil, loxerror.Error(
				fmt.Sprintf("Unknown URL values object key '%v'.", element.str),
			)
		}
		if len(values) == 0 {
			return nil, loxerror.Error(
				fmt.Sprintf(
					"No values associated with URL values object key '%v'.",
					element.str,
				),
			)
		}
		return NewLoxStringQuote(values[0]), nil
	}
	return nil, loxerror.Error(
		IndexMustBe("URL values objects", element, "a string"),
	)
}

func (l *LoxURLValues) Iterator() interfaces.Iterator {
	pairs := list.NewListCap[*LoxList](int64(len(l.values)))
	for key, values := range l.values {
		pair := list.NewListCap[any](2)
		pair.Add(NewLoxStringQuote(key))
		if len(values) > 0 {
			pair.Add(NewLoxStringQuote(values[0]))
		} else {
			pair.Add(EmptyLoxString())
		}
		pairs.Add(NewLoxList(pair))
	}
	return &LoxDictIterator{pairs, 0}
}

func (l *LoxURLValues) Length() int64 {
	return int64(len(l.values))
}

func (l *LoxURLValues) String() string {
	if len(l.values) == 0 {
		return fmt.Sprintf("<empty URL values at %p>", l)
	}
	firstIter := true
	var builder strings.Builder
	builder.WriteString("URL values {")
	for key, values := range l.values {
		if firstIter {
			firstIter = false
		} else {
			builder.WriteByte(' ')
		}
		builder.WriteString(getResult(key, key, false))
		builder.WriteByte('=')
		valuesLen := len(values)
		if valuesLen == 1 {
			value := values[0]
			builder.WriteString(getResult(value, value, false))
		} else {
			builder.WriteByte('[')
			for i, value := range values {
				builder.WriteString(getResult(value, value, false))
				if i < valuesLen-1 {
					builder.WriteByte(',')
					builder.WriteByte(' ')
				}
			}
			builder.WriteByte(']')
		}
	}
	builder.WriteByte('}')
	return fmt.Sprintf("<%v at %p>", builder.String(), l)
}

func (l *LoxURLValues) Type() string {
	return "url values"
}
