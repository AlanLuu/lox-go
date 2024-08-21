package ast

import (
	"fmt"
	"math/big"

	"github.com/AlanLuu/lox/bignum/bigfloat"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineBigFloatFuncs() {
	className := "bigfloat"
	bigFloatClass := NewLoxClass(className, nil, false)
	bigFloatFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native bigfloat fn %v at %p>", name, &s)
		}
		bigFloatClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bigfloat.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bigfloat.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	bigFloatFunc("new", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case int64:
			return bigfloat.New(float64(arg)), nil
		case float64:
			return bigfloat.New(arg), nil
		case *LoxString:
			bigFloat := &big.Float{}
			_, ok := bigFloat.SetString(arg.str)
			if !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Failed to convert '%v' to bigfloat.", arg.str))
			}
			return bigFloat, nil
		default:
			return argMustBeTypeAn(in.callToken, "new", "integer, float, or string")
		}
	})
	bigFloatFunc("toBigInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigFloat, ok := args[0].(*big.Float); ok {
			bigInt := &big.Int{}
			bigFloat.Int(bigInt)
			return bigInt, nil
		}
		return argMustBeType(in.callToken, "toBigInt", "bigfloat")
	})
	bigFloatFunc("toFloat", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigFloat, ok := args[0].(*big.Float); ok {
			float, _ := bigFloat.Float64()
			return float, nil
		}
		return argMustBeType(in.callToken, "toFloat", "bigfloat")
	})
	bigFloatFunc("toInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigFloat, ok := args[0].(*big.Float); ok {
			float, _ := bigFloat.Float64()
			return int64(float), nil
		}
		return argMustBeType(in.callToken, "toInt", "bigfloat")
	})
	bigFloatFunc("toString", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigFloat, ok := args[0].(*big.Float); ok {
			return NewLoxString(bigFloat.String(), '\''), nil
		}
		return argMustBeType(in.callToken, "toString", "bigfloat")
	})

	i.globals.Define(className, bigFloatClass)
}
