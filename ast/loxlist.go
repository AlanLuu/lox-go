package ast

import (
	"fmt"
	"reflect"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

const LIST_INDEX_MUST_BE_WHOLE_NUMBER = "Index value must be a whole number."
const LIST_INDEX_OUT_OF_RANGE = "List index out of range."

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
	case "insert":
		return listFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if index, ok := args[0].(int64); ok {
				if index < 0 || index > int64(len(l.elements)) {
					return nil, loxerror.RuntimeError(name, LIST_INDEX_OUT_OF_RANGE)
				}
				l.elements.AddAt(index, args[1])
				return nil, nil
			}
			return nil, loxerror.RuntimeError(name, LIST_INDEX_MUST_BE_WHOLE_NUMBER)
		})
	case "pop":
		return listFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 0:
				return l.elements.Pop(), nil
			case 1:
				if index, ok := args[0].(int64); ok {
					if index < 0 || index >= int64(len(l.elements)) {
						return nil, loxerror.RuntimeError(name, LIST_INDEX_OUT_OF_RANGE)
					}
					return l.elements.RemoveIndex(index), nil
				}
				return nil, loxerror.RuntimeError(name, LIST_INDEX_MUST_BE_WHOLE_NUMBER)
			}
			return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		})
	}
	return nil, loxerror.RuntimeError(name, "Lists have no property called '"+methodName+"'.")
}
