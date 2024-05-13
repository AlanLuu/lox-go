package ast

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type LoxString struct {
	str   string
	quote byte
}

func EmptyLoxString() *LoxString {
	return &LoxString{
		str:   "",
		quote: '\'',
	}
}

func StringIndexMustBeWholeNum(index any) string {
	return fmt.Sprintf("String index '%v' must be a whole number.", index)
}

func StringIndexOutOfRange(index int64) string {
	return fmt.Sprintf("String index %v out of range.", index)
}

func (l *LoxString) NewLoxString(str string) *LoxString {
	return &LoxString{
		str:   str,
		quote: l.quote,
	}
}

func (l *LoxString) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxString:
		return l.str == obj.str
	default:
		return false
	}
}

func (l *LoxString) Get(name token.Token) (any, error) {
	strFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (struct{ ProtoLoxCallable }, error) {
		s := struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string { return "<native string fn>" }
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
			padStr = getResult(arg, true)
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
	methodName := name.Lexeme
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'string.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
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
	case "endsWith":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return strings.HasSuffix(l.str, loxStr.str), nil
			}
			return argMustBeType("string")
		})
	case "index":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return int64(strings.Index(l.str, loxStr.str)), nil
			}
			return argMustBeType("string")
		})
	case "lower":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return &LoxString{strings.ToLower(l.str), l.quote}, nil
		})
	case "padEnd":
		return strFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if finalStrLen, ok := args[0].(int64); ok {
				paddedStr, useDoubleQuote := padString(l.str, finalStrLen, args[1], false)
				if useDoubleQuote {
					return &LoxString{paddedStr, '"'}, nil
				}
				return &LoxString{paddedStr, l.quote}, nil
			}
			return nil, loxerror.RuntimeError(name, "First argument to 'string.padEnd' must be a whole number.")
		})
	case "padStart":
		return strFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if finalStrLen, ok := args[0].(int64); ok {
				paddedStr, useDoubleQuote := padString(l.str, finalStrLen, args[1], true)
				if useDoubleQuote {
					return &LoxString{paddedStr, '"'}, nil
				}
				return &LoxString{paddedStr, l.quote}, nil
			}
			return nil, loxerror.RuntimeError(name, "First argument to 'string.padStart' must be a whole number.")
		})
	case "replace":
		return strFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if firstStr, firstStrOk := args[0].(*LoxString); firstStrOk {
				if secondStr, secondStrOk := args[1].(*LoxString); secondStrOk {
					newStr := strings.ReplaceAll(l.str, firstStr.str, secondStr.str)
					if strings.Contains(newStr, "'") {
						return &LoxString{newStr, '"'}, nil
					}
					return &LoxString{newStr, '\''}, nil
				}
				return nil, loxerror.RuntimeError(name, "Second argument to 'string.replace' must be a string.")
			}
			return nil, loxerror.RuntimeError(name, "First argument to 'string.replace' must be a string.")
		})
	case "split":
		return strFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				splitSlice := strings.Split(l.str, loxStr.str)
				loxList := list.NewList[Expr]()
				for _, str := range splitSlice {
					loxList.Add(&LoxString{str, '\''})
				}
				return &LoxList{loxList}, nil
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
				return &LoxString{strings.TrimSpace(l.str), l.quote}, nil
			case 1:
				if loxStr, ok := args[0].(*LoxString); ok {
					return &LoxString{strings.Trim(l.str, loxStr.str), l.quote}, nil
				}
				return argMustBeType("string")
			}
			return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		})
	case "toNum":
		return strFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			useParseFloat := func() (any, error) {
				result, resultErr := strconv.ParseFloat(l.str, 64)
				if resultErr != nil {
					return math.NaN(), nil
				}
				if util.FloatIsInt(result) {
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
					if len(l.str) > 1 {
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
				return argMustBeType("whole number")
			}
			return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		})
	case "upper":
		return strFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return &LoxString{strings.ToUpper(l.str), l.quote}, nil
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
				return &LoxString{finalStr, quote}, nil
			}
			return argMustBeType("whole number")
		})
	}
	return nil, loxerror.RuntimeError(name, "Strings have no property called '"+methodName+"'.")
}

func (l *LoxString) String() string {
	return l.str
}
