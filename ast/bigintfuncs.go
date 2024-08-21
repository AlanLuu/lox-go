package ast

import (
	"fmt"
	"math/big"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineBigIntFuncs() {
	className := "bigint"
	bigIntClass := NewLoxClass(className, nil, false)
	bigIntFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native bigint fn %v at %p>", name, &s)
		}
		bigIntClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bigint.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bigint.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	bigIntFunc("new", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case int64:
			return new(big.Int).SetInt64(arg), nil
		case float64:
			return new(big.Int).SetInt64(int64(arg)), nil
		case *LoxString:
			bigInt := &big.Int{}
			_, ok := bigInt.SetString(arg.str, 0)
			if !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Failed to convert '%v' to bigint.", arg.str))
			}
			return bigInt, nil
		default:
			return argMustBeTypeAn(in.callToken, "new", "integer, float, or string")
		}
	})
	bigIntFunc("isInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigInt, ok := args[0].(*big.Int); ok {
			return bigInt.IsInt64(), nil
		}
		return argMustBeType(in.callToken, "isInt", "bigint")
	})
	bigIntFunc("toBigFloat", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigInt, ok := args[0].(*big.Int); ok {
			return new(big.Float).SetInt(bigInt), nil
		}
		return argMustBeType(in.callToken, "toBigFloat", "bigint")
	})
	bigIntFunc("toFloat", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigInt, ok := args[0].(*big.Int); ok {
			return float64(bigInt.Int64()), nil
		}
		return argMustBeType(in.callToken, "toFloat", "bigint")
	})
	bigIntFunc("toInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigInt, ok := args[0].(*big.Int); ok {
			return bigInt.Int64(), nil
		}
		return argMustBeType(in.callToken, "toInt", "bigint")
	})
	bigIntFunc("toString", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigInt, ok := args[0].(*big.Int); ok {
			return NewLoxString(bigInt.String(), '\''), nil
		}
		return argMustBeType(in.callToken, "toString", "bigint")
	})

	i.globals.Define(className, bigIntClass)
}
