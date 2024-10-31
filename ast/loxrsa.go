package ast

import (
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"golang.org/x/crypto/ssh"
)

func LoxRSADecode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

func LoxRSAEncode(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

type LoxRSA struct {
	privKey     *rsa.PrivateKey
	pubKey      rsa.PublicKey
	bitSize     int
	precomputed bool
	methods     map[string]*struct{ ProtoLoxCallable }
}

func NewLoxRSA(bitSize int) (*LoxRSA, error) {
	privKey, err := rsa.GenerateKey(crand.Reader, bitSize)
	if err != nil {
		return nil, err
	}
	return &LoxRSA{
		privKey:     privKey,
		pubKey:      privKey.PublicKey,
		bitSize:     bitSize,
		precomputed: false,
		methods:     make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func NewLoxRSAPrivKeyBytes(bytes []byte) (*LoxRSA, error) {
	privKey, err := x509.ParsePKCS1PrivateKey(bytes)
	if err != nil {
		privKey2, err2 := x509.ParsePKCS8PrivateKey(bytes)
		if err2 != nil {
			return nil, err2
		}
		privKey = privKey2.(*rsa.PrivateKey)
	}
	return &LoxRSA{
		privKey:     privKey,
		pubKey:      privKey.PublicKey,
		bitSize:     privKey.N.BitLen(),
		precomputed: false,
		methods:     make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func NewLoxRSAPrivKeyStr(str string) (*LoxRSA, error) {
	decodedBytes, err := LoxRSADecode(str)
	if err != nil {
		return nil, err
	}
	return NewLoxRSAPrivKeyBytes(decodedBytes)
}

func NewLoxRSAPubKey(N *big.Int, E int) *LoxRSA {
	return &LoxRSA{
		privKey:     nil,
		pubKey:      rsa.PublicKey{N: N, E: E},
		bitSize:     N.BitLen(),
		precomputed: false,
		methods:     make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxRSAPubKeyBytes(bytes []byte) (*LoxRSA, error) {
	pubKey, err := x509.ParsePKCS1PublicKey(bytes)
	if err != nil {
		return nil, err
	}
	return &LoxRSA{
		privKey:     nil,
		pubKey:      *pubKey,
		bitSize:     pubKey.N.BitLen(),
		precomputed: false,
		methods:     make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func NewLoxRSAPubKeyStr(str string) (*LoxRSA, error) {
	decodedBytes, err := LoxRSADecode(str)
	if err != nil {
		return nil, err
	}
	return NewLoxRSAPubKeyBytes(decodedBytes)
}

func (l *LoxRSA) encodePrivKeyPKCS1() []byte {
	return x509.MarshalPKCS1PrivateKey(l.privKey)
}

func (l *LoxRSA) encodePubKeyPKCS1() []byte {
	return x509.MarshalPKCS1PublicKey(&l.pubKey)
}

func (l *LoxRSA) encodePrivKeyPKCS8() ([]byte, error) {
	return x509.MarshalPKCS8PrivateKey(l.privKey)
}

func (l *LoxRSA) encodePubKeyPKIX() ([]byte, error) {
	return x509.MarshalPKIXPublicKey(&l.pubKey)
}

func (l *LoxRSA) isKeyPair() bool {
	return l.privKey != nil
}

func (l *LoxRSA) precompute() error {
	if !l.isKeyPair() {
		return loxerror.Error("Can only call 'rsa.precompute' on RSA keypairs.")
	}
	if !l.precomputed {
		l.privKey.Precompute()
		l.precomputed = true
	}
	return nil
}

func (l *LoxRSA) precomputeForce() error {
	if !l.isKeyPair() {
		return loxerror.Error("Can only call 'rsa.precomputeForce' on RSA keypairs.")
	}
	l.privKey.Precompute()
	l.precomputed = true
	return nil
}

func (l *LoxRSA) toPubKey() {
	if l.privKey != nil {
		l.privKey = nil
		l.precomputed = false
	}
}

func (l *LoxRSA) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if method, ok := l.methods[lexemeName]; ok {
		return method, nil
	}
	rsaFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native rsa fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.methods[lexemeName]; !ok {
			l.methods[lexemeName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'rsa.%v' must be a %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'rsa.%v' must be an %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	accessMustBeKeypair := func() (any, error) {
		return nil, loxerror.RuntimeError(
			name,
			fmt.Sprintf(
				"Can only access 'rsa.%v' on RSA keypairs.",
				lexemeName,
			),
		)
	}
	accessMustBePrecomputedKeypair := func() (any, error) {
		return nil, loxerror.RuntimeError(
			name,
			fmt.Sprintf(
				"Can only access 'rsa.%v' on precomputed RSA keypairs.",
				lexemeName,
			),
		)
	}
	callMustBeKeypair := func() (any, error) {
		return nil, loxerror.RuntimeError(
			name,
			fmt.Sprintf(
				"Can only call 'rsa.%v' on RSA keypairs.",
				lexemeName,
			),
		)
	}
	getArgList := func(callback *LoxFunction, numArgs int) list.List[any] {
		argList := list.NewListLen[any](int64(numArgs))
		callbackArity := callback.arity()
		if callbackArity > numArgs {
			for i := 0; i < callbackArity-numArgs; i++ {
				argList.Add(nil)
			}
		}
		return argList
	}
	switch lexemeName {
	case "bitLen", "bitSize":
		return int64(l.bitSize), nil
	case "decryptOAEP":
		return rsaFunc(-1, func(i *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen != 2 && argsLen != 3 {
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
			}

			var callable LoxCallable
			var ciphertext []byte
			var label []byte

			argZeroErrMsg := "First argument to 'rsa.decryptOAEP' must be a function."
			argOneErrMsg := "Second argument to 'rsa.decryptOAEP' must be a buffer or string."
			switch arg := args[0].(type) {
			case *LoxClass:
				return nil, loxerror.RuntimeError(name, argZeroErrMsg)
			case LoxCallable:
				callable = arg
			default:
				return nil, loxerror.RuntimeError(name, argZeroErrMsg)
			}
			switch args[1].(type) {
			case *LoxBuffer:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name, argOneErrMsg)
			}
			if argsLen == 3 {
				argTwoErrMsg := "Third argument to 'rsa.decryptOAEP' must be a buffer or string."
				switch args[2].(type) {
				case *LoxBuffer:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name, argTwoErrMsg)
				}
			}

			if !l.isKeyPair() {
				return accessMustBeKeypair()
			}

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
				switch arg := args[1].(type) {
				case *LoxBuffer:
					ciphertext = make([]byte, 0, len(arg.elements))
					for _, element := range arg.elements {
						ciphertext = append(ciphertext, byte(element.(int64)))
					}
				case *LoxString:
					var err error
					ciphertext, err = LoxRSADecode(arg.str)
					if err != nil {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				}
				if argsLen == 3 {
					switch arg := args[2].(type) {
					case *LoxBuffer:
						label = make([]byte, 0, len(arg.elements))
						for _, element := range arg.elements {
							label = append(label, byte(element.(int64)))
						}
					case *LoxString:
						label = []byte(arg.str)
					}
				} else {
					label = []byte{}
				}

				plaintext, err := rsa.DecryptOAEP(
					result.hash,
					nil,
					l.privKey,
					ciphertext,
					label,
				)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				buffer := EmptyLoxBufferCap(int64(len(plaintext)))
				for _, b := range plaintext {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				}
				return buffer, nil
			default:
				return nil, loxerror.RuntimeError(name,
					"Function argument to 'rsa.decryptOAEP' must return a hash object.")
			}
		})
	case "decryptOAEPToStr":
		return rsaFunc(-1, func(i *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen != 2 && argsLen != 3 {
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
			}

			var callable LoxCallable
			var ciphertext []byte
			var label []byte

			argZeroErrMsg := "First argument to 'rsa.decryptOAEPToStr' must be a function."
			argOneErrMsg := "Second argument to 'rsa.decryptOAEPToStr' must be a buffer or string."
			switch arg := args[0].(type) {
			case *LoxClass:
				return nil, loxerror.RuntimeError(name, argZeroErrMsg)
			case LoxCallable:
				callable = arg
			default:
				return nil, loxerror.RuntimeError(name, argZeroErrMsg)
			}
			switch args[1].(type) {
			case *LoxBuffer:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name, argOneErrMsg)
			}
			if argsLen == 3 {
				argTwoErrMsg := "Third argument to 'rsa.decryptOAEPToStr' must be a buffer or string."
				switch args[2].(type) {
				case *LoxBuffer:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name, argTwoErrMsg)
				}
			}

			if !l.isKeyPair() {
				return accessMustBeKeypair()
			}

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
				switch arg := args[1].(type) {
				case *LoxBuffer:
					ciphertext = make([]byte, 0, len(arg.elements))
					for _, element := range arg.elements {
						ciphertext = append(ciphertext, byte(element.(int64)))
					}
				case *LoxString:
					var err error
					ciphertext, err = LoxRSADecode(arg.str)
					if err != nil {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				}
				if argsLen == 3 {
					switch arg := args[2].(type) {
					case *LoxBuffer:
						label = make([]byte, 0, len(arg.elements))
						for _, element := range arg.elements {
							label = append(label, byte(element.(int64)))
						}
					case *LoxString:
						label = []byte(arg.str)
					}
				} else {
					label = []byte{}
				}

				plaintext, err := rsa.DecryptOAEP(
					result.hash,
					nil,
					l.privKey,
					ciphertext,
					label,
				)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return NewLoxStringQuote(string(plaintext)), nil
			default:
				return nil, loxerror.RuntimeError(name,
					"Function argument to 'rsa.decryptOAEPToStr' must return a hash object.")
			}
		})
	case "decryptPKCS1v15", "decrypt":
		return rsaFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var ciphertext []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				if l.isKeyPair() {
					ciphertext = make([]byte, 0, len(arg.elements))
					for _, element := range arg.elements {
						ciphertext = append(ciphertext, byte(element.(int64)))
					}
				}
			case *LoxString:
				if l.isKeyPair() {
					var err error
					ciphertext, err = LoxRSADecode(arg.str)
					if err != nil {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				}
			default:
				return argMustBeType("buffer or string")
			}

			if !l.isKeyPair() {
				return accessMustBeKeypair()
			}
			plaintext, err := rsa.DecryptPKCS1v15(nil, l.privKey, ciphertext)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			buffer := EmptyLoxBufferCap(int64(len(plaintext)))
			for _, b := range plaintext {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "decryptPKCS1v15ToStr", "decryptToStr":
		return rsaFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var ciphertext []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				if l.isKeyPair() {
					ciphertext = make([]byte, 0, len(arg.elements))
					for _, element := range arg.elements {
						ciphertext = append(ciphertext, byte(element.(int64)))
					}
				}
			case *LoxString:
				if l.isKeyPair() {
					var err error
					ciphertext, err = LoxRSADecode(arg.str)
					if err != nil {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				}
			default:
				return argMustBeType("buffer or string")
			}

			if !l.isKeyPair() {
				return accessMustBeKeypair()
			}
			plaintext, err := rsa.DecryptPKCS1v15(nil, l.privKey, ciphertext)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxStringQuote(string(plaintext)), nil
		})
	case "dp":
		if !l.isKeyPair() {
			return accessMustBeKeypair()
		}
		if !l.precomputed {
			return accessMustBePrecomputedKeypair()
		}
		return new(big.Int).Set(l.privKey.Precomputed.Dp), nil
	case "dq":
		if !l.isKeyPair() {
			return accessMustBeKeypair()
		}
		if !l.precomputed {
			return accessMustBePrecomputedKeypair()
		}
		return new(big.Int).Set(l.privKey.Precomputed.Dq), nil
	case "encryptOAEP":
		return rsaFunc(-1, func(i *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen != 2 && argsLen != 3 {
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
			}

			var callable LoxCallable
			var plaintext []byte
			var label []byte

			argZeroErrMsg := "First argument to 'rsa.encryptOAEP' must be a function."
			argOneErrMsg := "Second argument to 'rsa.encryptOAEP' must be a buffer or string."
			switch arg := args[0].(type) {
			case *LoxClass:
				return nil, loxerror.RuntimeError(name, argZeroErrMsg)
			case LoxCallable:
				callable = arg
			default:
				return nil, loxerror.RuntimeError(name, argZeroErrMsg)
			}
			switch args[1].(type) {
			case *LoxBuffer:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name, argOneErrMsg)
			}
			if argsLen == 3 {
				argTwoErrMsg := "Third argument to 'rsa.encryptOAEP' must be a buffer or string."
				switch args[2].(type) {
				case *LoxBuffer:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name, argTwoErrMsg)
				}
			}

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
				switch arg := args[1].(type) {
				case *LoxBuffer:
					plaintext = make([]byte, 0, len(arg.elements))
					for _, element := range arg.elements {
						plaintext = append(plaintext, byte(element.(int64)))
					}
				case *LoxString:
					plaintext = []byte(arg.str)
				}
				if argsLen == 3 {
					switch arg := args[2].(type) {
					case *LoxBuffer:
						label = make([]byte, 0, len(arg.elements))
						for _, element := range arg.elements {
							label = append(label, byte(element.(int64)))
						}
					case *LoxString:
						label = []byte(arg.str)
					}
				} else {
					label = []byte{}
				}

				ciphertext, err := rsa.EncryptOAEP(
					result.hash,
					crand.Reader,
					&l.pubKey,
					plaintext,
					label,
				)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				buffer := EmptyLoxBufferCap(int64(len(ciphertext)))
				for _, b := range ciphertext {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				}
				return buffer, nil
			default:
				return nil, loxerror.RuntimeError(name,
					"Function argument to 'rsa.encryptOAEP' must return a hash object.")
			}
		})
	case "encryptOAEPToStr":
		return rsaFunc(-1, func(i *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen != 2 && argsLen != 3 {
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
			}

			var callable LoxCallable
			var plaintext []byte
			var label []byte

			argZeroErrMsg := "First argument to 'rsa.encryptOAEPToStr' must be a function."
			argOneErrMsg := "Second argument to 'rsa.encryptOAEPToStr' must be a buffer or string."
			switch arg := args[0].(type) {
			case *LoxClass:
				return nil, loxerror.RuntimeError(name, argZeroErrMsg)
			case LoxCallable:
				callable = arg
			default:
				return nil, loxerror.RuntimeError(name, argZeroErrMsg)
			}
			switch args[1].(type) {
			case *LoxBuffer:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name, argOneErrMsg)
			}
			if argsLen == 3 {
				argTwoErrMsg := "Third argument to 'rsa.encryptOAEPToStr' must be a buffer or string."
				switch args[2].(type) {
				case *LoxBuffer:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name, argTwoErrMsg)
				}
			}

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
				switch arg := args[1].(type) {
				case *LoxBuffer:
					plaintext = make([]byte, 0, len(arg.elements))
					for _, element := range arg.elements {
						plaintext = append(plaintext, byte(element.(int64)))
					}
				case *LoxString:
					plaintext = []byte(arg.str)
				}
				if argsLen == 3 {
					switch arg := args[2].(type) {
					case *LoxBuffer:
						label = make([]byte, 0, len(arg.elements))
						for _, element := range arg.elements {
							label = append(label, byte(element.(int64)))
						}
					case *LoxString:
						label = []byte(arg.str)
					}
				} else {
					label = []byte{}
				}

				ciphertext, err := rsa.EncryptOAEP(
					result.hash,
					crand.Reader,
					&l.pubKey,
					plaintext,
					label,
				)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return NewLoxStringQuote(LoxRSAEncode(ciphertext)), nil
			default:
				return nil, loxerror.RuntimeError(name,
					"Function argument to 'rsa.encryptOAEPToStr' must return a hash object.")
			}
		})
	case "encryptPKCS1v15", "encrypt":
		return rsaFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var plaintext []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				plaintext = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					plaintext = append(plaintext, byte(element.(int64)))
				}
			case *LoxString:
				plaintext = []byte(arg.str)
			default:
				return argMustBeType("buffer or string")
			}

			ciphertext, err := rsa.EncryptPKCS1v15(crand.Reader, &l.pubKey, plaintext)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			buffer := EmptyLoxBufferCap(int64(len(ciphertext)))
			for _, b := range ciphertext {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "encryptPKCS1v15ToStr", "encryptToStr":
		return rsaFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var plaintext []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				plaintext = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					plaintext = append(plaintext, byte(element.(int64)))
				}
			case *LoxString:
				plaintext = []byte(arg.str)
			default:
				return argMustBeType("buffer or string")
			}

			ciphertext, err := rsa.EncryptPKCS1v15(crand.Reader, &l.pubKey, plaintext)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxStringQuote(LoxRSAEncode(ciphertext)), nil
		})
	case "exponent", "e":
		return int64(l.pubKey.E), nil
	case "isKeyPair":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isKeyPair(), nil
		})
	case "modSize", "n":
		return new(big.Int).Set(l.pubKey.N), nil
	case "modSizeBytes":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.pubKey.Size()), nil
		})
	case "numPrimes":
		if !l.isKeyPair() {
			return accessMustBeKeypair()
		}
		return int64(len(l.privKey.Primes)), nil
	case "precompute":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.precompute()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "precomputeForce":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.precomputeForce()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "primes":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return callMustBeKeypair()
			}
			primes := list.NewListCapDouble[any](int64(len(l.privKey.Primes)))
			for _, prime := range l.privKey.Primes {
				primes.Add(new(big.Int).Set(prime))
			}
			return NewLoxList(primes), nil
		})
	case "privateExponent", "d":
		if !l.isKeyPair() {
			return accessMustBeKeypair()
		}
		return new(big.Int).Set(l.privKey.D), nil
	case "privKeyEquals":
		return rsaFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if keyPair, ok := args[0].(*LoxRSA); ok {
				if !l.isKeyPair() {
					return callMustBeKeypair()
				}
				if !keyPair.isKeyPair() {
					return argMustBeTypeAn("rsa keypair with a private key")
				}
				return l.privKey.Equal(keyPair.privKey), nil
			}
			return argMustBeTypeAn("rsa keypair")
		})
	case "privKeyPKCS1", "privKey":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return callMustBeKeypair()
			}
			privKey := l.encodePrivKeyPKCS1()
			buffer := EmptyLoxBufferCap(int64(len(privKey)))
			for _, b := range privKey {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "privKeyPKCS8":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return callMustBeKeypair()
			}
			privKey, err := l.encodePrivKeyPKCS8()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			buffer := EmptyLoxBufferCap(int64(len(privKey)))
			for _, b := range privKey {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "privKeyStrPKCS1", "privKeyStr":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return callMustBeKeypair()
			}
			privKey := l.encodePrivKeyPKCS1()
			return NewLoxStringQuote(LoxRSAEncode(privKey)), nil
		})
	case "privKeyStrPKCS8":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return callMustBeKeypair()
			}
			privKey, err := l.encodePrivKeyPKCS8()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxStringQuote(LoxRSAEncode(privKey)), nil
		})
	case "pubKeyEquals":
		return rsaFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if keyPair, ok := args[0].(*LoxRSA); ok {
				return l.pubKey.Equal(&keyPair.pubKey), nil
			}
			return argMustBeTypeAn("rsa keypair or public key")
		})
	case "pubKeyPKCS1", "pubKey":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pubKey := l.encodePubKeyPKCS1()
			buffer := EmptyLoxBufferCap(int64(len(pubKey)))
			for _, b := range pubKey {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "pubKeyPKIX":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pubKey, err := l.encodePubKeyPKIX()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			buffer := EmptyLoxBufferCap(int64(len(pubKey)))
			for _, b := range pubKey {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "pubKeyStr":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pubKey := l.encodePubKeyPKCS1()
			return NewLoxStringQuote(LoxRSAEncode(pubKey)), nil
		})
	case "pubKeyStrPKIX":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pubKey, err := l.encodePubKeyPKIX()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxStringQuote(LoxRSAEncode(pubKey)), nil
		})
	case "ssh":
		return rsaFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var path string
			argsLen := len(args)
			switch argsLen {
			case 0:
				var pathErr error
				path, pathErr = os.UserHomeDir()
				if pathErr != nil {
					return nil, loxerror.RuntimeError(name, pathErr.Error())
				}
			case 1:
				if loxStr, ok := args[0].(*LoxString); ok {
					path = loxStr.str
				} else {
					return argMustBeType("string")
				}
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}

			stat, statErr := os.Stat(path)
			if statErr != nil {
				return nil, loxerror.RuntimeError(name, statErr.Error())
			}
			if !stat.IsDir() {
				return nil, loxerror.RuntimeError(name,
					"Path argument to 'rsa.ssh' must refer to a directory.")
			}

			if l.isKeyPair() {
				pemKey, pemKeyErr := ssh.MarshalPrivateKey(l.privKey, "")
				if pemKeyErr != nil {
					return nil, loxerror.RuntimeError(name, pemKeyErr.Error())
				}
				sshPrivKey := pem.EncodeToMemory(pemKey)
				writeErr := os.WriteFile(filepath.Join(path, "id_rsa"), sshPrivKey, 0600)
				if writeErr != nil {
					return nil, loxerror.RuntimeError(name, writeErr.Error())
				}
			}
			sshPubKey, sshPubKeyErr := ssh.NewPublicKey(&l.pubKey)
			if sshPubKeyErr != nil {
				return nil, loxerror.RuntimeError(name, sshPubKeyErr.Error())
			}
			authorizedKey := ssh.MarshalAuthorizedKey(sshPubKey)
			writeErr := os.WriteFile(filepath.Join(path, "id_rsa.pub"), authorizedKey, 0644)
			if writeErr != nil {
				return nil, loxerror.RuntimeError(name, writeErr.Error())
			}

			return nil, nil
		})
	case "sshComment":
		return rsaFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var comment, path string
			argsLen := len(args)
			switch argsLen {
			case 1, 2:
				if _, ok := args[0].(*LoxString); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'rsa.sshComment' must be a string.")
				}
				comment = args[0].(*LoxString).str
				if argsLen == 2 {
					if _, ok := args[1].(*LoxString); !ok {
						return nil, loxerror.RuntimeError(name,
							"Second argument to 'rsa.sshComment' must be a string.")
					}
					path = args[1].(*LoxString).str
				} else {
					var pathErr error
					path, pathErr = os.UserHomeDir()
					if pathErr != nil {
						return nil, loxerror.RuntimeError(name, pathErr.Error())
					}
				}
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}

			stat, statErr := os.Stat(path)
			if statErr != nil {
				return nil, loxerror.RuntimeError(name, statErr.Error())
			}
			if !stat.IsDir() {
				return nil, loxerror.RuntimeError(name,
					"Path argument to 'rsa.sshComment' must refer to a directory.")
			}

			if l.isKeyPair() {
				pemKey, pemKeyErr := ssh.MarshalPrivateKey(l.privKey, comment)
				if pemKeyErr != nil {
					return nil, loxerror.RuntimeError(name, pemKeyErr.Error())
				}
				sshPrivKey := pem.EncodeToMemory(pemKey)
				writeErr := os.WriteFile(filepath.Join(path, "id_rsa"), sshPrivKey, 0600)
				if writeErr != nil {
					return nil, loxerror.RuntimeError(name, writeErr.Error())
				}
			}
			sshPubKey, sshPubKeyErr := ssh.NewPublicKey(&l.pubKey)
			if sshPubKeyErr != nil {
				return nil, loxerror.RuntimeError(name, sshPubKeyErr.Error())
			}
			authorizedKey := ssh.MarshalAuthorizedKey(sshPubKey)
			writeErr := os.WriteFile(filepath.Join(path, "id_rsa.pub"), authorizedKey, 0644)
			if writeErr != nil {
				return nil, loxerror.RuntimeError(name, writeErr.Error())
			}

			return nil, nil
		})
	case "toPubKey":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.toPubKey()
			return l, nil
		})
	case "validate":
		return rsaFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return callMustBeKeypair()
			}
			err := l.privKey.Validate()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	}
	var errorMsg string
	if l.isKeyPair() {
		errorMsg = "RSA keypairs have no property called '" + lexemeName + "'."
	} else {
		errorMsg = "RSA public keys have no property called '" + lexemeName + "'."
	}
	return nil, loxerror.RuntimeError(name, errorMsg)
}

func (l *LoxRSA) String() string {
	if !l.isKeyPair() {
		return fmt.Sprintf("<RSA public key at %p>", l)
	}
	return fmt.Sprintf("<RSA keypair at %p>", l)
}

func (l *LoxRSA) Type() string {
	return "rsa"
}
