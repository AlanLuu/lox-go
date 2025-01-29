package ast

import (
	"fmt"
	"time"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func defineDurationFields(durationClass *LoxClass) {
	durations := map[string]time.Duration{
		"zero":        time.Duration(0),
		"nanosecond":  time.Nanosecond,
		"microsecond": time.Microsecond,
		"millisecond": time.Millisecond,
		"second":      time.Second,
		"minute":      time.Minute,
		"hour":        time.Hour,
		"day":         24 * time.Hour,
		"year":        365 * 24 * time.Hour,
	}
	for key, value := range durations {
		durationClass.classProperties[key] = NewLoxDuration(value)
	}
}

func (i *Interpreter) defineDurationFuncs() {
	className := "Duration"
	durationClass := NewLoxClass(className, nil, false)
	durationFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native Duration class fn %v at %p>", name, &s)
		}
		durationClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'Duration.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	defineDurationFields(durationClass)
	durationFunc("days", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if days, ok := args[0].(int64); ok {
			return NewLoxDuration(time.Duration(days) * (24 * time.Hour)), nil
		}
		return argMustBeType(in.callToken, "days", "integer")
	})
	durationFunc("fromInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if durationInt, ok := args[0].(int64); ok {
			return NewLoxDuration(time.Duration(durationInt)), nil
		}
		return argMustBeType(in.callToken, "fromInt", "integer")
	})
	durationFunc("hours", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if hours, ok := args[0].(int64); ok {
			return NewLoxDuration(time.Duration(hours) * time.Hour), nil
		}
		return argMustBeType(in.callToken, "hours", "integer")
	})
	durationFunc("loop", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxDuration); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Duration.loop' must be a duration.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Duration.loop' must be a function.")
		}
		loxDuration := args[0].(*LoxDuration)
		callback := args[1].(*LoxFunction)
		stopCallback := false
		callbackChan := make(chan struct{}, 1)
		errorChan := make(chan error, 1)
		argList := getArgList(callback, 0)
		go func() {
			for !stopCallback {
				result, resultErr := callback.call(i, argList)
				if resultErr != nil && result == nil {
					errorChan <- resultErr
					break
				}
			}
			callbackChan <- struct{}{}
		}()
		select {
		case err := <-errorChan:
			return nil, err
		case <-time.After(loxDuration.duration):
			stopCallback = true
			<-callbackChan
		}
		return nil, nil
	})
	durationFunc("microseconds", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if microseconds, ok := args[0].(int64); ok {
			return NewLoxDuration(time.Duration(microseconds) * time.Microsecond), nil
		}
		return argMustBeType(in.callToken, "microseconds", "integer")
	})
	durationFunc("milliseconds", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if milliseconds, ok := args[0].(int64); ok {
			return NewLoxDuration(time.Duration(milliseconds) * time.Millisecond), nil
		}
		return argMustBeType(in.callToken, "milliseconds", "integer")
	})
	durationFunc("minutes", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if minutes, ok := args[0].(int64); ok {
			return NewLoxDuration(time.Duration(minutes) * time.Minute), nil
		}
		return argMustBeType(in.callToken, "minutes", "integer")
	})
	durationFunc("parse", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			duration, err := time.ParseDuration(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxDuration(duration), nil
		}
		return argMustBeType(in.callToken, "parse", "string")
	})
	durationFunc("since", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxDate, ok := args[0].(*LoxDate); ok {
			return NewLoxDuration(time.Since(loxDate.date)), nil
		}
		return argMustBeType(in.callToken, "since", "date")
	})
	durationFunc("seconds", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if seconds, ok := args[0].(int64); ok {
			return NewLoxDuration(time.Duration(seconds) * time.Second), nil
		}
		return argMustBeType(in.callToken, "seconds", "integer")
	})
	durationFunc("sleep", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxDuration, ok := args[0].(*LoxDuration); ok {
			time.Sleep(loxDuration.duration)
			return nil, nil
		}
		return argMustBeType(in.callToken, "sleep", "duration")
	})
	durationFunc("stopwatch", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxStopwatch(), nil
	})
	durationFunc("until", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxDate, ok := args[0].(*LoxDate); ok {
			return NewLoxDuration(time.Until(loxDate.date)), nil
		}
		return argMustBeType(in.callToken, "until", "date")
	})
	durationFunc("years", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if years, ok := args[0].(int64); ok {
			return NewLoxDuration(time.Duration(years) * (365 * 24 * time.Hour)), nil
		}
		return argMustBeType(in.callToken, "years", "integer")
	})

	i.globals.Define(className, durationClass)
}
