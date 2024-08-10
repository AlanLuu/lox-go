package ast

import (
	"fmt"
	"regexp"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineRegexFuncs() {
	className := "regex"
	regexClass := NewLoxClass(className, nil, false)
	regexFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native regex class fn %v at %p>", name, &s)
		}
		regexClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'regex class.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	regexFunc("compile", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			loxRegex, err := NewLoxRegexStr(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return loxRegex, nil
		}
		return argMustBeType(in.callToken, "compile", "string")
	})
	regexFunc("escape", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(regexp.QuoteMeta(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "escape", "string")
	})
	regexFunc("test", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'regex class.test' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'regex class.test' must be a string.")
		}
		pattern := args[0].(*LoxString).str
		s := args[1].(*LoxString).str
		matched, err := regexp.MatchString(pattern, s)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return matched, nil
	})

	i.globals.Define(className, regexClass)
}
