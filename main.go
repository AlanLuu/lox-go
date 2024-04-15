package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/AlanLuu/lox/ast"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/scanner"
	"github.com/chzyer/readline"
)

const PROMPT = ">>> "

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

func run(source string, scParam *scanner.Scanner) (any, error) {
	var sc *scanner.Scanner
	if scParam != nil {
		sc = scParam
	} else {
		sc = scanner.NewScanner(source)
	}
	scanErr := sc.ScanTokens()
	if scanErr != nil {
		return "", scanErr
	}

	parser := ast.NewParser(sc.Tokens)
	expr, parseErr := parser.Parse()
	if parseErr != nil {
		return "", parseErr
	}

	interpreter := ast.Interpreter{}
	value, valueErr := interpreter.Interpret(expr)
	if valueErr != nil {
		return "", valueErr
	}

	return value, nil
}

func processFile(filePath string) error {
	file, openFileError := os.Open(filePath)
	if openFileError != nil {
		return openFileError
	}

	sc := scanner.NewScanner("")
	textSc := bufio.NewScanner(file)
	for textSc.Scan() {
		line := strings.TrimSpace(textSc.Text())
		sc.IncreaseLineNum()
		if len(line) == 0 {
			continue
		}

		sc.SetSourceLine(line)
		result, resultError := run(line, sc)
		if resultError != nil {
			return resultError
		}

		printResult(result)
		sc.ResetForNextLine()
	}
	return nil
}

func interactiveMode() {
	l, _ := readline.NewEx(&readline.Config{
		Prompt:          PROMPT,
		InterruptPrompt: "^C",
	})
	defer l.Close()
	l.CaptureExitSignal()

	for {
		userInput, readError := l.Readline()
		if readError == readline.ErrInterrupt {
			continue
		} else if readError == io.EOF {
			break
		}

		userInput = strings.TrimSpace(userInput)
		if len(userInput) == 0 {
			continue
		}

		result, resultError := run(userInput, nil)
		if resultError != nil {
			loxerror.PrintErrorObject(resultError)
			continue
		}

		printResult(result)
	}
}

func main() {
	args := os.Args
	exprCLine := flag.String("c", "", "Read code from command line")
	flag.Parse()

	if *exprCLine != "" {
		result, resultError := run(*exprCLine, nil)
		if resultError != nil {
			loxerror.PrintErrorObject(resultError)
		} else {
			printResult(result)
		}
	} else if len(args) > 1 {
		possibleError := processFile(args[1])
		if possibleError != nil {
			loxerror.PrintErrorObject(possibleError)
			os.Exit(1)
		}
	} else {
		interactiveMode()
	}
}
