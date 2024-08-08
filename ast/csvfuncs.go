package ast

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineCSVFuncs() {
	className := "csv"
	csvClass := NewLoxClass(className, nil, false)
	csvFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native csv fn %v at %p>", name, &s)
		}
		csvClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'csv.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	csvFunc("reader", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 1 && argsLen != 2 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v", argsLen))
		}
		delimiter := ','
		if argsLen == 2 {
			switch args[0].(type) {
			case *LoxFile:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'csv.reader' must be a file or string.")
			}
			if loxStr, ok := args[1].(*LoxString); ok {
				if utf8.RuneCountInString(loxStr.str) != 1 {
					return nil, loxerror.RuntimeError(in.callToken,
						"Second argument to 'csv.reader' must be a single-character string.")
				}
				delimiter = []rune(loxStr.str)[0]
			} else {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'csv.reader' must be a string.")
			}
		}
		switch arg := args[0].(type) {
		case *LoxFile:
			if !arg.isRead() {
				return nil, loxerror.RuntimeError(in.callToken,
					"Cannot create CSV reader for file not in read mode.")
			}
			if arg.isBinary {
				return nil, loxerror.RuntimeError(in.callToken,
					"Cannot create CSV reader for file in binary read mode.")
			}
			return NewLoxCSVReaderDelimiter(arg.file, delimiter), nil
		case *LoxString:
			return NewLoxCSVReaderDelimiter(strings.NewReader(arg.str), delimiter), nil
		}
		return argMustBeType(in.callToken, "reader", "file or string")
	})

	i.globals.Define(className, csvClass)
}
