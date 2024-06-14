package ast

import (
	"encoding/base64"
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineBase64Funcs() {
	className := "base64"
	base64Class := NewLoxClass(className, nil, false)
	base64Func := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native base64 fn %v at %p>", name, &s)
		}
		base64Class.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'base64.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	base64Func("decode", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			result, decodeErr := base64.StdEncoding.DecodeString(loxStr.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
			return NewLoxStringQuote(string(result)), nil
		}
		return argMustBeType(in.callToken, "decode", "string")
	})
	base64Func("decodeToBuf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			result, decodeErr := base64.StdEncoding.DecodeString(loxStr.str)
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
	base64Func("encode", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *LoxString:
			encodedStr := base64.StdEncoding.EncodeToString([]byte(arg.str))
			return NewLoxString(encodedStr, '\''), nil
		case *LoxBuffer:
			byteList := list.NewList[byte]()
			for _, element := range arg.elements {
				byteList.Add(byte(element.(int64)))
			}
			encodedStr := base64.StdEncoding.EncodeToString([]byte(byteList))
			return NewLoxString(encodedStr, '\''), nil
		}
		return argMustBeType(in.callToken, "encode", "string or buffer")
	})

	i.globals.Define(className, base64Class)
}
