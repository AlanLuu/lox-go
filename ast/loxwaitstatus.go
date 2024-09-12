package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/syscalls"
	"github.com/AlanLuu/lox/token"
)

type LoxWaitStatus struct {
	status     syscalls.WaitStatus
	properties map[string]any
}

func NewLoxWaitStatus(status syscalls.WaitStatus) *LoxWaitStatus {
	return &LoxWaitStatus{
		status:     status,
		properties: make(map[string]any),
	}
}

func (l *LoxWaitStatus) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if property, ok := l.properties[lexemeName]; ok {
		return property, nil
	}
	waitStatusField := func(field any) (any, error) {
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = field
		}
		return field, nil
	}
	waitStatusFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native wait status fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = s
		}
		return s, nil
	}
	switch lexemeName {
	case "continued":
		return waitStatusFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.status.Continued(), nil
		})
	case "exited":
		return waitStatusFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.status.Exited(), nil
		})
	case "exitStatus":
		return waitStatusFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.status.ExitStatus()), nil
		})
	case "signaled":
		return waitStatusFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.status.Signaled(), nil
		})
	case "stopped":
		return waitStatusFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.status.Stopped(), nil
		})
	case "stopSignal":
		return waitStatusFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.status.StopSignal()), nil
		})
	case "waitStatus":
		return waitStatusField(l.status.WaitStatus)
	}
	return nil, loxerror.RuntimeError(name, "Wait statuses have no property called '"+lexemeName+"'.")
}

func (l *LoxWaitStatus) String() string {
	return fmt.Sprintf("<wait status: %v at %p>", l.status.WaitStatus, l)
}

func (l *LoxWaitStatus) Type() string {
	return "wait status"
}
