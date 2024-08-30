package ast

import (
	"fmt"
	"math"
	"strconv"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

func (i *Interpreter) defineFloatFuncs() {
	className := "Float"
	floatClass := NewLoxClass(className, nil, false)
	floatFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native Float fn %v at %p>", name, &s)
		}
		floatClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'Float.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	floatClass.classProperties["MAX"] = float64(math.MaxFloat64)
	floatClass.classProperties["MIN"] = float64(math.SmallestNonzeroFloat64)
	floatFunc("parseFloat", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			result, resultErr := strconv.ParseFloat(loxStr.str, 64)
			if resultErr != nil {
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Failed to convert '%v' to float.", loxStr.str))
			}
			return result, nil
		}
		return argMustBeType(in.callToken, "parseFloat", "string")
	})
	floatFunc("toInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if value, ok := args[0].(float64); ok {
			return int64(value), nil
		}
		return argMustBeType(in.callToken, "toInt", "float")
	})
	floatFunc("toString", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if value, ok := args[0].(float64); ok {
			return NewLoxString(util.FormatFloatZero(value), '\''), nil
		}
		return argMustBeType(in.callToken, "toString", "float")
	})

	i.globals.Define(className, floatClass)
}
