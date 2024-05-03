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
	methodName := name.Lexeme
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
			return nil, loxerror.RuntimeError(name, "Argument to 'list.extend' must be a list.")
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
	case "pop":
		return listFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 0:
				if len(l.elements) == 0 {
					return nil, loxerror.RuntimeError(name, "Cannot pop from empty list.")
				}
				return l.elements.Pop(), nil
			case 1:
				if index, ok := args[0].(int64); ok {
					if len(l.elements) == 0 {
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
	}
	return nil, loxerror.RuntimeError(name, "Lists have no property called '"+methodName+"'.")
}
