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
	"unicode/utf8"

	"github.com/AlanLuu/lox/env"
	"github.com/AlanLuu/lox/equatable"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type Interpreter struct {
	environment *env.Environment
	globals     *env.Environment
	locals      map[any]int
	blockDepth  int
}

func NewInterpreter() *Interpreter {
	interpreter := &Interpreter{
		globals:    env.NewEnvironment(),
		locals:     make(map[any]int),
		blockDepth: 0,
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
	nativeFunc("chr", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if codePointNum, ok := args[0].(int64); ok {
			codePoint := rune(codePointNum)
			character := string(codePoint)
			if codePoint == '\'' {
				return &LoxString{character, '"'}, nil
			}
			return &LoxString{character, '\''}, nil
		}
		return nil, loxerror.Error("Argument to 'chr' must be a whole number.")
	})
	nativeFunc("len", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		switch element := args[0].(type) {
		case *LoxString:
			return int64(utf8.RuneCountInString(element.str)), nil
		case *LoxList:
			return int64(len(element.elements)), nil
		}
		return nil, loxerror.Error(fmt.Sprintf("Cannot get length of type '%v'.", getType(args[0])))
	})
	nativeFunc("List", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if size, ok := args[0].(int64); ok {
			if size < 0 {
				return nil, loxerror.Error("Argument to 'List' cannot be negative.")
			}
			lst := list.NewList[Expr]()
			for index := int64(0); index < size; index++ {
				lst.Add(nil)
			}
			return &LoxList{lst}, nil
		}
		return nil, loxerror.Error("Argument to 'List' must be a whole number.")
	})
	nativeFunc("ord", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			if utf8.RuneCountInString(loxStr.str) == 1 {
				codePoint, _ := utf8.DecodeRuneInString(loxStr.str)
				if codePoint == utf8.RuneError {
					return nil, loxerror.Error(fmt.Sprintf("Failed to decode character '%v'.", loxStr.str))
				}
				return int64(codePoint), nil
			}
		}
		return nil, loxerror.Error("Argument to 'ord' must be a single character.")
	})
	nativeFunc("type", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		return getType(args[0]), nil
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
		result, resultErr := i.visitCallExpr(expr)
		if resultErr != nil {
			switch result := result.(type) {
			case Return:
				return result.FinalValue, nil
			}
			return nil, resultErr
		}
		return result, nil
	case Class:
		return i.visitClassStmt(expr)
	case Continue:
		return expr, errors.New("")
	case Expression:
		return i.visitExpressionStmt(expr)
	case For:
		return i.visitForStmt(expr)
	case Function:
		return i.visitFunctionStmt(expr)
	case FunctionExpr:
		return i.visitFunctionExpr(expr)
	case Get:
		return i.visitGetExpr(expr)
	case If:
		return i.visitIfStmt(expr)
	case Index:
		return i.visitIndexExpr(expr)
	case List:
		return i.visitListExpr(expr)
	case Print:
		return i.visitPrintingStmt(expr)
	case Return:
		return i.visitReturnStmt(expr)
	case Set:
		return i.visitSetExpr(expr)
	case SetList:
		return i.visitSetListExpr(expr)
	case String:
		return i.visitStringExpr(expr)
	case Super:
		return i.visitSuperExpr(expr)
	case This:
		return i.visitThisExpr(expr)
	case Var:
		return i.visitVarStmt(expr)
	case Variable:
		return i.visitVariableExpr(expr)
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
	case nil:
		return nil, nil
	}
	return nil, errors.New("critical error: unknown type found in AST")
}

func getType(element any) string {
	switch element := element.(type) {
	case nil:
		return "nil"
	case int64, float64:
		return "number"
	case bool:
		return "boolean"
	case *LoxString:
		return "string"
	case LoxClass:
		return "class"
	case *LoxFunction:
		return "function"
	case *LoxInstance:
		return element.String()
	case *LoxList:
		return "list"
	default:
		return "unknown type"
	}
}

