package ast

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxString struct {
	str   string
	quote byte
}

func EmptyLoxString() *LoxString {
	return &LoxString{
		str:   "",
		quote: '\'',
	}
}

func StringIndexMustBeWholeNum(index any) string {
	return fmt.Sprintf("String index '%v' must be a whole number.", index)
}

func StringIndexOutOfRange(index int64) string {
	return fmt.Sprintf("String index %v out of range.", index)
}

func (l *LoxString) NewLoxString(str string) *LoxString {
	return &LoxString{
		str:   str,
		quote: l.quote,
	}
}

func (l *LoxString) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxString:
		return l.str == obj.str
	default:
		return false
	}
}

func (l *LoxString) Get(name token.Token) (any, error) {
	strFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (struct{ ProtoLoxCallable }, error) {
		s := struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string { return "<native string fn>" }
		return s, nil
	}
	methodName := name.Lexeme
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'string.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "compare":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return int64(strings.Compare(l.str, loxStr.str)), nil
			}
			return argMustBeType("string")
		})
	case "contains":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return strings.Contains(l.str, loxStr.str), nil
			}
			return argMustBeType("string")
		})
	case "index":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return int64(strings.Index(l.str, loxStr.str)), nil
			}
			return argMustBeType("string")
		})
	case "lower":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return &LoxString{strings.ToLower(l.str), l.quote}, nil
		})
	case "split":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				splitSlice := strings.Split(l.str, loxStr.str)
				loxList := list.NewList[Expr]()
				for _, str := range splitSlice {
					loxList.Add(&LoxString{str, '\''})
				}
				return &LoxList{loxList}, nil
			}
			return argMustBeType("string")
		})
	case "strip":
		return strFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 0:
				return &LoxString{strings.TrimSpace(l.str), l.quote}, nil
			case 1:
				if loxStr, ok := args[0].(*LoxString); ok {
					return &LoxString{strings.Trim(l.str, loxStr.str), l.quote}, nil
				}
				return argMustBeType("string")
			}
			return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		})
	case "toNum":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			result, resultErr := strconv.ParseFloat(l.str, 64)
			if resultErr != nil {
				return math.NaN(), nil
			}
			resultAsInt := int64(result)
			if result == float64(resultAsInt) {
				return resultAsInt, nil
			}
			return result, nil
		})
	case "upper":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return &LoxString{strings.ToUpper(l.str), l.quote}, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Strings have no property called '"+methodName+"'.")
}

func (l *LoxString) String() string {
	return l.str
}
