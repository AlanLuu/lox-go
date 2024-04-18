package ast

import (
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type Parser struct {
	tokens  list.List[token.Token]
	current int
}

func NewParser(tokens list.List[token.Token]) *Parser {
	return &Parser{tokens, 0}
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
		}
		return nil, p.error(equals, "Invalid assignment target.")
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

func (p *Parser) check(tokenType token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().TokenType == tokenType
}

func (p *Parser) comparison() (Expr, error) {
	expr, termErr := p.term()
	if termErr != nil {
		return nil, termErr
	}
	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
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

func (p *Parser) consume(tokenType token.TokenType, message string) (token.Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}
	return token.Token{}, p.error(p.peek(), message)
}

func (p *Parser) declaration() (Stmt, error) {
	var value Stmt
	var err error
	if p.match(token.VAR) {
		value, err = p.varDeclaration()
	} else {
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
	if _, ok := expr.(Assign); !util.StdinFromTerminal() || ok {
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

func (p *Parser) factor() (Expr, error) {
	expr, unaryErr := p.unary()
	if unaryErr != nil {
		return nil, unaryErr
	}
	for p.match(token.SLASH, token.STAR) {
		operator := p.previous()
		right, comparisonErr := p.unary()
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
	case p.match(token.NUMBER, token.STRING):
		return Literal{Value: p.previous().Literal}, nil
	case p.match(token.IDENTIFIER):
		return Variable{Name: p.previous()}, nil
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

func (p *Parser) statement() (Stmt, error) {
	if p.match(token.IF) {
		return p.ifStatement()
	}
	if p.match(token.PRINT) {
		return p.printStatement()
	}
	if p.match(token.WHILE) {
		return p.whileStatement()
	}
	if p.match(token.LEFT_BRACE) {
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
	if p.match(token.BANG, token.MINUS) {
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
	return p.primary()
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
