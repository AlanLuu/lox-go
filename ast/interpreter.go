package ast

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/AlanLuu/lox/env"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type Interpreter struct {
	environment *env.Environment
	globals     *env.Environment
}

func NewInterpreter() *Interpreter {
	interpreter := &Interpreter{
		globals: env.NewEnvironment(),
	}
	interpreter.environment = interpreter.globals
	interpreter.defineNativeFuncs()
	return interpreter
}

func (i *Interpreter) defineNativeFuncs() {
	nativeFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string { return "<native fn>" }
		i.globals.Define(name, s)
	}
	nativeFunc("clock", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return float64(time.Now().UnixMilli()) / 1000, nil
	})
}

func (i *Interpreter) evaluate(expr any) (any, error) {
	switch expr := expr.(type) {
	case Assign:
		return i.visitAssignExpr(expr)
	case Block:
		return i.visitBlockStmt(expr)
	case Break:
		return expr, errors.New("")
	case Call:
		return i.visitCallExpr(expr)
	case Continue:
		return expr, errors.New("")
	case Expression:
		return i.visitExpressionStmt(expr)
	case For:
		return i.visitForStmt(expr)
	case Function:
		return i.visitFunctionStmt(expr)
	case If:
		return i.visitIfStmt(expr)
	case Print:
		return i.visitPrintingStmt(expr)
	case Var:
		return i.visitVarStmt(expr)
	case Variable:
		return i.visitVariableStmt(expr)
	case While:
		return i.visitWhileStmt(expr)
	case Binary:
		return i.visitBinaryExpr(expr)
	case Grouping:
		return i.visitGroupingExpr(expr)
	case Literal:
		return i.visitLiteralExpr(expr)
	case Logical:
		return i.visitLogicalExpr(expr)
	case Unary:
		return i.visitUnaryExpr(expr)
	}
	return nil, errors.New("critical error: unknown type found in AST")
}

func (i *Interpreter) Interpret(statements list.List[Stmt]) error {
	for _, statement := range statements {
		value, evalErr := i.evaluate(statement)
		if evalErr != nil {
			if value != nil {
				switch statement.(type) {
				case While, For:
					continue
				}
			}
			return evalErr
		}
	}
	return nil
}

func (i *Interpreter) isTruthy(obj any) bool {
	switch obj := obj.(type) {
	case nil:
		return false
	case bool:
		return obj
	case int64:
		return obj != 0
	case float64:
		return obj != 0.0 && !math.IsNaN(obj)
	case string:
		return len(obj) > 0
	}
	return true
}

func printResult(source any, isPrintStmt bool) {
	switch source := source.(type) {
	case nil:
		if isPrintStmt {
			fmt.Println("nil")
		}
	case float64:
		if math.IsInf(source, 1) {
			fmt.Println("Infinity")
		} else if math.IsInf(source, -1) {
			fmt.Println("-Infinity")
		} else {
			fmt.Println(source)
		}
	case string:
		if len(source) == 0 && !isPrintStmt {
			fmt.Println("\"\"")
		} else {
			fmt.Println(source)
		}
	default:
		fmt.Println(source)
	}
}

func printResultExpressionStmt(source any) {
	printResult(source, false)
}

func printResultPrintStmt(source any) {
	printResult(source, true)
}

func (i *Interpreter) visitAssignExpr(expr Assign) (any, error) {
	value, valueErr := i.evaluate(expr.Value)
	if valueErr != nil {
		return nil, valueErr
	}
	assignErr := i.environment.Assign(expr.Name, value)
	if assignErr != nil {
		return nil, assignErr
	}
	return value, nil
}

