package ast

import (
	"fmt"
	"reflect"

	"github.com/AlanLuu/lox/equatable"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func ListIndexMustBeWholeNum(index any) string {
	return fmt.Sprintf("List index '%v' must be a whole number.", index)
}

func ListIndexOutOfRange(index int64) string {
	return fmt.Sprintf("List index %v out of range.", index)
}

type LoxList struct {
	elements list.List[Expr]
}

func (l *LoxList) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxList:
		return reflect.DeepEqual(l.elements, obj.elements)
	default:
		return false
	}
}

func (l *LoxList) Get(name token.Token) (any, error) {
	listFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (struct{ ProtoLoxCallable }, error) {
		s := struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string { return "<native list fn>" }
		return s, nil
	}
	indexOf := func(obj any) int64 {
		if equatableObj, ok := obj.(equatable.Equatable); ok {
			for i, element := range l.elements {
				if equatableObj.Equals(element) {
					return int64(i)
				}
			}
		} else {
			for i, element := range l.elements {
				if obj == element {
					return int64(i)
				}
			}
		}
		return -1
	}
	getArgList := func(callback *LoxFunction, numArgs int) list.List[any] {
		argList := list.NewListLen[any](int64(numArgs))
		callbackArity := callback.arity()
		if callbackArity > numArgs {
			for i := 0; i < callbackArity-numArgs; i++ {
				argList.Add(nil)
			}
		}
		return argList
	}
	methodName := name.Lexeme
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'list.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "append":
		return listFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			l.elements.Add(args[0])
			return nil, nil
		})
	case "clear":
		return listFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.elements.Clear()
			return nil, nil
		})
	case "contains":
		return listFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			return indexOf(args[0]) >= 0, nil
		})
	case "count":
		return listFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			obj := args[0]
			count := int64(0)
			if equatableObj, ok := obj.(equatable.Equatable); ok {
				for _, element := range l.elements {
					if equatableObj.Equals(element) {
						count++
					}
				}
			} else {
				for _, element := range l.elements {
					if obj == element {
						count++
					}
				}
			}
			return count, nil
		})
	case "extend":
		return listFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if extendList, ok := args[0].(*LoxList); ok {
				for _, element := range extendList.elements {
					l.elements.Add(element)
				}
				return nil, nil
			}
			return argMustBeType("list")
		})
	case "filter":
		return listFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				newList := list.NewList[Expr]()
				for index, element := range l.elements {
					argList[0] = element
					argList[1] = int64(index)
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						newList.Clear()
						return nil, resultErr
					}
					if i.isTruthy(result) {
						newList.Add(element)
					}
				}
				return &LoxList{newList}, nil
			}
			return argMustBeType("function")
		})
	case "forEach":
		return listFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				for index, element := range l.elements {
					argList[0] = element
					argList[1] = int64(index)
					result, resultErr := callback.call(i, argList)
					if resultErr != nil && result == nil {
						return nil, resultErr
					}
				}
				return nil, nil
			}
			return argMustBeType("function")
		})
	case "index":
		return listFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			return indexOf(args[0]), nil
		})
	case "insert":
		return listFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if index, ok := args[0].(int64); ok {
				if index < 0 || index > int64(len(l.elements)) {
					return nil, loxerror.RuntimeError(name, ListIndexOutOfRange(index))
				}
				l.elements.AddAt(index, args[1])
				return nil, nil
			}
			return nil, loxerror.RuntimeError(name, ListIndexMustBeWholeNum(args[0]))
		})
	case "map":
		return listFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				newList := list.NewList[Expr]()
				for index, element := range l.elements {
					argList[0] = element
					argList[1] = int64(index)
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						newList.Add(resultReturn.FinalValue)
					} else if resultErr != nil {
						newList.Clear()
						return nil, resultErr
					} else {
						newList.Add(result)
					}
				}
				return &LoxList{newList}, nil
			}
			return argMustBeType("function")
		})
	case "pop":
		return listFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 0:
				if l.elements.IsEmpty() {
					return nil, loxerror.RuntimeError(name, "Cannot pop from empty list.")
				}
				return l.elements.Pop(), nil
			case 1:
				if index, ok := args[0].(int64); ok {
					if l.elements.IsEmpty() {
						return nil, loxerror.RuntimeError(name, "Cannot pop from empty list.")
					}
					if index < 0 || index >= int64(len(l.elements)) {
						return nil, loxerror.RuntimeError(name, ListIndexOutOfRange(index))
					}
					return l.elements.RemoveIndex(index), nil
				}
				return nil, loxerror.RuntimeError(name, ListIndexMustBeWholeNum(args[0]))
			}
			return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		})
	case "reduce":
		return listFunc(-1, func(i *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 || argsLen > 2 {
				return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
			if callback, ok := args[0].(*LoxFunction); ok {
				var value any
				switch argsLen {
				case 1:
					if len(l.elements) == 0 {
						return nil, loxerror.RuntimeError(name, "Cannot call 'list.reduce' on empty list without initial value.")
					}
					value = l.elements[0]
				case 2:
					value = args[1]
				}

				argList := getArgList(callback, 4)
				argList[3] = l
				for index, element := range l.elements {
					if index == 0 && argsLen == 1 {
						continue
					}
					argList[0] = value
					argList[1] = element
					argList[2] = index

					var valueErr error
					value, valueErr = callback.call(i, argList)
					if valueReturn, ok := value.(Return); ok {
						value = valueReturn.FinalValue
					} else if valueErr != nil {
						return nil, valueErr
					}
				}
				return value, nil
			}
			return nil, loxerror.RuntimeError(name, "First argument to 'list.reduce' must be a function.")
		})
	case "remove":
		return listFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			index := indexOf(args[0])
			if index >= 0 {
				l.elements.RemoveIndex(index)
				return true, nil
			}
			return false, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Lists have no property called '"+methodName+"'.")
}

func (l *LoxList) String() string {
	return getResult(l, true)
}
