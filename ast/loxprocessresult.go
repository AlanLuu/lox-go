package ast

import (
	"fmt"
	"os"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxProcessResult struct {
	state   *os.ProcessState
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxProcessResult(state *os.ProcessState) *LoxProcessResult {
	return &LoxProcessResult{
		state:   state,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxProcessResult) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if property, ok := l.methods[methodName]; ok {
		return property, nil
	}
	processResultFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native process result fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	switch methodName {
	case "exitCode":
		return processResultFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.state.ExitCode()), nil
		})
	case "exited":
		return processResultFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.state.Exited(), nil
		})
	case "pid":
		return processResultFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.state.Pid()), nil
		})
	case "success":
		return processResultFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.state.Success(), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Process results have no property called '"+methodName+"'.")
}

func (l *LoxProcessResult) String() string {
	return fmt.Sprintf("<process result at %p>", l)
}

func (l *LoxProcessResult) Type() string {
	return "process result"
}
