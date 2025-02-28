package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/list"
)

func defineStringFields(stringClass *LoxClass) {
	digits := "0123456789"
	stringClass.classProperties["digits"] = NewLoxString(digits, '\'')

	hexDigits := "0123456789abcdefABCDEF"
	stringClass.classProperties["hexDigits"] = NewLoxString(hexDigits, '\'')

	hexDigitsLower := "0123456789abcdef"
	stringClass.classProperties["hexDigitsLower"] = NewLoxString(hexDigitsLower, '\'')

	hexDigitsUpper := "0123456789ABCDEF"
	stringClass.classProperties["hexDigitsUpper"] = NewLoxString(hexDigitsUpper, '\'')

	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	stringClass.classProperties["letters"] = NewLoxString(letters, '\'')

	lowercase := "abcdefghijklmnopqrstuvwxyz"
	stringClass.classProperties["lowercase"] = NewLoxString(lowercase, '\'')

	octDigits := "01234567"
	stringClass.classProperties["octDigits"] = NewLoxString(octDigits, '\'')

	punctuation := "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	stringClass.classProperties["punctuation"] = NewLoxString(punctuation, '\'')

	qwertyLower := "qwertyuiopasdfghjklzxcvbnm"
	stringClass.classProperties["qwertyLower"] = NewLoxString(qwertyLower, '\'')

	qwertyUpper := "QWERTYUIOPASDFGHJKLZXCVBNM"
	stringClass.classProperties["qwertyUpper"] = NewLoxString(qwertyUpper, '\'')

	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	stringClass.classProperties["uppercase"] = NewLoxString(uppercase, '\'')
}

func (i *Interpreter) defineStringFuncs() {
	className := "String"
	stringClass := NewLoxClass(className, nil, false)
	stringFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native String class fn %v at %p>", name, &s)
		}
		stringClass.classProperties[name] = s
	}

	defineStringFields(stringClass)
	stringFunc("builder", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxStringBuilder(), nil
	})
	stringFunc("toString", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		var str string
		switch arg := args[0].(type) {
		case *LoxString:
			str = arg.str
		case fmt.Stringer:
			str = arg.String()
		default:
			str = fmt.Sprint(arg)
		}
		return NewLoxStringQuote(str), nil
	})

	i.globals.Define(className, stringClass)
}
