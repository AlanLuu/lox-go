package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) definePriQueueFuncs() {
	className := "pqueue"
	priQueueClass := NewLoxClass(className, nil, false)
	priQueueFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native priority queue class fn %v at %p>", name, &s)
		}
		priQueueClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'priority queue class.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	priQueueFunc("args", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"priority queue class.args: expected at least 2 arguments but got %v.",
					argsLen,
				),
			)
		}
		if argsLen%2 != 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"priority queue class.args: number of arguments cannot be an odd number.")
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
							"priority queue class.args: argument #%v must be an integer.",
							i+1,
						),
					)
				}
			} else {
				m[intVar] = arg
			}
		}
		return NewLoxPriorityQueueMap(m, loxPriorityQueueOpts_def), nil
	})
	priQueueFunc("argsReversed", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"priority queue class.argsReversed: expected at least 2 arguments but got %v.",
					argsLen,
				),
			)
		}
		if argsLen%2 != 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"priority queue class.argsReversed: number of arguments cannot be an odd number.")
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
							"priority queue class.argsReversed: argument #%v must be an integer.",
							i+1,
						),
					)
				}
			} else {
				m[intVar] = arg
			}
		}
		return NewLoxPriorityQueueMap(m, loxPriorityQueueOpts_reversed), nil
	})
	priQueueFunc("builder", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxPriorityQueueBuilder(), nil
	})
	priQueueFunc("dict", 1, func(in *Interpreter, args list.List[any]) (any, error) {
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
						"Dictionary argument to 'priority queue class.dict' must only have integer keys.")
				}
				value := pair[1]
				m[key] = value
			}
			return NewLoxPriorityQueueMap(m, loxPriorityQueueOpts_def), nil
		}
		return argMustBeType(in.callToken, "dict", "dictionary")
	})
	priQueueFunc("dictReversed", 1, func(in *Interpreter, args list.List[any]) (any, error) {
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
						"Dictionary argument to 'priority queue class.dictReversed' must only have integer keys.")
				}
				value := pair[1]
				m[key] = value
			}
			return NewLoxPriorityQueueMap(m, loxPriorityQueueOpts_reversed), nil
		}
		return argMustBeType(in.callToken, "dict", "dictionary")
	})
	priQueueFunc("dup", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxPriorityQueue(loxPriorityQueueOpts_allowDup), nil
	})
	priQueueFunc("dupReversed", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxPriorityQueue(loxPriorityQueueOpts_allowDupReversed), nil
	})
	priQueueFunc("new", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxPriorityQueue(loxPriorityQueueOpts_def), nil
	})
	priQueueFunc("newReversed", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxPriorityQueue(loxPriorityQueueOpts_reversed), nil
	})
	priQueueFunc("reversed", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxPriorityQueue(loxPriorityQueueOpts_reversed), nil
	})

	i.globals.Define(className, priQueueClass)
}
