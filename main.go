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

func run(sc *scanner.Scanner, interpreter *ast.Interpreter) error {
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

	textSc := bufio.NewScanner(file)
	textSc.Scan()
	var program strings.Builder
	for {
		line := strings.TrimSpace(textSc.Text())
		program.WriteString(line)
		if !textSc.Scan() {
			break
		}
		program.WriteByte('\n')
	}

	sc := scanner.NewScanner(program.String())
	interpreter := ast.NewInterpreter()
	resultError := run(sc, interpreter)
	if resultError != nil {
		return resultError
	}

	return nil
}

func interactiveMode() int {
	l, _ := readline.NewEx(&readline.Config{
		Prompt:          PROMPT,
		InterruptPrompt: "^C",
	})
	defer l.Close()
	l.CaptureExitSignal()

	stdinFromTerminal := func() bool {
		stat, _ := os.Stdin.Stat()
		return (stat.Mode() & os.ModeCharDevice) != 0
	}()
	sc := scanner.NewScanner("")
	interpreter := ast.NewInterpreter()
	for {
		userInput, readError := l.Readline()
		if readError == readline.ErrInterrupt {
			continue
		} else if readError == io.EOF {
			break
		}

		userInput = strings.TrimSpace(userInput)
		if len(userInput) != 0 {
			sc.SetSourceLine(userInput)
			resultError := run(sc, interpreter)
			if resultError != nil {
				loxerror.PrintErrorObject(resultError)
				if !stdinFromTerminal {
					return 1
				}
			}
		}

		sc.ResetForNextLine()
		if !stdinFromTerminal {
			sc.IncreaseLineNum()
		}
	}
	return 0
}

func main() {
	args := os.Args
	exprCLine := flag.String("c", "", "Read code from command line")
	flag.Parse()

	if *exprCLine != "" {
		sc := scanner.NewScanner(*exprCLine)
		resultError := run(sc, ast.NewInterpreter())
		if resultError != nil {
			loxerror.PrintErrorObject(resultError)
			os.Exit(1)
		}
	} else if len(args) > 1 {
		possibleError := processFile(args[1])
		if possibleError != nil {
			loxerror.PrintErrorObject(possibleError)
			os.Exit(1)
		}
	} else {
		os.Exit(interactiveMode())
	}
}
