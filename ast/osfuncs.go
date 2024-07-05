package ast

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
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
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'os.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	osFunc("chdir", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := os.Chdir(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "chdir", "string")
	})
	osFunc("chmod", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.chmod' must be a string.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.chmod' must be an integer.")
		}
		file := args[0].(*LoxString).str
		mode := args[1].(int64)
		err := os.Chmod(file, os.FileMode(mode))
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("clearenv", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		os.Clearenv()
		return nil, nil
	})
	osFunc("executable", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		exePath, err := os.Executable()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxStringQuote(exePath), nil
	})
	osFunc("exit", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		exitCode := 0
		argsLen := len(args)
		switch argsLen {
		case 0:
		case 1:
			if code, ok := args[0].(int64); ok {
				exitCode = int(code)
			} else {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'os.exit' must be an integer.")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		CloseInputFuncReadline()
		os.Exit(exitCode)
		return nil, nil
	})
	osFunc("getcwd", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxStringQuote(cwd), nil
	})
	osFunc("getenv", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		var defaultValue any
		switch argsLen {
		case 1:
		case 2:
			defaultValue = args[1]
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
		switch arg := args[0].(type) {
		case *LoxString:
			value, ok := os.LookupEnv(arg.str)
			if !ok {
				return defaultValue, nil
			}
			return NewLoxStringQuote(value), nil
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.getenv' must be a string.")
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
	osFunc("getgid", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(os.Getgid()), nil
	})
	osFunc("getpid", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(os.Getpid()), nil
	})
	osFunc("getppid", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(os.Getppid()), nil
	})
	osFunc("getuid", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(os.Getuid()), nil
	})
	osFunc("hostname", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxStringQuote(hostname), nil
	})
	osFunc("link", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.link' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.link' must be a string.")
		}
		target := args[0].(*LoxString).str
		linkName := args[1].(*LoxString).str
		err := os.Link(target, linkName)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("listdir", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var path string
		argsLen := len(args)
		switch argsLen {
		case 0:
			path = "."
		case 1:
			if _, ok := args[0].(*LoxString); !ok {
				return argMustBeType(in.callToken, "listdir", "string")
			}
			path = args[0].(*LoxString).str
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		dirList := list.NewList[any]()
		dir := os.DirFS(path)
		dirFunc := func(_ string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			name := d.Name()
			if name == "." {
				return nil
			}
			dirList.Add(NewLoxStringQuote(name))
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		err := fs.WalkDir(dir, ".", dirFunc)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxList(dirList), nil
	})
	osFunc("mkdir", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := os.Mkdir(loxStr.str, 0777)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "mkdir", "string")
	})
	osClass.classProperties["name"] = NewLoxString(runtime.GOOS, '\'')
	osFunc("open", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.open' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.open' must be a string.")
		}
		path := args[0].(*LoxString).str
		mode := args[1].(*LoxString).str
		loxFile, loxFileErr := NewLoxFileModeStr(path, mode)
		if loxFileErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, loxFileErr.Error())
		}
		return loxFile, nil
	})
	osFunc("remove", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := os.Remove(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "remove", "string")
	})
	osFunc("removeAll", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := os.RemoveAll(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "removeAll", "string")
	})
	osClass.classProperties["SEEK_SET"] = int64(0)
	osClass.classProperties["SEEK_CUR"] = int64(1)
	osClass.classProperties["SEEK_END"] = int64(2)
	osFunc("setenv", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.setenv' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.setenv' must be a string.")
		}
		key := args[0].(*LoxString).str
		value := args[1].(*LoxString).str
		err := os.Setenv(key, value)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("symlink", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.symlink' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.symlink' must be a string.")
		}
		target := args[0].(*LoxString).str
		linkName := args[1].(*LoxString).str
		err := os.Symlink(target, linkName)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("system", 1, func(in *Interpreter, args list.List[any]) (any, error) {
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
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
			}
			return int64(0), nil
		}
		return argMustBeType(in.callToken, "system", "string")
	})
	osFunc("touch", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			file, fileErr := os.Create(loxStr.str)
			if fileErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, fileErr.Error())
			}
			file.Close()
			return nil, nil
		}
		return argMustBeType(in.callToken, "touch", "string")
	})
	osFunc("unsetenv", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := os.Unsetenv(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "unsetenv", "string")
	})
	osFunc("username", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		currentUser, err := user.Current()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
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
