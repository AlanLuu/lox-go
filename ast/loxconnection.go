package ast

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxConnection struct {
	conn       net.Conn
	lineReader *bufio.Reader
	isClosed   bool
	methods    map[string]*struct{ ProtoLoxCallable }
}

func NewLoxConnection(conn net.Conn) *LoxConnection {
	if conn == nil {
		panic("in NewLoxConnection: conn is nil")
	}
	return &LoxConnection{
		conn:       conn,
		lineReader: nil,
		isClosed:   false,
		methods:    make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxConnection) close() error {
	if !l.isClosed {
		err := l.conn.Close()
		if err != nil {
			return err
		}
		l.isClosed = true
	}
	return nil
}

func (l *LoxConnection) initLineReader() {
	if l.lineReader == nil {
		l.lineReader = bufio.NewReader(l.conn)
	}
}

func (l *LoxConnection) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	connFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native connection object fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'connection.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "close":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.close()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "closeForce":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.close()
			return nil, nil
		})
	case "isClosed":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isClosed, nil
		})
	case "localAddrNetwork":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			localAddr := l.conn.LocalAddr()
			if localAddr == nil {
				return EmptyLoxString(), nil
			}
			return NewLoxStringQuote(localAddr.Network()), nil
		})
	case "localAddrString":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			localAddr := l.conn.LocalAddr()
			if localAddr == nil {
				return EmptyLoxString(), nil
			}
			return NewLoxStringQuote(localAddr.String()), nil
		})
	case "readBuffer":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			bytes, err := io.ReadAll(l.conn)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, value := range bytes {
				addErr := buffer.add(int64(value))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "readBufferLine":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.initLineReader()
			bytes, err := l.lineReader.ReadBytes('\n')
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, value := range bytes {
				addErr := buffer.add(int64(value))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "readBufferLineIter":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.initLineReader()
			var bytes []byte
			var err error
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				bytes, err = l.lineReader.ReadBytes('\n')
				return err == nil
			}
			iterator.nextMethod = func() any {
				buffer := EmptyLoxBufferCap(int64(len(bytes)))
				for _, value := range bytes {
					buffer.add(int64(value))
				}
				return buffer
			}
			return NewLoxIterator(iterator), nil
		})
	case "readString":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			bytes, err := io.ReadAll(l.conn)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxStringQuote(string(bytes)), nil
		})
	case "readStringLine":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.initLineReader()
			str, err := l.lineReader.ReadString('\n')
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxStringQuote(str), nil
		})
	case "readStringLineIter":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.initLineReader()
			var str string
			var err error
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				str, err = l.lineReader.ReadString('\n')
				return err == nil
			}
			iterator.nextMethod = func() any {
				return NewLoxStringQuote(str)
			}
			return NewLoxIterator(iterator), nil
		})
	case "readToFile":
		return connFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'connection.readToFile' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return argMustBeType("file or string")
			}
			bytes, err := io.ReadAll(l.conn)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			switch arg := args[0].(type) {
			case *LoxFile:
				_, err := arg.file.Write(bytes)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			case *LoxString:
				err := os.WriteFile(arg.str, bytes, 0666)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}
			return nil, nil
		})
	case "readToFileLine":
		return connFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'connection.readToFile' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return argMustBeType("file or string")
			}
			l.initLineReader()
			bytes, err := l.lineReader.ReadBytes('\n')
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			switch arg := args[0].(type) {
			case *LoxFile:
				_, err := arg.file.Write(bytes)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			case *LoxString:
				err := os.WriteFile(arg.str, bytes, 0666)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}
			return nil, nil
		})
	case "remoteAddrNetwork":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			localAddr := l.conn.RemoteAddr()
			if localAddr == nil {
				return EmptyLoxString(), nil
			}
			return NewLoxStringQuote(localAddr.Network()), nil
		})
	case "remoteAddrString":
		return connFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			localAddr := l.conn.RemoteAddr()
			if localAddr == nil {
				return EmptyLoxString(), nil
			}
			return NewLoxStringQuote(localAddr.String()), nil
		})
	case "setDeadline":
		return connFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDate, ok := args[0].(*LoxDate); ok {
				err := l.conn.SetDeadline(loxDate.date)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return nil, nil
			}
			return argMustBeType("date")
		})
	case "setReadDeadline":
		return connFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDate, ok := args[0].(*LoxDate); ok {
				err := l.conn.SetReadDeadline(loxDate.date)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return nil, nil
			}
			return argMustBeType("date")
		})
	case "setWriteDeadline":
		return connFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDate, ok := args[0].(*LoxDate); ok {
				err := l.conn.SetWriteDeadline(loxDate.date)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return nil, nil
			}
			return argMustBeType("date")
		})
	case "write", "send":
		return connFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var bytes []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				bytes = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"File argument to 'connection.%v' must be in read mode.",
							methodName,
						),
					)
				}
				var readErr error
				bytes, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				bytes = []byte(arg.str)
			default:
				return argMustBeType("buffer, file, or string")
			}
			numBytes, err := l.conn.Write(bytes)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return int64(numBytes), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Connection objects have no property called '"+methodName+"'.")
}

func (l *LoxConnection) String() string {
	return fmt.Sprintf("<connection object at %p>", l)
}

func (l *LoxConnection) Type() string {
	return "connection"
}
