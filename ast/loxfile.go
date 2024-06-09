package ast

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/AlanLuu/lox/ast/filemode"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxFile struct {
	file       *os.File
	name       string
	mode       filemode.FileMode
	isBinary   bool
	isClosed   bool
	properties map[string]any
}

func NewLoxFileModeStr(path string, modeStr string) (*LoxFile, error) {
	unknownMode := func() error {
		if len(modeStr) == 0 {
			return loxerror.Error("File mode cannot be blank.")
		}
		return loxerror.Error("Unknown file mode '" + modeStr + "'.")
	}
	switch len(modeStr) {
	case 1:
		if fileMode, ok := filemode.Modes[modeStr[0]]; ok {
			file, fileErr := filemode.Open(path, fileMode)
			if fileErr != nil {
				return nil, fileErr
			}
			return &LoxFile{
				file:       file,
				name:       file.Name(),
				mode:       fileMode,
				isBinary:   false,
				isClosed:   false,
				properties: make(map[string]any),
			}, nil
		}
	case 2:
		isBinary := false
		var fileMode filemode.FileMode
		firstLetter := modeStr[0]
		secondLetter := modeStr[1]
		fileMode1, ok1 := filemode.Modes[firstLetter]
		fileMode2, ok2 := filemode.Modes[secondLetter]
		switch {
		case ok1:
			switch {
			case secondLetter == 'b':
				isBinary = true
				fileMode = fileMode1
			case fileMode1 == filemode.READ && fileMode2 == filemode.WRITE:
				fallthrough
			case fileMode2 == filemode.READ && fileMode1 == filemode.WRITE:
				fileMode = filemode.READ_WRITE
			default:
				return nil, unknownMode()
			}
		case ok2:
			switch {
			case firstLetter == 'b':
				isBinary = true
				fileMode = fileMode2
			case fileMode1 == filemode.READ && fileMode2 == filemode.WRITE:
				fallthrough
			case fileMode2 == filemode.READ && fileMode1 == filemode.WRITE:
				fileMode = filemode.READ_WRITE
			default:
				return nil, unknownMode()
			}
		default:
			return nil, unknownMode()
		}

		file, fileErr := filemode.Open(path, fileMode)
		if fileErr != nil {
			return nil, fileErr
		}
		return &LoxFile{
			file:       file,
			name:       file.Name(),
			mode:       fileMode,
			isBinary:   isBinary,
			isClosed:   false,
			properties: make(map[string]any),
		}, nil
	}
	return nil, unknownMode()
}

func (l *LoxFile) close() {
	if !l.isClosed {
		l.file.Close()
		l.isClosed = true
	}
}

func (l *LoxFile) isRead() bool {
	return l.mode == filemode.READ || l.mode == filemode.READ_WRITE
}

func (l *LoxFile) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if field, ok := l.properties[lexemeName]; ok {
		return field, nil
	}
	fileField := func(field any) (any, error) {
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = field
		}
		return field, nil
	}
	fileFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native file fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = s
		}
		return s, nil
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'file.%v' must be an %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch lexemeName {
	case "close":
		return fileFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.close()
			return nil, nil
		})
	case "closed":
		return fileField(l.isClosed)
	case "isClosed":
		return fileFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isClosed, nil
		})
	case "name":
		return fileField(NewLoxStringQuote(l.name))
	case "read":
		return fileFunc(-1, func(in *Interpreter, args list.List[any]) (any, error) {
			if l.isClosed {
				return nil, loxerror.RuntimeError(in.callToken, "Cannot read from a closed file.")
			}
			if !l.isRead() {
				return nil, loxerror.RuntimeError(in.callToken, "Unsupported operation 'read'.")
			}
			var buffer []byte
			var bufferErr error
			var bufferSize int
			argsLen := len(args)
			switch argsLen {
			case 0:
				buffer, bufferErr = io.ReadAll(l.file)
			case 1:
				if _, ok := args[0].(int64); !ok {
					return argMustBeTypeAn("integer")
				}
				numBytes := int(args[0].(int64))
				buffer = make([]byte, numBytes)
				bufferSize, bufferErr = io.ReadAtLeast(l.file, buffer, numBytes)
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
			if bufferErr != nil {
				switch {
				case argsLen == 1 && (errors.Is(bufferErr, io.ErrUnexpectedEOF) ||
					errors.Is(bufferErr, io.EOF)):
				default:
					return nil, loxerror.RuntimeError(in.callToken, bufferErr.Error())
				}
			}
			if l.isBinary {
				loxBuffer := EmptyLoxBuffer()
				if argsLen == 0 {
					for _, element := range buffer {
						addErr := loxBuffer.add(int64(element))
						if addErr != nil {
							return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
						}
					}
				} else {
					for i := 0; i < bufferSize; i++ {
						addErr := loxBuffer.add(int64(buffer[i]))
						if addErr != nil {
							return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
						}
					}
				}
				return loxBuffer, nil
			}
			return NewLoxStringQuote(string(buffer)), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Files have no property called '"+lexemeName+"'.")
}

func (l *LoxFile) String() string {
	binaryMode := ""
	if l.isBinary {
		binaryMode = "b"
	}
	return fmt.Sprintf("<file name='%v' mode='%v%v' at %p>",
		l.name, filemode.ModeStrings[l.mode], binaryMode, l)
}

func (l *LoxFile) Type() string {
	return "file"
}
