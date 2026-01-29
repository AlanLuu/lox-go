package ast

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type LoxString struct {
	str     string
	quote   byte
	methods map[string]*struct{ ProtoLoxCallable }
}

type LoxStringIterator struct {
	loxStr *LoxString
	index  int64
}

func (l *LoxStringIterator) HasNext() bool {
	return l.index < l.loxStr.Length()
}

func (l *LoxStringIterator) Next() any {
	c := []rune(l.loxStr.str)[l.index]
	l.index++
	return NewLoxStringQuote(string(c))
}

func NewLoxString(str string, quote byte) *LoxString {
	return &LoxString{
		str:     str,
		quote:   quote,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxStringChar[T byte | rune](char T) *LoxString {
	str := string(char)
	if char == '\'' {
		return NewLoxString(str, '"')
	}
	return NewLoxString(str, '\'')
}

func NewLoxStringQuote(str string) *LoxString {
	if strings.Contains(str, "'") {
		return NewLoxString(str, '"')
	}
	return NewLoxString(str, '\'')
}

func EmptyLoxString() *LoxString {
	return NewLoxString("", '\'')
}

func StringIndexMustBeWholeNum(index any) string {
	indexVal := index
	format := "%v"
	switch index := index.(type) {
	case float64:
		if util.FloatIsInt(index) {
			format = "%.1f"
		} else {
			indexVal = util.FormatFloat(index)
		}
	}
	return fmt.Sprintf(
		"String index '"+format+"' must be an integer or bigint.",
		indexVal,
	)
}

func StringIndexOutOfRange(index int64) string {
	return fmt.Sprintf("String index %v out of range.", index)
}

func (l *LoxString) NewLoxString(str string) *LoxString {
	return NewLoxString(str, l.quote)
}

func (l *LoxString) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxString:
		if l == obj {
			return true
		}
		return l.str == obj.str
	default:
		return false
	}
}

