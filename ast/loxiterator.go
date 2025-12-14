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

type LoxCustomIterator struct {
	in                *Interpreter
	callToken         *token.Token
	hasNext           LoxCallable
	next              LoxCallable
	hasNextArgs       *LoxList
	nextArgs          *LoxList
	iter              *LoxIterator
	errWhenHasNextNil bool
	errWhenNextNil    bool
	legacyPanicOnErr  bool
	properties        map[string]any
}

func NewLoxCustomIterator(
	in *Interpreter,
	hasNext LoxCallable,
	next LoxCallable,
	hasNextArgs *LoxList,
	nextArgs *LoxList,
) *LoxCustomIterator {
	if in == nil {
		panic("in NewLoxCustomIterator: interpreter is nil")
	}
	if hasNextArgs == nil {
		hasNextArgs = EmptyLoxList()
	}
	if nextArgs == nil {
		nextArgs = EmptyLoxList()
	}
	return &LoxCustomIterator{
		in:                in,
		callToken:         in.callToken,
		hasNext:           hasNext,
		next:              next,
		hasNextArgs:       hasNextArgs,
		nextArgs:          nextArgs,
		iter:              nil,
		errWhenHasNextNil: true,
		errWhenNextNil:    true,
		legacyPanicOnErr:  true,
		properties:        make(map[string]any),
	}
}

func NewLoxCustomIteratorEmpty(in *Interpreter) *LoxCustomIterator {
	return NewLoxCustomIterator(in, nil, nil, nil, nil)
}

func (l *LoxCustomIterator) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if field, ok := l.properties[lexemeName]; ok {
		return field, nil
	}
	customIteratorFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native custom iterator fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'custom iterator.%v' must be a %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch lexemeName {
	case "errWhenHasNextNil":
		return l.errWhenHasNextNil, nil
	case "errWhenNextNil":
		return l.errWhenNextNil, nil
	case "hasNext":
		if l.hasNext == nil {
			return nil, nil
		}
		return l.hasNext, nil
	case "hasNextArgs":
		if l.hasNextArgs == nil {
			return nil, nil
		}
		return l.hasNextArgs, nil
	case "iter":
		if l.iter == nil {
			l.iter = NewLoxIterator(l)
		}
		return l.iter, nil
	case "next":
		if l.next == nil {
			return nil, nil
		}
		return l.next, nil
	case "nextArgs":
		if l.nextArgs == nil {
			return nil, nil
		}
		return l.nextArgs, nil
	case "setErrWhenHasNextNil":
		return customIteratorFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if errWhenHasNextNil, ok := args[0].(bool); ok {
				l.errWhenHasNextNil = errWhenHasNextNil
				return l, nil
			}
			return argMustBeType("boolean")
		})
	case "setErrWhenNextNil":
		return customIteratorFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if errWhenNextNil, ok := args[0].(bool); ok {
				l.errWhenNextNil = errWhenNextNil
				return l, nil
			}
			return argMustBeType("boolean")
		})
	case "setHasNext":
		return customIteratorFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				l.hasNext = callback
				return l, nil
			}
			return argMustBeType("function")
		})
	case "setHasNextArgs":
		return customIteratorFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if hasNextArgs, ok := args[0].(*LoxList); ok {
				l.hasNextArgs = hasNextArgs
				return l, nil
			}
			return argMustBeType("list")
		})
	case "setNext":
		return customIteratorFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				l.next = callback
				return l, nil
			}
			return argMustBeType("function")
		})
	case "setNextArgs":
		return customIteratorFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if nextArgs, ok := args[0].(*LoxList); ok {
				l.nextArgs = nextArgs
				return l, nil
			}
			return argMustBeType("list")
		})
	case "setNoErr":
		return customIteratorFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 0:
				l.errWhenHasNextNil = false
				l.errWhenNextNil = false
			case 1:
				if b, ok := args[0].(bool); ok {
					l.errWhenHasNextNil = !b
					l.errWhenNextNil = !b
				} else {
					return argMustBeType("boolean")
				}
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
			return l, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Custom iterators have no property called '"+lexemeName+"'.")
}

func (l *LoxCustomIterator) HasNext() bool {
	result, err := l.HasNextErr()
	if err != nil {
		if l.legacyPanicOnErr {
			panic(err)
		}
		fmt.Fprintln(os.Stderr, err.Error())
	}
	return result
}

func (l *LoxCustomIterator) HasNextErr() (bool, error) {
	if l.hasNext == nil {
		if !l.errWhenHasNextNil {
			return false, nil
		}
		const errStr = "Cannot iterate over custom iterator: hasNext is nil"
		if l.callToken == nil {
			return false, loxerror.Error(errStr)
		}
		return false, loxerror.RuntimeError(l.callToken, errStr)
	}
	if l.next == nil && l.errWhenNextNil {
		const errStr = "Cannot iterate over custom iterator: next is nil"
		if l.callToken == nil {
			return false, loxerror.Error(errStr)
		}
		return false, loxerror.RuntimeError(l.callToken, errStr)
	}
	numArgs := len(l.hasNextArgs.elements)
	callbackArity := l.hasNext.arity()
	if callbackArity > numArgs {
		n := callbackArity - numArgs
		defer func() {
			newLen := len(l.hasNextArgs.elements)
			l.hasNextArgs.elements = l.hasNextArgs.elements[:newLen-n]
		}()
		for i := 0; i < n; i++ {
			l.hasNextArgs.elements.Add(nil)
		}
	}
	result, resultErr := l.hasNext.call(l.in, l.hasNextArgs.elements)
	if resultReturn, ok := result.(Return); ok {
		result = resultReturn.FinalValue
	} else if resultErr != nil {
		return false, resultErr
	}
	return l.in.isTruthy(result), nil
}

func (l *LoxCustomIterator) Next() any {
	result, err := l.NextErr()
	if err != nil {
		if l.legacyPanicOnErr {
			panic(err)
		}
		fmt.Fprintln(os.Stderr, err.Error())
	}
	return result
}

func (l *LoxCustomIterator) NextErr() (any, error) {
	if l.next == nil {
		if !l.errWhenNextNil {
			return nil, nil
		}
		const errStr = "Cannot iterate over custom iterator: next is nil"
		if l.callToken == nil {
			return nil, loxerror.Error(errStr)
		}
		return nil, loxerror.RuntimeError(l.callToken, errStr)
	}
	numArgs := len(l.nextArgs.elements)
	callbackArity := l.next.arity()
	if callbackArity > numArgs {
		n := callbackArity - numArgs
		defer func() {
			newLen := len(l.nextArgs.elements)
			l.nextArgs.elements = l.nextArgs.elements[:newLen-n]
		}()
		for i := 0; i < n; i++ {
			l.nextArgs.elements.Add(nil)
		}
	}
	result, resultErr := l.next.call(l.in, l.nextArgs.elements)
	if resultReturn, ok := result.(Return); ok {
		result = resultReturn.FinalValue
	} else if resultErr != nil {
		return nil, resultErr
	}
	return result, nil
}

func (l *LoxCustomIterator) Iterator() interfaces.Iterator {
	return l
}

func (l *LoxCustomIterator) IteratorErr() interfaces.IteratorErr {
	return l
}

func (l *LoxCustomIterator) String() string {
	return fmt.Sprintf("<custom iterator at %p>", l)
}

func (l *LoxCustomIterator) Type() string {
	return "custom iterator"
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
