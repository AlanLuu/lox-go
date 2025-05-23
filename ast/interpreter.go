package ast

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/AlanLuu/lox/bignum/bigfloat"
	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/env"
	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/loxsignal"
	"github.com/AlanLuu/lox/scanner"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type Interpreter struct {
	environment *env.Environment
	globals     *env.Environment
	locals      map[any]int
	blockDepth  int
	callToken   *token.Token
}

func NewInterpreter() *Interpreter {
	interpreter := &Interpreter{
		globals:    env.NewEnvironment(),
		locals:     make(map[any]int),
		blockDepth: 0,
		callToken:  nil,
	}
	interpreter.environment = interpreter.globals
	interpreter.defineBase32Funcs()     //Defined in base32funcs.go
	interpreter.defineBase64Funcs()     //Defined in base64funcs.go
	interpreter.defineBigFloatFuncs()   //Defined in bigfloatfuncs.go
	interpreter.defineBigIntFuncs()     //Defined in bigintfuncs.go
	interpreter.defineBigMathFuncs()    //Defined in bigmathfuncs.go
	interpreter.defineClassCalledLox()  //Defined in classcalledlox.go
	interpreter.defineCryptoFuncs()     //Defined in cryptofuncs.go
	interpreter.defineCSVFuncs()        //Defined in csvfuncs.go
	interpreter.defineDateFuncs()       //Defined in datefuncs.go
	interpreter.defineDotenvFuncs()     //Defined in dotenvfuncs.go
	interpreter.defineDurationFuncs()   //Defined in durationfuncs.go
	interpreter.defineFloatFuncs()      //Defined in floatfuncs.go
	interpreter.defineGzipFuncs()       //Defined in gzipfuncs.go
	interpreter.defineHexFuncs()        //Defined in hexfuncs.go
	interpreter.defineHTMLFuncs()       //Defined in htmlfuncs.go
	interpreter.defineHTTPFuncs()       //Defined in httpfuncs.go
	interpreter.defineIntFuncs()        //Defined in intfuncs.go
	interpreter.defineIteratorFuncs()   //Defined in iteratorfuncs.go
	interpreter.defineJSONFuncs()       //Defined in jsonfuncs.go
	interpreter.defineLogFuncs()        //Defined in logfuncs.go
	interpreter.defineMathFuncs()       //Defined in mathfuncs.go
	interpreter.defineNativeFuncs()     //Defined in nativefuncs.go
	interpreter.defineNetFuncs()        //Defined in netfuncs.go
	interpreter.defineOSFuncs()         //Defined in osfuncs.go
	interpreter.defineProcessFuncs()    //Defined in processfuncs.go
	interpreter.defineRandFuncs()       //Defined in randfuncs.go
	interpreter.defineRegexFuncs()      //Defined in regexfuncs.go
	interpreter.defineStringFuncs()     //Defined in stringfuncs.go
	interpreter.defineTarFuncs()        //Defined in tarfuncs.go
	interpreter.defineUnsafeFuncs()     //Defined in unsafefuncs.go
	interpreter.defineUUIDFuncs()       //Defined in uuidfuncs.go
	interpreter.defineWebBrowserFuncs() //Defined in webbrowserfuncs.go
	interpreter.defineWindowsFuncs()    //Defined in windowsfuncs_windows.go
	interpreter.defineZipFuncs()        //Defined in zipfuncs.go
	return interpreter
}

func (i *Interpreter) evaluate(expr any) (any, error) {
	switch expr := expr.(type) {
	case Assert:
		return i.visitAssertStmt(expr)
	case Assign:
		return i.visitAssignExpr(expr)
	case BigNum:
		return i.visitBigNumExpr(expr)
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
	case Dict:
		return i.visitDictExpr(expr)
	case DoWhile:
		return i.visitDoWhileStmt(expr)
	case Enum:
		return i.visitEnumStmt(expr)
	case Expression:
		return i.visitExpressionStmt(expr)
	case For:
		return i.visitForStmt(expr)
	case ForEach:
		return i.visitForEachStmt(expr)
	case Function:
		return i.visitFunctionStmt(expr)
	case FunctionExpr:
		return i.visitFunctionExpr(expr)
	case Get:
		return i.visitGetExpr(expr)
	case If:
		return i.visitIfStmt(expr)
	case Import:
		return i.visitImportStmt(expr)
	case Index:
		return i.visitIndexExpr(expr)
	case List:
		return i.visitListExpr(expr)
	case Loop:
		return i.visitLoopStmt(expr)
	case Print:
		return i.visitPrintingStmt(expr)
	case Repeat:
		return i.visitRepeatStmt(expr)
	case Return:
		return i.visitReturnStmt(expr)
	case Set:
		return i.visitSetExpr(expr)
	case SetObject:
		return i.visitSetObjectExpr(expr)
	case String:
		return i.visitStringExpr(expr)
	case Super:
		return i.visitSuperExpr(expr)
	case Ternary:
		return i.visitTernaryExpr(expr)
	case This:
		return i.visitThisExpr(expr)
	case Throw:
		return i.visitThrowStmt(expr)
	case TryCatchFinally:
		return i.visitTryCatchFinallyStmt(expr)
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
	case int64:
		return "integer"
	case float64:
		return "float"
	case *big.Int:
		return "bigint"
	case *big.Float:
		return "bigfloat"
	case bool:
		return "boolean"
	case interfaces.Type:
		return element.Type()
	default:
		return "unknown"
	}
}

func (i *Interpreter) Interpret(statements list.List[Stmt], makeHandler bool) error {
	interrupted := false
	if util.StdinFromTerminal() && makeHandler {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		defer func() {
			if !interrupted {
				sigChan <- loxsignal.LoopSignal{}
				signal.Stop(sigChan)
			}
		}()
		go func() {
			sig := <-sigChan
			switch sig {
			case os.Interrupt:
				interrupted = true
				signal.Stop(sigChan)
			}
		}()
	}
	for _, statement := range statements {
		if interrupted {
			return nil
		}
		value, evalErr := i.evaluate(statement)
		if evalErr != nil {
			if value != nil {
				switch statement.(type) {
				case While, For, ForEach, DoWhile, Repeat, Loop, Call:
					continue
				}
			}
			return evalErr
		}
	}
	return nil
}

func (i *Interpreter) InterpretReturnLast(statements list.List[Stmt]) (any, error) {
	var lastValue any
	for _, statement := range statements {
		var value any
		var evalErr error
		switch statement := statement.(type) {
		case Expression:
			value, evalErr = i.visitExpressionStmtReturn(statement)
		default:
			value, evalErr = i.evaluate(statement)
		}
		lastValue = value
		if evalErr != nil {
			if value != nil {
				switch statement.(type) {
				case While, For, ForEach, DoWhile, Repeat, Loop, Call:
					continue
				}
			}
			return nil, evalErr
		}
	}
	return lastValue, nil
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
	case *big.Int:
		return !bigint.IsZero(obj)
	case *big.Float:
		return !bigfloat.IsZero(obj)
	case interfaces.Length:
		return obj.Length() > 0
	}
	return true
}

func selfReferential(source any) string {
	switch source.(type) {
	case *LoxDict, *LoxSet:
		return "{...}"
	case *LoxList:
		return "[...]"
	}
	return "..."
}

func getBufferElementHexResult(source any) string {
	switch source := source.(type) {
	case int64:
		return fmt.Sprintf("0x%x", source)
	default:
		return fmt.Sprint(source)
	}
}

