package ast

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/AlanLuu/lox/bignum/bigfloat"
	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

func CanBeDictKeyCheck(key any) (bool, string) {
	switch key := key.(type) {
	case *LoxBuffer, *LoxDict, *LoxList, *LoxSet:
		return false, fmt.Sprintf("Type '%v' cannot be used as dictionary key.", getType(key))
	}
	return true, ""
}

func UnknownDictKey(key any) string {
	formatStr := "Unknown key '%v'."
	switch key := key.(type) {
	case float64:
		return fmt.Sprintf(formatStr, util.FormatFloatZero(key))
	case *big.Int:
		return fmt.Sprintf(formatStr, bigint.String(key))
	case *big.Float:
		return fmt.Sprintf(formatStr, bigfloat.String(key))
	default:
		return fmt.Sprintf(formatStr, key)
	}
}

type LoxDict struct {
	entries map[any]any
	methods map[string]*struct{ ProtoLoxCallable }
}

type LoxDictIterator struct {
	pairs list.List[*LoxList]
	index int
}

func (l *LoxDictIterator) HasNext() bool {
	return l.index < len(l.pairs)
}

func (l *LoxDictIterator) Next() any {
	pair := l.pairs[l.index]
	l.index++
	return pair
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

func (l *LoxDict) Get(name *token.Token) (any, error) {
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
	case "copy":
		return dictFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newDict := NewLoxDict(make(map[any]any))
			for key, value := range l.entries {
				newDict.setKeyValue(key, value)
			}
			return newDict, nil
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
	case "isEmpty":
		return dictFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return len(l.entries) == 0, nil
		})
	case "keys":
		return dictFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			keys := list.NewList[any]()
			it := l.Iterator()
			for it.HasNext() {
				pair := it.Next().(*LoxList).elements
				keys.Add(pair[0])
			}
			return NewLoxList(keys), nil
		})
	case "removeKey":
		return dictFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			return l.removeKey(args[0]), nil
		})
	case "values":
		return dictFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			values := list.NewList[any]()
			it := l.Iterator()
			for it.HasNext() {
				pair := it.Next().(*LoxList).elements
				values.Add(pair[1])
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
	case *big.Int:
		value, ok = l.entries[NewLoxBigIntKey(key)]
	case *big.Float:
		value, ok = l.entries[NewLoxBigFloatKey(key)]
	case *LoxString:
		value, ok = l.entries[LoxStringStr{key.str, key.quote}]
	case *LoxRange:
		value, ok = l.entries[LoxRangeDictSetKey{key.start, key.stop, key.step}]
	default:
		value, ok = l.entries[key]
	}
	return value, ok
}

func (l *LoxDict) setKeyValue(key any, value any) {
	switch key := key.(type) {
	case *big.Int:
		l.entries[NewLoxBigIntKey(key)] = value
	case *big.Float:
		l.entries[NewLoxBigFloatKey(key)] = value
	case *LoxString:
		l.entries[LoxStringStr{key.str, key.quote}] = value
	case *LoxRange:
		l.entries[LoxRangeDictSetKey{key.start, key.stop, key.step}] = value
	default:
		l.entries[key] = value
	}
}

func (l *LoxDict) removeKey(key any) any {
	keyItem := key
	switch key := key.(type) {
	case *big.Int:
		keyItem = NewLoxBigIntKey(key)
	case *big.Float:
		keyItem = NewLoxBigFloatKey(key)
	case *LoxString:
		keyItem = LoxStringStr{key.str, key.quote}
	case *LoxRange:
		keyItem = LoxRangeDictSetKey{key.start, key.stop, key.step}
	}
	value, ok := l.entries[keyItem]
	if !ok {
		return nil
	}
	delete(l.entries, keyItem)
	return value
}

func (l *LoxDict) Iterator() interfaces.Iterator {
	pairs := list.NewListCap[*LoxList](int64(len(l.entries)))
	for key, value := range l.entries {
		pair := list.NewListCap[any](2)
		switch key := key.(type) {
		case LoxBigNumKey:
			pair.Add(key.getBigNum())
		case LoxStringStr:
			pair.Add(NewLoxString(key.str, key.quote))
		case LoxRangeDictSetKey:
			pair.Add(NewLoxRange(key.start, key.stop, key.step))
		default:
			pair.Add(key)
		}
		switch value := value.(type) {
		case LoxBigNumKey:
			pair.Add(value.getBigNum())
		case LoxStringStr:
			pair.Add(NewLoxString(value.str, value.quote))
		case LoxRangeDictSetKey:
			pair.Add(NewLoxRange(value.start, value.stop, value.step))
		default:
			pair.Add(value)
		}
		pairs.Add(NewLoxList(pair))
	}
	return &LoxDictIterator{pairs, 0}
}

func (l *LoxDict) Length() int64 {
	return int64(len(l.entries))
}

func (l *LoxDict) String() string {
	return getResult(l, l, true)
}

func (l *LoxDict) Type() string {
	return "dictionary"
}
