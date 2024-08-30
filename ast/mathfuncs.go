package ast

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/util"
)

func (i *Interpreter) defineMathFuncs() {
	className := "Math"
	mathClass := NewLoxClass(className, nil, false)
	mathFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native Math fn %v at %p>", name, &s)
		}
		mathClass.classProperties[name] = s
	}
	zeroArgFunc := func(name string, fun func() float64) {
		mathFunc(name, 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return fun(), nil
		})
	}
	oneArgFunc := func(name string, fun func(float64) float64) {
		mathFunc(name, 1, func(in *Interpreter, args list.List[any]) (any, error) {
			switch num := args[0].(type) {
			case int64:
				return fun(float64(num)), nil
			case float64:
				return fun(num), nil
			}
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Argument to 'Math.%v' must be an integer or float.", name))
		})
	}
	twoArgFunc := func(name string, fun func(float64, float64) float64) {
		mathFunc(name, 2, func(in *Interpreter, args list.List[any]) (any, error) {
			switch num1 := args[0].(type) {
			case int64:
				switch num2 := args[1].(type) {
				case int64:
					return fun(float64(num1), float64(num2)), nil
				case float64:
					return fun(float64(num1), num2), nil
				}
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Second argument to 'Math.%v' must be an integer or float.", name))
			case float64:
				switch num2 := args[1].(type) {
				case int64:
					return fun(num1, float64(num2)), nil
				case float64:
					return fun(num1, num2), nil
				}
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Second argument to 'Math.%v' must be an integer or float.", name))
			}
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("First argument to 'Math.%v' must be an integer or float.", name))
		})
	}
	zeroArgFuncs := map[string]func() float64{
		"random": rand.Float64,
	}
	oneArgFuncs := map[string]func(float64) float64{
		"acos":  math.Acos,
		"acosh": math.Acosh,
		"asin":  math.Asin,
		"asinh": math.Asinh,
		"atan":  math.Atan,
		"atanh": math.Atanh,
		"cbrt":  math.Cbrt,
		"ceil":  math.Ceil,
		"cos":   math.Cos,
		"cosh":  math.Cosh,
		"exp":   math.Exp,
		"log":   math.Log,
		"log10": math.Log10,
		"log1p": math.Log1p,
		"log2":  math.Log2,
		"round": math.Round,
		"sin":   math.Sin,
		"sinh":  math.Sinh,
		"sqrt":  math.Sqrt,
		"tan":   math.Tan,
		"tanh":  math.Tanh,
	}
	twoArgFuncs := map[string]func(float64, float64) float64{
		"atan2": math.Atan2,
		"hypot": math.Hypot,
		"logB": func(num float64, base float64) float64 {
			return math.Log(num) / math.Log(base)
		},
		"nthrt": func(num float64, n float64) float64 {
			return math.Pow(num, 1/n)
		},
	}
	constants := map[string]float64{
		"E":  math.E,
		"PI": math.Pi,
	}
	for name, fun := range zeroArgFuncs {
		zeroArgFunc(name, fun)
	}
	for name, fun := range oneArgFuncs {
		oneArgFunc(name, fun)
	}
	for name, fun := range twoArgFuncs {
		twoArgFunc(name, fun)
	}
	for name, constant := range constants {
		mathClass.classProperties[name] = util.IntOrFloat(constant)
	}

	mathFunc("abs", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch num := args[0].(type) {
		case int64:
			if num < 0 {
				return -num, nil
			}
			return num, nil
		case float64:
			return math.Abs(num), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'Math.abs' must be an integer or float.")
	})
	mathFunc("floor", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch num := args[0].(type) {
		case int64:
			return num, nil
		case float64:
			return util.IntOrFloat(math.Floor(num)), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'Math.floor' must be an integer or float.")
	})
	mathFunc("max", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		fun := math.Max
		secondArgMsg := "Second argument to 'Math.max' must be an integer or float."
		switch num1 := args[0].(type) {
		case int64:
			switch num2 := args[1].(type) {
			case int64:
				if num1 > num2 {
					return num1, nil
				}
				return num2, nil
			case float64:
				return fun(float64(num1), num2), nil
			}
			return nil, loxerror.RuntimeError(in.callToken, secondArgMsg)
		case float64:
			switch num2 := args[1].(type) {
			case int64:
				return fun(num1, float64(num2)), nil
			case float64:
				return fun(num1, num2), nil
			}
			return nil, loxerror.RuntimeError(in.callToken, secondArgMsg)
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"First argument to 'Math.max' must be an integer or float.")
	})
	mathFunc("min", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		fun := math.Min
		secondArgMsg := "Second argument to 'Math.min' must be an integer or float."
		switch num1 := args[0].(type) {
		case int64:
			switch num2 := args[1].(type) {
			case int64:
				if num1 < num2 {
					return num1, nil
				}
				return num2, nil
			case float64:
				return fun(float64(num1), num2), nil
			}
			return nil, loxerror.RuntimeError(in.callToken, secondArgMsg)
		case float64:
			switch num2 := args[1].(type) {
			case int64:
				return fun(num1, float64(num2)), nil
			case float64:
				return fun(num1, num2), nil
			}
			return nil, loxerror.RuntimeError(in.callToken, secondArgMsg)
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"First argument to 'Math.min' must be an integer or float.")
	})
	mathFunc("trunc", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch num := args[0].(type) {
		case int64:
			return num, nil
		case float64:
			return util.IntOrFloat(math.Trunc(num)), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'Math.trunc' must be an integer or float.")
	})

	i.globals.Define(className, mathClass)
}