func getResult(source any, originalSource any, isPrintStmt bool) string {
	switch source := source.(type) {
	case nil:
		return "nil"
	case *big.Int:
		return bigint.String(source)
	case *big.Float:
		return bigfloat.String(source)
	case float64:
		switch {
		case math.IsInf(source, 1):
			return "Infinity"
		case math.IsInf(source, -1):
			return "-Infinity"
		case util.FloatIsInt(source):
			return fmt.Sprintf("%.1f", source)
		default:
			return util.FormatFloat(source)
		}
	case *LoxString:
		if len(source.str) == 0 {
			if isPrintStmt {
				return ""
			} else {
				return fmt.Sprintf("%c%c", source.quote, source.quote)
			}
		} else {
			sourceStr := source.str
			_, originalIsString := originalSource.(*LoxString)
			if !originalIsString || !isPrintStmt {
				escapeChars := map[rune]string{
					'\a': "\\a",
					'\n': "\\n",
					'\r': "\\r",
					'\t': "\\t",
					'\b': "\\b",
					'\f': "\\f",
					'\v': "\\v",
				}
				sourceQuoteRune := rune(source.quote)
				var builder strings.Builder
				for _, c := range source.str {
					if escapeString, ok := escapeChars[c]; ok {
						builder.WriteString(escapeString)
					} else {
						if c == sourceQuoteRune {
							builder.WriteRune('\\')
						}
						builder.WriteRune(c)
					}
				}
				sourceStr = builder.String()
			}
			if isPrintStmt {
				return fmt.Sprint(sourceStr)
			} else {
				return fmt.Sprintf("%c%v%c", source.quote, sourceStr, source.quote)
			}
		}
	case LoxStringStr:
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
	case *LoxBuffer:
		sourceLen := len(source.elements)
		var bufferStr strings.Builder
		bufferStr.WriteString("Buffer [")
		for i, element := range source.elements {
			if element == originalSource {
				bufferStr.WriteString(selfReferential(originalSource))
			} else {
				bufferStr.WriteString(getBufferElementHexResult(element))
			}
			if i < sourceLen-1 {
				bufferStr.WriteString(", ")
			}
		}
		bufferStr.WriteByte(']')
		return bufferStr.String()
	case *LoxDict:
		sourceLen := len(source.entries)
		var dictStr strings.Builder
		dictStr.WriteByte('{')
		i := 0
		for key, value := range source.entries {
			if key == originalSource {
				dictStr.WriteString(selfReferential(originalSource))
			} else {
				dictStr.WriteString(getResult(key, originalSource, false))
			}
			dictStr.WriteString(": ")
			if value == originalSource {
				dictStr.WriteString(selfReferential(originalSource))
			} else {
				dictStr.WriteString(getResult(value, originalSource, false))
			}
			if i < sourceLen-1 {
				dictStr.WriteString(", ")
			}
			i++
		}
		dictStr.WriteByte('}')
		return dictStr.String()
	case *LoxList:
		sourceLen := len(source.elements)
		var listStr strings.Builder
		listStr.WriteByte('[')
		for i, element := range source.elements {
			if element == originalSource {
				listStr.WriteString(selfReferential(originalSource))
			} else {
				listStr.WriteString(getResult(element, originalSource, false))
			}
			if i < sourceLen-1 {
				listStr.WriteString(", ")
			}
		}
		listStr.WriteByte(']')
		return listStr.String()
	case *LoxQueue:
		sourceLen := source.elements.Len()
		var queueStr strings.Builder
		queueStr.WriteString("Queue [")
		i := 0
		for e := source.elements.Front(); e != nil; e = e.Next() {
			if e.Value == originalSource {
				queueStr.WriteString(selfReferential(originalSource))
			} else {
				queueStr.WriteString(getResult(e.Value, originalSource, false))
			}
			if i < sourceLen-1 {
				queueStr.WriteString(", ")
			}
			i++
		}
		queueStr.WriteByte(']')
		return queueStr.String()
	case *LoxDeque:
		sourceLen := source.elements.Len()
		var dequeStr strings.Builder
		dequeStr.WriteString("Deque [")
		i := 0
		for e := source.elements.Front(); e != nil; e = e.Next() {
			if e.Value == originalSource {
				dequeStr.WriteString(selfReferential(originalSource))
			} else {
				dequeStr.WriteString(getResult(e.Value, originalSource, false))
			}
			if i < sourceLen-1 {
				dequeStr.WriteString(", ")
			}
			i++
		}
		dequeStr.WriteByte(']')
		return dequeStr.String()
	case *LoxSet:
		if len(source.elements) == 0 {
			return "∅"
		}
		sourceLen := len(source.elements)
		var setStr strings.Builder
		setStr.WriteByte('{')
		i := 0
		for element := range source.elements {
			if element == originalSource {
				setStr.WriteString(selfReferential(originalSource))
			} else {
				setStr.WriteString(getResult(element, originalSource, false))
			}
			if i < sourceLen-1 {
				setStr.WriteString(", ")
			}
			i++
		}
		setStr.WriteByte('}')
		return setStr.String()
	default:
		return fmt.Sprint(source)
	}
}

