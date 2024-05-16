package ast

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
)

func (i *Interpreter) defineNativeFuncs() {
	nativeFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native fn %v at %p>", name, &s)
		}
		i.globals.Define(name, s)
	}
	nativeFunc("clock", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return float64(time.Now().UnixMilli()) / 1000, nil
	})
	nativeFunc("chr", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if codePointNum, ok := args[0].(int64); ok {
			codePoint := rune(codePointNum)
			character := string(codePoint)
			if codePoint == '\'' {
				return &LoxString{character, '"'}, nil
			}
			return &LoxString{character, '\''}, nil
		}
		return nil, loxerror.Error("Argument to 'chr' must be a whole number.")
	})
	nativeFunc("len", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		switch element := args[0].(type) {
		case *LoxString:
			return int64(utf8.RuneCountInString(element.str)), nil
		case *LoxList:
			return int64(len(element.elements)), nil
		}
		return nil, loxerror.Error(fmt.Sprintf("Cannot get length of type '%v'.", getType(args[0])))
	})
	nativeFunc("List", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if size, ok := args[0].(int64); ok {
			if size < 0 {
				return nil, loxerror.Error("Argument to 'List' cannot be negative.")
			}
			lst := list.NewList[Expr]()
			for index := int64(0); index < size; index++ {
				lst.Add(nil)
			}
			return NewLoxList(lst), nil
		}
		return nil, loxerror.Error("Argument to 'List' must be a whole number.")
	})
	nativeFunc("ord", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			if utf8.RuneCountInString(loxStr.str) == 1 {
				codePoint, _ := utf8.DecodeRuneInString(loxStr.str)
				if codePoint == utf8.RuneError {
					return nil, loxerror.Error(fmt.Sprintf("Failed to decode character '%v'.", loxStr.str))
				}
				return int64(codePoint), nil
			}
		}
		return nil, loxerror.Error("Argument to 'ord' must be a single character.")
	})
	nativeFunc("sleep", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		switch seconds := args[0].(type) {
		case int64:
			time.Sleep(time.Duration(seconds) * time.Second)
			return nil, nil
		case float64:
			duration, _ := time.ParseDuration(fmt.Sprintf("%vs", seconds))
			time.Sleep(duration)
			return nil, nil
		}
		return nil, loxerror.Error("Argument to 'sleep' must be a number.")
	})
	nativeFunc("type", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		return getType(args[0]), nil
	})
}
