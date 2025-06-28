package ast

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxBufScanner struct {
	scanner         *bufio.Scanner
	startedScanning bool
	origReader      io.Reader
	readerType      int
	buffer          []byte
	methods         map[string]*struct{ ProtoLoxCallable }
}

func NewLoxBufScanner(reader io.Reader) *LoxBufScanner {
	return NewLoxBufScannerSplitFunc(reader, nil)
}

func NewLoxBufScannerSplitFunc(reader io.Reader, split bufio.SplitFunc) *LoxBufScanner {
	if reader == nil {
		panic("in NewLoxBufScannerSplitFunc: reader is nil")
	}
	bufScanner := &LoxBufScanner{
		scanner:    bufio.NewScanner(reader),
		origReader: reader,
		methods:    make(map[string]*struct{ ProtoLoxCallable }),
	}
	if split != nil {
		bufScanner.scanner.Split(split)
	}
	return bufScanner
}

func NewLoxBufScannerType(reader io.Reader, readerType int) *LoxBufScanner {
	return NewLoxBufScannerSplitFuncType(reader, nil, readerType)
}

func NewLoxBufScannerSplitFuncType(
	reader io.Reader,
	split bufio.SplitFunc,
	readerType int,
) *LoxBufScanner {
	if reader == nil {
		panic("in NewLoxBufScannerSplitFuncType: reader is nil")
	}
	bufScanner := &LoxBufScanner{
		scanner:    bufio.NewScanner(reader),
		origReader: reader,
		readerType: readerType,
		methods:    make(map[string]*struct{ ProtoLoxCallable }),
	}
	if split != nil {
		bufScanner.scanner.Split(split)
	}
	return bufScanner
}

func NewLoxBufScannerStdin() *LoxBufScanner {
	return NewLoxBufScanner(os.Stdin)
}

func NewLoxBufScannerStdinSplitFunc(split bufio.SplitFunc) *LoxBufScanner {
	return NewLoxBufScannerSplitFunc(os.Stdin, split)
}

func (l *LoxBufScanner) setBuffer(size int) {
	if l.startedScanning || size < 0 || (l.buffer != nil && len(l.buffer) == size) {
		return
	}
	if l.buffer == nil {
		l.buffer = make([]byte, size)
		l.scanner.Buffer(l.buffer, size)
		return
	}
	if size <= len(l.buffer) || cap(l.buffer) >= size {
		l.buffer = l.buffer[:size]
	} else {
		l.buffer = l.buffer[:cap(l.buffer)]
		for cap(l.buffer) < size {
			l.buffer = append(l.buffer, 0)
			l.buffer = l.buffer[:cap(l.buffer)]
		}
		l.buffer = l.buffer[:size]
	}
	l.scanner.Buffer(l.buffer, size)
}

func (l *LoxBufScanner) scan() bool {
	if !l.startedScanning {
		l.startedScanning = true
	}
	return l.scanner.Scan()
}

func (l *LoxBufScanner) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	bufScannerFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native buffered scanner fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'buffered scanner.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "buffer":
		return bufScannerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if l.startedScanning {
					return nil, loxerror.RuntimeError(name,
						"Cannot call 'buffered scanner.buffer' when scanning has started.")
				}
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'buffered scanner.buffer' cannot be negative.")
				}
				l.setBuffer(int(num))
				return nil, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "bytes":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			bytes := l.scanner.Bytes()
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, b := range bytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "err":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.scanner.Err(); err != nil {
				return NewLoxError(err), nil
			}
			return nil, nil
		})
	case "errThrow":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.scanner.Err(); err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "scan":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.scan(), nil
		})
	case "scanBytes":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.scan() {
				return nil, nil
			}
			bytes := l.scanner.Bytes()
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, b := range bytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "scanBytesErr":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.scan() {
				if err := l.scanner.Err(); err != nil {
					return NewLoxError(err), nil
				}
				return nil, nil
			}
			bytes := l.scanner.Bytes()
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, b := range bytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "scanBytesErrThrow":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.scan() {
				if err := l.scanner.Err(); err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return nil, nil
			}
			bytes := l.scanner.Bytes()
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, b := range bytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "scanBytesIter":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				return l.scan()
			}
			iterator.nextMethod = func() any {
				bytes := l.scanner.Bytes()
				buffer := EmptyLoxBufferCap(int64(len(bytes)))
				for _, b := range bytes {
					addErr := buffer.add(int64(b))
					if addErr != nil {
						panic(addErr)
					}
				}
				return buffer
			}
			return NewLoxIterator(iterator), nil
		})
	case "scanErr":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.scan() {
				if err := l.scanner.Err(); err != nil {
					return NewLoxError(err), nil
				}
			}
			return nil, nil
		})
	case "scanErrThrow":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.scan() {
				if err := l.scanner.Err(); err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
			}
			return nil, nil
		})
	case "scanText":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.scan() {
				return nil, nil
			}
			return NewLoxStringQuote(l.scanner.Text()), nil
		})
	case "scanTextErr":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.scan() {
				if err := l.scanner.Err(); err != nil {
					return NewLoxError(err), nil
				}
				return nil, nil
			}
			return NewLoxStringQuote(l.scanner.Text()), nil
		})
	case "scanTextErrThrow":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.scan() {
				if err := l.scanner.Err(); err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return nil, nil
			}
			return NewLoxStringQuote(l.scanner.Text()), nil
		})
	case "scanTextIter":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				return l.scan()
			}
			iterator.nextMethod = func() any {
				return NewLoxStringQuote(l.scanner.Text())
			}
			return NewLoxIterator(iterator), nil
		})
	case "splitFunc":
		return bufScannerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				splitFunc, ok := splitFuncsMap[num]
				if !ok {
					return nil, loxerror.RuntimeError(name,
						"Invalid integer argument to 'buffered scanner.splitFunc'.")
				}
				l.scanner.Split(splitFunc)
				return nil, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "startedScanning":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.startedScanning, nil
		})
	case "text":
		return bufScannerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.scanner.Text()), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Buffered scanners have no property called '"+methodName+"'.")
}

func (l *LoxBufScanner) String() string {
	switch origReader := l.origReader.(type) {
	case *os.File:
		switch origReader {
		case os.Stdin:
			return fmt.Sprintf("<buffered stdin scanner at %p>", l)
		case os.Stdout:
			return fmt.Sprintf("<buffered stdout scanner at %p>", l)
		case os.Stderr:
			return fmt.Sprintf("<buffered stderr scanner at %p>", l)
		}
	case net.Conn:
		return fmt.Sprintf("<buffered connection scanner at %p>", l)
	case *strings.Reader:
		return fmt.Sprintf("<buffered string scanner at %p>", l)
	}

	switch l.readerType {
	case loxBufReaderHTTPRes:
		return fmt.Sprintf("<buffered HTTP response scanner at %p>", l)
	}

	return fmt.Sprintf("<buffered scanner at %p>", l)
}

func (l *LoxBufScanner) Type() string {
	return "buffered scanner"
}
