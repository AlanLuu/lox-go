package ast

import (
	"fmt"
	"reflect"
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

func (l *LoxSet) String() string {
	return getResult(l, l, true)
}
