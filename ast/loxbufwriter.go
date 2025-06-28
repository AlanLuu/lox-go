package ast

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxBufWriter struct {
	writer     *bufio.Writer
	origWriter io.Writer
	methods    map[string]*struct{ ProtoLoxCallable }
}

func NewLoxBufWriter(writer io.Writer) *LoxBufWriter {
	return NewLoxBufWriterSize(writer, loxBufDefaultSize)
}

func NewLoxBufWriterSize(writer io.Writer, size int) *LoxBufWriter {
	if writer == nil {
		panic("in NewLoxBufWriterSize: writer is nil")
	}
	return &LoxBufWriter{
		writer:     bufio.NewWriterSize(writer, size),
		origWriter: writer,
		methods:    make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxBufWriterStderr() *LoxBufWriter {
	return NewLoxBufWriter(os.Stderr)
}

func NewLoxBufWriterStderrSize(size int) *LoxBufWriter {
	return NewLoxBufWriterSize(os.Stderr, size)
}

func NewLoxBufWriterStdout() *LoxBufWriter {
	return NewLoxBufWriter(os.Stdout)
}

func NewLoxBufWriterStdoutSize(size int) *LoxBufWriter {
	return NewLoxBufWriterSize(os.Stdout, size)
}

func (l *LoxBufWriter) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	bufWriterFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native buffered writer fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'buffered writer.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "available":
		return bufWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.writer.Available()), nil
		})
	case "availableBuf":
		return bufWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			availableBuf := l.writer.AvailableBuffer()
			buffer := EmptyLoxBufferCap(int64(len(availableBuf)))
			for _, b := range availableBuf {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "buffered":
		return bufWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.writer.Buffered()), nil
		})
	case "copyBufReader":
		return bufWriterFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxBufReader, ok := args[0].(*LoxBufReader); ok {
				numBytes, err := io.Copy(l.writer, loxBufReader.reader)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return numBytes, nil
			}
			return argMustBeType("buffered reader")
		})
	case "flush":
		return bufWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.writer.Flush()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "flushBool":
		return bufWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.writer.Flush() == nil, nil
		})
	case "flushForce":
		return bufWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.writer.Flush()
			return nil, nil
		})
	case "isFile":
		return bufWriterFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxFile, ok := args[0].(*LoxFile); ok {
				if file, ok := l.origWriter.(*os.File); ok {
					return loxFile.file == file, nil
				}
				return false, nil
			}
			return argMustBeType("file")
		})
	case "isStderr":
		return bufWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if file, ok := l.origWriter.(*os.File); ok {
				return file == os.Stderr, nil
			}
			return false, nil
		})
	case "isStdout":
		return bufWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if file, ok := l.origWriter.(*os.File); ok {
				return file == os.Stdout, nil
			}
			return false, nil
		})
	case "reset":
		return bufWriterFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxFile, ok := args[0].(*LoxFile); ok {
				if !loxFile.isWrite() && !loxFile.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'buffered writer.reset' must be in write or append mode.")
				}
				l.writer.Reset(loxFile.file)
				return nil, nil
			}
			return argMustBeType("file")
		})
	case "size":
		return bufWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.writer.Size()), nil
		})
	case "write":
		return bufWriterFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var numBytes int64
			var err error
			switch arg := args[0].(type) {
			case int64:
				if arg < 0 || arg > 255 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'buffered writer.write' must be an integer between 0 and 255.")
				}
				numBytes = 1
				err = l.writer.WriteByte(byte(arg))
			case *LoxBuffer:
				bytes := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				var n int
				n, err = l.writer.Write(bytes)
				numBytes = int64(n)
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'buffered writer.write' must be in read mode.")
				}
				numBytes, err = io.Copy(l.writer, arg.file)
			case *LoxString:
				var n int
				n, err = l.writer.WriteString(arg.str)
				numBytes = int64(n)
			default:
				return nil, loxerror.RuntimeError(name,
					"Argument to 'buffered writer.write' must be an integer, buffer, file, or string.")
			}
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return numBytes, nil
		})
	case "writeArgs":
		return bufWriterFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			for i, arg := range args {
				var err error
				switch arg := arg.(type) {
				case int64:
					if arg < 0 || arg > 255 {
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								"Integer at argument #%v in 'buffered writer.writeArgs' "+
									"must be an integer between 0 and 255.",
								i+1,
							),
						)
					}
					err = l.writer.WriteByte(byte(arg))
				case *LoxBuffer:
					bytes := make([]byte, 0, len(arg.elements))
					for _, element := range arg.elements {
						bytes = append(bytes, byte(element.(int64)))
					}
					_, err = l.writer.Write(bytes)
				case *LoxFile:
					if !arg.isRead() {
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								"File at argument #%v in 'buffered writer.writeArgs' "+
									"must be in read mode.",
								i+1,
							),
						)
					}
					_, err = io.Copy(l.writer, arg.file)
				case *LoxString:
					_, err = l.writer.WriteString(arg.str)
				default:
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument #%v in 'buffered writer.writeArgs' "+
								"must be an integer, buffer, file, or string.",
							i+1,
						),
					)
				}
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}
			return nil, nil
		})
	case "writeList":
		return bufWriterFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxList, ok := args[0].(*LoxList); ok {
				for i, arg := range loxList.elements {
					var err error
					switch arg := arg.(type) {
					case int64:
						if arg < 0 || arg > 255 {
							return nil, loxerror.RuntimeError(
								name,
								fmt.Sprintf(
									"Integer list element at index #%v in "+
										"'buffered writer.writeList' "+
										"must be an integer between 0 and 255.",
									i,
								),
							)
						}
						err = l.writer.WriteByte(byte(arg))
					case *LoxBuffer:
						bytes := make([]byte, 0, len(arg.elements))
						for _, element := range arg.elements {
							bytes = append(bytes, byte(element.(int64)))
						}
						_, err = l.writer.Write(bytes)
					case *LoxFile:
						if !arg.isRead() {
							return nil, loxerror.RuntimeError(
								name,
								fmt.Sprintf(
									"File list element at index #%v in "+
										"'buffered writer.writeList' "+
										"must be in read mode.",
									i,
								),
							)
						}
						_, err = io.Copy(l.writer, arg.file)
					case *LoxString:
						_, err = l.writer.WriteString(arg.str)
					default:
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								"List element at index #%v in "+
									"'buffered writer.writeList' "+
									"must be an integer, buffer, file, or string.",
								i,
							),
						)
					}
					if err != nil {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				}
				return nil, nil
			}
			return argMustBeType("list")
		})
	}
	return nil, loxerror.RuntimeError(name, "Buffered writers have no property called '"+methodName+"'.")
}

func (l *LoxBufWriter) String() string {
	switch origWriter := l.origWriter.(type) {
	case *os.File:
		switch origWriter {
		case os.Stdin:
			return fmt.Sprintf("<buffered stdin writer at %p>", l)
		case os.Stdout:
			return fmt.Sprintf("<buffered stdout writer at %p>", l)
		case os.Stderr:
			return fmt.Sprintf("<buffered stderr writer at %p>", l)
		}
	case net.Conn:
		return fmt.Sprintf("<buffered connection writer at %p>", l)
	case *strings.Builder:
		return fmt.Sprintf("<buffered stringbuilder writer at %p>", l)
	}
	return fmt.Sprintf("<buffered writer at %p>", l)
}

func (l *LoxBufWriter) Type() string {
	return "buffered writer"
}