func printResultExpressionStmt(source any) {
	if source != nil {
		fmt.Println(getResult(source, source, false))
	}
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

func (i *Interpreter) visitAssertStmt(stmt Assert) (any, error) {
	assertValue, assertValueErr := i.evaluate(stmt.Value)
	if assertValueErr != nil {
		return nil, assertValueErr
	}
	if !i.isTruthy(assertValue) {
		return nil, loxerror.RuntimeError(stmt.AssertToken, "AssertionError")
	}
	return nil, nil
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

func (i *Interpreter) visitBigNumExpr(expr BigNum) (any, error) {
	if expr.IsFloat {
		bigFloat := &big.Float{}
		bigFloat.SetString(expr.NumStr)
		return bigFloat, nil
	} else {
		bigInt := &big.Int{}
		bigInt.SetString(expr.NumStr, 0)
		return bigInt, nil
	}
}

func (i *Interpreter) visitBinaryExpr(expr Binary) (any, error) {
	runtimeErrorWrapper := func(message string) error {
		return loxerror.RuntimeError(expr.Operator, message)
	}
	unknownOpStr := "unknown operator"
	unknownOp := func() error {
		return runtimeErrorWrapper(fmt.Sprintf("%v '%v'.", unknownOpStr, expr.Operator.Lexeme))
	}
	unknownOpOn := func(str string) error {
		return runtimeErrorWrapper(
			fmt.Sprintf("%v '%v' on %v.", unknownOpStr, expr.Operator.Lexeme, str),
		)
	}
	handleNumString := func(left float64, right *LoxString) (any, error) {
		switch expr.Operator.TokenType {
		case token.PLUS:
			if math.IsInf(left, 1) {
				return right.NewLoxString("Infinity" + right.str), nil
			} else if math.IsInf(left, -1) {
				return right.NewLoxString("-Infinity" + right.str), nil
			}
			return right.NewLoxString(util.FormatFloat(left) + right.str), nil
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
	handleBigNumString := func(left *big.Int, right *LoxString) (any, error) {
		switch expr.Operator.TokenType {
		case token.STAR:
			if bigint.IsZeroOrLess(left) || len(right.str) == 0 {
				return EmptyLoxString(), nil
			}
			var builder strings.Builder
			one := bigint.BoolMap[true]
			for i := big.NewInt(0); i.Cmp(left) < 0; i.Add(i, one) {
				builder.WriteString(right.str)
			}
			return right.NewLoxString(builder.String()), nil
		}
		return math.NaN(), nil
	}
	handleNumBuffer := func(left int64, right *LoxBuffer) (any, error) {
		switch expr.Operator.TokenType {
		case token.STAR:
			if left <= 0 || len(right.elements) == 0 {
				return EmptyLoxBuffer(), nil
			}
			newBuffer := EmptyLoxBufferCap(int64(len(right.elements)) * left)
			for i := int64(0); i < left; i++ {
				for _, element := range right.elements {
					addErr := newBuffer.add(element)
					if addErr != nil {
						return nil, runtimeErrorWrapper(addErr.Error())
					}
				}
			}
			return newBuffer, nil
		}
		return math.NaN(), nil
	}
	handleBigNumBuffer := func(left *big.Int, right *LoxBuffer) (any, error) {
		switch expr.Operator.TokenType {
		case token.STAR:
			if bigint.IsZeroOrLess(left) || len(right.elements) == 0 {
				return EmptyLoxBuffer(), nil
			}
			newBuffer := EmptyLoxBuffer()
			one := bigint.BoolMap[true]
			for i := big.NewInt(0); i.Cmp(left) < 0; i.Add(i, one) {
				for _, element := range right.elements {
					addErr := newBuffer.add(element)
					if addErr != nil {
						return nil, runtimeErrorWrapper(addErr.Error())
					}
				}
			}
			return newBuffer, nil
		}
		return math.NaN(), nil
	}
	handleNumList := func(left int64, right *LoxList) (any, error) {
		switch expr.Operator.TokenType {
		case token.STAR:
			if left <= 0 || len(right.elements) == 0 {
				return EmptyLoxList(), nil
			}
			newList := list.NewListCap[any](int64(len(right.elements)) * left)
			for i := int64(0); i < left; i++ {
				for _, element := range right.elements {
					newList.Add(element)
				}
			}
			return NewLoxList(newList), nil
		}
		return math.NaN(), nil
	}
	handleBigNumList := func(left *big.Int, right *LoxList) (any, error) {
		switch expr.Operator.TokenType {
		case token.STAR:
			if bigint.IsZeroOrLess(left) || len(right.elements) == 0 {
				return EmptyLoxList(), nil
			}
			newList := list.NewList[any]()
			one := bigint.BoolMap[true]
			for i := big.NewInt(0); i.Cmp(left) < 0; i.Add(i, one) {
				for _, element := range right.elements {
					newList.Add(element)
				}
			}
			return NewLoxList(newList), nil
		}
		return math.NaN(), nil
	}
	handleTwoBigInts := func(left *big.Int, right *big.Int) (any, error) {
		divideByZeroMsg := "Cannot divide bigint by 0."
		switch expr.Operator.TokenType {
		case token.PLUS:
			return new(big.Int).Add(left, right), nil
		case token.MINUS:
			return new(big.Int).Sub(left, right), nil
		case token.STAR:
			return new(big.Int).Mul(left, right), nil
		case token.SLASH:
			if bigint.IsZero(right) {
				return nil, runtimeErrorWrapper(divideByZeroMsg)
			}
			return new(big.Int).Div(left, right), nil
		case token.PERCENT:
			if bigint.IsZero(right) {
				return nil, runtimeErrorWrapper(divideByZeroMsg)
			}
			return new(big.Int).Rem(left, right), nil
		case token.DOUBLE_STAR:
			return new(big.Int).Exp(left, right, nil), nil
		case token.DOUBLE_LESS:
			if bigint.IsNegative(right) {
				return math.NaN(), nil
			}
			return new(big.Int).Lsh(left, uint(right.Uint64())), nil
		case token.LESS:
			return left.Cmp(right) < 0, nil
		case token.LESS_EQUAL:
			return left.Cmp(right) <= 0, nil
		case token.DOUBLE_GREATER:
			if bigint.IsNegative(right) {
				return math.NaN(), nil
			}
			return new(big.Int).Rsh(left, uint(right.Uint64())), nil
		case token.GREATER:
			return left.Cmp(right) > 0, nil
		case token.GREATER_EQUAL:
			return left.Cmp(right) >= 0, nil
		case token.AMPERSAND:
			return new(big.Int).And(left, right), nil
		case token.PIPE:
			return new(big.Int).Or(left, right), nil
		case token.CARET:
			return new(big.Int).Xor(left, right), nil
		default:
			return nil, unknownOpOn("bigints")
		}
	}
	handleTwoBigFloats := func(left *big.Float, right *big.Float) (any, error) {
		switch expr.Operator.TokenType {
		case token.PLUS:
			return new(big.Float).Add(left, right), nil
		case token.MINUS:
			return new(big.Float).Sub(left, right), nil
		case token.STAR:
			return new(big.Float).Mul(left, right), nil
		case token.SLASH:
			if (bigfloat.IsZero(left) && bigfloat.IsZero(right)) ||
				(left.IsInf() && right.IsInf()) {
				return math.NaN(), nil
			}
			result := new(big.Float).Quo(left, right)
			if result.IsInf() {
				if bigfloat.IsPositive(left) {
					return math.Inf(1), nil
				} else {
					return math.Inf(-1), nil
				}
			}
			return result, nil
		case token.PERCENT:
			if (bigfloat.IsZero(left) && bigfloat.IsZero(right)) ||
				(left.IsInf() && right.IsInf()) {
				return math.NaN(), nil
			}

			//mod(a, b) = a - (b * floor(a / b))
			quotient := new(big.Float).Quo(left, right)
			flooredQuotient := &big.Int{}
			quotient.Int(flooredQuotient)
			quotient.SetInt(flooredQuotient)
			quotient.Mul(quotient, right)
			return new(big.Float).Sub(left, quotient), nil
		case token.DOUBLE_LESS:
			leftInt := &big.Int{}
			rightInt := &big.Int{}
			left.Int(leftInt)
			right.Int(rightInt)
			if bigint.IsNegative(rightInt) {
				return math.NaN(), nil
			}
			return new(big.Int).Lsh(leftInt, uint(rightInt.Uint64())), nil
		case token.LESS:
			return left.Cmp(right) < 0, nil
		case token.LESS_EQUAL:
			return left.Cmp(right) <= 0, nil
		case token.DOUBLE_GREATER:
			leftInt := &big.Int{}
			rightInt := &big.Int{}
			left.Int(leftInt)
			right.Int(rightInt)
			if bigint.IsNegative(rightInt) {
				return math.NaN(), nil
			}
			return new(big.Int).Rsh(leftInt, uint(rightInt.Uint64())), nil
		case token.GREATER:
			return left.Cmp(right) > 0, nil
		case token.GREATER_EQUAL:
			return left.Cmp(right) >= 0, nil
		case token.AMPERSAND:
			leftInt := &big.Int{}
			rightInt := &big.Int{}
			left.Int(leftInt)
			right.Int(rightInt)
			return new(big.Int).And(leftInt, rightInt), nil
		case token.PIPE:
			leftInt := &big.Int{}
			rightInt := &big.Int{}
			left.Int(leftInt)
			right.Int(rightInt)
			return new(big.Int).Or(leftInt, rightInt), nil
		case token.CARET:
			leftInt := &big.Int{}
			rightInt := &big.Int{}
			left.Int(leftInt)
			right.Int(rightInt)
			return new(big.Int).Xor(leftInt, rightInt), nil
		default:
			return nil, unknownOpOn("bigfloats")
		}
	}
	handleTwoInts := func(left int64, right int64) (any, error) {
		var result any
		switch expr.Operator.TokenType {
		case token.PLUS:
			result = left + right
		case token.MINUS:
			result = left - right
		case token.STAR:
			result = left * right
		case token.SLASH:
			divResult := float64(left) / float64(right)
			if util.FloatIsInt(divResult) {
				result = int64(divResult)
			} else {
				result = divResult
			}
		case token.PERCENT:
			result = left % right
		case token.DOUBLE_STAR:
			result = int64(math.Pow(float64(left), float64(right)))
		case token.DOUBLE_LESS:
			if right >= 0 {
				result = left << right
			} else {
				result = math.NaN()
			}
		case token.LESS:
			result = left < right
		case token.LESS_EQUAL:
			result = left <= right
		case token.DOUBLE_GREATER:
			if right >= 0 {
				result = left >> right
			} else {
				result = math.NaN()
			}
		case token.GREATER:
			result = left > right
		case token.GREATER_EQUAL:
			result = left >= right
		case token.AMPERSAND:
			result = left & right
		case token.PIPE:
			result = left | right
		case token.CARET:
			result = left ^ right
		default:
			return nil, unknownOp()
		}
		return result, nil
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
			if int64(right) >= 0 {
				result = int64(left) << int64(right)
			} else {
				result = math.NaN()
			}
		case token.LESS:
			result = left < right
		case token.LESS_EQUAL:
			result = left <= right
		case token.DOUBLE_GREATER:
			if int64(right) >= 0 {
				result = int64(left) >> int64(right)
			} else {
				result = math.NaN()
			}
		case token.GREATER:
			result = left > right
		case token.GREATER_EQUAL:
			result = left >= right
		case token.AMPERSAND:
			result = int64(left) & int64(right)
		case token.PIPE:
			result = int64(left) | int64(right)
		case token.CARET:
			result = int64(left) ^ int64(right)
		default:
			return nil, unknownOp()
		}
		return result, nil
	}
	boolMap := map[bool]float64{
		true:  1,
		false: 0,
	}
	boolMapInt := map[bool]int64{
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
		leftEquatable, leftIsEquatable := left.(interfaces.Equatable)
		if leftIsEquatable {
			return leftEquatable.Equals(right), nil
		}
		switch left := left.(type) {
		case int64:
			switch right := right.(type) {
			case float64:
				return float64(left) == right, nil
			case *big.Int:
				return right.Cmp(big.NewInt(left)) == 0, nil
			case *big.Float:
				leftFloat := new(big.Float).SetInt(big.NewInt(left))
				return leftFloat.Cmp(right) == 0, nil
			}
		case float64:
			switch right := right.(type) {
			case int64:
				return left == float64(right), nil
			case float64:
				if math.IsNaN(left) && math.IsNaN(right) {
					return true, nil
				}
			case *big.Int:
				rightFloat := new(big.Float).SetInt(right)
				return rightFloat.Cmp(bigfloat.New(left)) == 0, nil
			case *big.Float:
				return right.Cmp(bigfloat.New(left)) == 0, nil
			}
		case *big.Int:
			switch right := right.(type) {
			case int64:
				return left.Cmp(big.NewInt(right)) == 0, nil
			case float64:
				leftFloat := new(big.Float).SetInt(left)
				return leftFloat.Cmp(bigfloat.New(right)) == 0, nil
			case *big.Int:
				return left.Cmp(right) == 0, nil
			case *big.Float:
				leftFloat := new(big.Float).SetInt(left)
				return leftFloat.Cmp(right) == 0, nil
			}
		case *big.Float:
			switch right := right.(type) {
			case int64:
				rightFloat := new(big.Float).SetInt(big.NewInt(right))
				return left.Cmp(rightFloat) == 0, nil
			case float64:
				return left.Cmp(bigfloat.New(right)) == 0, nil
			case *big.Int:
				rightFloat := new(big.Float).SetInt(right)
				return left.Cmp(rightFloat) == 0, nil
			case *big.Float:
				return left.Cmp(right) == 0, nil
			}
		}
		return left == right, nil
	}
	if expr.Operator.TokenType == token.BANG_EQUAL {
		leftEquatable, leftIsEquatable := left.(interfaces.Equatable)
		if leftIsEquatable {
			return !leftEquatable.Equals(right), nil
		}
		switch left := left.(type) {
		case int64:
			switch right := right.(type) {
			case float64:
				return float64(left) != right, nil
			case *big.Int:
				return right.Cmp(big.NewInt(left)) != 0, nil
			case *big.Float:
				leftFloat := new(big.Float).SetInt(big.NewInt(left))
				return leftFloat.Cmp(right) != 0, nil
			}
		case float64:
			switch right := right.(type) {
			case int64:
				return left != float64(right), nil
			case float64:
				if math.IsNaN(left) && math.IsNaN(right) {
					return false, nil
				}
			case *big.Int:
				rightFloat := new(big.Float).SetInt(right)
				return rightFloat.Cmp(bigfloat.New(left)) != 0, nil
			case *big.Float:
				return right.Cmp(bigfloat.New(left)) != 0, nil
			}
		case *big.Int:
			switch right := right.(type) {
			case int64:
				return left.Cmp(big.NewInt(right)) != 0, nil
			case float64:
				leftFloat := new(big.Float).SetInt(left)
				return leftFloat.Cmp(bigfloat.New(right)) != 0, nil
			case *big.Int:
				return left.Cmp(right) != 0, nil
			case *big.Float:
				leftFloat := new(big.Float).SetInt(left)
				return leftFloat.Cmp(right) != 0, nil
			}
		case *big.Float:
			switch right := right.(type) {
			case int64:
				rightFloat := new(big.Float).SetInt(big.NewInt(right))
				return left.Cmp(rightFloat) != 0, nil
			case float64:
				return left.Cmp(bigfloat.New(right)) != 0, nil
			case *big.Int:
				rightFloat := new(big.Float).SetInt(right)
				return left.Cmp(rightFloat) != 0, nil
			case *big.Float:
				return left.Cmp(right) != 0, nil
			}
		}
		return left != right, nil
	}

	if leftAsStringer, ok := left.(fmt.Stringer); ok {
		if _, ok := right.(*LoxString); ok && expr.Operator.TokenType == token.PLUS {
			left = NewLoxStringQuote(leftAsStringer.String())
		}
	}
	if rightAsStringer, ok := right.(fmt.Stringer); ok {
		if _, ok := left.(*LoxString); ok && expr.Operator.TokenType == token.PLUS {
			right = NewLoxStringQuote(rightAsStringer.String())
		}
	}
	switch left := left.(type) {
	case int64:
		switch right := right.(type) {
		case int64:
			return handleTwoInts(left, right)
		case float64:
			return handleTwoFloats(float64(left), right)
		case *big.Int:
			return handleTwoBigInts(big.NewInt(left), right)
		case *big.Float:
			return handleTwoBigFloats(bigfloat.New(float64(left)), right)
		case bool:
			return handleTwoInts(left, boolMapInt[right])
		case *LoxString:
			return handleNumString(float64(left), right)
		case *LoxBuffer:
			return handleNumBuffer(left, right)
		case *LoxList:
			return handleNumList(left, right)
		case nil:
			return handleTwoInts(left, 0)
		}
	case float64:
		switch right := right.(type) {
		case int64:
			return handleTwoFloats(left, float64(right))
		case float64:
			return handleTwoFloats(left, right)
		case *big.Int:
			return handleTwoBigFloats(bigfloat.New(left), new(big.Float).SetInt(right))
		case *big.Float:
			return handleTwoBigFloats(bigfloat.New(left), right)
		case bool:
			return handleTwoFloats(left, boolMap[right])
		case *LoxString:
			switch expr.Operator.TokenType {
			case token.PLUS:
				return handleNumString(left, right)
			}
		case nil:
			return handleTwoFloats(left, 0)
		}
	case *big.Int:
		switch right := right.(type) {
		case int64:
			return handleTwoBigInts(left, big.NewInt(right))
		case float64:
			return handleTwoBigFloats(new(big.Float).SetInt(left), bigfloat.New(right))
		case *big.Int:
			return handleTwoBigInts(left, right)
		case *big.Float:
			return handleTwoBigFloats(new(big.Float).SetInt(left), right)
		case bool:
			return handleTwoBigInts(left, bigint.BoolMap[right])
		case *LoxString:
			return handleBigNumString(left, right)
		case *LoxBuffer:
			return handleBigNumBuffer(left, right)
		case *LoxList:
			return handleBigNumList(left, right)
		case nil:
			return handleTwoBigInts(left, bigint.BoolMap[false])
		}
	case *big.Float:
		switch right := right.(type) {
		case int64:
			return handleTwoBigFloats(left, bigfloat.New(float64(right)))
		case float64:
			return handleTwoBigFloats(left, bigfloat.New(right))
		case *big.Int:
			return handleTwoBigFloats(left, new(big.Float).SetInt(right))
		case *big.Float:
			return handleTwoBigFloats(left, right)
		case bool:
			return handleTwoBigFloats(left, bigfloat.BoolMap[right])
		case nil:
			return handleTwoBigFloats(left, bigfloat.BoolMap[false])
		}
	case bool:
		switch right := right.(type) {
		case int64:
			return handleTwoInts(boolMapInt[left], right)
		case float64:
			return handleTwoFloats(boolMap[left], right)
		case *big.Int:
			return handleTwoBigInts(bigint.BoolMap[left], right)
		case *big.Float:
			return handleTwoBigFloats(bigfloat.BoolMap[left], right)
		case bool:
			return handleTwoInts(boolMapInt[left], boolMapInt[right])
		case *LoxString:
			switch expr.Operator.TokenType {
			case token.PLUS:
				return right.NewLoxString(strconv.FormatBool(left) + right.str), nil
			case token.STAR:
				return handleNumString(boolMap[left], right)
			}
		case *LoxBuffer:
			return handleNumBuffer(int64(boolMap[left]), right)
		case *LoxList:
			return handleNumList(int64(boolMap[left]), right)
		case nil:
			return handleTwoInts(boolMapInt[left], 0)
		}
	case *LoxString:
		switch expr.Operator.TokenType {
		case token.PLUS:
			switch right := right.(type) {
			case int64:
				return left.NewLoxString(left.str + util.FormatFloat(float64(right))), nil
			case float64:
				if math.IsInf(right, 1) {
					return left.NewLoxString("Infinity" + left.str), nil
				} else if math.IsInf(right, -1) {
					return left.NewLoxString("-Infinity" + left.str), nil
				}
				return left.NewLoxString(left.str + util.FormatFloat(right)), nil
			case bool:
				return left.NewLoxString(left.str + strconv.FormatBool(right)), nil
			case *LoxString:
				if left.quote == '"' || right.quote == '"' {
					return NewLoxString(left.str+right.str, '"'), nil
				}
				return NewLoxString(left.str+right.str, '\''), nil
			case *LoxBuffer:
				return left.NewLoxString(left.str + right.String()), nil
			case *LoxDict:
				return left.NewLoxString(left.str + right.String()), nil
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
			case *big.Int:
				return handleBigNumString(right, left)
			case bool:
				return repeat(left, int64(boolMap[right]))
			case nil:
				return EmptyLoxString(), nil
			}
		}
	case *LoxBuffer:
		switch expr.Operator.TokenType {
		case token.PLUS:
			switch right := right.(type) {
			case *LoxString:
				return right.NewLoxString(left.String() + right.str), nil
			case *LoxBuffer:
				capacity := int64(len(left.elements)) + int64(len(right.elements))
				newBuffer := EmptyLoxBufferCap(capacity)
				for _, element := range left.elements {
					addErr := newBuffer.add(element)
					if addErr != nil {
						return nil, runtimeErrorWrapper(addErr.Error())
					}
				}
				for _, element := range right.elements {
					addErr := newBuffer.add(element)
					if addErr != nil {
						return nil, runtimeErrorWrapper(addErr.Error())
					}
				}
				return newBuffer, nil
			}
		case token.STAR:
			repeat := func(left *LoxBuffer, right int64) (*LoxBuffer, error) {
				if right <= 0 || len(left.elements) == 0 {
					return EmptyLoxBuffer(), nil
				}
				newBuffer := EmptyLoxBufferCap(int64(len(left.elements)) * right)
				for i := int64(0); i < right; i++ {
					for _, element := range left.elements {
						addErr := newBuffer.add(element)
						if addErr != nil {
							return nil, runtimeErrorWrapper(addErr.Error())
						}
					}
				}
				return newBuffer, nil
			}
			switch right := right.(type) {
			case int64:
				return repeat(left, right)
			case *big.Int:
				return handleBigNumBuffer(right, left)
			case bool:
				return repeat(left, int64(boolMap[right]))
			case nil:
				return EmptyLoxBuffer(), nil
			}
		}
	case *LoxDict:
		switch expr.Operator.TokenType {
		case token.PLUS:
			switch right := right.(type) {
			case *LoxString:
				return right.NewLoxString(left.String() + right.str), nil
			}
		case token.PIPE:
			switch right := right.(type) {
			case *LoxDict:
				newDict := NewLoxDict(make(map[any]any))
				for key, value := range left.entries {
					newDict.setKeyValue(key, value)
				}
				for key, value := range right.entries {
					newDict.setKeyValue(key, value)
				}
				return newDict, nil
			}
		}
	case *LoxList:
		switch expr.Operator.TokenType {
		case token.PLUS:
			switch right := right.(type) {
			case *LoxString:
				return right.NewLoxString(left.String() + right.str), nil
			case *LoxList:
				capacity := int64(len(left.elements)) + int64(len(right.elements))
				newList := list.NewListCap[any](capacity)
				for _, element := range left.elements {
					newList.Add(element)
				}
				for _, element := range right.elements {
					newList.Add(element)
				}
				return NewLoxList(newList), nil
			}
		case token.STAR:
			repeat := func(left *LoxList, right int64) (*LoxList, error) {
				if right <= 0 || len(left.elements) == 0 {
					return EmptyLoxList(), nil
				}
				newList := list.NewListCap[any](int64(len(left.elements)) * right)
				for i := int64(0); i < right; i++ {
					for _, element := range left.elements {
						newList.Add(element)
					}
				}
				return NewLoxList(newList), nil
			}
			switch right := right.(type) {
			case int64:
				return repeat(left, right)
			case *big.Int:
				return handleBigNumList(right, left)
			case bool:
				return repeat(left, int64(boolMap[right]))
			case nil:
				return EmptyLoxList(), nil
			}
		}
	case *LoxSet:
		switch right := right.(type) {
		case *LoxSet:
			switch expr.Operator.TokenType {
			case token.PIPE:
				return left.union(right), nil
			case token.AMPERSAND:
				return left.intersection(right), nil
			case token.MINUS:
				return left.difference(right), nil
			case token.CARET:
				return left.symmetricDifference(right), nil
			case token.LESS:
				return left.isProperSubset(right), nil
			case token.LESS_EQUAL:
				return left.isSubset(right), nil
			case token.GREATER:
				return left.isProperSuperset(right), nil
			case token.GREATER_EQUAL:
				return left.isSuperset(right), nil
			}
		}
	case nil:
		switch right := right.(type) {
		case int64:
			return handleTwoInts(0, right)
		case float64:
			return handleTwoFloats(0, right)
		case *big.Int:
			return handleTwoBigInts(bigint.BoolMap[false], right)
		case *big.Float:
			return handleTwoBigFloats(bigfloat.BoolMap[false], right)
		case bool:
			return handleTwoInts(0, boolMapInt[right])
		case *LoxString:
			switch expr.Operator.TokenType {
			case token.PLUS:
				return right.NewLoxString("nil" + right.str), nil
			case token.STAR:
				return EmptyLoxString(), nil
			}
		case *LoxBuffer:
			switch expr.Operator.TokenType {
			case token.STAR:
				return EmptyLoxBuffer(), nil
			}
		case *LoxList:
			switch expr.Operator.TokenType {
			case token.STAR:
				return EmptyLoxList(), nil
			}
		case nil:
			return handleTwoInts(0, 0)
		}
	}

	return math.NaN(), nil
}

func (i *Interpreter) visitCallExpr(expr Call) (any, error) {
	callee, calleeErr := i.evaluate(expr.Callee)
	if calleeErr != nil {
		return nil, calleeErr
	}
	arguments := list.NewListCap[any](int64(len(expr.Arguments)))
	for _, argument := range expr.Arguments {
		switch argument := argument.(type) {
		case Spread:
			result, resultErr := i.evaluate(argument.Iterable)
			if resultErr != nil {
				arguments.Clear()
				return nil, resultErr
			}
			switch result := result.(type) {
			case interfaces.Iterable:
				it := result.Iterator()
				for it.HasNext() {
					arguments.Add(it.Next())
				}
			default:
				arguments.Clear()
				return nil, loxerror.RuntimeError(argument.SpreadToken,
					"Value after '...' must be an iterable.")
			}
		default:
			result, resultErr := i.evaluate(argument)
			if resultErr != nil {
				arguments.Clear()
				return nil, resultErr
			}
			arguments.Add(result)
		}
	}
	if function, ok := callee.(LoxCallable); ok {
		argsLen := len(arguments)
		arity := function.arity()
		if arity >= 0 && argsLen != arity {
			if arity == 1 {
				return nil, loxerror.RuntimeError(expr.Paren,
					fmt.Sprintf("Expected %v argument but got %v.", arity, argsLen),
				)
			}
			return nil, loxerror.RuntimeError(expr.Paren,
				fmt.Sprintf("Expected %v arguments but got %v.", arity, argsLen),
			)
		}
		switch function := function.(type) {
		case LoxBuiltInProtoCallable:
			arguments.AddAt(0, function.instance)
		}
		prevToken := i.callToken
		defer func() {
			i.callToken = prevToken
		}()
		i.callToken = expr.Paren
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
		case *LoxClass:
			superClass = evalObj
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
		function := &LoxFunction{method.Name.Lexeme, method.Function, i.environment, isInit, method.Function.VarArgPos}
		methods[method.Name.Lexeme] = function
	}

	classProperties := make(map[string]any)
	for _, method := range stmt.ClassMethods {
		function := &LoxFunction{method.Name.Lexeme, method.Function, i.environment, false, method.Function.VarArgPos}
		classProperties[method.Name.Lexeme] = function
	}
	for name, field := range stmt.ClassFields {
		value, valueErr := i.evaluate(field)
		if valueErr != nil {
			return nil, valueErr
		}
		classProperties[name] = value
	}

	instanceFields := make(map[string]any)
	for name, field := range stmt.InstanceFields {
		value, valueErr := i.evaluate(field)
		if valueErr != nil {
			return nil, valueErr
		}
		instanceFields[name] = value
	}

	loxClass := &LoxClass{
		stmt.Name.Lexeme,
		superClass,
		methods,
		make(map[string]*LoxFunction),
		classProperties,
		instanceFields,
		stmt.CanInstantiate,
		false,
	}
	i.environment.Assign(stmt.Name, loxClass)
	return nil, nil
}

func (i *Interpreter) visitDictExpr(expr Dict) (any, error) {
	dict := NewLoxDict(make(map[any]any))
	var tempKey any
	isKey := true
	for _, entry := range expr.Entries {
		switch entry := entry.(type) {
		case Spread:
			theEntry, entryErr := i.evaluate(entry.Iterable)
			if entryErr != nil {
				return nil, entryErr
			}
			switch theEntry := theEntry.(type) {
			case *LoxDict:
				it := theEntry.Iterator()
				for it.HasNext() {
					next := it.Next().(*LoxList).elements
					canBeKey, keyErr := CanBeDictKeyCheck(next[0])
					if !canBeKey {
						return nil, loxerror.RuntimeError(expr.DictToken, keyErr)
					}
					dict.setKeyValue(next[0], next[1])
				}
			case interfaces.Iterable:
				it := theEntry.Iterator()
				for index := int64(0); it.HasNext(); index++ {
					dict.setKeyValue(index, it.Next())
				}
			default:
				return nil, loxerror.RuntimeError(entry.SpreadToken,
					"Value after '...' must be an iterable.")
			}
			isKey = true
		default:
			theEntry, entryErr := i.evaluate(entry)
			if entryErr != nil {
				return nil, entryErr
			}
			if isKey {
				canBeKey, keyErr := CanBeDictKeyCheck(theEntry)
				if !canBeKey {
					return nil, loxerror.RuntimeError(expr.DictToken, keyErr)
				}
				tempKey = theEntry
			} else {
				dict.setKeyValue(tempKey, theEntry)
			}
			isKey = !isKey
		}
	}
	return dict, nil
}

func (i *Interpreter) visitDoWhileStmt(stmt DoWhile) (any, error) {
	firstIteration := true
	enteredLoop := false
	loopInterrupted := false
	var result any
	var conditionErr error
	for cond := true; cond; {
		if conditionErr != nil {
			return nil, conditionErr
		}
		if loopInterrupted {
			return nil, loxerror.RuntimeError(stmt.DoToken, "loop interrupted")
		}
		if !firstIteration && !enteredLoop {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)
			defer func() {
				if !loopInterrupted {
					sigChan <- loxsignal.LoopSignal{}
					signal.Stop(sigChan)
				}
			}()
			go func() {
				sig := <-sigChan
				switch sig {
				case os.Interrupt:
					loopInterrupted = true
					signal.Stop(sigChan)
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
		cond = conditionErr != nil || i.isTruthy(result)
		if firstIteration {
			firstIteration = false
		}
	}
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
				case While, For, ForEach, DoWhile, Repeat, Loop:
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
	return i.visitBlockStmtEnv(stmt, env.NewEnvironmentEnclosing(i.environment))
}

func (i *Interpreter) visitBlockStmtEnv(stmt Block, e *env.Environment) (any, error) {
	value, blockErr := i.executeBlock(stmt.Statements, e)
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
		_, isSetObject := stmt.Expression.(SetObject)
		if !isAssign && !isSet && !isSetObject {
			printResultExpressionStmt(value)
		}
	}
	return nil, nil
}

func (i *Interpreter) visitExpressionStmtReturn(stmt Expression) (any, error) {
	value, err := i.evaluate(stmt.Expression)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (i *Interpreter) visitEnumStmt(stmt Enum) (any, error) {
	enum := &LoxEnum{}
	enum.name = stmt.Name.Lexeme
	members := make(map[string]*LoxEnumMember)
	for _, memberToken := range stmt.Members {
		members[memberToken.Lexeme] = &LoxEnumMember{memberToken.Lexeme, enum}
	}
	enum.members = members
	i.environment.Define(stmt.Name.Lexeme, enum)
	return nil, nil
}

func (i *Interpreter) visitForStmt(stmt For) (any, error) {
	enteredLoop := false
	loopInterrupted := false

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
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt)
				defer func() {
					if !loopInterrupted {
						sigChan <- loxsignal.LoopSignal{}
						signal.Stop(sigChan)
					}
				}()
				go func() {
					sig := <-sigChan
					switch sig {
					case os.Interrupt:
						loopInterrupted = true
						signal.Stop(sigChan)
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
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt)
				defer func() {
					if !loopInterrupted {
						sigChan <- loxsignal.LoopSignal{}
						signal.Stop(sigChan)
					}
				}()
				go func() {
					sig := <-sigChan
					switch sig {
					case os.Interrupt:
						loopInterrupted = true
						signal.Stop(sigChan)
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

func (i *Interpreter) visitForEachStmt(stmt ForEach) (any, error) {
	inType, inTypeErr := i.evaluate(stmt.Iterable)
	if inTypeErr != nil {
		return nil, inTypeErr
	}
	if _, ok := inType.(interfaces.Iterable); !ok {
		return nil, loxerror.RuntimeError(stmt.ForEachToken,
			fmt.Sprintf("Type '%v' is not iterable.", getType(inType)))
	}
	iterator := inType.(interfaces.Iterable).Iterator()

	tempEnvironment := env.NewEnvironmentEnclosing(i.environment)
	isBlock := false
	switch stmt.Body.(type) {
	case Block:
		isBlock = true
	default:
		previous := i.environment
		i.environment = tempEnvironment
		defer func() {
			i.environment = previous
		}()
	}

	enteredLoop := false
	loopInterrupted := false
	for iterator.HasNext() {
		if loopInterrupted {
			return nil, loxerror.RuntimeError(stmt.ForEachToken, "loop interrupted")
		}
		if !enteredLoop {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)
			defer func() {
				if !loopInterrupted {
					sigChan <- loxsignal.LoopSignal{}
					signal.Stop(sigChan)
				}
			}()
			go func() {
				sig := <-sigChan
				switch sig {
				case os.Interrupt:
					loopInterrupted = true
					signal.Stop(sigChan)
				}
			}()
			enteredLoop = true
		}
		tempEnvironment.Define(stmt.VariableName.Lexeme, iterator.Next())
		var value any
		var evalErr error
		if isBlock {
			value, evalErr = i.visitBlockStmtEnv(stmt.Body.(Block), tempEnvironment)
		} else {
			value, evalErr = i.evaluate(stmt.Body)
		}
		if evalErr != nil {
			switch value := value.(type) {
			case Break, Return:
				return value, evalErr
			case Continue:
			default:
				return nil, evalErr
			}
		}
	}
	return nil, nil
}

func (i *Interpreter) visitFunctionExpr(expr FunctionExpr) (*LoxFunction, error) {
	return &LoxFunction{"", expr, i.environment, false, expr.VarArgPos}, nil
}

func (i *Interpreter) visitFunctionStmt(stmt Function) (any, error) {
	funcName := stmt.Name.Lexeme
	i.environment.Define(funcName, &LoxFunction{funcName, stmt.Function, i.environment, false, stmt.Function.VarArgPos})
	return nil, nil
}

func (i *Interpreter) visitGetExpr(expr Get) (any, error) {
	obj, objErr := i.evaluate(expr.Object)
	if objErr != nil {
		return nil, objErr
	}
	if obj, ok := obj.(LoxObject); ok {
		get, getErr := obj.Get(expr.Name)
		switch get := get.(type) {
		case interfaces.LazyType:
			evalErr := get.LazyTypeEval()
			if evalErr != nil {
				return nil, evalErr
			}
		}
		return get, getErr
	}
	return nil, loxerror.RuntimeError(expr.Name,
		fmt.Sprintf("Type '%v' does not have properties.", getType(obj)))
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

func (i *Interpreter) visitImportStmt(stmt Import) (any, error) {
	importFileObj, importFileErr := i.evaluate(stmt.ImportFile)
	if importFileErr != nil {
		return nil, importFileErr
	}

	if _, ok := importFileObj.(*LoxString); !ok {
		return nil, loxerror.RuntimeError(stmt.ImportToken,
			"Import file must be a string.")
	}

	importFilePath := importFileObj.(*LoxString).str
	importFile, openFileError := os.Open(importFilePath)
	if openFileError != nil {
		return nil, loxerror.RuntimeError(stmt.ImportToken,
			fmt.Sprintf("Could not find file '%v'.", importFilePath))
	}

	importErr := func(e error) (any, error) {
		return nil, loxerror.RuntimeError(stmt.ImportToken,
			fmt.Sprintf("Error when importing file '%v':\n%v",
				importFilePath, e.Error()))
	}

	importProgram, readErr := io.ReadAll(importFile)
	importFile.Close()
	if readErr != nil {
		return importErr(readErr)
	}
	importSc := scanner.NewScanner(string(importProgram))
	scanErr := importSc.ScanTokens()
	if scanErr != nil {
		return importErr(scanErr)
	}

	importParser := NewParser(importSc.Tokens)
	exprList, parseErr := importParser.Parse()
	defer exprList.Clear()
	if parseErr != nil {
		return importErr(parseErr)
	}

	previous := i.environment
	defer func() {
		i.environment = previous
	}()
	if len(stmt.ImportNamespace) > 0 {
		i.environment = env.NewEnvironment()
	} else {
		i.environment = i.globals
	}

	importResolver := NewResolver(i)
	resolverErr := importResolver.Resolve(exprList)
	if resolverErr != nil {
		return importErr(resolverErr)
	}

	valueErr := i.Interpret(exprList, false)
	if valueErr != nil {
		return importErr(valueErr)
	}

	if len(stmt.ImportNamespace) > 0 {
		nameSpaceClass := NewLoxClass(stmt.ImportNamespace, nil, false)
		values := i.environment.Values()
		for name, value := range values {
			nameSpaceClass.classProperties[name] = value
		}
		i.globals.Define(stmt.ImportNamespace, nameSpaceClass)
	}

	return true, nil
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

	indexEndVal, indexEndValErr := i.evaluate(expr.IndexEnd)
	if indexEndValErr != nil {
		return nil, indexEndValErr
	}

	invalidBigintErr := func(value *big.Int) (any, error) {
		return nil, loxerror.RuntimeError(expr.Bracket,
			fmt.Sprintf("bigint index value '%v' is out of range.", bigint.String(value)))
	}
	switch indexElement := indexElement.(type) {
	case *LoxString:
		if expr.IsSlice {
			if indexVal == nil {
				indexVal = int64(0)
			}
			if indexEndVal == nil {
				indexEndVal = int64(utf8.RuneCountInString(indexElement.str))
			}
			var indexValInt int64
			var indexEndValInt int64
			switch indexVal := indexVal.(type) {
			case int64:
				indexValInt = indexVal
			case *big.Int:
				if !indexVal.IsInt64() {
					return invalidBigintErr(indexVal)
				}
				indexValInt = indexVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, StringIndexMustBeWholeNum(indexVal))
			}
			switch indexEndVal := indexEndVal.(type) {
			case int64:
				indexEndValInt = indexEndVal
			case *big.Int:
				if !indexEndVal.IsInt64() {
					return invalidBigintErr(indexEndVal)
				}
				indexEndValInt = indexEndVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, StringIndexMustBeWholeNum(indexEndVal))
			}
			originalIndexValInt := indexValInt
			if indexValInt < 0 {
				indexValInt += int64(utf8.RuneCountInString(indexElement.str))
			}
			if indexEndValInt < 0 {
				indexEndValInt += int64(utf8.RuneCountInString(indexElement.str))
			}
			if indexEndValInt > int64(utf8.RuneCountInString(indexElement.str)) {
				indexEndValInt = int64(utf8.RuneCountInString(indexElement.str))
			}
			if indexValInt < 0 {
				return nil, loxerror.RuntimeError(expr.Bracket, StringIndexOutOfRange(originalIndexValInt))
			}
			if indexValInt > indexEndValInt {
				return EmptyLoxString(), nil
			}
			return NewLoxStringQuote(string([]rune(indexElement.str)[indexValInt:indexEndValInt])), nil
		} else {
			var indexValInt int64
			switch indexVal := indexVal.(type) {
			case int64:
				indexValInt = indexVal
			case *big.Int:
				if !indexVal.IsInt64() {
					return invalidBigintErr(indexVal)
				}
				indexValInt = indexVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, StringIndexMustBeWholeNum(indexVal))
			}
			originalIndexValInt := indexValInt
			if indexValInt < 0 {
				indexValInt += int64(utf8.RuneCountInString(indexElement.str))
			}
			if indexValInt < 0 || indexValInt >= int64(utf8.RuneCountInString(indexElement.str)) {
				return nil, loxerror.RuntimeError(expr.Bracket, StringIndexOutOfRange(originalIndexValInt))
			}
			str := string([]rune(indexElement.str)[indexValInt])
			if str == "'" {
				return NewLoxString(str, '"'), nil
			}
			return NewLoxString(str, '\''), nil
		}
	case *LoxBuffer:
		if expr.IsSlice {
			if indexVal == nil {
				indexVal = int64(0)
			}
			if indexEndVal == nil {
				indexEndVal = int64(len(indexElement.elements))
			}
			var indexValInt int64
			var indexEndValInt int64
			switch indexVal := indexVal.(type) {
			case int64:
				indexValInt = indexVal
			case *big.Int:
				if !indexVal.IsInt64() {
					return invalidBigintErr(indexVal)
				}
				indexValInt = indexVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, BufferIndexMustBeWholeNum(indexVal))
			}
			switch indexEndVal := indexEndVal.(type) {
			case int64:
				indexEndValInt = indexEndVal
			case *big.Int:
				if !indexEndVal.IsInt64() {
					return invalidBigintErr(indexEndVal)
				}
				indexEndValInt = indexEndVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, BufferIndexMustBeWholeNum(indexEndVal))
			}
			originalIndexValInt := indexValInt
			if indexValInt < 0 {
				indexValInt += int64(len(indexElement.elements))
			}
			if indexEndValInt < 0 {
				indexEndValInt += int64(len(indexElement.elements))
			}
			if indexEndValInt > int64(len(indexElement.elements)) {
				indexEndValInt = int64(len(indexElement.elements))
			}
			if indexValInt < 0 {
				return nil, loxerror.RuntimeError(expr.Bracket, BufferIndexOutOfRange(originalIndexValInt))
			}
			capacity := indexEndValInt - indexValInt
			if capacity < 0 {
				capacity = 0
			}
			listSlice := list.NewListCap[any](capacity)
			for i := indexValInt; i < indexEndValInt; i++ {
				listSlice.Add(indexElement.elements[i])
			}
			return NewLoxBuffer(listSlice), nil
		} else {
			var indexValInt int64
			switch indexVal := indexVal.(type) {
			case int64:
				indexValInt = indexVal
			case *big.Int:
				if !indexVal.IsInt64() {
					return invalidBigintErr(indexVal)
				}
				indexValInt = indexVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, BufferIndexMustBeWholeNum(indexVal))
			}
			originalIndexValInt := indexValInt
			if indexValInt < 0 {
				indexValInt += int64(len(indexElement.elements))
			}
			if indexValInt < 0 || indexValInt >= int64(len(indexElement.elements)) {
				return nil, loxerror.RuntimeError(expr.Bracket, BufferIndexOutOfRange(originalIndexValInt))
			}
			return indexElement.elements[indexValInt], nil
		}
	case *LoxDict:
		if expr.IsSlice {
			return nil, loxerror.RuntimeError(expr.Bracket, "Cannot use slice to index into dictionary.")
		}
		value, ok := indexElement.getValueByKey(indexVal)
		if !ok {
			return nil, loxerror.RuntimeError(expr.Bracket, UnknownDictKey(indexVal))
		}
		return value, nil
	case *LoxList:
		if expr.IsSlice {
			if indexVal == nil {
				indexVal = int64(0)
			}
			if indexEndVal == nil {
				indexEndVal = int64(len(indexElement.elements))
			}
			var indexValInt int64
			var indexEndValInt int64
			switch indexVal := indexVal.(type) {
			case int64:
				indexValInt = indexVal
			case *big.Int:
				if !indexVal.IsInt64() {
					return invalidBigintErr(indexVal)
				}
				indexValInt = indexVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, ListIndexMustBeWholeNum(indexVal))
			}
			switch indexEndVal := indexEndVal.(type) {
			case int64:
				indexEndValInt = indexEndVal
			case *big.Int:
				if !indexEndVal.IsInt64() {
					return invalidBigintErr(indexEndVal)
				}
				indexEndValInt = indexEndVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, ListIndexMustBeWholeNum(indexEndVal))
			}
			originalIndexValInt := indexValInt
			if indexValInt < 0 {
				indexValInt += int64(len(indexElement.elements))
			}
			if indexEndValInt < 0 {
				indexEndValInt += int64(len(indexElement.elements))
			}
			if indexEndValInt > int64(len(indexElement.elements)) {
				indexEndValInt = int64(len(indexElement.elements))
			}
			if indexValInt < 0 {
				return nil, loxerror.RuntimeError(expr.Bracket, ListIndexOutOfRange(originalIndexValInt))
			}
			capacity := indexEndValInt - indexValInt
			if capacity < 0 {
				capacity = 0
			}
			listSlice := list.NewListCap[any](capacity)
			for i := indexValInt; i < indexEndValInt; i++ {
				listSlice.Add(indexElement.elements[i])
			}
			return NewLoxList(listSlice), nil
		} else {
			var indexValInt int64
			switch indexVal := indexVal.(type) {
			case int64:
				indexValInt = indexVal
			case *big.Int:
				if !indexVal.IsInt64() {
					return invalidBigintErr(indexVal)
				}
				indexValInt = indexVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, ListIndexMustBeWholeNum(indexVal))
			}
			originalIndexValInt := indexValInt
			if indexValInt < 0 {
				indexValInt += int64(len(indexElement.elements))
			}
			if indexValInt < 0 || indexValInt >= int64(len(indexElement.elements)) {
				return nil, loxerror.RuntimeError(expr.Bracket, ListIndexOutOfRange(originalIndexValInt))
			}
			return indexElement.elements[indexValInt], nil
		}
	case *LoxRange:
		if expr.IsSlice {
			rangeLength := indexElement.Length()
			if indexVal == nil {
				indexVal = int64(0)
			}
			if indexEndVal == nil {
				indexEndVal = rangeLength
			}
			var indexValInt int64
			var indexEndValInt int64
			switch indexVal := indexVal.(type) {
			case int64:
				indexValInt = indexVal
			case *big.Int:
				if !indexVal.IsInt64() {
					return invalidBigintErr(indexVal)
				}
				indexValInt = indexVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, RangeIndexMustBeWholeNum(indexVal))
			}
			switch indexEndVal := indexEndVal.(type) {
			case int64:
				indexEndValInt = indexEndVal
			case *big.Int:
				if !indexEndVal.IsInt64() {
					return invalidBigintErr(indexEndVal)
				}
				indexEndValInt = indexEndVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, RangeIndexMustBeWholeNum(indexEndVal))
			}
			if indexValInt < 0 {
				indexValInt += rangeLength
			}
			if indexEndValInt < 0 {
				indexEndValInt += rangeLength
			}
			if indexEndValInt > rangeLength {
				indexEndValInt = rangeLength
			}
			return indexElement.getRange(indexValInt, indexEndValInt), nil
		} else {
			var indexValInt int64
			switch indexVal := indexVal.(type) {
			case int64:
				indexValInt = indexVal
			case *big.Int:
				if !indexVal.IsInt64() {
					return invalidBigintErr(indexVal)
				}
				indexValInt = indexVal.Int64()
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, RangeIndexMustBeWholeNum(indexVal))
			}
			originalIndexValInt := indexValInt
			rangeLength := indexElement.Length()
			if indexValInt < 0 {
				indexValInt += rangeLength
			}
			if indexValInt < 0 || indexValInt >= rangeLength {
				return nil, loxerror.RuntimeError(expr.Bracket, RangeIndexOutOfRange(originalIndexValInt))
			}
			return indexElement.get(indexValInt), nil
		}
	case *LoxBigRange:
		if expr.IsSlice {
			rangeLength := big.NewInt(indexElement.Length())
			setIndexVal, setEndIndexVal := false, false
			if indexVal == nil {
				setIndexVal = true
				indexVal = big.NewInt(0)
			}
			if indexEndVal == nil {
				setEndIndexVal = true
				indexEndVal = rangeLength
			}
			var indexValInt, indexEndValInt *big.Int
			switch indexVal := indexVal.(type) {
			case int64:
				indexValInt = big.NewInt(indexVal)
			case *big.Int:
				if setIndexVal {
					indexValInt = indexVal
				} else {
					indexValInt = new(big.Int).Set(indexVal)
				}
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, BigRangeIndexMustBeWholeNum(indexVal))
			}
			switch indexEndVal := indexEndVal.(type) {
			case int64:
				indexEndValInt = big.NewInt(indexEndVal)
			case *big.Int:
				if setEndIndexVal {
					indexEndValInt = indexEndVal
				} else {
					indexEndValInt = new(big.Int).Set(indexEndVal)
				}
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, BigRangeIndexMustBeWholeNum(indexVal))
			}
			if indexValInt.Cmp(bigint.Zero) < 0 {
				indexValInt.Add(indexValInt, rangeLength)
			}
			if indexEndValInt.Cmp(bigint.Zero) < 0 {
				indexEndValInt.Add(indexEndValInt, rangeLength)
			}
			if indexEndValInt.Cmp(rangeLength) > 0 {
				indexEndValInt = rangeLength
			}
			return indexElement.getRange(indexValInt, indexEndValInt), nil
		} else {
			var indexValInt *big.Int
			switch indexVal := indexVal.(type) {
			case int64:
				indexValInt = big.NewInt(indexVal)
			case *big.Int:
				indexValInt = new(big.Int).Set(indexVal)
			default:
				return nil, loxerror.RuntimeError(expr.Bracket, BigRangeIndexMustBeWholeNum(indexVal))
			}
			originalIndexValInt := new(big.Int).Set(indexValInt)
			rangeLength := big.NewInt(indexElement.Length())
			if indexValInt.Cmp(bigint.Zero) < 0 {
				indexValInt.Add(indexValInt, rangeLength)
			}
			if indexValInt.Cmp(bigint.Zero) < 0 || indexValInt.Cmp(rangeLength) >= 0 {
				return nil, loxerror.RuntimeError(expr.Bracket, BigRangeIndexOutOfRange(originalIndexValInt))
			}
			return indexElement.get(indexValInt), nil
		}
	}
	return nil, loxerror.RuntimeError(expr.Bracket,
		fmt.Sprintf("Cannot index into type '%v'.", getType(indexElement)))
}

