package ast

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxGZIPReaderIterator struct {
	reader  *gzip.Reader
	current [1]byte
	isAtEnd bool
	stop    bool
}

func (l *LoxGZIPReaderIterator) HasNext() bool {
	return !l.stop
}

func (l *LoxGZIPReaderIterator) Next() any {
	theByte := l.current[0]
	if !l.isAtEnd {
		_, err := l.reader.Read(l.current[:])
		if err != nil {
			l.isAtEnd = true
			if !errors.Is(err, io.EOF) {
				l.stop = true
			}
		}
	} else {
		l.stop = true
	}
	return int64(theByte)
}

type LoxGZIPReader struct {
	reader      *gzip.Reader
	isClosed    bool
	multistream bool
	methods     map[string]*struct{ ProtoLoxCallable }
}

func NewLoxGZIPReader(reader io.Reader) (*LoxGZIPReader, error) {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		if strings.Contains(err.Error(), "EOF") {
			err = loxerror.Error("Invalid gzip file.")
		}
		return nil, err
	}
	return &LoxGZIPReader{
		reader:      gzipReader,
		isClosed:    false,
		multistream: true,
		methods:     make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func (l *LoxGZIPReader) setMultistream(value bool) {
	l.reader.Multistream(value)
	l.multistream = value
}

func (l *LoxGZIPReader) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	gzipReaderFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native gzip reader fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'gzip reader.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'gzip reader.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	closedErr := func() (any, error) {
		return nil, loxerror.RuntimeError(name,
			fmt.Sprintf("Cannot call 'gzip reader.%v' on closed gzip reader objects.", methodName))
	}
	switch methodName {
	case "close":
		return gzipReaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isClosed {
				err := l.reader.Close()
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				l.isClosed = true
			}
			return nil, nil
		})
	case "isClosed":
		return gzipReaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isClosed, nil
		})
	case "isMultistream":
		return gzipReaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.multistream, nil
		})
	case "multistream":
		return gzipReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if value, ok := args[0].(bool); ok {
				l.setMultistream(value)
				return nil, nil
			}
			return argMustBeType("boolean")
		})
	case "read":
		return gzipReaderFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var data []byte
			var err error
			argsLen := len(args)
			switch argsLen {
			case 0:
				if l.isClosed {
					return closedErr()
				}
				data, err = io.ReadAll(l.reader)
			case 1:
				if numBytes, ok := args[0].(int64); ok {
					if l.isClosed {
						return closedErr()
					}
					if numBytes < 0 {
						return nil, loxerror.RuntimeError(name,
							"Argument to 'gzip reader.read' cannot be negative.")
					}
					data = make([]byte, numBytes)
					_, err = l.reader.Read(data)
				} else {
					return argMustBeTypeAn("integer")
				}
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			buffer := EmptyLoxBufferCap(int64(len(data)))
			for _, b := range data {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "readToFile":
		return gzipReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case *LoxFile:
				if l.isClosed {
					return closedErr()
				}
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'gzip reader.readToFile' must be in write or append mode.")
				}
			case *LoxString:
				if l.isClosed {
					return closedErr()
				}
			default:
				return argMustBeType("file or string")
			}
			data, err := io.ReadAll(l.reader)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			switch arg := args[0].(type) {
			case *LoxFile:
				_, err := arg.file.Write(data)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			case *LoxString:
				err := os.WriteFile(arg.str, data, 0666)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}
			return nil, nil
		})
	case "reset":
		return gzipReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			default:
				return argMustBeType("buffer or file")
			}
			if l.isClosed {
				return closedErr()
			}
			var reader io.Reader
			switch arg := args[0].(type) {
			case *LoxBuffer:
				data := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					data = append(data, byte(element.(int64)))
				}
				reader = bytes.NewBuffer(data)
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"Cannot reset gzip reader to file not in read mode.")
				}
				reader = arg.file
			}
			l.reader.Reset(reader)
			l.multistream = true
			return nil, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "gzip readers have no property called '"+methodName+"'.")
}

func (l *LoxGZIPReader) Iterator() interfaces.Iterator {
	iterator := &LoxGZIPReaderIterator{
		reader: l.reader,
	}
	numBytesRead, err := l.reader.Read(iterator.current[:])
	if err != nil {
		iterator.isAtEnd = true
		if numBytesRead == 0 || !errors.Is(err, io.EOF) {
			iterator.stop = true
		}
	}
	return iterator
}

func (l *LoxGZIPReader) String() string {
	return fmt.Sprintf("<gzip reader at %p>", l)
}

func (l *LoxGZIPReader) Type() string {
	return "gzip reader"
}
