package ast

import "fmt"

func CanBeKeyCheck(key any) (bool, string) {
	switch key := key.(type) {
	case *LoxDict, *LoxList:
		return false, fmt.Sprintf("Unhashable type '%v'.", getType(key))
	}
	return true, ""
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

func (l *LoxDict) String() string {
	return getResult(l, true)
}
