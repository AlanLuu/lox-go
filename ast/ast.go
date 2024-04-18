package ast

import (
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/token"
)

type Expr interface{}
type Stmt interface{}

type Assign struct {
	Name  token.Token
	Value Expr
}

type Binary struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

type Block struct {
	Statements list.List[Stmt]
}

type Expression struct {
	Expression Expr
}

type Grouping struct {
	Expression Expr
}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

type Literal struct {
	Value any
}

type Logical struct {
	Left     Expr
	Operator token.Token
	Right    Expr
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

type While struct {
	Condition  Expr
	Body       Stmt
	WhileToken token.Token
}