func (i *Interpreter) visitListExpr(expr List) (any, error) {
	elements := list.NewListCap[any](int64(len(expr.Elements)))
	for _, element := range expr.Elements {
		switch element := element.(type) {
		case Spread:
			evalResult, evalErr := i.evaluate(element.Iterable)
			if evalErr != nil {
				elements.Clear()
				return nil, evalErr
			}
			switch evalResult := evalResult.(type) {
			case interfaces.Iterable:
				it := evalResult.Iterator()
				for it.HasNext() {
					elements.Add(it.Next())
				}
			default:
				elements.Clear()
				return nil, loxerror.RuntimeError(element.SpreadToken,
					"Value after '...' must be an iterable.")
			}
		default:
			evalResult, evalErr := i.evaluate(element)
			if evalErr != nil {
				elements.Clear()
				return nil, evalErr
			}
			elements.Add(evalResult)
		}
	}
	return NewLoxList(elements), nil
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

func (i *Interpreter) visitLoopStmt(stmt Loop) (any, error) {
	loopBlock := stmt.LoopBlock.(Block)
	enteredLoop := false
	loopInterrupted := false
	for {
		if !enteredLoop {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)
			defer func() {
				if !loopInterrupted {
					sigChan <- loxsignal.LoopSignal{}
					signal.Stop(sigChan)
				}
			}()
			go func() {
				sig := <-sigChan
				switch sig {
				case os.Interrupt:
					loopInterrupted = true
					signal.Stop(sigChan)
				}
			}()
			enteredLoop = true
		}
		if loopInterrupted {
			return nil, loxerror.RuntimeError(stmt.LoopToken, "loop interrupted")
		}
		value, evalErr := i.visitBlockStmt(loopBlock)
		if evalErr != nil {
			switch value := value.(type) {
			case Break, Return:
				return value, evalErr
			case Continue:
			default:
				return nil, evalErr
			}
		}
	}
}

