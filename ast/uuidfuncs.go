package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/google/uuid"
)

func defineUUIDFields(uuidClass *LoxClass) {
	specialUUIDs := map[string]uuid.UUID{
		"dns":  uuid.NameSpaceDNS,
		"url":  uuid.NameSpaceURL,
		"oid":  uuid.NameSpaceOID,
		"null": uuid.Nil,
		"max":  uuid.Max,
	}
	for key, value := range specialUUIDs {
		uuidClass.classProperties[key] = NewLoxUUID(value)
	}

	uuidVariants := map[string]uuid.Variant{
		"invalid":   uuid.Invalid,
		"rfc4122":   uuid.RFC4122,
		"reserved":  uuid.Reserved,
		"microsoft": uuid.Microsoft,
		"future":    uuid.Future,
	}
	for key, value := range uuidVariants {
		uuidClass.classProperties[key] = int64(value)
	}
}

func (i *Interpreter) defineUUIDFuncs() {
	className := "UUID"
	uuidClass := NewLoxClass(className, nil, false)
	uuidFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native UUID fn %v at %p>", name, &s)
		}
		uuidClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'UUID.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	defineUUIDFields(uuidClass)
	uuidFunc("disableRandPool", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		uuid.DisableRandPool()
		return nil, nil
	})
	uuidFunc("enableRandPool", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		uuid.EnableRandPool()
		return nil, nil
	})
	uuidFunc("fromBytes", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if buffer, ok := args[0].(*LoxBuffer); ok {
			bytes := make([]byte, 0, len(buffer.elements))
			for _, element := range buffer.elements {
				bytes = append(bytes, byte(element.(int64)))
			}
			newUUID, err := uuid.FromBytes(bytes)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxUUID(newUUID), nil
		}
		return argMustBeType(in.callToken, "fromBytes", "buffer")
	})
	uuidFunc("mustValidate", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			err := uuid.Validate(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "mustValidate", "string")
	})
	uuidFunc("new", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		randUUID, err := NewLoxUUIDV4Random()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return randUUID, nil
	})
	uuidFunc("newV1", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		v1UUID, err := uuid.NewUUID()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxUUID(v1UUID), nil
	})
	uuidFunc("newV4", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		randUUID, err := NewLoxUUIDV4Random()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return randUUID, nil
	})
	uuidFunc("newV6", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		v6UUID, err := uuid.NewV6()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxUUID(v6UUID), nil
	})
	uuidFunc("newV7", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		v7UUID, err := uuid.NewV7()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxUUID(v7UUID), nil
	})
	uuidFunc("parse", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			newUUID, err := NewLoxUUIDParse(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return newUUID, nil
		}
		return argMustBeType(in.callToken, "parse", "string")
	})
	uuidFunc("parseBytes", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if buffer, ok := args[0].(*LoxBuffer); ok {
			bytes := make([]byte, 0, len(buffer.elements))
			for _, element := range buffer.elements {
				bytes = append(bytes, byte(element.(int64)))
			}
			newUUID, err := uuid.ParseBytes(bytes)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxUUID(newUUID), nil
		}
		return argMustBeType(in.callToken, "parseBytes", "buffer")
	})
	uuidFunc("validate", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return uuid.Validate(loxStr.str) == nil, nil
		}
		return argMustBeType(in.callToken, "validate", "string")
	})

	i.globals.Define(className, uuidClass)
}
