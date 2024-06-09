package ast

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

func IndexMustBeWholeNum(theType string, index any) string {
	indexVal := index
	format := "%v"
	switch index := index.(type) {
	case float64:
		if util.FloatIsInt(index) {
			format = "%.1f"
		} else {
			indexVal = util.FormatFloat(index)
		}
	}
	return fmt.Sprintf("%v index '"+format+"' must be an integer.", theType, indexVal)
}

func ListIndexMustBeWholeNum(index any) string {
	return IndexMustBeWholeNum("List", index)
}

func ListIndexOutOfRange(index int64) string {
	return fmt.Sprintf("List index %v out of range.", index)
}

type LoxList struct {
	elements list.List[any]
	methods  map[string]*struct{ ProtoLoxCallable }
}

func NewLoxList(elements list.List[any]) *LoxList {
	return &LoxList{
		elements: elements,
		methods:  make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func EmptyLoxList() *LoxList {
	return NewLoxList(list.NewList[any]())
}

func (l *LoxList) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxList:
		return reflect.DeepEqual(l.elements, obj.elements)
	default:
		return false
	}
}

func (l *LoxList) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	listFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native list fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	indexOf := func(obj any) int64 {
		if equatableObj, ok := obj.(interfaces.Equatable); ok {
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
	lastIndexOf := func(obj any) int64 {
		if equatableObj, ok := obj.(interfaces.Equatable); ok {
			for i := len(l.elements) - 1; i >= 0; i-- {
				if equatableObj.Equals(l.elements[i]) {
					return int64(i)
				}
			}
		} else {
			for i := len(l.elements) - 1; i >= 0; i-- {
				if obj == l.elements[i] {
					return int64(i)
				}
			}
		}
		return -1
	}
	removeElements := func(arg any) bool {
		removed := false
		for i := int64(len(l.elements)) - 1; i >= 0; i-- {
			remove := false
			switch element := l.elements[i].(type) {
			case interfaces.Equatable:
				remove = element.Equals(arg)
			default:
				remove = element == arg
			}
			if remove {
				removed = true
				l.elements.RemoveIndex(i)
			}
		}
		return removed
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
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'list.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "all":
		return listFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				for index, element := range l.elements {
					argList[0] = element
					argList[1] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return nil, resultErr
					}
					if !i.isTruthy(result) {
						return false, nil
					}
				}
				return true, nil
			}
			return argMustBeType("function")
		})
	case "any":
		return listFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				for index, element := range l.elements {
					argList[0] = element
					argList[1] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return nil, resultErr
					}
					if i.isTruthy(result) {
						return true, nil
					}
				}
				return false, nil
			}
			return argMustBeType("function")
		})
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
			if equatableObj, ok := obj.(interfaces.Equatable); ok {
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
				newList := list.NewList[any]()
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
				return NewLoxList(newList), nil
			}
			return argMustBeType("function")
		})
	case "find":
		return listFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				for index, element := range l.elements {
					argList[0] = element
					argList[1] = int64(index)
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return nil, resultErr
					}
					if i.isTruthy(result) {
						return element, nil
					}
				}
				return nil, nil
			}
			return argMustBeType("function")
		})
	case "findIndex":
		return listFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				for index, element := range l.elements {
					argList[0] = element
					argList[1] = int64(index)
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return nil, resultErr
					}
					if i.isTruthy(result) {
						return argList[1], nil
					}
				}
				return int64(-1), nil
			}
			return argMustBeType("function")
		})
	case "flatten":
		return listFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newList := list.NewList[any]()
			var flatten func(elements list.List[any]) error
			flatten = func(elements list.List[any]) error {
				for _, element := range elements {
					switch element := element.(type) {
					case *LoxList:
						if element == l {
							return loxerror.RuntimeError(name, "Cannot flatten self-referential list.")
						}
						flattenErr := flatten(element.elements)
						if flattenErr != nil {
							return flattenErr
						}
					default:
						newList.Add(element)
					}
				}
				return nil
			}
			flattenErr := flatten(l.elements)
			if flattenErr != nil {
				return nil, flattenErr
			}
			return NewLoxList(newList), nil
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
				originalIndex := index
				if index < 0 {
					index += int64(len(l.elements))
				}
				if index < 0 || index > int64(len(l.elements)) {
					return nil, loxerror.RuntimeError(name, ListIndexOutOfRange(originalIndex))
				}
				l.elements.AddAt(index, args[1])
				return nil, nil
			}
			return nil, loxerror.RuntimeError(name, ListIndexMustBeWholeNum(args[0]))
		})
	case "isEmpty":
		return listFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return len(l.elements) == 0, nil
		})
	case "join":
		return listFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				var quote byte = '\''
				var builder strings.Builder
				for index, element := range l.elements {
					switch element := element.(type) {
					case *LoxString:
						if quote != '"' && element.quote == '"' {
							quote = '"'
						}
					}
					elementAsStr := getResult(element, element, true)
					builder.WriteString(elementAsStr)
					if loxStr.str != "" && index < len(l.elements)-1 {
						builder.WriteString(loxStr.str)
					}
				}
				return NewLoxString(builder.String(), quote), nil
			}
			return argMustBeType("string")
		})
	case "lastIndex":
		return listFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			return lastIndexOf(args[0]), nil
		})
	case "map":
		return listFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				newList := list.NewList[any]()
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
				return NewLoxList(newList), nil
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
					originalIndex := index
					if index < 0 {
						index += int64(len(l.elements))
					}
					if index < 0 || index >= int64(len(l.elements)) {
						return nil, loxerror.RuntimeError(name, ListIndexOutOfRange(originalIndex))
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
				defer argList.Clear()
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
	case "removeAll":
		return listFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			removed := false
			for _, arg := range args {
				if removeElements(arg) {
					removed = true
				}
			}
			return removed, nil
		})
	case "removeAllList":
		return listFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxList, ok := args[0].(*LoxList); ok {
				if loxList == l {
					removed := len(l.elements) > 0
					l.elements.Clear()
					return removed, nil
				}
				removed := false
				for _, element := range loxList.elements {
					if removeElements(element) {
						removed = true
					}
				}
				return removed, nil
			}
			return argMustBeType("list")
		})
	case "shuffle":
		return listFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			rand.Shuffle(len(l.elements), func(a int, b int) {
				l.elements[a], l.elements[b] = l.elements[b], l.elements[a]
			})
			return nil, nil
		})
	case "toSet":
		return listFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newSet := EmptyLoxSet()
			for index, element := range l.elements {
				_, errStr := newSet.add(element)
				if len(errStr) > 0 {
					errStr = "Type '%v' at index %v cannot be used as set element."
					return nil, loxerror.RuntimeError(name,
						fmt.Sprintf(errStr, getType(element), index))
				}
			}
			return newSet, nil
		})
	case "with":
		return listFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if newIndex, ok := args[0].(int64); ok {
				originalNewIndex := newIndex
				if newIndex < 0 {
					newIndex += int64(len(l.elements))
				}
				if newIndex < 0 || newIndex >= int64(len(l.elements)) {
					return nil, loxerror.RuntimeError(name, ListIndexOutOfRange(originalNewIndex))
				}
				newElement := args[1]
				newList := list.NewList[any]()
				for oldIndex, oldElement := range l.elements {
					if int64(oldIndex) != newIndex {
						newList.Add(oldElement)
					} else {
						newList.Add(newElement)
					}
				}
				return NewLoxList(newList), nil
			}
			return nil, loxerror.RuntimeError(name, "First argument to 'list.with' must be an integer.")
		})
	}
	return nil, loxerror.RuntimeError(name, "Lists have no property called '"+methodName+"'.")
}

func (l *LoxList) Length() int64 {
	return int64(len(l.elements))
}

func (l *LoxList) String() string {
	return getResult(l, l, true)
}

func (l *LoxList) Type() string {
	return "list"
}
