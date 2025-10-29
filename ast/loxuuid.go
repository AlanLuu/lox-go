package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/google/uuid"
)

var ClockSequenceUUIDs = map[uuid.Version]struct{}{
	1: {},
	2: {},
}
var TimeBasedUUIDs = map[uuid.Version]struct{}{
	1: {},
	2: {},
	6: {},
	7: {},
}

type LoxUUID struct {
	uuid    uuid.UUID
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxUUID(theUUID uuid.UUID) *LoxUUID {
	return &LoxUUID{
		uuid:    theUUID,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxUUIDParse(str string) (*LoxUUID, error) {
	newUUID, err := uuid.Parse(str)
	if err != nil {
		return nil, err
	}
	return &LoxUUID{
		uuid:    newUUID,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func NewLoxUUIDV4Random() (*LoxUUID, error) {
	randUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return &LoxUUID{
		uuid:    randUUID,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func (l *LoxUUID) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxUUID:
		if l == obj {
			return true
		}
		return l.uuid == obj.uuid
	default:
		return false
	}
}

func (l *LoxUUID) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	uuidFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native uuid fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	switch methodName {
	case "bytes":
		return uuidFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			buffer := EmptyLoxBufferCap(16)
			for _, b := range l.uuid {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "clockSequence":
		return uuidFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if _, ok := ClockSequenceUUIDs[l.uuid.Version()]; !ok {
				return nil, loxerror.RuntimeError(name,
					"uuid.clockSequence: current UUID must be a version 1 or 2 UUID.")
			}
			return int64(l.uuid.ClockSequence()), nil
		})
	case "string":
		return uuidFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.uuid.String()), nil
		})
	case "time":
		return uuidFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if _, ok := TimeBasedUUIDs[l.uuid.Version()]; !ok {
				return nil, loxerror.RuntimeError(name,
					"uuid.time: current UUID must be a version 1, 2, 6, or 7 UUID.")
			}
			return int64(l.uuid.Time()), nil
		})
	case "urn":
		return uuidFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.uuid.URN()), nil
		})
	case "variant":
		return uuidFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.uuid.Variant()), nil
		})
	case "variantStr", "variantString":
		return uuidFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.uuid.Variant().String()), nil
		})
	case "version":
		return uuidFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.uuid.Version()), nil
		})
	case "versionStr", "versionString":
		return uuidFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.uuid.Version().String()), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "UUIDs have no property called '"+methodName+"'.")
}

func (l *LoxUUID) String() string {
	return fmt.Sprintf("<UUID id=%v>", l.uuid.String())
}

func (l *LoxUUID) Type() string {
	return "uuid"
}
