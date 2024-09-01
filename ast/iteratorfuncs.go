package ast

import (
	crand "crypto/rand"
	"fmt"
	"math/big"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
)

func defineIteratorFields(iteratorClass *LoxClass) {
	urandom := InfiniteLoxIterator{}
	urandom.nextMethod = func() any {
		numBig, numErr := crand.Int(crand.Reader, big.NewInt(256))
		if numErr != nil {
			panic(numErr)
		}
		return numBig.Int64()
	}
	iteratorClass.classProperties["urandom"] = NewLoxIterator(urandom)

	zeroes := InfiniteLoxIterator{}
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
	iteratorFunc("zip", -1, func(in *Interpreter, args list.List[any]) (any, error) {
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
		iterator := ProtoLoxIterator{}
		iterator.hasNextMethod = func() bool {
			if len(argIterators) == 0 {
				return false
			}
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
