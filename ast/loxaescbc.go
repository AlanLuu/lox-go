package ast

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxAESCBC struct {
	block   cipher.Block
	key     []byte
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxAESCBC(keyLenBytes int) (*LoxAESCBC, error) {
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
	return &LoxAESCBC{
		block:   block,
		key:     key,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func NewLoxAESCBCBytes(key []byte) (*LoxAESCBC, error) {
	keyLenBytes := len(key)
	if keyLenBytes != 16 && keyLenBytes != 24 && keyLenBytes != 32 {
		return nil, loxerror.Error("Key length in bytes must be 16, 24, or 32.")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &LoxAESCBC{
		block:   block,
		key:     key,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func (l *LoxAESCBC) aesTypeInt() int {
	return len(l.key) * 8
}

func (l *LoxAESCBC) aesTypeStr() string {
	return fmt.Sprintf("AES-%v", l.aesTypeInt())
}

func (l *LoxAESCBC) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	aescbcFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native aes-cbc fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'aes-cbc.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "base64", "b64", "keyStr":
		return aescbcFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(LoxAESEncode(l.key)), nil
		})
	case "bytes", "key":
		return aescbcFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
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
		return aescbcFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
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
						"File argument to 'aes-cbc.decrypt' must be in read mode.")
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
						"AES-CBC: AES ciphertext size must be at least %v bytes.",
						aes.BlockSize,
					),
				)
			}

			iv := ciphertext[:aes.BlockSize]
			ciphertext = ciphertext[aes.BlockSize:]
			if len(ciphertext)%aes.BlockSize != 0 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"AES-CBC: AES ciphertext size must be a multiple of %v.",
						aes.BlockSize,
					),
				)
			}

			mode := cipher.NewCBCDecrypter(l.block, iv)
			mode.CryptBlocks(ciphertext, ciphertext)
			textLen := len(ciphertext)
			if textLen > 0 && textLen%aes.BlockSize == 0 {
				paddingByte := ciphertext[textLen-1]
				ciphertext = ciphertext[:textLen-int(paddingByte)]
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
	case "decryptToFile":
		return aescbcFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"First argument to 'aes-cbc.decryptToFile' must be a buffer, file, or string.")
			}
			switch arg := args[1].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Second file argument to 'aes-cbc.decryptToFile' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'aes-cbc.decryptToFile' must be a file or string.")
			}

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
						"File argument to 'aes-cbc.decryptToFile' must be in read mode.")
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
						"AES-CBC: AES ciphertext size must be at least %v bytes.",
						aes.BlockSize,
					),
				)
			}

			iv := ciphertext[:aes.BlockSize]
			ciphertext = ciphertext[aes.BlockSize:]
			if len(ciphertext)%aes.BlockSize != 0 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"AES-CBC: AES ciphertext size must be a multiple of %v.",
						aes.BlockSize,
					),
				)
			}

			mode := cipher.NewCBCDecrypter(l.block, iv)
			mode.CryptBlocks(ciphertext, ciphertext)
			textLen := len(ciphertext)
			if textLen > 0 && textLen%aes.BlockSize == 0 {
				paddingByte := ciphertext[textLen-1]
				ciphertext = ciphertext[:textLen-int(paddingByte)]
			}

			switch arg := args[1].(type) {
			case *LoxFile:
				_, err := arg.file.Write(ciphertext)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			case *LoxString:
				err := os.WriteFile(arg.str, ciphertext, 0666)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}

			return nil, nil
		})
	case "decryptToStr":
		return aescbcFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
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
						"File argument to 'aes-cbc.decryptToStr' must be in read mode.")
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
						"AES-CBC: AES ciphertext size must be at least %v bytes.",
						aes.BlockSize,
					),
				)
			}

			iv := ciphertext[:aes.BlockSize]
			ciphertext = ciphertext[aes.BlockSize:]
			if len(ciphertext)%aes.BlockSize != 0 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"AES-CBC: AES ciphertext size must be a multiple of %v.",
						aes.BlockSize,
					),
				)
			}

			mode := cipher.NewCBCDecrypter(l.block, iv)
			mode.CryptBlocks(ciphertext, ciphertext)
			textLen := len(ciphertext)
			if textLen > 0 && textLen%aes.BlockSize == 0 {
				paddingByte := ciphertext[textLen-1]
				ciphertext = ciphertext[:textLen-int(paddingByte)]
			}
			return NewLoxStringQuote(string(ciphertext)), nil
		})
	case "encrypt":
		return aescbcFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
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
						"File argument to 'aes-cbc.encrypt' must be in read mode.")
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

			//Should never happen
			if aes.BlockSize <= 0 || aes.BlockSize > 255 {
				return nil, loxerror.RuntimeError(name,
					"AES-CBC: Internal block size error when encrypting.")
			}

			//Use PKCS#7 padding if plaintext length is not a multiple of AES block size
			//https://stackoverflow.com/a/13572751
			if len(plaintext)%aes.BlockSize != 0 {
				n := byte(aes.BlockSize - (len(plaintext) % aes.BlockSize))
				for len(plaintext)%aes.BlockSize != 0 {
					plaintext = append(plaintext, n)
				}
			} else {
				for i := 0; i < aes.BlockSize; i++ {
					plaintext = append(plaintext, byte(aes.BlockSize))
				}
			}

			ciphertext := make([]byte, len(plaintext)+aes.BlockSize)
			iv := ciphertext[:aes.BlockSize]
			if _, err := io.ReadFull(crand.Reader, iv); err != nil {
				return nil, loxerror.RuntimeError(name,
					"AES-CBC: Failed to generate random IV.")
			}

			mode := cipher.NewCBCEncrypter(l.block, iv)
			mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

			buffer := EmptyLoxBufferCap(int64(len(ciphertext)))
			for _, b := range ciphertext {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "encryptToFile":
		return aescbcFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"First argument to 'aes-cbc.encryptToFile' must be a buffer, file, or string.")
			}
			switch arg := args[1].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Second file argument to 'aes-cbc.encryptToFile' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'aes-cbc.encryptToFile' must be a file or string.")
			}

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
						"File argument to 'aes-cbc.encryptToStr' must be in read mode.")
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

			//Should never happen
			if aes.BlockSize <= 0 || aes.BlockSize > 255 {
				return nil, loxerror.RuntimeError(name,
					"AES-CBC: Internal block size error when encrypting.")
			}

			//Use PKCS#7 padding if plaintext length is not a multiple of AES block size
			//https://stackoverflow.com/a/13572751
			if len(plaintext)%aes.BlockSize != 0 {
				n := byte(aes.BlockSize - (len(plaintext) % aes.BlockSize))
				for len(plaintext)%aes.BlockSize != 0 {
					plaintext = append(plaintext, n)
				}
			} else {
				for i := 0; i < aes.BlockSize; i++ {
					plaintext = append(plaintext, byte(aes.BlockSize))
				}
			}

			ciphertext := make([]byte, len(plaintext)+aes.BlockSize)
			iv := ciphertext[:aes.BlockSize]
			if _, err := io.ReadFull(crand.Reader, iv); err != nil {
				return nil, loxerror.RuntimeError(name,
					"AES-CBC: Failed to generate random IV.")
			}

			mode := cipher.NewCBCEncrypter(l.block, iv)
			mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

			switch arg := args[1].(type) {
			case *LoxFile:
				_, err := arg.file.Write(ciphertext)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			case *LoxString:
				err := os.WriteFile(arg.str, ciphertext, 0666)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}

			return nil, nil
		})
	case "encryptToStr":
		return aescbcFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
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
						"File argument to 'aes-cbc.encryptToStr' must be in read mode.")
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

			//Should never happen
			if aes.BlockSize <= 0 || aes.BlockSize > 255 {
				return nil, loxerror.RuntimeError(name,
					"AES-CBC: Internal block size error when encrypting.")
			}

			//Use PKCS#7 padding if plaintext length is not a multiple of AES block size
			//https://stackoverflow.com/a/13572751
			if len(plaintext)%aes.BlockSize != 0 {
				n := byte(aes.BlockSize - (len(plaintext) % aes.BlockSize))
				for len(plaintext)%aes.BlockSize != 0 {
					plaintext = append(plaintext, n)
				}
			} else {
				for i := 0; i < aes.BlockSize; i++ {
					plaintext = append(plaintext, byte(aes.BlockSize))
				}
			}

			ciphertext := make([]byte, len(plaintext)+aes.BlockSize)
			iv := ciphertext[:aes.BlockSize]
			if _, err := io.ReadFull(crand.Reader, iv); err != nil {
				return nil, loxerror.RuntimeError(name,
					"AES-CBC: Failed to generate random IV.")
			}

			mode := cipher.NewCBCEncrypter(l.block, iv)
			mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
			return NewLoxStringQuote(LoxAESEncode(ciphertext)), nil
		})
	case "hex":
		return aescbcFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxString(hex.EncodeToString(l.key), '\''), nil
		})
	case "typeInt":
		return aescbcFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.aesTypeInt()), nil
		})
	case "typeStr":
		return aescbcFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.aesTypeStr()), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "AES-CBC objects have no property called '"+methodName+"'.")
}

func (l *LoxAESCBC) String() string {
	return fmt.Sprintf("<AES-CBC object at %p>", l)
}

func (l *LoxAESCBC) Type() string {
	return "aes-cbc"
}
