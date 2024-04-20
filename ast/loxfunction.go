package ast

import (
	"github.com/AlanLuu/lox/env"
	"github.com/AlanLuu/lox/list"
)

type LoxFunction struct {
	declaration Function
}

func (f LoxFunction) arity() int {
	return len(f.declaration.Params)
}

func (f LoxFunction) call(interpreter *Interpreter, arguments list.List[any]) (any, error) {
	environment := env.NewEnvironmentEnclosing(interpreter.globals)
	for i := 0; i < len(f.declaration.Params); i++ {
		environment.Define(f.declaration.Params[i].Lexeme, arguments[i])
	}
	_, blockErr := interpreter.executeBlock(f.declaration.Body, environment)
	if blockErr != nil {
		return nil, blockErr
	}
	return nil, nil
}

func (f LoxFunction) String() string {
	return "<fn " + f.declaration.Name.Lexeme + ">"
}
