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
			program.WriteByte('\n')
			if stdinFromTerminal {
				scopeLevel += (strings.Count(userInput, "{") - strings.Count(userInput, "}"))
				if scopeLevel <= 0 {
					break
				}
			}
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

	if *exprCLine != "" {
		os.Stdin.Close()
		sc := scanner.NewScanner(*exprCLine)
		resultError := run(sc, ast.NewInterpreter())
		if resultError != nil {
			loxerror.PrintErrorObject(resultError)
			os.Exit(1)
		}
	} else if len(args) > 1 {
		os.Stdin.Close()
		possibleError := processFile(args[1])
		if possibleError != nil {
			loxerror.PrintErrorObject(possibleError)
			os.Exit(1)
		}
	} else {
		os.Exit(interactiveMode())
	}
}