func (i *Interpreter) Interpret(statements list.List[Stmt]) error {
	for _, statement := range statements {
		value, evalErr := i.evaluate(statement)
		if evalErr != nil {
			if value != nil {
				switch statement.(type) {
				case While, For, Call:
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
	case *LoxString:
		return len(obj.str) > 0
	}
	return true
}

func getResult(source any, isPrintStmt bool) string {
	switch source := source.(type) {
	case nil:
		return "nil"
	case float64:
		if math.IsInf(source, 1) {
			return "Infinity"
		} else if math.IsInf(source, -1) {
			return "-Infinity"
		} else {
			return fmt.Sprint(source)
		}
	case *LoxString:
		if len(source.str) == 0 {
			if isPrintStmt {
				return ""
			} else {
				return fmt.Sprintf("%c%c", source.quote, source.quote)
			}
		} else {
			if isPrintStmt {
				return fmt.Sprint(source.str)
			} else {
				return fmt.Sprintf("%c%v%c", source.quote, source.str, source.quote)
			}
		}
	case *LoxList:
		sourceLen := len(source.elements)
		var listStr strings.Builder
		listStr.WriteByte('[')
		for i, element := range source.elements {
			if element == source {
				listStr.WriteString("[...]")
			} else {
				listStr.WriteString(getResult(element, false))
			}
			if i < sourceLen-1 {
				listStr.WriteString(", ")
			}
		}
		listStr.WriteByte(']')
		return listStr.String()
	default:
		return fmt.Sprint(source)
	}
}

func printResultExpressionStmt(source any) {
	if source != nil {
		fmt.Println(getResult(source, false))
	}
}

func printResultPrintStmt(source any) {
	fmt.Println(getResult(source, true))
}

func (i *Interpreter) Resolve(expr Expr, depth int) {
	switch expr := expr.(type) {
	case Assign:
		i.locals[expr.Name] = depth
	case Variable:
		i.locals[expr.Name] = depth
	default:
		i.locals[expr] = depth
	}
}

func (i *Interpreter) visitAssignExpr(expr Assign) (any, error) {
	value, valueErr := i.evaluate(expr.Value)
	if valueErr != nil {
		return nil, valueErr
	}
	distance, ok := i.locals[expr.Name]
	if ok {
		i.environment.AssignAt(distance, expr.Name, value)
	} else {
		assignErr := i.globals.Assign(expr.Name, value)
		if assignErr != nil {
			return nil, assignErr
		}
	}
	return value, nil
}

func (i *Interpreter) visitBinaryExpr(expr Binary) (any, error) {
	runtimeErrorWrapper := func(message string) error {
		return loxerror.RuntimeError(expr.Operator, message)
	}
	unknownOpStr := "unknown operator"
	unknownOp := func() error {
		return runtimeErrorWrapper(fmt.Sprintf("%v '%v'.", unknownOpStr, expr.Operator.Lexeme))
	}
	handleNumString := func(left float64, right *LoxString) (any, error) {
		switch expr.Operator.TokenType {
		case token.PLUS:
			if math.IsInf(left, 1) {
				return right.NewLoxString("Infinity" + right.str), nil
			} else if math.IsInf(left, -1) {
				return right.NewLoxString("-Infinity" + right.str), nil
			}
			return right.NewLoxString(strconv.FormatFloat(left, 'f', -1, 64) + right.str), nil
		case token.STAR:
			if left <= 0 {
				return EmptyLoxString(), nil
			}
			if util.FloatIsInt(left) {
				return right.NewLoxString(strings.Repeat(right.str, int(left))), nil
			}
		}
		return math.NaN(), nil
	}
	handleNumList := func(left float64, right *LoxList) (any, error) {
		switch expr.Operator.TokenType {
		case token.PLUS:
			return handleNumString(left, &LoxString{right.String(), '\''})
		case token.STAR:
			if left <= 0 || len(right.elements) == 0 {
				return EmptyLoxList(), nil
			}
			if util.FloatIsInt(left) {
				newList := list.NewList[Expr]()
				for i := 0; i < int(left); i++ {
					for _, element := range right.elements {
						newList.Add(element)
					}
				}
				return &LoxList{newList}, nil
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
		case token.PERCENT:
			result = math.Mod(left, right)
		case token.DOUBLE_STAR:
			result = math.Pow(left, right)
		case token.DOUBLE_LESS:
			result = int64(left) << int64(right)
		case token.LESS:
			result = left < right
		case token.LESS_EQUAL:
			result = left <= right
		case token.DOUBLE_GREATER:
			result = int64(left) >> int64(right)
		case token.GREATER:
			result = left > right
		case token.GREATER_EQUAL:
			result = left >= right
		case token.AND_SYMBOL:
			result = int64(left) & int64(right)
		case token.OR_SYMBOL:
			result = int64(left) | int64(right)
		case token.CARET:
			result = int64(left) ^ int64(right)
		default:
			return nil, unknownOp()
		}

		switch result := result.(type) {
		case float64:
			if !math.IsInf(result, 1) &&
				!math.IsInf(result, -1) &&
				util.FloatIsInt(left) &&
				util.FloatIsInt(right) &&
				util.FloatIsInt(result) {
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
		leftEquatable, leftIsEquatable := left.(equatable.Equatable)
		if leftIsEquatable {
			return leftEquatable.Equals(right), nil
		}
		switch left := left.(type) {
		case int64:
			switch right := right.(type) {
			case float64:
				return float64(left) == right, nil
			}
		case float64:
			switch right := right.(type) {
			case int64:
				return left == float64(right), nil
			}
		}
		return left == right, nil
	}
	if expr.Operator.TokenType == token.BANG_EQUAL {
		leftEquatable, leftIsEquatable := left.(equatable.Equatable)
		if leftIsEquatable {
			return !leftEquatable.Equals(right), nil
		}
		switch left := left.(type) {
		case int64:
			switch right := right.(type) {
			case float64:
				return float64(left) != right, nil
			}
		case float64:
			switch right := right.(type) {
			case int64:
				return left != float64(right), nil
			}
		}
		return left != right, nil
	}

	if leftAsStringer, ok := left.(fmt.Stringer); ok {
		switch left.(type) {
		case *LoxString:
		case *LoxList:
		default:
			left = &LoxString{leftAsStringer.String(), '\''}
		}
	}
	if rightAsStringer, ok := right.(fmt.Stringer); ok {
		switch right.(type) {
		case *LoxString:
		case *LoxList:
		default:
			right = &LoxString{rightAsStringer.String(), '\''}
		}
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
		case *LoxString:
			return handleNumString(float64(left), right)
		case *LoxList:
			return handleNumList(float64(left), right)
		case nil:
			return handleTwoFloats(float64(left), 0)
		}
	case float64:
		switch right := right.(type) {
		case int64:
			return handleTwoFloats(float64(left), float64(right))
		case float64:
			return handleTwoFloats(float64(left), float64(right))
		case bool:
			return handleTwoFloats(float64(left), boolMap[right])
		case *LoxString:
			return handleNumString(left, right)
		case *LoxList:
			return handleNumList(left, right)
		case nil:
			return handleTwoFloats(float64(left), 0)
		}
	case bool:
		switch right := right.(type) {
		case int64:
			return handleTwoFloats(boolMap[left], float64(right))
		case float64:
			return handleTwoFloats(boolMap[left], float64(right))
		case bool:
			return handleTwoFloats(boolMap[left], boolMap[right])
		case *LoxString:
			switch expr.Operator.TokenType {
			case token.PLUS:
				return right.NewLoxString(strconv.FormatBool(left) + right.str), nil
			case token.STAR:
				return handleNumString(boolMap[left], right)
			}
		case *LoxList:
			switch expr.Operator.TokenType {
			case token.PLUS:
				return &LoxString{strconv.FormatBool(left) + right.String(), '\''}, nil
			case token.STAR:
				return handleNumList(boolMap[left], right)
			}
		case nil:
			return handleTwoFloats(boolMap[left], 0)
		}
	case *LoxString:
		switch expr.Operator.TokenType {
		case token.PLUS:
			switch right := right.(type) {
			case int64:
				return left.NewLoxString(left.str + strconv.FormatFloat(float64(right), 'f', -1, 64)), nil
			case float64:
				if math.IsInf(right, 1) {
					return left.NewLoxString("Infinity" + left.str), nil
				} else if math.IsInf(right, -1) {
					return left.NewLoxString("-Infinity" + left.str), nil
				}
				return left.NewLoxString(left.str + strconv.FormatFloat(right, 'f', -1, 64)), nil
			case bool:
				return left.NewLoxString(left.str + strconv.FormatBool(right)), nil
			case *LoxString:
				if left.quote == '"' || right.quote == '"' {
					return &LoxString{left.str + right.str, '"'}, nil
				}
				return &LoxString{left.str + right.str, '\''}, nil
			case *LoxList:
				return left.NewLoxString(left.str + right.String()), nil
			case nil:
				return left.NewLoxString(left.str + "nil"), nil
			}
		case token.STAR:
			repeat := func(left *LoxString, right int64) (*LoxString, error) {
				if right <= 0 {
					return EmptyLoxString(), nil
				}
				return left.NewLoxString(strings.Repeat(left.str, int(right))), nil
			}
			switch right := right.(type) {
			case int64:
				return repeat(left, right)
			case bool:
				return repeat(left, int64(boolMap[right]))
			case nil:
				return EmptyLoxString(), nil
			}
		}
	case *LoxList:
		switch expr.Operator.TokenType {
		case token.PLUS:
			switch right := right.(type) {
			case int64:
				return &LoxString{left.String() + strconv.FormatFloat(float64(right), 'f', -1, 64), '\''}, nil
			case float64:
				if math.IsInf(right, 1) {
					return &LoxString{left.String() + "Infinity", '\''}, nil
				} else if math.IsInf(right, -1) {
					return &LoxString{left.String() + "-Infinity", '\''}, nil
				}
				return &LoxString{left.String() + strconv.FormatFloat(right, 'f', -1, 64), '\''}, nil
			case bool:
				return &LoxString{left.String() + strconv.FormatBool(right), '\''}, nil
			case *LoxString:
				return right.NewLoxString(left.String() + right.str), nil
			case *LoxList:
				newList := list.NewList[Expr]()
				for _, element := range left.elements {
					newList.Add(element)
				}
				for _, element := range right.elements {
					newList.Add(element)
				}
				return &LoxList{newList}, nil
			case nil:
				return &LoxString{left.String() + "nil", '\''}, nil
			}
		case token.STAR:
			repeat := func(left *LoxList, right int64) (*LoxList, error) {
				if right <= 0 || len(left.elements) == 0 {
					return EmptyLoxList(), nil
				}
				newList := list.NewList[Expr]()
				for i := int64(0); i < right; i++ {
					for _, element := range left.elements {
						newList.Add(element)
					}
				}
				return &LoxList{newList}, nil
			}
			switch right := right.(type) {
			case int64:
				return repeat(left, right)
			case bool:
				return repeat(left, int64(boolMap[right]))
			case nil:
				return EmptyLoxList(), nil
			}
		}
	case nil:
		switch right := right.(type) {
		case int64:
			return handleTwoFloats(0, float64(right))
		case float64:
			return handleTwoFloats(0, float64(right))
		case bool:
			return handleTwoFloats(0, boolMap[right])
		case *LoxString:
			switch expr.Operator.TokenType {
			case token.PLUS:
				return right.NewLoxString("nil" + right.str), nil
			case token.STAR:
				return EmptyLoxString(), nil
			}
		case *LoxList:
			switch expr.Operator.TokenType {
			case token.PLUS:
				return &LoxString{"nil" + right.String(), '\''}, nil
			case token.STAR:
				return EmptyLoxList(), nil
			}
		case nil:
			return handleTwoFloats(0, 0)
		}
	}

	return math.NaN(), nil
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
		if arity >= 0 && argsLen != arity {
			return nil, loxerror.RuntimeError(expr.Paren,
				fmt.Sprintf("Expected %v arguments but got %v.", arity, argsLen),
			)
		}
		return function.call(i, arguments)
	}
	return nil, loxerror.RuntimeError(expr.Paren, "Can only call functions and classes.")
}

func (i *Interpreter) visitClassStmt(stmt Class) (any, error) {
	var superClass *LoxClass
	if stmt.SuperClass != nil {
		evalObj, evalObjErr := i.evaluate(*stmt.SuperClass)
		if evalObjErr != nil {
			return nil, evalObjErr
		}
		switch evalObj := evalObj.(type) {
		case LoxClass:
			superClass = &evalObj
		default:
			return nil, loxerror.RuntimeError(stmt.SuperClass.Name, "Superclass must be a class.")
		}
	}
	i.environment.Define(stmt.Name.Lexeme, nil)
	if stmt.SuperClass != nil {
		environment := env.NewEnvironmentEnclosing(i.environment)
		environment.Define("super", superClass)
		previous := i.environment
		i.environment = environment
		defer func() {
			i.environment = previous
		}()
	}

	methods := make(map[string]*LoxFunction)
	for _, method := range stmt.Methods {
		isInit := method.Name.Lexeme == "init"
		function := &LoxFunction{method.Name.Lexeme, method.Function, i.environment, isInit}
		methods[method.Name.Lexeme] = function
	}

	classMethods := make(map[string]any)
	for _, method := range stmt.ClassMethods {
		function := &LoxFunction{method.Name.Lexeme, method.Function, i.environment, false}
		classMethods[method.Name.Lexeme] = function
	}

	loxClass := LoxClass{stmt.Name.Lexeme, superClass, methods, classMethods}
	i.environment.Assign(stmt.Name, loxClass)
	return nil, nil
}

func (i *Interpreter) executeBlock(statements list.List[Stmt], environment *env.Environment) (any, error) {
	i.blockDepth++
	previous := i.environment
	i.environment = environment
	defer func() {
		i.environment = previous
		i.blockDepth--
	}()
	for _, statement := range statements {
		value, evalErr := i.evaluate(statement)
		if evalErr != nil {
			if value != nil {
				switch statement.(type) {
				case While, For:
					if _, ok := value.(Return); !ok {
						continue
					}
				}
			}
			switch value := value.(type) {
			case Break, Continue, Return:
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
		case Break, Continue, Return:
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
	if util.StdinFromTerminal() && i.blockDepth <= 0 {
		_, isAssign := stmt.Expression.(Assign)
		_, isSet := stmt.Expression.(Set)
		_, isSetList := stmt.Expression.(SetList)
		if !isAssign && !isSet && !isSetList {
			printResultExpressionStmt(value)
		}
	}
	return nil, nil
}

func (i *Interpreter) visitForStmt(stmt For) (any, error) {
	enteredLoop := false
	loopInterrupted := false
	catchInterruptSignal := func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		go func() {
			sig := <-sigChan
			switch sig {
			case os.Interrupt:
				loopInterrupted = true
			}
		}()
	}

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
			if !enteredLoop {
				catchInterruptSignal()
				enteredLoop = true
			}
			value, evalErr := i.evaluate(stmt.Body)
			if evalErr != nil {
				switch value := value.(type) {
				case Break, Return:
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
			if !enteredLoop {
				catchInterruptSignal()
				enteredLoop = true
			}
			value, evalErr := i.evaluate(stmt.Body)
			if evalErr != nil {
				switch value := value.(type) {
				case Break, Return:
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

func (i *Interpreter) visitFunctionExpr(expr FunctionExpr) (*LoxFunction, error) {
	return &LoxFunction{"", expr, i.environment, false}, nil
}

func (i *Interpreter) visitFunctionStmt(stmt Function) (any, error) {
	funcName := stmt.Name.Lexeme
	i.environment.Define(funcName, &LoxFunction{funcName, stmt.Function, i.environment, false})
	return nil, nil
}

func (i *Interpreter) visitGetExpr(expr Get) (any, error) {
	obj, objErr := i.evaluate(expr.Object)
	if objErr != nil {
		return nil, objErr
	}
	if obj, ok := obj.(LoxObject); ok {
		return obj.Get(expr.Name)
	}
	return nil, loxerror.RuntimeError(expr.Name, "Only classes, instances, lists, and strings have properties.")
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
			case Break, Continue, Return:
				return value, evalErr
			}
			return nil, evalErr
		}
	} else if stmt.ElseBranch != nil {
		value, evalErr := i.evaluate(stmt.ElseBranch)
		if evalErr != nil {
			switch value := value.(type) {
			case Break, Continue, Return:
				return value, evalErr
			}
			return nil, evalErr
		}
	}
	return nil, nil
}

func (i *Interpreter) visitIndexExpr(expr Index) (any, error) {
	indexElement, indexElementErr := i.evaluate(expr.IndexElement)
	if indexElementErr != nil {
		return nil, indexElementErr
	}

	indexVal, indexValErr := i.evaluate(expr.Index)
	if indexValErr != nil {
		return nil, indexValErr
	}
	if indexVal == nil {
		indexVal = int64(0)
	}

	indexEndVal, indexEndValErr := i.evaluate(expr.IndexEnd)
	if indexEndValErr != nil {
		return nil, indexEndValErr
	}

	switch indexElement := indexElement.(type) {
	case *LoxString:
		if _, ok := indexVal.(int64); !ok {
			return nil, loxerror.RuntimeError(expr.Bracket, StringIndexMustBeWholeNum(indexVal))
		}
		indexValInt := indexVal.(int64)
		if expr.IsSlice {
			if indexEndVal == nil {
				indexEndVal = int64(len(indexElement.str))
			}
			if _, ok := indexEndVal.(int64); !ok {
				return nil, loxerror.RuntimeError(expr.Bracket, StringIndexMustBeWholeNum(indexEndVal))
			}
			indexEndValInt := indexEndVal.(int64)
			if indexEndValInt > int64(len(indexElement.str)) {
				indexEndValInt = int64(len(indexElement.str))
			}
			if indexValInt < 0 {
				return nil, loxerror.RuntimeError(expr.Bracket, StringIndexOutOfRange(indexValInt))
			}
			if indexValInt > indexEndValInt {
				return EmptyLoxString(), nil
			}
			return &LoxString{indexElement.str[indexValInt:indexEndValInt], '\''}, nil
		} else {
			if indexValInt < 0 || indexValInt >= int64(len(indexElement.str)) {
				return nil, loxerror.RuntimeError(expr.Bracket, StringIndexOutOfRange(indexValInt))
			}
			str := string(indexElement.str[indexValInt])
			if str == "'" {
				return &LoxString{str, '"'}, nil
			}
			return &LoxString{str, '\''}, nil
		}
	case *LoxList:
		if _, ok := indexVal.(int64); !ok {
			return nil, loxerror.RuntimeError(expr.Bracket, ListIndexMustBeWholeNum(indexVal))
		}
		indexValInt := indexVal.(int64)
		if expr.IsSlice {
			if indexEndVal == nil {
				indexEndVal = int64(len(indexElement.elements))
			}
			if _, ok := indexEndVal.(int64); !ok {
				return nil, loxerror.RuntimeError(expr.Bracket, ListIndexMustBeWholeNum(indexEndVal))
			}
			indexEndValInt := indexEndVal.(int64)
			if indexEndValInt > int64(len(indexElement.elements)) {
				indexEndValInt = int64(len(indexElement.elements))
			}
			if indexValInt < 0 {
				return nil, loxerror.RuntimeError(expr.Bracket, ListIndexOutOfRange(indexValInt))
			}
			listSlice := list.NewList[Expr]()
			for i := indexValInt; i < indexEndValInt; i++ {
				listSlice.Add(indexElement.elements[i])
			}
			return &LoxList{listSlice}, nil
		} else {
			if indexValInt < 0 || indexValInt >= int64(len(indexElement.elements)) {
				return nil, loxerror.RuntimeError(expr.Bracket, ListIndexOutOfRange(indexValInt))
			}
			return indexElement.elements[indexValInt], nil
		}
	}
	return nil, loxerror.RuntimeError(expr.Bracket, "Can only index into lists and strings.")
}

func (i *Interpreter) visitListExpr(expr List) (any, error) {
	for index, element := range expr.Elements {
		evalResult, evalErr := i.evaluate(element)
		if evalErr != nil {
			return nil, evalErr
		}
		expr.Elements[index] = evalResult
	}
	return &LoxList{expr.Elements}, nil
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

func (i *Interpreter) visitReturnStmt(stmt Return) (any, error) {
	var value any
	var valueErr error
	if stmt.Value != nil {
		value, valueErr = i.evaluate(stmt.Value)
		if valueErr != nil {
			return nil, valueErr
		}
	}
	stmt.FinalValue = value
	return stmt, errors.New("")
}

func (i *Interpreter) visitSetExpr(expr Set) (any, error) {
	obj, objErr := i.evaluate(expr.Object)
	if objErr != nil {
		return nil, objErr
	}
	evaluateAndSet := func(setFunc func(value any)) (any, error) {
		value, valueErr := i.evaluate(expr.Value)
		if valueErr != nil {
			return nil, valueErr
		}
		setFunc(value)
		return value, nil
	}
	if instance, ok := obj.(*LoxInstance); ok {
		return evaluateAndSet(func(value any) {
			instance.Set(expr.Name, value)
		})
	}
	if class, ok := obj.(LoxClass); ok {
		return evaluateAndSet(func(value any) {
			class.classProperties[expr.Name.Lexeme] = value
		})
	}
	return nil, loxerror.RuntimeError(expr.Name, "Only classes and instances have properties that can be set.")
}

func (i *Interpreter) visitSetListExpr(expr SetList) (any, error) {
	indexes := list.NewList[int64]()
	defer indexes.Clear()

	node := expr.Object
	for index, ok := node.(Index); ok; {
		indexNum, indexNumErr := i.evaluate(index.Index)
		if indexNumErr != nil {
			return nil, indexNumErr
		}
		switch indexNum := indexNum.(type) {
		case int64:
			indexes.Add(indexNum)
		default:
			return nil, loxerror.RuntimeError(expr.Name, ListIndexMustBeWholeNum(indexNum))
		}
		node = index.IndexElement
		index, ok = node.(Index)
	}

	variable, variableErr := i.evaluate(node)
	if variableErr != nil {
		return nil, variableErr
	}
	if variable, ok := variable.(*LoxList); ok {
		value, valueErr := i.evaluate(expr.Value)
		if valueErr != nil {
			return nil, valueErr
		}
		for loopIndex := len(indexes) - 1; loopIndex >= 0; loopIndex-- {
			position := indexes[loopIndex]
			if loopIndex > 0 {
				if position < 0 || position >= int64(len(variable.elements)) {
					return nil, loxerror.RuntimeError(expr.Name, ListIndexOutOfRange(position))
				}
				variable, ok = variable.elements[position].(*LoxList)
				if !ok {
					return nil, loxerror.RuntimeError(expr.Name, "Can only assign to list indexes.")
				}
			} else {
				if position < 0 || position >= int64(len(variable.elements)) {
					return nil, loxerror.RuntimeError(expr.Name, ListIndexOutOfRange(position))
				}
				variable.elements[position] = value
			}
		}
		return value, nil
	}
	return nil, loxerror.RuntimeError(expr.Name, "Can only assign to list indexes.")
}

func (i *Interpreter) visitStringExpr(expr String) (any, error) {
	return &LoxString{expr.Str, expr.Quote}, nil
}

func (i *Interpreter) visitSuperExpr(expr Super) (any, error) {
	distance := i.locals[expr]
	superClass := i.environment.GetAtStr(distance, "super").(*LoxClass)
	object := i.environment.GetAtStr(distance-1, "this").(*LoxInstance)
	method, ok := superClass.findMethod(expr.Method.Lexeme)
	if !ok {
		return nil, loxerror.RuntimeError(expr.Method, "Undefined property '"+expr.Method.Lexeme+"'.")
	}
	return method.bind(object), nil
}

func (i *Interpreter) visitThisExpr(expr This) (any, error) {
	distance, ok := i.locals[expr]
	if ok {
		return i.environment.GetAt(distance, expr.Keyword)
	} else {
		return i.globals.Get(expr.Keyword)
	}
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
	case token.TILDE:
		switch right := right.(type) {
		case int64:
			return ^right, nil
		case float64:
			return ^int64(right), nil
		}
		return math.NaN(), nil
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

func (i *Interpreter) visitVariableExpr(expr Variable) (any, error) {
	distance, ok := i.locals[expr.Name]
	if ok {
		return i.environment.GetAt(distance, expr.Name)
	} else {
		return i.globals.Get(expr.Name)
	}
}

func (i *Interpreter) visitWhileStmt(stmt While) (any, error) {
	enteredLoop := false
	loopInterrupted := false
	for result, conditionErr := i.evaluate(stmt.Condition); conditionErr != nil || i.isTruthy(result); {
		if conditionErr != nil {
			return nil, conditionErr
		}
		if loopInterrupted {
			return nil, loxerror.RuntimeError(stmt.WhileToken, "loop interrupted")
		}
		if !enteredLoop {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)
			go func() {
				sig := <-sigChan
				switch sig {
				case os.Interrupt:
					loopInterrupted = true
				}
			}()
			enteredLoop = true
		}
		value, evalErr := i.evaluate(stmt.Body)
		if evalErr != nil {
			switch value := value.(type) {
			case Break, Return:
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
