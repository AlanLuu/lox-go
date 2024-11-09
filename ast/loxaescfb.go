package ast

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func LoxAESDecode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

func LoxAESEncode(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

type LoxAESCFB struct {
	block   cipher.Block
	key     []byte
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxAESCFB(keyLenBytes int) (*LoxAESCFB, error) {
	if keyLenBytes != 128 && keyLenBytes != 192 && keyLenBytes != 256 {
		return nil, loxerror.Error("AES integer argument must be 128, 192, or 256.")
	}
	key := make([]byte, keyLenBytes/8)
	if _, err := io.ReadFull(crand.Reader, key); err != nil {
		return nil, loxerror.Error("Failed to generate random AES key.")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &LoxAESCFB{
		block:   block,
		key:     key,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func NewLoxAESCFBBytes(key []byte) (*LoxAESCFB, error) {
	keyLenBytes := len(key)
	if keyLenBytes != 16 && keyLenBytes != 24 && keyLenBytes != 32 {
		return nil, loxerror.Error("Key length in bytes must be 16, 24, or 32.")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &LoxAESCFB{
		block:   block,
		key:     key,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func (l *LoxAESCFB) aesTypeInt() int {
	return len(l.key) * 8
}

func (l *LoxAESCFB) aesTypeStr() string {
	return fmt.Sprintf("AES-%v", l.aesTypeInt())
}

func (l *LoxAESCFB) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	aescfbFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native aes-cfb fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'aes-cfb.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "base64", "b64", "keyStr":
		return aescfbFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(LoxAESEncode(l.key)), nil
		})
	case "bytes", "key":
		return aescfbFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			buffer := EmptyLoxBufferCap(int64(len(l.key)))
			for _, b := range l.key {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "decrypt":
		return aescfbFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var ciphertext []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				ciphertext = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					ciphertext = append(ciphertext, byte(element.(int64)))
				}
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'aes-cfb.decrypt' must be in read mode.")
				}
				var readErr error
				ciphertext, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				var decodeErr error
				ciphertext, decodeErr = LoxAESDecode(arg.str)
				if decodeErr != nil {
					return nil, loxerror.RuntimeError(name, decodeErr.Error())
				}
			default:
				return argMustBeType("buffer, file, or string")
			}
			if len(ciphertext) < aes.BlockSize {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"AES-CFB: AES ciphertext size must be at least %v bytes.",
						aes.BlockSize,
					),
				)
			}

			iv := ciphertext[:aes.BlockSize]
			ciphertext = ciphertext[aes.BlockSize:]
			stream := cipher.NewCFBDecrypter(l.block, iv)
			stream.XORKeyStream(ciphertext, ciphertext)

			buffer := EmptyLoxBufferCap(int64(len(ciphertext)))
			for _, b := range ciphertext {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "decryptToStr":
		return aescfbFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var ciphertext []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				ciphertext = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					ciphertext = append(ciphertext, byte(element.(int64)))
				}
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'aes-cfb.decryptToStr' must be in read mode.")
				}
				var readErr error
				ciphertext, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				var decodeErr error
				ciphertext, decodeErr = LoxAESDecode(arg.str)
				if decodeErr != nil {
					return nil, loxerror.RuntimeError(name, decodeErr.Error())
				}
			default:
				return argMustBeType("buffer, file, or string")
			}
			if len(ciphertext) < aes.BlockSize {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"AES-CFB: AES ciphertext size must be at least %v bytes.",
						aes.BlockSize,
					),
				)
			}
			iv := ciphertext[:aes.BlockSize]
			ciphertext = ciphertext[aes.BlockSize:]
			stream := cipher.NewCFBDecrypter(l.block, iv)
			stream.XORKeyStream(ciphertext, ciphertext)
			return NewLoxStringQuote(string(ciphertext)), nil
		})
	case "encrypt":
		return aescfbFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var plaintext []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				plaintext = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					plaintext = append(plaintext, byte(element.(int64)))
				}
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'aes-cfb.encrypt' must be in read mode.")
				}
				var readErr error
				plaintext, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				plaintext = []byte(arg.str)
			default:
				return argMustBeType("buffer, file, or string")
			}
			ciphertext := make([]byte, len(plaintext)+aes.BlockSize)
			iv := ciphertext[:aes.BlockSize]
			if _, err := io.ReadFull(crand.Reader, iv); err != nil {
				return nil, loxerror.RuntimeError(name,
					"AES-CFB: Failed to generate random IV.")
			}

			stream := cipher.NewCFBEncrypter(l.block, iv)
			stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

			buffer := EmptyLoxBufferCap(int64(len(ciphertext)))
			for _, b := range ciphertext {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "encryptToStr":
		return aescfbFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var plaintext []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				plaintext = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					plaintext = append(plaintext, byte(element.(int64)))
				}
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'aes-cfb.encryptToStr' must be in read mode.")
				}
				var readErr error
				plaintext, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				plaintext = []byte(arg.str)
			default:
				return argMustBeType("buffer, file, or string")
			}
			ciphertext := make([]byte, len(plaintext)+aes.BlockSize)
			iv := ciphertext[:aes.BlockSize]
			if _, err := io.ReadFull(crand.Reader, iv); err != nil {
				return nil, loxerror.RuntimeError(name,
					"AES-CFB: Failed to generate random IV.")
			}
			stream := cipher.NewCFBEncrypter(l.block, iv)
			stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
			return NewLoxStringQuote(LoxAESEncode(ciphertext)), nil
		})
	case "hex":
		return aescfbFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxString(hex.EncodeToString(l.key), '\''), nil
		})
	case "typeInt":
		return aescfbFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.aesTypeInt()), nil
		})
	case "typeStr":
		return aescfbFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.aesTypeStr()), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "AES-CFB objects have no property called '"+methodName+"'.")
}

func (l *LoxAESCFB) String() string {
	return fmt.Sprintf("<AES-CFB object at %p>", l)
}

func (l *LoxAESCFB) Type() string {
	return "aes-cfb"
}
