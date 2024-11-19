package ast

import (
	"bytes"
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

const (
	ZIP_USE_BUFFER = 1 + iota
)

func defineZipFields(zipClass *LoxClass) {
	zipFields := map[string]int64{
		"USE_BUFFER": ZIP_USE_BUFFER,
	}
	for key, value := range zipFields {
		zipClass.classProperties[key] = value
	}
}

func (i *Interpreter) defineZipFuncs() {
	className := "zip"
	zipClass := NewLoxClass(className, nil, false)
	zipFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native zip fn %v at %p>", name, &s)
		}
		zipClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'zip.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	defineZipFields(zipClass)
	zipFunc("writer", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *LoxFile:
			if !arg.isWrite() && !arg.isAppend() {
				return nil, loxerror.RuntimeError(in.callToken,
					"Cannot create zip writer for file not in write or append mode.")
			}
			return NewLoxZIPWriter(arg.file), nil
		case int64:
			switch arg {
			case ZIP_USE_BUFFER:
				return NewLoxZIPWriterBytes(new(bytes.Buffer)), nil
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"Integer argument to 'zip.writer' must be equal to the field 'zip.USE_BUFFER'.")
			}
		default:
			return argMustBeType(in.callToken, "writer", "file or the field 'zip.USE_BUFFER'")
		}
	})

	i.globals.Define(className, zipClass)
}
