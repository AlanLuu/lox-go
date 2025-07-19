package ast

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxStringReader struct {
	reader  *strings.Reader
	str     string
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxStringReader(str string) *LoxStringReader {
	return &LoxStringReader{
		reader:  strings.NewReader(str),
		str:     str,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxStringReader) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	stringreaderFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native string reader fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'string reader.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'string reader.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "read":
		return stringreaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxBuffer, ok := args[0].(*LoxBuffer); ok {
				loxBufLen := len(loxBuffer.elements)
				bytes := make([]byte, loxBufLen)
				numBytes, err := l.reader.Read(bytes)
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
	case "readAt":
		return stringreaderFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxBuffer); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'string reader.readAt' must be a buffer.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'string reader.readAt' must be an integer.")
			}
			loxBuffer := args[0].(*LoxBuffer)
			offset := args[1].(int64)
			loxBufLen := len(loxBuffer.elements)
			bytes := make([]byte, loxBufLen)
			numBytes, err := l.reader.ReadAt(bytes, offset)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			for i := 0; i < loxBufLen; i++ {
				loxBuffer.elements[i] = int64(bytes[i])
			}
			return int64(numBytes), nil
		})
	case "readByte":
		return stringreaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			b, err := l.reader.ReadByte()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return int64(b), nil
		})
	case "readBytes":
		return stringreaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'string reader.readBytes' cannot be negative.")
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
		return stringreaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'string reader.readBytesIter' cannot be negative.")
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
		return stringreaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			r, _, err := l.reader.ReadRune()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxStringQuote(string(r)), nil
		})
	case "readChars":
		return stringreaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'string reader.readChars' cannot be negative.")
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
		return stringreaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'string reader.readCharsIter' cannot be negative.")
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
	case "readToFile":
		return stringreaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxFile, ok := args[0].(*LoxFile); ok {
				if !loxFile.isWrite() && !loxFile.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'string reader.readToFile' must be in write or append mode.")
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
		return stringreaderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.reader.Reset(loxStr.str)
				return nil, nil
			}
			return argMustBeType("string")
		})
	case "seek":
		return stringreaderFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'string reader.seek' must be an integer.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'string reader.seek' must be an integer.")
			}
			offset := args[0].(int64)
			whence := int(args[1].(int64))
			position, seekErr := l.reader.Seek(offset, whence)
			if seekErr != nil {
				return nil, loxerror.RuntimeError(name, seekErr.Error())
			}
			return position, nil
		})
	case "size":
		return stringreaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.reader.Size(), nil
		})
	case "sizeutf8":
		return stringreaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(utf8.RuneCountInString(l.str)), nil
		})
	case "unreadByte":
		return stringreaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.reader.UnreadByte()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "unreadByteBool":
		return stringreaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.reader.UnreadByte() == nil, nil
		})
	case "unreadChar":
		return stringreaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.reader.UnreadRune()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "unreadCharBool":
		return stringreaderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.reader.UnreadRune() == nil, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "String readers have no property called '"+methodName+"'.")
}

func (l *LoxStringReader) Length() int64 {
	return int64(l.reader.Len())
}

func (l *LoxStringReader) String() string {
	return fmt.Sprintf("<string reader object at %p>", l)
}

func (l *LoxStringReader) Type() string {
	return "string reader"
}
