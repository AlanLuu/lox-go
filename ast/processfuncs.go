package ast

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/util"
)

func (i *Interpreter) defineProcessFuncs() {
	className := "process"
	processClass := NewLoxClass(className, nil, false)
	processFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native process class fn %v at %p>", name, &s)
		}
		processClass.classProperties[name] = s
	}
	getExecCmd := func(funcName string, isShell bool, in *Interpreter, args list.List[any]) (*exec.Cmd, error) {
		argsLen := len(args)
		if argsLen == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Expected at least 1 argument but got 0.")
		}
		var cmdArgs []string
		setCmdArgs := func(elementsLen int) {
			if isShell {
				cmdArgs = make([]string, 0, elementsLen+2)
				if util.IsWindows() {
					cmdArgs = append(cmdArgs, "cmd")
					cmdArgs = append(cmdArgs, "/c")
				} else {
					cmdArgs = append(cmdArgs, "sh")
					cmdArgs = append(cmdArgs, "-c")
				}
			} else {
				cmdArgs = make([]string, 0, elementsLen)
			}
		}
		switch cmd := args[0].(type) {
		case *LoxList:
			if argsLen != 1 {
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Only 1 argument can be passed when passing a list to '%v'.", funcName))
			}
			elementsLen := len(cmd.elements)
			if elementsLen == 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("List argument to '%v' must not be empty.", funcName))
			}
			setCmdArgs(elementsLen)
			for _, element := range cmd.elements {
				switch element := element.(type) {
				case *LoxString:
					cmdArgs = append(cmdArgs, element.str)
				default:
					cmdArgs = nil
					return nil, loxerror.RuntimeError(in.callToken,
						fmt.Sprintf("List argument to '%v' must only have strings.", funcName))
				}
			}
		case *LoxString:
			setCmdArgs(argsLen)
			for _, arg := range args {
				switch arg := arg.(type) {
				case *LoxString:
					cmdArgs = append(cmdArgs, arg.str)
				default:
					cmdArgs = nil
					return nil, loxerror.RuntimeError(in.callToken,
						fmt.Sprintf("Arguments to '%v' must be strings.", funcName))
				}
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Arguments to '%v' must be a list or strings of arguments.", funcName))
		}
		if len(cmdArgs) == 1 {
			return exec.Command(cmdArgs[0]), nil
		}
		return exec.Command(cmdArgs[0], cmdArgs[1:]...), nil
	}
	methodName := func(name string) string {
		return "process class." + name
	}
	setStd := func(cmd *exec.Cmd) {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	processFunc("new", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		cmd, err := getExecCmd(methodName("new"), false, in, args)
		if err != nil {
			return nil, err
		}
		return NewLoxProcess(cmd), nil
	})
	processFunc("newReusable", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		cmd, err := getExecCmd(methodName("newReusable"), false, in, args)
		if err != nil {
			return nil, err
		}
		return NewLoxProcessReusable(cmd), nil
	})
	processFunc("newReusableSetStd", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		cmd, err := getExecCmd(methodName("newReusableSetStd"), false, in, args)
		if err != nil {
			return nil, err
		}
		setStd(cmd)
		return NewLoxProcessReusable(cmd), nil
	})
	processFunc("newSetStd", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		cmd, err := getExecCmd(methodName("newSetStd"), false, in, args)
		if err != nil {
			return nil, err
		}
		setStd(cmd)
		return NewLoxProcess(cmd), nil
	})
	processFunc("newShell", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		cmd, err := getExecCmd(methodName("newShell"), true, in, args)
		if err != nil {
			return nil, err
		}
		return NewLoxProcess(cmd), nil
	})
	processFunc("newShellReusable", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		cmd, err := getExecCmd(methodName("newShellReusable"), true, in, args)
		if err != nil {
			return nil, err
		}
		return NewLoxProcessReusable(cmd), nil
	})
	processFunc("newShellReusableSetStd", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		cmd, err := getExecCmd(methodName("newShellReusableSetStd"), true, in, args)
		if err != nil {
			return nil, err
		}
		setStd(cmd)
		return NewLoxProcessReusable(cmd), nil
	})
	processFunc("newShellSetStd", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		cmd, err := getExecCmd(methodName("newShellSetStd"), true, in, args)
		if err != nil {
			return nil, err
		}
		setStd(cmd)
		return NewLoxProcess(cmd), nil
	})
	processFunc("run", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		cmd, err := getExecCmd(methodName("run"), false, in, args)
		if err != nil {
			return nil, err
		}
		process := NewLoxProcess(cmd)
		if err := process.run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return NewLoxProcessResult(exitErr.ProcessState), nil
			} else {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
		}
		return NewLoxProcessResult(process.process.ProcessState), nil
	})
	processFunc("runSetStd", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		cmd, err := getExecCmd(methodName("runSetStd"), false, in, args)
		if err != nil {
			return nil, err
		}
		setStd(cmd)
		process := NewLoxProcess(cmd)
		if err := process.run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return NewLoxProcessResult(exitErr.ProcessState), nil
			} else {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
		}
		return NewLoxProcessResult(process.process.ProcessState), nil
	})
	processFunc("runShell", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		cmd, err := getExecCmd(methodName("runShell"), true, in, args)
		if err != nil {
			return nil, err
		}
		process := NewLoxProcess(cmd)
		if err := process.run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return NewLoxProcessResult(exitErr.ProcessState), nil
			} else {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
		}
		return NewLoxProcessResult(process.process.ProcessState), nil
	})

	i.globals.Define(className, processClass)
}
