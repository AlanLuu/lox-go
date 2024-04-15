package ast

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/token"
)

type Interpreter struct{}

func (i Interpreter) evaluate(expr any) (any, error) {
	switch expr := expr.(type) {
	case Expression:
		return i.visitExpressionStmt(expr)
	case Print:
		return i.visitPrintingStmt(expr)
	case Binary:
		return i.VisitBinaryExpr(expr)
	case Grouping:
		return i.VisitGroupingExpr(expr)
	case Literal:
		return i.VisitLiteralExpr(expr)
	case Unary:
		return i.VisitUnaryExpr(expr)
	}
	return nil, errors.New("critical error: unknown type found in AST")
}

func (i Interpreter) Interpret(statements list.List[Stmt]) error {
	for _, statement := range statements {
		_, evalErr := i.evaluate(statement)
		if evalErr != nil {
			return evalErr
		}
	}
	return nil
}

func (i Interpreter) isTruthy(obj any) bool {
	switch obj := obj.(type) {
	case nil:
		return false
	case bool:
		return obj
	case int64:
		return obj != 0
	case string:
		return len(obj) > 0
	}
	return true
}

func printResult(source any) {
	switch source := source.(type) {
	case nil:
		fmt.Println("nil")
	case float64:
		if math.IsInf(source, 1) {
			fmt.Println("Infinity")
		} else if math.IsInf(source, -1) {
			fmt.Println("-Infinity")
		} else {
			fmt.Println(source)
		}
	case string:
		if len(source) == 0 {
			fmt.Println("\"\"")
		} else {
			fmt.Println(source)
		}
	default:
		fmt.Println(source)
	}
}

func runtimeError(theToken token.Token, message string) error {
	errorStr := message + "\n[line " + fmt.Sprint(theToken.Line) + "]"
	return errors.New(errorStr)
}

func (i Interpreter) VisitBinaryExpr(expr Binary) (any, error) {
	floatIsInt := func(f float64) bool {
		return f == float64(int64(f))
	}
	runtimeErrorWrapper := func(message string) error {
		return runtimeError(expr.Operator, message)
	}
	unknownOpStr := "unknown operator"
	unknownOp := func() error {
		return runtimeErrorWrapper(unknownOpStr)
	}
	handleNumString := func(left float64, right string) (any, error) {
		switch expr.Operator.TokenType {
		case token.PLUS:
			if math.IsInf(left, 1) {
				return "Infinity" + right, nil
			} else if math.IsInf(left, -1) {
				return "-Infinity" + right, nil
			}
			return strconv.FormatFloat(left, 'f', -1, 64) + right, nil
		case token.STAR:
			if left <= 0 {
				return "", nil
			}
			if floatIsInt(left) {
				return strings.Repeat(right, int(left)), nil
			}
		}
		return math.NaN(), nil
	}
	handleTwoFloats := func(left float64, right float64) (any, error) {
		var result any
		switch expr.Operator.TokenType {
		case token.PLUS:
			result = left + right
		case token.MINUS:
			result = left - right
		case token.STAR:
			result = left * right
		case token.SLASH:
			result = left / right
		case token.LESS:
			result = left < right
		case token.LESS_EQUAL:
			result = left <= right
		case token.GREATER:
			result = left > right
		case token.GREATER_EQUAL:
			result = left >= right
		default:
			return nil, unknownOp()
		}

		switch result := result.(type) {
		case float64:
			if !math.IsInf(result, 1) &&
				!math.IsInf(result, -1) &&
				floatIsInt(left) &&
				floatIsInt(right) &&
				floatIsInt(result) {
				return int64(result), nil
			}
		}

		return result, nil
	}
	boolMap := map[bool]float64{
		true:  1,
		false: 0,
	}

	left, leftErr := i.evaluate(expr.Left)
	if leftErr != nil {
		return nil, leftErr
	}
	right, rightErr := i.evaluate(expr.Right)
	if rightErr != nil {
		return nil, rightErr
	}

	if expr.Operator.TokenType == token.EQUAL_EQUAL {
		return left == right, nil
	}
	if expr.Operator.TokenType == token.BANG_EQUAL {
		return left != right, nil
	}
	switch left := left.(type) {
	case int64:
		switch right := right.(type) {
		case int64:
			return handleTwoFloats(float64(left), float64(right))
		case float64:
			return handleTwoFloats(float64(left), float64(right))
		case bool:
			return handleTwoFloats(float64(left), boolMap[right])
		case string:
			return handleNumString(float64(left), right)
		}
	case float64:
		switch right := right.(type) {
		case int64:
			return handleTwoFloats(float64(left), float64(right))
		case float64:
			return handleTwoFloats(float64(left), float64(right))
		case bool:
			return handleTwoFloats(float64(left), boolMap[right])
		case string:
			return handleNumString(left, right)
		}
	case bool:
		switch right := right.(type) {
		case int64:
			return handleTwoFloats(boolMap[left], float64(right))
		case float64:
			return handleTwoFloats(boolMap[left], float64(right))
		case bool:
			return handleTwoFloats(boolMap[left], boolMap[right])
		case string:
			switch expr.Operator.TokenType {
			case token.PLUS:
				return strconv.FormatBool(left) + right, nil
			case token.STAR:
				return handleNumString(boolMap[left], right)
			}
			return math.NaN(), nil
		}
	case string:
		switch expr.Operator.TokenType {
		case token.PLUS:
			switch right := right.(type) {
			case int64:
				return left + strconv.FormatFloat(float64(right), 'f', -1, 64), nil
			case float64:
				if math.IsInf(right, 1) {
					return "Infinity" + left, nil
				} else if math.IsInf(right, -1) {
					return "-Infinity" + left, nil
				}
				return left + strconv.FormatFloat(right, 'f', -1, 64), nil
			case bool:
				return left + strconv.FormatBool(right), nil
			case string:
				return left + right, nil
			}
		case token.STAR:
			repeat := func(left string, right int64) (string, error) {
				if right <= 0 {
					return "", nil
				}
				return strings.Repeat(left, int(right)), nil
			}
			switch right := right.(type) {
			case int64:
				return repeat(left, right)
			case bool:
				return repeat(left, int64(boolMap[right]))
			}
		}
		return math.NaN(), nil
	}

	return nil, runtimeErrorWrapper("operands must be numbers, strings, or booleans")
}

func (i Interpreter) visitExpressionStmt(stmt Expression) (any, error) {
	i.evaluate(stmt.Expression)
	return nil, nil
}

func (i Interpreter) VisitGroupingExpr(expr Grouping) (any, error) {
	return i.evaluate(expr.Expression)
}

func (i Interpreter) VisitLiteralExpr(expr Literal) (any, error) {
	return expr.Value, nil
}

func (i Interpreter) visitPrintingStmt(stmt Print) (any, error) {
	value, evalErr := i.evaluate(stmt.Expression)
	if evalErr != nil {
		return nil, evalErr
	}
	printResult(value)
	return nil, nil
}

func (i Interpreter) VisitUnaryExpr(expr Unary) (any, error) {
	right, evalErr := i.evaluate(expr.Right)
	if evalErr != nil {
		return nil, evalErr
	}
	switch expr.Operator.TokenType {
	case token.MINUS:
		switch right := right.(type) {
		case int64:
			return -right, nil
		case float64:
			return -right, nil
		}
		return math.NaN(), nil
	case token.BANG:
		return !i.isTruthy(right), nil
	}

	return nil, nil
}
