package ast

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/util"
)

func (i *Interpreter) defineUnsafeFuncs() {
	if !util.UnsafeMode {
		return
	}
	className := "unsafe"
	unsafeClass := NewLoxClass(className, nil, false)
	unsafeFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native unsafe fn %v at %p>", name, &s)
		}
		unsafeClass.classProperties[name] = s
	}

	unsafeFunc("threadFunc", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'unsafe.threadFunc' must be an integer.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'unsafe.threadFunc' must be a function.")
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
	unsafeFunc("threadFuncs", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Expected at least 2 arguments but got 0.")
		}
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'unsafe.threadFuncs' must be an integer.")
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
						"Argument number %v to 'unsafe.threadFuncs' must be a function.",
						i+1,
					),
				)
			}
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

	i.globals.Define(className, unsafeClass)
}
