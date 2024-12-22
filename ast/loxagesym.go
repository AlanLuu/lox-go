package ast

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func LoxAgeEncryptionDecode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

func LoxAgeEncryptionEncode(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

type LoxAgeSymmetric struct {
	password     string
	initPassword bool
	methods      map[string]*struct{ ProtoLoxCallable }
}

func NewLoxAgeSymmetric() *LoxAgeSymmetric {
	return &LoxAgeSymmetric{
		password:     "",
		initPassword: false,
		methods:      make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxAgeSymmetricPassword(password string) *LoxAgeSymmetric {
	if len(password) == 0 {
		fmt.Fprintln(
			os.Stderr,
			"WARNING: empty password specified when creating age symmetric encryption object.",
		)
	}
	return &LoxAgeSymmetric{
		password:     password,
		initPassword: true,
		methods:      make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxAgeSymmetric) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	ageSymFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native age symmetric encryption fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'age symmetric.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	passwordErr := func(isArg bool) (any, error) {
		var first string
		if isArg || !l.initPassword {
			first = "Password argument "
		} else {
			first = "Password "
		}
		second := fmt.Sprintf("to 'age symmetric.%v' cannot be empty.", methodName)
		return nil, loxerror.RuntimeError(name, first+second)
	}
	switch methodName {
	case "decrypt":
		return ageSymFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 1, 2:
				switch args[0].(type) {
				case *LoxBuffer:
				case *LoxFile:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name,
						"First argument to 'age symmetric.decrypt' must be a buffer, file, or string.")
				}

				var password string = l.password
				if argsLen == 2 {
					switch arg := args[1].(type) {
					case *LoxString:
						password = arg.str
					default:
						return nil, loxerror.RuntimeError(name,
							"Second argument to 'age symmetric.decrypt' must be a string.")
					}
				} else if !l.initPassword {
					return nil, loxerror.RuntimeError(name,
						"Must specify password argument to 'age symmetric.decrypt'.")
				}
				if len(password) == 0 {
					return passwordErr(argsLen == 2)
				}

				identity, err := age.NewScryptIdentity(password)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
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
							"File argument to 'age symmetric.decrypt' must be in read mode.")
					}
					var readErr error
					ciphertext, readErr = io.ReadAll(arg.file)
					if readErr != nil {
						return nil, loxerror.RuntimeError(name, readErr.Error())
					}
				case *LoxString:
					var decodeErr error
					ciphertext, decodeErr = LoxAgeEncryptionDecode(arg.str)
					if decodeErr != nil {
						return nil, loxerror.RuntimeError(name, decodeErr.Error())
					}
				}

				readBuffer := bytes.NewBuffer(ciphertext)
				r, err := age.Decrypt(readBuffer, identity)
				if err != nil {
					switch err.(type) {
					case *age.NoIdentityMatchError:
						return nil, loxerror.RuntimeError(name,
							"age symmetric.decrypt: incorrect password")
					default:
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				}

				writeBuffer := new(bytes.Buffer)
				_, err = io.Copy(writeBuffer, r)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}

				bytes := writeBuffer.Bytes()
				buffer := EmptyLoxBufferCap(int64(len(bytes)))
				for _, b := range bytes {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				}
				return buffer, nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		})
	case "decryptToFile":
		return ageSymFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 2, 3:
				switch args[0].(type) {
				case *LoxBuffer:
				case *LoxFile:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name,
						"First argument to 'age symmetric.decryptToFile' must be a buffer, file, or string.")
				}
				switch arg := args[1].(type) {
				case *LoxFile:
					if !arg.isWrite() && !arg.isAppend() {
						return nil, loxerror.RuntimeError(name,
							"Second file argument to 'age symmetric.decryptToFile' must be in write or append mode.")
					}
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'age symmetric.decryptToFile' must be a file or string.")
				}

				var password string = l.password
				if argsLen == 3 {
					switch arg := args[2].(type) {
					case *LoxString:
						password = arg.str
					default:
						return nil, loxerror.RuntimeError(name,
							"Third argument to 'age symmetric.decryptToFile' must be a string.")
					}
				} else if !l.initPassword {
					return nil, loxerror.RuntimeError(name,
						"Must specify password argument to 'age symmetric.decryptToFile'.")
				}
				if len(password) == 0 {
					return passwordErr(argsLen == 3)
				}

				identity, err := age.NewScryptIdentity(password)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
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
							"First file argument to 'age symmetric.decryptToFile' must be in read mode.")
					}
					var readErr error
					ciphertext, readErr = io.ReadAll(arg.file)
					if readErr != nil {
						return nil, loxerror.RuntimeError(name, readErr.Error())
					}
				case *LoxString:
					var decodeErr error
					ciphertext, decodeErr = LoxAgeEncryptionDecode(arg.str)
					if decodeErr != nil {
						return nil, loxerror.RuntimeError(name, decodeErr.Error())
					}
				}

				readBuffer := bytes.NewBuffer(ciphertext)
				r, err := age.Decrypt(readBuffer, identity)
				if err != nil {
					switch err.(type) {
					case *age.NoIdentityMatchError:
						return nil, loxerror.RuntimeError(name,
							"age symmetric.decryptToFile: incorrect password")
					default:
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				}

				writeBuffer := new(bytes.Buffer)
				_, err = io.Copy(writeBuffer, r)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}

				bytes := writeBuffer.Bytes()
				switch arg := args[1].(type) {
				case *LoxFile:
					_, err := arg.file.Write(bytes)
					if err != nil {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				case *LoxString:
					err := os.WriteFile(arg.str, bytes, 0666)
					if err != nil {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				}

				return nil, nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
			}
		})
	case "decryptToStr":
		return ageSymFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 1, 2:
				switch args[0].(type) {
				case *LoxBuffer:
				case *LoxFile:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name,
						"First argument to 'age symmetric.decryptToStr' must be a buffer, file, or string.")
				}

				var password string = l.password
				if argsLen == 2 {
					switch arg := args[1].(type) {
					case *LoxString:
						password = arg.str
					default:
						return nil, loxerror.RuntimeError(name,
							"Second argument to 'age symmetric.decryptToStr' must be a string.")
					}
				} else if !l.initPassword {
					return nil, loxerror.RuntimeError(name,
						"Must specify password argument to 'age symmetric.decryptToStr'.")
				}
				if len(password) == 0 {
					return passwordErr(argsLen == 2)
				}

				identity, err := age.NewScryptIdentity(password)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
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
							"File argument to 'age symmetric.decryptToStr' must be in read mode.")
					}
					var readErr error
					ciphertext, readErr = io.ReadAll(arg.file)
					if readErr != nil {
						return nil, loxerror.RuntimeError(name, readErr.Error())
					}
				case *LoxString:
					var decodeErr error
					ciphertext, decodeErr = LoxAgeEncryptionDecode(arg.str)
					if decodeErr != nil {
						return nil, loxerror.RuntimeError(name, decodeErr.Error())
					}
				}

				readBuffer := bytes.NewBuffer(ciphertext)
				r, err := age.Decrypt(readBuffer, identity)
				if err != nil {
					switch err.(type) {
					case *age.NoIdentityMatchError:
						return nil, loxerror.RuntimeError(name,
							"age symmetric.decryptToStr: incorrect password")
					default:
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				}

				writeBuffer := new(bytes.Buffer)
				_, err = io.Copy(writeBuffer, r)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}

				return NewLoxStringQuote(writeBuffer.String()), nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		})
	case "encrypt":
		return ageSymFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 1, 2:
				switch args[0].(type) {
				case *LoxBuffer:
				case *LoxFile:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name,
						"First argument to 'age symmetric.encrypt' must be a buffer, file, or string.")
				}

				var password string = l.password
				if argsLen == 2 {
					switch arg := args[1].(type) {
					case *LoxString:
						password = arg.str
					default:
						return nil, loxerror.RuntimeError(name,
							"Second argument to 'age symmetric.encrypt' must be a string.")
					}
				} else if !l.initPassword {
					return nil, loxerror.RuntimeError(name,
						"Must specify password argument to 'age symmetric.encrypt'.")
				}
				if len(password) == 0 {
					return passwordErr(argsLen == 2)
				}

				recipient, err := age.NewScryptRecipient(password)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
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
							"File argument to 'age symmetric.encrypt' must be in read mode.")
					}
					var readErr error
					plaintext, readErr = io.ReadAll(arg.file)
					if readErr != nil {
						return nil, loxerror.RuntimeError(name, readErr.Error())
					}
				case *LoxString:
					plaintext = []byte(arg.str)
				}

				bytesBuffer := new(bytes.Buffer)
				w, err := age.Encrypt(bytesBuffer, recipient)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				_, err = w.Write(plaintext)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				err = w.Close()
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}

				bytes := bytesBuffer.Bytes()
				buffer := EmptyLoxBufferCap(int64(len(bytes)))
				for _, b := range bytes {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				}
				return buffer, nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		})
	case "encryptToFile":
		return ageSymFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 2, 3:
				switch args[0].(type) {
				case *LoxBuffer:
				case *LoxFile:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name,
						"First argument to 'age symmetric.encryptToFile' must be a buffer, file, or string.")
				}
				switch arg := args[1].(type) {
				case *LoxFile:
					if !arg.isWrite() && !arg.isAppend() {
						return nil, loxerror.RuntimeError(name,
							"Second file argument to 'age symmetric.encryptToFile' must be in write or append mode.")
					}
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'age symmetric.encryptToFile' must be a file or string.")
				}

				var password string = l.password
				if argsLen == 3 {
					switch arg := args[2].(type) {
					case *LoxString:
						password = arg.str
					default:
						return nil, loxerror.RuntimeError(name,
							"Third argument to 'age symmetric.encryptToFile' must be a string.")
					}
				} else if !l.initPassword {
					return nil, loxerror.RuntimeError(name,
						"Must specify password argument to 'age symmetric.encryptToFile'.")
				}
				if len(password) == 0 {
					return passwordErr(argsLen == 3)
				}

				recipient, err := age.NewScryptRecipient(password)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
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
							"First file argument to 'age symmetric.encryptToFile' must be in read mode.")
					}
					var readErr error
					plaintext, readErr = io.ReadAll(arg.file)
					if readErr != nil {
						return nil, loxerror.RuntimeError(name, readErr.Error())
					}
				case *LoxString:
					plaintext = []byte(arg.str)
				}

				bytesBuffer := new(bytes.Buffer)
				w, err := age.Encrypt(bytesBuffer, recipient)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				_, err = w.Write(plaintext)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				err = w.Close()
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}

				bytes := bytesBuffer.Bytes()
				switch arg := args[1].(type) {
				case *LoxFile:
					_, err := arg.file.Write(bytes)
					if err != nil {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				case *LoxString:
					err := os.WriteFile(arg.str, bytes, 0666)
					if err != nil {
						return nil, loxerror.RuntimeError(name, err.Error())
					}
				}

				return nil, nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
			}
		})
	case "encryptToStr":
		return ageSymFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 1, 2:
				switch args[0].(type) {
				case *LoxBuffer:
				case *LoxFile:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name,
						"First argument to 'age symmetric.encryptToStr' must be a buffer, file, or string.")
				}

				var password string = l.password
				if argsLen == 2 {
					switch arg := args[1].(type) {
					case *LoxString:
						password = arg.str
					default:
						return nil, loxerror.RuntimeError(name,
							"Second argument to 'age symmetric.encryptToStr' must be a string.")
					}
				} else if !l.initPassword {
					return nil, loxerror.RuntimeError(name,
						"Must specify password argument to 'age symmetric.encryptToStr'.")
				}
				if len(password) == 0 {
					return passwordErr(argsLen == 2)
				}

				recipient, err := age.NewScryptRecipient(password)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
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
							"File argument to 'age symmetric.encryptToStr' must be in read mode.")
					}
					var readErr error
					plaintext, readErr = io.ReadAll(arg.file)
					if readErr != nil {
						return nil, loxerror.RuntimeError(name, readErr.Error())
					}
				case *LoxString:
					plaintext = []byte(arg.str)
				}

				bytesBuffer := new(bytes.Buffer)
				w, err := age.Encrypt(bytesBuffer, recipient)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				_, err = w.Write(plaintext)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				err = w.Close()
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}

				return NewLoxStringQuote(
					LoxAgeEncryptionEncode(bytesBuffer.Bytes()),
				), nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		})
	case "hasInitPassword":
		return ageSymFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.initPassword, nil
		})
	case "initPassword", "password":
		return ageSymFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.initPassword {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Cannot call 'age symmetric.%v' without having an initial password set.",
						methodName,
					),
				)
			}
			return NewLoxStringQuote(l.password), nil
		})
	case "removeInitPassword", "removePassword":
		return ageSymFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.password = ""
			l.initPassword = false
			return nil, nil
		})
	case "setInitPassword", "setPassword":
		return ageSymFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				if len(loxStr.str) == 0 {
					return nil, loxerror.RuntimeError(name,
						"String argument to 'age symmetric.setInitPassword' cannot be empty.")
				}
				l.password = loxStr.str
				l.initPassword = true
				return nil, nil
			}
			return argMustBeType("string")
		})
	}
	return nil, loxerror.RuntimeError(name, "age symmetric encryption objects have no property called '"+methodName+"'.")
}

func (l *LoxAgeSymmetric) String() string {
	return fmt.Sprintf("<age symmetric encryption object at %p>", l)
}

func (l *LoxAgeSymmetric) Type() string {
	return "age symmetric"
}
