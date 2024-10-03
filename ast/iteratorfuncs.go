package ast

import (
	crand "crypto/rand"
	"fmt"
	"math/big"

	"github.com/AlanLuu/lox/bignum/bigfloat"
	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func defineIteratorFields(iteratorClass *LoxClass) {
	urandom := InfiniteIterator{}
	urandom.nextMethod = func() any {
		numBig, numErr := crand.Int(crand.Reader, bigint.TwoFiveSix)
		if numErr != nil {
			panic(numErr)
		}
		return numBig.Int64()
	}
	iteratorClass.classProperties["urandom"] = NewLoxIterator(urandom)

	zeroes := InfiniteIterator{}
	zeroes.nextMethod = func() any {
		return int64(0)
	}
	iteratorClass.classProperties["zeroes"] = NewLoxIterator(zeroes)
}

func (i *Interpreter) defineIteratorFuncs() {
	className := "Iterator"
	iteratorClass := NewLoxClass(className, nil, false)
	iteratorFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native Iterator class fn %v at %p>", name, &s)
		}
		iteratorClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'Iterator.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	defineIteratorFields(iteratorClass)
	iteratorFunc("args", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		index := 0
		iterator := ProtoIterator{}
		iterator.hasNextMethod = func() bool {
			return index < argsLen
		}
		iterator.nextMethod = func() any {
			element := args[index]
			index++
			return element
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("batched", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.batched' is not iterable.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.batched' must be an integer.")
		}
		length := args[1].(int64)
		if length <= 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.batched' must be at least 1.")
		}
		iterableIterator := args[0].(interfaces.Iterable).Iterator()
		iterator := ProtoIterator{}
		iterator.hasNextMethod = func() bool {
			return iterableIterator.HasNext()
		}
		iterator.nextMethod = func() any {
			elements := list.NewListCap[any](length)
			for i := int64(0); i < length; i++ {
				if !iterableIterator.HasNext() {
					break
				}
				elements.Add(iterableIterator.Next())
			}
			return NewLoxList(elements)
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("chain", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return EmptyLoxIterator(), nil
		}
		argIterators := list.NewListCap[interfaces.Iterator](int64(len(args)))
		for _, arg := range args {
			switch arg := arg.(type) {
			case interfaces.Iterable:
				argIterators.Add(arg.Iterator())
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"All arguments to 'Iterator.chain' must be iterables.")
			}
		}
		iterator := ProtoIterator{}
		iteratorIndex := 0
		iterator.hasNextMethod = func() bool {
			if !argIterators[iteratorIndex].HasNext() {
				for i := iteratorIndex + 1; i < len(argIterators); i++ {
					if argIterators[i].HasNext() {
						iteratorIndex = i
						return true
					}
				}
				return false
			}
			return true
		}
		iterator.nextMethod = func() any {
			return argIterators[iteratorIndex].Next()
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("countFloat", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var start, step any
		argsLen := len(args)
		switch argsLen {
		case 1, 2:
			switch args[0].(type) {
			case int64:
			case *big.Int:
			case float64:
			case *big.Float:
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'Iterator.countFloat' must be an integer, bigint, float, or bigfloat.")
			}
			start = args[0]
			if argsLen == 2 {
				switch args[1].(type) {
				case float64:
				case *big.Float:
				default:
					return nil, loxerror.RuntimeError(in.callToken,
						"Second argument to 'Iterator.countFloat' must be a float or bigfloat.")
				}
				step = args[1]
			} else {
				step = float64(1.0)
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
		iterator := InfiniteIterator{}
		switch start := start.(type) {
		case int64:
			startFloat := float64(start)
			firstIteration := true
			switch step := step.(type) {
			case float64:
				iterator.nextMethod = func() any {
					var num any
					if firstIteration {
						num = start
						firstIteration = false
					} else {
						num = startFloat
					}
					startFloat += step
					return num
				}
			case *big.Float:
				iterator.nextMethod = func() any {
					var num any
					if firstIteration {
						num = start
						firstIteration = false
					} else {
						num = startFloat
					}
					stepFloat, _ := step.Float64()
					startFloat += stepFloat
					return num
				}
			}
		case *big.Int:
			bigFloatStart := new(big.Float).SetInt(start)
			firstIteration := true
			switch step := step.(type) {
			case float64:
				bigFloatStep := big.NewFloat(step)
				iterator.nextMethod = func() any {
					var num any
					if firstIteration {
						num = new(big.Int).Set(start)
						firstIteration = false
					} else {
						num = new(big.Float).Set(bigFloatStart)
					}
					bigFloatStart.Add(bigFloatStart, bigFloatStep)
					return num
				}
			case *big.Float:
				iterator.nextMethod = func() any {
					var num any
					if firstIteration {
						num = new(big.Int).Set(start)
						firstIteration = false
					} else {
						num = new(big.Float).Set(bigFloatStart)
					}
					bigFloatStart.Add(bigFloatStart, step)
					return num
				}
			}
		case float64:
			switch step := step.(type) {
			case float64:
				iterator.nextMethod = func() any {
					num := start
					start += step
					return num
				}
			case *big.Float:
				iterator.nextMethod = func() any {
					num := start
					stepFloat, _ := step.Float64()
					start += stepFloat
					return num
				}
			}
		case *big.Float:
			bigFloatStart := new(big.Float).Set(start)
			switch step := step.(type) {
			case float64:
				bigFloatStep := big.NewFloat(step)
				iterator.nextMethod = func() any {
					num := new(big.Float).Set(bigFloatStart)
					bigFloatStart.Add(bigFloatStart, bigFloatStep)
					return num
				}
			case *big.Float:
				iterator.nextMethod = func() any {
					num := new(big.Float).Set(bigFloatStart)
					bigFloatStart.Add(bigFloatStart, step)
					return num
				}
			}
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("countInt", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var start, step any
		argsLen := len(args)
		switch argsLen {
		case 1, 2:
			switch args[0].(type) {
			case int64:
			case *big.Int:
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'Iterator.countInt' must be an integer or bigint.")
			}
			start = args[0]
			if argsLen == 2 {
				switch args[1].(type) {
				case int64:
				case *big.Int:
				default:
					return nil, loxerror.RuntimeError(in.callToken,
						"Second argument to 'Iterator.countInt' must be an integer or bigint.")
				}
				step = args[1]
			} else {
				step = int64(1)
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
		iterator := InfiniteIterator{}
		switch start := start.(type) {
		case int64:
			switch step := step.(type) {
			case int64:
				iterator.nextMethod = func() any {
					num := start
					start += step
					return num
				}
			case *big.Int:
				iterator.nextMethod = func() any {
					num := start
					start += step.Int64()
					return num
				}
			}
		case *big.Int:
			bigIntStart := new(big.Int).Set(start)
			switch step := step.(type) {
			case int64:
				bigIntStep := big.NewInt(step)
				iterator.nextMethod = func() any {
					num := new(big.Int).Set(bigIntStart)
					bigIntStart.Add(bigIntStart, bigIntStep)
					return num
				}
			case *big.Int:
				iterator.nextMethod = func() any {
					num := new(big.Int).Set(bigIntStart)
					bigIntStart.Add(bigIntStart, step)
					return num
				}
			}
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("cycle", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if iterable, ok := args[0].(interfaces.Iterable); ok {
			elements := list.NewList[any]()
			it := iterable.Iterator()
			atLeastOne := false
			newIterator := ProtoIterator{}
			newIterator.hasNextMethod = func() bool {
				if !atLeastOne && it.HasNext() {
					atLeastOne = true
				}
				return atLeastOne
			}
			elementsIndex := 0
			newIterator.nextMethod = func() any {
				var next any
				if it.HasNext() {
					next = it.Next()
					elements.Add(next)
				} else {
					next = elements[elementsIndex]
					elementsIndex = (elementsIndex + 1) % len(elements)
				}
				return next
			}
			return NewLoxIterator(newIterator), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Type '%v' is not iterable.", getType(args[0])))
	})
	iteratorFunc("enumerate", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var it interfaces.Iterator
		var index any
		argsLen := len(args)
		switch argsLen {
		case 1, 2:
			if _, ok := args[0].(interfaces.Iterable); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'Iterator.enumerate' is not iterable.")
			}
			it = args[0].(interfaces.Iterable).Iterator()
			if argsLen == 2 {
				switch args[1].(type) {
				case int64:
				case *big.Int:
				case float64:
				case *big.Float:
				default:
					return nil, loxerror.RuntimeError(in.callToken,
						"Second argument to 'Iterator.enumerate' must be an integer, bigint, float, or bigfloat.")
				}
				index = args[1]
			} else {
				index = int64(0)
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
		iterable := ProtoIterator{}
		iterable.hasNextMethod = func() bool {
			return it.HasNext()
		}
		switch index := index.(type) {
		case int64:
			iterable.nextMethod = func() any {
				entry := list.NewListCap[any](2)
				entry.Add(index)
				index++
				entry.Add(it.Next())
				return NewLoxList(entry)
			}
		case *big.Int:
			indexCopy := new(big.Int).Set(index)
			iterable.nextMethod = func() any {
				entry := list.NewListCap[any](2)
				entry.Add(new(big.Int).Set(indexCopy))
				indexCopy.Add(indexCopy, bigint.One)
				entry.Add(it.Next())
				return NewLoxList(entry)
			}
		case float64:
			iterable.nextMethod = func() any {
				entry := list.NewListCap[any](2)
				entry.Add(index)
				index++
				entry.Add(it.Next())
				return NewLoxList(entry)
			}
		case *big.Float:
			indexCopy := new(big.Float).Set(index)
			iterable.nextMethod = func() any {
				entry := list.NewListCap[any](2)
				entry.Add(new(big.Float).Set(indexCopy))
				indexCopy.Add(indexCopy, bigfloat.One)
				entry.Add(it.Next())
				return NewLoxList(entry)
			}
		}
		return NewLoxIterator(iterable), nil
	})
	iteratorFunc("infiniteArg", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		arg := args[0]
		iterator := InfiniteIterator{}
		iterator.nextMethod = func() any {
			return arg
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("pairwise", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if iterable, ok := args[0].(interfaces.Iterable); ok {
			it := iterable.Iterator()
			if !it.HasNext() {
				return EmptyLoxIterator(), nil
			}
			first := it.Next()
			if !it.HasNext() {
				return EmptyLoxIterator(), nil
			}
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				return it.HasNext()
			}
			iterator.nextMethod = func() any {
				pair := list.NewListCap[any](2)
				pair.Add(first)
				next := it.Next()
				pair.Add(next)
				first = next
				return NewLoxList(pair)
			}
			return NewLoxIterator(iterator), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Type '%v' is not iterable.", getType(args[0])))
	})
	iteratorFunc("repeat", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var element any
		var repeatCount *big.Int
		var isInfinite bool
		argsLen := len(args)
		switch argsLen {
		case 1, 2:
			element = args[0]
			if argsLen == 2 {
				switch arg := args[1].(type) {
				case int64:
					repeatCount = big.NewInt(arg)
				case *big.Int:
					repeatCount = new(big.Int).Set(arg)
				default:
					return nil, loxerror.RuntimeError(in.callToken,
						"Second argument to 'Iterator.repeat' must be an integer or bigint.")
				}
			} else {
				isInfinite = true
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
		iterator := ProtoIterator{}
		if isInfinite {
			iterator.hasNextMethod = func() bool {
				return true
			}
			iterator.nextMethod = func() any {
				return element
			}
		} else {
			count := big.NewInt(0)
			iterator.hasNextMethod = func() bool {
				return count.Cmp(repeatCount) < 0
			}
			iterator.nextMethod = func() any {
				count.Add(count, bigint.One)
				return element
			}
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("reversed", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.ReverseIterable); ok {
			return NewLoxIterator(element.ReverseIterator()), nil
		}
		return argMustBeType(in.callToken, "reversed", "buffer, list, or string")
	})
	iteratorFunc("zip", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return EmptyLoxIterator(), nil
		}
		argIterators := list.NewListCap[interfaces.Iterator](int64(len(args)))
		for _, arg := range args {
			switch arg := arg.(type) {
			case interfaces.Iterable:
				argIterators.Add(arg.Iterator())
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"All arguments to 'Iterator.zip' must be iterables.")
			}
		}
		iterator := ProtoIterator{}
		iterator.hasNextMethod = func() bool {
			for _, argIterator := range argIterators {
				if !argIterator.HasNext() {
					return false
				}
			}
			return true
		}
		iterator.nextMethod = func() any {
			elements := list.NewListCap[any](int64(len(argIterators)))
			for _, argIterator := range argIterators {
				elements.Add(argIterator.Next())
			}
			return NewLoxList(elements)
		}
		return NewLoxIterator(iterator), nil
	})

	i.globals.Define(className, iteratorClass)
}
