package ast

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"filippo.io/age"
	"filippo.io/age/armor"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxAgeAsymmetric struct {
	privKey      *age.X25519Identity
	pubKey       *age.X25519Recipient
	creationDate time.Time
	methods      map[string]*struct{ ProtoLoxCallable }
}

func NewLoxAgeAsymmetric() (*LoxAgeAsymmetric, error) {
	privKey, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, err
	}
	return NewLoxAgeAsymmetricPrivKey(privKey, true), nil
}

func NewLoxAgeAsymmetricPrivKey(privKey *age.X25519Identity, isNew bool) *LoxAgeAsymmetric {
	var creationDate time.Time
	if isNew {
		creationDate = time.Now()
	}
	return &LoxAgeAsymmetric{
		privKey:      privKey,
		pubKey:       privKey.Recipient(),
		creationDate: creationDate,
		methods:      make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxAgeAsymmetricPrivKeyStr(str string) (*LoxAgeAsymmetric, error) {
	privKey, err := age.ParseX25519Identity(str)
	if err != nil {
		return nil, err
	}
	return NewLoxAgeAsymmetricPrivKey(privKey, false), nil
}

func NewLoxAgeAsymmetricPubKey(pubKey *age.X25519Recipient) *LoxAgeAsymmetric {
	return &LoxAgeAsymmetric{
		privKey:      nil,
		pubKey:       pubKey,
		creationDate: time.Time{},
		methods:      make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxAgeAsymmetricPubKeyStr(str string) (*LoxAgeAsymmetric, error) {
	pubKey, err := age.ParseX25519Recipient(str)
	if err != nil {
		return nil, err
	}
	return NewLoxAgeAsymmetricPubKey(pubKey), nil
}

func (l *LoxAgeAsymmetric) hasCreationDate() bool {
	return l.creationDate != time.Time{}
}

func (l *LoxAgeAsymmetric) isKeyPair() bool {
	return l.privKey != nil
}

func (l *LoxAgeAsymmetric) isKeyPairWithCreationDate() bool {
	return l.isKeyPair() && l.hasCreationDate()
}

func (l *LoxAgeAsymmetric) toPubKey() {
	if l.privKey != nil {
		l.privKey = nil
	}
}

func (l *LoxAgeAsymmetric) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	ageAsymFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native age asymmetric encryption fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'age asymmetric.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	callMustBeKeypair := func() (any, error) {
		return nil, loxerror.RuntimeError(
			name,
			fmt.Sprintf(
				"Can only call 'age asymmetric.%v' on age asymmetric keypairs.",
				methodName,
			),
		)
	}
	switch methodName {
	case "creationDate":
		return ageAsymFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPairWithCreationDate() {
				return nil, loxerror.RuntimeError(name,
					"Can only call 'age asymmetric.creationDate' on newly-generated age asymmetric keypairs.")
			}
			return NewLoxDate(l.creationDate), nil
		})
	case "decrypt":
		return ageAsymFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return argMustBeType("buffer, file, or string")
			}

			if !l.isKeyPair() {
				return callMustBeKeypair()
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
						"File argument to 'age asymmetric.decrypt' must be in read mode.")
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
			r, err := age.Decrypt(readBuffer, l.privKey)
			if err != nil {
				switch err.(type) {
				case *age.NoIdentityMatchError:
					return nil, loxerror.RuntimeError(name,
						"age asymmetric.decrypt: failed to decrypt using current private key")
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
		})
	case "decryptPEM", "decryptPem":
		return ageAsymFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return argMustBeType("buffer, file, or string")
			}

			if !l.isKeyPair() {
				return callMustBeKeypair()
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
						"File argument to 'age asymmetric.decryptPEM' must be in read mode.")
				}
				var readErr error
				ciphertext, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				ciphertext = []byte(arg.str)
			}

			readBuffer := bytes.NewBuffer(ciphertext)
			pemReader := armor.NewReader(readBuffer)
			r, err := age.Decrypt(pemReader, l.privKey)
			if err != nil {
				switch err.(type) {
				case *age.NoIdentityMatchError:
					return nil, loxerror.RuntimeError(name,
						"age asymmetric.decryptPEM: failed to decrypt using current private key")
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
		})
	case "decryptPEMToFile", "decryptPemToFile":
		return ageAsymFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"First argument to 'age asymmetric.decryptPEMToFile' must be a buffer, file, or string.")
			}
			switch arg := args[1].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Second file argument to 'age asymmetric.decryptPEMToFile' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'age asymmetric.decryptPEMToFile' must be a file or string.")
			}

			if !l.isKeyPair() {
				return callMustBeKeypair()
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
						"File argument to 'age asymmetric.decryptPEMToFile' must be in read mode.")
				}
				var readErr error
				ciphertext, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				ciphertext = []byte(arg.str)
			}

			readBuffer := bytes.NewBuffer(ciphertext)
			pemReader := armor.NewReader(readBuffer)
			r, err := age.Decrypt(pemReader, l.privKey)
			if err != nil {
				switch err.(type) {
				case *age.NoIdentityMatchError:
					return nil, loxerror.RuntimeError(name,
						"age asymmetric.decryptPEMToFile: failed to decrypt using current private key")
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
		})
	case "decryptPEMToStr", "decryptPemToStr":
		return ageAsymFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return argMustBeType("buffer, file, or string")
			}

			if !l.isKeyPair() {
				return callMustBeKeypair()
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
						"File argument to 'age asymmetric.decryptPEMToStr' must be in read mode.")
				}
				var readErr error
				ciphertext, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				ciphertext = []byte(arg.str)
			}

			readBuffer := bytes.NewBuffer(ciphertext)
			pemReader := armor.NewReader(readBuffer)
			r, err := age.Decrypt(pemReader, l.privKey)
			if err != nil {
				switch err.(type) {
				case *age.NoIdentityMatchError:
					return nil, loxerror.RuntimeError(name,
						"age asymmetric.decryptPEMToStr: failed to decrypt using current private key")
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
		})
	case "decryptToFile":
		return ageAsymFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"First argument to 'age asymmetric.decryptToFile' must be a buffer, file, or string.")
			}
			switch arg := args[1].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Second file argument to 'age asymmetric.decryptToFile' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'age asymmetric.decryptToFile' must be a file or string.")
			}

			if !l.isKeyPair() {
				return callMustBeKeypair()
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
						"File argument to 'age asymmetric.decryptToFile' must be in read mode.")
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
			r, err := age.Decrypt(readBuffer, l.privKey)
			if err != nil {
				switch err.(type) {
				case *age.NoIdentityMatchError:
					return nil, loxerror.RuntimeError(name,
						"age asymmetric.decryptToFile: failed to decrypt using current private key")
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
		})
	case "decryptToStr":
		return ageAsymFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return argMustBeType("buffer, file, or string")
			}

			if !l.isKeyPair() {
				return callMustBeKeypair()
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
						"File argument to 'age asymmetric.decryptToStr' must be in read mode.")
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
			r, err := age.Decrypt(readBuffer, l.privKey)
			if err != nil {
				switch err.(type) {
				case *age.NoIdentityMatchError:
					return nil, loxerror.RuntimeError(name,
						"age asymmetric.decryptToStr: failed to decrypt using current private key")
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
		})
	case "encrypt":
		return ageAsymFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
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
						"File argument to 'age asymmetric.encrypt' must be in read mode.")
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

			bytesBuffer := new(bytes.Buffer)
			w, err := age.Encrypt(bytesBuffer, l.pubKey)
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
		})
	case "encryptPEMToFile", "encryptPemToFile":
		return ageAsymFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"First argument to 'age asymmetric.encryptPEMToFile' must be a buffer, file, or string.")
			}
			switch arg := args[1].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Second file argument to 'age asymmetric.encryptPEMToFile' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'age asymmetric.encryptPEMToFile' must be a file or string.")
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
						"File argument to 'age asymmetric.encryptPEMToFile' must be in read mode.")
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
			pemWriter := armor.NewWriter(bytesBuffer)
			w, err := age.Encrypt(pemWriter, l.pubKey)
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
			err = pemWriter.Close()
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
		})
	case "encryptToFile":
		return ageAsymFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"First argument to 'age asymmetric.encryptToFile' must be a buffer, file, or string.")
			}
			switch arg := args[1].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Second file argument to 'age asymmetric.encryptToFile' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'age asymmetric.encryptToFile' must be a file or string.")
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
						"File argument to 'age asymmetric.encryptToFile' must be in read mode.")
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
			w, err := age.Encrypt(bytesBuffer, l.pubKey)
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
		})
	case "encryptToPEM", "encryptToPem":
		return ageAsymFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
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
						"File argument to 'age asymmetric.encryptToPEM' must be in read mode.")
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

			builder := new(strings.Builder)
			pemWriter := armor.NewWriter(builder)
			w, err := age.Encrypt(pemWriter, l.pubKey)
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
			err = pemWriter.Close()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}

			return NewLoxStringQuote(builder.String()), nil
		})
	case "encryptToStr":
		return ageAsymFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
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
						"File argument to 'age asymmetric.encryptToStr' must be in read mode.")
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

			bytesBuffer := new(bytes.Buffer)
			w, err := age.Encrypt(bytesBuffer, l.pubKey)
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
		})
	case "export":
		return ageAsymFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			getOutput := func() []byte {
				var builder bytes.Buffer
				if l.isKeyPairWithCreationDate() {
					timeFormat := l.creationDate.Format(time.RFC3339)
					builder.WriteString(fmt.Sprintf("# created: %s\n", timeFormat))
				}
				builder.WriteString(fmt.Sprintf("# public key: %s\n", l.pubKey))
				if l.isKeyPair() {
					builder.WriteString(fmt.Sprintf("%s\n", l.privKey))
				}
				return builder.Bytes()
			}
			switch arg := args[0].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'age asymmetric.export' must be in write or append mode.")
				}
				_, err := arg.file.Write(getOutput())
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			case *LoxString:
				err := os.WriteFile(arg.str, getOutput(), 0666)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			default:
				return argMustBeType("file or string")
			}
			return nil, nil
		})
	case "hasCreationDate":
		return ageAsymFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.hasCreationDate(), nil
		})
	case "isKeyPair":
		return ageAsymFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isKeyPair(), nil
		})
	case "isNewKeyPair":
		return ageAsymFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isKeyPairWithCreationDate(), nil
		})
	case "privKey":
		return ageAsymFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isKeyPair() {
				return callMustBeKeypair()
			}
			return NewLoxStringQuote(l.privKey.String()), nil
		})
	case "pubKey":
		return ageAsymFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.pubKey.String()), nil
		})
	case "toPubKey":
		return ageAsymFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.toPubKey()
			return l, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "age asymmetric encryption objects have no property called '"+methodName+"'.")
}

func (l *LoxAgeAsymmetric) String() string {
	return fmt.Sprintf("<age asymmetric encryption object at %p>", l)
}

func (l *LoxAgeAsymmetric) Type() string {
	return "age asymmetric"
}
