package ast

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/scanner"
	"github.com/AlanLuu/lox/util"
	"github.com/chzyer/readline"
	"github.com/mattn/go-isatty"
)

var inputSc *bufio.Scanner
var inputReadline *readline.Instance

func CloseInputFuncReadline() {
	if inputReadline != nil {
		inputReadline.Close()
	}
}

func OSExit(code int) {
	CloseInputFuncReadline()
	os.Exit(code)
}

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
	numToBaseStr := func(num int64, prefix string, base int) (*LoxString, error) {
		var builder strings.Builder
		if num < 0 {
			builder.WriteRune('-')
			builder.WriteString(prefix)
			builder.WriteString(strconv.FormatInt(num, base)[1:])
		} else {
			builder.WriteString(prefix)
			builder.WriteString(strconv.FormatInt(num, base))
		}
		return NewLoxString(builder.String(), '\''), nil
	}
	nativeFunc("arity", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if callable, ok := args[0].(LoxCallable); ok {
			return int64(callable.arity()), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'arity' must be a function or class.")
	})
	nativeFunc("bigrange", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		switch argsLen {
		case 1:
			switch stop := args[0].(type) {
			case int64:
				return NewLoxBigRangeStop(big.NewInt(stop)), nil
			case *big.Int:
				return NewLoxBigRangeStop(stop), nil
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'bigrange' must be an integer or bigint.")
			}
		case 2, 3:
			var start, stop, step *big.Int
			switch arg1 := args[0].(type) {
			case int64:
				start = big.NewInt(arg1)
			case *big.Int:
				start = new(big.Int).Set(arg1)
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'bigrange' must be an integer or bigint.")
			}
			switch arg2 := args[1].(type) {
			case int64:
				stop = big.NewInt(arg2)
			case *big.Int:
				stop = new(big.Int).Set(arg2)
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'bigrange' must be an integer or bigint.")
			}
			if argsLen == 3 {
				switch arg3 := args[2].(type) {
				case int64:
					if arg3 == 0 {
						return nil, loxerror.RuntimeError(in.callToken,
							"Third argument to 'bigrange' cannot be 0.")
					}
					step = big.NewInt(arg3)
				case *big.Int:
					if arg3.Cmp(bigint.Zero) == 0 {
						return nil, loxerror.RuntimeError(in.callToken,
							"Third argument to 'bigrange' cannot be 0n.")
					}
					step = new(big.Int).Set(arg3)
				default:
					return nil, loxerror.RuntimeError(in.callToken,
						"Third argument to 'bigrange' must be an integer or bigint.")
				}
			} else {
				step = big.NewInt(1)
			}
			return NewLoxBigRange(start, stop, step), nil
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1, 2, or 3 arguments but got %v.", argsLen))
		}
	})
	nativeFunc("bin", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			return numToBaseStr(num, "0b", 2)
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'bin' must be an integer.")
	})
	nativeFunc("Buffer", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		buffer := EmptyLoxBufferCap(int64(len(args)))
		for _, element := range args {
			addErr := buffer.add(element)
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
		}
		return buffer, nil
	})
	nativeFunc("BufferCap", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if capacity, ok := args[0].(int64); ok {
			if capacity < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'BufferCap' cannot be negative.")
			}
			return EmptyLoxBufferCap(capacity), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'BufferCap' must be an integer.")
	})
	nativeFunc("BufferZero", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if size, ok := args[0].(int64); ok {
			if size < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'BufferZero' cannot be negative.")
			}
			buffer := EmptyLoxBufferCap(size)
			for index := int64(0); index < size; index++ {
				buffer.add(int64(0))
			}
			return buffer, nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'BufferZero' must be an integer.")
	})
	nativeFunc("cap", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Capacity); ok {
			return element.Capacity(), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Cannot get capacity of type '%v'.", getType(args[0])))
	})
	nativeFunc("chr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if codePointNum, ok := args[0].(int64); ok {
			codePoint := rune(codePointNum)
			character := string(codePoint)
			if codePoint == '\'' {
				return NewLoxString(character, '"'), nil
			}
			return NewLoxString(character, '\''), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'chr' must be an integer.")
	})
	nativeFunc("clock", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return float64(time.Now().UnixMilli()) / 1000, nil
	})
	nativeFunc("Deque", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		deque := NewLoxDeque()
		for _, element := range args {
			deque.pushBack(element)
		}
		return deque, nil
	})
	nativeFunc("DequeIterable", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Iterable); ok {
			deque := NewLoxDeque()
			it := element.Iterator()
			for it.HasNext() {
				deque.pushBack(it.Next())
			}
			return deque, nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("DequeIterable: type '%v' is not iterable.", getType(args[0])))
	})
	nativeFunc("DictIterable", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Iterable); ok {
			dict := EmptyLoxDict()
			it := element.Iterator()
			switch element.(type) {
			case *LoxDict:
				for it.HasNext() {
					pair := it.Next().(*LoxList).elements
					dict.setKeyValue(pair[0], pair[1])
				}
			default:
				for index := int64(0); it.HasNext(); index++ {
					dict.setKeyValue(index, it.Next())
				}
			}
			return dict, nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("DictIterable: type '%v' is not iterable.", getType(args[0])))
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

			previous := i.environment
			defer func() {
				i.environment = previous
			}()
			i.environment = i.globals

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
	nativeFunc("hex", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			return numToBaseStr(num, "0x", 16)
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'hex' must be an integer.")
	})
	nativeFunc("input", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var prompt any = ""
		argsLen := len(args)
		switch argsLen {
		case 0:
		case 1:
			prompt = args[0]
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}

		var userInput string
		fd := os.Stdin.Fd()
		if isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd) {
			if inputReadline == nil {
				inputReadline, _ = readline.NewEx(&readline.Config{
					Prompt:          getResult(prompt, prompt, true),
					InterruptPrompt: "^C",
				})
			} else {
				inputReadline.SetPrompt(getResult(prompt, prompt, true))
			}

			var readError error
			userInput, readError = inputReadline.Readline()
			switch readError {
			case readline.ErrInterrupt:
				return nil, loxerror.RuntimeError(in.callToken, "Keyboard interrupt")
			case io.EOF:
				return nil, nil
			}
		} else {
			if inputSc == nil {
				inputSc = bufio.NewScanner(os.Stdin)
			}
			if !inputSc.Scan() {
				return nil, nil
			}
			userInput = inputSc.Text()
		}

		if strings.Contains(userInput, "'") {
			return NewLoxString(userInput, '"'), nil
		}
		return NewLoxString(userInput, '\''), nil
	})
	nativeFunc("isinstance", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[1].(*LoxClass); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'isinstance' must be a class.")
		}
		switch instance := args[0].(type) {
		case *LoxInstance:
			class := args[1].(*LoxClass)
			for c := instance.class; c != nil; c = c.superClass {
				if c == class {
					return true, nil
				}
			}
		}
		return false, nil
	})
	nativeFunc("iterator", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Iterable); ok {
			return NewLoxIterator(element.Iterator()), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("iterator: type '%v' is not iterable.", getType(args[0])))
	})
	nativeFunc("len", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Length); ok {
			return element.Length(), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("Cannot get length of type '%v'.", getType(args[0])))
	})
	nativeFunc("List", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if size, ok := args[0].(int64); ok {
			if size < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'List' cannot be negative.")
			}
			lst := list.NewListCap[any](size)
			for index := int64(0); index < size; index++ {
				lst.Add(nil)
			}
			return NewLoxList(lst), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'List' must be an integer.")
	})
	nativeFunc("ListCap", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if capacity, ok := args[0].(int64); ok {
			if capacity < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'ListCap' cannot be negative.")
			}
			return NewLoxList(list.NewListCap[any](capacity)), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'ListCap' must be an integer.")
	})
	nativeFunc("ListIterable", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Iterable); ok {
			lst := list.NewList[any]()
			it := element.Iterator()
			for it.HasNext() {
				lst.Add(it.Next())
			}
			return NewLoxList(lst), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("ListIterable: type '%v' is not iterable.", getType(args[0])))
	})
	nativeFunc("ListZero", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if size, ok := args[0].(int64); ok {
			if size < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'ListZero' cannot be negative.")
			}
			lst := list.NewListCap[any](size)
			for index := int64(0); index < size; index++ {
				lst.Add(int64(0))
			}
			return NewLoxList(lst), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'ListZero' must be an integer.")
	})
	nativeFunc("oct", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			return numToBaseStr(num, "0o", 8)
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'oct' must be an integer.")
	})
	nativeFunc("ord", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			if utf8.RuneCountInString(loxStr.str) == 1 {
				codePoint, _ := utf8.DecodeRuneInString(loxStr.str)
				if codePoint == utf8.RuneError {
					return nil, loxerror.RuntimeError(in.callToken,
						fmt.Sprintf("Failed to decode character '%v'.", loxStr.str))
				}
				return int64(codePoint), nil
			}
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'ord' must be a single character.")
	})
	nativeFunc("Queue", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		queue := NewLoxQueue()
		for _, element := range args {
			queue.add(element)
		}
		return queue, nil
	})
	nativeFunc("QueueIterable", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Iterable); ok {
			queue := NewLoxQueue()
			it := element.Iterator()
			for it.HasNext() {
				queue.add(it.Next())
			}
			return queue, nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("QueueIterable: type '%v' is not iterable.", getType(args[0])))
	})
	nativeFunc("range", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		switch argsLen {
		case 1:
			if stop, ok := args[0].(int64); ok {
				return NewLoxRangeStop(stop), nil
			}
			return nil, loxerror.RuntimeError(in.callToken,
				"Argument to 'range' must be an integer.")
		case 2, 3:
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'range' must be an integer.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'range' must be an integer.")
			}
			start := args[0].(int64)
			stop := args[1].(int64)
			var step int64
			if argsLen == 3 {
				if _, ok := args[2].(int64); !ok {
					return nil, loxerror.RuntimeError(in.callToken,
						"Third argument to 'range' must be an integer.")
				}
				step = args[2].(int64)
				if step == 0 {
					return nil, loxerror.RuntimeError(in.callToken,
						"Third argument to 'range' cannot be 0.")
				}
			} else {
				step = 1
			}
			return NewLoxRange(start, stop, step), nil
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1, 2, or 3 arguments but got %v.", argsLen))
		}
	})
	nativeFunc("repeatFunc", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'repeatFunc' must be an integer.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'repeatFunc' must be a function.")
		}
		times := args[0].(int64)
		if times > 0 {
			callback := args[1].(*LoxFunction)
			argList := getArgList(callback, 0)
			defer argList.Clear()
			for i := int64(0); i < times; i++ {
				result, resultErr := callback.call(in, argList)
				if resultErr != nil && result == nil {
					return nil, resultErr
				}
			}
		}
		return nil, nil
	})
	nativeFunc("Ring", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		if len(args) == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Ring: expected at least 1 argument but got 0.")
		}
		return NewLoxRingArgs(args...), nil
	})
	nativeFunc("RingIterable", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Iterable); ok {
			iterElements := list.NewList[any]()
			it := element.Iterator()
			for it.HasNext() {
				iterElements.Add(it.Next())
			}
			if len(iterElements) == 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Iterable argument to 'RingIterable' cannot be empty.")
			}
			return NewLoxRingArgs(iterElements...), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("RingIterable: type '%v' is not iterable.", getType(args[0])))
	})
	nativeFunc("RingList", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxList, ok := args[0].(*LoxList); ok {
			if len(loxList.elements) == 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"List argument to 'RingList' cannot be empty.")
			}
			return NewLoxRingArgs(loxList.elements...), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'RingList' must be a list.")
	})
	nativeFunc("RingNil", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			if num <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'RingNil' cannot be less than 1.")
			}
			return NewLoxRingNils(int(num)), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'RingNil' must be an integer.")
	})
	nativeFunc("Set", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		set := EmptyLoxSet()
		for _, element := range args {
			_, errStr := set.add(element)
			if len(errStr) > 0 {
				return nil, loxerror.RuntimeError(in.callToken, errStr)
			}
		}
		return set, nil
	})
	nativeFunc("SetIterable", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Iterable); ok {
			set := EmptyLoxSet()
			it := element.Iterator()
			for it.HasNext() {
				_, errStr := set.add(it.Next())
				if len(errStr) > 0 {
					return nil, loxerror.RuntimeError(in.callToken, errStr)
				}
			}
			return set, nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("SetIterable: type '%v' is not iterable.", getType(args[0])))
	})
	nativeFunc("sleep", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch seconds := args[0].(type) {
		case int64:
			time.Sleep(time.Duration(seconds) * time.Second)
			return nil, nil
		case float64:
			duration, _ := time.ParseDuration(fmt.Sprintf("%vs", seconds))
			time.Sleep(duration)
			return nil, nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'sleep' must be an integer or float.")
	})
	nativeFunc("sum", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if element, ok := args[0].(interfaces.Iterable); ok {
			sum := &LoxInternalSum{int64(0)}
			it := element.Iterator()
			for it.HasNext() {
				err := sum.sum(it.Next())
				if err != nil {
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
			}
			return sum.element, nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("sum: type '%v' is not iterable.", getType(args[0])))
	})
	nativeFunc("threadFunc", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'threadFunc' must be an integer.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'threadFunc' must be a function.")
		}
		if !util.UnsafeMode {
			return nil, loxerror.RuntimeError(in.callToken,
				"Cannot call 'threadFunc' in non-unsafe mode.")
		}
		times := args[0].(int64)
		if times > 0 {
			type errStruct struct {
				err error
				num int64
			}
			errorChan := make(chan errStruct, times)
			callbackChan := make(chan struct{}, times)
			callback := args[1].(*LoxFunction)
			for i := int64(0); i < times; i++ {
				go func(num int64) {
					argList := getArgList(callback, 1)
					argList[0] = num
					result, resultErr := callback.call(in, argList)
					if resultErr != nil && result == nil {
						errorChan <- errStruct{resultErr, num}
					} else {
						callbackChan <- struct{}{}
					}
					argList.Clear()
				}(i + 1)
			}
			for i := int64(0); i < times; i++ {
				select {
				case errStruct := <-errorChan:
					fmt.Fprintf(
						os.Stderr,
						"Runtime error in thread #%v: %v\n",
						errStruct.num,
						strings.ReplaceAll(errStruct.err.Error(), "\n", " "),
					)
				case <-callbackChan:
				}
			}
		}
		return nil, nil
	})
	nativeFunc("threadFuncs", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Expected at least 2 arguments but got 0.")
		}
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'threadFuncs' must be an integer.")
		}
		if argsLen == 1 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Expected at least 2 arguments but got 1.")
		}
		numCallbacks := int64(argsLen) - 1
		callbacks := list.NewListCap[*LoxFunction](numCallbacks)
		for i := 1; i < argsLen; i++ {
			switch arg := args[i].(type) {
			case *LoxFunction:
				callbacks.Add(arg)
			default:
				callbacks.Clear()
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"Argument number %v to 'threadFuncs' must be a function.",
						i+1,
					),
				)
			}
		}
		if !util.UnsafeMode {
			callbacks.Clear()
			return nil, loxerror.RuntimeError(in.callToken,
				"Cannot call 'threadFuncs' in non-unsafe mode.")
		}
		times := args[0].(int64)
		if times > 0 {
			type errStruct struct {
				err error
				num int64
			}
			numThreads := times * numCallbacks
			errorChan := make(chan errStruct, numThreads)
			callbackChan := make(chan struct{}, numThreads)
			threadNum := int64(1)
			for i := int64(0); i < times; i++ {
				for _, callback := range callbacks {
					go func(num int64) {
						argList := getArgList(callback, 1)
						argList[0] = num
						result, resultErr := callback.call(in, argList)
						if resultErr != nil && result == nil {
							errorChan <- errStruct{resultErr, num}
						} else {
							callbackChan <- struct{}{}
						}
						argList.Clear()
					}(threadNum)
					threadNum++
				}
			}
			for i := int64(0); i < numThreads; i++ {
				select {
				case errStruct := <-errorChan:
					fmt.Fprintf(
						os.Stderr,
						"Runtime error in thread #%v: %v\n",
						errStruct.num,
						strings.ReplaceAll(errStruct.err.Error(), "\n", " "),
					)
				case <-callbackChan:
				}
			}
		}
		callbacks.Clear()
		return nil, nil
	})
	nativeFunc("type", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		return NewLoxString(getType(args[0]), '\''), nil
	})
}
