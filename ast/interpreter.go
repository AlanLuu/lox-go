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
	nativeFunc("len", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		switch element := args[0].(type) {
		case string:
			return int64(len(element)), nil
		case list.List[Expr]:
			return int64(len(element)), nil
		}
		return nil, loxerror.Error(fmt.Sprintf("Cannot get length of type '%v'.", getType(args[0])))
	})
	nativeFunc("List", 1, func(_ *Interpreter, args list.List[any]) (any, error) {
		if size, ok := args[0].(int64); ok {
			if size < 0 {
				return nil, loxerror.Error("Argument to List function cannot be negative.")
			}
			lst := list.NewList[Expr]()
			for index := int64(0); index < size; index++ {
				lst.Add(nil)
			}
			return lst, nil
		}
		return nil, loxerror.Error("Expected argument of type 'number' to List function.")
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
	case string:
		return "string"
	case LoxClass:
		return "class"
	case LoxFunction:
		return "function"
	case *LoxInstance:
		return element.String()
	case list.List[Expr]:
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
	case string:
		return len(obj) > 0
	}
	return true
}

func getResult(source any, isPrintStmt bool, alwaysPrint bool) string {
	switch source := source.(type) {
	case nil:
		if isPrintStmt || alwaysPrint {
			return "nil"
		} else {
			return ""
		}
	case float64:
		if math.IsInf(source, 1) {
			return "Infinity"
		} else if math.IsInf(source, -1) {
			return "-Infinity"
		} else {
			return fmt.Sprint(source)
		}
	case string:
		if len(source) == 0 && (!isPrintStmt || alwaysPrint) {
			return "\"\""
		} else {
			return fmt.Sprint(source)
		}
	case list.List[Expr]:
		sourceLen := len(source)
		var listStr strings.Builder
		listStr.WriteByte('[')
		for i, element := range source {
			listStr.WriteString(getResult(element, isPrintStmt, true))
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
	fmt.Println(getResult(source, false, false))
}

func printResultPrintStmt(source any) {
	fmt.Println(getResult(source, true, false))
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

	if leftAsStringer, ok := left.(fmt.Stringer); ok {
		left = leftAsStringer.String()
	}
	if rightAsStringer, ok := right.(fmt.Stringer); ok {
		right = rightAsStringer.String()
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
		case string:
			return handleNumString(left, right)
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
		case string:
			switch expr.Operator.TokenType {
			case token.PLUS:
				return strconv.FormatBool(left) + right, nil
			case token.STAR:
				return handleNumString(boolMap[left], right)
			}
		case nil:
			return handleTwoFloats(boolMap[left], 0)
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
			case nil:
				return left + "nil", nil
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
			case nil:
				return "", nil
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
		case string:
			switch expr.Operator.TokenType {
			case token.PLUS:
				return "nil" + right, nil
			case token.STAR:
				return "", nil
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
		if argsLen != arity {
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

	methods := make(map[string]LoxFunction)
	for _, method := range stmt.Methods {
		isInit := method.Name.Lexeme == "init"
		function := LoxFunction{method.Name.Lexeme, method.Function, i.environment, isInit}
		methods[method.Name.Lexeme] = function
	}

	loxClass := LoxClass{stmt.Name.Lexeme, superClass, methods}
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
		if !isAssign && !isSet {
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

func (i *Interpreter) visitFunctionExpr(expr FunctionExpr) (LoxFunction, error) {
	return LoxFunction{"", expr, i.environment, false}, nil
}

func (i *Interpreter) visitFunctionStmt(stmt Function) (any, error) {
	funcName := stmt.Name.Lexeme
	i.environment.Define(funcName, LoxFunction{funcName, stmt.Function, i.environment, false})
	return nil, nil
}

func (i *Interpreter) visitGetExpr(expr Get) (any, error) {
	obj, objErr := i.evaluate(expr.Object)
	if objErr != nil {
		return nil, objErr
	}
	if obj, ok := obj.(*LoxInstance); ok {
		return obj.Get(expr.Name)
	}
	return nil, loxerror.RuntimeError(expr.Name, "Only instances have properties.")
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
	if _, ok := indexVal.(int64); !ok {
		return nil, loxerror.RuntimeError(expr.Bracket, "Index value is not an integer.")
	}
	indexValInt := indexVal.(int64)
	switch indexElement := indexElement.(type) {
	case string:
		if indexValInt < 0 || indexValInt >= int64(len(indexElement)) {
			return nil, loxerror.RuntimeError(expr.Bracket, "String index out of range.")
		}
		return string(indexElement[indexValInt]), nil
	case list.List[Expr]:
		if indexValInt < 0 || indexValInt >= int64(len(indexElement)) {
			return nil, loxerror.RuntimeError(expr.Bracket, "List index out of range.")
		}
		return indexElement[indexValInt], nil
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
	return expr.Elements, nil
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
	if obj, ok := obj.(*LoxInstance); ok {
		value, valueErr := i.evaluate(expr.Value)
		if valueErr != nil {
			return nil, valueErr
		}
		obj.Set(expr.Name, value)
		return value, nil
	}
	return nil, loxerror.RuntimeError(expr.Name, "Only instances have fields.")
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
			return nil, loxerror.RuntimeError(expr.Name, "Index value is not an integer.")
		}
		node = index.IndexElement
		index, ok = node.(Index)
	}

	variable, variableErr := i.evaluate(node)
	if variableErr != nil {
		return nil, variableErr
	}
	if variable, ok := variable.(list.List[Expr]); ok {
		value, valueErr := i.evaluate(expr.Value)
		if valueErr != nil {
			return nil, valueErr
		}
		for loopIndex := len(indexes) - 1; loopIndex >= 0; loopIndex-- {
			position := indexes[loopIndex]
			if loopIndex > 0 {
				if position < 0 || position >= int64(len(variable)) {
					return nil, loxerror.RuntimeError(expr.Name, "List index out of range.")
				}
				variable, ok = variable[position].(list.List[Expr])
				if !ok {
					return nil, loxerror.RuntimeError(expr.Name, "Can only assign to list indexes.")
				}
			} else {
				if position < 0 || position >= int64(len(variable)) {
					return nil, loxerror.RuntimeError(expr.Name, "List index out of range.")
				}
				variable[position] = value
			}
		}
		return value, nil
	}
	return nil, loxerror.RuntimeError(expr.Name, "Can only assign to list indexes.")
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
