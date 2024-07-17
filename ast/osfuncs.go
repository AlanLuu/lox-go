package ast

import (
	crand "crypto/rand"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"syscall"

	"github.com/AlanLuu/lox/ast/filemode"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/syscalls"
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
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'os.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	stdStream := func(stream *os.File, mode filemode.FileMode, isBinary bool) *LoxFile {
		return &LoxFile{
			file:       stream,
			name:       stream.Name(),
			mode:       mode,
			isBinary:   isBinary,
			isClosed:   false,
			properties: make(map[string]any),
		}
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
	osFunc("chown", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.chown' must be a string.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.chown' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'os.chown' must be an integer.")
		}
		if util.IsWindows() {
			return nil, loxerror.RuntimeError(in.callToken,
				"'os.chown' is unsupported on Windows.")
		}
		file := args[0].(*LoxString).str
		uid := int(args[1].(int64))
		gid := int(args[2].(int64))
		err := os.Chown(file, uid, gid)
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
	osFunc("getegid", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(os.Getegid()), nil
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
	osFunc("geteuid", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(os.Geteuid()), nil
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
	osFunc("kill", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		switch argsLen {
		case 1, 2:
			if pid, ok := args[0].(int64); ok {
				if argsLen == 2 {
					if _, ok := args[1].(int64); !ok {
						return nil, loxerror.RuntimeError(in.callToken,
							"Second argument to 'os.kill' must be an integer.")
					}
				}
				noSuchProcessMsg := "No such process with PID %v."
				process, err := os.FindProcess(int(pid))
				if err != nil {
					if util.IsWindows() {
						return nil, loxerror.RuntimeError(in.callToken,
							fmt.Sprintf(noSuchProcessMsg, pid))
					}
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}

				/*
					os.FindProcess always returns a process object on Unix systems
					regardless of whether a process with the specified PID actually
					exists, so check for its existence by checking if there is a
					process with the specified PID running or not
				*/
				if !util.IsWindows() {
					err = process.Signal(syscall.Signal(0))
					if err != nil {
						return nil, loxerror.RuntimeError(in.callToken,
							fmt.Sprintf(noSuchProcessMsg, pid))
					}
				}

				if argsLen == 2 {
					sigNum := args[1].(int64)
					err = process.Signal(syscall.Signal(sigNum))
				} else {
					err = process.Kill()
				}

				if err != nil {
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
				return nil, nil
			} else if argsLen == 2 {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'os.kill' must be an integer.")
			}
			return nil, loxerror.RuntimeError(in.callToken,
				"Argument to 'os.kill' must be an integer.")
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
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
	osFunc("mkfifo", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := syscalls.Mkfifo(loxStr.str, 0666)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "mkfifo", "string")
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
	osFunc("pipe", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		r, w, err := os.Pipe()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		files := list.NewList[any]()
		files.Add(&LoxFile{
			file:       r,
			name:       r.Name(),
			mode:       filemode.READ,
			isBinary:   false,
			isClosed:   false,
			properties: make(map[string]any),
		})
		files.Add(&LoxFile{
			file:       w,
			name:       w.Name(),
			mode:       filemode.WRITE,
			isBinary:   false,
			isClosed:   false,
			properties: make(map[string]any),
		})
		return NewLoxList(files), nil
	})
	osFunc("pipeBin", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		r, w, err := os.Pipe()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		files := list.NewList[any]()
		files.Add(&LoxFile{
			file:       r,
			name:       r.Name(),
			mode:       filemode.READ,
			isBinary:   true,
			isClosed:   false,
			properties: make(map[string]any),
		})
		files.Add(&LoxFile{
			file:       w,
			name:       w.Name(),
			mode:       filemode.WRITE,
			isBinary:   true,
			isClosed:   false,
			properties: make(map[string]any),
		})
		return NewLoxList(files), nil
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
	osFunc("rename", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.rename' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.rename' must be a string.")
		}
		oldPath := args[0].(*LoxString).str
		newPath := args[1].(*LoxString).str
		err := os.Rename(oldPath, newPath)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
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
	osFunc("setgid", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if gid, ok := args[0].(int64); ok {
			err := syscalls.Setgid(int(gid))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeTypeAn(in.callToken, "setgid", "integer")
	})
	osFunc("setuid", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if uid, ok := args[0].(int64); ok {
			err := syscalls.Setuid(int(uid))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeTypeAn(in.callToken, "setuid", "integer")
	})
	osClass.classProperties["stderr"] = stdStream(os.Stderr, filemode.WRITE, false)
	osClass.classProperties["stdin"] = stdStream(os.Stdin, filemode.READ, false)
	osClass.classProperties["stdout"] = stdStream(os.Stdout, filemode.WRITE, false)
	osClass.classProperties["stderrBin"] = stdStream(os.Stderr, filemode.WRITE, true)
	osClass.classProperties["stdinBin"] = stdStream(os.Stdin, filemode.READ, true)
	osClass.classProperties["stdoutBin"] = stdStream(os.Stdout, filemode.WRITE, true)
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
	osFunc("urandom", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if numBytes, ok := args[0].(int64); ok {
			if numBytes < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'os.urandom' cannot be negative.")
			}
			buffer := EmptyLoxBuffer()
			for i := int64(0); i < numBytes; i++ {
				numBig, numErr := crand.Int(crand.Reader, big.NewInt(256))
				if numErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, numErr.Error())
				}
				addErr := buffer.add(numBig.Int64())
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
			return buffer, nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			"Argument to 'os.urandom' must be an integer.")
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
