package ast

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxZIPWriter struct {
	writer      *zip.Writer
	bytesBuffer *bytes.Buffer
	isClosed    bool
	fileNames   map[string]struct{ isDir bool }
	methods     map[string]*struct{ ProtoLoxCallable }
}

func NewLoxZIPWriter(writer io.Writer) *LoxZIPWriter {
	return &LoxZIPWriter{
		writer:      zip.NewWriter(writer),
		bytesBuffer: nil,
		isClosed:    false,
		fileNames:   make(map[string]struct{ isDir bool }),
		methods:     make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxZIPWriterBytes(bytesBuffer *bytes.Buffer) *LoxZIPWriter {
	zipWriter := NewLoxZIPWriter(bytesBuffer)
	zipWriter.bytesBuffer = bytesBuffer
	return zipWriter
}

func (l *LoxZIPWriter) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	zipWriterFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native zip writer fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'zip writer.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	closedErr := func() (any, error) {
		return nil, loxerror.RuntimeError(name,
			fmt.Sprintf("Cannot call 'zip writer.%v' on closed zip writer objects.", methodName))
	}
	fileExistsErr := func(fileName string, fileNameStruct struct{ isDir bool }) (any, error) {
		var errStr string
		if fileNameStruct.isDir {
			errStr = fmt.Sprintf("zip writer.%v: '%v' already exists as a directory.", methodName, fileName)
		} else {
			errStr = fmt.Sprintf("zip writer.%v: '%v' already exists as a file.", methodName, fileName)
		}
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "addFile":
		return zipWriterFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"First argument to 'zip writer.addFile' must be a string.")
			}
			switch args[1].(type) {
			case *LoxBuffer:
			case *LoxFile:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'zip writer.addFile' must be a buffer, file, or string.")
			}

			if l.isClosed {
				return closedErr()
			}

			fileName := strings.Trim(args[0].(*LoxString).str, "/\\ ")
			var builder strings.Builder
			for _, c := range fileName {
				switch c {
				case '/', '\\':
					builder.WriteRune(os.PathSeparator)
				default:
					builder.WriteRune(c)
				}
			}
			fileName = builder.String()
			if fileNameStruct, ok := l.fileNames[fileName]; ok {
				return fileExistsErr(fileName, fileNameStruct)
			} else {
				l.fileNames[fileName] = struct{ isDir bool }{false}
			}

			var content []byte
			switch arg := args[1].(type) {
			case *LoxBuffer:
				content = make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					content = append(content, byte(element.(int64)))
				}
			case *LoxFile:
				if !arg.isRead() {
					delete(l.fileNames, fileName)
					return nil, loxerror.RuntimeError(name,
						"First file argument to 'zip writer.addFile' must be in read mode.")
				}
				var readErr error
				content, readErr = io.ReadAll(arg.file)
				if readErr != nil {
					delete(l.fileNames, fileName)
					return nil, loxerror.RuntimeError(name, readErr.Error())
				}
			case *LoxString:
				content = []byte(arg.str)
			}
			writer, err := l.writer.Create(fileName)
			if err != nil {
				delete(l.fileNames, fileName)
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			_, err = writer.Write(content)
			if err != nil {
				delete(l.fileNames, fileName)
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l, nil
		})
	case "buffer":
		return zipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.bytesBuffer == nil {
				return nil, loxerror.RuntimeError(name,
					"ZIP file is not being written to a buffer.")
			}
			bytes := l.bytesBuffer.Bytes()
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, b := range bytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "close":
		return zipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isClosed {
				err := l.writer.Close()
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				l.isClosed = true
			}
			return nil, nil
		})
	case "fileNames":
		return zipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			fileNamesLen := len(l.fileNames)
			if fileNamesLen == 0 {
				return EmptyLoxList(), nil
			}
			fileNames := make([]string, 0, fileNamesLen)
			for fileName := range l.fileNames {
				fileNames = append(fileNames, fileName)
			}
			slices.Sort(fileNames)
			fileNamesList := list.NewListCap[any](int64(fileNamesLen))
			for _, fileName := range fileNames {
				fileNamesList.Add(NewLoxStringQuote(fileName))
			}
			return NewLoxList(fileNamesList), nil
		})
	case "flush":
		return zipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.writer.Flush()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l, nil
		})
	case "isBuffer":
		return zipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.bytesBuffer != nil, nil
		})
	case "isClosed":
		return zipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isClosed, nil
		})
	case "mkdir":
		return zipWriterFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				if l.isClosed {
					return closedErr()
				}
				dirName := strings.Trim(loxStr.str, "/\\ ")
				var builder strings.Builder
				for _, c := range dirName {
					switch c {
					case '/', '\\':
						builder.WriteRune(os.PathSeparator)
					default:
						builder.WriteRune(c)
					}
				}
				dirName = builder.String()
				if fileNameStruct, ok := l.fileNames[dirName]; ok {
					return fileExistsErr(dirName, fileNameStruct)
				} else {
					l.fileNames[dirName] = struct{ isDir bool }{true}
				}
				_, err := l.writer.Create(dirName + "/")
				if err != nil {
					delete(l.fileNames, dirName)
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return l, nil
			}
			return argMustBeType("string")
		})
	case "printFileNames":
		return zipWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			fileNamesLen := len(l.fileNames)
			if fileNamesLen > 0 {
				fileNames := make([]string, 0, fileNamesLen)
				for fileName := range l.fileNames {
					fileNames = append(fileNames, fileName)
				}
				slices.Sort(fileNames)
				for _, fileName := range fileNames {
					fmt.Println(fileName)
				}
			}
			return nil, nil
		})
	case "setComment":
		return zipWriterFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				err := l.writer.SetComment(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return l, nil
			}
			return argMustBeType("string")
		})
	}
	return nil, loxerror.RuntimeError(name, "ZIP writers have no property called '"+methodName+"'.")
}

func (l *LoxZIPWriter) String() string {
	return fmt.Sprintf("<zip writer at %p>", l)
}

func (l *LoxZIPWriter) Type() string {
	return "zip writer"
}
