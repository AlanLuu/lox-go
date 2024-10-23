package ast

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/fernet/fernet-go"
)

type LoxFernet struct {
	key     *fernet.Key
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxFernet() (*LoxFernet, error) {
	key := fernet.Key([32]byte{})
	err := key.Generate()
	if err != nil {
		return nil, err
	}
	return &LoxFernet{
		key:     &key,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func NewLoxFernetFromBytes(bytes [32]byte) *LoxFernet {
	key := fernet.Key(bytes)
	return &LoxFernet{
		key:     &key,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxFernetFromString(keyStr string) (*LoxFernet, error) {
	key, err := fernet.DecodeKey(keyStr)
	if err != nil {
		return nil, err
	}
	return &LoxFernet{
		key:     key,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func (l *LoxFernet) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	fernetFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native fernet fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'fernet.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	decryptFailMsg := func(element any) string {
		switch element.(type) {
		case *LoxBuffer:
			return "Fernet: failed to decrypt buffer."
		case *LoxFile:
			return "Fernet: failed to decrypt file."
		case *LoxString:
			return "Fernet: failed to decrypt string."
		default:
			return ""
		}
	}
	switch methodName {
	case "base64", "b64":
		return fernetFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.key.Encode()), nil
		})
	case "bytes":
		return fernetFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			buffer := EmptyLoxBufferCap(32)
			for _, b := range l.key {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "decrypt":
		return fernetFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var bytes []byte
			var decryptedBytes []byte

			switch arg := args[0].(type) {
			case *LoxBuffer:
				bytes = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'fernet.decrypt' must be in read mode.")
				}
				var readErr error
				bytes, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				bytes = []byte(arg.str)
			default:
				return argMustBeType("buffer, file, or string")
			}

			decryptedBytes = fernet.VerifyAndDecrypt(bytes, 0, []*fernet.Key{l.key})
			if decryptedBytes == nil {
				return nil, loxerror.RuntimeError(name, decryptFailMsg(args[0]))
			}
			buffer := EmptyLoxBufferCap(int64(len(decryptedBytes)))
			for _, b := range decryptedBytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "decryptToFile":
		return fernetFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"First argument to 'fernet.decryptToFile' must be a buffer, file, or string.")
			}
			switch arg := args[1].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Second file argument to 'fernet.decryptToFile' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'fernet.decryptToFile' must be a file or string.")
			}

			var bytes []byte
			var decryptedBytes []byte
			switch arg := args[0].(type) {
			case *LoxBuffer:
				bytes = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"First file argument to 'fernet.decryptToFile' must be in read mode.")
				}
				var readErr error
				bytes, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				bytes = []byte(arg.str)
			}

			decryptedBytes = fernet.VerifyAndDecrypt(bytes, 0, []*fernet.Key{l.key})
			if decryptedBytes == nil {
				return nil, loxerror.RuntimeError(name, decryptFailMsg(args[0]))
			}
			switch arg := args[1].(type) {
			case *LoxFile:
				_, err := arg.file.Write(decryptedBytes)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			case *LoxString:
				err := os.WriteFile(arg.str, decryptedBytes, 0666)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}

			return nil, nil
		})
	case "decryptToStr":
		return fernetFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var bytes []byte
			var decryptedBytes []byte

			switch arg := args[0].(type) {
			case *LoxBuffer:
				bytes = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'fernet.decryptToStr' must be in read mode.")
				}
				var readErr error
				bytes, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				bytes = []byte(arg.str)
			default:
				return argMustBeType("buffer, file, or string")
			}

			decryptedBytes = fernet.VerifyAndDecrypt(bytes, 0, []*fernet.Key{l.key})
			if decryptedBytes == nil {
				return nil, loxerror.RuntimeError(name, decryptFailMsg(args[0]))
			}
			return NewLoxStringQuote(string(decryptedBytes)), nil
		})
	case "encrypt":
		return fernetFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var encryptedBytes []byte
			var err error

			switch arg := args[0].(type) {
			case *LoxBuffer:
				bytes := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				encryptedBytes, err = fernet.EncryptAndSign(bytes, l.key)
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'fernet.encrypt' must be in read mode.")
				}
				bytes, readErr := io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
				encryptedBytes, err = fernet.EncryptAndSign(bytes, l.key)
			case *LoxString:
				encryptedBytes, err = fernet.EncryptAndSign([]byte(arg.str), l.key)
			default:
				return argMustBeType("buffer, file, or string")
			}

			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			buffer := EmptyLoxBufferCap(int64(len(encryptedBytes)))
			for _, b := range encryptedBytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "encryptToFile":
		return fernetFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"First argument to 'fernet.encryptToFile' must be a buffer, file, or string.")
			}
			switch arg := args[1].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Second file argument to 'fernet.encryptToFile' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'fernet.encryptToFile' must be a file or string.")
			}

			var encryptedBytes []byte
			var err error
			switch arg := args[0].(type) {
			case *LoxBuffer:
				bytes := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				encryptedBytes, err = fernet.EncryptAndSign(bytes, l.key)
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"First file argument to 'fernet.encryptToFile' must be in read mode.")
				}
				bytes, readErr := io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
				encryptedBytes, err = fernet.EncryptAndSign(bytes, l.key)
			case *LoxString:
				encryptedBytes, err = fernet.EncryptAndSign([]byte(arg.str), l.key)
			}
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}

			switch arg := args[1].(type) {
			case *LoxFile:
				_, err := arg.file.Write(encryptedBytes)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			case *LoxString:
				err := os.WriteFile(arg.str, encryptedBytes, 0666)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}

			return nil, nil
		})
	case "encryptToStr":
		return fernetFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var encryptedBytes []byte
			var err error

			switch arg := args[0].(type) {
			case *LoxBuffer:
				bytes := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				encryptedBytes, err = fernet.EncryptAndSign(bytes, l.key)
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'fernet.encryptToStr' must be in read mode.")
				}
				bytes, readErr := io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
				encryptedBytes, err = fernet.EncryptAndSign(bytes, l.key)
			case *LoxString:
				encryptedBytes, err = fernet.EncryptAndSign([]byte(arg.str), l.key)
			default:
				return argMustBeType("buffer, file, or string")
			}

			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxString(string(encryptedBytes), '\''), nil
		})
	case "hex":
		return fernetFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxString(hex.EncodeToString(l.key[:]), '\''), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Fernet objects have no property called '"+methodName+"'.")
}

func (l *LoxFernet) String() string {
	return fmt.Sprintf("[fernet key: %v]", l.key.Encode())
}

func (l *LoxFernet) Type() string {
	return "fernet"
}
