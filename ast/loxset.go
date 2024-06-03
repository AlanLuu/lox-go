package ast

import (
	"fmt"
	"reflect"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func CanBeSetElementCheck(element any) (bool, string) {
	switch element := element.(type) {
	case *LoxDict, *LoxList, *LoxSet:
		return false, fmt.Sprintf("Type '%v' cannot be used as set element.", getType(element))
	}
	return true, ""
}

type LoxSet struct {
	elements map[any]bool
	methods  map[string]*struct{ ProtoLoxCallable }
}

func NewLoxSet(elements map[any]bool) *LoxSet {
	return &LoxSet{
		elements: elements,
		methods:  make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func EmptyLoxSet() *LoxSet {
	return NewLoxSet(make(map[any]bool))
}

func (l *LoxSet) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxSet:
		return reflect.DeepEqual(l.elements, obj.elements)
	default:
		return false
	}
}

func (l *LoxSet) Get(name token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	setFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native set fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'set.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "add":
		return setFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			ok, errStr := l.add(args[0])
			if len(errStr) > 0 {
				return nil, loxerror.RuntimeError(name, errStr)
			}
			return ok, nil
		})
	case "clear":
		return setFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			for element := range l.elements {
				delete(l.elements, element)
			}
			return nil, nil
		})
	case "contains":
		return setFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			ok, errStr := CanBeSetElementCheck(args[0])
			if !ok {
				return nil, loxerror.RuntimeError(name, errStr)
			}
			return l.contains(args[0]), nil
		})
	case "copy":
		return setFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newSet := EmptyLoxSet()
			for element := range l.elements {
				newSet.add(element)
			}
			return newSet, nil
		})
	case "isDisjoint":
		return setFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if set, ok := args[0].(*LoxSet); ok {
				return l.isDisjoint(set), nil
			}
			return argMustBeType("set")
		})
	case "isEmpty":
		return setFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isEmpty(), nil
		})
	case "remove":
		return setFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			ok, errStr := CanBeSetElementCheck(args[0])
			if !ok {
				return nil, loxerror.RuntimeError(name, errStr)
			}
			return l.remove(args[0]), nil
		})
	case "toList":
		return setFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newList := list.NewList[any]()
			for element := range l.elements {
				newList.Add(element)
			}
			return NewLoxList(newList), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Sets have no property called '"+methodName+"'.")
}

func (l *LoxSet) add(element any) (bool, string) {
	var theElement any
	switch element := element.(type) {
	case *LoxString:
		theElement = LoxStringStr{element.str, element.quote}
	default:
		canBeElement, elementErr := CanBeSetElementCheck(element)
		if !canBeElement {
			return false, elementErr
		}
		theElement = element
	}
	if l.elements[theElement] {
		return false, ""
	}
	l.elements[theElement] = true
	return true, ""
}

func (l *LoxSet) contains(element any) bool {
	var theElement any
	switch element := element.(type) {
	case *LoxString:
		theElement = LoxStringStr{element.str, element.quote}
	default:
		theElement = element
	}
	return l.elements[theElement]
}

func (l *LoxSet) difference(other *LoxSet) *LoxSet {
	newSet := EmptyLoxSet()
	for element := range l.elements {
		if !other.elements[element] {
			newSet.add(element)
		}
	}
	return newSet
}

func (l *LoxSet) intersection(other *LoxSet) *LoxSet {
	newSet := EmptyLoxSet()
	for element := range l.elements {
		if other.elements[element] {
			newSet.add(element)
		}
	}
	return newSet
}

func (l *LoxSet) isDisjoint(other *LoxSet) bool {
	for element := range l.elements {
		if other.elements[element] {
			return false
		}
	}
	return true
}

func (l *LoxSet) isEmpty() bool {
	return len(l.elements) == 0
}

func (l *LoxSet) isSubset(other *LoxSet) bool {
	for element := range l.elements {
		if !other.elements[element] {
			return false
		}
	}
	return true
}

func (l *LoxSet) isProperSubset(other *LoxSet) bool {
	return l.isSubset(other) && !l.isSuperset(other)
}

func (l *LoxSet) isSuperset(other *LoxSet) bool {
	for element := range other.elements {
		if !l.elements[element] {
			return false
		}
	}
	return true
}

func (l *LoxSet) isProperSuperset(other *LoxSet) bool {
	return l.isSuperset(other) && !l.isSubset(other)
}

func (l *LoxSet) remove(element any) bool {
	if l.elements[element] {
		delete(l.elements, element)
		return true
	}
	return false
}

func (l *LoxSet) symmetricDifference(other *LoxSet) *LoxSet {
	newSet := EmptyLoxSet()
	for element := range l.elements {
		if !other.elements[element] {
			newSet.add(element)
		}
	}
	for element := range other.elements {
		if !l.elements[element] {
			newSet.add(element)
		}
	}
	return newSet
}

func (l *LoxSet) union(other *LoxSet) *LoxSet {
	newSet := EmptyLoxSet()
	for element := range l.elements {
		newSet.add(element)
	}
	for element := range other.elements {
		newSet.add(element)
	}
	return newSet
}

func (l *LoxSet) Length() int64 {
	return int64(len(l.elements))
}

func (l *LoxSet) String() string {
	return getResult(l, l, true)
}
