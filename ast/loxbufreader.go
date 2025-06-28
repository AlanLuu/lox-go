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

const loxBufDefaultSize = 4096
const (
	loxBufReaderHTTPRes = iota + 1
)

type LoxBufReader struct {
	reader     *bufio.Reader
	origReader io.Reader
	readerType int
	methods    map[string]*struct{ ProtoLoxCallable }
}

func NewLoxBufReader(reader io.Reader) *LoxBufReader {
	return NewLoxBufReaderSize(reader, loxBufDefaultSize)
}

func NewLoxBufReaderSize(reader io.Reader, size int) *LoxBufReader {
	if reader == nil {
		panic("in NewLoxBufReaderSize: reader is nil")
	}
	return &LoxBufReader{
		reader:     bufio.NewReaderSize(reader, size),
		origReader: reader,
		methods:    make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxBufReaderType(reader io.Reader, readerType int) *LoxBufReader {
	return NewLoxBufReaderSizeType(reader, loxBufDefaultSize, readerType)
}

func NewLoxBufReaderSizeType(reader io.Reader, size, readerType int) *LoxBufReader {
	if reader == nil {
		panic("in NewLoxBufReaderSizeWithType: reader is nil")
	}
	return &LoxBufReader{
		reader:     bufio.NewReaderSize(reader, size),
		origReader: reader,
		readerType: readerType,
		methods:    make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxBufReaderStdin() *LoxBufReader {
	return NewLoxBufReader(os.Stdin)
}

func NewLoxBufReaderStdinSize(size int) *LoxBufReader {
	return NewLoxBufReaderSize(os.Stdin, size)
}

func (l *LoxBufReader) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	bufReaderFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native buffered reader fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'buffered reader.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'buffered reader.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "buffered":
		return bufReaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.reader.Buffered()), nil
		})
	case "copyToBufWriter":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxBufWriter, ok := args[0].(*LoxBufWriter); ok {
				numBytes, err := io.Copy(loxBufWriter.writer, l.reader)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return numBytes, nil
			}
			return argMustBeType("buffered writer")
		})
	case "discard":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				numBytes, err := l.reader.Discard(int(num))
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return int64(numBytes), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "peek":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				peekBytes, err := l.reader.Peek(int(num))
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				buffer := EmptyLoxBufferCap(int64(len(peekBytes)))
				for _, b := range peekBytes {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				}
				return buffer, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "read":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxBuffer, ok := args[0].(*LoxBuffer); ok {
				loxBufLen := len(loxBuffer.elements)
				bytes := make([]byte, loxBufLen)
				numBytes, err := io.ReadFull(l.reader, bytes)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				for i := 0; i < loxBufLen; i++ {
					loxBuffer.elements[i] = int64(bytes[i])
				}
				return int64(numBytes), nil
			}
			return argMustBeType("buffer")
		})
	case "readBuffer", "readBuf":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				if len(loxStr.str) != 1 {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"String argument to 'buffered reader.%v' must be exactly 1 byte long.",
							methodName,
						),
					)
				}
				bytes, err := l.reader.ReadBytes(loxStr.str[0])
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				buffer := EmptyLoxBufferCap(int64(len(bytes)))
				for _, b := range bytes {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				}
				return buffer, nil
			}
			return argMustBeType("string")
		})
	case "readByte":
		return bufReaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			b, err := l.reader.ReadByte()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return int64(b), nil
		})
	case "readBytes":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'buffered reader.readBytes' cannot be negative.")
				}
				buffer := EmptyLoxBufferCap(num)
				for i := int64(0); i < num; i++ {
					b, err := l.reader.ReadByte()
					if err != nil {
						if i == 0 {
							buffer = nil
							return EmptyLoxBuffer(), nil
						}
						break
					}
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				}
				return buffer, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "readBytesIter":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'buffered reader.readBytesIter' cannot be negative.")
				}
				var b byte
				var err error
				var count int64 = 0
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					if count >= num || err != nil {
						return false
					}
					b, err = l.reader.ReadByte()
					count++
					return err == nil
				}
				iterator.nextMethod = func() any {
					return int64(b)
				}
				return NewLoxIterator(iterator), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "readChar":
		return bufReaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			r, _, err := l.reader.ReadRune()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxStringQuote(string(r)), nil
		})
	case "readChars":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'buffered reader.readChars' cannot be negative.")
				}
				var builder strings.Builder
				for i := int64(0); i < num; i++ {
					r, _, err := l.reader.ReadRune()
					if err != nil {
						if i == 0 {
							return EmptyLoxString(), nil
						}
						break
					}
					builder.WriteRune(r)
				}
				return NewLoxStringQuote(builder.String()), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "readCharsIter":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'buffered reader.readCharsIter' cannot be negative.")
				}
				var r rune
				var err error
				var count int64 = 0
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					if count >= num || err != nil {
						return false
					}
					r, _, err = l.reader.ReadRune()
					count++
					return err == nil
				}
				iterator.nextMethod = func() any {
					return NewLoxStringQuote(string(r))
				}
				return NewLoxIterator(iterator), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "readString", "readStr":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				if len(loxStr.str) != 1 {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"String argument to 'buffered reader.%v' must be exactly 1 byte long.",
							methodName,
						),
					)
				}
				s, err := l.reader.ReadString(loxStr.str[0])
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return NewLoxStringQuote(s), nil
			}
			return argMustBeType("string")
		})
	case "readStrings", "readStrs":
		return bufReaderFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"First argument to 'buffered reader.%v' must be a string.",
						methodName,
					),
				)
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Second argument to 'buffered reader.%v' must be an integer.",
						methodName,
					),
				)
			}
			str := args[0].(*LoxString).str
			if len(str) != 1 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"String argument to 'buffered reader.%v' must be exactly 1 byte long.",
						methodName,
					),
				)
			}
			num := args[1].(int64)
			if num < 0 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Integer argument to 'buffered reader.%v' cannot be negative.",
						methodName,
					),
				)
			}
			strList := list.NewListCap[any](num)
			for i := int64(0); i < num; i++ {
				s, err := l.reader.ReadString(str[0])
				if err != nil {
					if i == 0 {
						strList = nil
						return EmptyLoxList(), nil
					}
					break
				}
				strList.Add(NewLoxStringQuote(s))
			}
			return NewLoxList(strList), nil
		})
	case "readStringsIter", "readStrsIter":
		return bufReaderFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"First argument to 'buffered reader.%v' must be a string.",
						methodName,
					),
				)
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Second argument to 'buffered reader.%v' must be an integer.",
						methodName,
					),
				)
			}
			str := args[0].(*LoxString).str
			if len(str) != 1 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"String argument to 'buffered reader.%v' must be exactly 1 byte long.",
						methodName,
					),
				)
			}
			num := args[1].(int64)
			if num < 0 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Integer argument to 'buffered reader.%v' cannot be negative.",
						methodName,
					),
				)
			}
			var s string
			var err error
			var count int64 = 0
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				if count >= num || err != nil {
					return false
				}
				s, err = l.reader.ReadString(str[0])
				count++
				return err == nil
			}
			iterator.nextMethod = func() any {
				return NewLoxStringQuote(s)
			}
			return NewLoxIterator(iterator), nil
		})
	case "readToFile":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxFile, ok := args[0].(*LoxFile); ok {
				if !loxFile.isWrite() && !loxFile.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'buffered reader.readToFile' must be in write or append mode.")
				}
				numBytes, err := io.Copy(loxFile.file, l.reader)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return numBytes, nil
			}
			return argMustBeType("file")
		})
	case "reset":
		return bufReaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxFile, ok := args[0].(*LoxFile); ok {
				if !loxFile.isRead() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'buffered reader.reset' must be in read mode.")
				}
				l.reader.Reset(loxFile.file)
				return nil, nil
			}
			return argMustBeType("file")
		})
	case "size":
		return bufReaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.reader.Size()), nil
		})
	case "unreadByte":
		return bufReaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.reader.UnreadByte()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "unreadByteBool":
		return bufReaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.reader.UnreadByte() == nil, nil
		})
	case "unreadChar":
		return bufReaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.reader.UnreadRune()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "unreadCharBool":
		return bufReaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.reader.UnreadRune() == nil, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Buffered readers have no property called '"+methodName+"'.")
}

func (l *LoxBufReader) String() string {
	switch origReader := l.origReader.(type) {
	case *os.File:
		switch origReader {
		case os.Stdin:
			return fmt.Sprintf("<buffered stdin reader at %p>", l)
		case os.Stdout:
			return fmt.Sprintf("<buffered stdout reader at %p>", l)
		case os.Stderr:
			return fmt.Sprintf("<buffered stderr reader at %p>", l)
		}
	case net.Conn:
		return fmt.Sprintf("<buffered connection reader at %p>", l)
	case *strings.Reader:
		return fmt.Sprintf("<buffered string reader at %p>", l)
	}

	switch l.readerType {
	case loxBufReaderHTTPRes:
		return fmt.Sprintf("<buffered HTTP response reader at %p>", l)
	}

	return fmt.Sprintf("<buffered reader at %p>", l)
}

func (l *LoxBufReader) Type() string {
	return "buffered reader"
}
