package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxEnumMember struct {
	name string
	enum *LoxEnum
}

func (l *LoxEnumMember) String() string {
	return l.enum.name + "." + l.name
}

func (l *LoxEnumMember) Type() string {
	return l.enum.name
}

type LoxEnum struct {
	name    string
	members map[string]*LoxEnumMember
}

func NewLoxEnum(name string, members map[string]*LoxEnumMember) *LoxEnum {
	return &LoxEnum{
		name:    name,
		members: members,
	}
}

func (l *LoxEnum) Get(name token.Token) (any, error) {
	enumMember, ok := l.members[name.Lexeme]
	if !ok {
		return nil, loxerror.RuntimeError(name,
			fmt.Sprintf("Unknown enum member '%v.%v'.", l.name, name.Lexeme))
	}
	return enumMember, nil
}

func (l *LoxEnum) String() string {
	return fmt.Sprintf("<enum %v at %p>", l.name, l)
}

func (l *LoxEnum) Type() string {
	return "enum"
}