func (i *Interpreter) visitBinaryExpr(expr Binary) (any, error) {
	floatIsInt := func(f float64) bool {
		return f == float64(int64(f))
	}
	runtimeErrorWrapper := func(message string) error {
		return loxerror.RuntimeError(expr.Operator, message)
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

func (i *Interpreter) visitCallExpr(expr Call) (any, error) {
	callee, calleeErr := i.evaluate(expr.Callee)
	if calleeErr != nil {
		return nil, calleeErr
	}
	arguments := list.NewList[any]()
	for _, argument := range expr.Arguments {
		result, resultErr := i.evaluate(argument)
		if resultErr != nil {
			arguments.Clear()
			return nil, resultErr
		}
		arguments.Add(result)
	}
	if function, ok := callee.(LoxCallable); ok {
		argsLen := len(arguments)
		arity := function.arity()
		if argsLen != arity {
			return nil, loxerror.RuntimeError(expr.Paren,
				fmt.Sprintf("Expected %v arguments but got %v.", arity, argsLen),
			)
		}
		return function.call(i, arguments)
	}
	return nil, loxerror.RuntimeError(expr.Paren, "Can only call functions and classes.")
}

func (i *Interpreter) executeBlock(statements list.List[Stmt], environment *env.Environment) (any, error) {
	previous := i.environment
	i.environment = environment
	defer func() {
		i.environment = previous
	}()
	for _, statement := range statements {
		value, evalErr := i.evaluate(statement)
		if evalErr != nil {
			if value != nil {
				switch statement.(type) {
				case While, For:
					continue
				}
			}
			switch value := value.(type) {
			case Break, Continue:
				return value, evalErr
			}
			return nil, evalErr
		}
	}
	return nil, nil
}

func (i *Interpreter) visitBlockStmt(stmt Block) (any, error) {
	value, blockErr := i.executeBlock(stmt.Statements, env.NewEnvironmentEnclosing(i.environment))
	if blockErr != nil {
		switch value := value.(type) {
		case Break, Continue:
			return value, blockErr
		}
		return nil, blockErr
	}
	return nil, nil
}

func (i *Interpreter) visitExpressionStmt(stmt Expression) (any, error) {
	value, err := i.evaluate(stmt.Expression)
	if err != nil {
		return nil, err
	}
	if util.StdinFromTerminal() {
		_, ok := stmt.Expression.(Assign)
		if !ok {
			printResultExpressionStmt(value)
		}
	}
	return nil, nil
}

func (i *Interpreter) visitForStmt(stmt For) (any, error) {
	loopInterrupted := false
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		sig := <-sigChan
		switch sig {
		case os.Interrupt:
			loopInterrupted = true
		}
	}()

	tempEnvironment := env.NewEnvironmentEnclosing(i.environment)
	previous := i.environment
	i.environment = tempEnvironment
	defer func() {
		i.environment = previous
	}()

	if stmt.Initializer != nil {
		_, initializerErr := i.evaluate(stmt.Initializer)
		if initializerErr != nil {
			return nil, initializerErr
		}
	}
	if stmt.Condition != nil {
		for result, conditionErr := i.evaluate(stmt.Condition); conditionErr != nil || i.isTruthy(result); {
			if conditionErr != nil {
				return nil, conditionErr
			}
			if loopInterrupted {
				return nil, loxerror.RuntimeError(stmt.ForToken, "loop interrupted")
			}
			value, evalErr := i.evaluate(stmt.Body)
			if evalErr != nil {
				switch value := value.(type) {
				case Break:
					return value, evalErr
				case Continue:
				default:
					return nil, evalErr
				}
			}
			if stmt.Increment != nil {
				_, incrementErr := i.evaluate(stmt.Increment)
				if incrementErr != nil {
					return nil, incrementErr
				}
			}
			result, conditionErr = i.evaluate(stmt.Condition)
		}
	} else {
		for {
			if loopInterrupted {
				return nil, loxerror.RuntimeError(stmt.ForToken, "loop interrupted")
			}
			value, evalErr := i.evaluate(stmt.Body)
			if evalErr != nil {
				switch value := value.(type) {
				case Break:
					return value, evalErr
				case Continue:
				default:
					return nil, evalErr
				}
			}
			if stmt.Increment != nil {
				_, incrementErr := i.evaluate(stmt.Increment)
				if incrementErr != nil {
					return nil, incrementErr
				}
			}
		}
	}
	return nil, nil
}

func (i *Interpreter) visitFunctionStmt(stmt Function) (any, error) {
	function := LoxFunction{stmt}
	i.environment.Define(stmt.Name.Lexeme, function)
	return nil, nil
}

func (i *Interpreter) visitGroupingExpr(expr Grouping) (any, error) {
	return i.evaluate(expr.Expression)
}

func (i *Interpreter) visitIfStmt(stmt If) (any, error) {
	condition, conditionErr := i.evaluate(stmt.Condition)
	if conditionErr != nil {
		return nil, conditionErr
	}
	if i.isTruthy(condition) {
		value, evalErr := i.evaluate(stmt.ThenBranch)
		if evalErr != nil {
			switch value := value.(type) {
			case Break, Continue:
				return value, evalErr
			}
			return nil, evalErr
		}
	} else if stmt.ElseBranch != nil {
		value, evalErr := i.evaluate(stmt.ElseBranch)
		if evalErr != nil {
			switch value := value.(type) {
			case Break, Continue:
				return value, evalErr
			}
			return nil, evalErr
		}
	}
	return nil, nil
}

func (i *Interpreter) visitLiteralExpr(expr Literal) (any, error) {
	return expr.Value, nil
}

func (i *Interpreter) visitLogicalExpr(expr Logical) (any, error) {
	left, leftErr := i.evaluate(expr.Left)
	if leftErr != nil {
		return nil, leftErr
	}
	if expr.Operator.TokenType == token.OR {
		if i.isTruthy(left) {
			return left, nil
		}
	} else if !i.isTruthy(left) {
		return left, nil
	}
	return i.evaluate(expr.Right)
}

func (i *Interpreter) visitPrintingStmt(stmt Print) (any, error) {
	value, evalErr := i.evaluate(stmt.Expression)
	if evalErr != nil {
		return nil, evalErr
	}
	printResultPrintStmt(value)
	return nil, nil
}

func (i *Interpreter) visitUnaryExpr(expr Unary) (any, error) {
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

func (i *Interpreter) visitVarStmt(stmt Var) (any, error) {
	var value any
	var err error
	if stmt.Initializer != nil {
		value, err = i.evaluate(stmt.Initializer)
		if err != nil {
			return nil, err
		}
	}
	i.environment.Define(stmt.Name.Lexeme, value)
	return nil, nil
}

func (i *Interpreter) visitVariableStmt(expr Variable) (any, error) {
	return i.environment.Get(expr.Name)
}

func (i *Interpreter) visitWhileStmt(stmt While) (any, error) {
	loopInterrupted := false
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		sig := <-sigChan
		switch sig {
		case os.Interrupt:
			loopInterrupted = true
		}
	}()
	for result, conditionErr := i.evaluate(stmt.Condition); conditionErr != nil || i.isTruthy(result); {
		if conditionErr != nil {
			return nil, conditionErr
		}
		if loopInterrupted {
			return nil, loxerror.RuntimeError(stmt.WhileToken, "loop interrupted")
		}
		value, evalErr := i.evaluate(stmt.Body)
		if evalErr != nil {
			switch value := value.(type) {
			case Break:
				return value, evalErr
			case Continue:
			default:
				return nil, evalErr
			}
		}
		result, conditionErr = i.evaluate(stmt.Condition)
	}
	return nil, nil
}
