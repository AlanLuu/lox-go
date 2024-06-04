package ast

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
)

func (i *Interpreter) defineOSFuncs() {
	className := "os"
	osClass := NewLoxClass(className, nil, false)
	osFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native os fn %v at %p>", name, &s)
		}
		osClass.classProperties[name] = s
	}
	argMustBeType := func(name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'os.%v' must be a %v.", name, theType)
		return nil, loxerror.Error(errStr)
	}

	osFunc("chdir", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := os.Chdir(loxStr.str)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		return argMustBeType("chdir", "string")
	})
	osFunc("chmod", 2, func(_ *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.Error("First argument to 'os.chmod' must be a string.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.Error("Second argument to 'os.chmod' must be an integer.")
		}
		file := args[0].(*LoxString).str
		mode := args[1].(int64)
		err := os.Chmod(file, os.FileMode(mode))
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	osFunc("exit", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		exitCode := 0
		argsLen := len(args)
		switch argsLen {
		case 0:
		case 1:
			if code, ok := args[0].(int64); ok {
				exitCode = int(code)
			} else {
				return nil, loxerror.Error("Argument to 'os.exit' must be an integer.")
			}
		default:
			return nil, loxerror.Error(fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		os.Exit(exitCode)
		return nil, nil
	})
	osFunc("getcwd", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		return NewLoxStringQuote(cwd), nil
	})
	osFunc("getuid", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(os.Getuid()), nil
	})
	osFunc("hostname", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		return NewLoxStringQuote(hostname), nil
	})
	osFunc("mkdir", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := os.Mkdir(loxStr.str, 0777)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		return argMustBeType("mkdir", "string")
	})
	osClass.classProperties["name"] = NewLoxString(runtime.GOOS, '\'')
	osFunc("remove", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := os.Remove(loxStr.str)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		return argMustBeType("remove", "string")
	})
	osFunc("removeAll", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := os.RemoveAll(loxStr.str)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		return argMustBeType("removeAll", "string")
	})
	osFunc("system", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("cmd", "/c", loxStr.str)
			} else {
				cmd = exec.Command("sh", "-c", loxStr.str)
			}
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					return int64(exitErr.ExitCode()), nil
				} else {
					return nil, err
				}
			}
			return int64(0), nil
		}
		return argMustBeType("system", "string")
	})
	osFunc("touch", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			file, fileErr := os.Create(loxStr.str)
			if fileErr != nil {
				return nil, fileErr
			}
			file.Close()
			return nil, nil
		}
		return argMustBeType("touch", "string")
	})
	osFunc("username", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		currentUser, err := user.Current()
		if err != nil {
			return nil, err
		}
		return NewLoxStringQuote(currentUser.Username), nil
	})

	i.globals.Define(className, osClass)
}
