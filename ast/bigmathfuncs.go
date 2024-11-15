package ast

import (
	"fmt"
	"math/big"
	"math/rand"

	"github.com/AlanLuu/lox/bignum/bigfloat"
	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineBigMathFuncs() {
	className := "bigmath"
	bigMathClass := NewLoxClass(className, nil, false)
	bigMathFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native bigmath fn %v at %p>", name, &s)
		}
		bigMathClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bigmath.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	bigMathFunc("abs", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *big.Int:
			return new(big.Int).Abs(arg), nil
		case *big.Float:
			return new(big.Float).Abs(arg), nil
		default:
			return argMustBeType(in.callToken, "abs", "bigint or bigfloat")
		}
	})
	bigMathFunc("ceil", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *big.Int:
			return new(big.Int).Set(arg), nil
		case *big.Float:
			intPart, frac := new(big.Int), new(big.Float)
			arg.Int(intPart)
			frac.Sub(arg, new(big.Float).SetInt(intPart))
			if bigfloat.IsPositive(frac) {
				intPart.Add(intPart, bigint.One)
			}
			return intPart, nil
		default:
			return argMustBeType(in.callToken, "ceil", "bigint or bigfloat")
		}
	})
	bigMathFunc("dim", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		switch first := args[0].(type) {
		case *big.Int:
			switch second := args[1].(type) {
			case *big.Int:
				result := new(big.Int).Sub(first, second)
				if bigint.IsNegative(result) {
					return big.NewInt(0), nil
				}
				return result, nil
			case *big.Float:
				result := new(big.Float).Sub(new(big.Float).SetInt(first), second)
				if bigfloat.IsNegative(result) {
					return bigfloat.New(0.0), nil
				}
				return result, nil
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'bigmath.dim' must be a bigint or bigfloat.")
			}
		case *big.Float:
			switch second := args[1].(type) {
			case *big.Int:
				result := new(big.Float).Sub(first, new(big.Float).SetInt(second))
				if bigfloat.IsNegative(result) {
					return bigfloat.New(0.0), nil
				}
				return result, nil
			case *big.Float:
				result := new(big.Float).Sub(first, second)
				if bigfloat.IsNegative(result) {
					return bigfloat.New(0.0), nil
				}
				return result, nil
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'bigmath.dim' must be a bigint or bigfloat.")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bigmath.dim' must be a bigint or bigfloat.")
		}
	})
	bigMathFunc("divmod", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*big.Int); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bigmath.divmod' must be a bigint.")
		}
		if _, ok := args[1].(*big.Int); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bigmath.divmod' must be a bigint.")
		}
		first := args[0].(*big.Int)
		second := args[1].(*big.Int)
		if bigint.IsZero(second) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Cannot divide bigint by 0.")
		}
		quotient, modulus := new(big.Int).DivMod(first, second, new(big.Int))
		resultList := list.NewListCap[any](2)
		resultList.Add(quotient)
		resultList.Add(modulus)
		return NewLoxList(resultList), nil
	})
	bigMathFunc("factorial", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigInt, ok := args[0].(*big.Int); ok {
			if bigint.IsNegative(bigInt) {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'bigmath.factorial' cannot be negative.")
			}
			if bigint.IsZero(bigInt) || bigInt.Cmp(bigint.One) == 0 {
				return big.NewInt(1), nil
			}
			result := new(big.Int).Set(bigInt)
			for i := new(big.Int).Sub(bigInt, bigint.One); i.Cmp(bigint.One) > 0; i.Sub(i, bigint.One) {
				result.Mul(result, i)
			}
			return result, nil
		}
		return argMustBeType(in.callToken, "factorial", "bigint")
	})
	bigMathFunc("floor", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *big.Int:
			return new(big.Int).Set(arg), nil
		case *big.Float:
			result, accuracy := arg.Int(nil)
			if bigint.IsNegative(result) && accuracy != big.Exact {
				result.Sub(result, bigint.One)
			}
			return result, nil
		default:
			return argMustBeType(in.callToken, "floor", "bigint or bigfloat")
		}
	})
	bigMathFunc("gcd", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*big.Int); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bigmath.gcd' must be a bigint.")
		}
		if _, ok := args[1].(*big.Int); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bigmath.gcd' must be a bigint.")
		}
		first := args[0].(*big.Int)
		second := args[1].(*big.Int)
		return new(big.Int).GCD(nil, nil, first, second), nil
	})
	bigMathFunc("lcm", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*big.Int); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bigmath.lcm' must be a bigint.")
		}
		if _, ok := args[1].(*big.Int); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bigmath.lcm' must be a bigint.")
		}
		first := args[0].(*big.Int)
		second := args[1].(*big.Int)
		if bigint.IsZero(first) && bigint.IsZero(second) {
			return big.NewInt(0), nil
		}
		result := new(big.Int)
		result.Mul(first, second)
		if bigint.IsNegative(result) {
			result.Neg(result)
		}
		result.Div(result, new(big.Int).GCD(nil, nil, first, second))
		return result, nil
	})
	bigMathFunc("mantexp", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigFloat, ok := args[0].(*big.Float); ok {
			mant := new(big.Float)
			exp := bigFloat.MantExp(mant)
			resultList := list.NewListCap[any](2)
			resultList.Add(mant)
			resultList.Add(int64(exp))
			return NewLoxList(resultList), nil
		}
		return argMustBeType(in.callToken, "mantexp", "bigfloat")
	})
	bigMathFunc("max", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		var first *big.Float
		var second *big.Float
		firstIsInt, secondIsInt := false, false
		switch arg := args[0].(type) {
		case *big.Int:
			firstIsInt = true
			first = new(big.Float).SetInt(arg)
		case *big.Float:
			first = new(big.Float).Set(arg)
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bigmath.max' must be a bigint or bigfloat.")
		}
		switch arg := args[1].(type) {
		case *big.Int:
			secondIsInt = true
			second = new(big.Float).SetInt(arg)
		case *big.Float:
			second = new(big.Float).Set(arg)
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bigmath.max' must be a bigint or bigfloat.")
		}
		if first.Cmp(second) > 0 {
			if firstIsInt {
				return new(big.Int).Set(args[0].(*big.Int)), nil
			}
			return first, nil
		} else {
			if secondIsInt {
				return new(big.Int).Set(args[1].(*big.Int)), nil
			}
			return second, nil
		}
	})
	bigMathFunc("min", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		var first *big.Float
		var second *big.Float
		firstIsInt, secondIsInt := false, false
		switch arg := args[0].(type) {
		case *big.Int:
			firstIsInt = true
			first = new(big.Float).SetInt(arg)
		case *big.Float:
			first = new(big.Float).Set(arg)
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bigmath.min' must be a bigint or bigfloat.")
		}
		switch arg := args[1].(type) {
		case *big.Int:
			secondIsInt = true
			second = new(big.Float).SetInt(arg)
		case *big.Float:
			second = new(big.Float).Set(arg)
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bigmath.min' must be a bigint or bigfloat.")
		}
		if first.Cmp(second) < 0 {
			if firstIsInt {
				return new(big.Int).Set(args[0].(*big.Int)), nil
			}
			return first, nil
		} else {
			if secondIsInt {
				return new(big.Int).Set(args[1].(*big.Int)), nil
			}
			return second, nil
		}
	})
	bigMathFunc("mod", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*big.Int); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bigmath.mod' must be a bigint.")
		}
		if _, ok := args[1].(*big.Int); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bigmath.mod' must be a bigint.")
		}
		first := args[0].(*big.Int)
		second := args[1].(*big.Int)
		if bigint.IsZero(second) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Cannot divide bigint by 0.")
		}
		return new(big.Int).Mod(first, second), nil
	})
	bigMathFunc("quorem", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*big.Int); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bigmath.quorem' must be a bigint.")
		}
		if _, ok := args[1].(*big.Int); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bigmath.quorem' must be a bigint.")
		}
		first := args[0].(*big.Int)
		second := args[1].(*big.Int)
		if bigint.IsZero(second) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Cannot divide bigint by 0.")
		}
		quotient, remainder := new(big.Int).QuoRem(first, second, new(big.Int))
		resultList := list.NewListCap[any](2)
		resultList.Add(quotient)
		resultList.Add(remainder)
		return NewLoxList(resultList), nil
	})
	bigMathFunc("random", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return bigfloat.New(rand.Float64()), nil
	})
	bigMathFunc("round", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *big.Int:
			return new(big.Int).Set(arg), nil
		case *big.Float:
			intPart, frac := new(big.Int), new(big.Float)
			arg.Int(intPart)
			frac.Sub(arg, new(big.Float).SetInt(intPart))
			half := bigfloat.New(0.5)
			if arg.Sign() >= 0 {
				if frac.Cmp(half) >= 0 {
					intPart.Add(intPart, bigint.One)
				}
			} else if frac.Cmp(new(big.Float).Neg(half)) <= 0 {
				intPart.Sub(intPart, bigint.One)
			}
			return intPart, nil
		default:
			return argMustBeType(in.callToken, "round", "bigint or bigfloat")
		}
	})
	bigMathFunc("sqrt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var arg *big.Float
		switch arg1 := args[0].(type) {
		case *big.Int:
			arg = new(big.Float).SetInt(arg1)
		case *big.Float:
			arg = new(big.Float).Set(arg1)
		default:
			return argMustBeType(in.callToken, "sqrt", "bigint or bigfloat")
		}
		if bigfloat.IsNegative(arg) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Argument to 'bigmath.sqrt' cannot be negative.")
		}
		return arg.Sqrt(arg), nil
	})
	bigMathFunc("sqrtint", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bigInt, ok := args[0].(*big.Int); ok {
			if bigint.IsNegative(bigInt) {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'bigmath.sqrtint' cannot be negative.")
			}
			return new(big.Int).Sqrt(bigInt), nil
		}
		return argMustBeType(in.callToken, "sqrtint", "bigint")
	})

	i.globals.Define(className, bigMathClass)
}