func (i *Interpreter) visitPrintingStmt(stmt Print) (any, error) {
	value, evalErr := i.evaluate(stmt.Expression)
	if evalErr != nil {
		return nil, evalErr
	}
	if stmt.Stderr {
		if stmt.NewLine {
			fmt.Fprintln(os.Stderr, getResult(value, value, true))
		} else {
			fmt.Fprint(os.Stderr, getResult(value, value, true))
		}
	} else {
		if stmt.NewLine {
			fmt.Println(getResult(value, value, true))
		} else {
			fmt.Print(getResult(value, value, true))
		}
	}
	return nil, nil
}

func (i *Interpreter) visitRepeatStmt(stmt Repeat) (any, error) {
	expr, exprErr := i.evaluate(stmt.Expression)
	if exprErr != nil {
		return nil, exprErr
	}
	var repeatTimes int64 = 0
	var repeatTimesBigInt *big.Int
	useBigInt := false
	switch expr := expr.(type) {
	case int64:
		repeatTimes = expr
	case *big.Int:
		repeatTimesBigInt = expr
		useBigInt = true
	case bool:
		if expr {
			repeatTimes = 1
		}
	case nil:
	default:
		return nil, loxerror.RuntimeError(stmt.RepeatToken,
			"Repeat statement expression must be an integer or bigint.")
	}
	enteredLoop := false
	loopInterrupted := false
	if useBigInt {
		times := repeatTimesBigInt
		one := bigint.BoolMap[true]
		for count := big.NewInt(0); count.Cmp(times) < 0; count.Add(count, one) {
			if !enteredLoop {
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt)
				defer func() {
					if !loopInterrupted {
						sigChan <- loxsignal.LoopSignal{}
						signal.Stop(sigChan)
					}
				}()
				go func() {
					sig := <-sigChan
					switch sig {
					case os.Interrupt:
						loopInterrupted = true
						signal.Stop(sigChan)
					}
				}()
				enteredLoop = true
			}
			if loopInterrupted {
				return nil, loxerror.RuntimeError(stmt.RepeatToken, "loop interrupted")
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
		}
	} else {
		for count := int64(0); count < repeatTimes; count++ {
			if !enteredLoop {
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt)
				defer func() {
					if !loopInterrupted {
						sigChan <- loxsignal.LoopSignal{}
						signal.Stop(sigChan)
					}
				}()
				go func() {
					sig := <-sigChan
					switch sig {
					case os.Interrupt:
						loopInterrupted = true
						signal.Stop(sigChan)
					}
				}()
				enteredLoop = true
			}
			if loopInterrupted {
				return nil, loxerror.RuntimeError(stmt.RepeatToken, "loop interrupted")
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
		}
	}
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
	if class, ok := obj.(*LoxClass); ok {
		return evaluateAndSet(func(value any) {
			class.classProperties[expr.Name.Lexeme] = value
		})
	}
	return nil, loxerror.RuntimeError(expr.Name, "Only classes and instances have properties that can be set.")
}

