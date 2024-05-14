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

type Break struct{}

type Block struct {
	Statements list.List[Stmt]
}

type Call struct {
	Callee    Expr
	Paren     token.Token
	Arguments list.List[Expr]
}

type Class struct {
	Name           token.Token
	SuperClass     *Variable
	Methods        list.List[Function]
	ClassMethods   list.List[Function]
	ClassFields    map[string]Expr
	InstanceFields map[string]Expr
}

type Continue struct{}

type Expression struct {
	Expression Expr
}

type For struct {
	Initializer Stmt
	Condition   Expr
	Increment   Expr
	Body        Stmt
	ForToken    token.Token
}

type Function struct {
	Name     token.Token
	Function FunctionExpr
}

type FunctionExpr struct {
	Params list.List[token.Token]
	Body   list.List[Stmt]
}

type Get struct {
	Object Expr
	Name   token.Token
}

type Grouping struct {
	Expression Expr
}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

type Index struct {
	IndexElement Expr
	Bracket      token.Token
	Index        Expr
	IndexEnd     Expr
	IsSlice      bool
}

type List struct {
	Elements list.List[Expr]
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

type Return struct {
	Keyword    token.Token
	Value      Expr
	FinalValue any
}

type Set struct {
	Object Expr
	Name   token.Token
	Value  Expr
}

type SetList struct {
	Set
}

type String struct {
	Str   string
	Quote byte
}

type Super struct {
	Keyword token.Token
	Method  token.Token
}

type This struct {
	Keyword token.Token
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
