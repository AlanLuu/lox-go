package ast

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxProcessError struct {
	msg string
}

func (l LoxProcessError) Error() string {
	return l.msg
}

type LoxProcessOptions struct {
	reusable bool
}

type LoxProcess struct {
	process        *exec.Cmd
	originalStdin  io.Reader
	originalStdout io.Writer
	originalStderr io.Writer
	cmdArgStr      string
	reusable       bool
	started        bool
	waited         bool
	methods        map[string]*struct{ ProtoLoxCallable }
}

func NewLoxProcessFields(process *exec.Cmd, options LoxProcessOptions) *LoxProcess {
	return &LoxProcess{
		process:        process,
		originalStdin:  process.Stdin,
		originalStdout: process.Stdout,
		originalStderr: process.Stderr,
		cmdArgStr:      "[" + strings.Join(process.Args, ", ") + "]",
		reusable:       options.reusable,
		started:        false,
		waited:         false,
		methods:        make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxProcess(process *exec.Cmd) *LoxProcess {
	return NewLoxProcessFields(process, LoxProcessOptions{
		reusable: false,
	})
}

func NewLoxProcessReusable(process *exec.Cmd) *LoxProcess {
	return NewLoxProcessFields(process, LoxProcessOptions{
		reusable: true,
	})
}

func (l *LoxProcess) combinedOutput() ([]byte, error) {
	if l.started && !l.reusable {
		return nil, LoxProcessError{"Cannot run process that has already been executed."}
	}
	if l.process.Stdout != nil {
		return nil, LoxProcessError{"Cannot get output of process with stdout already set."}
	}
	if l.process.Stderr != nil {
		return nil, LoxProcessError{"Cannot get output of process with stderr already set."}
	}
	l.started = true
	l.waited = true
	return l.process.CombinedOutput()
}

func (l *LoxProcess) kill() error {
	if !l.started {
		return LoxProcessError{"Cannot kill process that is not executing."}
	}
	if l.waited {
		return LoxProcessError{"Cannot kill process that has already been waited on."}
	}
	err := l.process.Process.Kill()
	if err == nil {
		l.wait()
	}
	return err
}

func (l *LoxProcess) output() ([]byte, error) {
	if l.started && !l.reusable {
		return nil, LoxProcessError{"Cannot run process that has already been executed."}
	}
	if l.process.Stdout != nil {
		return nil, LoxProcessError{"Cannot get output of process with stdout already set."}
	}
	l.started = true
	l.waited = true
	return l.process.Output()
}

func (l *LoxProcess) resetProcessCmd() {
	var process *exec.Cmd
	switch len(l.process.Args) {
	case 0: //Should never happen
		panic("in resetProcessCmd: args slice is empty")
	case 1:
		process = exec.Command(l.process.Args[0])
	default:
		process = exec.Command(l.process.Args[0], l.process.Args[1:]...)
	}
	process.Stdin = l.originalStdin
	process.Stdout = l.originalStdout
	process.Stderr = l.originalStderr
	l.process = process
	l.started = false
	l.waited = false
}

func (l *LoxProcess) run() error {
	if l.started && !l.reusable {
		return LoxProcessError{"Cannot run process that has already been executed."}
	}
	l.started = true
	l.waited = true
	return l.process.Run()
}

func (l *LoxProcess) setStderr(writer io.Writer) {
	l.process.Stderr = writer
	l.originalStderr = writer
}

func (l *LoxProcess) setStdin(reader io.Reader) {
	l.process.Stdin = reader
	l.originalStdin = reader
}

func (l *LoxProcess) setStdout(writer io.Writer) {
	l.process.Stdout = writer
	l.originalStdout = writer
}

func (l *LoxProcess) start() error {
	if l.started {
		if !l.reusable && l.waited {
			return LoxProcessError{"Cannot start process that has already been executed."}
		} else {
			return LoxProcessError{"Cannot start process that is already executing."}
		}
	}
	l.started = true
	return l.process.Start()
}

func (l *LoxProcess) wait() error {
	if !l.started {
		return LoxProcessError{"Cannot wait on process that is not executing."}
	}
	if l.waited {
		return LoxProcessError{"Cannot wait on process that has already been waited on."}
	}
	l.waited = true
	return l.process.Wait()
}

func (l *LoxProcess) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if property, ok := l.methods[methodName]; ok {
		return property, nil
	}
	processFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native process fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'process.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "args":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			argsList := list.NewListCap[any](int64(len(l.process.Args)))
			for _, arg := range l.process.Args {
				argsList.Add(NewLoxStringQuote(arg))
			}
			return NewLoxList(argsList), nil
		})
	case "dir":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.process.Dir), nil
		})
	case "combinedOutput":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			output, err := l.combinedOutput()
			resetStd := true
			if l.reusable {
				defer func() {
					if resetStd {
						l.resetProcessCmd()
					}
				}()
			}
			if err != nil {
				switch err := err.(type) {
				case *exec.ExitError:
					if !strings.Contains(err.Error(), "exit status") {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				default:
					if _, ok := err.(LoxProcessError); ok {
						resetStd = false
					}
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}
			return NewLoxStringQuote(string(output)), nil
		})
	case "combinedOutputBuf":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			output, err := l.combinedOutput()
			resetStd := true
			if l.reusable {
				defer func() {
					if resetStd {
						l.resetProcessCmd()
					}
				}()
			}
			if err != nil {
				switch err := err.(type) {
				case *exec.ExitError:
					if !strings.Contains(err.Error(), "exit status") {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				default:
					if _, ok := err.(LoxProcessError); ok {
						resetStd = false
					}
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}
			buffer := EmptyLoxBufferCapDouble(int64(len(output)))
			for _, element := range output {
				addErr := buffer.add(int64(element))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "isReusable":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.reusable, nil
		})
	case "isRunning":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.started && !l.waited, nil
		})
	case "kill":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.kill()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxProcessResult(l.process.ProcessState), nil
		})
	case "output":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			output, err := l.output()
			resetStd := true
			if l.reusable {
				defer func() {
					if resetStd {
						l.resetProcessCmd()
					}
				}()
			}
			if err != nil {
				switch err := err.(type) {
				case *exec.ExitError:
					if !strings.Contains(err.Error(), "exit status") {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				default:
					if _, ok := err.(LoxProcessError); ok {
						resetStd = false
					}
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}
			return NewLoxStringQuote(string(output)), nil
		})
	case "outputBuf":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			output, err := l.output()
			resetStd := true
			if l.reusable {
				defer func() {
					if resetStd {
						l.resetProcessCmd()
					}
				}()
			}
			if err != nil {
				switch err := err.(type) {
				case *exec.ExitError:
					if !strings.Contains(err.Error(), "exit status") {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				default:
					if _, ok := err.(LoxProcessError); ok {
						resetStd = false
					}
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}
			buffer := EmptyLoxBufferCapDouble(int64(len(output)))
			for _, element := range output {
				addErr := buffer.add(int64(element))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "path":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.process.Path), nil
		})
	case "run":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.reusable {
				defer l.resetProcessCmd()
			}
			if err := l.run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					return NewLoxProcessResult(exitErr.ProcessState), nil
				} else {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}
			return NewLoxProcessResult(l.process.ProcessState), nil
		})
	case "setArgs":
		return processFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxList, ok := args[0].(*LoxList); ok {
				args := make([]string, 0, len(loxList.elements))
				for _, element := range loxList.elements {
					switch element := element.(type) {
					case *LoxString:
						args = append(args, element.str)
					default:
						args = nil
						return nil, loxerror.RuntimeError(name,
							"List argument to 'process.setArgs' must only have strings.")
					}
				}
				l.process.Args = args
				return l, nil
			}
			return argMustBeType("list")
		})
	case "setDir":
		return processFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.process.Dir = loxStr.str
				return l, nil
			}
			return argMustBeType("string")
		})
	case "setPath":
		return processFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.process.Path = loxStr.str
				return l, nil
			}
			return argMustBeType("string")
		})
	case "setReusable":
		return processFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if reusable, ok := args[0].(bool); ok {
				if reusable && l.started {
					l.resetProcessCmd()
				}
				l.reusable = reusable
				return l, nil
			}
			return argMustBeType("boolean")
		})
	case "setStderr":
		return processFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxFile, ok := args[0].(*LoxFile); ok {
				l.setStderr(loxFile.file)
				return l, nil
			}
			return argMustBeType("file")
		})
	case "setStdin":
		return processFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxFile, ok := args[0].(*LoxFile); ok {
				l.setStdin(loxFile.file)
				return l, nil
			}
			return argMustBeType("file")
		})
	case "setStdout":
		return processFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxFile, ok := args[0].(*LoxFile); ok {
				l.setStdout(loxFile.file)
				return l, nil
			}
			return argMustBeType("file")
		})
	case "start":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.start(); err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "started":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.started, nil
		})
	case "wait":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.reusable {
				defer l.resetProcessCmd()
			}
			if err := l.wait(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					return NewLoxProcessResult(exitErr.ProcessState), nil
				} else {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}
			return NewLoxProcessResult(l.process.ProcessState), nil
		})
	case "waited":
		return processFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.waited, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Processes have no property called '"+methodName+"'.")
}

func (l *LoxProcess) String() string {
	return fmt.Sprintf("<process cmd=\"%v\" at %p", l.cmdArgStr, l)
}

func (l *LoxProcess) Type() string {
	return "process"
}