func (i *Interpreter) visitSetObjectExpr(expr SetObject) (any, error) {
	indexes := list.NewList[any]()
	defer indexes.Clear()

	node := expr.Object
	for index, ok := node.(Index); ok; {
		indexVal, indexValErr := i.evaluate(index.Index)
		if indexValErr != nil {
			return nil, indexValErr
		}
		indexes.Add(indexVal)
		node = index.IndexElement
		index, ok = node.(Index)
	}

	variable, variableErr := i.evaluate(node)
	if variableErr != nil {
		return nil, variableErr
	}
	assignErrMsg := "Can only assign to buffer, dictionary, and list indexes."
	switch variable := variable.(type) {
	case *LoxBuffer:
		value, valueErr := i.evaluate(expr.Value)
		if valueErr != nil {
			return nil, valueErr
		}
		if len(indexes) > 1 {
			return nil, loxerror.RuntimeError(expr.Name, BufferNestedElementErrMsg)
		}
		index := indexes[len(indexes)-1]
		switch index := index.(type) {
		case int64:
			bufferSetErr := variable.setIndex(index, value)
			if bufferSetErr != nil {
				return nil, loxerror.RuntimeError(expr.Name, bufferSetErr.Error())
			}
			return value, nil
		default:
			return nil, loxerror.RuntimeError(expr.Name, BufferIndexMustBeWholeNum(index))
		}
	case *LoxDict:
		value, valueErr := i.evaluate(expr.Value)
		if valueErr != nil {
			return nil, valueErr
		}
		for loopIndex := len(indexes) - 1; loopIndex >= 0; loopIndex-- {
			index := indexes[loopIndex]
			if loopIndex > 0 {
				var ok bool
				var keyValue any
				keyValue, ok = variable.getValueByKey(index)
				if !ok {
					return nil, loxerror.RuntimeError(expr.Name, assignErrMsg)
				}
				variable, ok = keyValue.(*LoxDict)
				if !ok {
					return nil, loxerror.RuntimeError(expr.Name, assignErrMsg)
				}
			} else {
				canBeKey, keyErr := CanBeDictKeyCheck(index)
				if !canBeKey {
					return nil, loxerror.RuntimeError(expr.Name, keyErr)
				}
				variable.setKeyValue(index, value)
			}
		}
		return value, nil
	case *LoxList:
		value, valueErr := i.evaluate(expr.Value)
		if valueErr != nil {
			return nil, valueErr
		}
		for loopIndex := len(indexes) - 1; loopIndex >= 0; loopIndex-- {
			index := indexes[loopIndex]
			switch index := index.(type) {
			case int64:
				originalIndex := index
				if index < 0 {
					index += int64(len(variable.elements))
				}
				if loopIndex > 0 {
					if index < 0 || index >= int64(len(variable.elements)) {
						return nil, loxerror.RuntimeError(expr.Name, ListIndexOutOfRange(originalIndex))
					}
					var ok bool
					variable, ok = variable.elements[index].(*LoxList)
					if !ok {
						return nil, loxerror.RuntimeError(expr.Name, assignErrMsg)
					}
				} else {
					if index < 0 || index >= int64(len(variable.elements)) {
						return nil, loxerror.RuntimeError(expr.Name, ListIndexOutOfRange(originalIndex))
					}
					variable.elements[index] = value
				}
			default:
				return nil, loxerror.RuntimeError(expr.Name, ListIndexMustBeWholeNum(index))
			}
		}
		return value, nil
	}
	return nil, loxerror.RuntimeError(expr.Name, assignErrMsg)
}

