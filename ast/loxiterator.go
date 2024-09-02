package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type ProtoIterator struct {
	hasNextMethod func() bool
	nextMethod    func() any
}

func (l ProtoIterator) HasNext() bool {
	return l.hasNextMethod()
}

func (l ProtoIterator) Next() any {
	return l.nextMethod()
}

type InfiniteIterator struct {
	nextMethod func() any
}

func (l InfiniteIterator) HasNext() bool {
	return true
}

func (l InfiniteIterator) Next() any {
	return l.nextMethod()
}

type EmptyIterator struct{}

func (l EmptyIterator) HasNext() bool {
	return false
}

func (l EmptyIterator) Next() any {
	return nil
}

type LoxIterator struct {
	iterator interfaces.Iterator
	methods  map[string]*struct{ ProtoLoxCallable }
}

func NewLoxIterator(iterator interfaces.Iterator) *LoxIterator {
	return &LoxIterator{
		iterator: iterator,
		methods:  make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func EmptyLoxIterator() *LoxIterator {
	return NewLoxIterator(EmptyIterator{})
}

func (l *LoxIterator) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	iteratorFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native iterator fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	switch methodName {
	case "hasNext":
		return iteratorFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.HasNext(), nil
		})
	case "next":
		return iteratorFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.HasNext() {
				return nil, loxerror.RuntimeError(name, "StopIteration")
			}
			return l.Next(), nil
		})
	case "toList":
		return iteratorFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newList := list.NewList[any]()
			for l.HasNext() {
				newList.Add(l.Next())
			}
			return NewLoxList(newList), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Iterators have no property called '"+methodName+"'.")
}

func (l *LoxIterator) HasNext() bool {
	return l.iterator.HasNext()
}

func (l *LoxIterator) Next() any {
	return l.iterator.Next()
}

func (l *LoxIterator) Iterator() interfaces.Iterator {
	return l
}

func (l *LoxIterator) String() string {
	return fmt.Sprintf("<iterator object at %p>", l)
}

func (l *LoxIterator) Type() string {
	return "iterator"
}
