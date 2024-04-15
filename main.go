package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"strings"

	"github.com/AlanLuu/lox/ast"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/scanner"
	"github.com/chzyer/readline"
)

const PROMPT = ">>> "

func run(source string, scParam *scanner.Scanner) error {
	var sc *scanner.Scanner
	if scParam != nil {
		sc = scParam
	} else {
		sc = scanner.NewScanner(source)
	}
	scanErr := sc.ScanTokens()
	if scanErr != nil {
		return scanErr
	}

	parser := ast.NewParser(sc.Tokens)
	exprList, parseErr := parser.Parse()
	defer exprList.Clear()
	if parseErr != nil {
		return parseErr
	}

	interpreter := ast.Interpreter{}
	valueErr := interpreter.Interpret(exprList)
	if valueErr != nil {
		return valueErr
	}

	return nil
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
		resultError := run(line, sc)
		if resultError != nil {
			return resultError
		}

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

		resultError := run(userInput, nil)
		if resultError != nil {
			loxerror.PrintErrorObject(resultError)
		}
	}
}

func main() {
	args := os.Args
	exprCLine := flag.String("c", "", "Read code from command line")
	flag.Parse()

	if *exprCLine != "" {
		resultError := run(*exprCLine, nil)
		if resultError != nil {
			loxerror.PrintErrorObject(resultError)
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
