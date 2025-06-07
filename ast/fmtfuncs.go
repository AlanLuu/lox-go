package ast

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineFmtFuncs() {
	className := "fmt"
	fmtClass := NewLoxClass(className, nil, false)
	fmtFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native fmt class fn %v at %p>", name, &s)
		}
		fmtClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'fmt.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'fmt.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	argsToStrings := func(args []any) []any {
		ret := make([]any, 0, len(args))
		for _, arg := range args {
			ret = append(ret, getResult(arg, arg, true))
		}
		return ret
	}
	argsToStringsAsStrSlice := func(args []any) []string {
		ret := make([]string, 0, len(args))
		for _, arg := range args {
			ret = append(ret, getResult(arg, arg, true))
		}
		return ret
	}
	scanfResultsLen := func(formatStr string) int {
		percentCount := strings.Count(formatStr, "%")
		doublePercentCount := strings.Count(formatStr, "%%")
		percentSpaceCount := strings.Count(formatStr, "% ")
		return percentCount - (doublePercentCount * 2) - percentSpaceCount
	}
	parseFormatStr := func(formatStr string, useBytes bool) []any {
		results := []any{}
		formatRunes := []rune(formatStr)
		formatRunesLen := len(formatRunes)
	outer:
		for i := 0; i < formatRunesLen; i++ {
			if formatRunes[i] == '%' {
				if i++; i >= formatRunesLen {
					break outer
				}
				switch formatRunes[i] {
				case '+', '-', '#', ' ':
					if i++; i >= formatRunesLen {
						break outer
					}
					for {
						switch formatRunes[i] {
						case '+', '-', '#', ' ':
							if i++; i >= formatRunesLen {
								break outer
							}
							continue
						}
						break
					}
				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					seenDot := false
					for {
						if i++; i >= formatRunesLen {
							break outer
						}
						if formatRunes[i] == '.' {
							if seenDot {
								break
							}
							seenDot = true
							if i++; i >= formatRunesLen {
								break outer
							}
						}
						if !(formatRunes[i] >= '0' && formatRunes[i] <= '9') {
							break
						}
					}
				case '.':
					for {
						if i++; i >= formatRunesLen {
							break outer
						}
						if !(formatRunes[i] >= '0' && formatRunes[i] <= '9') {
							break
						}
					}
				}
				switch formatRunes[i] {
				case 'b', 'c', 'd', 'o', 'O', 'x', 'X', 'U':
					var ii int64
					results = append(results, &ii)
				case 'e', 'E', 'f', 'F', 'g', 'G':
					var f float64
					results = append(results, &f)
				case 't':
					var b bool
					results = append(results, &b)
				case '%':
				default:
					if useBytes {
						var b []byte
						results = append(results, &b)
					} else {
						var s string
						results = append(results, &s)
					}
				}
			}
		}
		return results
	}
	resultsToLoxList := func(results []any) (*LoxList, error) {
		resultsList := list.NewListCap[any](int64(len(results)))
		for _, result := range results {
			switch result := result.(type) {
			case *int64:
				resultsList.Add(*result)
			case *float64:
				resultsList.Add(*result)
			case *bool:
				resultsList.Add(*result)
			case *string:
				resultsList.Add(NewLoxStringQuote(*result))
			case *[]byte:
				byteSlice := *result
				buffer := EmptyLoxBufferCap(int64(len(byteSlice)))
				for _, b := range byteSlice {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, addErr
					}
				}
				resultsList.Add(buffer)
			}
		}
		return NewLoxList(resultsList), nil
	}

	fmtFunc("append", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.append: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxBuffer); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.append' must be a buffer.")
		}
		loxBuffer := args[0].(*LoxBuffer)
		bytes := make([]byte, 0, len(loxBuffer.elements))
		for _, element := range loxBuffer.elements {
			bytes = append(bytes, byte(element.(int64)))
		}
		result := fmt.Append(bytes, argsToStrings(args[1:])...)
		if len(result) > len(loxBuffer.elements) {
			for i := len(loxBuffer.elements); i < len(result); i++ {
				addErr := loxBuffer.add(int64(result[i]))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
		}
		return nil, nil
	})
	fmtFunc("appendRet", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.appendRet: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxBuffer); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.appendRet' must be a buffer.")
		}
		loxBuffer := args[0].(*LoxBuffer)
		bytes := make([]byte, 0, len(loxBuffer.elements))
		for _, element := range loxBuffer.elements {
			bytes = append(bytes, byte(element.(int64)))
		}
		result := fmt.Append(bytes, argsToStrings(args[1:])...)
		retBuffer := EmptyLoxBufferCap(int64(len(result)))
		for _, value := range result {
			addErr := retBuffer.add(int64(value))
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
		}
		return retBuffer, nil
	})
	fmtFunc("appendf", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"fmt.appendf: expected at least 2 arguments but got %v.",
					len(args),
				),
			)
		}
		if _, ok := args[0].(*LoxBuffer); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.appendf' must be a buffer.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'fmt.appendf' must be a string.")
		}
		loxBuffer := args[0].(*LoxBuffer)
		bytes := make([]byte, 0, len(loxBuffer.elements))
		for _, element := range loxBuffer.elements {
			bytes = append(bytes, byte(element.(int64)))
		}
		formatStr := args[1].(*LoxString).str
		result := fmt.Appendf(bytes, formatStr, args[2:]...)
		if len(result) > len(loxBuffer.elements) {
			for i := len(loxBuffer.elements); i < len(result); i++ {
				addErr := loxBuffer.add(int64(result[i]))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
		}
		return nil, nil
	})
	fmtFunc("appendfRet", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"fmt.appendfRet: expected at least 2 arguments but got %v.",
					len(args),
				),
			)
		}
		if _, ok := args[0].(*LoxBuffer); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.appendfRet' must be a buffer.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'fmt.appendfRet' must be a string.")
		}
		loxBuffer := args[0].(*LoxBuffer)
		bytes := make([]byte, 0, len(loxBuffer.elements))
		for _, element := range loxBuffer.elements {
			bytes = append(bytes, byte(element.(int64)))
		}
		formatStr := args[1].(*LoxString).str
		result := fmt.Appendf(bytes, formatStr, args[2:]...)
		retBuffer := EmptyLoxBufferCap(int64(len(result)))
		for _, value := range result {
			addErr := retBuffer.add(int64(value))
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
		}
		return retBuffer, nil
	})
	fmtFunc("appendfln", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"fmt.appendfln: expected at least 2 arguments but got %v.",
					len(args),
				),
			)
		}
		if _, ok := args[0].(*LoxBuffer); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.appendfln' must be a buffer.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'fmt.appendfln' must be a string.")
		}
		loxBuffer := args[0].(*LoxBuffer)
		bytes := make([]byte, 0, len(loxBuffer.elements))
		for _, element := range loxBuffer.elements {
			bytes = append(bytes, byte(element.(int64)))
		}
		formatStr := args[1].(*LoxString).str
		result := fmt.Appendf(bytes, formatStr, args[2:]...)
		result = append(result, '\n')
		if len(result) > len(loxBuffer.elements) {
			for i := len(loxBuffer.elements); i < len(result); i++ {
				addErr := loxBuffer.add(int64(result[i]))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
		}
		return nil, nil
	})
	fmtFunc("appendflnRet", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"fmt.appendflnRet: expected at least 2 arguments but got %v.",
					len(args),
				),
			)
		}
		if _, ok := args[0].(*LoxBuffer); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.appendflnRet' must be a buffer.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'fmt.appendflnRet' must be a string.")
		}
		loxBuffer := args[0].(*LoxBuffer)
		bytes := make([]byte, 0, len(loxBuffer.elements))
		for _, element := range loxBuffer.elements {
			bytes = append(bytes, byte(element.(int64)))
		}
		formatStr := args[1].(*LoxString).str
		result := fmt.Appendf(bytes, formatStr, args[2:]...)
		result = append(result, '\n')
		retBuffer := EmptyLoxBufferCap(int64(len(result)))
		for _, value := range result {
			addErr := retBuffer.add(int64(value))
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
		}
		return retBuffer, nil
	})
	fmtFunc("appendln", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.appendln: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxBuffer); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.appendln' must be a buffer.")
		}
		loxBuffer := args[0].(*LoxBuffer)
		bytes := make([]byte, 0, len(loxBuffer.elements))
		for _, element := range loxBuffer.elements {
			bytes = append(bytes, byte(element.(int64)))
		}
		result := fmt.Appendln(bytes, argsToStrings(args[1:])...)
		if len(result) > len(loxBuffer.elements) {
			for i := len(loxBuffer.elements); i < len(result); i++ {
				addErr := loxBuffer.add(int64(result[i]))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
		}
		return nil, nil
	})
	fmtFunc("appendlnRet", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.appendlnRet: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxBuffer); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.appendlnRet' must be a buffer.")
		}
		loxBuffer := args[0].(*LoxBuffer)
		bytes := make([]byte, 0, len(loxBuffer.elements))
		for _, element := range loxBuffer.elements {
			bytes = append(bytes, byte(element.(int64)))
		}
		result := fmt.Appendln(bytes, argsToStrings(args[1:])...)
		retBuffer := EmptyLoxBufferCap(int64(len(result)))
		for _, value := range result {
			addErr := retBuffer.add(int64(value))
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
		}
		return retBuffer, nil
	})
	fmtFunc("error", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argStrs := argsToStringsAsStrSlice(args)
		return NewLoxError(errors.New(strings.Join(argStrs, " "))), nil
	})
	fmtFunc("errorf", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.errorf: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.errorf' must be a string.")
		}
		return NewLoxError(
			fmt.Errorf(
				args[0].(*LoxString).str,
				args[1:]...,
			),
		), nil
	})
	fmtFunc("errorfln", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.errorfln: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.errorfln' must be a string.")
		}
		return NewLoxError(
			fmt.Errorf(
				args[0].(*LoxString).str+"\n",
				args[1:]...,
			),
		), nil
	})
	fmtFunc("errorln", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argStrs := argsToStringsAsStrSlice(args)
		return NewLoxError(errors.New(strings.Join(argStrs, " ") + "\n")), nil
	})
	fmtFunc("fprint", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.fprint: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.fprint' must be a file.")
		}
		fileArg := args[0].(*LoxFile)
		if !fileArg.isWrite() && !fileArg.isAppend() {
			return nil, loxerror.RuntimeError(in.callToken,
				"First file argument to 'fmt.fprint' must be in write or append mode.")
		}
		fmt.Fprint(fileArg.file, argsToStrings(args[1:])...)
		return nil, nil
	})
	fmtFunc("fprintCheckErr", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.fprintCheckErr: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.fprintCheckErr' must be a file.")
		}
		fileArg := args[0].(*LoxFile)
		if !fileArg.isWrite() && !fileArg.isAppend() {
			return nil, loxerror.RuntimeError(in.callToken,
				"First file argument to 'fmt.fprintCheckErr' must be in write or append mode.")
		}
		_, err := fmt.Fprint(fileArg.file, argsToStrings(args[1:])...)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	fmtFunc("fprintf", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"fmt.fprintf: expected at least 2 arguments but got %v.",
					len(args),
				),
			)
		}
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.fprintf' must be a file.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'fmt.fprintf' must be a string.")
		}
		fileArg := args[0].(*LoxFile)
		if !fileArg.isWrite() && !fileArg.isAppend() {
			return nil, loxerror.RuntimeError(in.callToken,
				"First file argument to 'fmt.fprintf' must be in write or append mode.")
		}
		formatStr := args[1].(*LoxString).str
		fmt.Fprintf(fileArg.file, formatStr, args[2:]...)
		return nil, nil
	})
	fmtFunc("fprintfCheckErr", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"fmt.fprintfCheckErr: expected at least 2 arguments but got %v.",
					len(args),
				),
			)
		}
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.fprintfCheckErr' must be a file.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'fmt.fprintfCheckErr' must be a string.")
		}
		fileArg := args[0].(*LoxFile)
		if !fileArg.isWrite() && !fileArg.isAppend() {
			return nil, loxerror.RuntimeError(in.callToken,
				"First file argument to 'fmt.fprintfCheckErr' must be in write or append mode.")
		}
		formatStr := args[1].(*LoxString).str
		_, err := fmt.Fprintf(fileArg.file, formatStr, args[2:]...)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	fmtFunc("fprintfln", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"fmt.fprintfln: expected at least 2 arguments but got %v.",
					len(args),
				),
			)
		}
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.fprintfln' must be a file.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'fmt.fprintfln' must be a string.")
		}
		fileArg := args[0].(*LoxFile)
		if !fileArg.isWrite() && !fileArg.isAppend() {
			return nil, loxerror.RuntimeError(in.callToken,
				"First file argument to 'fmt.fprintfln' must be in write or append mode.")
		}
		formatStr := args[1].(*LoxString).str
		fmt.Fprintf(fileArg.file, formatStr+"\n", args[2:]...)
		return nil, nil
	})
	fmtFunc("fprintflnCheckErr", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) < 2 {
			return nil, loxerror.RuntimeError(
				in.callToken,
				fmt.Sprintf(
					"fmt.fprintflnCheckErr: expected at least 2 arguments but got %v.",
					len(args),
				),
			)
		}
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.fprintflnCheckErr' must be a file.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'fmt.fprintflnCheckErr' must be a string.")
		}
		fileArg := args[0].(*LoxFile)
		if !fileArg.isWrite() && !fileArg.isAppend() {
			return nil, loxerror.RuntimeError(in.callToken,
				"First file argument to 'fmt.fprintflnCheckErr' must be in write or append mode.")
		}
		formatStr := args[1].(*LoxString).str
		_, err := fmt.Fprintf(fileArg.file, formatStr+"\n", args[2:]...)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	fmtFunc("fprintln", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.fprintln: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.fprintln' must be a file.")
		}
		fileArg := args[0].(*LoxFile)
		if !fileArg.isWrite() && !fileArg.isAppend() {
			return nil, loxerror.RuntimeError(in.callToken,
				"First file argument to 'fmt.fprintln' must be in write or append mode.")
		}
		fmt.Fprintln(fileArg.file, argsToStrings(args[1:])...)
		return nil, nil
	})
	fmtFunc("fprintlnCheckErr", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.fprintlnCheckErr: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.fprintlnCheckErr' must be a file.")
		}
		fileArg := args[0].(*LoxFile)
		if !fileArg.isWrite() && !fileArg.isAppend() {
			return nil, loxerror.RuntimeError(in.callToken,
				"First file argument to 'fmt.fprintlnCheckErr' must be in write or append mode.")
		}
		_, err := fmt.Fprintln(fileArg.file, argsToStrings(args[1:])...)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	fmtFunc("print", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		fmt.Print(argsToStrings(args)...)
		return nil, nil
	})
	fmtFunc("printCheckErr", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		_, err := fmt.Print(argsToStrings(args)...)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	fmtFunc("printf", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.printf: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.printf' must be a string.")
		}
		fmt.Printf(args[0].(*LoxString).str, args[1:]...)
		return nil, nil
	})
	fmtFunc("printfCheckErr", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.printfCheckErr: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.printfCheckErr' must be a string.")
		}
		_, err := fmt.Printf(args[0].(*LoxString).str, args[1:]...)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	fmtFunc("printfln", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.printfln: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.printfln' must be a string.")
		}
		fmt.Printf(args[0].(*LoxString).str+"\n", args[1:]...)
		return nil, nil
	})
	fmtFunc("printflnCheckErr", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.printflnCheckErr: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.printflnCheckErr' must be a string.")
		}
		_, err := fmt.Printf(args[0].(*LoxString).str+"\n", args[1:]...)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	fmtFunc("println", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		fmt.Println(argsToStrings(args)...)
		return nil, nil
	})
	fmtFunc("printlnCheckErr", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		_, err := fmt.Println(argsToStrings(args)...)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	fmtFunc("scanBool", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'fmt.scanBool' cannot be 0 or negative.")
			}
			results := make([]any, arg)
			for i := int64(0); i < arg; i++ {
				var b bool
				results[i] = &b
			}
			_, err := fmt.Scan(results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*bool))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeTypeAn(in.callToken, "scanBool", "integer")
	})
	fmtFunc("scanBuf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'fmt.scanBuf' cannot be 0 or negative.")
			}
			results := make([]any, arg)
			for i := int64(0); i < arg; i++ {
				var b []byte
				results[i] = &b
			}
			_, err := fmt.Scan(results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				byteSlice := *result.(*[]byte)
				buffer := EmptyLoxBufferCap(int64(len(byteSlice)))
				for _, b := range byteSlice {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
					}
				}
				resultsList.Add(buffer)
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeTypeAn(in.callToken, "scanBuf", "integer")
	})
	fmtFunc("scanFloat", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'fmt.scanFloat' cannot be 0 or negative.")
			}
			results := make([]any, arg)
			for i := int64(0); i < arg; i++ {
				var f float64
				results[i] = &f
			}
			_, err := fmt.Scan(results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*float64))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeTypeAn(in.callToken, "scanFloat", "integer")
	})
	fmtFunc("scanInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'fmt.scanInt' cannot be 0 or negative.")
			}
			results := make([]any, arg)
			for i := int64(0); i < arg; i++ {
				var ii int64
				results[i] = &ii
			}
			_, err := fmt.Scan(results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*int64))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeTypeAn(in.callToken, "scanInt", "integer")
	})
	fmtFunc("scanStr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'fmt.scanStr' cannot be 0 or negative.")
			}
			results := make([]any, arg)
			for i := int64(0); i < arg; i++ {
				var s string
				results[i] = &s
			}
			_, err := fmt.Scan(results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(NewLoxStringQuote(*result.(*string)))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeTypeAn(in.callToken, "scanStr", "integer")
	})
	fmtFunc("scanf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			results := parseFormatStr(formatStr, false)
			_, err := fmt.Scanf(formatStr, results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			loxList, addErr := resultsToLoxList(results)
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
			return loxList, nil
		}
		return argMustBeType(in.callToken, "scanf", "string")
	})
	fmtFunc("scanfBool", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			resultsLen := scanfResultsLen(formatStr)
			if resultsLen <= 0 {
				return EmptyLoxList(), nil
			}
			results := make([]any, resultsLen)
			for i := 0; i < resultsLen; i++ {
				var b bool
				results[i] = &b
			}
			_, err := fmt.Scanf(formatStr, results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*bool))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "scanfBool", "string")
	})
	fmtFunc("scanfBuf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			resultsLen := scanfResultsLen(formatStr)
			if resultsLen <= 0 {
				return EmptyLoxList(), nil
			}
			results := make([]any, resultsLen)
			for i := 0; i < resultsLen; i++ {
				var b []byte
				results[i] = &b
			}
			_, err := fmt.Scanf(formatStr, results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				byteSlice := *result.(*[]byte)
				buffer := EmptyLoxBufferCap(int64(len(byteSlice)))
				for _, b := range byteSlice {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
					}
				}
				resultsList.Add(buffer)
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "scanfBuf", "string")
	})
	fmtFunc("scanfBufVerbs", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			results := parseFormatStr(formatStr, true)
			_, err := fmt.Scanf(formatStr, results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			loxList, addErr := resultsToLoxList(results)
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
			return loxList, nil
		}
		return argMustBeType(in.callToken, "scanfBufVerbs", "string")
	})
	fmtFunc("scanfFloat", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			resultsLen := scanfResultsLen(formatStr)
			if resultsLen <= 0 {
				return EmptyLoxList(), nil
			}
			results := make([]any, resultsLen)
			for i := 0; i < resultsLen; i++ {
				var f float64
				results[i] = &f
			}
			_, err := fmt.Scanf(formatStr, results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*float64))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "scanfFloat", "string")
	})
	fmtFunc("scanfInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			resultsLen := scanfResultsLen(formatStr)
			if resultsLen <= 0 {
				return EmptyLoxList(), nil
			}
			results := make([]any, resultsLen)
			for i := 0; i < resultsLen; i++ {
				var ii int64
				results[i] = &ii
			}
			_, err := fmt.Scanf(formatStr, results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*int64))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "scanfInt", "string")
	})
	fmtFunc("scanfStr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			resultsLen := scanfResultsLen(formatStr)
			if resultsLen <= 0 {
				return EmptyLoxList(), nil
			}
			results := make([]any, resultsLen)
			for i := 0; i < resultsLen; i++ {
				var s string
				results[i] = &s
			}
			_, err := fmt.Scanf(formatStr, results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(NewLoxStringQuote(*result.(*string)))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "scanfStr", "string")
	})
	fmtFunc("scanfln", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			results := parseFormatStr(formatStr, false)
			_, err := fmt.Scanf(formatStr+"\n", results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			loxList, addErr := resultsToLoxList(results)
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
			return loxList, nil
		}
		return argMustBeType(in.callToken, "scanfln", "string")
	})
	fmtFunc("scanflnBool", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			resultsLen := scanfResultsLen(formatStr)
			if resultsLen <= 0 {
				return EmptyLoxList(), nil
			}
			results := make([]any, resultsLen)
			for i := 0; i < resultsLen; i++ {
				var b bool
				results[i] = &b
			}
			_, err := fmt.Scanf(formatStr+"\n", results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*bool))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "scanflnBool", "string")
	})
	fmtFunc("scanflnBuf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			resultsLen := scanfResultsLen(formatStr)
			if resultsLen <= 0 {
				return EmptyLoxList(), nil
			}
			results := make([]any, resultsLen)
			for i := 0; i < resultsLen; i++ {
				var b []byte
				results[i] = &b
			}
			_, err := fmt.Scanf(formatStr+"\n", results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				byteSlice := *result.(*[]byte)
				buffer := EmptyLoxBufferCap(int64(len(byteSlice)))
				for _, b := range byteSlice {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
					}
				}
				resultsList.Add(buffer)
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "scanflnBuf", "string")
	})
	fmtFunc("scanflnBufVerbs", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			results := parseFormatStr(formatStr, true)
			_, err := fmt.Scanf(formatStr+"\n", results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			loxList, addErr := resultsToLoxList(results)
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
			return loxList, nil
		}
		return argMustBeType(in.callToken, "scanflnBufVerbs", "string")
	})
	fmtFunc("scanflnFloat", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			resultsLen := scanfResultsLen(formatStr)
			if resultsLen <= 0 {
				return EmptyLoxList(), nil
			}
			results := make([]any, resultsLen)
			for i := 0; i < resultsLen; i++ {
				var f float64
				results[i] = &f
			}
			_, err := fmt.Scanf(formatStr+"\n", results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*float64))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "scanflnFloat", "string")
	})
	fmtFunc("scanflnInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			resultsLen := scanfResultsLen(formatStr)
			if resultsLen <= 0 {
				return EmptyLoxList(), nil
			}
			results := make([]any, resultsLen)
			for i := 0; i < resultsLen; i++ {
				var ii int64
				results[i] = &ii
			}
			_, err := fmt.Scanf(formatStr+"\n", results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*int64))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "scanflnInt", "string")
	})
	fmtFunc("scanflnStr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(*LoxString); ok {
			formatStr := arg.str
			resultsLen := scanfResultsLen(formatStr)
			if resultsLen <= 0 {
				return EmptyLoxList(), nil
			}
			results := make([]any, resultsLen)
			for i := 0; i < resultsLen; i++ {
				var s string
				results[i] = &s
			}
			_, err := fmt.Scanf(formatStr+"\n", results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(NewLoxStringQuote(*result.(*string)))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "scanflnStr", "string")
	})
	fmtFunc("scanlnBool", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'fmt.scanlnBool' cannot be 0 or negative.")
			}
			results := make([]any, arg)
			for i := int64(0); i < arg; i++ {
				var b bool
				results[i] = &b
			}
			_, err := fmt.Scanln(results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*bool))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeTypeAn(in.callToken, "scanlnBool", "integer")
	})
	fmtFunc("scanlnBuf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'fmt.scanlnBuf' cannot be 0 or negative.")
			}
			results := make([]any, arg)
			for i := int64(0); i < arg; i++ {
				var b []byte
				results[i] = &b
			}
			_, err := fmt.Scanln(results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				byteSlice := *result.(*[]byte)
				buffer := EmptyLoxBufferCap(int64(len(byteSlice)))
				for _, b := range byteSlice {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
					}
				}
				resultsList.Add(buffer)
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeTypeAn(in.callToken, "scanlnBuf", "integer")
	})
	fmtFunc("scanlnFloat", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'fmt.scanlnFloat' cannot be 0 or negative.")
			}
			results := make([]any, arg)
			for i := int64(0); i < arg; i++ {
				var f float64
				results[i] = &f
			}
			_, err := fmt.Scanln(results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*float64))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeTypeAn(in.callToken, "scanlnFloat", "integer")
	})
	fmtFunc("scanlnInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'fmt.scanlnInt' cannot be 0 or negative.")
			}
			results := make([]any, arg)
			for i := int64(0); i < arg; i++ {
				var ii int64
				results[i] = &ii
			}
			_, err := fmt.Scanln(results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(*result.(*int64))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeTypeAn(in.callToken, "scanlnInt", "integer")
	})
	fmtFunc("scanlnStr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'fmt.scanlnStr' cannot be 0 or negative.")
			}
			results := make([]any, arg)
			for i := int64(0); i < arg; i++ {
				var s string
				results[i] = &s
			}
			_, err := fmt.Scanln(results...)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(NewLoxStringQuote(*result.(*string)))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeTypeAn(in.callToken, "scanlnStr", "integer")
	})
	fmtFunc("sprint", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		return NewLoxStringQuote(fmt.Sprint(argsToStrings(args)...)), nil
	})
	fmtFunc("sprintBuf", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		resultStr := fmt.Sprint(argsToStrings(args)...)
		resultStrLen := len(resultStr)
		buffer := EmptyLoxBufferCap(int64(resultStrLen))
		for i := 0; i < resultStrLen; i++ {
			addErr := buffer.add(int64(resultStr[i]))
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
		}
		return buffer, nil
	})
	fmtFunc("sprintf", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.sprintf: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.sprintf' must be a string.")
		}
		return NewLoxStringQuote(
			fmt.Sprintf(
				args[0].(*LoxString).str,
				args[1:]...,
			),
		), nil
	})
	fmtFunc("sprintfBuf", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.sprintfBuf: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.sprintfBuf' must be a string.")
		}
		resultStr := fmt.Sprintf(
			args[0].(*LoxString).str,
			args[1:]...,
		)
		resultStrLen := len(resultStr)
		buffer := EmptyLoxBufferCap(int64(resultStrLen))
		for i := 0; i < resultStrLen; i++ {
			addErr := buffer.add(int64(resultStr[i]))
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
		}
		return buffer, nil
	})
	fmtFunc("sprintfln", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.sprintfln: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.sprintfln' must be a string.")
		}
		return NewLoxStringQuote(
			fmt.Sprintf(
				args[0].(*LoxString).str+"\n",
				args[1:]...,
			),
		), nil
	})
	fmtFunc("sprintflnBuf", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"fmt.sprintflnBuf: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'fmt.sprintflnBuf' must be a string.")
		}
		resultStr := fmt.Sprintf(
			args[0].(*LoxString).str+"\n",
			args[1:]...,
		)
		resultStrLen := len(resultStr)
		buffer := EmptyLoxBufferCap(int64(resultStrLen))
		for i := 0; i < resultStrLen; i++ {
			addErr := buffer.add(int64(resultStr[i]))
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
		}
		return buffer, nil
	})
	fmtFunc("sprintln", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		return NewLoxStringQuote(fmt.Sprintln(argsToStrings(args)...)), nil
	})
	fmtFunc("sprintlnBuf", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		resultStr := fmt.Sprintln(argsToStrings(args)...)
		resultStrLen := len(resultStr)
		buffer := EmptyLoxBufferCap(int64(resultStrLen))
		for i := 0; i < resultStrLen; i++ {
			addErr := buffer.add(int64(resultStr[i]))
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
		}
		return buffer, nil
	})

	i.globals.Define(className, fmtClass)
}
