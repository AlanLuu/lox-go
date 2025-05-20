package ast

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func defineLogFields(logClass *LoxClass) {
	flags := map[string]int64{
		"Ldate":         log.Ldate,
		"Ltime":         log.Ltime,
		"Lmicroseconds": log.Lmicroseconds,
		"Llongfile":     log.Llongfile,
		"Lshortfile":    log.Lshortfile,
		"LUTC":          log.LUTC,
		"Lmsgprefix":    log.Lmsgprefix,
		"LstdFlags":     log.LstdFlags,
	}
	for key, value := range flags {
		logClass.classProperties[key] = value
	}
}

func (i *Interpreter) defineLogFuncs() {
	className := "log"
	logClass := NewLoxClass(className, nil, false)
	logFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native log fn %v at %p>", name, &s)
		}
		logClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'log.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'log.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	results := func(args []any) []any {
		elements := make([]any, 0, len(args))
		for _, arg := range args {
			elements = append(elements, getResult(arg, arg, true))
		}
		return elements
	}

	defineLogFields(logClass)
	logFunc("fatal", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		log.Println(results(args)...)
		OSExit(1)
		return nil, nil
	})
	logFunc("flags", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(log.Flags()), nil
	})
	logFunc("logger", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		switch argsLen {
		case 0:
			return DefaultLoxLogger, nil
		case 2:
			twoArgsMsg := "When passing 2 arguments, "
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					twoArgsMsg+"first argument to 'log.logger' must be a string.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					twoArgsMsg+"second argument to 'log.logger' must be an integer.")
			}
			prefix := args[0].(*LoxString).str
			flag := int(args[1].(int64))
			return NewLoxLoggerArgs(os.Stderr, prefix, flag), nil
		case 3:
			threeArgsMsg := "When passing 3 arguments, "
			if _, ok := args[0].(*LoxFile); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					threeArgsMsg+"first argument to 'log.logger' must be a file.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					threeArgsMsg+"second argument to 'log.logger' must be a string.")
			}
			if _, ok := args[2].(int64); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					threeArgsMsg+"third argument to 'log.logger' must be an integer.")
			}
			file := args[0].(*LoxFile).file
			prefix := args[1].(*LoxString).str
			flag := int(args[2].(int64))
			return NewLoxLoggerArgs(file, prefix, flag), nil
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0, 2, or 3 arguments but got %v.", argsLen))
		}
	})
	logFunc("loggerFromDefault", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		defaultLogger := log.Default()
		return NewLoxLoggerArgs(
			defaultLogger.Writer(),
			defaultLogger.Prefix(),
			defaultLogger.Flags(),
		), nil
	})
	logFunc("outputIs", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxFile, ok := args[0].(*LoxFile); ok {
			return log.Writer() == loxFile.file, nil
		}
		return argMustBeType(in.callToken, "outputIs", "file")
	})
	logFunc("prefix", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxStringQuote(log.Prefix()), nil
	})
	logFunc("print", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		log.Print(results(args)...)
		return nil, nil
	})
	logFunc("println", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		log.Println(results(args)...)
		return nil, nil
	})
	logFunc("setFlags", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if flag, ok := args[0].(int64); ok {
			log.SetFlags(int(flag))
			return nil, nil
		}
		return argMustBeTypeAn(in.callToken, "setFlags", "integer")
	})
	logFunc("setOutput", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxFile, ok := args[0].(*LoxFile); ok {
			log.SetOutput(loxFile.file)
			return nil, nil
		}
		return argMustBeType(in.callToken, "setOutput", "file")
	})
	logFunc("setPrefix", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			log.SetPrefix(loxStr.str)
			return nil, nil
		}
		return argMustBeType(in.callToken, "setPrefix", "string")
	})
	logFunc("sprint", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		var builder strings.Builder
		prevWriter := log.Writer()
		log.SetOutput(&builder)
		log.Print(results(args)...)
		log.SetOutput(prevWriter)
		return NewLoxStringQuote(strings.TrimRight(builder.String(), "\n")), nil
	})
	logFunc("sprintln", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		var builder strings.Builder
		prevWriter := log.Writer()
		log.SetOutput(&builder)
		log.Println(results(args)...)
		log.SetOutput(prevWriter)
		return NewLoxStringQuote(builder.String()), nil
	})

	i.globals.Define(className, logClass)
}