func (i *Interpreter) visitStringExpr(expr String) (any, error) {
	return NewLoxString(expr.Str, expr.Quote), nil
}

func (i *Interpreter) visitSuperExpr(expr Super) (any, error) {
	distance := i.locals[expr]
	superClass := i.environment.GetAtStr(distance, "super").(*LoxClass)
	object := i.environment.GetAtStr(distance-1, "this")
	switch object := object.(type) {
	case *LoxInstance:
		method, ok := superClass.findMethod(expr.Method.Lexeme)
		if ok {
			return method.bind(object), nil
		}
		field, ok := superClass.findInstanceField(expr.Method.Lexeme)
		if ok {
			switch method := field.(type) {
			case *struct{ ProtoLoxCallable }:
				return LoxBuiltInProtoCallable{object, method}, nil
			}
		}
	case *LoxClass:
		field, ok := superClass.Get(expr.Method)
		if ok == nil {
			switch field := field.(type) {
			case *LoxFunction:
				return field.bind(object), nil
			}
			return field, nil
		}
	}
	return nil, loxerror.RuntimeError(expr.Method, "Undefined property '"+expr.Method.Lexeme+"'.")
}

func (i *Interpreter) visitTernaryExpr(expr Ternary) (any, error) {
	condition, conditionErr := i.evaluate(expr.Condition)
	if conditionErr != nil {
		return nil, conditionErr
	}
	if i.isTruthy(condition) {
		return i.evaluate(expr.TrueExpr)
	} else {
		return i.evaluate(expr.FalseExpr)
	}
}

