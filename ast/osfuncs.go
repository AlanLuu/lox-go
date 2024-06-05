package ast

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/util"
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
	osFunc("getenv", -1, func(_ *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		var defaultValue any
		switch argsLen {
		case 1:
		case 2:
			defaultValue = args[1]
		default:
			return nil, loxerror.Error(fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
		switch arg := args[0].(type) {
		case *LoxString:
			value, ok := os.LookupEnv(arg.str)
			if !ok {
				return defaultValue, nil
			}
			return NewLoxStringQuote(value), nil
		default:
			return nil, loxerror.Error("First argument to 'os.getenv' must be a string.")
		}
	})
	osFunc("getenvs", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		envsDict := EmptyLoxDict()
		envs := os.Environ()
		for _, env := range envs {
			envSplit := strings.Split(env, "=")
			key := envSplit[0]
			value := envSplit[1]
			envsDict.setKeyValue(NewLoxStringQuote(key), NewLoxStringQuote(value))
		}
		return envsDict, nil
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
	osFunc("setenv", 2, func(_ *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.Error("First argument to 'os.setenv' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.Error("Second argument to 'os.setenv' must be a string.")
		}
		key := args[0].(*LoxString).str
		value := args[1].(*LoxString).str
		err := os.Setenv(key, value)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	osFunc("system", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			var cmd *exec.Cmd
			if util.IsWindows() {
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
	osFunc("unsetenv", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := os.Unsetenv(loxStr.str)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		return argMustBeType("unsetenv", "string")
	})
	osFunc("username", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		currentUser, err := user.Current()
		if err != nil {
			return nil, err
		}
		username := currentUser.Username
		if util.IsWindows() && strings.Contains(username, "\\") {
			contents := strings.Split(username, "\\")
			return NewLoxStringQuote(contents[len(contents)-1]), nil
		}
		return NewLoxStringQuote(username), nil
	})

	i.globals.Define(className, osClass)
}
