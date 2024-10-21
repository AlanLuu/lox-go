package ast

import (
	"fmt"
	"math"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type Parser struct {
	tokens    list.List[*token.Token]
	current   int
	loopDepth int
}

func NewParser(tokens list.List[*token.Token]) *Parser {
	return &Parser{tokens, 0, 0}
}

func (p *Parser) advance() *token.Token {
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

func (p *Parser) assertStatement() (Stmt, error) {
	assertToken := p.previous()
	assertExpr, assertExprErr := p.expression()
	if assertExprErr != nil {
		return nil, assertExprErr
	}
	_, semiColonErr := p.consume(token.SEMICOLON, "Expected ';' after value.")
	if semiColonErr != nil {
		return nil, semiColonErr
	}
	return Assert{assertExpr, assertToken}, nil
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
			return SetObject{set}, nil
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
	for p.match(token.AMPERSAND) {
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
	for p.match(token.PIPE) {
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
			if p.match(token.COLON) {
				var indexEnd Expr
				if p.peek().TokenType != token.RIGHT_BRACKET {
					var indexEndErr error
					indexEnd, indexEndErr = p.expression()
					if indexEndErr != nil {
						return nil, indexEndErr
					}
				}
				rightBracket, rightBracketErr := p.consume(token.RIGHT_BRACKET, "Expected ']' after index.")
				if rightBracketErr != nil {
					return nil, rightBracketErr
				}
				expr = Index{
					IndexElement: expr,
					Bracket:      rightBracket,
					Index:        nil,
					IndexEnd:     indexEnd,
					IsSlice:      true,
				}
				continue
			}
			index, indexErr := p.expression()
			if indexErr != nil {
				return nil, indexErr
			}
			var indexEnd Expr
			isSlice := false
			if p.match(token.COLON) {
				isSlice = true
				if p.peek().TokenType != token.RIGHT_BRACKET {
					var indexEndErr error
					indexEnd, indexEndErr = p.expression()
					if indexEndErr != nil {
						return nil, indexEndErr
					}
				}
			}
			rightBracket, rightBracketErr := p.consume(token.RIGHT_BRACKET, "Expected ']' after index.")
			if rightBracketErr != nil {
				return nil, rightBracketErr
			}
			expr = Index{
				IndexElement: expr,
				Bracket:      rightBracket,
				Index:        index,
				IndexEnd:     indexEnd,
				IsSlice:      isSlice,
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

func (p *Parser) classDeclaration(canInstantiate bool) (Stmt, error) {
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
	classMethods := list.NewList[Function]()
	classFields := make(map[string]Expr)
	instanceFields := make(map[string]Expr)
	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		isStatic := false
		if p.match(token.STATIC) {
			isStatic = true
		}
		name, nameErr := p.consume(token.IDENTIFIER, "Expected identifier name.")
		if nameErr != nil {
			return nil, nameErr
		}
		if p.match(token.EQUAL) {
			expr, exprErr := p.expression()
			if exprErr != nil {
				return nil, exprErr
			}
			if isStatic {
				_, semiColonErr := p.consume(token.SEMICOLON, "Expected ';' after static field initializer.")
				if semiColonErr != nil {
					return nil, semiColonErr
				}
				classFields[name.Lexeme] = expr
			} else {
				_, semiColonErr := p.consume(token.SEMICOLON, "Expected ';' after instance field initializer.")
				if semiColonErr != nil {
					return nil, semiColonErr
				}
				instanceFields[name.Lexeme] = expr
			}
		} else {
			method, methodErr := p.functionBody("method", true)
			if methodErr != nil {
				return nil, methodErr
			}
			if isStatic {
				classMethods.Add(Function{Name: name, Function: method})
			} else {
				methods.Add(Function{Name: name, Function: method})
			}
		}
	}

	_, rightBraceErr := p.consume(token.RIGHT_BRACE, "Expected '}' after class body.")
	if rightBraceErr != nil {
		return nil, rightBraceErr
	}
	return Class{
		Name:           className,
		SuperClass:     superClass,
		Methods:        methods,
		ClassMethods:   classMethods,
		ClassFields:    classFields,
		InstanceFields: instanceFields,
		CanInstantiate: canInstantiate,
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

func (p *Parser) consume(tokenType token.TokenType, message string) (*token.Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}
	return nil, p.error(p.peek(), message)
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
		value, err = p.classDeclaration(true)
	case p.match(token.ENUM):
		value, err = p.enumDeclaration()
	case p.match(token.STATIC):
		_, classErr := p.consume(token.CLASS, "Expected 'class' after 'static'.")
		if classErr != nil {
			return nil, classErr
		}
		value, err = p.classDeclaration(false)
	default:
		value, err = p.statement(false)
	}
	if err != nil {
		p.synchronize()
		return nil, err
	}
	return value, nil
}

func (p *Parser) dict() (Expr, error) {
	entries := list.NewList[Expr]()
	trailingComma := false
	if !p.check(token.RIGHT_BRACE) {
		for cond := true; cond; cond = p.match(token.COMMA) {
			if p.match(token.RIGHT_BRACE) {
				trailingComma = true
				break
			}
			if p.match(token.ELLIPSIS) {
				spreadDict, spreadDictErr := p.or()
				if spreadDictErr != nil {
					return nil, spreadDictErr
				}
				entries.Add(Spread{spreadDict, p.previous()})
			} else {
				key, keyErr := p.or()
				if keyErr != nil {
					return nil, keyErr
				}
				_, colonErr := p.consume(token.COLON, "Expected ':' after dictionary key.")
				if colonErr != nil {
					return nil, colonErr
				}
				value, valueErr := p.or()
				if valueErr != nil {
					return nil, valueErr
				}
				entries.Add(key)
				entries.Add(value)
			}
		}
	}
	if !trailingComma {
		_, rightBraceErr := p.consume(token.RIGHT_BRACE, "Expected '}' after dict.")
		if rightBraceErr != nil {
			return nil, rightBraceErr
		}
	}
	return Dict{entries, p.previous()}, nil
}

func (p *Parser) isDict() bool {
	originalPos := p.current
	defer func() {
		p.current = originalPos
	}()

	if p.isAtEnd() {
		return false
	}
	if p.match(token.RIGHT_BRACE) || p.match(token.ELLIPSIS) {
		return true
	}
	_, exprErr := p.or()
	if exprErr != nil {
		return false
	}
	if p.match(token.COLON) {
		return true
	}
	return false
}

func (p *Parser) doWhileStatement() (Stmt, error) {
	p.loopDepth++
	defer func() {
		p.loopDepth--
	}()
	doToken := p.previous()
	body, bodyErr := p.statement(true)
	if bodyErr != nil {
		return nil, bodyErr
	}
	_, whileErr := p.consume(token.WHILE, "Expected 'while' after body.")
	if whileErr != nil {
		return nil, whileErr
	}
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
	_, semiColonErr := p.consume(token.SEMICOLON, "Expected ';' at end of 'do while' statement.")
	if semiColonErr != nil {
		return nil, semiColonErr
	}
	return DoWhile{condition, body, doToken}, nil
}

func (p *Parser) enumDeclaration() (Stmt, error) {
	enumName, enumNameErr := p.consume(token.IDENTIFIER, "Expected enum name.")
	if enumNameErr != nil {
		return nil, enumNameErr
	}
	_, leftBraceErr := p.consume(token.LEFT_BRACE, "Expected '{' before enum body.")
	if leftBraceErr != nil {
		return nil, leftBraceErr
	}

	trailingComma := false
	enumMembers := list.NewList[*token.Token]()
	if !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		for cond := true; cond; cond = p.match(token.COMMA) {
			if p.match(token.RIGHT_BRACE) {
				trailingComma = true
				break
			}
			enumMember, enumMemberErr := p.consume(token.IDENTIFIER, "Expected enum member name.")
			if enumMemberErr != nil {
				return nil, enumMemberErr
			}
			enumMembers.Add(enumMember)
		}
	}

	if !trailingComma {
		_, rightBraceErr := p.consume(token.RIGHT_BRACE, "Expected '}' after enum body.")
		if rightBraceErr != nil {
			return nil, rightBraceErr
		}
	}
	return Enum{enumName, enumMembers}, nil
}

func (p *Parser) error(theToken *token.Token, message string) error {
	var theError error
	if theToken.TokenType == token.EOF {
		theError = loxerror.GiveError(theToken.Line, " at end", message)
	} else {
		theError = loxerror.GiveError(theToken.Line, " at '"+theToken.Lexeme+"'", message)
	}
	return theError
}

func (p *Parser) expression() (Expr, error) {
	return p.ternary()
}

func (p *Parser) expressionStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, isAssign := expr.(Assign)
	_, isSet := expr.(Set)
	_, isSetObject := expr.(SetObject)
	if !util.StdinFromTerminal() || isAssign || isSet || isSetObject {
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
			spread := p.match(token.ELLIPSIS)
			expr, exprErr := p.expression()
			if exprErr != nil {
				arguments.Clear()
				return nil, exprErr
			}
			if spread {
				arguments.Add(Spread{expr, p.previous()})
			} else {
				arguments.Add(expr)
			}
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

	body, bodyErr := p.statement(true)
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

func (p *Parser) forEachStatement() (Stmt, error) {
	p.loopDepth++
	defer func() {
		p.loopDepth--
	}()

	forEachToken := p.previous()
	_, leftParenErr := p.consume(token.LEFT_PAREN, "Expected '(' after 'foreach'.")
	if leftParenErr != nil {
		return nil, leftParenErr
	}

	_, varErr := p.consume(token.VAR, "Expected 'var' after '('.")
	if varErr != nil {
		return nil, varErr
	}

	variableName, variableNameErr := p.consume(token.IDENTIFIER, "Expected identifier name after 'var'.")
	if variableNameErr != nil {
		return nil, variableNameErr
	}

	inErrMsg := "Expected 'in' after identifier name."
	if !p.match(token.IDENTIFIER) {
		return nil, p.error(p.peek(), inErrMsg)
	}

	inKeyword := p.previous()
	if inKeyword.Lexeme != "in" {
		return nil, p.error(inKeyword, inErrMsg)
	}

	iterable, iterableErr := p.expression()
	if iterableErr != nil {
		return nil, iterableErr
	}

	_, rightParenErr := p.consume(token.RIGHT_PAREN, "Expected ')' after expression.")
	if rightParenErr != nil {
		return nil, rightParenErr
	}

	body, bodyErr := p.statement(true)
	if bodyErr != nil {
		return nil, bodyErr
	}

	return ForEach{
		VariableName: variableName,
		Iterable:     iterable,
		Body:         body,
		ForEachToken: forEachToken,
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
	if p.loopDepth > 0 {
		prevLoopDepth := p.loopDepth
		defer func() {
			p.loopDepth = prevLoopDepth
		}()
		p.loopDepth = 0
	}
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

	parameters := list.NewList[*token.Token]()
	varArgPos := -1
	if !p.check(token.RIGHT_PAREN) {
		varArgPosCount := 0
		for cond := true; cond; cond = p.match(token.COMMA) {
			if len(parameters) >= 255 {
				loxerror.PrintErrorObject(p.error(p.peek(), "Can't have more than 255 parameters."))
			}
			if p.match(token.ELLIPSIS) {
				if varArgPos >= 0 {
					return emptyFuncNode, p.error(p.peek(), fmt.Sprintf("Can't have multiple varargs in %v.", kind))
				}
				varArgPos = varArgPosCount
			}
			paramName, paramNameErr := p.consume(token.IDENTIFIER, "Expected parameter name.")
			if paramNameErr != nil {
				parameters.Clear()
				return emptyFuncNode, paramNameErr
			}
			parameters.Add(paramName)
			varArgPosCount++
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
		Params:    parameters,
		Body:      block,
		VarArgPos: varArgPos,
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
	thenBranch, thenBranchErr := p.statement(true)
	if thenBranchErr != nil {
		return nil, thenBranchErr
	}
	var elseBranch Stmt
	if p.match(token.ELSE) {
		var elseBranchErr error
		elseBranch, elseBranchErr = p.statement(true)
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

func (p *Parser) importStatement() (Stmt, error) {
	importToken := p.previous()
	importFile, importFileErr := p.expression()
	if importFileErr != nil {
		return nil, importFileErr
	}
	expectedErrMsg := "Expected ';' or 'as' after value."
	importNamespace := ""
	if p.match(token.IDENTIFIER) {
		asKeyword := p.previous()
		if asKeyword.Lexeme == "as" {
			asIdentifier, asIdentifierErr := p.consume(token.IDENTIFIER,
				"Expected identifier name after 'as'.")
			if asIdentifierErr != nil {
				return nil, asIdentifierErr
			}
			importNamespace = asIdentifier.Lexeme
			expectedErrMsg = "Expected ';' after identifier."
		} else {
			return nil, p.error(asKeyword, expectedErrMsg)
		}
	}
	_, semiColonErr := p.consume(token.SEMICOLON, expectedErrMsg)
	if semiColonErr != nil {
		return nil, semiColonErr
	}
	return Import{importFile, importNamespace, importToken}, nil
}

func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == token.EOF
}

func (p *Parser) list() (Expr, error) {
	elements := list.NewList[Expr]()
	if !p.check(token.RIGHT_BRACKET) {
		for cond := true; cond; cond = p.match(token.COMMA) {
			spread := p.match(token.ELLIPSIS)
			expr, exprErr := p.expression()
			if exprErr != nil {
				elements.Clear()
				return nil, exprErr
			}
			if spread {
				elements.Add(Spread{expr, p.previous()})
			} else {
				elements.Add(expr)
			}
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

func (p *Parser) peek() *token.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() *token.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) primary() (Expr, error) {
	switch {
	case p.match(token.FALSE):
		return Literal{Value: false}, nil
	case p.match(token.TRUE):
		return Literal{Value: true}, nil
	case p.match(token.INFINITY):
		return Literal{Value: math.Inf(1)}, nil
	case p.match(token.NAN):
		return Literal{Value: math.NaN()}, nil
	case p.match(token.NIL):
		return Literal{Value: nil}, nil
	case p.match(token.NUMBER):
		return Literal{Value: p.previous().Literal}, nil
	case p.match(token.BIG_NUMBER):
		previous := p.previous()
		return BigNum{
			NumStr:  previous.Literal.(string),
			IsFloat: previous.Quote == 255,
		}, nil
	case p.match(token.STRING):
		previous := p.previous()
		return String{Str: previous.Literal.(string), Quote: previous.Quote}, nil
	case p.match(token.IDENTIFIER):
		return Variable{Name: p.previous()}, nil
	case p.match(token.FUN):
		return p.functionBody("function", false)
	case p.match(token.LEFT_BRACE):
		return p.dict()
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

func (p *Parser) printStatement(newLine bool) (Stmt, error) {
	value, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, consumeErr := p.consume(token.SEMICOLON, "Expected ';' after value.")
	if consumeErr != nil {
		return nil, consumeErr
	}
	return Print{Expression: value, NewLine: newLine}, nil
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

func (p *Parser) statement(alwaysBlock bool) (Stmt, error) {
	switch {
	case p.match(token.ASSERT):
		return p.assertStatement()
	case p.match(token.BREAK):
		return p.breakStatement()
	case p.match(token.CONTINUE):
		return p.continueStatement()
	case p.match(token.DO):
		return p.doWhileStatement()
	case p.match(token.FOR):
		return p.forStatement()
	case p.match(token.FOREACH):
		return p.forEachStatement()
	case p.match(token.IF):
		return p.ifStatement()
	case p.match(token.IMPORT):
		return p.importStatement()
	case p.match(token.PRINT):
		return p.printStatement(true)
	case p.match(token.PUT):
		return p.printStatement(false)
	case p.match(token.RETURN):
		return p.returnStatement()
	case p.match(token.THROW):
		return p.throwStatement()
	case p.match(token.TRY):
		return p.tryCatchFinallyStatement()
	case p.match(token.WHILE):
		return p.whileStatement()
	case p.match(token.LEFT_BRACE):
		if alwaysBlock || !p.isDict() {
			blockList, blockErr := p.block()
			if blockErr != nil {
				return nil, blockErr
			}
			return Block{Statements: blockList}, nil
		}
		p.current--
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

func (p *Parser) throwStatement() (Stmt, error) {
	throwToken := p.previous()
	throwExpr, throwExprErr := p.expression()
	if throwExprErr != nil {
		return nil, throwExprErr
	}
	_, semiColonErr := p.consume(token.SEMICOLON, "Expected ';' after value.")
	if semiColonErr != nil {
		return nil, semiColonErr
	}
	return Throw{throwExpr, throwToken}, nil
}

func (p *Parser) tryCatchFinallyStatement() (Stmt, error) {
	_, leftBraceErr := p.consume(token.LEFT_BRACE, "Expected '{' after 'try'.")
	if leftBraceErr != nil {
		return nil, leftBraceErr
	}
	tryBlockList, tryBlockListErr := p.block()
	if tryBlockListErr != nil {
		return nil, tryBlockListErr
	}

	var catchBlockList list.List[Stmt] = nil
	var catchName *token.Token
	foundCatchBlock := false
	if p.match(token.CATCH) {
		foundCatchBlock = true
		leftBraceErrMsg := "Expected '(' or '{' after 'catch'."
		if p.match(token.LEFT_PAREN) {
			var catchNameErr error
			catchName, catchNameErr = p.consume(token.IDENTIFIER, "Expected identifier name.")
			if catchNameErr != nil {
				return nil, catchNameErr
			}
			_, rightParenErr := p.consume(token.RIGHT_PAREN, "Expected ')' after identifier name.")
			if rightParenErr != nil {
				return nil, rightParenErr
			}
			leftBraceErrMsg = "Expected '{' before catch body."
		}
		_, leftBraceErr = p.consume(token.LEFT_BRACE, leftBraceErrMsg)
		if leftBraceErr != nil {
			return nil, leftBraceErr
		}
		var catchBlockListErr error
		catchBlockList, catchBlockListErr = p.block()
		if catchBlockListErr != nil {
			return nil, catchBlockListErr
		}
	}

	var finallyBlockList list.List[Stmt] = nil
	foundFinallyBlock := false
	if p.match(token.FINALLY) {
		foundFinallyBlock = true
		_, leftBraceErr = p.consume(token.LEFT_BRACE, "Expected '{' after 'finally'.")
		if leftBraceErr != nil {
			return nil, leftBraceErr
		}
		var finallyBlockListErr error
		finallyBlockList, finallyBlockListErr = p.block()
		if finallyBlockListErr != nil {
			return nil, finallyBlockListErr
		}
	}

	if !foundCatchBlock && !foundFinallyBlock {
		return nil, p.error(p.peek(), "Expected 'catch' or 'finally' after try block.")
	}
	if !foundCatchBlock {
		return TryCatchFinally{
			Block{Statements: tryBlockList},
			catchName,
			nil,
			Block{Statements: finallyBlockList},
		}, nil
	}
	if !foundFinallyBlock {
		return TryCatchFinally{
			Block{Statements: tryBlockList},
			catchName,
			Block{Statements: catchBlockList},
			nil,
		}, nil
	}
	return TryCatchFinally{
		Block{Statements: tryBlockList},
		catchName,
		Block{Statements: catchBlockList},
		Block{Statements: finallyBlockList},
	}, nil
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

func (p *Parser) ternary() (Expr, error) {
	condition, conditionErr := p.assignment()
	if conditionErr != nil {
		return nil, conditionErr
	}
	if p.match(token.QUESTION) {
		trueExpr, trueExprErr := p.expression()
		if trueExprErr != nil {
			return nil, trueExprErr
		}
		_, colonErr := p.consume(token.COLON, "Expected ':' after second expression in ternary operator.")
		if colonErr != nil {
			return nil, colonErr
		}
		falseExpr, falseExprErr := p.ternary()
		if falseExprErr != nil {
			return nil, falseExprErr
		}
		return Ternary{condition, trueExpr, falseExpr}, nil
	}
	return condition, nil
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
	body, bodyErr := p.statement(true)
	if bodyErr != nil {
		return nil, bodyErr
	}
	return While{
		Condition:  condition,
		Body:       body,
		WhileToken: whileToken,
	}, nil
}
