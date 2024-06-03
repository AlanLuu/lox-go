package ast

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/scanner"
	"github.com/chzyer/readline"
)

var textSc *bufio.Scanner

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
				return NewLoxString(character, '"'), nil
			}
			return NewLoxString(character, '\''), nil
		}
		return nil, loxerror.Error("Argument to 'chr' must be an integer.")
	})
	nativeFunc("eval", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if codeStr, ok := args[0].(*LoxString); ok {
			importSc := scanner.NewScanner(codeStr.str)
			scanErr := importSc.ScanTokens()
			if scanErr != nil {
				return nil, scanErr
			}

			importParser := NewParser(importSc.Tokens)
			exprList, parseErr := importParser.Parse()
			defer exprList.Clear()
			if parseErr != nil {
				return nil, parseErr
			}

			importResolver := NewResolver(i)
			resolverErr := importResolver.Resolve(exprList)
			if resolverErr != nil {
				return nil, resolverErr
			}

			evalValue, valueErr := i.InterpretReturnLast(exprList)
			if valueErr != nil {
				return nil, valueErr
			}
			return evalValue, nil
		}
		return args[0], nil
	})
	nativeFunc("input", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		var prompt any = ""
		argsLen := len(args)
		switch argsLen {
		case 0:
		case 1:
			prompt = args[0]
		default:
			return nil, loxerror.Error(fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}

		var userInput string
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			l, _ := readline.NewEx(&readline.Config{
				Prompt:          getResult(prompt, prompt, true),
				InterruptPrompt: "^C",
			})
			defer l.Close()

			var readError error
			userInput, readError = l.Readline()
			switch readError {
			case readline.ErrInterrupt, io.EOF:
				return nil, nil
			}
		} else {
			if textSc == nil {
				textSc = bufio.NewScanner(os.Stdin)
			}
			if !textSc.Scan() {
				return nil, nil
			}
			userInput = textSc.Text()
		}

		if strings.Contains(userInput, "'") {
			return NewLoxString(userInput, '"'), nil
		}
		return NewLoxString(userInput, '\''), nil
	})
	nativeFunc("len", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Length); ok {
			return element.Length(), nil
		}
		return nil, loxerror.Error(fmt.Sprintf("Cannot get length of type '%v'.", getType(args[0])))
	})
	nativeFunc("List", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if size, ok := args[0].(int64); ok {
			if size < 0 {
				return nil, loxerror.Error("Argument to 'List' cannot be negative.")
			}
			lst := list.NewList[any]()
			for index := int64(0); index < size; index++ {
				lst.Add(nil)
			}
			return NewLoxList(lst), nil
		}
		return nil, loxerror.Error("Argument to 'List' must be an integer.")
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
	nativeFunc("Set", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		set := EmptyLoxSet()
		for _, element := range args {
			_, errStr := set.add(element)
			if len(errStr) > 0 {
				return nil, loxerror.Error(errStr)
			}
		}
		return set, nil
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
		return nil, loxerror.Error("Argument to 'sleep' must be an integer or float.")
	})
	nativeFunc("type", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		return NewLoxString(getType(args[0]), '\''), nil
	})
}