func (i *Interpreter) visitThisExpr(expr This) (any, error) {
	distance, ok := i.locals[expr]
	if ok {
		return i.environment.GetAt(distance, expr.Keyword)
	} else {
		return i.globals.Get(expr.Keyword)
	}
}

func (i *Interpreter) visitThrowStmt(stmt Throw) (any, error) {
	throwValue, throwValueErr := i.evaluate(stmt.Value)
	if throwValueErr != nil {
		return nil, throwValueErr
	}
	var throwValueStr string
	switch throwValue := throwValue.(type) {
	case *LoxError:
		return nil, throwValue.theError
	case *LoxString:
		throwValueStr = throwValue.str
	default:
		//Use string representation of throw expression as error message
		result, _ := i.visitBinaryExpr(Binary{
			Literal{NewLoxString("", '\'')},
			&token.Token{
				TokenType: token.PLUS,
				Lexeme:    "+",
			},
			Literal{throwValue},
		})
		throwValueStr = result.(*LoxString).str
	}
	return nil, loxerror.RuntimeError(stmt.ThrowToken, throwValueStr)
}

func (i *Interpreter) visitTryCatchFinallyStmt(stmt TryCatchFinally) (any, error) {
	finallyBlock := func(originalAny any, originalErr error) (any, error) {
		if stmt.FinallyBlock != nil {
			finallyValue, finallyErr := i.visitBlockStmt(stmt.FinallyBlock.(Block))
			if finallyErr != nil {
				switch finallyValue := finallyValue.(type) {
				case Break, Continue, Return:
					return finallyValue, finallyErr
				}
				return nil, finallyErr
			}
		}
		return originalAny, originalErr
	}
	tryValue, tryErr := i.visitBlockStmt(stmt.TryBlock.(Block))
	if tryErr != nil {
		switch tryValue := tryValue.(type) {
		case Break, Continue, Return:
			return finallyBlock(tryValue, tryErr)
		}
		if stmt.CatchBlock != nil {
			var catchValue any
			var catchErr error
			if stmt.CatchName != nil {
				catchBlockEnv := env.NewEnvironmentEnclosing(i.environment)
				catchBlockEnv.Define(stmt.CatchName.Lexeme, NewLoxError(tryErr))
				catchValue, catchErr = i.visitBlockStmtEnv(stmt.CatchBlock.(Block), catchBlockEnv)
			} else {
				catchValue, catchErr = i.visitBlockStmt(stmt.CatchBlock.(Block))
			}
			if catchErr != nil {
				switch catchValue := catchValue.(type) {
				case Break, Continue, Return:
					return finallyBlock(catchValue, catchErr)
				}
				return finallyBlock(nil, catchErr)
			}
		} else {
			return finallyBlock(nil, tryErr)
		}
	}
	return finallyBlock(nil, nil)
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
		case *big.Int:
			return new(big.Int).Neg(right), nil
		case *big.Float:
			return new(big.Float).Neg(right), nil
		case bool:
			if right {
				return int64(-1), nil
			}
			return int64(0), nil
		case nil:
			return int64(0), nil
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
		case *big.Int:
			return new(big.Int).Not(right), nil
		case *big.Float:
			bigInt := &big.Int{}
			right.Int(bigInt)
			return bigInt.Not(bigInt), nil
		case bool:
			if right {
				return ^int64(1), nil
			}
			return ^int64(0), nil
		case nil:
			return ^int64(0), nil
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
	var variable any
	var variableErr error
	distance, ok := i.locals[expr.Name]
	if ok {
		variable, variableErr = i.environment.GetAt(distance, expr.Name)
	} else {
		variable, variableErr = i.globals.Get(expr.Name)
	}
	switch variable := variable.(type) {
	case interfaces.LazyType:
		evalErr := variable.LazyTypeEval()
		if evalErr != nil {
			return nil, evalErr
		}
	}
	return variable, variableErr
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
			defer func() {
				if !loopInterrupted {
					sigChan <- loxsignal.LoopSignal{}
					signal.Stop(sigChan)
				}
			}()
			go func() {
				sig := <-sigChan
				switch sig {
				case os.Interrupt:
					loopInterrupted = true
					signal.Stop(sigChan)
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
