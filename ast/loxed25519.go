package ast

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func LoxEd25519Decode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

func LoxEd25519Encode(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

type LoxEd25519 struct {
	pubKey  ed25519.PublicKey
	privKey ed25519.PrivateKey
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxEd25519() (*LoxEd25519, error) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}
	return &LoxEd25519{
		pubKey:  pubKey,
		privKey: privKey,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func NewLoxEd25519PrivKey(privKey ed25519.PrivateKey) (*LoxEd25519, error) {
	currentSize := len(privKey)
	privKeySize := ed25519.PrivateKeySize
	if currentSize != privKeySize {
		return nil, loxerror.Error(
			fmt.Sprintf(
				"Ed25519 private key size must be %v bytes and not %v bytes.",
				privKeySize,
				currentSize,
			),
		)
	}
	return &LoxEd25519{
		pubKey:  privKey.Public().(ed25519.PublicKey),
		privKey: privKey,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func NewLoxEd25519PrivKeySeed(seedBytes []byte) (*LoxEd25519, error) {
	currentSize := len(seedBytes)
	seedSize := ed25519.SeedSize
	if currentSize != seedSize {
		return nil, loxerror.Error(
			fmt.Sprintf(
				"Ed25519 private key seed size must be %v bytes and not %v bytes.",
				seedSize,
				currentSize,
			),
		)
	}
	return NewLoxEd25519PrivKey(ed25519.NewKeyFromSeed(seedBytes))
}

func NewLoxEd25519PubKey(pubKey ed25519.PublicKey) (*LoxEd25519, error) {
	currentSize := len(pubKey)
	pubKeySize := ed25519.PublicKeySize
	if currentSize != pubKeySize {
		return nil, loxerror.Error(
			fmt.Sprintf(
				"Ed25519 public key size must be %v bytes and not %v bytes.",
				pubKeySize,
				currentSize,
			),
		)
	}
	return &LoxEd25519{
		pubKey:  pubKey,
		privKey: nil,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func (l *LoxEd25519) isKeyPair() bool {
	return l.privKey != nil
}

func (l *LoxEd25519) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	ed25519Func := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native ed25519 fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'ed25519.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'ed25519.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "isKeyPair":
		return ed25519Func(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isKeyPair(), nil
		})
	case "pubKey":
		return ed25519Func(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			buffer := EmptyLoxBufferCap(int64(len(l.pubKey)))
			for _, b := range l.pubKey {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "pubKeyEquals":
		return ed25519Func(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if keyPair, ok := args[0].(*LoxEd25519); ok {
				return l.pubKey.Equal(keyPair.pubKey), nil
			}
			return argMustBeTypeAn("ed25519 keypair or public key")
		})
	case "pubKeyStr":
		return ed25519Func(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(LoxEd25519Encode(l.pubKey)), nil
		})
	case "privKey":
		return ed25519Func(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return nil, loxerror.RuntimeError(name,
					"Can only call 'ed25519.privKey' on ed25519 keypairs.")
			}
			buffer := EmptyLoxBufferCap(int64(len(l.privKey)))
			for _, b := range l.privKey {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "privKeyEquals":
		return ed25519Func(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if keyPair, ok := args[0].(*LoxEd25519); ok {
				if !l.isKeyPair() {
					return nil, loxerror.RuntimeError(name,
						"Can only call 'ed25519.privKeyEquals' on ed25519 keypairs.")
				}
				if !keyPair.isKeyPair() {
					return argMustBeTypeAn("ed25519 keypair with a private key")
				}
				return l.privKey.Equal(keyPair.privKey), nil
			}
			return argMustBeTypeAn("ed25519 keypair")
		})
	case "privKeyStr":
		return ed25519Func(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return nil, loxerror.RuntimeError(name,
					"Can only call 'ed25519.privKeyStr' on ed25519 keypairs.")
			}
			return NewLoxStringQuote(LoxEd25519Encode(l.privKey)), nil
		})
	case "seed":
		return ed25519Func(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return nil, loxerror.RuntimeError(name,
					"Can only call 'ed25519.seed' on ed25519 keypairs.")
			}
			seedBytes := l.privKey.Seed()
			buffer := EmptyLoxBufferCap(int64(len(seedBytes)))
			for _, b := range seedBytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "seedB64", "seedBase64":
		return ed25519Func(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Can only call 'ed25519.%v' on ed25519 keypairs.",
						methodName,
					),
				)
			}
			seedBytes := l.privKey.Seed()
			return NewLoxString(base64.StdEncoding.EncodeToString(seedBytes), '\''), nil
		})
	case "seedEncoded":
		return ed25519Func(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return nil, loxerror.RuntimeError(name,
					"Can only call 'ed25519.seedEncoded' on ed25519 keypairs.")
			}
			seedBytes := l.privKey.Seed()
			return NewLoxStringQuote(LoxEd25519Encode(seedBytes)), nil
		})
	case "sign":
		return ed25519Func(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var message []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				if l.isKeyPair() {
					message = make([]byte, 0, len(arg.elements))
					for _, element := range arg.elements {
						message = append(message, byte(element.(int64)))
					}
				}
			case *LoxString:
				if l.isKeyPair() {
					message = []byte(arg.str)
				}
			default:
				return argMustBeType("buffer or string")
			}

			if !l.isKeyPair() {
				return nil, loxerror.RuntimeError(name,
					"Can only call 'ed25519.sign' on ed25519 keypairs.")
			}
			signature := ed25519.Sign(l.privKey, message)
			buffer := EmptyLoxBufferCap(int64(len(signature)))
			for _, b := range signature {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "signToStr":
		return ed25519Func(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var message []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				if l.isKeyPair() {
					message = make([]byte, 0, len(arg.elements))
					for _, element := range arg.elements {
						message = append(message, byte(element.(int64)))
					}
				}
			case *LoxString:
				if l.isKeyPair() {
					message = []byte(arg.str)
				}
			default:
				return argMustBeType("buffer or string")
			}

			if !l.isKeyPair() {
				return nil, loxerror.RuntimeError(name,
					"Can only call 'ed25519.signToStr' on ed25519 keypairs.")
			}
			signature := ed25519.Sign(l.privKey, message)
			return NewLoxStringQuote(LoxEd25519Encode(signature)), nil
		})
	case "verify":
		return ed25519Func(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"First argument to 'ed25519.verify' must be a buffer or string.")
			}
			switch args[1].(type) {
			case *LoxBuffer:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'ed25519.verify' must be a buffer or string.")
			}

			var message []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				message = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					message = append(message, byte(element.(int64)))
				}
			case *LoxString:
				message = []byte(arg.str)
			}

			var signature []byte
			switch arg := args[1].(type) {
			case *LoxBuffer:
				signature = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					signature = append(signature, byte(element.(int64)))
				}
			case *LoxString:
				var err error
				signature, err = LoxEd25519Decode(arg.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}

			return ed25519.Verify(l.pubKey, message, signature), nil
		})
	}
	var errorMsg string
	if l.isKeyPair() {
		errorMsg = "Ed25519 keypairs have no property called '" + methodName + "'."
	} else {
		errorMsg = "Ed25519 public keys have no property called '" + methodName + "'."
	}
	return nil, loxerror.RuntimeError(name, errorMsg)
}

func (l *LoxEd25519) String() string {
	if !l.isKeyPair() {
		return fmt.Sprintf("<ed25519 public key at %p>", l)
	}
	return fmt.Sprintf("<ed25519 keypair at %p>", l)
}

func (l *LoxEd25519) Type() string {
	return "ed25519"
}
