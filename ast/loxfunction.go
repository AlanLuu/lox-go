package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/env"
	"github.com/AlanLuu/lox/list"
)

type LoxFunction struct {
	name          string
	declaration   FunctionExpr
	closure       *env.Environment
	isInitializer bool
	varArgPos     int
}

func (f *LoxFunction) arity() int {
	if f.hasVarArg() {
		return -1
	}
	return len(f.declaration.Params)
}

func (f *LoxFunction) bind(instance any) *LoxFunction {
	environment := env.NewEnvironmentEnclosing(f.closure)
	environment.Define("this", instance)
	return &LoxFunction{f.name, f.declaration, environment, f.isInitializer, f.varArgPos}
}

func (f *LoxFunction) call(interpreter *Interpreter, arguments list.List[any]) (any, error) {
	environment := env.NewEnvironmentEnclosing(f.closure)
	if f.hasVarArg() {
		for i := 0; i < len(f.declaration.Params); i++ {
			if i > f.varArgPos {
				environment.Define(f.declaration.Params[i].Lexeme, nil)
			} else if i == f.varArgPos {
				argsLen := len(arguments)
				varArgs := list.NewListCap[any](int64(argsLen - i))
				for j := i; j < argsLen; j++ {
					varArgs.Add(arguments[j])
				}
				environment.Define(f.declaration.Params[i].Lexeme, NewLoxList(varArgs))
			} else {
				environment.Define(f.declaration.Params[i].Lexeme, arguments[i])
			}
		}
	} else {
		for i := 0; i < len(f.declaration.Params); i++ {
			environment.Define(f.declaration.Params[i].Lexeme, arguments[i])
		}
	}
	retValue, blockErr := interpreter.executeBlock(f.declaration.Body, environment)
	if blockErr != nil {
		switch retValue := retValue.(type) {
		case Return:
			if f.isInitializer {
				return f.closure.GetAtStr(0, "this"), nil
			}
			return retValue, blockErr
		}
		return nil, blockErr
	}
	if f.isInitializer {
		return f.closure.GetAtStr(0, "this"), nil
	}
	return nil, nil
}

func (f *LoxFunction) hasVarArg() bool {
	return f.varArgPos >= 0
}

func (f *LoxFunction) String() string {
	if len(f.name) == 0 {
		return fmt.Sprintf("<fn at %p>", f)
	}
	return fmt.Sprintf("<fn %v at %p>", f.name, f)
}

func (f *LoxFunction) Type() string {
	return "function"
}
