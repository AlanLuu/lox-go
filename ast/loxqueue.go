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

type LoxQueueIterator struct {
	head *linkedlist.Element
}

func (l *LoxQueueIterator) HasNext() bool {
	return l.head != nil
}

func (l *LoxQueueIterator) Next() any {
	element := l.head.Value
	l.head = l.head.Next()
	return element
}

type LoxQueue struct {
	elements *linkedlist.List
	methods  map[string]*struct{ ProtoLoxCallable }
}

func NewLoxQueue() *LoxQueue {
	return &LoxQueue{
		elements: linkedlist.New(),
		methods:  make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxQueue) add(element any) {
	l.elements.PushBack(element)
}

func (l *LoxQueue) back() any {
	element := l.elements.Back()
	if element == nil {
		return nil
	}
	return element.Value
}

func (l *LoxQueue) contains(element any) bool {
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

func (l *LoxQueue) peek() any {
	element := l.elements.Front()
	if element == nil {
		return nil
	}
	return element.Value
}

func (l *LoxQueue) remove() (any, error) {
	element := l.elements.Front()
	if element == nil {
		return nil, loxerror.Error("Cannot remove from empty queue.")
	}
	return l.elements.Remove(element), nil
}

func (l *LoxQueue) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxQueue:
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

func (l *LoxQueue) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if field, ok := l.methods[methodName]; ok {
		return field, nil
	}
	queueFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native queue fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	switch methodName {
	case "add", "enqueue":
		return queueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			l.add(args[0])
			return nil, nil
		})
	case "back", "rear":
		return queueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.back(), nil
		})
	case "clear":
		return queueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.elements.Init()
			return nil, nil
		})
	case "contains":
		return queueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			return l.contains(args[0]), nil
		})
	case "dequeue", "remove":
		return queueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			element, err := l.remove()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return element, nil
		})
	case "isEmpty":
		return queueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.Length() == 0, nil
		})
	case "peek":
		return queueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.peek(), nil
		})
	case "toList":
		return queueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newList := list.NewListCapDouble[any](l.Length())
			for e := l.elements.Front(); e != nil; e = e.Next() {
				newList.Add(e.Value)
			}
			return NewLoxList(newList), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Queues have no property called '"+methodName+"'.")
}

func (l *LoxQueue) Iterator() interfaces.Iterator {
	return &LoxQueueIterator{l.elements.Front()}
}

func (l *LoxQueue) Length() int64 {
	return int64(l.elements.Len())
}

func (l *LoxQueue) String() string {
	return getResult(l, l, true)
}

func (l *LoxQueue) Type() string {
	return "queue"
}