func (l *LoxString) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	strFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native string fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	padString := func(initialStr string, finalStrLen int64, arg any, padAtStart bool) (string, bool) {
		var builder strings.Builder
		if !padAtStart {
			builder.WriteString(initialStr)
		}

		var padStr string
		switch arg := arg.(type) {
		case string:
			padStr = arg
		default:
			padStr = getResult(arg, arg, true)
		}

		useDoubleQuote := false
		padStrLen := int64(len(padStr))
		if padStrLen > 0 {
			offset := finalStrLen - int64(len(initialStr))
			for i := int64(0); i < offset; i++ {
				b := padStr[i%padStrLen]
				if !useDoubleQuote && b == '\'' {
					useDoubleQuote = true
				}
				builder.WriteByte(b)
			}
		}
		if padAtStart {
			builder.WriteString(initialStr)
		}

		return builder.String(), useDoubleQuote
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'string.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'string.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "caesar":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if shift, ok := args[0].(int64); ok {
				var upperA, upperZ, lowerA, lowerZ int64 = 65, 90, 97, 122
				var builder strings.Builder
				for _, c := range l.str {
					cc := int64(c)
					if cc >= upperA && cc <= upperZ {
						builder.WriteRune(rune(((cc-upperA)+shift)%26 + upperA))
					} else if cc >= lowerA && cc <= lowerZ {
						builder.WriteRune(rune(((cc-lowerA)+shift)%26 + lowerA))
					} else {
						builder.WriteRune(c)
					}
				}
				return NewLoxString(builder.String(), l.quote), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "capitalize":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			switch utf8.RuneCountInString(l.str) {
			case 0:
				return EmptyLoxString(), nil
			case 1:
				return NewLoxString(strings.ToUpper(l.str), l.quote), nil
			}
			runes := []rune(l.str)
			newStr := strings.ToUpper(string(runes[0])) + strings.ToLower(string(runes[1:]))
			return NewLoxString(newStr, l.quote), nil
		})
	case "center":
		return strFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var width int64
			var fillStr string = " "
			argsLen := len(args)
			switch argsLen {
			case 1, 2:
				if arg, ok := args[0].(int64); ok {
					width = arg
				} else {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'string.center' must be an integer.")
				}
				if argsLen == 2 {
					const errStr = "Second argument to 'string.center' must be a single character."
					if arg, ok := args[1].(*LoxString); ok {
						if arg.Length() != 1 {
							return nil, loxerror.RuntimeError(name, errStr)
						}
						fillStr = arg.str
					} else {
						return nil, loxerror.RuntimeError(name, errStr)
					}
				}
			default:
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Expected 1 or 2 arguments but got %v.",
						argsLen,
					),
				)
			}
			sLen := l.Length()
			if width <= sLen {
				return l, nil
			}
			s := l.str
			b := sLen%2 == 0
			for sLen < width {
				if b {
					s = fillStr + s
				} else {
					s += fillStr
				}
				sLen++
				if sLen < 0 {
					break
				}
				b = !b
			}
			return NewLoxStringQuote(s), nil
		})
	case "compare":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return int64(strings.Compare(l.str, loxStr.str)), nil
			}
			return argMustBeType("string")
		})
	case "contains":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return strings.Contains(l.str, loxStr.str), nil
			}
			return argMustBeType("string")
		})
	case "count":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return int64(strings.Count(l.str, loxStr.str)), nil
			}
			return argMustBeType("string")
		})
	case "cut":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				before, after, _ := strings.Cut(l.str, loxStr.str)
				pair := list.NewListCap[any](2)
				pair.Add(NewLoxStringQuote(before))
				pair.Add(NewLoxStringQuote(after))
				return NewLoxList(pair), nil
			}
			return argMustBeType("string")
		})
	case "cutPrefix":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				s, _ := strings.CutPrefix(l.str, loxStr.str)
				return NewLoxStringQuote(s), nil
			}
			return argMustBeType("string")
		})
	case "cutSuffix":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				s, _ := strings.CutSuffix(l.str, loxStr.str)
				return NewLoxStringQuote(s), nil
			}
			return argMustBeType("string")
		})
	case "endsWith":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return strings.HasSuffix(l.str, loxStr.str), nil
			}
			return argMustBeType("string")
		})
	case "equalsIgnoreCase":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return strings.EqualFold(l.str, loxStr.str), nil
			}
			return argMustBeType("string")
		})
	case "fields":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			fields := strings.Fields(l.str)
			fieldsList := list.NewListCap[any](int64(len(fields)))
			for _, field := range fields {
				fieldsList.Add(NewLoxStringQuote(field))
			}
			return NewLoxList(fieldsList), nil
		})
	case "filter":
		return strFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				var builder strings.Builder
				for index, char := range l.str {
					argList[0] = NewLoxStringChar(char)
					argList[1] = int64(index)
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return nil, resultErr
					}
					if i.isTruthy(result) {
						builder.WriteRune(char)
					}
				}
				return NewLoxStringQuote(builder.String()), nil
			}
			return argMustBeType("function")
		})
	case "index":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return int64(strings.Index(l.str, loxStr.str)), nil
			}
			return argMustBeType("string")
		})
	case "indexFrom":
		return strFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'string.indexFrom' must be a string.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'string.indexFrom' must be an integer.")
			}
			loxStr := args[0].(*LoxString)
			fromIndex := args[1].(int64)
			if fromIndex <= 0 {
				return int64(strings.Index(l.str, loxStr.str)), nil
			}
			if fromIndex > int64(utf8.RuneCountInString(l.str)) {
				return int64(-1), nil
			}
			runes := []rune(l.str)
			result := strings.Index(string(runes[fromIndex:]), loxStr.str)
			if result == -1 {
				return int64(-1), nil
			}
			return int64(result) + fromIndex, nil
		})
	case "isAlnum", "isalnum":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.str == "" {
				return false, nil
			}
			for _, c := range l.str {
				if !unicode.IsLetter(c) && !unicode.IsDigit(c) && !unicode.IsNumber(c) {
					return false, nil
				}
			}
			return true, nil
		})
	case "isAlpha", "isalpha":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.str == "" {
				return false, nil
			}
			for _, c := range l.str {
				if !unicode.IsLetter(c) {
					return false, nil
				}
			}
			return true, nil
		})
	case "isAscii", "isascii":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			for _, c := range l.str {
				if c > unicode.MaxASCII {
					return false, nil
				}
			}
			return true, nil
		})
	case "isDigit", "isdigit":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.str == "" {
				return false, nil
			}
			for _, c := range l.str {
				if !unicode.IsDigit(c) {
					return false, nil
				}
			}
			return true, nil
		})
	case "isEmpty":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return len(l.str) == 0, nil
		})
	case "isLower", "islower":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.str == "" {
				return false, nil
			}
			for _, c := range l.str {
				if !unicode.IsLower(c) {
					return false, nil
				}
			}
			return true, nil
		})
	case "isNumeric", "isnumeric":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.str == "" {
				return false, nil
			}
			for _, c := range l.str {
				if !unicode.IsNumber(c) {
					return false, nil
				}
			}
			return true, nil
		})
	case "isPrintable", "isprintable":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			for _, c := range l.str {
				if !unicode.IsPrint(c) {
					return false, nil
				}
			}
			return true, nil
		})
	case "isSpace", "isspace":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.str == "" {
				return false, nil
			}
			for _, c := range l.str {
				if !unicode.IsSpace(c) {
					return false, nil
				}
			}
			return true, nil
		})
	case "isUpper", "isupper":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.str == "" {
				return false, nil
			}
			for _, c := range l.str {
				if !unicode.IsUpper(c) {
					return false, nil
				}
			}
			return true, nil
		})
	case "join":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if iterable, ok := args[0].(interfaces.Iterable); ok {
				iterator := iterable.Iterator()
				iteratorWithErr, isErrIter := iterator.(interfaces.IteratorErr)
				hasNext := func() (bool, error) {
					if isErrIter {
						return iteratorWithErr.HasNextErr()
					}
					return iterator.HasNext(), nil
				}
				next := func() (any, error) {
					if isErrIter {
						return iteratorWithErr.NextErr()
					}
					return iterator.Next(), nil
				}
				ok, hasNextErr := hasNext()
				if hasNextErr != nil {
					return nil, loxerror.RuntimeError(name, hasNextErr.Error())
				}
				if !ok {
					return EmptyLoxString(), nil
				}
				var builder strings.Builder
				for {
					value, nextErr := next()
					if nextErr != nil {
						return nil, loxerror.RuntimeError(name, nextErr.Error())
					}
					builder.WriteString(getResult(value, value, true))
					ok, hasNextErr = hasNext()
					if hasNextErr != nil {
						return nil, loxerror.RuntimeError(name, hasNextErr.Error())
					}
					if !ok {
						break
					}
					builder.WriteString(l.str)
				}
				return NewLoxStringQuote(builder.String()), nil
			}
			return nil, loxerror.RuntimeError(name,
				"Argument to 'string.join' is not iterable.")
		})
	case "lastIndex":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return int64(strings.LastIndex(l.str, loxStr.str)), nil
			}
			return argMustBeType("string")
		})
	case "lastIndexFrom":
		return strFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'string.lastIndexFrom' must be a string.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'string.lastIndexFrom' must be an integer.")
			}
			loxStr := args[0].(*LoxString)
			fromIndex := args[1].(int64)
			origStrLen := int64(utf8.RuneCountInString(l.str))
			if fromIndex >= origStrLen {
				return int64(strings.LastIndex(l.str, loxStr.str)), nil
			}
			if fromIndex < 0 {
				return int64(-1), nil
			}
			substrLen := int64(utf8.RuneCountInString(loxStr.str))
			maxStartPos := origStrLen - substrLen
			if fromIndex > maxStartPos {
				fromIndex = maxStartPos
			}
			endPos := fromIndex + substrLen
			if endPos > origStrLen {
				endPos = origStrLen
			}
			runes := []rune(l.str)
			result := int64(strings.LastIndex(string(runes[:endPos]), loxStr.str))
			if result != -1 && result <= fromIndex {
				return result, nil
			}
			return int64(-1), nil
		})
	case "ljust":
		return strFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var width int64
			var fillStr string = " "
			argsLen := len(args)
			switch argsLen {
			case 1, 2:
				if arg, ok := args[0].(int64); ok {
					width = arg
				} else {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'string.ljust' must be an integer.")
				}
				if argsLen == 2 {
					const errStr = "Second argument to 'string.ljust' must be a single character."
					if arg, ok := args[1].(*LoxString); ok {
						if arg.Length() != 1 {
							return nil, loxerror.RuntimeError(name, errStr)
						}
						fillStr = arg.str
					} else {
						return nil, loxerror.RuntimeError(name, errStr)
					}
				}
			default:
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Expected 1 or 2 arguments but got %v.",
						argsLen,
					),
				)
			}
			sLen := l.Length()
			if width <= sLen {
				return l, nil
			}
			var builder strings.Builder
			builder.WriteString(l.str)
			for sLen < width {
				builder.WriteString(fillStr)
				sLen++
				if sLen < 0 {
					break
				}
			}
			return NewLoxStringQuote(builder.String()), nil
		})
	case "lower":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxString(strings.ToLower(l.str), l.quote), nil
		})
	case "lstrip":
		return strFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 0:
				return NewLoxString(strings.TrimLeftFunc(l.str, unicode.IsSpace), l.quote), nil
			case 1:
				if loxStr, ok := args[0].(*LoxString); ok {
					return NewLoxString(strings.TrimLeft(l.str, loxStr.str), l.quote), nil
				}
				return argMustBeType("string")
			}
			return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		})
	case "padEnd":
		return strFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if finalStrLen, ok := args[0].(int64); ok {
				paddedStr, useDoubleQuote := padString(l.str, finalStrLen, args[1], false)
				if useDoubleQuote {
					return NewLoxString(paddedStr, '"'), nil
				}
				return NewLoxString(paddedStr, l.quote), nil
			}
			return nil, loxerror.RuntimeError(name, "First argument to 'string.padEnd' must be an integer.")
		})
	case "padStart":
		return strFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if finalStrLen, ok := args[0].(int64); ok {
				paddedStr, useDoubleQuote := padString(l.str, finalStrLen, args[1], true)
				if useDoubleQuote {
					return NewLoxString(paddedStr, '"'), nil
				}
				return NewLoxString(paddedStr, l.quote), nil
			}
			return nil, loxerror.RuntimeError(name, "First argument to 'string.padStart' must be an integer.")
		})
	case "quote":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(strconv.Quote(l.str)), nil
		})
	case "replace":
		return strFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if firstStr, firstStrOk := args[0].(*LoxString); firstStrOk {
				if secondStr, secondStrOk := args[1].(*LoxString); secondStrOk {
					newStr := strings.ReplaceAll(l.str, firstStr.str, secondStr.str)
					if strings.Contains(newStr, "'") {
						return NewLoxString(newStr, '"'), nil
					}
					return NewLoxString(newStr, '\''), nil
				}
				return nil, loxerror.RuntimeError(name, "Second argument to 'string.replace' must be a string.")
			}
			return nil, loxerror.RuntimeError(name, "First argument to 'string.replace' must be a string.")
		})
	case "reversed":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			runes := []rune(l.str)
			var builder strings.Builder
			for i := len(runes) - 1; i >= 0; i-- {
				builder.WriteRune(runes[i])
			}
			return NewLoxStringQuote(builder.String()), nil
		})
	case "reversedWords":
		return strFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var delimiter string
			argsLen := len(args)
			switch argsLen {
			case 0:
				delimiter = " "
			case 1:
				if loxStr, ok := args[0].(*LoxString); ok {
					delimiter = loxStr.str
				} else {
					return argMustBeType("string")
				}
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
			words := strings.Split(l.str, delimiter)
			var builder strings.Builder
			for i := len(words) - 1; i >= 0; i-- {
				builder.WriteString(words[i])
				if i > 0 {
					builder.WriteString(delimiter)
				}
			}
			return NewLoxStringQuote(builder.String()), nil
		})
	case "rjust":
		return strFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var width int64
			var fillStr string = " "
			argsLen := len(args)
			switch argsLen {
			case 1, 2:
				if arg, ok := args[0].(int64); ok {
					width = arg
				} else {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'string.rjust' must be an integer.")
				}
				if argsLen == 2 {
					const errStr = "Second argument to 'string.rjust' must be a single character."
					if arg, ok := args[1].(*LoxString); ok {
						if arg.Length() != 1 {
							return nil, loxerror.RuntimeError(name, errStr)
						}
						fillStr = arg.str
					} else {
						return nil, loxerror.RuntimeError(name, errStr)
					}
				}
			default:
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Expected 1 or 2 arguments but got %v.",
						argsLen,
					),
				)
			}
			sLen := l.Length()
			if width <= sLen {
				return l, nil
			}
			s := l.str
			for sLen < width {
				s = fillStr + s
				sLen++
				if sLen < 0 {
					break
				}
			}
			return NewLoxStringQuote(s), nil
		})
	case "rot13":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var upperA, upperZ, lowerA, lowerZ rune = 65, 90, 97, 122
			var builder strings.Builder
			for _, c := range l.str {
				if c >= upperA && c <= upperZ {
					builder.WriteRune(((c-upperA)+13)%26 + upperA)
				} else if c >= lowerA && c <= lowerZ {
					builder.WriteRune(((c-lowerA)+13)%26 + lowerA)
				} else {
					builder.WriteRune(c)
				}
			}
			return NewLoxString(builder.String(), l.quote), nil
		})
	case "rot18":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var upperA, upperZ, lowerA, lowerZ rune = 65, 90, 97, 122
			var zeroPos, ninePos rune = 48, 57
			var builder strings.Builder
			for _, c := range l.str {
				if c >= upperA && c <= upperZ {
					builder.WriteRune(((c-upperA)+13)%26 + upperA)
				} else if c >= lowerA && c <= lowerZ {
					builder.WriteRune(((c-lowerA)+13)%26 + lowerA)
				} else if c >= zeroPos && c <= ninePos {
					builder.WriteRune(((c-zeroPos)+5)%10 + zeroPos)
				} else {
					builder.WriteRune(c)
				}
			}
			return NewLoxString(builder.String(), l.quote), nil
		})
	case "rot47":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var lower, upper rune = 33, 126
			var builder strings.Builder
			for _, c := range l.str {
				if c >= lower && c <= upper {
					builder.WriteRune(((c-lower)+47)%94 + lower)
				} else {
					builder.WriteRune(c)
				}
			}
			return NewLoxString(builder.String(), l.quote), nil
		})
	case "rstrip":
		return strFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 0:
				return NewLoxString(strings.TrimRightFunc(l.str, unicode.IsSpace), l.quote), nil
			case 1:
				if loxStr, ok := args[0].(*LoxString); ok {
					return NewLoxString(strings.TrimRight(l.str, loxStr.str), l.quote), nil
				}
				return argMustBeType("string")
			}
			return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		})
	case "shuffled":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			runes := []rune(l.str)
			rand.Shuffle(len(runes), func(a int, b int) {
				runes[a], runes[b] = runes[b], runes[a]
			})
			return NewLoxString(string(runes), l.quote), nil
		})
	case "split":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				splitSlice := strings.Split(l.str, loxStr.str)
				loxList := list.NewListCap[any](int64(len(splitSlice)))
				for _, str := range splitSlice {
					loxList.Add(NewLoxStringQuote(str))
				}
				return NewLoxList(loxList), nil
			}
			return argMustBeType("string")
		})
	case "splitAfter":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				splitSlice := strings.SplitAfter(l.str, loxStr.str)
				loxList := list.NewListCap[any](int64(len(splitSlice)))
				for _, str := range splitSlice {
					loxList.Add(NewLoxStringQuote(str))
				}
				return NewLoxList(loxList), nil
			}
			return argMustBeType("string")
		})
	case "startsWith":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return strings.HasPrefix(l.str, loxStr.str), nil
			}
			return argMustBeType("string")
		})
	case "strip":
		return strFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 0:
				return NewLoxString(strings.TrimSpace(l.str), l.quote), nil
			case 1:
				if loxStr, ok := args[0].(*LoxString); ok {
					return NewLoxString(strings.Trim(l.str, loxStr.str), l.quote), nil
				}
				return argMustBeType("string")
			}
			return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		})
	case "swapcase":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var upperA, upperZ, lowerA, lowerZ rune = 65, 90, 97, 122
			var offset rune = 32
			var builder strings.Builder
			for _, c := range l.str {
				if c >= upperA && c <= upperZ {
					builder.WriteRune(c + offset)
				} else if c >= lowerA && c <= lowerZ {
					builder.WriteRune(c - offset)
				} else {
					builder.WriteRune(c)
				}
			}
			return NewLoxString(builder.String(), l.quote), nil
		})
	case "title":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			words := strings.Split(l.str, " ")
			wordsLen := len(words)
			var builder strings.Builder
			for index, word := range words {
				switch utf8.RuneCountInString(word) {
				case 0:
				case 1:
					builder.WriteString(strings.ToUpper(word))
				default:
					runes := []rune(word)
					finalWord := strings.ToUpper(string(runes[0])) + strings.ToLower(string(runes[1:]))
					builder.WriteString(finalWord)
				}
				if index < wordsLen-1 {
					builder.WriteRune(' ')
				}
			}
			return NewLoxString(builder.String(), l.quote), nil
		})
	case "toBuffer":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			b := make([]byte, 4)
			buffer := EmptyLoxBufferCap(int64(len(l.str)))
			for _, r := range l.str {
				for i := 0; i < utf8.EncodeRune(b, r); i++ {
					addErr := buffer.add(int64(b[i]))
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				}
			}
			return buffer, nil
		})
	case "toList":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newList := list.NewListCap[any](int64(utf8.RuneCountInString(l.str)))
			for _, c := range l.str {
				newList.Add(NewLoxStringQuote(string(c)))
			}
			return NewLoxList(newList), nil
		})
	case "toNum":
		return strFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			useParseFloat := func() (any, error) {
				result, resultErr := strconv.ParseFloat(l.str, 64)
				if resultErr != nil {
					return math.NaN(), nil
				}
				if util.FloatIsInt(result) && !strings.Contains(l.str, ".") {
					return int64(result), nil
				}
				return result, nil
			}
			argsLen := len(args)
			switch argsLen {
			case 0:
				return useParseFloat()
			case 1:
				if base, ok := args[0].(int64); ok {
					if base == 10 {
						return useParseFloat()
					}
					bases := map[byte]int64{
						'b': 2,
						'B': 2,
						'o': 8,
						'O': 8,
						'x': 16,
						'X': 16,
					}
					var result int64
					var resultErr error
					if utf8.RuneCountInString(l.str) > 1 {
						if l.str[0] == '0' {
							if bases[l.str[1]] != 0 && (base == bases[l.str[1]] || base == 0) {
								//String starts with 0b, 0B, 0o, 0O, 0x, 0X
								result, resultErr = strconv.ParseInt(l.str, 0, 64)
							} else if base == 0 {
								//Treat strings starting with '0' as base 10 when base argument is 0
								result, resultErr = strconv.ParseInt(l.str, 10, 64)
							} else {
								//Multi digit strings that start with '0'
								result, resultErr = strconv.ParseInt(l.str, int(base), 64)
							}
						} else {
							//Multi digit strings that don't start with '0'
							result, resultErr = strconv.ParseInt(l.str, int(base), 64)
						}
					} else {
						//Single digit strings
						result, resultErr = strconv.ParseInt(l.str, int(base), 64)
					}
					if resultErr != nil {
						return math.NaN(), nil
					}
					return result, nil
				}
				return argMustBeTypeAn("integer")
			}
			return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		})
	case "toSet":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newSet := EmptyLoxSet()
			for _, c := range l.str {
				_, errStr := newSet.add(NewLoxStringQuote(string(c)))
				if len(errStr) > 0 {
					return nil, loxerror.RuntimeError(name, errStr)
				}
			}
			return newSet, nil
		})
	case "unquote":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			lStr := l.str
			arr := []rune(lStr)
			arrLen := len(arr)
			if arrLen > 0 {
				if arr[0] == '`' && arr[arrLen-1] == '`' {
					return l, nil
				}
				if arr[0] == '\'' && arr[arrLen-1] == '\'' {
					arr[0] = '"'
					arr[arrLen-1] = '"'
					lStr = string(arr)
				}
			}
			s, err := strconv.Unquote(lStr)
			if err != nil {
				return l, nil
			}
			return NewLoxStringQuote(s), nil
		})
	case "upper":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxString(strings.ToUpper(l.str), l.quote), nil
		})
	case "zfill":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if finalStrLen, ok := args[0].(int64); ok {
				var paddedStr string
				var useDoubleQuote bool
				var strSign byte = 0
				if len(l.str) > 0 && (l.str[0] == '-' || l.str[0] == '+') {
					strSign = l.str[0]
					paddedStr, useDoubleQuote = padString(l.str[1:], finalStrLen-1, "0", true)
				} else {
					paddedStr, useDoubleQuote = padString(l.str, finalStrLen, "0", true)
				}

				var finalStr string
				var quote byte = '\''
				if useDoubleQuote {
					quote = '"'
				}
				if strSign != 0 {
					var builder strings.Builder
					builder.WriteByte(strSign)
					builder.WriteString(paddedStr)
					finalStr = builder.String()
				} else {
					finalStr = paddedStr
				}
				return NewLoxString(finalStr, quote), nil
			}
			return argMustBeTypeAn("integer")
		})
	}
	return nil, loxerror.RuntimeError(name, "Strings have no property called '"+methodName+"'.")
}

