package ast

import (
	"fmt"
	"os"
	"strings"

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

type ProtoIteratorErr struct {
	hasNextMethod    func() (bool, error)
	nextMethod       func() (any, error)
	legacyPanicOnErr bool
}

func (l ProtoIteratorErr) HasNext() bool {
	result, err := l.hasNextMethod()
	if err != nil {
		if l.legacyPanicOnErr {
			panic(err)
		}
		fmt.Fprintln(os.Stderr, err.Error())
	}
	return result
}

func (l ProtoIteratorErr) HasNextErr() (bool, error) {
	return l.hasNextMethod()
}

func (l ProtoIteratorErr) Next() any {
	result, err := l.nextMethod()
	if err != nil {
		if l.legacyPanicOnErr {
			panic(err)
		}
		fmt.Fprintln(os.Stderr, err.Error())
	}
	return result
}

func (l ProtoIteratorErr) NextErr() (any, error) {
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

type InfiniteIteratorErr struct {
	nextMethod       func() (any, error)
	legacyPanicOnErr bool
}

func (l InfiniteIteratorErr) HasNext() bool {
	return true
}

func (l InfiniteIteratorErr) HasNextErr() (bool, error) {
	return true, nil
}

func (l InfiniteIteratorErr) Next() any {
	result, err := l.nextMethod()
	if err != nil {
		if l.legacyPanicOnErr {
			panic(err)
		}
		fmt.Fprintln(os.Stderr, err.Error())
	}
	return result
}

func (l InfiniteIteratorErr) NextErr() (any, error) {
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
	tag      string
	methods  map[string]*struct{ ProtoLoxCallable }
}

func NewLoxIterator(iterator interfaces.Iterator) *LoxIterator {
	return NewLoxIteratorTag(iterator, "")
}

func NewLoxIteratorTag(iterator interfaces.Iterator, tag string) *LoxIterator {
	return &LoxIterator{
		iterator: iterator,
		tag:      tag,
		methods:  make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func EmptyLoxIterator() *LoxIterator {
	return NewLoxIterator(EmptyIterator{})
}

func (l *LoxIterator) assertIterErr() interfaces.IteratorErr {
	return assertIterErr(l.iterator)
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
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'iterator.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'iterator.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	formatErr := func(err error) string {
		return strings.ReplaceAll(err.Error(), "\n", " ")
	}
	switch methodName {
	case "hasNext":
		return iteratorFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if iterWithErr := l.assertIterErr(); iterWithErr != nil {
				result, hasNextErr := iterWithErr.HasNextErr()
				if hasNextErr != nil {
					return nil, loxerror.RuntimeError(name,
						"iterator.hasNext error: "+formatErr(hasNextErr))
				}
				return result, nil
			}
			return l.HasNext(), nil
		})
	case "isEmptyType":
		return iteratorFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			switch l.iterator.(type) {
			case EmptyIterator:
				return true, nil
			}
			return false, nil
		})
	case "next":
		return iteratorFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if iterWithErr := l.assertIterErr(); iterWithErr != nil {
				ok, hasNextErr := iterWithErr.HasNextErr()
				if hasNextErr != nil {
					return nil, loxerror.RuntimeError(name,
						"iterator.next error: "+formatErr(hasNextErr))
				}
				if !ok {
					return nil, loxerror.RuntimeError(name, "StopIteration")
				}
				result, nextErr := iterWithErr.NextErr()
				if nextErr != nil {
					return nil, loxerror.RuntimeError(name,
						"iterator.next error: "+formatErr(nextErr))
				}
				return result, nil
			}
			if !l.HasNext() {
				return nil, loxerror.RuntimeError(name, "StopIteration")
			}
			return l.Next(), nil
		})
	case "tag":
		return iteratorFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.tag), nil
		})
	case "tagIs":
		return iteratorFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return l.tag == loxStr.str, nil
			}
			return argMustBeType("string")
		})
	case "take":
		return iteratorFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if length, ok := args[0].(int64); ok {
				if length < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'iterator.take' cannot be negative.")
				}
				newList := list.NewListCap[any](length)
				if iterWithErr := l.assertIterErr(); iterWithErr != nil {
					for i := int64(0); i < length; i++ {
						ok, hasNextErr := iterWithErr.HasNextErr()
						if hasNextErr != nil {
							fmt.Fprintf(
								os.Stderr,
								"iterator.take error: %v\n",
								formatErr(hasNextErr),
							)
							continue
						}
						if !ok {
							break
						}
						result, nextErr := iterWithErr.NextErr()
						if nextErr != nil {
							fmt.Fprintf(
								os.Stderr,
								"iterator.take error: %v\n",
								formatErr(nextErr),
							)
							continue
						}
						newList.Add(result)
					}
				} else {
					for i := int64(0); i < length; i++ {
						if !l.HasNext() {
							break
						}
						newList.Add(l.Next())
					}
				}
				return NewLoxList(newList), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "toList":
		return iteratorFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 0:
				newList := list.NewList[any]()
				if iterWithErr := l.assertIterErr(); iterWithErr != nil {
					for {
						ok, hasNextErr := iterWithErr.HasNextErr()
						if hasNextErr != nil {
							return nil, loxerror.RuntimeError(name,
								"iterator.toList error: "+formatErr(hasNextErr))
						}
						if !ok {
							break
						}
						result, nextErr := iterWithErr.NextErr()
						if nextErr != nil {
							return nil, loxerror.RuntimeError(name,
								"iterator.toList error: "+formatErr(nextErr))
						}
						newList.Add(result)
					}
				} else {
					for l.HasNext() {
						newList.Add(l.Next())
					}
				}
				return NewLoxList(newList), nil
			case 1:
				if length, ok := args[0].(int64); ok {
					if length < 0 {
						return nil, loxerror.RuntimeError(name,
							"Argument to 'iterator.toList' cannot be negative.")
					}
					newList := list.NewListCap[any](length)
					if iterWithErr := l.assertIterErr(); iterWithErr != nil {
						for i := int64(0); i < length; i++ {
							ok, hasNextErr := iterWithErr.HasNextErr()
							if hasNextErr != nil {
								return nil, loxerror.RuntimeError(name,
									"iterator.toList error: "+formatErr(hasNextErr))
							}
							if !ok {
								break
							}
							result, nextErr := iterWithErr.NextErr()
							if nextErr != nil {
								return nil, loxerror.RuntimeError(name,
									"iterator.toList error: "+formatErr(nextErr))
							}
							newList.Add(result)
						}
					} else {
						for i := int64(0); i < length; i++ {
							if !l.HasNext() {
								break
							}
							newList.Add(l.Next())
						}
					}
					return NewLoxList(newList), nil
				}
				return argMustBeTypeAn("integer")
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
		})
	}
	return nil, loxerror.RuntimeError(name, "Iterators have no property called '"+methodName+"'.")
}

func (l *LoxIterator) HasNext() bool {
	return l.iterator.HasNext()
}

func (l *LoxIterator) HasNextErr() (bool, error) {
	if iterWithErr := l.assertIterErr(); iterWithErr != nil {
		return iterWithErr.HasNextErr()
	}
	return l.iterator.HasNext(), nil
}

func (l *LoxIterator) Next() any {
	return l.iterator.Next()
}

func (l *LoxIterator) NextErr() (any, error) {
	if iterWithErr := l.assertIterErr(); iterWithErr != nil {
		return iterWithErr.NextErr()
	}
	return l.iterator.Next(), nil
}

func (l *LoxIterator) Iterator() interfaces.Iterator {
	return l
}

func (l *LoxIterator) IteratorErr() interfaces.IteratorErr {
	return l
}

func (l *LoxIterator) String() string {
	if l.tag != "" {
		return fmt.Sprintf("<%v iterator object at %p>", l.tag, l)
	}
	return fmt.Sprintf("<iterator object at %p>", l)
}

func (l *LoxIterator) Type() string {
	return "iterator"
}
