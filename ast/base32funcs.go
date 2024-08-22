package ast

import (
	"encoding/base32"
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineBase32Funcs() {
	className := "base32"
	base32Class := NewLoxClass(className, nil, false)
	base32Func := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native base32 fn %v at %p>", name, &s)
		}
		base32Class.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'base32.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	base32Func("decode", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			result, decodeErr := base32.StdEncoding.DecodeString(loxStr.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
			return NewLoxStringQuote(string(result)), nil
		}
		return argMustBeType(in.callToken, "decode", "string")
	})
	base32Func("decodeToBuf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			result, decodeErr := base32.StdEncoding.DecodeString(loxStr.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
			buffer := EmptyLoxBuffer()
			for _, value := range result {
				addErr := buffer.add(int64(value))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
			return buffer, nil
		}
		return argMustBeType(in.callToken, "decodeToBuf", "string")
	})
	base32Func("encode", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *LoxString:
			encodedStr := base32.StdEncoding.EncodeToString([]byte(arg.str))
			return NewLoxString(encodedStr, '\''), nil
		case *LoxBuffer:
			byteList := list.NewListCap[byte](int64(len(arg.elements)))
			for _, element := range arg.elements {
				byteList.Add(byte(element.(int64)))
			}
			encodedStr := base32.StdEncoding.EncodeToString([]byte(byteList))
			return NewLoxString(encodedStr, '\''), nil
		}
		return argMustBeType(in.callToken, "encode", "string or buffer")
	})

	i.globals.Define(className, base32Class)
}
