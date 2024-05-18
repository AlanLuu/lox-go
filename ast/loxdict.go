package ast

import (
	"fmt"
	"reflect"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func CanBeKeyCheck(key any) (bool, string) {
	switch key := key.(type) {
	case *LoxDict, *LoxList:
		return false, fmt.Sprintf("Unhashable type '%v'.", getType(key))
	}
	return true, ""
}

func UnknownKey(key any) string {
	return fmt.Sprintf("Unknown key '%v'.", key)
}

type LoxDictString struct {
	str   string
	quote byte
}

type LoxDict struct {
	entries map[any]any
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxDict(entries map[any]any) *LoxDict {
	return &LoxDict{
		entries: entries,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func EmptyLoxDict() *LoxDict {
	return NewLoxDict(make(map[any]any))
}

func (l *LoxDict) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxDict:
		return reflect.DeepEqual(l.entries, obj.entries)
	default:
		return false
	}
}

func (l *LoxDict) Get(name token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	dictFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native dictionary fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	switch methodName {
	case "clear":
		return dictFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			for key := range l.entries {
				delete(l.entries, key)
			}
			return nil, nil
		})
	case "containsKey":
		return dictFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			_, ok := l.getValueByKey(args[0])
			return ok, nil
		})
	case "get":
		return dictFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 1:
				value, ok := l.getValueByKey(args[0])
				if !ok {
					return nil, nil
				}
				return value, nil
			case 2:
				value, ok := l.getValueByKey(args[0])
				if !ok {
					return args[1], nil
				}
				return value, nil
			}
			return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		})
	case "keys":
		return dictFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			keys := list.NewList[Expr]()
			for key := range l.entries {
				keys.Add(key)
			}
			return NewLoxList(keys), nil
		})
	case "removeKey":
		return dictFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			return l.removeKey(args[0]), nil
		})
	case "values":
		return dictFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			values := list.NewList[Expr]()
			for _, value := range l.entries {
				values.Add(value)
			}
			return NewLoxList(values), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Dictionaries have no property called '"+methodName+"'.")
}

func (l *LoxDict) getValueByKey(key any) (any, bool) {
	var value any
	var ok bool
	switch key := key.(type) {
	case *LoxString:
		value, ok = l.entries[LoxDictString{key.str, key.quote}]
	default:
		value, ok = l.entries[key]
	}
	return value, ok
}

func (l *LoxDict) setKeyValue(key any, value any) {
	switch key := key.(type) {
	case *LoxString:
		l.entries[LoxDictString{key.str, key.quote}] = value
	default:
		l.entries[key] = value
	}
}

func (l *LoxDict) removeKey(key any) any {
	keyItem := key
	switch key := key.(type) {
	case *LoxString:
		keyItem = LoxDictString{key.str, key.quote}
	}
	value, ok := l.entries[keyItem]
	if !ok {
		return nil
	}
	delete(l.entries, keyItem)
	return value
}

func (l *LoxDict) String() string {
	return getResult(l, true)
}
