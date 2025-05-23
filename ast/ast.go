package ast

import (
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/token"
)

type Expr interface{}
type Stmt interface{}

type Assert struct {
	Value       Expr
	AssertToken *token.Token
}

type Assign struct {
	Name  *token.Token
	Value Expr
}

type BigNum struct {
	NumStr  string
	IsFloat bool
}

type Binary struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
}

type Break struct{}

type Block struct {
	Statements list.List[Stmt]
}

type Call struct {
	Callee    Expr
	Paren     *token.Token
	Arguments list.List[Expr]
}

type Class struct {
	Name           *token.Token
	SuperClass     *Variable
	Methods        list.List[Function]
	ClassMethods   list.List[Function]
	ClassFields    map[string]Expr
	InstanceFields map[string]Expr
	CanInstantiate bool
}

type Continue struct{}

type Dict struct {
	Entries   list.List[Expr]
	DictToken *token.Token
}

type DoWhile struct {
	Condition Expr
	Body      Stmt
	DoToken   *token.Token
}

type Enum struct {
	Name    *token.Token
	Members list.List[*token.Token]
}

type Expression struct {
	Expression Expr
}

type For struct {
	Initializer Stmt
	Condition   Expr
	Increment   Expr
	Body        Stmt
	ForToken    *token.Token
}

type ForEach struct {
	VariableName *token.Token
	Iterable     Expr
	Body         Stmt
	ForEachToken *token.Token
}

type Function struct {
	Name     *token.Token
	Function FunctionExpr
}

type FunctionExpr struct {
	Params    list.List[*token.Token]
	Body      list.List[Stmt]
	VarArgPos int
}

type Get struct {
	Object Expr
	Name   *token.Token
}

type Grouping struct {
	Expression Expr
}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

type Import struct {
	ImportFile      Expr
	ImportNamespace string
	ImportToken     *token.Token
}

type Index struct {
	IndexElement Expr
	Bracket      *token.Token
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
	Operator *token.Token
	Right    Expr
}

type Loop struct {
	LoopBlock Stmt
	LoopToken *token.Token
}

type Print struct {
	Expression Expr
	NewLine    bool
	Stderr     bool
}

type Repeat struct {
	Expression  Expr
	Body        Stmt
	RepeatToken *token.Token
}

type Return struct {
	Keyword    *token.Token
	Value      Expr
	FinalValue any
}

type Set struct {
	Object Expr
	Name   *token.Token
	Value  Expr
}

type SetObject struct {
	Set
}

type Spread struct {
	Iterable    Expr
	SpreadToken *token.Token
}

type String struct {
	Str   string
	Quote byte
}

type Super struct {
	Keyword *token.Token
	Method  *token.Token
}

type Ternary struct {
	Condition Expr
	TrueExpr  Expr
	FalseExpr Expr
}

type This struct {
	Keyword *token.Token
}

type Throw struct {
	Value      Expr
	ThrowToken *token.Token
}

type TryCatchFinally struct {
	TryBlock     Stmt
	CatchName    *token.Token
	CatchBlock   Stmt
	FinallyBlock Stmt
}

type Unary struct {
	Operator *token.Token
	Right    Expr
}

type Var struct {
	Name        *token.Token
	Initializer Expr
}

type Variable struct {
	Name *token.Token
}

type While struct {
	Condition  Expr
	Body       Stmt
	WhileToken *token.Token
}
