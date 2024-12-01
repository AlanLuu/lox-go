package ast

import (
	crand "crypto/rand"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/AlanLuu/lox/ast/filemode"
	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/syscalls"
	"github.com/AlanLuu/lox/syscalls/linuxsyscalls"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
	"github.com/mattn/go-isatty"
)

func cmdArgsToLoxList() *LoxList {
	args := flag.Args()
	argvList := list.NewListCap[any](int64(len(args)) + 1)
	execPath, err := os.Executable()
	if err == nil {
		argvList.Add(NewLoxStringQuote(execPath))
	} else {
		argvList.Add(EmptyLoxString())
	}

	for _, arg := range args {
		argvList.Add(NewLoxStringQuote(arg))
	}

	return NewLoxList(argvList)
}

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
			stat:       nil,
			properties: make(map[string]any),
		}
	}

	osClass.classProperties["argv"] = cmdArgsToLoxList()
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
	osFunc("chroot", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := syscalls.Chroot(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "chroot", "string")
	})
	osFunc("clearenv", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		os.Clearenv()
		return nil, nil
	})
	osFunc("close", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if fd, ok := args[0].(int64); ok {
			err := syscalls.Close(int(fd))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeTypeAn(in.callToken, "close", "integer")
	})
	osFunc("closeRange", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.closeRange' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.closeRange' must be an integer.")
		}
		low := args[0].(int64)
		high := args[1].(int64)
		for fd := low; fd < high; fd++ {
			syscalls.Close(int(fd))
		}
		return nil, nil
	})
	osFunc("copy", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.copy' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.copy' must be a string.")
		}

		sourceStr := args[0].(*LoxString).str
		sourceStat, sourceStatErr := os.Stat(sourceStr)
		if sourceStatErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, sourceStatErr.Error())
		}
		if sourceStat.IsDir() {
			return nil, loxerror.RuntimeError(in.callToken,
				"Cannot use 'os.copy' to copy directory.")
		}

		destStr := args[1].(*LoxString).str
		destStat, destStatErr := os.Stat(destStr)
		if destStatErr == nil && destStat.IsDir() {
			var pathSep string
			if strings.Contains(sourceStr, "/") {
				pathSep = "/"
			} else if util.IsWindows() && strings.Contains(sourceStr, "\\") {
				pathSep = "\\"
			}
			if len(pathSep) > 0 {
				splitList := strings.Split(sourceStr, pathSep)
				index := len(splitList) - 1
				for index > 0 && splitList[index] == "" {
					index--
				}
				destStr = filepath.Join(destStr, splitList[index])
			} else {
				destStr = filepath.Join(destStr, sourceStr)
			}
		}

		source, sourceErr := os.Open(sourceStr)
		if sourceErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, sourceErr.Error())
		}
		defer source.Close()

		dest, destErr := os.OpenFile(destStr, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if destErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, destErr.Error())
		}
		defer dest.Close()

		numBytes, copyErr := io.Copy(dest, source)
		if copyErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, copyErr.Error())
		}
		return numBytes, nil
	})
	osFunc("dup", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if fd, ok := args[0].(int64); ok {
			newFd, err := syscalls.Dup(int(fd))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return int64(newFd), nil
		}
		return argMustBeTypeAn(in.callToken, "dup", "integer")
	})
	osFunc("dup2", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.dup2' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.dup2' must be an integer.")
		}
		oldfd := args[0].(int64)
		newfd := args[1].(int64)
		err := syscalls.Dup2(int(oldfd), int(newfd))
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return newfd, nil
	})
	osClass.classProperties["devNull"] = NewLoxLazyFile(os.DevNull, filemode.READ_WRITE, false)
	osClass.classProperties["devNullBin"] = NewLoxLazyFile(os.DevNull, filemode.READ_WRITE, true)
	if !util.IsWindows() {
		osClass.classProperties["devFull"] = NewLoxLazyFile("/dev/full", filemode.READ_WRITE, false)
		osClass.classProperties["devFullBin"] = NewLoxLazyFile("/dev/full", filemode.READ_WRITE, true)
		osClass.classProperties["devRandom"] = NewLoxLazyFile("/dev/random", filemode.READ_WRITE, true)
		osClass.classProperties["devUrandom"] = NewLoxLazyFile("/dev/urandom", filemode.READ_WRITE, true)
		osClass.classProperties["devZero"] = NewLoxLazyFile("/dev/zero", filemode.READ_WRITE, true)
	}
	osFunc("execl", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen == 0 || argsLen == 1 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected at least 2 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.execl' must be a string.")
		}

		argv := list.NewListCap[string](int64(argsLen - 1))
		for i := 1; i < argsLen; i++ {
			switch element := args[i].(type) {
			case *LoxString:
				argv.Add(element.str)
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken,
					"All arguments to 'os.execl' after the first must be strings.")
			}
		}
		path := args[0].(*LoxString).str
		err := syscalls.Execv(path, argv)
		if err != nil {
			errMsg := strings.Replace(err.Error(), "execv", "execl", 1)
			return nil, loxerror.RuntimeError(in.callToken, errMsg)
		}
		return nil, nil
	})
	osFunc("execle", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen < 3 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected at least 3 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.execle' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.execle' must be a string.")
		}

		argv := list.NewListCap[string](int64(argsLen - 2))
		lastArgErrMsg := "Last argument to 'os.execle' must be a dictionary."
		envDictIndex := -1
		for i := 1; i <= argsLen; i++ {
			if i != argsLen && envDictIndex != -1 {
				argv.Clear()
				for ; i < argsLen; i++ {
					switch args[i].(type) {
					case *LoxDict:
						return nil, loxerror.RuntimeError(in.callToken,
							"Only one dictionary argument can be passed to 'os.execle'.")
					}
				}
				return nil, loxerror.RuntimeError(in.callToken, lastArgErrMsg)
			} else if i == argsLen {
				if envDictIndex == -1 {
					argv.Clear()
					return nil, loxerror.RuntimeError(in.callToken, lastArgErrMsg)
				}
				break
			}
			switch element := args[i].(type) {
			case *LoxString:
				argv.Add(element.str)
			case *LoxDict:
				envDictIndex = i
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken,
					"All arguments to 'os.execle' after the second must be strings or a dictionary.")
			}
		}

		strDictErrMsg := "Environment dictionary in 'os.execle' must only have strings."
		envDict := args[envDictIndex].(*LoxDict)
		envp := list.NewList[string]()
		it := envDict.Iterator()
		for it.HasNext() {
			var builder strings.Builder
			pair := it.Next().(*LoxList).elements
			switch key := pair[0].(type) {
			case *LoxString:
				builder.WriteString(key.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			builder.WriteRune('=')
			switch value := pair[1].(type) {
			case *LoxString:
				builder.WriteString(value.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			envp.Add(builder.String())
		}

		path := args[0].(*LoxString).str
		err := syscalls.Execve(path, argv, envp)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("execlp", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen == 0 || argsLen == 1 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected at least 2 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.execlp' must be a string.")
		}

		argv := list.NewListCap[string](int64(argsLen - 1))
		for i := 1; i < argsLen; i++ {
			switch element := args[i].(type) {
			case *LoxString:
				argv.Add(element.str)
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken,
					"All arguments to 'os.execlp' after the first must be strings.")
			}
		}
		file := args[0].(*LoxString).str
		err := syscalls.Execvp(file, argv)
		if err != nil {
			errMsg := strings.Replace(err.Error(), "execvp", "execlp", 1)
			return nil, loxerror.RuntimeError(in.callToken, errMsg)
		}
		return nil, nil
	})
	osFunc("execlpe", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen < 3 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected at least 3 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.execlpe' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.execlpe' must be a string.")
		}

		argv := list.NewListCap[string](int64(argsLen - 2))
		lastArgErrMsg := "Last argument to 'os.execlpe' must be a dictionary."
		envDictIndex := -1
		for i := 1; i <= argsLen; i++ {
			if i != argsLen && envDictIndex != -1 {
				argv.Clear()
				for ; i < argsLen; i++ {
					switch args[i].(type) {
					case *LoxDict:
						return nil, loxerror.RuntimeError(in.callToken,
							"Only one dictionary argument can be passed to 'os.execlpe'.")
					}
				}
				return nil, loxerror.RuntimeError(in.callToken, lastArgErrMsg)
			} else if i == argsLen {
				if envDictIndex == -1 {
					argv.Clear()
					return nil, loxerror.RuntimeError(in.callToken, lastArgErrMsg)
				}
				break
			}
			switch element := args[i].(type) {
			case *LoxString:
				argv.Add(element.str)
			case *LoxDict:
				envDictIndex = i
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken,
					"All arguments to 'os.execlpe' after the second must be strings or a dictionary.")
			}
		}

		strDictErrMsg := "Environment dictionary in 'os.execlpe' must only have strings."
		envDict := args[envDictIndex].(*LoxDict)
		envp := list.NewList[string]()
		it := envDict.Iterator()
		for it.HasNext() {
			var builder strings.Builder
			pair := it.Next().(*LoxList).elements
			switch key := pair[0].(type) {
			case *LoxString:
				builder.WriteString(key.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			builder.WriteRune('=')
			switch value := pair[1].(type) {
			case *LoxString:
				builder.WriteString(value.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			envp.Add(builder.String())
		}

		file := args[0].(*LoxString).str
		err := syscalls.Execvpe(file, argv, envp)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("executable", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		exePath, err := os.Executable()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxStringQuote(exePath), nil
	})
	osFunc("execv", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.execv' must be a string.")
		}
		if _, ok := args[1].(*LoxList); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.execv' must be a list.")
		}

		strListErrMsg := "Second argument to 'os.execv' must be a list of strings."
		argvList := args[1].(*LoxList).elements
		if argvList.IsEmpty() {
			return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
		}
		argv := list.NewListCap[string](int64(len(argvList)))
		for _, element := range argvList {
			switch element := element.(type) {
			case *LoxString:
				argv.Add(element.str)
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
			}
		}
		path := args[0].(*LoxString).str
		err := syscalls.Execv(path, argv)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("execve", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.execve' must be a string.")
		}
		if _, ok := args[1].(*LoxList); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.execve' must be a list.")
		}
		if _, ok := args[2].(*LoxDict); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'os.execve' must be a dictionary.")
		}

		strListErrMsg := "Second argument to 'os.execve' must be a list of strings."
		argvList := args[1].(*LoxList).elements
		if argvList.IsEmpty() {
			return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
		}
		argv := list.NewListCap[string](int64(len(argvList)))
		for _, element := range argvList {
			switch element := element.(type) {
			case *LoxString:
				argv.Add(element.str)
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
			}
		}

		strDictErrMsg := "Third argument to 'os.execve' must be a dictionary with only strings."
		envDict := args[2].(*LoxDict)
		envp := list.NewList[string]()
		it := envDict.Iterator()
		for it.HasNext() {
			var builder strings.Builder
			pair := it.Next().(*LoxList).elements
			switch key := pair[0].(type) {
			case *LoxString:
				builder.WriteString(key.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			builder.WriteRune('=')
			switch value := pair[1].(type) {
			case *LoxString:
				builder.WriteString(value.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			envp.Add(builder.String())
		}

		path := args[0].(*LoxString).str
		err := syscalls.Execve(path, argv, envp)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("execvp", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.execvp' must be a string.")
		}
		if _, ok := args[1].(*LoxList); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.execvp' must be a list.")
		}

		strListErrMsg := "Second argument to 'os.execvp' must be a list of strings."
		argvList := args[1].(*LoxList).elements
		if argvList.IsEmpty() {
			return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
		}
		argv := list.NewListCap[string](int64(len(argvList)))
		for _, element := range argvList {
			switch element := element.(type) {
			case *LoxString:
				argv.Add(element.str)
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
			}
		}
		file := args[0].(*LoxString).str
		err := syscalls.Execvp(file, argv)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("execvpe", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.execvpe' must be a string.")
		}
		if _, ok := args[1].(*LoxList); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.execvpe' must be a list.")
		}
		if _, ok := args[2].(*LoxDict); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'os.execvpe' must be a dictionary.")
		}

		strListErrMsg := "Second argument to 'os.execvpe' must be a list of strings."
		argvList := args[1].(*LoxList).elements
		if argvList.IsEmpty() {
			return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
		}
		argv := list.NewListCap[string](int64(len(argvList)))
		for _, element := range argvList {
			switch element := element.(type) {
			case *LoxString:
				argv.Add(element.str)
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
			}
		}

		strDictErrMsg := "Third argument to 'os.execvpe' must be a dictionary with only strings."
		envDict := args[2].(*LoxDict)
		envp := list.NewList[string]()
		it := envDict.Iterator()
		for it.HasNext() {
			var builder strings.Builder
			pair := it.Next().(*LoxList).elements
			switch key := pair[0].(type) {
			case *LoxString:
				builder.WriteString(key.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			builder.WriteRune('=')
			switch value := pair[1].(type) {
			case *LoxString:
				builder.WriteString(value.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			envp.Add(builder.String())
		}

		file := args[0].(*LoxString).str
		err := syscalls.Execvpe(file, argv, envp)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("expandEnv", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(os.ExpandEnv(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "expandEnv", "string")
	})
	osFunc("expandHome", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			pathStr := loxStr.str
			if pathStr == "~" || strings.HasPrefix(pathStr, "~/") {
				homeDir, homeIsSet := os.LookupEnv("HOME")
				if !homeIsSet || util.IsWindows() {
					currentUser, err := user.Current()
					if err != nil {
						return nil, loxerror.RuntimeError(in.callToken, err.Error())
					}
					homeDir = currentUser.HomeDir
				}
				if pathStr == "~" {
					pathStr = homeDir
				} else {
					pathStr = filepath.Join(homeDir, pathStr[2:])
				}
			}
			return NewLoxStringQuote(pathStr), nil
		}
		return argMustBeType(in.callToken, "expandHome", "string")
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
	osFunc("fallocate", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		switch args[0].(type) {
		case int64:
		case *LoxFile:
		case *LoxString:
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.fallocate' must be an integer, file, or string.")
		}

		var size int64
		switch arg := args[1].(type) {
		case int64:
			size = arg
		case *LoxString:
			runes := []rune(arg.str)
			runesLen := len(runes)
			if runesLen < 2 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second string argument to 'os.fallocate' must have at least 2 characters.")
			}
			lastIndex := runesLen - 1

			sizeNumStr := string(runes[:lastIndex])
			sizeNum, convertErr := strconv.ParseInt(sizeNumStr, 10, 64)
			if convertErr != nil {
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Failed to convert '%v' to integer.", sizeNumStr))
			}

			const kb = 1024
			const mb = 1024 * 1024
			const gb = 1024 * 1024 * 1024
			sizes := map[rune]int64{
				'b': kb / 2,
				'k': kb,
				'K': kb,
				'm': mb,
				'M': mb,
				'g': gb,
				'G': gb,
			}
			sizeChar := runes[lastIndex]
			if sizeChar >= '0' && sizeChar <= '9' {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second string argument to 'os.fallocate' must end with a size suffix.")
			}
			if sizeConstant, ok := sizes[sizeChar]; ok {
				size = sizeConstant * sizeNum
			} else {
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("'%v' is not a valid size suffix.", string(sizeChar)))
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.fallocate' must be an integer or string.")
		}

		if size <= 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Size argument to 'os.fallocate' cannot be 0 or negative.")
		}

		var err error
		switch arg := args[0].(type) {
		case int64:
			if arg < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"File descriptor argument to 'os.fallocate' cannot be negative.")
			}
			badFileDesc := "bad file descriptor"
			if util.IsLinux() {
				err = linuxsyscalls.Fallocate(int(arg), 0, 0, size)
				if err != nil {
					file := os.NewFile(uintptr(arg), "fallocate")
					if file == nil {
						return nil, loxerror.RuntimeError(in.callToken, badFileDesc)
					}

					//Check if the file can actually be written to
					_, testWriteErr := file.WriteString("")
					if testWriteErr != nil {
						return nil, loxerror.RuntimeError(in.callToken, testWriteErr.Error())
					}

					stat, statErr := file.Stat()
					if statErr != nil {
						return nil, loxerror.RuntimeError(in.callToken, statErr.Error())
					}
					statSize := stat.Size()
					if size > statSize {
						_, err = file.Write(make([]byte, size-statSize))
					}
				}
			} else {
				file := os.NewFile(uintptr(arg), "fallocate")
				if file == nil {
					return nil, loxerror.RuntimeError(in.callToken, badFileDesc)
				}

				//Check if the file can actually be written to
				_, testWriteErr := file.WriteString("")
				if testWriteErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, testWriteErr.Error())
				}

				stat, statErr := file.Stat()
				if statErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, statErr.Error())
				}
				statSize := stat.Size()
				if size > statSize {
					_, err = file.Write(make([]byte, size-statSize))
				}
			}
		case *LoxFile:
			if !arg.isWrite() && !arg.isAppend() {
				return nil, loxerror.RuntimeError(in.callToken,
					"File argument to 'os.fallocate' must be in write or append mode.")
			}
			fd := int(arg.file.Fd())
			switch {
			case util.IsLinux():
				err = linuxsyscalls.Fallocate(fd, 0, 0, size)
				if err != nil {
					switch {
					case arg.mode == filemode.READ_WRITE:
						stat, statErr := arg.file.Stat()
						if statErr != nil {
							return nil, loxerror.RuntimeError(in.callToken, statErr.Error())
						}
						statSize := stat.Size()
						if size > statSize {
							originalOffset, seekErr1 := arg.file.Seek(0, io.SeekStart)
							if seekErr1 != nil {
								return nil, loxerror.RuntimeError(in.callToken, seekErr1.Error())
							}
							_, seekErr2 := arg.file.Seek(0, io.SeekEnd)
							if seekErr2 != nil {
								return nil, loxerror.RuntimeError(in.callToken, seekErr2.Error())
							}
							_, err = arg.file.Write(make([]byte, size-statSize))
							_, seekErr3 := arg.file.Seek(originalOffset, io.SeekStart)
							if seekErr3 != nil {
								return nil, loxerror.RuntimeError(in.callToken, seekErr3.Error())
							}
						}
					default:
						stat, statErr := arg.file.Stat()
						if statErr != nil {
							return nil, loxerror.RuntimeError(in.callToken, statErr.Error())
						}
						statSize := stat.Size()
						if size > statSize {
							_, err = arg.file.Write(make([]byte, size-statSize))
						}
					}
				}
			case arg.mode == filemode.READ_WRITE:
				stat, statErr := arg.file.Stat()
				if statErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, statErr.Error())
				}
				statSize := stat.Size()
				if size > statSize {
					originalOffset, seekErr1 := arg.file.Seek(0, io.SeekStart)
					if seekErr1 != nil {
						return nil, loxerror.RuntimeError(in.callToken, seekErr1.Error())
					}
					_, seekErr2 := arg.file.Seek(0, io.SeekEnd)
					if seekErr2 != nil {
						return nil, loxerror.RuntimeError(in.callToken, seekErr2.Error())
					}
					_, err = arg.file.Write(make([]byte, size-statSize))
					_, seekErr3 := arg.file.Seek(originalOffset, io.SeekStart)
					if seekErr3 != nil {
						return nil, loxerror.RuntimeError(in.callToken, seekErr3.Error())
					}
				}
			default:
				stat, statErr := arg.file.Stat()
				if statErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, statErr.Error())
				}
				statSize := stat.Size()
				if size > statSize {
					_, err = arg.file.Write(make([]byte, size-statSize))
				}
			}
		case *LoxString:
			file, fileErr := os.OpenFile(
				arg.str,
				os.O_WRONLY|os.O_CREATE|os.O_APPEND,
				0666,
			)
			if fileErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, fileErr.Error())
			}
			defer file.Close()
			if util.IsLinux() {
				err = linuxsyscalls.Fallocate(int(file.Fd()), 0, 0, size)
				if err != nil {
					stat, statErr := file.Stat()
					if statErr != nil {
						return nil, loxerror.RuntimeError(in.callToken, statErr.Error())
					}
					statSize := stat.Size()
					if size > statSize {
						_, err = file.Write(make([]byte, size-statSize))
					}
				}
			} else {
				stat, statErr := file.Stat()
				if statErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, statErr.Error())
				}
				statSize := stat.Size()
				if size > statSize {
					_, err = file.Write(make([]byte, size-statSize))
				}
			}
		}

		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("fchdir", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if fd, ok := args[0].(int64); ok {
			err := syscalls.Fchdir(int(fd))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeTypeAn(in.callToken, "fchdir", "integer")
	})
	osFunc("fchmod", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.fchmod' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.fchmod' must be an integer.")
		}
		fd := int(args[0].(int64))
		mode := uint32(args[1].(int64))
		err := syscalls.Fchmod(fd, mode)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("fchown", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.fchown' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.fchown' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'os.fchown' must be an integer.")
		}
		fd := int(args[0].(int64))
		uid := int(args[1].(int64))
		gid := int(args[2].(int64))
		err := syscalls.Fchown(fd, uid, gid)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("forkExec", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.forkExec' must be a string.")
		}
		if _, ok := args[1].(*LoxList); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.forkExec' must be a list.")
		}

		strListErrMsg := "Second argument to 'os.forkExec' must be a list of strings."
		argvList := args[1].(*LoxList).elements
		if argvList.IsEmpty() {
			return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
		}
		argv := list.NewListCap[string](int64(len(argvList)))
		for _, element := range argvList {
			switch element := element.(type) {
			case *LoxString:
				argv.Add(element.str)
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
			}
		}

		argv0 := args[0].(*LoxString).str
		pid, err := syscalls.ForkExecFd(argv0, argv)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return int64(pid), nil
	})
	osFunc("forkExecve", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.forkExecve' must be a string.")
		}
		if _, ok := args[1].(*LoxList); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.forkExecve' must be a list.")
		}
		if _, ok := args[2].(*LoxDict); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'os.forkExecve' must be a dictionary.")
		}

		strListErrMsg := "Second argument to 'os.forkExecve' must be a list of strings."
		argvList := args[1].(*LoxList).elements
		if argvList.IsEmpty() {
			return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
		}
		argv := list.NewListCap[string](int64(len(argvList)))
		for _, element := range argvList {
			switch element := element.(type) {
			case *LoxString:
				argv.Add(element.str)
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
			}
		}

		strDictErrMsg := "Third argument to 'os.forkExecve' must be a dictionary with only strings."
		envDict := args[2].(*LoxDict)
		envp := list.NewList[string]()
		it := envDict.Iterator()
		for it.HasNext() {
			var builder strings.Builder
			pair := it.Next().(*LoxList).elements
			switch key := pair[0].(type) {
			case *LoxString:
				builder.WriteString(key.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			builder.WriteRune('=')
			switch value := pair[1].(type) {
			case *LoxString:
				builder.WriteString(value.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			envp.Add(builder.String())
		}

		argv0 := args[0].(*LoxString).str
		pid, err := syscalls.ForkExecveFd(argv0, argv, envp)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return int64(pid), nil
	})
	osFunc("forkExecvp", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.forkExecvp' must be a string.")
		}
		if _, ok := args[1].(*LoxList); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.forkExecvp' must be a list.")
		}

		strListErrMsg := "Second argument to 'os.forkExecvp' must be a list of strings."
		argvList := args[1].(*LoxList).elements
		if argvList.IsEmpty() {
			return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
		}
		argv := list.NewListCap[string](int64(len(argvList)))
		for _, element := range argvList {
			switch element := element.(type) {
			case *LoxString:
				argv.Add(element.str)
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
			}
		}

		argv0 := args[0].(*LoxString).str
		pid, err := syscalls.ForkExecvpFd(argv0, argv)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return int64(pid), nil
	})
	osFunc("forkExecvpe", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.forkExecvpe' must be a string.")
		}
		if _, ok := args[1].(*LoxList); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.forkExecvpe' must be a list.")
		}
		if _, ok := args[2].(*LoxDict); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'os.forkExecvpe' must be a dictionary.")
		}

		strListErrMsg := "Second argument to 'os.forkExecvpe' must be a list of strings."
		argvList := args[1].(*LoxList).elements
		if argvList.IsEmpty() {
			return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
		}
		argv := list.NewListCap[string](int64(len(argvList)))
		for _, element := range argvList {
			switch element := element.(type) {
			case *LoxString:
				argv.Add(element.str)
			default:
				argv.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strListErrMsg)
			}
		}

		strDictErrMsg := "Third argument to 'os.forkExecvpe' must be a dictionary with only strings."
		envDict := args[2].(*LoxDict)
		envp := list.NewList[string]()
		it := envDict.Iterator()
		for it.HasNext() {
			var builder strings.Builder
			pair := it.Next().(*LoxList).elements
			switch key := pair[0].(type) {
			case *LoxString:
				builder.WriteString(key.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			builder.WriteRune('=')
			switch value := pair[1].(type) {
			case *LoxString:
				builder.WriteString(value.str)
			default:
				envp.Clear()
				return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
			}
			envp.Add(builder.String())
		}

		argv0 := args[0].(*LoxString).str
		pid, err := syscalls.ForkExecvpeFd(argv0, argv, envp)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return int64(pid), nil
	})
	osFunc("fsync", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if fd, ok := args[0].(int64); ok {
			err := syscalls.Fsync(int(fd))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeTypeAn(in.callToken, "fsync", "integer")
	})
	osFunc("ftruncate", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.ftruncate' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.ftruncate' must be an integer.")
		}
		fd := int(args[0].(int64))
		size := args[1].(int64)
		err := syscalls.Ftruncate(fd, size)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
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
	osFunc("getgroups", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		groups, err := syscalls.Getgroups()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		groupsList := list.NewListCap[any](int64(len(groups)))
		for _, group := range groups {
			groupsList.Add(int64(group))
		}
		return NewLoxList(groupsList), nil
	})
	osFunc("getpagesize", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(syscalls.Getpagesize()), nil
	})
	osFunc("getpid", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(os.Getpid()), nil
	})
	osFunc("getppid", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(os.Getppid()), nil
	})
	osFunc("getsid", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if pid, ok := args[0].(int64); ok {
			sid, err := syscalls.Getsid(int(pid))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return int64(sid), nil
		}
		return argMustBeTypeAn(in.callToken, "getsid", "integer")
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
	osFunc("isatty", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			fd := uintptr(num)
			return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd), nil
		}
		return argMustBeTypeAn(in.callToken, "isatty", "integer")
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
	osFunc("lchown", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.lchown' must be a string.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.lchown' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'os.lchown' must be an integer.")
		}
		if util.IsWindows() {
			return nil, loxerror.RuntimeError(in.callToken,
				"'os.lchown' is unsupported on Windows.")
		}
		link := args[0].(*LoxString).str
		uid := int(args[1].(int64))
		gid := int(args[2].(int64))
		err := os.Lchown(link, uid, gid)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
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
	osFunc("mkdirp", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := os.MkdirAll(loxStr.str, 0777)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "mkdirp", "string")
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
	osFunc("mktemp", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		dir := ""
		argsLen := len(args)
		switch argsLen {
		case 0:
		case 1:
			if loxStr, ok := args[0].(*LoxString); ok {
				dir = loxStr.str
			} else {
				return argMustBeType(in.callToken, "mktemp", "string")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		tempFile, err := os.CreateTemp(dir, "lox.tmp.")
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return &LoxFile{
			file:       tempFile,
			name:       tempFile.Name(),
			mode:       filemode.READ_WRITE,
			isBinary:   false,
			stat:       nil,
			properties: make(map[string]any),
		}, nil
	})
	osFunc("mktempBin", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		dir := ""
		argsLen := len(args)
		switch argsLen {
		case 0:
		case 1:
			if loxStr, ok := args[0].(*LoxString); ok {
				dir = loxStr.str
			} else {
				return argMustBeType(in.callToken, "mktempBin", "string")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		tempFile, err := os.CreateTemp(dir, "lox.tmp.")
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return &LoxFile{
			file:       tempFile,
			name:       tempFile.Name(),
			mode:       filemode.READ_WRITE,
			isBinary:   true,
			stat:       nil,
			properties: make(map[string]any),
		}, nil
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
		files := list.NewListCap[any](2)
		files.Add(&LoxFile{
			file:       r,
			name:       r.Name(),
			mode:       filemode.READ,
			isBinary:   false,
			stat:       nil,
			properties: make(map[string]any),
		})
		files.Add(&LoxFile{
			file:       w,
			name:       w.Name(),
			mode:       filemode.WRITE,
			isBinary:   false,
			stat:       nil,
			properties: make(map[string]any),
		})
		return NewLoxList(files), nil
	})
	osFunc("pipeBin", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		r, w, err := os.Pipe()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		files := list.NewListCap[any](2)
		files.Add(&LoxFile{
			file:       r,
			name:       r.Name(),
			mode:       filemode.READ,
			isBinary:   true,
			stat:       nil,
			properties: make(map[string]any),
		})
		files.Add(&LoxFile{
			file:       w,
			name:       w.Name(),
			mode:       filemode.WRITE,
			isBinary:   true,
			stat:       nil,
			properties: make(map[string]any),
		})
		return NewLoxList(files), nil
	})
	osFunc("read", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.read' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.read' must be an integer.")
		}

		fd := args[0].(int64)
		numBytes := args[1].(int64)
		bytes := make([]byte, numBytes)
		_, err := syscalls.Read(int(fd), bytes)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}

		buffer := EmptyLoxBufferCap(int64(len(bytes)))
		for _, element := range bytes {
			bufErr := buffer.add(int64(element))
			if bufErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, bufErr.Error())
			}
		}
		return buffer, nil
	})
	osFunc("readFile", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			bytes, err := os.ReadFile(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxStringQuote(string(bytes)), nil
		}
		return argMustBeType(in.callToken, "readFile", "string")
	})
	osFunc("readFileBin", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			bytes, err := os.ReadFile(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			loxBuffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, element := range bytes {
				addErr := loxBuffer.add(int64(element))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
			return loxBuffer, nil
		}
		return argMustBeType(in.callToken, "readFileBin", "string")
	})
	osFunc("readLink", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			dest, err := os.Readlink(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxStringQuote(dest), nil
		}
		return argMustBeType(in.callToken, "readLink", "string")
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
	osFunc("setegid", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if egid, ok := args[0].(int64); ok {
			err := syscalls.Setegid(int(egid))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeTypeAn(in.callToken, "setegid", "integer")
	})
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
	osFunc("seteuid", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if euid, ok := args[0].(int64); ok {
			err := syscalls.Seteuid(int(euid))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeTypeAn(in.callToken, "seteuid", "integer")
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
	osFunc("setgroups", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxList, ok := args[0].(*LoxList); ok {
			gids := list.NewList[int]()
			for _, element := range loxList.elements {
				switch element := element.(type) {
				case int64:
					gids.Add(int(element))
				default:
					gids.Clear()
					return nil, loxerror.RuntimeError(in.callToken,
						"Argument to 'os.setgroups' must be a list of integers.")
				}
			}
			err := syscalls.Setgroups(gids)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "setgroups", "list")
	})
	osFunc("setregid", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.setregid' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.setregid' must be an integer.")
		}
		rgid := args[0].(int64)
		egid := args[1].(int64)
		err := syscalls.Setregid(int(rgid), int(egid))
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("setresgid", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.setresgid' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.setresgid' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'os.setresgid' must be an integer.")
		}
		rgid := args[0].(int64)
		egid := args[1].(int64)
		sgid := args[2].(int64)
		err := linuxsyscalls.Setresgid(int(rgid), int(egid), int(sgid))
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("setresuid", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.setresuid' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.setresuid' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'os.setresuid' must be an integer.")
		}
		ruid := args[0].(int64)
		euid := args[1].(int64)
		suid := args[2].(int64)
		err := linuxsyscalls.Setresuid(int(ruid), int(euid), int(suid))
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("setreuid", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.setreuid' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.setreuid' must be an integer.")
		}
		ruid := args[0].(int64)
		euid := args[1].(int64)
		err := syscalls.Setreuid(int(ruid), int(euid))
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("setsid", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		pid, err := syscalls.Setsid()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return int64(pid), nil
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
	osFunc("sync", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		if util.IsWindows() {
			return nil, loxerror.RuntimeError(in.callToken,
				"'os.sync' is unsupported on Windows.")
		}
		syscalls.Sync()
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
	osFunc("tee", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.tee' must be a string.")
		}

		elementStr := getResult(args[0], args[0], true)
		path := args[1].(*LoxString).str
		file, openErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if openErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, openErr.Error())
		}
		defer file.Close()

		_, writeErr := file.WriteString(elementStr)
		if writeErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, writeErr.Error())
		}
		_, writeErr = os.Stdout.WriteString(elementStr)
		if writeErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, writeErr.Error())
		}
		if util.StdinFromTerminal() && []rune(elementStr)[len(elementStr)-1] != '\n' {
			fmt.Println()
		}
		return nil, nil
	})
	osFunc("teeAppend", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.teeAppend' must be a string.")
		}

		elementStr := getResult(args[0], args[0], true)
		path := args[1].(*LoxString).str
		file, openErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if openErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, openErr.Error())
		}
		defer file.Close()

		_, writeErr := file.WriteString(elementStr)
		if writeErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, writeErr.Error())
		}
		_, writeErr = os.Stdout.WriteString(elementStr)
		if writeErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, writeErr.Error())
		}
		if util.StdinFromTerminal() && []rune(elementStr)[len(elementStr)-1] != '\n' {
			fmt.Println()
		}
		return nil, nil
	})
	osFunc("tempdir", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxStringQuote(os.TempDir()), nil
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
	osFunc("truncate", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.truncate' must be a string.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.truncate' must be an integer.")
		}
		path := args[0].(*LoxString).str
		size := args[1].(int64)
		err := os.Truncate(path, size)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("umask", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if mask, ok := args[0].(int64); ok {
			if util.IsWindows() {
				return nil, loxerror.RuntimeError(in.callToken,
					"'os.umask' is unsupported on Windows.")
			}
			return int64(syscalls.Umask(int(mask))), nil
		}
		return argMustBeTypeAn(in.callToken, "umask", "integer")
	})
	osFunc("uname", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		result, err := syscalls.Uname()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		dict := EmptyLoxDict()
		setDict := func(key string, value string) {
			dict.setKeyValue(NewLoxString(key, '\''), NewLoxStringQuote(value))
		}
		setDict("sysname", result.Sysname)
		setDict("nodename", result.Nodename)
		setDict("release", result.Release)
		setDict("version", result.Version)
		setDict("machine", result.Machine)
		return dict, nil
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
			buffer := EmptyLoxBufferCap(numBytes)
			for i := int64(0); i < numBytes; i++ {
				numBig, numErr := crand.Int(crand.Reader, bigint.TwoFiveSix)
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
	osFunc("userCacheDir", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxStringQuote(cacheDir), nil
	})
	osFunc("userConfigDir", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxStringQuote(configDir), nil
	})
	osFunc("userHomeDir", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxStringQuote(homeDir), nil
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
	osFunc("wait", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		pid, waitStatus, err := syscalls.Wait()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		l := list.NewListCap[any](2)
		l.Add(int64(pid))
		l.Add(NewLoxWaitStatus(waitStatus))
		return NewLoxList(l), nil
	})
	osFunc("write", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.write' must be an integer.")
		}
		if _, ok := args[1].(*LoxBuffer); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.write' must be a buffer.")
		}

		fd := args[0].(int64)
		buffer := args[1].(*LoxBuffer)
		bytes := list.NewListCap[byte](int64(len(buffer.elements)))
		for _, element := range buffer.elements {
			bytes.Add(byte(element.(int64)))
		}

		numBytesWritten, err := syscalls.Write(int(fd), bytes)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return int64(numBytesWritten), nil
	})
	osFunc("writeFile", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.writeFile' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.writeFile' must be a string.")
		}
		name := args[0].(*LoxString).str
		data := args[1].(*LoxString).str
		err := os.WriteFile(name, []byte(data), 0666)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	osFunc("writeFileBin", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'os.writeFileBin' must be a string.")
		}
		if _, ok := args[1].(*LoxBuffer); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'os.writeFileBin' must be a buffer.")
		}

		name := args[0].(*LoxString).str
		buffer := args[1].(*LoxBuffer)
		data := make([]byte, 0, len(buffer.elements))
		for _, element := range buffer.elements {
			data = append(data, byte(element.(int64)))
		}

		err := os.WriteFile(name, data, 0666)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})

	i.globals.Define(className, osClass)
}
