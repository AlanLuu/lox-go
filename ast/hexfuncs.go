package ast

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"

	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineHexFuncs() {
	className := "hexstr"
	hexClass := NewLoxClass(className, nil, false)
	hexFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native hexstr fn %v at %p>", name, &s)
		}
		hexClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'hexstr.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	hexFunc("decode", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			result, decodeErr := hex.DecodeString(loxStr.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
			buffer := EmptyLoxBufferCap(int64(len(result)))
			for _, value := range result {
				addErr := buffer.add(int64(value))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
			return buffer, nil
		}
		return argMustBeType(in.callToken, "decode", "string")
	})
	hexFunc("decodeToStr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			result, decodeErr := hex.DecodeString(loxStr.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
			return NewLoxStringQuote(string(result)), nil
		}
		return argMustBeType(in.callToken, "decodeToStr", "string")
	})
	hexFunc("dump", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *LoxString:
			hexDump := hex.Dump([]byte(arg.str))
			return NewLoxStringQuote(hexDump), nil
		case *LoxBuffer:
			byteList := list.NewListCapDouble[byte](int64(len(arg.elements)))
			for _, element := range arg.elements {
				byteList.Add(byte(element.(int64)))
			}
			hexDump := hex.Dump([]byte(byteList))
			return NewLoxStringQuote(hexDump), nil
		}
		return argMustBeType(in.callToken, "dump", "string or buffer")
	})
	hexFunc("encode", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *LoxString:
			hexStr := hex.EncodeToString([]byte(arg.str))
			return NewLoxString(hexStr, '\''), nil
		case *LoxBuffer:
			byteList := list.NewListCapDouble[byte](int64(len(arg.elements)))
			for _, element := range arg.elements {
				byteList.Add(byte(element.(int64)))
			}
			hexStr := hex.EncodeToString([]byte(byteList))
			return NewLoxString(hexStr, '\''), nil
		}
		return argMustBeType(in.callToken, "encode", "string or buffer")
	})
	hexFunc("tobigint", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			matched, _ := regexp.MatchString("^[0-9a-fA-F]+$", loxStr.str)
			if !matched {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument string must only contain 0-9, a-f, and A-F.")
			}
			hexMapping := map[rune]int64{
				'0': 0,
				'1': 1,
				'2': 2,
				'3': 3,
				'4': 4,
				'5': 5,
				'6': 6,
				'7': 7,
				'8': 8,
				'9': 9,
				'a': 10,
				'A': 10,
				'b': 11,
				'B': 11,
				'c': 12,
				'C': 12,
				'd': 13,
				'D': 13,
				'e': 14,
				'E': 14,
				'f': 15,
				'F': 15,
			}
			runes := []rune(loxStr.str)
			bigInt := big.NewInt(0)
			j := big.NewInt(0)
			value := big.NewInt(0)
			sixteen := big.NewInt(16)
			for i := len(runes) - 1; i >= 0; i-- {
				value.SetInt64(hexMapping[runes[i]])
				sixteen.Exp(sixteen, j, nil)
				sixteen.Mul(sixteen, value)
				bigInt.Add(bigInt, sixteen)
				sixteen.SetInt64(16)
				j.Add(j, bigint.One)
			}
			return bigInt, nil
		}
		return argMustBeType(in.callToken, "tobigint", "string")
	})

	i.globals.Define(className, hexClass)
}
