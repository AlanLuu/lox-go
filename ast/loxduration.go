package ast

import (
	"fmt"
	"time"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxDuration struct {
	duration time.Duration
	methods  map[string]*struct{ ProtoLoxCallable }
}

func NewLoxDuration(duration time.Duration) *LoxDuration {
	return &LoxDuration{
		duration: duration,
		methods:  make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxDuration) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxDuration:
		return l.duration == obj.duration
	default:
		return false
	}
}

func (l *LoxDuration) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	durationFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native duration fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'duration.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'duration.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "abs":
		return durationFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxDuration(l.duration.Abs()), nil
		})
	case "add":
		return durationFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDuration, ok := args[0].(*LoxDuration); ok {
				return NewLoxDuration(l.duration + loxDuration.duration), nil
			}
			return argMustBeType("duration")
		})
	case "div":
		return durationFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDuration, ok := args[0].(*LoxDuration); ok {
				if loxDuration.duration == 0 {
					return nil, loxerror.RuntimeError(name,
						"Cannot divide duration by a duration of 0.")
				}
				return NewLoxDuration(l.duration / loxDuration.duration), nil
			}
			return argMustBeType("duration")
		})
	case "hours":
		return durationFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.duration.Hours(), nil
		})
	case "int":
		return durationFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.duration), nil
		})
	case "microseconds":
		return durationFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.duration.Microseconds(), nil
		})
	case "milliseconds":
		return durationFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.duration.Milliseconds(), nil
		})
	case "minutes":
		return durationFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.duration.Minutes(), nil
		})
	case "mod":
		return durationFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDuration, ok := args[0].(*LoxDuration); ok {
				if loxDuration.duration == 0 {
					return nil, loxerror.RuntimeError(name,
						"Cannot divide duration by a duration of 0.")
				}
				return NewLoxDuration(l.duration % loxDuration.duration), nil
			}
			return argMustBeType("duration")
		})
	case "mul":
		return durationFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDuration, ok := args[0].(*LoxDuration); ok {
				return NewLoxDuration(l.duration * loxDuration.duration), nil
			}
			return argMustBeType("duration")
		})
	case "nanoseconds":
		return durationFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.duration.Nanoseconds(), nil
		})
	case "round":
		return durationFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDuration, ok := args[0].(*LoxDuration); ok {
				return NewLoxDuration(l.duration.Round(loxDuration.duration)), nil
			}
			return argMustBeType("duration")
		})
	case "scale", "times":
		return durationFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if scale, ok := args[0].(int64); ok {
				return NewLoxDuration(l.duration * time.Duration(scale)), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "seconds":
		return durationFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.duration.Seconds(), nil
		})
	case "sleep":
		return durationFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			time.Sleep(l.duration)
			return nil, nil
		})
	case "string":
		return durationFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxString(l.duration.String(), '\''), nil
		})
	case "sub":
		return durationFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDuration, ok := args[0].(*LoxDuration); ok {
				return NewLoxDuration(l.duration - loxDuration.duration), nil
			}
			return argMustBeType("duration")
		})
	case "truncate":
		return durationFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDuration, ok := args[0].(*LoxDuration); ok {
				return NewLoxDuration(l.duration.Truncate(loxDuration.duration)), nil
			}
			return argMustBeType("duration")
		})
	}
	return nil, loxerror.RuntimeError(name, "Durations have no property called '"+methodName+"'.")
}

func (l *LoxDuration) String() string {
	return fmt.Sprintf("<duration: %v>", l.duration)
}

func (l *LoxDuration) Type() string {
	return "duration"
}
