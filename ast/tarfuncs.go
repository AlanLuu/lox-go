package ast

import (
	"bytes"
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

const (
	TAR_USE_BUFFER = 1 + iota
)

func defineTarFields(tarClass *LoxClass) {
	tarFields := map[string]int64{
		"USE_BUFFER": TAR_USE_BUFFER,
	}
	for key, value := range tarFields {
		tarClass.classProperties[key] = value
	}
}

func (i *Interpreter) defineTarFuncs() {
	className := "tar"
	tarClass := NewLoxClass(className, nil, false)
	tarFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native tar fn %v at %p>", name, &s)
		}
		tarClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'tar.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	defineTarFields(tarClass)
	tarFunc("writer", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *LoxFile:
			if !arg.isWrite() && !arg.isAppend() {
				return nil, loxerror.RuntimeError(in.callToken,
					"Cannot create tar writer for file not in write or append mode.")
			}
			return NewLoxTarWriter(arg.file), nil
		case int64:
			switch arg {
			case TAR_USE_BUFFER:
				return NewLoxTarWriterBytes(new(bytes.Buffer)), nil
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"Integer argument to 'tar.writer' must be equal to the field 'tar.USE_BUFFER'.")
			}
		default:
			return argMustBeType(in.callToken, "writer", "file or the field 'tar.USE_BUFFER'")
		}
	})

	i.globals.Define(className, tarClass)
}
