package ast

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/AlanLuu/lox/ast/filemode"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type LoxDirRoot struct {
	root     *os.Root
	isClosed bool
	chain    bool
	methods  map[string]*struct{ ProtoLoxCallable }
}

func NewLoxDirRoot(root *os.Root) *LoxDirRoot {
	return &LoxDirRoot{
		root:     root,
		isClosed: false,
		chain:    false,
		methods:  make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxDirRoot) close() {
	if !l.isClosed {
		l.root.Close()
		l.isClosed = true
	}
}

func (l *LoxDirRoot) nilOrChain() any {
	if l.chain {
		return l
	}
	return nil
}

func (l *LoxDirRoot) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if method, ok := l.methods[lexemeName]; ok {
		return method, nil
	}
	dirRootFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native dir root fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.methods[lexemeName]; !ok {
			l.methods[lexemeName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'dir root.%v' must be a %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch lexemeName {
	case "chdir":
		return dirRootFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := os.Chdir(l.root.Name())
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l.nilOrChain(), nil
		})
	case "chmod":
		return dirRootFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dir root.chmod' must be a string.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dir root.chmod' must be an integer.")
			}
			file := args[0].(*LoxString).str
			mode := args[1].(int64)
			err := l.root.Chmod(file, os.FileMode(mode))
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l.nilOrChain(), nil
		})
	case "chown":
		return dirRootFunc(3, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dir root.chown' must be a string.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dir root.chown' must be an integer.")
			}
			if _, ok := args[2].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Third argument to 'dir root.chown' must be an integer.")
			}
			if util.IsWindows() {
				return nil, loxerror.RuntimeError(name,
					"'dir root.chown' is unsupported on Windows.")
			}
			file := args[0].(*LoxString).str
			uid := int(args[1].(int64))
			gid := int(args[2].(int64))
			err := l.root.Chown(file, uid, gid)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l.nilOrChain(), nil
		})
	case "close":
		return dirRootFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.close()
			return l.nilOrChain(), nil
		})
	case "create":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				file, err := l.root.Create(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return &LoxFile{
					file:       file,
					name:       file.Name(),
					mode:       filemode.READ_WRITE,
					isBinary:   false,
					stat:       nil,
					properties: make(map[string]any),
				}, nil
			}
			return argMustBeType("string")
		})
	case "createBin":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				file, err := l.root.Create(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return &LoxFile{
					file:       file,
					name:       file.Name(),
					mode:       filemode.READ_WRITE,
					isBinary:   true,
					stat:       nil,
					properties: make(map[string]any),
				}, nil
			}
			return argMustBeType("string")
		})
	case "isClosed":
		return dirRootFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isClosed, nil
		})
	case "lchown":
		return dirRootFunc(3, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dir root.lchown' must be a string.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dir root.lchown' must be an integer.")
			}
			if _, ok := args[2].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Third argument to 'dir root.lchown' must be an integer.")
			}
			if util.IsWindows() {
				return nil, loxerror.RuntimeError(name,
					"'dir root.lchown' is unsupported on Windows.")
			}
			file := args[0].(*LoxString).str
			uid := int(args[1].(int64))
			gid := int(args[2].(int64))
			err := l.root.Lchown(file, uid, gid)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l.nilOrChain(), nil
		})
	case "link":
		return dirRootFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dir root.link' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dir root.link' must be a string.")
			}
			target := args[0].(*LoxString).str
			linkName := args[1].(*LoxString).str
			err := l.root.Link(target, linkName)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l.nilOrChain(), nil
		})
	case "mkdir":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				err := l.root.Mkdir(loxStr.str, 0777)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return l.nilOrChain(), nil
			}
			return argMustBeType("string")
		})
	case "mkdirp":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				err := l.root.MkdirAll(loxStr.str, 0777)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return l.nilOrChain(), nil
			}
			return argMustBeType("string")
		})
	case "name":
		return dirRootFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.root.Name()), nil
		})
	case "open":
		return dirRootFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dir root.open' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dir root.open' must be a string.")
			}
			path := args[0].(*LoxString).str
			mode := args[1].(*LoxString).str
			loxFile, err := NewLoxFileModeStr(path, mode, l.root)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return loxFile, nil
		})
	case "readFile":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				bytes, err := l.root.ReadFile(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return NewLoxStringQuote(string(bytes)), nil
			}
			return argMustBeType("string")
		})
	case "readFileBin":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				bytes, err := l.root.ReadFile(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				loxBuffer := EmptyLoxBufferCap(int64(len(bytes)))
				for _, element := range bytes {
					addErr := loxBuffer.add(int64(element))
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				}
				return loxBuffer, nil
			}
			return argMustBeType("string")
		})
	case "readLink":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				dest, err := l.root.Readlink(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return NewLoxStringQuote(dest), nil
			}
			return argMustBeType("string")
		})
	case "remove":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				err := l.root.Remove(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return l.nilOrChain(), nil
			}
			return argMustBeType("string")
		})
	case "removeAll":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				err := l.root.RemoveAll(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return l.nilOrChain(), nil
			}
			return argMustBeType("string")
		})
	case "rename":
		return dirRootFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dir root.rename' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dir root.rename' must be a string.")
			}
			oldPath := args[0].(*LoxString).str
			newPath := args[1].(*LoxString).str
			err := l.root.Rename(oldPath, newPath)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l.nilOrChain(), nil
		})
	case "rmrf":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				remove := func(path string) {
					if err := os.Remove(path); err != nil {
						fmt.Fprintln(os.Stderr, err.Error())
					}
				}
				isWindows := util.IsWindows()
				var dotRegex *regexp.Regexp
				if isWindows {
					dotRegex = regexp.MustCompile(`^[.][\/\\]*$`)
				} else {
					dotRegex = regexp.MustCompile(`^[.][\/]*$`)
				}
				dirs := list.NewList[string]()
				deletedEmptyDir := false
				dirFunc := func(path string, d fs.DirEntry, err error) error {
					if isWindows {
						path = strings.ReplaceAll(path, "/", "\\")
					}
					if err != nil {
						fmt.Fprintln(os.Stderr, err.Error())
					}
					if !dirs.IsEmpty() {
						if last := dirs.Peek(); !strings.HasPrefix(path, last) {
							remove(last)
							dirs.Pop()
						}
					}
					if d == nil || dotRegex.MatchString(path) {
						return nil
					}
					if d.IsDir() {
						if err := os.Remove(path); err != nil {
							var errCode syscall.Errno
							if isWindows {
								//Windows error code 145 is ERROR_DIR_NOT_EMPTY
								errCode = 145
							} else {
								errCode = syscall.ENOTEMPTY
							}
							var pathErr *os.PathError
							if errors.As(err, &pathErr) && pathErr.Err == errCode {
								dirs.Add(path)
							} else {
								fmt.Fprintln(os.Stderr, err.Error())
							}
						} else {
							deletedEmptyDir = true
						}
					} else {
						remove(path)
					}
					if deletedEmptyDir {
						deletedEmptyDir = false
						return filepath.SkipDir
					}
					return nil
				}
				if err := filepath.WalkDir(loxStr.str, dirFunc); err != nil {
					fmt.Fprintf(
						os.Stderr,
						"dir root.rmrf: %v\n",
						err.Error(),
					)
				}
				for i := len(dirs) - 1; i >= 0; i-- {
					remove(dirs[i])
				}
				dirs.Clear()
				return l.nilOrChain(), nil
			}
			return argMustBeType("string")
		})
	case "setChain":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if chain, ok := args[0].(bool); ok {
				l.chain = chain
				return l, nil
			}
			return argMustBeType("boolean")
		})
	case "stat":
		return dirRootFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := loxStr.str
				stat, statErr := l.root.Stat(str)
				if statErr != nil {
					return nil, loxerror.RuntimeError(name, statErr.Error())
				}
				return NewLoxFileInfoPathFileInfo(str, stat), nil
			}
			return argMustBeType("string")
		})
	case "symlink":
		return dirRootFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dir root.symlink' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dir root.symlink' must be a string.")
			}
			target := args[0].(*LoxString).str
			linkName := args[1].(*LoxString).str
			err := l.root.Symlink(target, linkName)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l.nilOrChain(), nil
		})
	case "writeFile":
		return dirRootFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dir root.writeFile' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dir root.writeFile' must be a string.")
			}
			fileName := args[0].(*LoxString).str
			data := args[1].(*LoxString).str
			err := l.root.WriteFile(fileName, []byte(data), 0666)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l.nilOrChain(), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Dir roots have no property called '"+lexemeName+"'.")
}

func (l *LoxDirRoot) String() string {
	if l.root == nil {
		return fmt.Sprintf("<nil dir root at %p>", l)
	}
	return fmt.Sprintf("<dir root name='%v' at %p>", l.root.Name(), l)
}

func (l *LoxDirRoot) Type() string {
	return "dir root"
}
