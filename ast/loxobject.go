package ast

import "github.com/AlanLuu/lox/token"

type LoxObject interface {
	Get(name token.Token) (any, error)
}
