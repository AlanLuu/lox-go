package ast

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

const (
	filepath_skipDir int64 = iota + 1
	filepath_skipAll
)

func defineFilepathFields(filepathClass *LoxClass) {
	filepathClass.classProperties["listSep"] = NewLoxStringQuote(
		string(filepath.ListSeparator),
	)
	filepathClass.classProperties["sep"] = NewLoxStringQuote(
		string(filepath.Separator),
	)
	filepathClass.classProperties["skipDir"] = filepath_skipDir
	filepathClass.classProperties["skipAll"] = filepath_skipAll
}

func (i *Interpreter) defineFilepathFuncs() {
	className := "filepath"
	filepathClass := NewLoxClass(className, nil, false)
	filepathFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native filepath class fn %v at %p>", name, &s)
		}
		filepathClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'filepath.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	defineFilepathFields(filepathClass)
	filepathFunc("abs", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			result, err := filepath.Abs(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxStringQuote(result), nil
		}
		return argMustBeType(in.callToken, "abs", "string")
	})
	filepathFunc("base", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(filepath.Base(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "base", "string")
	})
	filepathFunc("clean", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(filepath.Clean(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "clean", "string")
	})
	filepathFunc("dir", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(filepath.Dir(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "dir", "string")
	})
	filepathFunc("evalSymlinks", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			result, err := filepath.EvalSymlinks(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxStringQuote(result), nil
		}
		return argMustBeType(in.callToken, "evalSymlinks", "string")
	})
	filepathFunc("ext", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(filepath.Ext(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "ext", "string")
	})
	filepathFunc("exts", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			s := []rune(loxStr.str)
			last := len(s)
			resultsList := list.NewList[any]()
			for i := last - 1; i >= 0 && !util.IsPathSep(s[i]); i-- {
				if s[i] == '.' {
					resultsList.Add(NewLoxStringQuote(string(s[i:last])))
					last = i
				}
			}
			slices.Reverse(resultsList)
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "exts", "string")
	})
	filepathFunc("fileInfo", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			str := loxStr.str
			stat, statErr := os.Lstat(str)
			if statErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, statErr.Error())
			}
			return NewLoxFileInfoPathFileInfo(str, stat), nil
		}
		return argMustBeType(in.callToken, "fileInfo", "string")
	})
	filepathFunc("fileInfoNil", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxFileInfo(loxStr.str, nil, nil), nil
		}
		return argMustBeType(in.callToken, "fileInfoNil", "string")
	})
	filepathFunc("fromSlash", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(filepath.FromSlash(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "fromSlash", "string")
	})
	filepathFunc("glob", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			globs, err := filepath.Glob(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			globsList := list.NewListCap[any](int64(len(globs)))
			for _, glob := range globs {
				globsList.Add(NewLoxStringQuote(glob))
			}
			return NewLoxList(globsList), nil
		}
		return argMustBeType(in.callToken, "glob", "string")
	})
	filepathFunc("isAbs", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return filepath.IsAbs(loxStr.str), nil
		}
		return argMustBeType(in.callToken, "isAbs", "string")
	})
	filepathFunc("isLocal", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return filepath.IsLocal(loxStr.str), nil
		}
		return argMustBeType(in.callToken, "isLocal", "string")
	})
	filepathFunc("join", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var strArgs []string
		for i, arg := range args {
			switch arg := arg.(type) {
			case *LoxString:
				if i == 0 {
					strArgs = make([]string, 0, len(args))
				}
				strArgs = append(strArgs, arg.str)
			default:
				strArgs = nil
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"Argument #%v in 'filepath.join' must be a string.",
						i+1,
					),
				)
			}
		}
		return NewLoxStringQuote(filepath.Join(strArgs...)), nil
	})
	filepathFunc("match", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'filepath.match' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'filepath.match' must be a string.")
		}
		pattern := args[0].(*LoxString).str
		name := args[1].(*LoxString).str
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return matched, nil
	})
	filepathFunc("split", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			dir, file := filepath.Split(loxStr.str)
			pair := list.NewListCap[any](2)
			pair.Add(NewLoxStringQuote(dir))
			pair.Add(NewLoxStringQuote(file))
			return NewLoxList(pair), nil
		}
		return argMustBeType(in.callToken, "split", "string")
	})
	filepathFunc("splitList", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			results := filepath.SplitList(loxStr.str)
			resultsList := list.NewListCap[any](int64(len(results)))
			for _, result := range results {
				resultsList.Add(NewLoxStringQuote(result))
			}
			return NewLoxList(resultsList), nil
		}
		return argMustBeType(in.callToken, "splitList", "string")
	})
	filepathFunc("stem", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			s := []rune(loxStr.str)
			sLen := len(s)
			end := sLen - 1
			start := end
			foundEnd := false
			foundStart := false
			for start > 0 {
				if !foundEnd {
					if s[end] != '.' {
						end--
					} else {
						foundEnd = true
					}
				}
				if util.IsPathSep(s[start]) {
					foundStart = true
					break
				}
				start--
			}
			if foundStart || start == -1 {
				start++
			}
			if !foundEnd || end == sLen-1 {
				end = sLen
			}
			return NewLoxStringQuote(string(s[start:end])), nil
		}
		return argMustBeType(in.callToken, "stem", "string")
	})
	filepathFunc("toSlash", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(filepath.ToSlash(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "toSlash", "string")
	})
	filepathFunc("volumeName", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(filepath.VolumeName(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "volumeName", "string")
	})
	filepathFunc("walk", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'filepath.walk' must be a string.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'filepath.walk' must be a function.")
		}
		pathArg := args[0].(*LoxString).str
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 3)
		defer argList.Clear()
		var index int64 = 0
		dirFunc := func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return loxerror.RuntimeError(in.callToken, err.Error())
			}
			argList[0] = NewLoxStringQuote(path)
			argList[1] = NewLoxFileInfoPathDirEntry(path, d)
			argList[2] = index
			index++
			result, resultErr := callback.call(i, argList)
			if resultReturn, ok := result.(Return); ok {
				result = resultReturn.FinalValue
			} else if resultErr != nil {
				return resultErr
			}
			if result, ok := result.(int64); ok {
				switch result {
				case filepath_skipDir:
					return fs.SkipDir
				case filepath_skipAll:
					return fs.SkipAll
				}
			}
			return nil
		}
		return nil, filepath.WalkDir(pathArg, dirFunc)
	})
	filepathFunc("walkFileInfo", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'filepath.walkFileInfo' must be a string.")
		}
		if _, ok := args[1].(*LoxFunction); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'filepath.walkFileInfo' must be a function.")
		}
		pathArg := args[0].(*LoxString).str
		callback := args[1].(*LoxFunction)
		argList := getArgList(callback, 2)
		defer argList.Clear()
		var index int64 = 0
		dirFunc := func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return loxerror.RuntimeError(in.callToken, err.Error())
			}
			argList[0] = NewLoxFileInfoPathDirEntry(path, d)
			argList[1] = index
			index++
			result, resultErr := callback.call(i, argList)
			if resultReturn, ok := result.(Return); ok {
				result = resultReturn.FinalValue
			} else if resultErr != nil {
				return resultErr
			}
			if result, ok := result.(int64); ok {
				switch result {
				case filepath_skipDir:
					return fs.SkipDir
				case filepath_skipAll:
					return fs.SkipAll
				}
			}
			return nil
		}
		return nil, filepath.WalkDir(pathArg, dirFunc)
	})
	filepathFunc("walkIter", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			callToken := in.callToken
			firstIter := true
			fileChan := make(chan *LoxFileInfo, 1)
			errorChan := make(chan error, 1)
			signalChan := make(chan struct{}, 1)
			doneChan := make(chan struct{}, 1)
			var current *LoxFileInfo
			iterator := ProtoIteratorErr{legacyPanicOnErr: true}
			iterator.hasNextMethod = func() (bool, error) {
				if firstIter {
					firstIter = false
					go func() {
						filepath.WalkDir(
							loxStr.str,
							func(path string, d fs.DirEntry, err error) error {
								<-signalChan
								if err != nil {
									errorChan <- err
									return err
								}
								fileChan <- NewLoxFileInfoPathDirEntry(
									path, d,
								)
								return nil
							},
						)
						doneChan <- struct{}{}
					}()
				}
				signalChan <- struct{}{}
				select {
				case err := <-errorChan:
					return false, loxerror.RuntimeError(
						callToken,
						err.Error(),
					)
				case current = <-fileChan:
					return true, nil
				case <-doneChan:
					return false, nil
				}
			}
			iterator.nextMethod = func() (any, error) {
				return current, nil
			}
			return NewLoxIterator(iterator), nil
		}
		return argMustBeType(in.callToken, "walkIter", "string")
	})
	filepathFunc("walkStrIter", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			callToken := in.callToken
			firstIter := true
			pathChan := make(chan string, 1)
			errorChan := make(chan error, 1)
			signalChan := make(chan struct{}, 1)
			doneChan := make(chan struct{}, 1)
			var currentPath string
			iterator := ProtoIteratorErr{legacyPanicOnErr: true}
			iterator.hasNextMethod = func() (bool, error) {
				if firstIter {
					firstIter = false
					go func() {
						filepath.WalkDir(
							loxStr.str,
							func(path string, _ fs.DirEntry, err error) error {
								<-signalChan
								if err != nil {
									errorChan <- err
									return err
								}
								pathChan <- path
								return nil
							},
						)
						doneChan <- struct{}{}
					}()
				}
				signalChan <- struct{}{}
				select {
				case err := <-errorChan:
					return false, loxerror.RuntimeError(
						callToken,
						err.Error(),
					)
				case currentPath = <-pathChan:
					return true, nil
				case <-doneChan:
					return false, nil
				}
			}
			iterator.nextMethod = func() (any, error) {
				return NewLoxStringQuote(currentPath), nil
			}
			return NewLoxIterator(iterator), nil
		}
		return argMustBeType(in.callToken, "walkStrIter", "string")
	})

	i.globals.Define(className, filepathClass)
}
