package ast

import (
	"encoding/hex"
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

	bigIntFunc("bitSize", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigInt, ok := args[0].(*big.Int); ok {
			return int64(bigInt.BitLen()), nil
		}
		return argMustBeType(in.callToken, "bitSize", "bigint")
	})
	bigIntFunc("bytes", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigInt, ok := args[0].(*big.Int); ok {
			bigIntBytes := bigInt.Bytes()
			buffer := EmptyLoxBufferCap(int64(len(bigIntBytes)))
			for _, b := range bigIntBytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
			return buffer, nil
		}
		return argMustBeType(in.callToken, "bytes", "bigint")
	})
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
	bigIntFunc("probablyPrime", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var bigInt *big.Int
		var n = 10
		argsLen := len(args)
		switch argsLen {
		case 1, 2:
			if arg, ok := args[0].(*big.Int); ok {
				bigInt = arg
			} else if argsLen == 2 {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'bigint.probablyPrime' must be a bigint.")
			} else {
				return argMustBeType(in.callToken, "probablyPrime", "bigint")
			}
			if argsLen == 2 {
				if arg, ok := args[1].(int64); ok {
					n = int(arg)
				} else {
					return nil, loxerror.RuntimeError(in.callToken,
						"Second argument to 'bigint.probablyPrime' must be an integer.")
				}
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
		if n < 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bigint.probablyPrime' cannot be negative.")
		}
		return bigInt.ProbablyPrime(n), nil
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
	bigIntFunc("toHexStr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigInt, ok := args[0].(*big.Int); ok {
			return NewLoxString(hex.EncodeToString(bigInt.Bytes()), '\''), nil
		}
		return argMustBeType(in.callToken, "toHexStr", "bigint")
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
