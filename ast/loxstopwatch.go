package ast

import (
	"fmt"
	"time"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func TimeZeroTime() time.Time {
	return time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
}

type LoxStopwatch struct {
	startTime time.Time
	stopTime  time.Time
	started   bool
	methods   map[string]*struct{ ProtoLoxCallable }
}

func NewLoxStopwatch() *LoxStopwatch {
	zeroTime := TimeZeroTime()
	return &LoxStopwatch{
		startTime: zeroTime,
		stopTime:  zeroTime,
		started:   false,
		methods:   make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxStopwatch) currentTime() time.Duration {
	if l.started {
		return time.Since(l.startTime)
	}
	return l.stopTime.Sub(l.startTime)
}

func (l *LoxStopwatch) start() {
	if l.started {
		return
	}
	if l.isZero() {
		l.startTime = time.Now()
	} else {
		l.startTime = time.Now().Add(-l.currentTime())
	}
	l.started = true
}

func (l *LoxStopwatch) isZero() bool {
	return l.startTime.IsZero() && l.stopTime.IsZero()
}

func (l *LoxStopwatch) reset() {
	zeroTime := TimeZeroTime()
	l.startTime = zeroTime
	l.stopTime = zeroTime
	l.started = false
}

func (l *LoxStopwatch) stop() {
	if !l.started {
		return
	}
	l.stopTime = time.Now()
	l.started = false
}

func (l *LoxStopwatch) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxStopwatch:
		if l == obj {
			return true
		}
		return l.currentTime() == obj.currentTime()
	default:
		return false
	}
}

func (l *LoxStopwatch) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	stopwatchFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native stopwatch fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	switch methodName {
	case "duration":
		return stopwatchFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxDuration(l.currentTime()), nil
		})
	case "hours":
		return stopwatchFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.currentTime().Hours(), nil
		})
	case "isReset":
		return stopwatchFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isZero(), nil
		})
	case "microseconds":
		return stopwatchFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.currentTime().Microseconds(), nil
		})
	case "milliseconds":
		return stopwatchFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.currentTime().Milliseconds(), nil
		})
	case "minutes":
		return stopwatchFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.currentTime().Minutes(), nil
		})
	case "reset":
		return stopwatchFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.reset()
			return nil, nil
		})
	case "seconds":
		return stopwatchFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.currentTime().Seconds(), nil
		})
	case "start":
		return stopwatchFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.start()
			return nil, nil
		})
	case "stop":
		return stopwatchFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.stop()
			return nil, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Stopwatches have no property called '"+methodName+"'.")
}

func (l *LoxStopwatch) String() string {
	return fmt.Sprintf("<stopwatch: %v>", l.currentTime())
}

func (l *LoxStopwatch) Type() string {
	return "stopwatch"
}
