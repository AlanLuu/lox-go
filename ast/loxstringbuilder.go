package ast

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxStringBuilder struct {
	builder strings.Builder
	loxStr  *LoxString
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxStringBuilder() *LoxStringBuilder {
	return &LoxStringBuilder{
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxStringBuilder) Capacity() int64 {
	return int64(l.builder.Cap())
}

func (l *LoxStringBuilder) clear() {
	l.builder.Reset()
	l.loxStr = nil
}

func (l *LoxStringBuilder) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	stringbuilderFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native stringbuilder fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'stringbuilder.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'stringbuilder.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "append":
		return stringbuilderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.builder.WriteString(loxStr.str)
				return l, nil
			}
			return argMustBeType("string")
		})
	case "appendBuf":
		return stringbuilderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxBuffer, ok := args[0].(*LoxBuffer); ok {
				bytes := make([]byte, 0, len(loxBuffer.elements))
				for _, element := range loxBuffer.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				l.builder.Write(bytes)
				return l, nil
			}
			return argMustBeType("buffer")
		})
	case "appendBuilder":
		return stringbuilderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if sb, ok := args[0].(*LoxStringBuilder); ok {
				l.builder.WriteString(sb.builder.String())
				return l, nil
			}
			return argMustBeType("stringbuilder")
		})
	case "appendByte":
		return stringbuilderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if byteNum, ok := args[0].(int64); ok {
				if byteNum < 0 || byteNum > 255 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'stringbuilder.appendByte' must be an integer between 0 and 255.")
				}
				l.builder.WriteByte(byte(byteNum))
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "appendCodePoint":
		return stringbuilderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if codePoint, ok := args[0].(int64); ok {
				if codePoint < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'stringbuilder.appendCodePoint' cannot be negative.")
				}
				if codePoint > utf8.MaxRune {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument to 'stringbuilder.appendCodePoint' cannot be larger than %v.",
							int(utf8.MaxRune),
						),
					)
				}
				l.builder.WriteRune(rune(codePoint))
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "buffer":
		return stringbuilderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			bytes := []byte(l.builder.String())
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, value := range bytes {
				addErr := buffer.add(int64(value))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "cap", "capacity":
		return stringbuilderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.builder.Cap()), nil
		})
	case "clear":
		return stringbuilderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.clear()
			return l, nil
		})
	case "equals":
		return stringbuilderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if sb, ok := args[0].(*LoxStringBuilder); ok {
				return l.builder.String() == sb.builder.String(), nil
			}
			return argMustBeType("stringbuilder")
		})
	case "grow":
		return stringbuilderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if growFactor, ok := args[0].(int64); ok {
				if growFactor < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'stringbuilder.grow' cannot be negative.")
				}
				l.builder.Grow(int(growFactor))
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "len", "length":
		return stringbuilderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.builder.Len() == 0 {
				return int64(0), nil
			}
			return int64(utf8.RuneCountInString(l.builder.String())), nil
		})
	case "numBytes":
		return stringbuilderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.builder.Len()), nil
		})
	case "string":
		return stringbuilderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			str := l.builder.String()
			if l.loxStr == nil || str != l.loxStr.str {
				l.loxStr = NewLoxStringQuote(str)
			}
			return l.loxStr, nil
		})
	case "stringNew":
		return stringbuilderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.builder.String()), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Stringbuilders have no property called '"+methodName+"'.")
}

func (l *LoxStringBuilder) Length() int64 {
	if l.builder.Len() == 0 {
		return 0
	}
	return int64(utf8.RuneCountInString(l.builder.String()))
}

func (l *LoxStringBuilder) String() string {
	return fmt.Sprintf("<stringbuilder object at %p>", l)
}

func (l *LoxStringBuilder) Type() string {
	return "stringbuilder"
}
