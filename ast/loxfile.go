package ast

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/AlanLuu/lox/ast/filemode"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
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
	case 3:
		if strings.Contains(modeStr, "r") &&
			strings.Contains(modeStr, "w") &&
			strings.Contains(modeStr, "b") {

			fileMode := filemode.READ_WRITE
			file, fileErr := filemode.Open(path, fileMode)
			if fileErr != nil {
				return nil, fileErr
			}
			return &LoxFile{
				file:       file,
				name:       file.Name(),
				mode:       fileMode,
				isBinary:   true,
				isClosed:   false,
				properties: make(map[string]any),
			}, nil
		}
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

func (l *LoxFile) isWrite() bool {
	return l.mode == filemode.WRITE || l.mode == filemode.READ_WRITE
}

func (l *LoxFile) isAppend() bool {
	return l.mode == filemode.APPEND
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
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'file.%v' must be a %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'file.%v' must be an %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch lexemeName {
	case "chdir":
		return fileFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.file.Chdir()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "chmod":
		return fileFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if mode, ok := args[0].(int64); ok {
				err := l.file.Chmod(os.FileMode(mode))
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return nil, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "close":
		return fileFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.close()
			return nil, nil
		})
	case "closed":
		return fileField(l.isClosed)
	case "fd":
		return int64(l.file.Fd()), nil
	case "flush":
		return fileFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			flushErr := l.file.Sync()
			if flushErr != nil {
				return nil, loxerror.RuntimeError(name, flushErr.Error())
			}
			return nil, nil
		})
	case "isClosed":
		return fileFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isClosed, nil
		})
	case "isDir":
		return fileFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			stat, statErr := l.file.Stat()
			if statErr != nil {
				return nil, loxerror.RuntimeError(name, statErr.Error())
			}
			return stat.IsDir(), nil
		})
	case "mode":
		binaryMode := ""
		if l.isBinary {
			binaryMode = "b"
		}
		return fileField(NewLoxString(filemode.ModeStrings[l.mode]+binaryMode, '\''))
	case "name":
		return fileField(NewLoxStringQuote(l.name))
	case "read":
		return fileFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if l.isClosed {
				return nil, loxerror.RuntimeError(name, "Cannot read from a closed file.")
			}
			if !l.isRead() {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'read' for file not in read mode.")
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
				if numBytes >= 0 {
					buffer = make([]byte, numBytes)
					bufferSize, bufferErr = io.ReadAtLeast(l.file, buffer, numBytes)
				} else {
					buffer, bufferErr = io.ReadAll(l.file)
				}
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
			if bufferErr != nil {
				switch {
				case argsLen == 1 && (errors.Is(bufferErr, io.ErrUnexpectedEOF) ||
					errors.Is(bufferErr, io.EOF)):
				default:
					return nil, loxerror.RuntimeError(name, bufferErr.Error())
				}
			}
			if l.isBinary {
				loxBuffer := EmptyLoxBuffer()
				if argsLen == 0 {
					for _, element := range buffer {
						addErr := loxBuffer.add(int64(element))
						if addErr != nil {
							return nil, loxerror.RuntimeError(name, addErr.Error())
						}
					}
				} else {
					for i := 0; i < bufferSize; i++ {
						addErr := loxBuffer.add(int64(buffer[i]))
						if addErr != nil {
							return nil, loxerror.RuntimeError(name, addErr.Error())
						}
					}
				}
				return loxBuffer, nil
			}
			return NewLoxStringQuote(string(buffer)), nil
		})
	case "readByte":
		return fileFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.isClosed {
				return nil, loxerror.RuntimeError(name, "Cannot read from a closed file.")
			}
			if !l.isRead() {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readByte' for file not in read mode.")
			}
			if !l.isBinary {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readByte' for file not in binary mode.")
			}
			b := make([]byte, 1)
			_, readErr := l.file.Read(b)
			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					return nil, nil
				}
				return nil, loxerror.RuntimeError(name, readErr.Error())
			}
			return int64(b[0]), nil
		})
	case "readChar":
		return fileFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.isClosed {
				return nil, loxerror.RuntimeError(name, "Cannot read from a closed file.")
			}
			if !l.isRead() {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readChar' for file not in read mode.")
			}
			if l.isBinary {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readChar' for file in binary mode.")
			}
			var b [4]byte
			for i := 0; i < len(b); i++ {
				_, readErr := l.file.Read(b[i : i+1])
				if readErr != nil {
					if errors.Is(readErr, io.EOF) {
						return nil, nil
					}
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
				if r, _ := utf8.DecodeRune(b[:i+1]); r != utf8.RuneError {
					if r == '\'' {
						return NewLoxString(string(r), '"'), nil
					}
					return NewLoxString(string(r), '\''), nil
				}
			}
			return nil, loxerror.RuntimeError(name,
				fmt.Sprintf("Invalid character encoding found with bytes '%v'.", b))
		})
	case "readLine":
		return fileFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.isClosed {
				return nil, loxerror.RuntimeError(name, "Cannot read from a closed file.")
			}
			if !l.isRead() {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readLine' for file not in read mode.")
			}
			if l.isBinary {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readLine' for file in binary mode.")
			}
			var quote byte = '\''
			var builder strings.Builder
			b := make([]byte, 1)
			_, readErr := l.file.Read(b)
			for readErr == nil && (b[0] == '\r' || b[0] == '\n') {
				for b[0] == '\r' && readErr == nil {
					_, readErr = l.file.Read(b)
					if readErr != nil {
						break
					}
					if b[0] != '\n' {
						builder.WriteByte('\r')
						break
					}
					_, readErr = l.file.Read(b)
				}
				for b[0] == '\n' && readErr == nil {
					_, readErr = l.file.Read(b)
				}
			}
			for b[0] != '\n' && readErr == nil {
				if quote == '\'' && b[0] == '\'' {
					quote = '"'
				}
				if b[0] == '\r' {
					_, readErr = l.file.Read(b)
					if b[0] != '\n' {
						builder.WriteByte('\r')
					}
				} else {
					builder.WriteByte(b[0])
					_, readErr = l.file.Read(b)
				}
			}
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return nil, loxerror.RuntimeError(name, readErr.Error())
			}
			return NewLoxString(builder.String(), quote), nil
		})
	case "readLines":
		return fileFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			numLines := -1
			argsLen := len(args)
			switch argsLen {
			case 0:
			case 1:
				if _, ok := args[0].(int64); !ok {
					return argMustBeTypeAn("integer")
				}
				numLines = int(args[0].(int64))
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
			if l.isClosed {
				return nil, loxerror.RuntimeError(name, "Cannot read from a closed file.")
			}
			if !l.isRead() {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readLines' for file not in read mode.")
			}
			if l.isBinary {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readLines' for file in binary mode.")
			}
			lines := list.NewList[any]()
			if numLines == 0 {
				return NewLoxList(lines), nil
			}
			var quote byte = '\''
			var builder strings.Builder
			b := make([]byte, 1)
			_, readErr := l.file.Read(b)
		outer:
			for readErr == nil && (numLines < 0 || len(lines) < numLines) {
				switch {
				case quote == '\'' && b[0] == '\'':
					quote = '"'
				case b[0] == '\r':
					_, readErr = l.file.Read(b)
					if b[0] != '\n' {
						builder.WriteByte('\r')
						continue
					} else if builder.Len() > 0 {
						lines.Add(NewLoxString(builder.String(), quote))
						builder.Reset()
					}
					quote = '\''
					if readErr != nil {
						break outer
					}
				case b[0] == '\n':
					if builder.Len() > 0 {
						lines.Add(NewLoxString(builder.String(), quote))
						builder.Reset()
					}
					quote = '\''
				default:
					builder.WriteByte(b[0])
				}
				if numLines < 0 || len(lines) < numLines {
					_, readErr = l.file.Read(b)
					if readErr != nil && builder.Len() > 0 {
						lines.Add(NewLoxString(builder.String(), quote))
					}
				}
			}
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				lines.Clear()
				return nil, loxerror.RuntimeError(name, readErr.Error())
			}
			return NewLoxList(lines), nil
		})
	case "readNewLine":
		return fileFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.isClosed {
				return nil, loxerror.RuntimeError(name, "Cannot read from a closed file.")
			}
			if !l.isRead() {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readNewLine' for file not in read mode.")
			}
			if l.isBinary {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readNewLine' for file in binary mode.")
			}
			var quote byte = '\''
			var builder strings.Builder
			b := make([]byte, 1)
			_, readErr := l.file.Read(b)
			for b[0] != '\n' && readErr == nil {
				if quote == '\'' && b[0] == '\'' {
					quote = '"'
				}
				builder.WriteByte(b[0])
				_, readErr = l.file.Read(b)
			}
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return nil, loxerror.RuntimeError(name, readErr.Error())
			}
			if b[0] == '\n' {
				builder.WriteByte('\n')
			}
			return NewLoxString(builder.String(), quote), nil
		})
	case "readNewLines":
		return fileFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			numLines := -1
			argsLen := len(args)
			switch argsLen {
			case 0:
			case 1:
				if _, ok := args[0].(int64); !ok {
					return argMustBeTypeAn("integer")
				}
				numLines = int(args[0].(int64))
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
			if l.isClosed {
				return nil, loxerror.RuntimeError(name, "Cannot read from a closed file.")
			}
			if !l.isRead() {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readNewLines' for file not in read mode.")
			}
			if l.isBinary {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'readNewLines' for file in binary mode.")
			}
			lines := list.NewList[any]()
			if numLines == 0 {
				return NewLoxList(lines), nil
			}
			var quote byte = '\''
			var builder strings.Builder
			b := make([]byte, 1)
			_, readErr := l.file.Read(b)
			for readErr == nil && (numLines < 0 || len(lines) < numLines) {
				switch {
				case quote == '\'' && b[0] == '\'':
					quote = '"'
				case b[0] == '\n':
					builder.WriteByte('\n')
					lines.Add(NewLoxString(builder.String(), quote))
					builder.Reset()
					quote = '\''
				default:
					builder.WriteByte(b[0])
				}
				if numLines < 0 || len(lines) < numLines {
					_, readErr = l.file.Read(b)
					if readErr != nil && builder.Len() > 0 {
						lines.Add(NewLoxString(builder.String(), quote))
					}
				}
			}
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				lines.Clear()
				return nil, loxerror.RuntimeError(name, readErr.Error())
			}
			return NewLoxList(lines), nil
		})
	case "seek":
		return fileFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'file.seek' must be an integer.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'file.seek' must be an integer.")
			}
			offset := args[0].(int64)
			whence := int(args[1].(int64))
			if whence < 0 || whence > 2 {
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Invalid whence value '%v' for 'file.seek'.", whence))
			}
			if l.isAppend() {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'seek' for file in append mode.")
			}
			position, seekErr := l.file.Seek(offset, whence)
			if seekErr != nil {
				return nil, loxerror.RuntimeError(name, seekErr.Error())
			}
			return position, nil
		})
	case "write":
		return fileFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if l.isClosed {
				return nil, loxerror.RuntimeError(name, "Cannot write to a closed file.")
			}
			if !l.isWrite() && !l.isAppend() {
				return nil, loxerror.RuntimeError(name,
					"Unsupported operation 'write' for file not in write or append mode.")
			}
			switch arg := args[0].(type) {
			case *LoxBuffer:
				if !l.isBinary {
					return argMustBeType("string")
				}
				byteList := list.NewList[byte]()
				for _, element := range arg.elements {
					byteList.Add(byte(element.(int64)))
				}
				numBytes, writeErr := l.file.Write([]byte(byteList))
				if writeErr != nil {
					return nil, loxerror.RuntimeError(name, writeErr.Error())
				}
				return int64(numBytes), nil
			case *LoxString:
				if l.isBinary {
					return argMustBeType("buffer")
				}
				numBytes, writeErr := l.file.WriteString(arg.str)
				if writeErr != nil {
					return nil, loxerror.RuntimeError(name, writeErr.Error())
				}
				return int64(numBytes), nil
			default:
				if l.isBinary {
					return argMustBeType("buffer")
				}
				return argMustBeType("string")
			}
		})
	case "writeByte":
		return fileFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if value, ok := args[0].(int64); ok {
				if l.isClosed {
					return nil, loxerror.RuntimeError(name, "Cannot write to a closed file.")
				}
				if !l.isWrite() && !l.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Unsupported operation 'writeByte' for file not in write or append mode.")
				}
				if !l.isBinary {
					return nil, loxerror.RuntimeError(name,
						"Unsupported operation 'writeByte' for file not in binary mode.")
				}
				if value < 0 || value > 255 {
					return nil, loxerror.RuntimeError(name,
						fmt.Sprintf("Invalid byte value '%v'.", value))
				}
				b := make([]byte, 1)
				b[0] = byte(value)
				_, writeErr := l.file.Write([]byte(b))
				if writeErr != nil {
					return nil, loxerror.RuntimeError(name, writeErr.Error())
				}
				return nil, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "writeLine":
		return fileFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				if l.isClosed {
					return nil, loxerror.RuntimeError(name, "Cannot write to a closed file.")
				}
				if !l.isWrite() && !l.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Unsupported operation 'writeLine' for file not in write or append mode.")
				}
				if l.isBinary {
					return nil, loxerror.RuntimeError(name,
						"Unsupported operation 'writeLine' for file in binary mode.")
				}
				var numBytes int
				var writeErr error
				if util.IsWindows() {
					numBytes, writeErr = l.file.WriteString(loxStr.str + "\r\n")
				} else {
					numBytes, writeErr = l.file.WriteString(loxStr.str + "\n")
				}
				if writeErr != nil {
					return nil, loxerror.RuntimeError(name, writeErr.Error())
				}
				return int64(numBytes), nil
			}
			return argMustBeType("string")
		})
	case "writeLines":
		return fileFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxList, ok := args[0].(*LoxList); ok {
				if l.isClosed {
					return nil, loxerror.RuntimeError(name, "Cannot write to a closed file.")
				}
				if !l.isWrite() && !l.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Unsupported operation 'writeLines' for file not in write or append mode.")
				}
				if l.isBinary {
					return nil, loxerror.RuntimeError(name,
						"Unsupported operation 'writeLines' for file in binary mode.")
				}
				numBytes := 0
				for _, element := range loxList.elements {
					var strToWrite string
					switch element := element.(type) {
					case *LoxString:
						strToWrite = element.str
					case fmt.Stringer:
						strToWrite = element.String()
					default:
						strToWrite = fmt.Sprint(element)
					}
					bytes, writeErr := l.file.WriteString(strToWrite)
					if writeErr != nil {
						return nil, loxerror.RuntimeError(name, writeErr.Error())
					}
					numBytes += bytes
				}
				return int64(numBytes), nil
			}
			return argMustBeType("list")
		})
	case "writeNewLines":
		return fileFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxList, ok := args[0].(*LoxList); ok {
				if l.isClosed {
					return nil, loxerror.RuntimeError(name, "Cannot write to a closed file.")
				}
				if !l.isWrite() && !l.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"Unsupported operation 'writeNewLines' for file not in write or append mode.")
				}
				if l.isBinary {
					return nil, loxerror.RuntimeError(name,
						"Unsupported operation 'writeNewLines' for file in binary mode.")
				}
				numBytes := 0
				for _, element := range loxList.elements {
					var strToWrite string
					switch element := element.(type) {
					case *LoxString:
						strToWrite = element.str
					case fmt.Stringer:
						strToWrite = element.String()
					default:
						strToWrite = fmt.Sprint(element)
					}
					var bytes int
					var writeErr error
					if util.IsWindows() {
						bytes, writeErr = l.file.WriteString(strToWrite + "\r\n")
					} else {
						bytes, writeErr = l.file.WriteString(strToWrite + "\n")
					}
					if writeErr != nil {
						return nil, loxerror.RuntimeError(name, writeErr.Error())
					}
					numBytes += bytes
				}
				return int64(numBytes), nil
			}
			return argMustBeType("list")
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
