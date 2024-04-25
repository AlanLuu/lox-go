package ast

import (
	"github.com/AlanLuu/lox/env"
	"github.com/AlanLuu/lox/list"
)

type LoxFunction struct {
	name        string
	declaration FunctionExpr
	closure     *env.Environment
}

func (f LoxFunction) arity() int {
	return len(f.declaration.Params)
}

func (f LoxFunction) bind(instance *LoxInstance) LoxFunction {
	environment := env.NewEnvironmentEnclosing(f.closure)
	environment.Define("this", instance)
	return LoxFunction{f.name, f.declaration, environment}
}

func (f LoxFunction) call(interpreter *Interpreter, arguments list.List[any]) (any, error) {
	environment := env.NewEnvironmentEnclosing(f.closure)
	for i := 0; i < len(f.declaration.Params); i++ {
		environment.Define(f.declaration.Params[i].Lexeme, arguments[i])
	}
	retValue, blockErr := interpreter.executeBlock(f.declaration.Body, environment)
	if blockErr != nil {
		switch retValue := retValue.(type) {
		case Return:
			return retValue, blockErr
		}
		return nil, blockErr
	}
	return nil, nil
}

func (f LoxFunction) String() string {
	if len(f.name) == 0 {
		return "<fn>"
	}
	return "<fn " + f.name + ">"
}
