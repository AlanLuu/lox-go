package ast

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxLoggerIterator struct {
	loxLogger *LoxLogger
	index     int
}

func (l *LoxLoggerIterator) HasNext() bool {
	return l.index < len(l.loxLogger.savedLogs)
}

func (l *LoxLoggerIterator) Next() any {
	logStr := l.loxLogger.savedLogs[l.index]
	l.index++
	return NewLoxStringQuote(logStr)
}

var DefaultLoxLogger = NewLoxLogger(log.Default())

type LoxLogger struct {
	logger    *log.Logger
	isDefault bool
	savedLogs []string
	methods   map[string]*struct{ ProtoLoxCallable }
}

func NewLoxLogger(logger *log.Logger) *LoxLogger {
	return &LoxLogger{
		logger:    logger,
		isDefault: logger == log.Default(),
		savedLogs: []string{},
		methods:   make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxLoggerArgs(writer io.Writer, prefix string, flag int) *LoxLogger {
	return NewLoxLogger(log.New(writer, prefix, flag))
}

func (l *LoxLogger) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	loggerFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native logger fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'logger.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'logger.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	fatal := func() {
		CloseInputFuncReadline()
		os.Exit(1)
	}
	results := func(args []any) []any {
		elements := make([]any, 0, len(args))
		for _, arg := range args {
			elements = append(elements, getResult(arg, arg, true))
		}
		return elements
	}
	switch methodName {
	case "clearSavedLogs":
		return loggerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			(*list.List[string])(&l.savedLogs).Clear()
			return nil, nil
		})
	case "fatal":
		return loggerFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			l.logger.Println(results(args)...)
			fatal()
			return nil, nil
		})
	case "flags":
		return loggerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.logger.Flags()), nil
		})
	case "outputIs":
		return loggerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxFile, ok := args[0].(*LoxFile); ok {
				return l.logger.Writer() == loxFile.file, nil
			}
			return argMustBeType("file")
		})
	case "prefix":
		return loggerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.logger.Prefix()), nil
		})
	case "println":
		return loggerFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			l.logger.Println(results(args)...)
			return nil, nil
		})
	case "printSave":
		return loggerFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var builder strings.Builder
			prevWriter := l.logger.Writer()
			l.logger.SetOutput(&builder)
			l.logger.Println(results(args)...)
			l.logger.SetOutput(prevWriter)
			l.savedLogs = append(
				l.savedLogs,
				strings.TrimRight(builder.String(), "\n"),
			)
			return nil, nil
		})
	case "printSaved":
		return loggerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			writer := l.logger.Writer()
			for _, logStr := range l.savedLogs {
				fmt.Fprintln(writer, logStr)
			}
			return nil, nil
		})
	case "savedGet", "getSaved":
		return loggerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if index, ok := args[0].(int64); ok {
				if index < 0 || index >= int64(len(l.savedLogs)) {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"logger.%v: index %v out of range.",
							methodName,
							index,
						),
					)
				}
				return NewLoxStringQuote(l.savedLogs[index]), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "savedLen":
		return loggerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(len(l.savedLogs)), nil
		})
	case "savedLogs":
		return loggerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			logsList := list.NewListCap[any](int64(len(l.savedLogs)))
			for _, logStr := range l.savedLogs {
				logsList.Add(NewLoxStringQuote(logStr))
			}
			return NewLoxList(logsList), nil
		})
	case "setFlags":
		return loggerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if flag, ok := args[0].(int64); ok {
				l.logger.SetFlags(int(flag))
				return nil, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "setOutput":
		return loggerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxFile, ok := args[0].(*LoxFile); ok {
				l.logger.SetOutput(loxFile.file)
				return nil, nil
			}
			return argMustBeType("file")
		})
	case "setPrefix":
		return loggerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.logger.SetPrefix(loxStr.str)
				return nil, nil
			}
			return argMustBeType("string")
		})
	case "sprintln":
		return loggerFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var builder strings.Builder
			prevWriter := l.logger.Writer()
			l.logger.SetOutput(&builder)
			l.logger.Println(results(args)...)
			l.logger.SetOutput(prevWriter)
			return NewLoxStringQuote(strings.TrimRight(builder.String(), "\n")), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Logger objects have no property called '"+methodName+"'.")
}

func (l *LoxLogger) Iterator() interfaces.Iterator {
	return &LoxLoggerIterator{l, 0}
}

func (l *LoxLogger) String() string {
	if l.isDefault {
		return fmt.Sprintf("<default logger object at %p>", l)
	}
	return fmt.Sprintf("<logger object at %p>", l)
}

func (l *LoxLogger) Type() string {
	return "logger"
}
