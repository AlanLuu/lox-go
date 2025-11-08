package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"runtime"
	"strings"

	"github.com/AlanLuu/lox/ast"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/scanner"
	"github.com/AlanLuu/lox/util"
	"github.com/chzyer/readline"
)

const LOX_PROGRAM_NAME = "lox-go"

const PROMPT = ">>> "
const NEXT_LINE_PROMPT = "... "

//go:embed loxcode/*
var loxCodeFS embed.FS

func flagsProvided() map[string]struct{} {
	m := map[string]struct{}{}
	flag.Visit(func(f *flag.Flag) {
		m[f.Name] = struct{}{}
	})
	return m
}

func printVersion() {
	var builder strings.Builder
	builder.WriteString(LOX_PROGRAM_NAME)
	if hash := util.GitHashShort(); hash != "" {
		builder.WriteByte('-')
		builder.WriteString(hash)
	}
	builder.WriteByte(' ')
	builder.WriteByte('(')
	builder.WriteString(runtime.Version())
	builder.WriteString("-" + runtime.GOOS + "-" + runtime.GOARCH)
	builder.WriteByte(')')
	fmt.Println(builder.String())
}

func usageFunc(writer io.Writer) func() {
	return func() {
		usage :=
			`Usage: lox [OPTIONS] [FILE]

OPTIONS:
	-c <code>
		Execute Lox code from command line argument
	-i
		Drop into REPL mode after running Lox code
	--disable-loxcode, -dl
		Disable execution of all Lox files that are bundled inside this interpreter executable
	--unsafe
		Enable unsafe mode, allowing access to functions that can potentially crash this interpreter
	-h, --help
		Print this usage message and exit
	-v, --version
		Print version information and exit
`
		fmt.Fprint(writer, usage)
	}
}

func runLoxCode(interpreter *ast.Interpreter) error {
	if util.DisableLoxCode {
		return nil
	}
	dirFunc := func(path string, d fs.DirEntry, _ error) error {
		if !d.IsDir() {
			program, err := loxCodeFS.ReadFile(path)
			if err != nil {
				fmt.Fprintf(
					os.Stderr,
					"Warning: failed to read Lox file '%v'.\n",
					path,
				)
				return nil
			}

			sc := scanner.NewScanner(string(program))
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

			valueErr := interpreter.Interpret(exprList, true)
			if valueErr != nil {
				return valueErr
			}
		}
		return nil
	}
	return fs.WalkDir(loxCodeFS, ".", dirFunc)
}

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

	valueErr := interpreter.Interpret(exprList, true)
	if valueErr != nil {
		return valueErr
	}

	return nil
}

func processFile(filePath string) (*ast.Interpreter, error) {
	file, openFileError := os.Open(filePath)
	if openFileError != nil {
		return nil, openFileError
	}

	program, readErr := io.ReadAll(file)
	file.Close()
	if readErr != nil {
		return nil, readErr
	}

	interpreter := ast.NewInterpreter()
	runLoxCodeErr := runLoxCode(interpreter)
	if runLoxCodeErr != nil {
		return interpreter, runLoxCodeErr
	}

	sc := scanner.NewScanner(string(program))
	return interpreter, run(sc, interpreter)
}

func interactiveMode(interpreter *ast.Interpreter) int {
	util.InteractiveMode = true
	runLoxCodeErr := runLoxCode(interpreter)
	if runLoxCodeErr != nil {
		loxerror.PrintErrorObject(runLoxCodeErr)
		return 1
	}
	if util.StdinFromTerminal() {
		l, readlineErr := readline.NewEx(&readline.Config{
			Prompt:          PROMPT,
			InterruptPrompt: "^C",
		})
		if readlineErr != nil {
			//Should never happen
			fmt.Fprintf(
				os.Stderr,
				"Failed to launch REPL due to following error: %v\n",
				readlineErr.Error(),
			)
			return 1
		}
		defer l.Close()
		numSpacesIndent := 2
	outer:
		for {
			var program strings.Builder
			scopeLevel := 0
			for {
				if scopeLevel > 0 {
					indent := strings.Repeat(" ", numSpacesIndent*scopeLevel)
					l.SetPrompt(NEXT_LINE_PROMPT + indent)
				} else {
					l.SetPrompt(PROMPT)
				}

				userInput, readError := l.Readline()
				if readError == readline.ErrInterrupt {
					continue
				} else if readError == io.EOF {
					break outer
				}

				if len(userInput) == 0 {
					if scopeLevel > 0 {
						program.WriteByte('\n')
					}
					continue
				}
				userInput = strings.TrimSpace(userInput)
				program.WriteString(userInput)
				leftBraceCount, rightBraceCount := util.CountBraces(userInput)
				scopeLevel += (leftBraceCount - rightBraceCount)
				if scopeLevel <= 0 {
					break
				}
				program.WriteByte('\n')
			}

			sc := scanner.NewScanner(program.String())
			resultError := run(sc, interpreter)
			if resultError != nil {
				loxerror.PrintErrorObject(resultError)
			}
		}
	} else {
		program, readErr := io.ReadAll(os.Stdin)
		if readErr != nil {
			loxerror.PrintErrorObject(readErr)
			return 1
		}

		sc := scanner.NewScanner(string(program))
		resultError := run(sc, interpreter)
		if resultError != nil {
			loxerror.PrintErrorObject(resultError)
			return 1
		}
	}

	return 0
}

func main() {
	var (
		exprCLine       = flag.String("c", "", "")
		interactive     = flag.Bool("i", false, "")
		disableLoxCode  = flag.Bool("disable-loxcode", false, "")
		disableLoxCode2 = flag.Bool("dl", false, "")
		unsafe          = flag.Bool("unsafe", false, "")
		helpFlag1       = flag.Bool("h", false, "")
		helpFlag2       = flag.Bool("help", false, "")
		versionFlag1    = flag.Bool("v", false, "")
		versionflag2    = flag.Bool("version", false, "")
	)
	flag.Usage = usageFunc(os.Stderr)
	flag.Parse()
	if *helpFlag1 || *helpFlag2 {
		usageFunc(os.Stdout)()
		os.Exit(0)
	}
	if *versionFlag1 || *versionflag2 {
		printVersion()
		os.Exit(0)
	}

	args := flag.Args()
	util.DisableLoxCode = *disableLoxCode || *disableLoxCode2
	util.UnsafeMode = *unsafe
	exitCode := 0
	flagsMap := flagsProvided()
	var interpreter *ast.Interpreter
	if _, ok := flagsMap["c"]; ok {
		line := strings.TrimSpace(*exprCLine)
		if line != "" {
			sc := scanner.NewScanner(line)
			interpreter = ast.NewInterpreter()
			runLoxCodeErr := runLoxCode(interpreter)
			if runLoxCodeErr == nil {
				resultError := run(sc, interpreter)
				if resultError != nil {
					loxerror.PrintErrorObject(resultError)
					exitCode = 1
				}
			} else {
				loxerror.PrintErrorObject(runLoxCodeErr)
				exitCode = 1
			}
		}
	} else if len(args) > 0 && args[0] != "-" {
		var possibleError error
		interpreter, possibleError = processFile(args[0])
		if possibleError != nil {
			loxerror.PrintErrorObject(possibleError)
			exitCode = 1
		}
	} else {
		exitCode = interactiveMode(ast.NewInterpreter())
	}

	if *interactive && !util.InteractiveMode {
		if interpreter != nil {
			exitCode = interactiveMode(interpreter)
		} else {
			exitCode = interactiveMode(ast.NewInterpreter())
		}
	}

	ast.OSExit(exitCode)
}
