package ast

import (
	"fmt"
	"net"
	"os"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type LoxListener struct {
	listener net.Listener
	isClosed bool
	methods  map[string]*struct{ ProtoLoxCallable }
}

func NewLoxListener(listener net.Listener) *LoxListener {
	if listener == nil {
		panic("in NewLoxListener: listener is nil")
	}
	return &LoxListener{
		listener: listener,
		isClosed: false,
		methods:  make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxListener) close() error {
	if !l.isClosed {
		err := l.listener.Close()
		if err != nil {
			return err
		}
		l.isClosed = true
	}
	return nil
}

func (l *LoxListener) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	listenerFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native listener object fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'listener.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "accept":
		return listenerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			conn, err := l.listener.Accept()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxConnection(conn), nil
		})
	case "acceptFunc":
		return listenerFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				if !util.UnsafeMode {
					return nil, loxerror.RuntimeError(name,
						"Cannot call 'listener.acceptFunc' in non-unsafe mode.")
				}
				var num int64 = 1
				for {
					conn, err := l.listener.Accept()
					if err != nil {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
					go func() {
						argList := getArgList(callback, 2)
						defer argList.Clear()
						argList[0] = NewLoxConnection(conn)
						argList[1] = num
						num++
						result, resultErr := callback.call(i, argList)
						if resultErr != nil && result == nil {
							errStr := loxerror.RuntimeError(
								name,
								resultErr.Error(),
							).Error()
							fmt.Fprintf(
								os.Stderr,
								"listener.acceptFunc: error in callback: %v\n",
								errStr,
							)
						}
					}()
				}
			}
			return argMustBeType("function")
		})
	case "acceptIter":
		return listenerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var conn net.Conn
			var err error
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				conn, err = l.listener.Accept()
				return err == nil
			}
			iterator.nextMethod = func() any {
				return NewLoxConnection(conn)
			}
			return NewLoxIterator(iterator), nil
		})
	case "addrNetwork":
		return listenerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			addr := l.listener.Addr()
			if addr == nil {
				return EmptyLoxString(), nil
			}
			return NewLoxStringQuote(addr.Network()), nil
		})
	case "addrString":
		return listenerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			addr := l.listener.Addr()
			if addr == nil {
				return EmptyLoxString(), nil
			}
			return NewLoxStringQuote(addr.String()), nil
		})
	case "close":
		return listenerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.close()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "closeForce":
		return listenerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.close()
			return nil, nil
		})
	case "isClosed":
		return listenerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isClosed, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Listener objects have no property called '"+methodName+"'.")
}

func (l *LoxListener) String() string {
	return fmt.Sprintf("<listener object at %p>", l)
}

func (l *LoxListener) Type() string {
	return "listener"
}
