package ast

import (
	"errors"

	"github.com/AlanLuu/lox/ast/classtype"
	"github.com/AlanLuu/lox/ast/functiontype"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type Resolver struct {
	Interpreter     *Interpreter
	Scopes          list.List[map[string]bool]
	CurrentFunction functiontype.FunctionType
	CurrentClass    classtype.ClassType
}

func NewResolver(interpreter *Interpreter) *Resolver {
	return &Resolver{
		Interpreter:     interpreter,
		Scopes:          list.NewList[map[string]bool](),
		CurrentFunction: functiontype.NONE,
		CurrentClass:    classtype.NONE,
	}
}

func (r *Resolver) beginScope() {
	r.Scopes.Add(make(map[string]bool))
}

func (r *Resolver) declare(name token.Token) error {
	if r.Scopes.IsEmpty() {
		return nil
	}
	scope := r.Scopes.Peek()
	if _, ok := scope[name.Lexeme]; ok {
		return loxerror.RuntimeError(name, "Already a variable with this name in this scope.")
	}
	scope[name.Lexeme] = false
	return nil
}

func (r *Resolver) define(name token.Token) {
	if r.Scopes.IsEmpty() {
		return
	}
	scope := r.Scopes.Peek()
	scope[name.Lexeme] = true
}

func (r *Resolver) endScope() {
	r.Scopes.Pop()
}

func (r *Resolver) Resolve(statements list.List[Stmt]) error {
	for _, stmt := range statements {
		resolveErr := r.resolveStmt(stmt)
		if resolveErr != nil {
			return resolveErr
		}
	}
	return nil
}

func (r *Resolver) resolveExpr(expr Expr) error {
	switch expr := expr.(type) {
	case Assign:
		return r.visitAssignExpr(expr)
	case Binary:
		return r.visitBinaryExpr(expr)
	case Call:
		return r.visitCallExpr(expr)
	case FunctionExpr:
		return r.visitFunctionExpr(expr)
	case Get:
		return r.visitGetExpr(expr)
	case Grouping:
		return r.visitGroupingExpr(expr)
	case Index:
		return r.visitIndexExpr(expr)
	case List:
		return r.visitListExpr(expr)
	case Literal:
		return nil
	case Logical:
		return r.visitLogicalExpr(expr)
	case Set:
		return r.visitSetExpr(expr)
	case SetList:
		return r.visitSetExpr(expr.Set)
	case String:
		return nil
	case Super:
		return r.visitSuperExpr(expr)
	case This:
		return r.visitThisExpr(expr)
	case Unary:
		return r.visitUnaryExpr(expr)
	case Variable:
		return r.visitVariableExpr(expr)
	case nil:
		return nil
	}
	return errors.New("unknown expression found in resolver")
}

func (r *Resolver) resolveStmt(stmt Stmt) error {
	switch stmt := stmt.(type) {
	case Block:
		return r.visitBlockStmt(stmt)
	case Break, Continue:
		return nil
	case Class:
		return r.visitClassStmt(stmt)
	case For:
		return r.visitForStmt(stmt)
	case Function:
		return r.visitFunctionStmt(stmt)
	case Expression:
		return r.visitExpressionStmt(stmt)
	case If:
		return r.visitIfStmt(stmt)
	case Print:
		return r.visitPrintStmt(stmt)
	case Return:
		return r.visitReturnStmt(stmt)
	case Var:
		return r.visitVarStmt(stmt)
	case While:
		return r.visitWhileStmt(stmt)
	case nil:
		return nil
	}
	return errors.New("unknown statement found in resolver")
}

func (r *Resolver) resolveFunction(fnExpr FunctionExpr, fnType functiontype.FunctionType) error {
	enclosingFunction := r.CurrentFunction
	r.CurrentFunction = fnType
	defer func() {
		r.CurrentFunction = enclosingFunction
	}()

	r.beginScope()
	defer r.endScope()
	for _, param := range fnExpr.Params {
		declareErr := r.declare(param)
		if declareErr != nil {
			return declareErr
		}
		r.define(param)
	}
	return r.Resolve(fnExpr.Body)
}

func (r *Resolver) resolveLocal(expr Expr, name token.Token) {
	for i := len(r.Scopes) - 1; i >= 0; i-- {
		scope := r.Scopes[i]
		if _, ok := scope[name.Lexeme]; ok {
			r.Interpreter.Resolve(expr, len(r.Scopes)-1-i)
			return
		}
	}
}

func (r *Resolver) visitAssignExpr(expr Assign) error {
	resolveErr := r.resolveExpr(expr.Value)
	if resolveErr != nil {
		return resolveErr
	}
	r.resolveLocal(expr, expr.Name)
	return nil
}

func (r *Resolver) visitBinaryExpr(expr Binary) error {
	resolveErr := r.resolveExpr(expr.Left)
	if resolveErr != nil {
		return resolveErr
	}
	return r.resolveExpr(expr.Right)
}

func (r *Resolver) visitBlockStmt(stmt Block) error {
	r.beginScope()
	defer r.endScope()
	return r.Resolve(stmt.Statements)
}

func (r *Resolver) visitCallExpr(expr Call) error {
	resolveErr := r.resolveExpr(expr.Callee)
	if resolveErr != nil {
		return resolveErr
	}
	for _, argument := range expr.Arguments {
		resolveErr = r.resolveExpr(argument)
		if resolveErr != nil {
			return resolveErr
		}
	}
	return nil
}

func (r *Resolver) visitClassStmt(stmt Class) error {
	enclosingClass := r.CurrentClass
	r.CurrentClass = classtype.CLASS
	defer func() {
		r.CurrentClass = enclosingClass
	}()

	declareErr := r.declare(stmt.Name)
	if declareErr != nil {
		return declareErr
	}

	r.define(stmt.Name)
	if stmt.SuperClass != nil {
		if stmt.Name.Lexeme == stmt.SuperClass.Name.Lexeme {
			return loxerror.RuntimeError(stmt.SuperClass.Name, "A class can't inherit from itself.")
		}
		r.CurrentClass = classtype.SUBCLASS
		resolveErr := r.resolveExpr(*stmt.SuperClass)
		if resolveErr != nil {
			return resolveErr
		}
		r.beginScope()
		r.Scopes.Peek()["super"] = true
	}
	r.beginScope()
	r.Scopes.Peek()["this"] = true
	for _, method := range stmt.Methods {
		declaration := functiontype.METHOD
		if method.Name.Lexeme == "init" {
			declaration = functiontype.INITIALIZER
		}
		resolveErr := r.resolveFunction(method.Function, declaration)
		if resolveErr != nil {
			return resolveErr
		}
	}
	for _, method := range stmt.ClassMethods {
		r.beginScope()
		r.Scopes.Peek()["this"] = true
		resolveErr := r.resolveFunction(method.Function, functiontype.METHOD)
		if resolveErr != nil {
			return resolveErr
		}
		r.endScope()
	}
	for _, field := range stmt.ClassFields {
		resolveErr := r.resolveExpr(field)
		if resolveErr != nil {
			return resolveErr
		}
	}
	for _, field := range stmt.InstanceFields {
		resolveErr := r.resolveExpr(field)
		if resolveErr != nil {
			return resolveErr
		}
	}

	r.endScope()
	if stmt.SuperClass != nil {
		r.endScope()
	}
	return nil
}

func (r *Resolver) visitExpressionStmt(stmt Expression) error {
	return r.resolveExpr(stmt.Expression)
}

func (r *Resolver) visitForStmt(stmt For) error {
	r.beginScope()
	defer r.endScope()
	resolveErr := r.resolveStmt(stmt.Initializer)
	if resolveErr != nil {
		return resolveErr
	}
	resolveErr = r.resolveExpr(stmt.Condition)
	if resolveErr != nil {
		return resolveErr
	}
	resolveErr = r.resolveExpr(stmt.Increment)
	if resolveErr != nil {
		return resolveErr
	}
	return r.resolveStmt(stmt.Body)
}

func (r *Resolver) visitFunctionExpr(expr FunctionExpr) error {
	return r.resolveFunction(expr, functiontype.FUNCTION)
}

func (r *Resolver) visitFunctionStmt(stmt Function) error {
	declareErr := r.declare(stmt.Name)
	if declareErr != nil {
		return declareErr
	}
	r.define(stmt.Name)
	return r.resolveFunction(stmt.Function, functiontype.FUNCTION)
}

func (r *Resolver) visitGetExpr(expr Get) error {
	return r.resolveExpr(expr.Object)
}

func (r *Resolver) visitGroupingExpr(expr Grouping) error {
	return r.resolveExpr(expr.Expression)
}

func (r *Resolver) visitIndexExpr(expr Index) error {
	resolveErr := r.resolveExpr(expr.IndexElement)
	if resolveErr != nil {
		return resolveErr
	}
	resolveErr = r.resolveExpr(expr.Index)
	if resolveErr != nil {
		return resolveErr
	}
	return r.resolveExpr(expr.IndexEnd)
}

func (r *Resolver) visitListExpr(expr List) error {
	for _, element := range expr.Elements {
		resolveErr := r.resolveExpr(element)
		if resolveErr != nil {
			return resolveErr
		}
	}
	return nil
}

func (r *Resolver) visitIfStmt(stmt If) error {
	resolveErr := r.resolveExpr(stmt.Condition)
	if resolveErr != nil {
		return resolveErr
	}
	resolveErr = r.resolveStmt(stmt.ThenBranch)
	if resolveErr != nil {
		return resolveErr
	}
	if stmt.ElseBranch != nil {
		resolveErr = r.resolveStmt(stmt.ElseBranch)
		if resolveErr != nil {
			return resolveErr
		}
	}
	return nil
}

func (r *Resolver) visitLogicalExpr(expr Logical) error {
	resolveErr := r.resolveExpr(expr.Left)
	if resolveErr != nil {
		return resolveErr
	}
	return r.resolveExpr(expr.Right)
}

func (r *Resolver) visitPrintStmt(stmt Print) error {
	return r.resolveExpr(stmt.Expression)
}

func (r *Resolver) visitReturnStmt(stmt Return) error {
	if r.CurrentFunction == functiontype.NONE {
		return loxerror.RuntimeError(stmt.Keyword, "Can't return from top-level code.")
	}
	if stmt.Value != nil {
		if r.CurrentFunction == functiontype.INITIALIZER {
			return loxerror.RuntimeError(stmt.Keyword, "Can't return a value from an initializer.")
		}
		return r.resolveExpr(stmt.Value)
	}
	return nil
}

func (r *Resolver) visitSetExpr(expr Set) error {
	resolveErr := r.resolveExpr(expr.Value)
	if resolveErr != nil {
		return resolveErr
	}
	return r.resolveExpr(expr.Object)
}

func (r *Resolver) visitSuperExpr(expr Super) error {
	if r.CurrentClass == classtype.NONE {
		return loxerror.RuntimeError(expr.Keyword, "Can't use 'super' outside of a class.")
	} else if r.CurrentClass != classtype.SUBCLASS {
		return loxerror.RuntimeError(expr.Keyword, "Can't use 'super' in a class with no superclass.")
	}
	r.resolveLocal(expr, expr.Keyword)
	return nil
}

func (r *Resolver) visitThisExpr(expr This) error {
	if r.CurrentClass == classtype.NONE {
		return loxerror.RuntimeError(expr.Keyword, "Can't use 'this' outside of a class.")
	}
	r.resolveLocal(expr, expr.Keyword)
	return nil
}

func (r *Resolver) visitUnaryExpr(expr Unary) error {
	return r.resolveExpr(expr.Right)
}

func (r *Resolver) visitVarStmt(stmt Var) error {
	declareErr := r.declare(stmt.Name)
	if declareErr != nil {
		return declareErr
	}
	if stmt.Initializer != nil {
		resolveErr := r.resolveExpr(stmt.Initializer)
		if resolveErr != nil {
			return resolveErr
		}
	}
	r.define(stmt.Name)
	return nil
}

func (r *Resolver) visitVariableExpr(expr Variable) error {
	if !r.Scopes.IsEmpty() {
		scopes := r.Scopes.Peek()
		value, ok := scopes[expr.Name.Lexeme]
		if ok && !value {
			return loxerror.RuntimeError(expr.Name, "Can't read local variable in its own initializer.")
		}
	}
	r.resolveLocal(expr, expr.Name)
	return nil
}

func (r *Resolver) visitWhileStmt(stmt While) error {
	resolveErr := r.resolveExpr(stmt.Condition)
	if resolveErr != nil {
		return resolveErr
	}
	return r.resolveStmt(stmt.Body)
}
