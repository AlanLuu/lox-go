package ast

import (
	"fmt"
	"hash"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxHash struct {
	hash       hash.Hash
	hashFunc   func() hash.Hash
	hashType   string
	properties map[string]any
}

func NewLoxHash(theHash hash.Hash, hashFunc func() hash.Hash, hashType string) *LoxHash {
	return &LoxHash{
		hash:       theHash,
		hashFunc:   hashFunc,
		hashType:   hashType,
		properties: make(map[string]any),
	}
}

func (l *LoxHash) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if field, ok := l.properties[lexemeName]; ok {
		return field, nil
	}
	hashField := func(field any) (any, error) {
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = field
		}
		return field, nil
	}
	hashFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native hash fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'hash.%v' must be a %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch lexemeName {
	case "blockSize":
		return int64(l.hash.BlockSize()), nil
	case "digest":
		return hashFunc(0, func(in *Interpreter, _ list.List[any]) (any, error) {
			digest := l.hash.Sum(nil)
			buffer := EmptyLoxBufferCap(int64(len(digest)))
			for _, element := range digest {
				addErr := buffer.add(int64(element))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "hex":
		fallthrough
	case "hexDigest":
		return hashFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			hexDigest := fmt.Sprintf("%x", l.hash.Sum(nil))
			return NewLoxString(hexDigest, '\''), nil
		})
	case "reset":
		return hashFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.hash.Reset()
			return l, nil
		})
	case "size":
		return int64(l.hash.Size()), nil
	case "type":
		return hashField(NewLoxString(l.hashType, '\''))
	case "update":
		return hashFunc(1, func(in *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case *LoxBuffer:
				bytes := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					bytes = append(bytes, byte(element.(int64)))
				}
				l.hash.Write(bytes)
			case *LoxString:
				l.hash.Write([]byte(arg.str))
			default:
				return argMustBeType("buffer or string")
			}
			return l, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Hashes have no property called '"+lexemeName+"'.")
}

func (l *LoxHash) String() string {
	return fmt.Sprintf("<%v hash object at %p>", l.hashType, l)
}

func (l *LoxHash) Type() string {
	return "hash"
}
