package main

import (
	"flag"
	"io"
	"os"
	"strings"

	"github.com/AlanLuu/lox/ast"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/scanner"
	"github.com/AlanLuu/lox/util"
	"github.com/chzyer/readline"
)

const PROMPT = ">>> "
const NEXT_LINE_PROMPT = "... "

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

	resolver := ast.NewResolver(interpreter)
	resolverErr := resolver.Resolve(exprList)
	if resolverErr != nil {
		return resolverErr
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

	program, readErr := io.ReadAll(file)
	file.Close()
	if readErr != nil {
		return readErr
	}
	sc := scanner.NewScanner(string(program))
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

	interpreter := ast.NewInterpreter()
	stdinFromTerminal := util.StdinFromTerminal()
	numSpacesIndent := 2

outer:
	for {
		var program strings.Builder
		scopeLevel := 0
		for {
			if stdinFromTerminal {
				if scopeLevel > 0 {
					indent := strings.Repeat(" ", numSpacesIndent*scopeLevel)
					l.SetPrompt(NEXT_LINE_PROMPT + indent)
				} else {
					l.SetPrompt(PROMPT)
				}
			}

			userInput, readError := l.Readline()
			if readError == readline.ErrInterrupt {
				continue
			} else if readError == io.EOF {
				if stdinFromTerminal {
					break outer
				} else {
					break
				}
			}

			if len(userInput) == 0 {
				if !stdinFromTerminal || scopeLevel > 0 {
					program.WriteByte('\n')
				}
				continue
			}
			userInput = strings.TrimSpace(userInput)
			program.WriteString(userInput)
			if stdinFromTerminal {
				leftBraceCount, rightBraceCount := util.CountBraces(userInput)
				scopeLevel += (leftBraceCount - rightBraceCount)
				if scopeLevel <= 0 {
					break
				}
			}
			program.WriteByte('\n')
		}

		sc := scanner.NewScanner(program.String())
		resultError := run(sc, interpreter)
		if resultError != nil {
			loxerror.PrintErrorObject(resultError)
			if !stdinFromTerminal {
				return 1
			}
		}

		if !stdinFromTerminal {
			break
		}
	}

	return 0
}

func main() {
	args := os.Args
	exprCLine := flag.String("c", "", "Read code from command line")
	flag.Parse()

	exitCode := 0
	if *exprCLine != "" {
		sc := scanner.NewScanner(*exprCLine)
		resultError := run(sc, ast.NewInterpreter())
		if resultError != nil {
			loxerror.PrintErrorObject(resultError)
			exitCode = 1
		}
	} else if len(args) > 1 {
		possibleError := processFile(args[1])
		if possibleError != nil {
			loxerror.PrintErrorObject(possibleError)
			exitCode = 1
		}
	} else {
		util.InteractiveMode = true
		exitCode = interactiveMode()
	}

	ast.CloseInputFuncReadline()
	os.Exit(exitCode)
}
