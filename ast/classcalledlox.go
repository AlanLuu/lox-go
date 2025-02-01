package ast

import (
	"fmt"
	"runtime"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

func (i *Interpreter) defineClassCalledLox() {
	className := "lox"
	classCalledLox := NewLoxClass(className, nil, false)
	classCalledLoxFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native class called lox fn %v at %p>", name, &s)
		}
		classCalledLox.classProperties[name] = s
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'lox.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	classCalledLoxFunc("gc", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		switch argsLen {
		case 0:
			runtime.GC()
		case 1:
			if num, ok := args[0].(int64); ok {
				for i := int64(0); i < num; i++ {
					runtime.GC()
				}
			} else {
				return argMustBeTypeAn(in.callToken, "gc", "integer")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		return nil, nil
	})
	classCalledLoxFunc("globals", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		dict := EmptyLoxDict()
		for key, value := range in.globals.Values() {
			dict.setKeyValue(NewLoxString(key, '\''), value)
		}
		return dict, nil
	})
	classCalledLoxFunc("locals", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		dict := EmptyLoxDict()
		if in.environment != in.globals {
			for key, value := range in.environment.Values() {
				dict.setKeyValue(NewLoxString(key, '\''), value)
			}
		}
		return dict, nil
	})
	classCalledLox.classProperties["ranloxcode"] = !util.DisableLoxCode
	classCalledLox.classProperties["unsafe"] = util.UnsafeMode

	i.globals.Define(className, classCalledLox)
}
