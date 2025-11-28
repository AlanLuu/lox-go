package ast

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type loxFileInfo_name string

func (l loxFileInfo_name) Name() string {
	return string(l)
}

func (l loxFileInfo_name) IsDir() bool {
	stat, statErr := os.Lstat(string(l))
	if statErr != nil {
		return false
	}
	return stat.IsDir()
}

type LoxFileInfo struct {
	path     loxFileInfo_name
	dirEntry fs.DirEntry
	fileInfo fs.FileInfo
	methods  map[string]*struct{ ProtoLoxCallable }
}

func NewLoxFileInfo(
	path string,
	dirEntry fs.DirEntry,
	fileInfo fs.FileInfo,
) *LoxFileInfo {
	return &LoxFileInfo{
		path:     loxFileInfo_name(path),
		dirEntry: dirEntry,
		fileInfo: fileInfo,
		methods:  make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxFileInfoPathDirEntry(
	path string,
	dirEntry fs.DirEntry,
) *LoxFileInfo {
	return NewLoxFileInfo(path, dirEntry, nil)
}

func NewLoxFileInfoPathFileInfo(
	path string,
	fileInfo fs.FileInfo,
) *LoxFileInfo {
	return NewLoxFileInfo(path, nil, fileInfo)
}

func (l *LoxFileInfo) commonFuncs() interface {
	Name() string
	IsDir() bool
} {
	switch {
	case l.dirEntry != nil:
		return l.dirEntry
	case l.fileInfo != nil:
		return l.fileInfo
	default:
		return l.path
	}
}

func (l *LoxFileInfo) hasNil() bool {
	return l.dirEntry == nil && l.fileInfo == nil
}

func (l *LoxFileInfo) initFileInfoField() error {
	if l.fileInfo != nil {
		return nil
	}
	var err error
	if l.dirEntry == nil {
		l.fileInfo, err = os.Lstat(string(l.path))
	} else {
		l.fileInfo, err = l.dirEntry.Info()
	}
	if err != nil {
		l.fileInfo = nil
		return loxerror.Error(
			fmt.Sprintf(
				"error when calling method: %v",
				err.Error(),
			),
		)
	}
	return nil
}

func (l *LoxFileInfo) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if field, ok := l.methods[methodName]; ok {
		return field, nil
	}
	fileInfoFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native file info object fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	switch methodName {
	case "fileMode":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.fileMode: %v",
						err.Error(),
					),
				)
			}
			return int64(l.fileInfo.Mode()), nil
		})
	case "hasNil":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.hasNil(), nil
		})
	case "isAppendOnly":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.isAppendOnly: %v",
						err.Error(),
					),
				)
			}
			return l.fileInfo.Mode()&fs.ModeAppend != 0, nil
		})
	case "isCharDevice":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.isCharDevice: %v",
						err.Error(),
					),
				)
			}
			return l.fileInfo.Mode()&fs.ModeCharDevice != 0, nil
		})
	case "isDevice":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.isDevice: %v",
						err.Error(),
					),
				)
			}
			return l.fileInfo.Mode()&fs.ModeDevice != 0, nil
		})
	case "isDir":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.hasNil() {
				if err := l.initFileInfoField(); err != nil {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"fileinfo.isDir: %v",
							err.Error(),
						),
					)
				}
			}
			return l.commonFuncs().IsDir(), nil
		})
	case "isNamedPipe":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.isNamedPipe: %v",
						err.Error(),
					),
				)
			}
			return l.fileInfo.Mode()&fs.ModeNamedPipe != 0, nil
		})
	case "isRegular":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.isRegular: %v",
						err.Error(),
					),
				)
			}
			return l.fileInfo.Mode().IsRegular(), nil
		})
	case "isSetgid":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.isSetgid: %v",
						err.Error(),
					),
				)
			}
			return l.fileInfo.Mode()&fs.ModeSetgid != 0, nil
		})
	case "isSetuid":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.isSetuid: %v",
						err.Error(),
					),
				)
			}
			return l.fileInfo.Mode()&fs.ModeSetuid != 0, nil
		})
	case "isSocket":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.isSocket: %v",
						err.Error(),
					),
				)
			}
			return l.fileInfo.Mode()&fs.ModeSocket != 0, nil
		})
	case "isSticky":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.isSticky: %v",
						err.Error(),
					),
				)
			}
			return l.fileInfo.Mode()&fs.ModeSticky != 0, nil
		})
	case "isSymlink":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.isSymlink: %v",
						err.Error(),
					),
				)
			}
			return l.fileInfo.Mode()&fs.ModeSymlink != 0, nil
		})
	case "modeStr":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.modeStr: %v",
						err.Error(),
					),
				)
			}
			return NewLoxStringQuote(l.fileInfo.Mode().String()), nil
		})
	case "modTime":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.modTime: %v",
						err.Error(),
					),
				)
			}
			return NewLoxDate(l.fileInfo.ModTime()), nil
		})
	case "name":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.commonFuncs().Name()), nil
		})
	case "path", "fullName":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(string(l.path)), nil
		})
	case "size":
		return fileInfoFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.initFileInfoField(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"fileinfo.size: %v",
						err.Error(),
					),
				)
			}
			return l.fileInfo.Size(), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "File info objects have no property called '"+methodName+"'.")
}

func (l *LoxFileInfo) String() string {
	if l.hasNil() {
		return fmt.Sprintf(
			"<nil file info name='%v' at %p>",
			string(l.path),
			l,
		)
	}
	return fmt.Sprintf(
		"<file info name='%v' at %p>",
		string(l.path),
		l,
	)
}

func (l *LoxFileInfo) Type() string {
	return "file info"
}
