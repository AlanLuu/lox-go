package ast

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"golang.org/x/crypto/bcrypt"
)

func (i *Interpreter) defineCryptoFuncs() {
	className := "crypto"
	cryptoClass := NewLoxClass(className, nil, false)
	cryptoFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native crypto fn %v at %p>", name, &s)
		}
		cryptoClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'crypto.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	cryptoFunc("bcrypt", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var password []byte
		var cost int
		argsLen := len(args)
		switch argsLen {
		case 1:
			if _, ok := args[0].(*LoxString); !ok {
				return argMustBeType(in.callToken, "bcrypt", "string")
			}
			password = []byte(args[0].(*LoxString).str)
			cost = bcrypt.DefaultCost
		case 2:
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'crypto.bcrypt' must be a string.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'crypto.bcrypt' must be an integer.")
			}
			password = []byte(args[0].(*LoxString).str)
			cost = int(args[1].(int64))
		}
		hash, hashErr := bcrypt.GenerateFromPassword(password, cost)
		if hashErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, hashErr.Error())
		}
		return NewLoxString(string(hash), '\''), nil
	})
	cryptoFunc("bcryptVerify", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'crypto.bcryptVerify' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'crypto.bcryptVerify' must be a string.")
		}
		password := []byte(args[0].(*LoxString).str)
		hash := []byte(args[1].(*LoxString).str)
		return bcrypt.CompareHashAndPassword(hash, password) == nil, nil
	})
	cryptoFunc("md5", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		argsLen := len(args)
		switch argsLen {
		case 0:
			hashObj = md5.New()
		case 1:
			switch arg := args[0].(type) {
			case *LoxBuffer:
				hashObj = md5.New()
				bytes := []byte{}
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				hashObj.Write(bytes)
			case *LoxString:
				hashObj = md5.New()
				hashObj.Write([]byte(arg.str))
			default:
				return argMustBeType(in.callToken, "md5", "buffer or string")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		return NewLoxHash(hashObj, "md5"), nil
	})
	cryptoFunc("sha1", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		argsLen := len(args)
		switch argsLen {
		case 0:
			hashObj = sha1.New()
		case 1:
			switch arg := args[0].(type) {
			case *LoxBuffer:
				hashObj = sha1.New()
				bytes := []byte{}
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				hashObj.Write(bytes)
			case *LoxString:
				hashObj = sha1.New()
				hashObj.Write([]byte(arg.str))
			default:
				return argMustBeType(in.callToken, "sha1", "buffer or string")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		return NewLoxHash(hashObj, "sha1"), nil
	})
	cryptoFunc("sha224", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		argsLen := len(args)
		switch argsLen {
		case 0:
			hashObj = sha256.New224()
		case 1:
			switch arg := args[0].(type) {
			case *LoxBuffer:
				hashObj = sha256.New224()
				bytes := []byte{}
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				hashObj.Write(bytes)
			case *LoxString:
				hashObj = sha256.New224()
				hashObj.Write([]byte(arg.str))
			default:
				return argMustBeType(in.callToken, "sha224", "buffer or string")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		return NewLoxHash(hashObj, "sha224"), nil
	})
	cryptoFunc("sha256", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		argsLen := len(args)
		switch argsLen {
		case 0:
			hashObj = sha256.New()
		case 1:
			switch arg := args[0].(type) {
			case *LoxBuffer:
				hashObj = sha256.New()
				bytes := []byte{}
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				hashObj.Write(bytes)
			case *LoxString:
				hashObj = sha256.New()
				hashObj.Write([]byte(arg.str))
			default:
				return argMustBeType(in.callToken, "sha256", "buffer or string")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		return NewLoxHash(hashObj, "sha256"), nil
	})
	cryptoFunc("sha384", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		argsLen := len(args)
		switch argsLen {
		case 0:
			hashObj = sha512.New384()
		case 1:
			switch arg := args[0].(type) {
			case *LoxBuffer:
				hashObj = sha512.New384()
				bytes := []byte{}
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				hashObj.Write(bytes)
			case *LoxString:
				hashObj = sha512.New384()
				hashObj.Write([]byte(arg.str))
			default:
				return argMustBeType(in.callToken, "sha384", "buffer or string")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		return NewLoxHash(hashObj, "sha384"), nil
	})
	cryptoFunc("sha512", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		argsLen := len(args)
		switch argsLen {
		case 0:
			hashObj = sha512.New()
		case 1:
			switch arg := args[0].(type) {
			case *LoxBuffer:
				hashObj = sha512.New()
				bytes := []byte{}
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				hashObj.Write(bytes)
			case *LoxString:
				hashObj = sha512.New()
				hashObj.Write([]byte(arg.str))
			default:
				return argMustBeType(in.callToken, "sha512", "buffer or string")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		return NewLoxHash(hashObj, "sha512"), nil
	})

	i.globals.Define(className, cryptoClass)
}
