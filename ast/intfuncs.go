package ast

import (
	"fmt"
	"math"
	"strconv"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineIntFuncs() {
	className := "Integer"
	intClass := NewLoxClass(className, nil, false)
	intFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native Integer fn %v at %p>", name, &s)
		}
		intClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'Integer.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'Integer.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	intClass.classProperties["MAX"] = int64(math.MaxInt64)
	intClass.classProperties["MIN"] = int64(math.MinInt64)
	intFunc("parseInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			result, resultErr := strconv.ParseInt(loxStr.str, 0, 64)
			if resultErr != nil {
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Failed to convert '%v' to integer.", loxStr.str))
			}
			return result, nil
		}
		return argMustBeType(in.callToken, "parseInt", "string")
	})
	intFunc("toFloat", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if value, ok := args[0].(int64); ok {
			return float64(value), nil
		}
		return argMustBeTypeAn(in.callToken, "toFloat", "integer")
	})
	intFunc("toString", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if value, ok := args[0].(int64); ok {
			return NewLoxString(fmt.Sprint(value), '\''), nil
		}
		return argMustBeTypeAn(in.callToken, "toString", "integer")
	})

	i.globals.Define(className, intClass)
}
