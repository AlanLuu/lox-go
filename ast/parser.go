package ast

import (
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
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
	return p.equality()
}

func (p *Parser) expressionStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, consumeErr := p.consume(token.SEMICOLON, "Expected ';' after expression.")
	if consumeErr != nil {
		return nil, consumeErr
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

func (p *Parser) Parse() (list.List[Stmt], error) {
	statements := list.NewList[Stmt]()
	for !p.isAtEnd() {
		statement, err := p.statement()
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
	if p.match(token.PRINT) {
		return p.printStatement()
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
