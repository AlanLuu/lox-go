package ast

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/syscalls"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type LoxTarWriter struct {
	writer      *tar.Writer
	bytesBuffer *bytes.Buffer
	isClosed    bool
	fileNames   map[string]struct{ isDir bool }
	methods     map[string]*struct{ ProtoLoxCallable }
}

func NewLoxTarWriter(writer io.Writer) *LoxTarWriter {
	return &LoxTarWriter{
		writer:      tar.NewWriter(writer),
		bytesBuffer: nil,
		isClosed:    false,
		fileNames:   make(map[string]struct{ isDir bool }),
		methods:     make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxTarWriterBytes(bytesBuffer *bytes.Buffer) *LoxTarWriter {
	tarWriter := NewLoxTarWriter(bytesBuffer)
	tarWriter.bytesBuffer = bytesBuffer
	return tarWriter
}

func (l *LoxTarWriter) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	tarWriterFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native tar writer fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'tar writer.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	closedErr := func() (any, error) {
		return nil, loxerror.RuntimeError(name,
			fmt.Sprintf("Cannot call 'tar writer.%v' on closed tar writer objects.", methodName))
	}
	fileExistsErr := func(fileName string, fileNameStruct struct{ isDir bool }) (any, error) {
		var errStr string
		if fileNameStruct.isDir {
			errStr = fmt.Sprintf("tar writer.%v: '%v' already exists as a directory.", methodName, fileName)
		} else {
			errStr = fmt.Sprintf("tar writer.%v: '%v' already exists as a file.", methodName, fileName)
		}
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "addFile":
		return tarWriterFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 1:
				oneArgMsg := "When passing 1 argument, "
				if loxFile, ok := args[0].(*LoxFile); ok {
					if l.isClosed {
						return closedErr()
					}
					if !loxFile.isRead() {
						return nil, loxerror.RuntimeError(name,
							oneArgMsg+"file argument to 'tar writer.addFile' must be in read mode.")
					}

					file := loxFile.file
					stat, statErr := file.Stat()
					if statErr != nil {
						return nil, loxerror.RuntimeError(name, statErr.Error())
					}

					if stat.IsDir() {
						return nil, loxerror.RuntimeError(name,
							oneArgMsg+"cannot call 'tar writer.addFile' on a directory.")
					}

					content, readErr := io.ReadAll(file)
					if readErr != nil {
						return nil, loxerror.RuntimeError(name, readErr.Error())
					}

					statName := strings.Trim(stat.Name(), "/\\ ")
					statName = strings.ReplaceAll(statName, "\\", "/")
					if fileNameStruct, ok := l.fileNames[statName]; ok {
						return fileExistsErr(statName, fileNameStruct)
					} else {
						l.fileNames[statName] = struct{ isDir bool }{false}
					}

					writeHeaderErr := l.writer.WriteHeader(&tar.Header{
						Name:     statName,
						Mode:     int64(stat.Mode().Perm()),
						ModTime:  stat.ModTime(),
						Size:     int64(len(content)),
						Typeflag: tar.TypeReg,
					})
					if writeHeaderErr != nil {
						delete(l.fileNames, statName)
						return nil, loxerror.RuntimeError(name, writeHeaderErr.Error())
					}
					_, writeErr := l.writer.Write(content)
					if writeErr != nil {
						delete(l.fileNames, statName)
						return nil, loxerror.RuntimeError(name, writeErr.Error())
					}
				} else {
					return nil, loxerror.RuntimeError(name,
						oneArgMsg+"argument to 'tar writer.addFile' must be a file.")
				}
			case 2:
				twoArgMsg := "When passing 2 arguments, "
				switch args[0].(type) {
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name,
						twoArgMsg+"first argument to 'tar writer.addFile' must be a string.")
				}
				switch args[1].(type) {
				case *LoxBuffer:
				case *LoxFile:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(name,
						twoArgMsg+"second argument to 'tar writer.addFile' must be a buffer, file, or string.")
				}

				if l.isClosed {
					return closedErr()
				}

				fileName := strings.Trim(args[0].(*LoxString).str, "/\\ ")
				fileName = strings.ReplaceAll(fileName, "\\", "/")
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
							twoArgMsg+"first file argument to 'tar writer.addFile' must be in read mode.")
					}
					stat, statErr := arg.file.Stat()
					if statErr != nil {
						delete(l.fileNames, fileName)
						return nil, loxerror.RuntimeError(name, statErr.Error())
					}
					if stat.IsDir() {
						delete(l.fileNames, fileName)
						return nil, loxerror.RuntimeError(name,
							twoArgMsg+"cannot call 'tar writer.addFile' on a directory.")
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

				var headerMode int64
				if !util.IsWindows() {
					umask := syscalls.Umask(0)
					syscalls.Umask(umask)
					headerMode = int64(0666 & ^umask)
				} else {
					headerMode = 0644
				}

				writeHeaderErr := l.writer.WriteHeader(&tar.Header{
					Name:     fileName,
					Mode:     headerMode,
					ModTime:  time.Now(),
					Size:     int64(len(content)),
					Typeflag: tar.TypeReg,
				})
				if writeHeaderErr != nil {
					delete(l.fileNames, fileName)
					return nil, loxerror.RuntimeError(name, writeHeaderErr.Error())
				}
				_, writeErr := l.writer.Write(content)
				if writeErr != nil {
					delete(l.fileNames, fileName)
					return nil, loxerror.RuntimeError(name, writeErr.Error())
				}
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
			return l, nil
		})
	case "buffer":
		return tarWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.bytesBuffer == nil {
				return nil, loxerror.RuntimeError(name,
					"Tar file is not being written to a buffer.")
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
		return tarWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
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
		return tarWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
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
		return tarWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.writer.Flush()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l, nil
		})
	case "isBuffer":
		return tarWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.bytesBuffer != nil, nil
		})
	case "isClosed":
		return tarWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isClosed, nil
		})
	case "mkdir":
		return tarWriterFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				if l.isClosed {
					return closedErr()
				}
				dirName := strings.Trim(loxStr.str, "/\\ ")
				dirName = strings.ReplaceAll(dirName, "\\", "/")
				if fileNameStruct, ok := l.fileNames[dirName]; ok {
					return fileExistsErr(dirName, fileNameStruct)
				} else {
					l.fileNames[dirName] = struct{ isDir bool }{true}
				}
				var headerMode int64
				if !util.IsWindows() {
					umask := syscalls.Umask(0)
					syscalls.Umask(umask)
					headerMode = int64(0777 & ^umask)
				} else {
					headerMode = 0755
				}
				writeHeaderErr := l.writer.WriteHeader(&tar.Header{
					Name:     dirName,
					Mode:     headerMode,
					ModTime:  time.Now(),
					Typeflag: tar.TypeDir,
				})
				if writeHeaderErr != nil {
					delete(l.fileNames, dirName)
					return nil, loxerror.RuntimeError(name, writeHeaderErr.Error())
				}
				return l, nil
			}
			return argMustBeType("string")
		})
	case "printFileNames":
		return tarWriterFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
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
	}
	return nil, loxerror.RuntimeError(name, "Tar writers have no property called '"+methodName+"'.")
}

func (l *LoxTarWriter) String() string {
	return fmt.Sprintf("<tar writer at %p>", l)
}

func (l *LoxTarWriter) Type() string {
	return "tar writer"
}
