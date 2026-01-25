package ast

import (
	"fmt"
	"os"
	"strings"
	"sync"

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
		if _, ok := args[1].(LoxCallable); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'unsafe.threadFunc' must be a function.")
		}
		times := args[0].(int64)
		if times > 0 {
			callback := args[1].(LoxCallable)
			var wg sync.WaitGroup
			for i := int64(0); i < times; i++ {
				wg.Add(1)
				go func(num int64) {
					defer wg.Done()
					argList := getArgList(callback, 1)
					argList[0] = num
					result, resultErr := callback.call(in, argList)
					if resultErr != nil && result == nil {
						fmt.Fprintf(
							os.Stderr,
							"Runtime error in thread #%v: %v\n",
							num,
							strings.ReplaceAll(resultErr.Error(), "\n", " "),
						)
					}
					argList.Clear()
				}(i + 1)
			}
			wg.Wait()
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
		callbacks := list.NewListCap[LoxCallable](numCallbacks)
		for i := 1; i < argsLen; i++ {
			switch arg := args[i].(type) {
			case LoxCallable:
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
			callback := args[1].(LoxCallable)
			var wg sync.WaitGroup
			for i := int64(0); i < times; i++ {
				wg.Add(1)
				go func(num int64) {
					defer wg.Done()
					argList := getArgList(callback, 1)
					argList[0] = num
					result, resultErr := callback.call(in, argList)
					if resultErr != nil && result == nil {
						fmt.Fprintf(
							os.Stderr,
							"Runtime error in thread #%v: %v\n",
							num,
							strings.ReplaceAll(resultErr.Error(), "\n", " "),
						)
					}
					argList.Clear()
				}(i + 1)
			}
			wg.Wait()
		}
		callbacks.Clear()
		return nil, nil
	})

	i.globals.Define(className, unsafeClass)
}
