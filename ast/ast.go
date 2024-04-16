package ast

import (
	"github.com/AlanLuu/lox/token"
)

type Expr interface{}
type Stmt interface{}

type Binary struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

type Expression struct {
	Expression Expr
}

type Grouping struct {
	Expression Expr
}

type Literal struct {
	Value any
}

type Print struct {
	Expression Expr
}

type Unary struct {
	Operator token.Token
	Right    Expr
}

type Var struct {
	Name        token.Token
	Initializer Expr
}

type Variable struct {
	Name token.Token
}
