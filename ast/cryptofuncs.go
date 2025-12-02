package ast

import (
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	crand "crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"math/big"

	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var LoxCryptoHashes = map[string]crypto.Hash{
	"md5":    crypto.MD5,
	"sha1":   crypto.SHA1,
	"sha224": crypto.SHA224,
	"sha256": crypto.SHA256,
	"sha384": crypto.SHA384,
	"sha512": crypto.SHA512,
}

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
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'crypto.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	cryptoFunc("aescbc", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var aesCBC *LoxAESCBC
		var err error
		switch arg := args[0].(type) {
		case int64:
			aesCBC, err = NewLoxAESCBC(int(arg))
		case *LoxBuffer:
			keyBytes := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				keyBytes = append(keyBytes, byte(element.(int64)))
			}
			aesCBC, err = NewLoxAESCBCBytes(keyBytes)
		case *LoxString:
			keyBytes, decodeErr := LoxAESDecode(arg.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
			aesCBC, err = NewLoxAESCBCBytes(keyBytes)
		default:
			return argMustBeType(in.callToken, "aescbc", "buffer, integer, or string")
		}
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken,
				"crypto.aescbc: "+err.Error())
		}
		return aesCBC, nil
	})
	cryptoFunc("aescbchex", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			keyBytes, decodeErr := hex.DecodeString(loxStr.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
			aesCBC, err := NewLoxAESCBCBytes(keyBytes)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken,
					"crypto.aescbchex: "+err.Error())
			}
			return aesCBC, nil
		}
		return argMustBeType(in.callToken, "aescbchex", "string")
	})
	cryptoFunc("aescfb", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var aesCFB *LoxAESCFB
		var err error
		switch arg := args[0].(type) {
		case int64:
			aesCFB, err = NewLoxAESCFB(int(arg))
		case *LoxBuffer:
			keyBytes := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				keyBytes = append(keyBytes, byte(element.(int64)))
			}
			aesCFB, err = NewLoxAESCFBBytes(keyBytes)
		case *LoxString:
			keyBytes, decodeErr := LoxAESDecode(arg.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
			aesCFB, err = NewLoxAESCFBBytes(keyBytes)
		default:
			return argMustBeType(in.callToken, "aescfb", "buffer, integer, or string")
		}
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken,
				"crypto.aescfb: "+err.Error())
		}
		return aesCFB, nil
	})
	cryptoFunc("aescfbhex", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			keyBytes, decodeErr := hex.DecodeString(loxStr.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
			aesCFB, err := NewLoxAESCFBBytes(keyBytes)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken,
					"crypto.aescfbhex: "+err.Error())
			}
			return aesCFB, nil
		}
		return argMustBeType(in.callToken, "aescfbhex", "string")
	})
	cryptoFunc("ageasym", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var result *LoxAgeAsymmetric
		var err error
		argsLen := len(args)
		switch argsLen {
		case 0:
			result, err = NewLoxAgeAsymmetric()
		case 1:
			if loxStr, ok := args[0].(*LoxString); ok {
				result, err = NewLoxAgeAsymmetricPrivKeyStr(loxStr.str)
			} else {
				return argMustBeType(in.callToken, "ageasym", "string")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return result, nil
	})
	cryptoFunc("ageasympub", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			result, err := NewLoxAgeAsymmetricPubKeyStr(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return result, nil
		}
		return argMustBeType(in.callToken, "ageasympub", "string")
	})
	cryptoFunc("agesym", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		switch argsLen {
		case 0:
			return NewLoxAgeSymmetric(), nil
		case 1:
			if loxStr, ok := args[0].(*LoxString); ok {
				if len(loxStr.str) == 0 {
					return nil, loxerror.RuntimeError(in.callToken,
						"String argument to 'crypto.agesym' cannot be empty.")
				}
				return NewLoxAgeSymmetricPassword(loxStr.str), nil
			}
			return argMustBeType(in.callToken, "agesym", "string")
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
	})
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
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
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
	cryptoFunc("ed25519", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		keyPair, err := NewLoxEd25519()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return keyPair, nil
	})
	cryptoFunc("ed25519priv", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var privKeyBytes []byte
		switch arg := args[0].(type) {
		case *LoxBuffer:
			privKeyBytes = make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				privKeyBytes = append(privKeyBytes, byte(element.(int64)))
			}
		case *LoxString:
			var decodeErr error
			privKeyBytes, decodeErr = LoxEd25519Decode(arg.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
		default:
			return argMustBeType(in.callToken, "ed25519priv", "buffer or string")
		}
		privKey, err := NewLoxEd25519PrivKey(privKeyBytes)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return privKey, nil
	})
	cryptoFunc("ed25519pub", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var pubKeyBytes []byte
		switch arg := args[0].(type) {
		case *LoxBuffer:
			pubKeyBytes = make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				pubKeyBytes = append(pubKeyBytes, byte(element.(int64)))
			}
		case *LoxString:
			var decodeErr error
			pubKeyBytes, decodeErr = LoxEd25519Decode(arg.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
		default:
			return argMustBeType(in.callToken, "ed25519pub", "buffer or string")
		}
		pubKey, err := NewLoxEd25519PubKey(pubKeyBytes)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return pubKey, nil
	})
	cryptoFunc("ed25519seed", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var seedBytes []byte
		switch arg := args[0].(type) {
		case *LoxBuffer:
			seedBytes = make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				seedBytes = append(seedBytes, byte(element.(int64)))
			}
		case *LoxString:
			var decodeErr error
			seedBytes, decodeErr = LoxEd25519Decode(arg.str)
			if decodeErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, decodeErr.Error())
			}
		default:
			return argMustBeType(in.callToken, "ed25519seed", "buffer or string")
		}
		keyPair, err := NewLoxEd25519PrivKeySeed(seedBytes)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return keyPair, nil
	})
	cryptoFunc("fernet", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var fernet *LoxFernet
		var err error
		argsLen := len(args)
		switch argsLen {
		case 0:
			fernet, err = NewLoxFernet()
		case 1:
			switch arg := args[0].(type) {
			case *LoxBuffer:
				if len(arg.elements) != 32 {
					return nil, loxerror.RuntimeError(in.callToken,
						"Buffer argument to 'crypto.fernet' must be of length 32.")
				}
				key := [32]byte{}
				for i := 0; i < 32; i++ {
					key[i] = byte(arg.elements[i].(int64))
				}
				fernet = NewLoxFernetFromBytes(key)
			case *LoxString:
				fernet, err = NewLoxFernetFromString(arg.str)
			default:
				return argMustBeType(in.callToken, "fernet", "buffer or string")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return fernet, nil
	})
	cryptoFunc("flip", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		randInt, err := crand.Int(crand.Reader, bigint.Two)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return randInt.Int64() == 0, nil
	})
	cryptoFunc("hmac", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		argZeroErrMsg := "First argument to 'crypto.hmac' must be a function."
		argOneErrMsg := "Second argument to 'crypto.hmac' must be a buffer or string."
		switch args[0].(type) {
		case *LoxClass:
			return nil, loxerror.RuntimeError(in.callToken, argZeroErrMsg)
		case LoxCallable:
		default:
			return nil, loxerror.RuntimeError(in.callToken, argZeroErrMsg)
		}
		switch args[1].(type) {
		case *LoxBuffer:
		case *LoxString:
		default:
			return nil, loxerror.RuntimeError(in.callToken, argOneErrMsg)
		}

		callable := args[0].(LoxCallable)
		var result any
		switch callable := callable.(type) {
		case *LoxFunction:
			argList := getArgList(callable, 0)
			callResult, resultErr := callable.call(i, argList)
			if callresultReturn, ok := callResult.(Return); ok {
				result = callresultReturn.FinalValue
			} else if resultErr != nil {
				return nil, resultErr
			}
		default:
			var resultErr error
			result, resultErr = callable.call(i, list.NewList[any]())
			if resultErr != nil {
				return nil, resultErr
			}
		}

		switch result := result.(type) {
		case *LoxHash:
			if result.hashFunc == nil {
				return nil, loxerror.RuntimeError(in.callToken,
					"Function argument to 'crypto.hmac' returned unknown hash type.")
			}
			var key []byte
			switch arg := args[1].(type) {
			case *LoxBuffer:
				key = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					key = append(key, byte(element.(int64)))
				}
			case *LoxString:
				key = []byte(arg.str)
			}
			hashObj := hmac.New(result.hashFunc, key)
			return NewLoxHash(hashObj, result.hashFunc, "hmac-"+result.hashType), nil
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"Function argument to 'crypto.hmac' must return a hash object.")
		}
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
				bytes := make([]byte, 0, len(arg.elements))
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
		return NewLoxHash(hashObj, md5.New, "md5"), nil
	})
	cryptoFunc("md5sum", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		switch arg := args[0].(type) {
		case *LoxBuffer:
			hashObj = md5.New()
			bytes := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				bytes = append(bytes, byte(element.(int64)))
			}
			hashObj.Write(bytes)
		case *LoxString:
			hashObj = md5.New()
			hashObj.Write([]byte(arg.str))
		default:
			return argMustBeType(in.callToken, "md5sum", "buffer or string")
		}
		hexDigest := fmt.Sprintf("%x", hashObj.Sum(nil))
		return NewLoxString(hexDigest, '\''), nil
	})
	cryptoFunc("prime", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if numBits, ok := args[0].(int64); ok {
			if numBits < 2 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Integer argument to 'crypto.prime' must be an integer >= 2.")
			}
			prime, err := crand.Prime(crand.Reader, int(numBits))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return prime, nil
		}
		return argMustBeTypeAn(in.callToken, "prime", "integer")
	})
	cryptoFunc("randBigInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		const errMsg = "argument to 'crypto.randBigInt' cannot be 0 or negative."
		var max *big.Int
		switch arg := args[0].(type) {
		case int64:
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Integer "+errMsg)
			}
			max = big.NewInt(arg)
		case *big.Int:
			if arg.Cmp(bigint.Zero) <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Bigint "+errMsg)
			}
			max = new(big.Int).Set(arg)
		default:
			return argMustBeTypeAn(in.callToken, "randBigInt", "integer or bigint")
		}
		result, err := crand.Int(crand.Reader, max)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return result, nil
	})
	cryptoFunc("randBigInts", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		switch args[0].(type) {
		case int64:
		case *big.Int:
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'crypto.randBigInts' must be an integer or bigint.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'crypto.randBigInts' must be an integer.")
		}
		const errMsg = "argument to 'crypto.randBigInts' cannot be 0 or negative."
		var max *big.Int
		switch arg := args[0].(type) {
		case int64:
			if arg <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"First integer "+errMsg)
			}
			max = big.NewInt(arg)
		case *big.Int:
			if arg.Cmp(bigint.Zero) <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"First bigint "+errMsg)
			}
			max = new(big.Int).Set(arg)
		}
		times := args[1].(int64)
		if times < 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'crypto.randBigInts' cannot be negative.")
		}
		nums := list.NewListCap[any](times)
		for i := int64(0); i < times; i++ {
			result, err := crand.Int(crand.Reader, max)
			if err != nil {
				nums.Clear()
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			nums.Add(result)
		}
		return NewLoxList(nums), nil
	})
	cryptoFunc("randInt", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if max, ok := args[0].(int64); ok {
			if max <= 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'crypto.randInt' cannot be 0 or negative.")
			}
			result, err := crand.Int(crand.Reader, big.NewInt(max))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return result.Int64(), nil
		}
		return argMustBeTypeAn(in.callToken, "randInt", "integer")
	})
	for _, s := range []string{"randomUUID", "randUUID"} {
		cryptoFunc(s, 0, func(in *Interpreter, _ list.List[any]) (any, error) {
			randUUID, err := uuid.NewRandom()
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxString(randUUID.String(), '\''), nil
		})
	}
	cryptoFunc("rsa", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if bitSize, ok := args[0].(int64); ok {
			keyPair, err := NewLoxRSA(int(bitSize))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return keyPair, nil
		}
		return argMustBeType(in.callToken, "rsa", "integer")
	})
	cryptoFunc("rsapriv", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var keyPair *LoxRSA
		var err error
		switch arg := args[0].(type) {
		case *LoxBuffer:
			privKeyBytes := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				privKeyBytes = append(privKeyBytes, byte(element.(int64)))
			}
			keyPair, err = NewLoxRSAPrivKeyBytes(privKeyBytes)
		case *LoxString:
			keyPair, err = NewLoxRSAPrivKeyStr(arg.str)
		default:
			return argMustBeType(in.callToken, "rsapriv", "buffer or string")
		}
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return keyPair, nil
	})
	cryptoFunc("rsapub", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var keyPair *LoxRSA
		var err error
		switch arg := args[0].(type) {
		case *LoxBuffer:
			privKeyBytes := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				privKeyBytes = append(privKeyBytes, byte(element.(int64)))
			}
			keyPair, err = NewLoxRSAPubKeyBytes(privKeyBytes)
		case *LoxString:
			keyPair, err = NewLoxRSAPubKeyStr(arg.str)
		default:
			return argMustBeType(in.callToken, "rsapub", "buffer or string")
		}
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return keyPair, nil
	})
	cryptoFunc("rsapubne", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*big.Int); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'crypto.rsapubne' must be a bigint.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'crypto.rsapubne' must be an integer.")
		}
		N := args[0].(*big.Int)
		E := int(args[1].(int64))
		return NewLoxRSAPubKey(N, E), nil
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
				bytes := make([]byte, 0, len(arg.elements))
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
		return NewLoxHash(hashObj, sha1.New, "sha1"), nil
	})
	cryptoFunc("sha1sum", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		switch arg := args[0].(type) {
		case *LoxBuffer:
			hashObj = sha1.New()
			bytes := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				bytes = append(bytes, byte(element.(int64)))
			}
			hashObj.Write(bytes)
		case *LoxString:
			hashObj = sha1.New()
			hashObj.Write([]byte(arg.str))
		default:
			return argMustBeType(in.callToken, "sha1sum", "buffer or string")
		}
		hexDigest := fmt.Sprintf("%x", hashObj.Sum(nil))
		return NewLoxString(hexDigest, '\''), nil
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
				bytes := make([]byte, 0, len(arg.elements))
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
		return NewLoxHash(hashObj, sha256.New224, "sha224"), nil
	})
	cryptoFunc("sha224sum", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		switch arg := args[0].(type) {
		case *LoxBuffer:
			hashObj = sha256.New224()
			bytes := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				bytes = append(bytes, byte(element.(int64)))
			}
			hashObj.Write(bytes)
		case *LoxString:
			hashObj = sha256.New224()
			hashObj.Write([]byte(arg.str))
		default:
			return argMustBeType(in.callToken, "sha224sum", "buffer or string")
		}
		hexDigest := fmt.Sprintf("%x", hashObj.Sum(nil))
		return NewLoxString(hexDigest, '\''), nil
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
				bytes := make([]byte, 0, len(arg.elements))
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
		return NewLoxHash(hashObj, sha256.New, "sha256"), nil
	})
	cryptoFunc("sha256sum", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		switch arg := args[0].(type) {
		case *LoxBuffer:
			hashObj = sha256.New()
			bytes := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				bytes = append(bytes, byte(element.(int64)))
			}
			hashObj.Write(bytes)
		case *LoxString:
			hashObj = sha256.New()
			hashObj.Write([]byte(arg.str))
		default:
			return argMustBeType(in.callToken, "sha256sum", "buffer or string")
		}
		hexDigest := fmt.Sprintf("%x", hashObj.Sum(nil))
		return NewLoxString(hexDigest, '\''), nil
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
				bytes := make([]byte, 0, len(arg.elements))
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
		return NewLoxHash(hashObj, sha512.New384, "sha384"), nil
	})
	cryptoFunc("sha384sum", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		switch arg := args[0].(type) {
		case *LoxBuffer:
			hashObj = sha512.New384()
			bytes := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				bytes = append(bytes, byte(element.(int64)))
			}
			hashObj.Write(bytes)
		case *LoxString:
			hashObj = sha512.New384()
			hashObj.Write([]byte(arg.str))
		default:
			return argMustBeType(in.callToken, "sha384sum", "buffer or string")
		}
		hexDigest := fmt.Sprintf("%x", hashObj.Sum(nil))
		return NewLoxString(hexDigest, '\''), nil
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
				bytes := make([]byte, 0, len(arg.elements))
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
		return NewLoxHash(hashObj, sha512.New, "sha512"), nil
	})
	cryptoFunc("sha512sum", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var hashObj hash.Hash
		switch arg := args[0].(type) {
		case *LoxBuffer:
			hashObj = sha512.New()
			bytes := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				bytes = append(bytes, byte(element.(int64)))
			}
			hashObj.Write(bytes)
		case *LoxString:
			hashObj = sha512.New()
			hashObj.Write([]byte(arg.str))
		default:
			return argMustBeType(in.callToken, "sha512sum", "buffer or string")
		}
		hexDigest := fmt.Sprintf("%x", hashObj.Sum(nil))
		return NewLoxString(hexDigest, '\''), nil
	})

	i.globals.Define(className, cryptoClass)
}