func (l *LoxString) Index(element any) (any, error) {
	switch element := element.(type) {
	case int64:
		return l.IndexInt(element)
	case *big.Int:
		numInt, err := checkValidBigint(element)
		if err != nil {
			return nil, err
		}
		return l.IndexInt(numInt)
	}
	return nil, loxerror.Error(StringIndexMustBeWholeNum(element))
}

func (l *LoxString) IndexSlice(first, second any) (any, error) {
	var firstInt, secondInt int64
	switch first := first.(type) {
	case nil:
		firstInt = 0
	case int64:
		firstInt = first
	case *big.Int:
		var err error
		firstInt, err = checkValidBigint(first)
		if err != nil {
			return nil, err
		}
	default:
		return nil, loxerror.Error(StringIndexMustBeWholeNum(first))
	}
	switch second := second.(type) {
	case nil:
		secondInt = l.Length()
	case int64:
		secondInt = second
	case *big.Int:
		var err error
		secondInt, err = checkValidBigint(second)
		if err != nil {
			return nil, err
		}
	default:
		return nil, loxerror.Error(StringIndexMustBeWholeNum(second))
	}
	return l.IndexIntSlice(firstInt, secondInt)
}

func (l *LoxString) IndexInt(index int64) (any, error) {
	originalIndex := index
	lLen := l.Length()
	index = convertNegIndex(lLen, index)
	if index < 0 || index >= lLen {
		return nil, loxerror.Error(StringIndexOutOfRange(originalIndex))
	}
	str := string([]rune(l.str)[index])
	if str == "'" {
		return NewLoxString(str, '"'), nil
	}
	return NewLoxString(str, '\''), nil
}

func (l *LoxString) IndexIntSlice(first, second int64) (any, error) {
	lLen := l.Length()
	first = max(convertNegIndex(lLen, first), 0)
	second = convertSliceIndex(lLen, second)
	if first > second {
		return EmptyLoxString(), nil
	}
	return NewLoxStringQuote(
		string([]rune(l.str)[first:second]),
	), nil
}

func (l *LoxString) Iterator() interfaces.Iterator {
	return &LoxStringIterator{l, 0}
}

func (l *LoxString) Length() int64 {
	return int64(utf8.RuneCountInString(l.str))
}

func (l *LoxString) ReverseIterator() interfaces.Iterator {
	iterator := ProtoIterator{}
	index := l.Length() - 1
	iterator.hasNextMethod = func() bool {
		return index >= 0
	}
	iterator.nextMethod = func() any {
		c := []rune(l.str)[index]
		index--
		return NewLoxStringQuote(string(c))
	}
	return iterator
}

func (l *LoxString) String() string {
	return l.str
}

func (l *LoxString) Type() string {
	return "string"
}
