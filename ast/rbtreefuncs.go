package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineRBTreeFuncs() {
	className := "rbtree"
	rbTreeClass := NewLoxClass(className, nil, false)
	rbTreeFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native rbtree class fn %v at %p>", name, &s)
		}
		rbTreeClass.classProperties[name] = s
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'rbtree class.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	rbTreeFunc("args", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"rbtree class.args: expected at least 1 argument but got 0.")
		}
		return NewLoxRBTreeArgs(args...), nil
	})
	rbTreeFunc("dictInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxDict, ok := args[0].(*LoxDict); ok {
			m := map[int64]any{}
			it := loxDict.Iterator()
			for it.HasNext() {
				pair := it.Next().(*LoxList).elements
				var key int64
				switch pairKey := pair[0].(type) {
				case int64:
					key = pairKey
				default:
					return nil, loxerror.RuntimeError(in.callToken,
						"Dictionary argument to 'rbtree class.dictInt' must only have integer keys.")
				}
				value := pair[1]
				m[key] = value
			}
			return NewLoxRBTreeMapInt(m), nil
		}
		return argMustBeTypeAn(in.callToken, "dictInt", "dictionary")
	})
	rbTreeFunc("dictStr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxDict, ok := args[0].(*LoxDict); ok {
			m := map[string]any{}
			it := loxDict.Iterator()
			for it.HasNext() {
				pair := it.Next().(*LoxList).elements
				var key string
				switch pairKey := pair[0].(type) {
				case *LoxString:
					key = pairKey.str
				default:
					return nil, loxerror.RuntimeError(in.callToken,
						"Dictionary argument to 'rbtree class.dictStr' must only have string keys.")
				}
				value := pair[1]
				m[key] = value
			}
			return NewLoxRBTreeMapStr(m), nil
		}
		return argMustBeTypeAn(in.callToken, "dictStr", "dictionary")
	})
	rbTreeFunc("int", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxRBTreeIntKeys(), nil
	})
	rbTreeFunc("intArgs", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"rbtree class.intArgs: expected at least 2 arguments but got %v.",
					argsLen,
				),
			)
		}
		if argsLen%2 != 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"rbtree class.intArgs: number of arguments cannot be an odd number.")
		}
		m := map[int64]any{}
		var intVar int64
		for i, arg := range args {
			if i%2 == 0 {
				switch arg := arg.(type) {
				case int64:
					intVar = arg
				default:
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"rbtree class.intArgs: argument #%v must be an integer.",
							i+1,
						),
					)
				}
			} else {
				m[intVar] = arg
			}
		}
		return NewLoxRBTreeMapInt(m), nil
	})
	rbTreeFunc("iter", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Iterable); ok {
			elements := []any{}
			switch it := element.Iterator().(type) {
			case interfaces.IteratorErr:
				for {
					ok, hasNextErr := it.HasNextErr()
					if hasNextErr != nil {
						return nil, loxerror.RuntimeError(in.callToken, hasNextErr.Error())
					}
					if !ok {
						break
					}
					next, nextErr := it.NextErr()
					if nextErr != nil {
						return nil, loxerror.RuntimeError(in.callToken, nextErr.Error())
					}
					elements = append(elements, next)
				}
			default:
				for it.HasNext() {
					elements = append(elements, it.Next())
				}
			}
			return NewLoxRBTreeArgs(elements...), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("rbtree class.iter: type '%v' is not iterable.", getType(args[0])))
	})
	rbTreeFunc("list", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxList, ok := args[0].(*LoxList); ok {
			return NewLoxRBTreeArgs(loxList.elements...), nil
		}
		return argMustBeTypeAn(in.callToken, "list", "list")
	})
	rbTreeFunc("rbtree", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxRBTree, ok := args[0].(*LoxRBTree); ok {
			return NewLoxRBTree(loxRBTree.tree, loxRBTree.keyType), nil
		}
		return argMustBeTypeAn(in.callToken, "rbtree", "rb tree")
	})
	rbTreeFunc("str", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxRBTreeStringKeys(), nil
	})
	rbTreeFunc("strArgs", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"rbtree class.strArgs: expected at least 2 arguments but got %v.",
					argsLen,
				),
			)
		}
		if argsLen%2 != 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"rbtree class.strArgs: number of arguments cannot be an odd number.")
		}
		m := map[string]any{}
		var str string
		for i, arg := range args {
			if i%2 == 0 {
				switch arg := arg.(type) {
				case *LoxString:
					str = arg.str
				default:
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"rbtree class.strArgs: argument #%v must be a string.",
							i+1,
						),
					)
				}
			} else {
				m[str] = arg
			}
		}
		return NewLoxRBTreeMapStr(m), nil
	})

	i.globals.Define(className, rbTreeClass)
}
