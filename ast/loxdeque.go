package ast

import (
	linkedlist "container/list"
	"fmt"
	"reflect"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxDequeIterator struct {
	head *linkedlist.Element
}

func (l *LoxDequeIterator) HasNext() bool {
	return l.head != nil
}

func (l *LoxDequeIterator) Next() any {
	element := l.head.Value
	l.head = l.head.Next()
	return element
}

type LoxDeque struct {
	elements *linkedlist.List
	methods  map[string]*struct{ ProtoLoxCallable }
}

func NewLoxDeque() *LoxDeque {
	return &LoxDeque{
		elements: linkedlist.New(),
		methods:  make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxDeque) contains(element any) bool {
	for e := l.elements.Front(); e != nil; e = e.Next() {
		var condition bool
		if e.Value == element {
			condition = true
		} else if first, ok := element.(interfaces.Equatable); ok {
			condition = first.Equals(e.Value)
		} else if second, ok := e.Value.(interfaces.Equatable); ok {
			condition = second.Equals(element)
		} else {
			condition = reflect.DeepEqual(e.Value, element)
		}
		if condition {
			return true
		}
	}
	return false
}

func (l *LoxDeque) back() any {
	element := l.elements.Back()
	if element == nil {
		return nil
	}
	return element.Value
}

func (l *LoxDeque) front() any {
	element := l.elements.Front()
	if element == nil {
		return nil
	}
	return element.Value
}

func (l *LoxDeque) pushBack(element any) {
	l.elements.PushBack(element)
}

func (l *LoxDeque) pushFront(element any) {
	l.elements.PushFront(element)
}

func (l *LoxDeque) removeBack() (any, error) {
	element := l.elements.Back()
	if element == nil {
		return nil, loxerror.Error("Cannot remove from empty deque.")
	}
	return l.elements.Remove(element), nil
}

func (l *LoxDeque) removeFront() (any, error) {
	element := l.elements.Front()
	if element == nil {
		return nil, loxerror.Error("Cannot remove from empty deque.")
	}
	return l.elements.Remove(element), nil
}

func (l *LoxDeque) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxDeque:
		e1 := l.elements.Front()
		e2 := obj.elements.Front()
		for e1 != nil && e2 != nil {
			var condition bool
			if e1.Value == e2.Value {
				condition = true
			} else if first, ok := e1.Value.(interfaces.Equatable); ok {
				condition = first.Equals(e2.Value)
			} else if second, ok := e2.Value.(interfaces.Equatable); ok {
				condition = second.Equals(e1.Value)
			} else {
				condition = reflect.DeepEqual(e1.Value, e2.Value)
			}
			if !condition {
				return false
			}
			e1 = e1.Next()
			e2 = e2.Next()
		}
		return (e1 == nil) == (e2 == nil)
	default:
		return false
	}
}

func (l *LoxDeque) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if field, ok := l.methods[methodName]; ok {
		return field, nil
	}
	dequeFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native deque fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	switch methodName {
	case "back", "rear":
		return dequeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.back(), nil
		})
	case "clear":
		return dequeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.elements.Init()
			return nil, nil
		})
	case "contains":
		return dequeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			return l.contains(args[0]), nil
		})
	case "front":
		return dequeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.front(), nil
		})
	case "isEmpty":
		return dequeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.Length() == 0, nil
		})
	case "pushBack":
		return dequeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			l.pushBack(args[0])
			return nil, nil
		})
	case "pushFront":
		return dequeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			l.pushFront(args[0])
			return nil, nil
		})
	case "removeBack":
		return dequeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			element, err := l.removeBack()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return element, nil
		})
	case "removeFront":
		return dequeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			element, err := l.removeFront()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return element, nil
		})
	case "toList":
		return dequeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newList := list.NewListCapDouble[any](l.Length())
			for e := l.elements.Front(); e != nil; e = e.Next() {
				newList.Add(e.Value)
			}
			return NewLoxList(newList), nil
		})
	case "toListReversed":
		return dequeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newList := list.NewListCapDouble[any](l.Length())
			for e := l.elements.Back(); e != nil; e = e.Prev() {
				newList.Add(e.Value)
			}
			return NewLoxList(newList), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Deques have no property called '"+methodName+"'.")
}

func (l *LoxDeque) Iterator() interfaces.Iterator {
	return &LoxDequeIterator{l.elements.Front()}
}

func (l *LoxDeque) ReverseIterator() interfaces.Iterator {
	iterator := struct {
		ProtoIterator
		current *linkedlist.Element
	}{current: l.elements.Back()}
	iterator.hasNextMethod = func() bool {
		return iterator.current != nil
	}
	iterator.nextMethod = func() any {
		element := iterator.current.Value
		iterator.current = iterator.current.Prev()
		return element
	}
	return iterator
}

func (l *LoxDeque) Length() int64 {
	return int64(l.elements.Len())
}

func (l *LoxDeque) String() string {
	return getResult(l, l, true)
}

func (l *LoxDeque) Type() string {
	return "deque"
}
