package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type Parser struct {
	tokens    list.List[token.Token]
	current   int
	loopDepth int
}

func NewParser(tokens list.List[token.Token]) *Parser {
	return &Parser{tokens, 0, 0}
}

func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) and() (Expr, error) {
	expr, exprErr := p.equality()
	if exprErr != nil {
		return nil, exprErr
	}
	for p.match(token.AND) {
		operator := p.previous()
		right, equalityErr := p.equality()
		if equalityErr != nil {
			return nil, equalityErr
		}
		expr = Logical{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) assignment() (Expr, error) {
	expr, exprErr := p.or()
	if exprErr != nil {
		return nil, exprErr
	}

	if p.match(token.EQUAL) {
		equals := p.previous()
		value, valueErr := p.assignment()
		if valueErr != nil {
			return nil, valueErr
		}
		switch expr := expr.(type) {
		case Variable:
			return Assign{Name: expr.Name, Value: value}, nil
		case Get:
			return Set{
				Object: expr.Object,
				Name:   expr.Name,
				Value:  value,
			}, nil
		case Index:
			set := Set{
				Object: expr,
				Name:   expr.Bracket,
				Value:  value,
			}
			return SetList{set}, nil
		}
		return nil, p.error(equals, "Invalid assignment target.")
	}

	return expr, nil
}

func (p *Parser) bitShift() (Expr, error) {
	expr, termErr := p.term()
	if termErr != nil {
		return nil, termErr
	}
	for p.match(token.DOUBLE_LESS, token.DOUBLE_GREATER) {
		operator := p.previous()
		right, termErr := p.term()
		if termErr != nil {
			return nil, termErr
		}
		expr = Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) bitwiseAnd() (Expr, error) {
	expr, bitShiftErr := p.bitShift()
	if bitShiftErr != nil {
		return nil, bitShiftErr
	}
	for p.match(token.AND_SYMBOL) {
		operator := p.previous()
		right, bitShiftErr := p.bitShift()
		if bitShiftErr != nil {
			return nil, bitShiftErr
		}
		expr = Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) bitwiseOr() (Expr, error) {
	expr, bitwiseXorErr := p.bitwiseXor()
	if bitwiseXorErr != nil {
		return nil, bitwiseXorErr
	}
	for p.match(token.OR_SYMBOL) {
		operator := p.previous()
		right, bitwiseXorErr := p.bitwiseXor()
		if bitwiseXorErr != nil {
			return nil, bitwiseXorErr
		}
		expr = Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) bitwiseXor() (Expr, error) {
	expr, bitwiseAndErr := p.bitwiseAnd()
	if bitwiseAndErr != nil {
		return nil, bitwiseAndErr
	}
	for p.match(token.CARET) {
		operator := p.previous()
		right, bitwiseAndErr := p.bitwiseAnd()
		if bitwiseAndErr != nil {
			return nil, bitwiseAndErr
		}
		expr = Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) block() (list.List[Stmt], error) {
	statements := list.NewList[Stmt]()
	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		declaration, declarationErr := p.declaration()
		if declarationErr != nil {
			statements.Clear()
			return statements, declarationErr
		}
		statements.Add(declaration)
	}
	_, consumeErr := p.consume(token.RIGHT_BRACE, "Expected '}' after block.")
	if consumeErr != nil {
		statements.Clear()
		return statements, consumeErr
	}
	return statements, nil
}

func (p *Parser) breakStatement() (Stmt, error) {
	breakToken := p.previous()
	if p.loopDepth <= 0 {
		return nil, p.error(breakToken, "Illegal break statement.")
	}
	_, consumeErr := p.consume(token.SEMICOLON, "Expected ';' after 'break'.")
	if consumeErr != nil {
		return nil, consumeErr
	}
	return Break{}, nil
}

func (p *Parser) call() (Expr, error) {
	expr, exprErr := p.primary()
	if exprErr != nil {
		return nil, exprErr
	}
	for {
		if p.match(token.LEFT_PAREN) {
			finishCallExpr, finishCallExprErr := p.finishCall(expr)
			if finishCallExprErr != nil {
				return nil, finishCallExprErr
			}
			expr = finishCallExpr
		} else if p.match(token.DOT) {
			name, nameErr := p.consume(token.IDENTIFIER, "Expected property name after '.'.")
			if nameErr != nil {
				return nil, nameErr
			}
			expr = Get{Object: expr, Name: name}
		} else if p.match(token.LEFT_BRACKET) {
			index, indexErr := p.expression()
			if indexErr != nil {
				return nil, indexErr
			}
			rightBracket, rightBracketErr := p.consume(token.RIGHT_BRACKET, "Expected ']' after index.")
			if rightBracketErr != nil {
				return nil, rightBracketErr
			}
			expr = Index{
				IndexElement: expr,
				Bracket:      rightBracket,
				Index:        index,
			}
		} else {
			break
		}
	}
	return expr, nil
}

func (p *Parser) check(tokenType token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().TokenType == tokenType
}

func (p *Parser) classDeclaration() (Stmt, error) {
	className, classNameErr := p.consume(token.IDENTIFIER, "Expected class name.")
	if classNameErr != nil {
		return nil, classNameErr
	}

	var superClass *Variable
	if p.match(token.LESS) {
		_, superClassNameErr := p.consume(token.IDENTIFIER, "Expected superclass name.")
		if superClassNameErr != nil {
			return nil, superClassNameErr
		}
		superClass = &Variable{p.previous()}
	}

	_, leftBraceErr := p.consume(token.LEFT_BRACE, "Expected '{' before class body.")
	if leftBraceErr != nil {
		return nil, leftBraceErr
	}

	methods := list.NewList[Function]()
	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		methodName, methodNameErr := p.consume(token.IDENTIFIER, "Expected method name.")
		if methodNameErr != nil {
			return nil, methodNameErr
		}
		method, methodErr := p.functionBody("method", true)
		if methodErr != nil {
			return nil, methodErr
		}
		methods.Add(Function{Name: methodName, Function: method})
	}

	_, rightBraceErr := p.consume(token.RIGHT_BRACE, "Expected '}' after class body.")
	if rightBraceErr != nil {
		return nil, rightBraceErr
	}
	return Class{
		Name:       className,
		SuperClass: superClass,
		Methods:    methods,
	}, nil
}

func (p *Parser) comparison() (Expr, error) {
	expr, bitwiseOrErr := p.bitwiseOr()
	if bitwiseOrErr != nil {
		return nil, bitwiseOrErr
	}
	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.previous()
		right, bitwiseOrErr := p.bitwiseOr()
		if bitwiseOrErr != nil {
			return nil, bitwiseOrErr
		}
		expr = Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) consume(tokenType token.TokenType, message string) (token.Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}
	return token.Token{}, p.error(p.peek(), message)
}

func (p *Parser) continueStatement() (Stmt, error) {
	continueToken := p.previous()
	if p.loopDepth <= 0 {
		return nil, p.error(continueToken, "Illegal continue statement.")
	}
	_, consumeErr := p.consume(token.SEMICOLON, "Expected ';' after 'continue'.")
	if consumeErr != nil {
		return nil, consumeErr
	}
	return Continue{}, nil
}

func (p *Parser) declaration() (Stmt, error) {
	var value Stmt
	var err error
	switch {
	case p.match(token.VAR):
		value, err = p.varDeclaration()
	case p.match(token.FUN):
		value, err = p.function("function")
	case p.match(token.CLASS):
		value, err = p.classDeclaration()
	default:
		value, err = p.statement()
	}
	if err != nil {
		p.synchronize()
		return nil, err
	}
	return value, nil
}

func (p *Parser) error(theToken token.Token, message string) error {
	var theError error
	if theToken.TokenType == token.EOF {
		theError = loxerror.GiveError(theToken.Line, " at end", message)
	} else {
		theError = loxerror.GiveError(theToken.Line, " at '"+theToken.Lexeme+"'", message)
	}
	return theError
}

func (p *Parser) expression() (Expr, error) {
	return p.assignment()
}

func (p *Parser) expressionStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, isAssign := expr.(Assign)
	_, isSet := expr.(Set)
	_, isSetList := expr.(SetList)
	if !util.StdinFromTerminal() || isAssign || isSet || isSetList {
		_, consumeErr := p.consume(token.SEMICOLON, "Expected ';' after expression.")
		if consumeErr != nil {
			return nil, consumeErr
		}
	} else {
		p.consume(token.SEMICOLON, "")
	}
	return Expression{Expression: expr}, nil
}

func (p *Parser) equality() (Expr, error) {
	expr, comparisonErr := p.comparison()
	if comparisonErr != nil {
		return nil, comparisonErr
	}
	for p.match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.previous()
		right, comparisonErr := p.comparison()
		if comparisonErr != nil {
			return nil, comparisonErr
		}
		expr = Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) exponent() (Expr, error) {
	expr, callErr := p.call()
	if callErr != nil {
		return nil, callErr
	}
	for p.match(token.DOUBLE_STAR) {
		operator := p.previous()
		right, unaryErr := p.unary()
		if unaryErr != nil {
			return nil, unaryErr
		}
		expr = Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) factor() (Expr, error) {
	expr, unaryErr := p.unary()
	if unaryErr != nil {
		return nil, unaryErr
	}
	for p.match(token.SLASH, token.STAR, token.PERCENT) {
		operator := p.previous()
		right, unaryErr := p.unary()
		if unaryErr != nil {
			return nil, unaryErr
		}
		expr = Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) finishCall(callee Expr) (Expr, error) {
	arguments := list.NewList[Expr]()
	if !p.check(token.RIGHT_PAREN) {
		for cond := true; cond; cond = p.match(token.COMMA) {
			if len(arguments) >= 255 {
				loxerror.PrintErrorObject(p.error(p.peek(), "Can't have more than 255 arguments."))
			}
			expr, exprErr := p.expression()
			if exprErr != nil {
				arguments.Clear()
				return nil, exprErr
			}
			arguments.Add(expr)
		}
	}
	paren, parenErr := p.consume(token.RIGHT_PAREN, "Expected ')' after arguments.")
	if parenErr != nil {
		return nil, parenErr
	}
	return Call{
		Callee:    callee,
		Paren:     paren,
		Arguments: arguments,
	}, nil
}

func (p *Parser) forStatement() (Stmt, error) {
	p.loopDepth++
	defer func() {
		p.loopDepth--
	}()

	forToken := p.previous()
	_, leftParenErr := p.consume(token.LEFT_PAREN, "Expected '(' after 'for'.")
	if leftParenErr != nil {
		return nil, leftParenErr
	}

	var initializer Stmt
	var initializerErr error
	switch {
	case p.match(token.SEMICOLON):
		initializer = nil
	case p.match(token.VAR):
		initializer, initializerErr = p.varDeclaration()
		if initializerErr != nil {
			return nil, initializerErr
		}
	default:
		initializer, initializerErr = p.expressionStatement()
		if initializerErr != nil {
			return nil, initializerErr
		}
	}

	var condition Expr
	var conditionErr error
	if !p.check(token.SEMICOLON) {
		condition, conditionErr = p.expression()
		if conditionErr != nil {
			return nil, conditionErr
		}
	}
	_, conditionSemicolonErr := p.consume(token.SEMICOLON, "Expected ';' after loop condition.")
	if conditionSemicolonErr != nil {
		return nil, conditionSemicolonErr
	}

	var increment Expr
	var incrementErr error
	if !p.check(token.RIGHT_PAREN) {
		increment, incrementErr = p.expression()
		if incrementErr != nil {
			return nil, incrementErr
		}
	}
	_, incrementSemicolonErr := p.consume(token.RIGHT_PAREN, "Expected ')' after for clauses.")
	if incrementSemicolonErr != nil {
		return nil, incrementSemicolonErr
	}

	body, bodyErr := p.statement()
	if bodyErr != nil {
		return nil, bodyErr
	}
	return For{
		Initializer: initializer,
		Condition:   condition,
		Increment:   increment,
		Body:        body,
		ForToken:    forToken,
	}, nil
}

func (p *Parser) function(kind string) (Function, error) {
	emptyFuncNode := Function{}
	name, nameErr := p.consume(token.IDENTIFIER, fmt.Sprintf("Expected %v name.", kind))
	if nameErr != nil {
		return emptyFuncNode, nameErr
	}
	funcBody, funcBodyErr := p.functionBody(kind, true)
	if funcBodyErr != nil {
		return emptyFuncNode, funcBodyErr
	}
	return Function{Name: name, Function: funcBody}, nil
}

func (p *Parser) functionBody(kind string, funcHasName bool) (FunctionExpr, error) {
	emptyFuncNode := FunctionExpr{}
	if funcHasName {
		_, leftParenErr := p.consume(token.LEFT_PAREN, fmt.Sprintf("Expected '(' after %v name.", kind))
		if leftParenErr != nil {
			return emptyFuncNode, leftParenErr
		}
	} else {
		var leftParenErr error
		if kind == "function" {
			_, leftParenErr = p.consume(token.LEFT_PAREN, "Expected '(' after 'fun'.")
		} else {
			_, leftParenErr = p.consume(token.LEFT_PAREN, "Expected '('.")
		}
		if leftParenErr != nil {
			return emptyFuncNode, leftParenErr
		}
	}

	parameters := list.NewList[token.Token]()
	if !p.check(token.RIGHT_PAREN) {
		for cond := true; cond; cond = p.match(token.COMMA) {
			if len(parameters) >= 255 {
				loxerror.PrintErrorObject(p.error(p.peek(), "Can't have more than 255 parameters."))
			}
			paramName, paramNameErr := p.consume(token.IDENTIFIER, "Expected parameter name.")
			if paramNameErr != nil {
				parameters.Clear()
				return emptyFuncNode, paramNameErr
			}
			parameters.Add(paramName)
		}
	}

	_, rightParenErr := p.consume(token.RIGHT_PAREN, "Expected ')' after parameters.")
	if rightParenErr != nil {
		return emptyFuncNode, rightParenErr
	}

	var block list.List[Stmt]
	var blockErr error
	if !funcHasName && p.match(token.ARROW) {
		expr, exprErr := p.expression()
		if exprErr != nil {
			return emptyFuncNode, exprErr
		}
		block = list.NewList[Stmt]()
		block.Add(Return{Value: expr})
	} else {
		var leftBraceErrMsg string
		if funcHasName {
			leftBraceErrMsg = fmt.Sprintf("Expected '{' before %v body.", kind)
		} else {
			leftBraceErrMsg = fmt.Sprintf("Expected '{' or '=>' before %v body.", kind)
		}
		_, leftBraceErr := p.consume(token.LEFT_BRACE, leftBraceErrMsg)
		if leftBraceErr != nil {
			return emptyFuncNode, leftBraceErr
		}
		block, blockErr = p.block()
		if blockErr != nil {
			return emptyFuncNode, blockErr
		}
	}
	return FunctionExpr{
		Params: parameters,
		Body:   block,
	}, nil
}

func (p *Parser) ifStatement() (Stmt, error) {
	_, leftParenErr := p.consume(token.LEFT_PAREN, "Expected '(' after 'if'.")
	if leftParenErr != nil {
		return nil, leftParenErr
	}
	condition, conditionErr := p.expression()
	if conditionErr != nil {
		return nil, conditionErr
	}
	_, rightParenErr := p.consume(token.RIGHT_PAREN, "Expected ')' after if condition.")
	if rightParenErr != nil {
		return nil, rightParenErr
	}
	thenBranch, thenBranchErr := p.statement()
	if thenBranchErr != nil {
		return nil, thenBranchErr
	}
	var elseBranch Stmt
	if p.match(token.ELSE) {
		var elseBranchErr error
		elseBranch, elseBranchErr = p.statement()
		if elseBranchErr != nil {
			return nil, elseBranchErr
		}
	}
	return If{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil
}

func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == token.EOF
}

func (p *Parser) list() (Expr, error) {
	elements := list.NewList[Expr]()
	if !p.check(token.RIGHT_BRACKET) {
		for cond := true; cond; cond = p.match(token.COMMA) {
			expr, exprErr := p.expression()
			if exprErr != nil {
				elements.Clear()
				return nil, exprErr
			}
			elements.Add(expr)
		}
	}
	_, rightBracketErr := p.consume(token.RIGHT_BRACKET, "Expected ']' after list.")
	if rightBracketErr != nil {
		return nil, rightBracketErr
	}
	return List{Elements: elements}, nil
}

func (p *Parser) match(tokenTypes ...token.TokenType) bool {
	for _, tokenType := range tokenTypes {
		if p.check(tokenType) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) or() (Expr, error) {
	expr, exprErr := p.and()
	if exprErr != nil {
		return nil, exprErr
	}
	for p.match(token.OR) {
		operator := p.previous()
		right, andErr := p.and()
		if andErr != nil {
			return nil, andErr
		}
		expr = Logical{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) Parse() (list.List[Stmt], error) {
	statements := list.NewList[Stmt]()
	for !p.isAtEnd() {
		statement, err := p.declaration()
		if err != nil {
			return statements, err
		}
		statements.Add(statement)
	}
	return statements, nil
}

func (p *Parser) peek() token.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() token.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) primary() (Expr, error) {
	switch {
	case p.match(token.FALSE):
		return Literal{Value: false}, nil
	case p.match(token.TRUE):
		return Literal{Value: true}, nil
	case p.match(token.NIL):
		return Literal{Value: nil}, nil
	case p.match(token.NUMBER):
		return Literal{Value: p.previous().Literal}, nil
	case p.match(token.STRING):
		previous := p.previous()
		return Literal{Value: previous.Literal, Quote: previous.Quote}, nil
	case p.match(token.IDENTIFIER):
		return Variable{Name: p.previous()}, nil
	case p.match(token.FUN):
		return p.functionBody("function", false)
	case p.match(token.LEFT_BRACKET):
		return p.list()
	case p.match(token.SUPER):
		keyword := p.previous()
		_, dotErr := p.consume(token.DOT, "Expected '.' after 'super'.")
		if dotErr != nil {
			return nil, dotErr
		}
		method, methodErr := p.consume(token.IDENTIFIER, "Expected superclass method name.")
		if methodErr != nil {
			return nil, methodErr
		}
		return Super{Keyword: keyword, Method: method}, nil
	case p.match(token.THIS):
		return This{Keyword: p.previous()}, nil
	case p.match(token.LEFT_PAREN):
		expr, expressionErr := p.expression()
		if expressionErr != nil {
			return nil, expressionErr
		}
		_, consumeErr := p.consume(token.RIGHT_PAREN, "Expected ')' after expression.")
		if consumeErr != nil {
			return nil, consumeErr
		}
		return Grouping{Expression: expr}, nil
	}
	return nil, p.error(p.peek(), "Expected expression.")
}

func (p *Parser) printStatement() (Stmt, error) {
	value, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, consumeErr := p.consume(token.SEMICOLON, "Expected ';' after value.")
	if consumeErr != nil {
		return nil, consumeErr
	}
	return Print{Expression: value}, nil
}

func (p *Parser) returnStatement() (Stmt, error) {
	keyword := p.previous()
	var value Expr
	var valueErr error
	if !p.check(token.SEMICOLON) {
		value, valueErr = p.expression()
		if valueErr != nil {
			return nil, valueErr
		}
	}
	_, consumeErr := p.consume(token.SEMICOLON, "Expected ';' after return value.")
	if consumeErr != nil {
		return nil, consumeErr
	}
	return Return{Keyword: keyword, Value: value}, nil
}

func (p *Parser) statement() (Stmt, error) {
	switch {
	case p.match(token.BREAK):
		return p.breakStatement()
	case p.match(token.CONTINUE):
		return p.continueStatement()
	case p.match(token.FOR):
		return p.forStatement()
	case p.match(token.IF):
		return p.ifStatement()
	case p.match(token.PRINT):
		return p.printStatement()
	case p.match(token.RETURN):
		return p.returnStatement()
	case p.match(token.WHILE):
		return p.whileStatement()
	case p.match(token.LEFT_BRACE):
		blockList, blockErr := p.block()
		if blockErr != nil {
			return nil, blockErr
		}
		return Block{Statements: blockList}, nil
	}
	return p.expressionStatement()
}

func (p *Parser) synchronize() {
	p.advance()
	for !p.isAtEnd() {
		if p.previous().TokenType == token.SEMICOLON {
			return
		}

		switch p.peek().TokenType {
		case token.CLASS:
			fallthrough
		case token.FUN:
			fallthrough
		case token.VAR:
			fallthrough
		case token.FOR:
			fallthrough
		case token.IF:
			fallthrough
		case token.WHILE:
			fallthrough
		case token.PRINT:
			fallthrough
		case token.RETURN:
			return
		}

		p.advance()
	}
}

func (p *Parser) term() (Expr, error) {
	expr, factorErr := p.factor()
	if factorErr != nil {
		return nil, factorErr
	}
	for p.match(token.MINUS, token.PLUS) {
		operator := p.previous()
		right, comparisonErr := p.factor()
		if comparisonErr != nil {
			return nil, comparisonErr
		}
		expr = Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) unary() (Expr, error) {
	if p.match(token.BANG, token.MINUS, token.TILDE) {
		operator := p.previous()
		right, unaryErr := p.unary()
		if unaryErr != nil {
			return nil, unaryErr
		}
		return Unary{
			Operator: operator,
			Right:    right,
		}, nil
	}
	return p.exponent()
}

func (p *Parser) varDeclaration() (Stmt, error) {
	name, varConsumeErr := p.consume(token.IDENTIFIER, "Expected variable name.")
	if varConsumeErr != nil {
		return nil, varConsumeErr
	}

	var initializer Expr = nil
	var initializerErr error = nil
	if p.match(token.EQUAL) {
		initializer, initializerErr = p.expression()
		if initializerErr != nil {
			return nil, initializerErr
		}
	}

	_, semiConsumeErr := p.consume(token.SEMICOLON, "Expected ';' after variable declaration.")
	if semiConsumeErr != nil {
		return nil, semiConsumeErr
	}

	return Var{Name: name, Initializer: initializer}, nil
}

func (p *Parser) whileStatement() (Stmt, error) {
	p.loopDepth++
	defer func() {
		p.loopDepth--
	}()
	whileToken := p.previous()
	_, leftParenErr := p.consume(token.LEFT_PAREN, "Expected '(' after 'while'.")
	if leftParenErr != nil {
		return nil, leftParenErr
	}
	condition, conditionErr := p.expression()
	if conditionErr != nil {
		return nil, conditionErr
	}
	_, rightParenErr := p.consume(token.RIGHT_PAREN, "Expected ')' after condition.")
	if rightParenErr != nil {
		return nil, rightParenErr
	}
	body, bodyErr := p.statement()
	if bodyErr != nil {
		return nil, bodyErr
	}
	return While{
		Condition:  condition,
		Body:       body,
		WhileToken: whileToken,
	}, nil
}
