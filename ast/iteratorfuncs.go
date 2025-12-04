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
	urandom := InfiniteIteratorErr{legacyPanicOnErr: true}
	urandom.nextMethod = func() (any, error) {
		numBig, numErr := crand.Int(crand.Reader, bigint.TwoFiveSix)
		if numErr != nil {
			return nil, numErr
		}
		return numBig.Int64(), nil
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

	defineIteratorFields(iteratorClass)
	iteratorFunc("accumulate", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 2 && argsLen != 3 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.accumulate' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.accumulate' must be a function.")
		}
		it := args[0].(interfaces.Iterable).Iterator()
		itWithErr, isErrIter := it.(interfaces.IteratorErr)
		hasNext := func() (bool, error) {
			if isErrIter {
				return itWithErr.HasNextErr()
			}
			return it.HasNext(), nil
		}
		next := func() (any, error) {
			if isErrIter {
				return itWithErr.NextErr()
			}
			return it.Next(), nil
		}
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 3)
		var index int64 = 0
		var value any
		firstIterOuter := true
		iterator := ProtoIteratorErr{legacyPanicOnErr: true}
		iterator.hasNextMethod = func() (bool, error) {
			if argList == nil {
				return false, nil
			}
			firstIter := firstIterOuter
			firstIterOuter = false
			if firstIter && argsLen == 3 {
				value = args[2]
				return true, nil
			}
			ok, hasNextErr := hasNext()
			if hasNextErr != nil {
				return false, hasNextErr
			}
			if !ok {
				argList.Clear()
				return false, nil
			}
			element, nextErr := next()
			if nextErr != nil {
				return false, nextErr
			}
			if firstIter {
				value = element
			} else {
				argList[0] = value
				argList[1] = element
				argList[2] = index
				index++
				var valueErr error
				value, valueErr = callback.call(i, argList)
				if valueReturn, ok := value.(Return); ok {
					value = valueReturn.FinalValue
				} else if valueErr != nil {
					return false, valueErr
				}
			}
			return true, nil
		}
		iterator.nextMethod = func() (any, error) {
			return value, nil
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("accumulateAdd", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 1 && argsLen != 2 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(interfaces.Iterable); !ok {
			if argsLen == 1 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'Iterator.accumulateAdd' is not iterable.")
			}
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.accumulateAdd' is not iterable.")
		}
		it := args[0].(interfaces.Iterable).Iterator()
		itWithErr, isErrIter := it.(interfaces.IteratorErr)
		hasNext := func() (bool, error) {
			if isErrIter {
				return itWithErr.HasNextErr()
			}
			return it.HasNext(), nil
		}
		next := func() (any, error) {
			if isErrIter {
				return itWithErr.NextErr()
			}
			return it.Next(), nil
		}
		var value any
		firstIterOuter := true
		iterator := ProtoIteratorErr{legacyPanicOnErr: true}
		iterator.hasNextMethod = func() (bool, error) {
			firstIter := firstIterOuter
			firstIterOuter = false
			if firstIter && argsLen == 2 {
				value = args[1]
				return true, nil
			}
			ok, hasNextErr := hasNext()
			if hasNextErr != nil {
				return false, hasNextErr
			}
			if !ok {
				return false, nil
			}
			element, nextErr := next()
			if nextErr != nil {
				return false, nextErr
			}
			if firstIter {
				value = element
			} else {
				var addErr error
				value, addErr = i.visitBinaryExpr(Binary{
					Literal{value},
					&token.Token{
						TokenType: token.PLUS,
						Lexeme:    "+",
					},
					Literal{element},
				})
				if addErr != nil {
					return false, addErr
				}
			}
			return true, nil
		}
		iterator.nextMethod = func() (any, error) {
			return value, nil
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("all", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if iterable, ok := args[0].(interfaces.Iterable); ok {
			it := iterable.Iterator()
			switch it := it.(type) {
			case interfaces.IteratorErr:
				for {
					ok, hasNextErr := it.HasNextErr()
					if hasNextErr != nil {
						return nil, hasNextErr
					}
					if !ok {
						break
					}
					result, nextErr := it.NextErr()
					if nextErr != nil {
						return nil, nextErr
					}
					if !i.isTruthy(result) {
						return false, nil
					}
				}
			default:
				for it.HasNext() {
					if !i.isTruthy(it.Next()) {
						return false, nil
					}
				}
			}
			return true, nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Iterator.all: type '%v' is not iterable.", getType(args[0])))
	})
	iteratorFunc("allFunc", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.allFunc' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.allFunc' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		defer argList.Clear()
		var index int64 = 0
		switch it := arg0Iterator.(type) {
		case interfaces.IteratorErr:
			for ; ; index++ {
				ok, hasNextErr := it.HasNextErr()
				if hasNextErr != nil {
					return nil, hasNextErr
				}
				if !ok {
					break
				}
				element, nextErr := it.NextErr()
				if nextErr != nil {
					return nil, nextErr
				}
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
		default:
			for ; it.HasNext(); index++ {
				argList[0] = it.Next()
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
		}
		return true, nil
	})
	iteratorFunc("any", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if iterable, ok := args[0].(interfaces.Iterable); ok {
			it := iterable.Iterator()
			switch it := it.(type) {
			case interfaces.IteratorErr:
				for {
					ok, hasNextErr := it.HasNextErr()
					if hasNextErr != nil {
						return nil, hasNextErr
					}
					if !ok {
						break
					}
					result, nextErr := it.NextErr()
					if nextErr != nil {
						return nil, nextErr
					}
					if i.isTruthy(result) {
						return true, nil
					}
				}
			default:
				for it.HasNext() {
					if i.isTruthy(it.Next()) {
						return true, nil
					}
				}
			}
			return false, nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Iterator.any: type '%v' is not iterable.", getType(args[0])))
	})
	iteratorFunc("anyFunc", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.anyFunc' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.anyFunc' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		defer argList.Clear()
		var index int64 = 0
		switch it := arg0Iterator.(type) {
		case interfaces.IteratorErr:
			for ; ; index++ {
				ok, hasNextErr := it.HasNextErr()
				if hasNextErr != nil {
					return nil, hasNextErr
				}
				if !ok {
					break
				}
				element, nextErr := it.NextErr()
				if nextErr != nil {
					return nil, nextErr
				}
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
		default:
			for ; it.HasNext(); index++ {
				argList[0] = it.Next()
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
		}
		return false, nil
	})
	iteratorFunc("args", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
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
	iteratorFunc("count", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.count' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.count' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		arg0IteratorWithErr, isErrIter := arg0Iterator.(interfaces.IteratorErr)
		callback := args[1].(*LoxFunction)
		hasNext := func() (bool, error) {
			if isErrIter {
				return arg0IteratorWithErr.HasNextErr()
			}
			return arg0Iterator.HasNext(), nil
		}
		next := func() (any, error) {
			if isErrIter {
				return arg0IteratorWithErr.NextErr()
			}
			return arg0Iterator.Next(), nil
		}
		var count int64 = 0
		argList := getArgList(callback, 2)
		defer argList.Clear()
		for index := int64(0); ; index++ {
			if count < 0 {
				//Integer overflow, return max 64-bit signed value
				return int64((1 << 63) - 1), nil
			}
			ok, hasNextErr := hasNext()
			if hasNextErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, hasNextErr.Error())
			}
			if !ok {
				return count, nil
			}
			var nextErr error
			argList[0], nextErr = next()
			if nextErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, nextErr.Error())
			}
			argList[1] = index
			result, resultErr := callback.call(i, argList)
			if resultReturn, ok := result.(Return); ok {
				result = resultReturn.FinalValue
			} else if resultErr != nil {
				return nil, resultErr
			}
			if i.isTruthy(result) {
				count++
			}
		}
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
	iteratorFunc("countTrue", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if iterable, ok := args[0].(interfaces.Iterable); ok {
			it := iterable.Iterator()
			itWithErr, isErrIter := it.(interfaces.IteratorErr)
			hasNext := func() (bool, error) {
				if isErrIter {
					return itWithErr.HasNextErr()
				}
				return it.HasNext(), nil
			}
			next := func() (any, error) {
				if isErrIter {
					return itWithErr.NextErr()
				}
				return it.Next(), nil
			}
			var count int64 = 0
			for {
				if count < 0 {
					//Integer overflow, return max 64-bit signed value
					return int64((1 << 63) - 1), nil
				}
				ok, hasNextErr := hasNext()
				if hasNextErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, hasNextErr.Error())
				}
				if !ok {
					return count, nil
				}
				element, nextErr := next()
				if nextErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, nextErr.Error())
				}
				if i.isTruthy(element) {
					count++
				}
			}
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Iterator.countTrue: type '%v' is not iterable.", getType(args[0])))
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
					if elementsIndex >= len(elements)-1 {
						elementsIndex = 0
					} else {
						elementsIndex++
					}
				}
				return next
			}
			return NewLoxIterator(newIterator), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Iterator.cycle: type '%v' is not iterable.", getType(args[0])))
	})
	iteratorFunc("dropuntil", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.dropuntil' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.dropuntil' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		loopStopped := false
		var index int64 = 0
		var current any
		iterator := ProtoIteratorErr{legacyPanicOnErr: true}
		switch arg0Iterator := arg0Iterator.(type) {
		case interfaces.IteratorErr:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				if loopStopped {
					ok, hasNextErr := arg0Iterator.HasNextErr()
					if hasNextErr != nil {
						return false, hasNextErr
					}
					if !ok {
						argList.Clear()
						return false, nil
					}
					element, nextErr := arg0Iterator.NextErr()
					if nextErr != nil {
						return false, nextErr
					}
					current = element
					return true, nil
				}
				for {
					ok, hasNextErr := arg0Iterator.HasNextErr()
					if hasNextErr != nil {
						return false, hasNextErr
					}
					if !ok {
						argList.Clear()
						return false, nil
					}
					element, nextErr := arg0Iterator.NextErr()
					if nextErr != nil {
						return false, nextErr
					}
					argList[0] = element
					argList[1] = index
					index++
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return false, resultErr
					}
					if i.isTruthy(result) {
						loopStopped = true
						current = element
						return true, nil
					}
				}
			}
		default:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				if loopStopped {
					if !arg0Iterator.HasNext() {
						argList.Clear()
						return false, nil
					}
					current = arg0Iterator.Next()
					return true, nil
				}
				for {
					if !arg0Iterator.HasNext() {
						argList.Clear()
						return false, nil
					}
					element := arg0Iterator.Next()
					argList[0] = element
					argList[1] = index
					index++
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return false, resultErr
					}
					if i.isTruthy(result) {
						loopStopped = true
						current = element
						return true, nil
					}
				}
			}
		}
		iterator.nextMethod = func() (any, error) {
			return current, nil
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("dropwhile", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.dropwhile' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.dropwhile' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		loopStopped := false
		var index int64 = 0
		var current any
		iterator := ProtoIteratorErr{legacyPanicOnErr: true}
		switch arg0Iterator := arg0Iterator.(type) {
		case interfaces.IteratorErr:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				if loopStopped {
					ok, hasNextErr := arg0Iterator.HasNextErr()
					if hasNextErr != nil {
						return false, hasNextErr
					}
					if !ok {
						argList.Clear()
						return false, nil
					}
					element, nextErr := arg0Iterator.NextErr()
					if nextErr != nil {
						return false, nextErr
					}
					current = element
					return true, nil
				}
				for {
					ok, hasNextErr := arg0Iterator.HasNextErr()
					if hasNextErr != nil {
						return false, hasNextErr
					}
					if !ok {
						argList.Clear()
						return false, nil
					}
					element, nextErr := arg0Iterator.NextErr()
					if nextErr != nil {
						return false, nextErr
					}
					argList[0] = element
					argList[1] = index
					index++
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return false, resultErr
					}
					if !i.isTruthy(result) {
						loopStopped = true
						current = element
						return true, nil
					}
				}
			}
		default:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				if loopStopped {
					if !arg0Iterator.HasNext() {
						argList.Clear()
						return false, nil
					}
					current = arg0Iterator.Next()
					return true, nil
				}
				for {
					if !arg0Iterator.HasNext() {
						argList.Clear()
						return false, nil
					}
					element := arg0Iterator.Next()
					argList[0] = element
					argList[1] = index
					index++
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return false, resultErr
					}
					if !i.isTruthy(result) {
						loopStopped = true
						current = element
						return true, nil
					}
				}
			}
		}
		iterator.nextMethod = func() (any, error) {
			return current, nil
		}
		return NewLoxIterator(iterator), nil
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
	iteratorFunc("filter", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.filter' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.filter' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		var index int64 = 0
		var current any
		iterator := ProtoIteratorErr{legacyPanicOnErr: true}
		switch arg0Iterator := arg0Iterator.(type) {
		case interfaces.IteratorErr:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				for {
					ok, hasNextErr := arg0Iterator.HasNextErr()
					if hasNextErr != nil {
						return false, hasNextErr
					}
					if !ok {
						argList.Clear()
						return false, nil
					}
					element, nextErr := arg0Iterator.NextErr()
					if nextErr != nil {
						return false, nextErr
					}
					argList[0] = element
					argList[1] = index
					index++
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return false, resultErr
					}
					if i.isTruthy(result) {
						current = element
						return true, nil
					}
				}
			}
		default:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				for {
					if !arg0Iterator.HasNext() {
						argList.Clear()
						return false, nil
					}
					element := arg0Iterator.Next()
					argList[0] = element
					argList[1] = index
					index++
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return false, resultErr
					}
					if i.isTruthy(result) {
						current = element
						return true, nil
					}
				}
			}
		}
		iterator.nextMethod = func() (any, error) {
			return current, nil
		}
		return NewLoxIteratorTag(iterator, "filter"), nil
	})
	iteratorFunc("filterfalse", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.filterfalse' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.filterfalse' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		var index int64 = 0
		var current any
		iterator := ProtoIteratorErr{legacyPanicOnErr: true}
		switch arg0Iterator := arg0Iterator.(type) {
		case interfaces.IteratorErr:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				for {
					ok, hasNextErr := arg0Iterator.HasNextErr()
					if hasNextErr != nil {
						return false, hasNextErr
					}
					if !ok {
						argList.Clear()
						return false, nil
					}
					element, nextErr := arg0Iterator.NextErr()
					if nextErr != nil {
						return false, nextErr
					}
					argList[0] = element
					argList[1] = index
					index++
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return false, resultErr
					}
					if !i.isTruthy(result) {
						current = element
						return true, nil
					}
				}
			}
		default:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				for {
					if !arg0Iterator.HasNext() {
						argList.Clear()
						return false, nil
					}
					element := arg0Iterator.Next()
					argList[0] = element
					argList[1] = index
					index++
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return false, resultErr
					}
					if !i.isTruthy(result) {
						current = element
						return true, nil
					}
				}
			}
		}
		iterator.nextMethod = func() (any, error) {
			return current, nil
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("filterfalseonly", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if iterable, ok := args[0].(interfaces.Iterable); ok {
			var current any
			stopped := false
			it := iterable.Iterator()
			iterator := ProtoIteratorErr{legacyPanicOnErr: true}
			switch it := it.(type) {
			case interfaces.IteratorErr:
				iterator.hasNextMethod = func() (bool, error) {
					if stopped {
						return false, nil
					}
					for {
						ok, hasNextErr := it.HasNextErr()
						if hasNextErr != nil {
							return false, hasNextErr
						}
						if !ok {
							stopped = true
							return false, nil
						}
						element, nextErr := it.NextErr()
						if nextErr != nil {
							return false, nextErr
						}
						if !i.isTruthy(element) {
							current = element
							return true, nil
						}
					}
				}
			default:
				iterator.hasNextMethod = func() (bool, error) {
					if stopped {
						return false, nil
					}
					for {
						if !it.HasNext() {
							stopped = true
							return false, nil
						}
						element := it.Next()
						if !i.isTruthy(element) {
							current = element
							return true, nil
						}
					}
				}
			}
			iterator.nextMethod = func() (any, error) {
				return current, nil
			}
			return NewLoxIterator(iterator), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Iterator.filterfalseonly: type '%v' is not iterable.", getType(args[0])))
	})
	iteratorFunc("filtertrueonly", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if iterable, ok := args[0].(interfaces.Iterable); ok {
			var current any
			stopped := false
			it := iterable.Iterator()
			iterator := ProtoIteratorErr{legacyPanicOnErr: true}
			switch it := it.(type) {
			case interfaces.IteratorErr:
				iterator.hasNextMethod = func() (bool, error) {
					if stopped {
						return false, nil
					}
					for {
						ok, hasNextErr := it.HasNextErr()
						if hasNextErr != nil {
							return false, hasNextErr
						}
						if !ok {
							stopped = true
							return false, nil
						}
						element, nextErr := it.NextErr()
						if nextErr != nil {
							return false, nextErr
						}
						if i.isTruthy(element) {
							current = element
							return true, nil
						}
					}
				}
			default:
				iterator.hasNextMethod = func() (bool, error) {
					if stopped {
						return false, nil
					}
					for {
						if !it.HasNext() {
							stopped = true
							return false, nil
						}
						element := it.Next()
						if i.isTruthy(element) {
							current = element
							return true, nil
						}
					}
				}
			}
			iterator.nextMethod = func() (any, error) {
				return current, nil
			}
			return NewLoxIterator(iterator), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Iterator.filtertrueonly: type '%v' is not iterable.", getType(args[0])))
	})
	iteratorFunc("getuntil", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.getuntil' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.getuntil' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		var index int64 = 0
		var current any
		iterator := ProtoIteratorErr{legacyPanicOnErr: true}
		switch arg0Iterator := arg0Iterator.(type) {
		case interfaces.IteratorErr:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				ok, hasNextErr := arg0Iterator.HasNextErr()
				if hasNextErr != nil {
					return false, hasNextErr
				}
				if !ok {
					argList.Clear()
					return false, nil
				}
				element, nextErr := arg0Iterator.NextErr()
				if nextErr != nil {
					return false, nextErr
				}
				argList[0] = element
				argList[1] = index
				index++
				result, resultErr := callback.call(i, argList)
				if resultReturn, ok := result.(Return); ok {
					result = resultReturn.FinalValue
				} else if resultErr != nil {
					return false, resultErr
				}
				if i.isTruthy(result) {
					argList.Clear()
					return false, nil
				}
				current = element
				return true, nil
			}
		default:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				if !arg0Iterator.HasNext() {
					argList.Clear()
					return false, nil
				}
				element := arg0Iterator.Next()
				argList[0] = element
				argList[1] = index
				index++
				result, resultErr := callback.call(i, argList)
				if resultReturn, ok := result.(Return); ok {
					result = resultReturn.FinalValue
				} else if resultErr != nil {
					return false, resultErr
				}
				if i.isTruthy(result) {
					argList.Clear()
					return false, nil
				}
				current = element
				return true, nil
			}
		}
		iterator.nextMethod = func() (any, error) {
			return current, nil
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("getuntillast", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.getuntillast' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.getuntillast' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		var index int64 = 0
		var current any
		iterator := ProtoIteratorErr{legacyPanicOnErr: true}
		switch arg0Iterator := arg0Iterator.(type) {
		case interfaces.IteratorErr:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				ok, hasNextErr := arg0Iterator.HasNextErr()
				if hasNextErr != nil {
					return false, hasNextErr
				}
				if !ok {
					argList.Clear()
					return false, nil
				}
				element, nextErr := arg0Iterator.NextErr()
				if nextErr != nil {
					return false, nextErr
				}
				argList[0] = element
				argList[1] = index
				index++
				result, resultErr := callback.call(i, argList)
				if resultReturn, ok := result.(Return); ok {
					result = resultReturn.FinalValue
				} else if resultErr != nil {
					return false, resultErr
				}
				if i.isTruthy(result) {
					argList.Clear()
				}
				current = element
				return true, nil
			}
		default:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				if !arg0Iterator.HasNext() {
					argList.Clear()
					return false, nil
				}
				element := arg0Iterator.Next()
				argList[0] = element
				argList[1] = index
				index++
				result, resultErr := callback.call(i, argList)
				if resultReturn, ok := result.(Return); ok {
					result = resultReturn.FinalValue
				} else if resultErr != nil {
					return false, resultErr
				}
				if i.isTruthy(result) {
					argList.Clear()
				}
				current = element
				return true, nil
			}
		}
		iterator.nextMethod = func() (any, error) {
			return current, nil
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("getwhile", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.getwhile' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.getwhile' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		var index int64 = 0
		var current any
		iterator := ProtoIteratorErr{legacyPanicOnErr: true}
		switch arg0Iterator := arg0Iterator.(type) {
		case interfaces.IteratorErr:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				ok, hasNextErr := arg0Iterator.HasNextErr()
				if hasNextErr != nil {
					return false, hasNextErr
				}
				if !ok {
					argList.Clear()
					return false, nil
				}
				element, nextErr := arg0Iterator.NextErr()
				if nextErr != nil {
					return false, nextErr
				}
				argList[0] = element
				argList[1] = index
				index++
				result, resultErr := callback.call(i, argList)
				if resultReturn, ok := result.(Return); ok {
					result = resultReturn.FinalValue
				} else if resultErr != nil {
					return false, resultErr
				}
				if !i.isTruthy(result) {
					argList.Clear()
					return false, nil
				}
				current = element
				return true, nil
			}
		default:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				if !arg0Iterator.HasNext() {
					argList.Clear()
					return false, nil
				}
				element := arg0Iterator.Next()
				argList[0] = element
				argList[1] = index
				index++
				result, resultErr := callback.call(i, argList)
				if resultReturn, ok := result.(Return); ok {
					result = resultReturn.FinalValue
				} else if resultErr != nil {
					return false, resultErr
				}
				if !i.isTruthy(result) {
					argList.Clear()
					return false, nil
				}
				current = element
				return true, nil
			}
		}
		iterator.nextMethod = func() (any, error) {
			return current, nil
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("getwhilelast", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.getwhilelast' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.getwhilelast' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		var index int64 = 0
		var current any
		iterator := ProtoIteratorErr{legacyPanicOnErr: true}
		switch arg0Iterator := arg0Iterator.(type) {
		case interfaces.IteratorErr:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				ok, hasNextErr := arg0Iterator.HasNextErr()
				if hasNextErr != nil {
					return false, hasNextErr
				}
				if !ok {
					argList.Clear()
					return false, nil
				}
				element, nextErr := arg0Iterator.NextErr()
				if nextErr != nil {
					return false, nextErr
				}
				argList[0] = element
				argList[1] = index
				index++
				result, resultErr := callback.call(i, argList)
				if resultReturn, ok := result.(Return); ok {
					result = resultReturn.FinalValue
				} else if resultErr != nil {
					return false, resultErr
				}
				if !i.isTruthy(result) {
					argList.Clear()
				}
				current = element
				return true, nil
			}
		default:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				if !arg0Iterator.HasNext() {
					argList.Clear()
					return false, nil
				}
				element := arg0Iterator.Next()
				argList[0] = element
				argList[1] = index
				index++
				result, resultErr := callback.call(i, argList)
				if resultReturn, ok := result.(Return); ok {
					result = resultReturn.FinalValue
				} else if resultErr != nil {
					return false, resultErr
				}
				if !i.isTruthy(result) {
					argList.Clear()
				}
				current = element
				return true, nil
			}
		}
		iterator.nextMethod = func() (any, error) {
			return current, nil
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("infiniteArg", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		arg := args[0]
		iterator := InfiniteIterator{}
		iterator.nextMethod = func() any {
			return arg
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("infiniteArgs", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Expected at least 1 argument but got 0.")
		}
		index := 0
		iterator := InfiniteIterator{}
		iterator.nextMethod = func() any {
			element := args[index]
			if index >= argsLen-1 {
				index = 0
			} else {
				index++
			}
			return element
		}
		return NewLoxIterator(iterator), nil
	})
	iteratorFunc("length", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if iterable, ok := args[0].(interfaces.Iterable); ok {
			it := iterable.Iterator()
			itWithErr, isErrIter := it.(interfaces.IteratorErr)
			hasNext := func() (bool, error) {
				if isErrIter {
					return itWithErr.HasNextErr()
				}
				return it.HasNext(), nil
			}
			next := func() (any, error) {
				if isErrIter {
					return itWithErr.NextErr()
				}
				return it.Next(), nil
			}
			var count int64 = 0
			for {
				if count < 0 {
					//Integer overflow, return max 64-bit signed value
					return int64((1 << 63) - 1), nil
				}
				ok, hasNextErr := hasNext()
				if hasNextErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, hasNextErr.Error())
				}
				if !ok {
					return count, nil
				}
				_, nextErr := next()
				if nextErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, nextErr.Error())
				}
				count++
			}
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Iterator.length: type '%v' is not iterable.", getType(args[0])))
	})
	iteratorFunc("map", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.map' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.map' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		var index int64 = 0
		iterator := ProtoIteratorErr{legacyPanicOnErr: true}
		switch arg0Iterator := arg0Iterator.(type) {
		case interfaces.IteratorErr:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				ok, hasNextErr := arg0Iterator.HasNextErr()
				if hasNextErr != nil {
					return false, hasNextErr
				}
				if !ok {
					argList.Clear()
					return false, nil
				}
				return true, nil
			}
			iterator.nextMethod = func() (any, error) {
				element, nextErr := arg0Iterator.NextErr()
				if nextErr != nil {
					return nil, nextErr
				}
				argList[0] = element
				argList[1] = index
				index++
				result, resultErr := callback.call(i, argList)
				if resultReturn, ok := result.(Return); ok {
					return resultReturn.FinalValue, nil
				} else if resultErr != nil {
					return nil, resultErr
				} else {
					return result, nil
				}
			}
		default:
			iterator.hasNextMethod = func() (bool, error) {
				if argList == nil {
					return false, nil
				}
				if !arg0Iterator.HasNext() {
					argList.Clear()
					return false, nil
				}
				return true, nil
			}
			iterator.nextMethod = func() (any, error) {
				argList[0] = arg0Iterator.Next()
				argList[1] = index
				index++
				result, resultErr := callback.call(i, argList)
				if resultReturn, ok := result.(Return); ok {
					return resultReturn.FinalValue, nil
				} else if resultErr != nil {
					return nil, resultErr
				} else {
					return result, nil
				}
			}
		}
		return NewLoxIteratorTag(iterator, "map"), nil
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
			fmt.Sprintf("Iterator.pairwise: type '%v' is not iterable.", getType(args[0])))
	})
	iteratorFunc("reduce", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 2 && argsLen != 3 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.reduce' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.reduce' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		arg0IteratorWithErr, isErrIter := arg0Iterator.(interfaces.IteratorErr)
		callback := args[1].(*LoxFunction)
		hasNext := func() (bool, error) {
			if isErrIter {
				return arg0IteratorWithErr.HasNextErr()
			}
			return arg0Iterator.HasNext(), nil
		}
		next := func() (any, error) {
			if isErrIter {
				return arg0IteratorWithErr.NextErr()
			}
			return arg0Iterator.Next(), nil
		}
		var value any
		switch argsLen {
		case 2:
			ok, hasNextErr := hasNext()
			if hasNextErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, hasNextErr.Error())
			}
			if !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Cannot call 'Iterator.reduce' on empty iterable without initial value.")
			}
			var nextErr error
			value, nextErr = next()
			if nextErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, nextErr.Error())
			}
		case 3:
			value = args[2]
		}
		argList := getArgList(callback, 3)
		defer argList.Clear()
		for index := int64(0); ; index++ {
			if index == 0 && argsLen == 2 {
				continue
			}
			ok, hasNextErr := hasNext()
			if hasNextErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, hasNextErr.Error())
			}
			if !ok {
				return value, nil
			}
			argList[0] = value
			var nextErr error
			argList[1], nextErr = next()
			if nextErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, nextErr.Error())
			}
			argList[2] = index
			var valueErr error
			value, valueErr = callback.call(i, argList)
			if valueReturn, ok := value.(Return); ok {
				value = valueReturn.FinalValue
			} else if valueErr != nil {
				return nil, valueErr
			}
		}
	})
	iteratorFunc("reduceRight", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 2 && argsLen != 3 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(interfaces.Iterable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Iterator.reduceRight' is not iterable.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Iterator.reduceRight' must be a function.")
		}
		arg0Iterator := args[0].(interfaces.Iterable).Iterator()
		arg0IteratorWithErr, isErrIter := arg0Iterator.(interfaces.IteratorErr)
		callback := args[1].(*LoxFunction)
		hasNext := func() (bool, error) {
			if isErrIter {
				return arg0IteratorWithErr.HasNextErr()
			}
			return arg0Iterator.HasNext(), nil
		}
		next := func() (any, error) {
			if isErrIter {
				return arg0IteratorWithErr.NextErr()
			}
			return arg0Iterator.Next(), nil
		}
		iterElements := list.NewList[any]()
		defer iterElements.Clear()
		for {
			ok, hasNextErr := hasNext()
			if hasNextErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, hasNextErr.Error())
			}
			if !ok {
				if len(iterElements) == 0 {
					return nil, loxerror.RuntimeError(in.callToken,
						"Cannot call 'Iterator.reduceRight' on empty iterable without initial value.")
				}
				break
			}
			element, nextErr := next()
			if nextErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, nextErr.Error())
			}
			iterElements.Add(element)
		}
		lastIndex := int64(len(iterElements) - 1)
		var value any
		switch argsLen {
		case 2:
			value = iterElements[lastIndex]
		case 3:
			value = args[2]
		}
		argList := getArgList(callback, 3)
		defer argList.Clear()
		for index := lastIndex; index >= 0; index-- {
			if index == lastIndex && argsLen == 2 {
				continue
			}
			argList[0] = value
			argList[1] = iterElements[index]
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
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Iterator.reversed: type '%v' is not reverse-iterable.", getType(args[0])))
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
