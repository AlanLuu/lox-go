package ast

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxGZIPWriter struct {
	writer      *gzip.Writer
	bytesBuffer *bytes.Buffer
	isClosed    bool
	methods     map[string]*struct{ ProtoLoxCallable }
}

func NewLoxGZIPWriter(writer io.Writer) *LoxGZIPWriter {
	return &LoxGZIPWriter{
		writer:      gzip.NewWriter(writer),
		bytesBuffer: nil,
		isClosed:    false,
		methods:     make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxGZIPWriterLevel(writer io.Writer, level int) (*LoxGZIPWriter, error) {
	writerLevel, err := gzip.NewWriterLevel(writer, level)
	if err != nil {
		return nil, err
	}
	return &LoxGZIPWriter{
		writer:      writerLevel,
		bytesBuffer: nil,
		isClosed:    false,
		methods:     make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func NewLoxGZIPWriterBytes(bytesBuffer *bytes.Buffer) *LoxGZIPWriter {
	gzipWriter := NewLoxGZIPWriter(bytesBuffer)
	gzipWriter.bytesBuffer = bytesBuffer
	return gzipWriter
}

func NewLoxGZIPWriterBytesLevel(bytesBuffer *bytes.Buffer, level int) (*LoxGZIPWriter, error) {
	gzipWriter, err := NewLoxGZIPWriterLevel(bytesBuffer, level)
	if err != nil {
		return nil, err
	}
	gzipWriter.bytesBuffer = bytesBuffer
	return gzipWriter, nil
}

func (l *LoxGZIPWriter) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	gzipWriterFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native gzip writer fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'gzip writer.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	closedErr := func() (any, error) {
		return nil, loxerror.RuntimeError(name,
			fmt.Sprintf("Cannot call 'gzip writer.%v' on closed gzip writer objects.", methodName))
	}
	switch methodName {
	case "buffer":
		return gzipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.bytesBuffer == nil {
				return nil, loxerror.RuntimeError(name,
					"gzip file is not being written to a buffer.")
			}
			bytes := l.bytesBuffer.Bytes()
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, b := range bytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "close":
		return gzipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isClosed {
				err := l.writer.Close()
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				l.isClosed = true
			}
			return nil, nil
		})
	case "flush":
		return gzipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.writer.Flush()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l, nil
		})
	case "isBuffer":
		return gzipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.bytesBuffer != nil, nil
		})
	case "isClosed":
		return gzipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isClosed, nil
		})
	case "reset":
		return gzipWriterFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxFile:
			case int64:
			default:
				return argMustBeType("file or the field 'gzip.USE_BUFFER'")
			}
			if l.isClosed {
				return closedErr()
			}
			var writer io.Writer
			switch arg := args[0].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Cannot reset gzip writer to file not in write or append mode.")
				}
				writer = arg.file
				l.bytesBuffer = nil
			case int64:
				switch arg {
				case GZIP_USE_BUFFER:
					bytesBuffer := new(bytes.Buffer)
					writer = bytesBuffer
					l.bytesBuffer = bytesBuffer
				default:
					return nil, loxerror.RuntimeError(name,
						"Integer argument to 'gzip writer.reset' must be equal to the field 'gzip.USE_BUFFER'.")
				}
			}
			l.writer.Reset(writer)
			return l, nil
		})
	case "write":
		return gzipWriterFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Argument to 'gzip writer.write' must be a buffer, file, or string.")
			}
			if l.isClosed {
				return closedErr()
			}
			var data []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				data = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					data = append(data, byte(element.(int64)))
				}
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'gzip writer.write' must be in read mode.")
				}
				var readErr error
				data, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				data = []byte(arg.str)
			}
			_, err := l.writer.Write(data)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "gzip writers have no property called '"+methodName+"'.")
}

func (l *LoxGZIPWriter) String() string {
	return fmt.Sprintf("<gzip writer at %p>", l)
}

func (l *LoxGZIPWriter) Type() string {
	return "gzip writer"
}
