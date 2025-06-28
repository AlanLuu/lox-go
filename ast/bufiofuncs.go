package ast

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

var splitFuncsMap = map[int64]bufio.SplitFunc{
	0: bufio.ScanLines,
	1: bufio.ScanBytes,
	2: bufio.ScanRunes,
	3: bufio.ScanWords,
}

func defineBufioFields(bufioClass *LoxClass) {
	splitFuncs := [...]string{
		"scanLines",
		"scanBytes",
		"scanChars",
		"scanWords",
	}
	for i, key := range splitFuncs {
		bufioClass.classProperties[key] = int64(i)
	}
	bufioClass.classProperties["maxScanTokenSize"] =
		int64(bufio.MaxScanTokenSize)
}

func (i *Interpreter) defineBufioFuncs() {
	className := "bufio"
	bufioClass := NewLoxClass(className, nil, false)
	bufioFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native bufio class fn %v at %p>", name, &s)
		}
		bufioClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bufio.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bufio.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	defineBufioFields(bufioClass)
	bufioFunc("reader", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxFile, ok := args[0].(*LoxFile); ok {
			if !loxFile.isRead() {
				return nil, loxerror.RuntimeError(in.callToken,
					"File argument to 'bufio.reader' must be in read mode.")
			}
			return NewLoxBufReader(loxFile.file), nil
		}
		return argMustBeType(in.callToken, "reader", "file")
	})
	bufioFunc("readerConn", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxConn, ok := args[0].(*LoxConnection); ok {
			return NewLoxBufReader(loxConn.conn), nil
		}
		return argMustBeType(in.callToken, "readerConn", "connection object")
	})
	bufioFunc("readerConnSize", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxConnection); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bufio.readerConnSize' must be a connection object.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bufio.readerConnSize' must be an integer.")
		}
		size := args[1].(int64)
		if size < 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Integer argument to 'bufio.readerConnSize' cannot be negative.")
		}
		loxConn := args[0].(*LoxConnection)
		return NewLoxBufReaderSize(loxConn.conn, int(size)), nil
	})
	bufioFunc("readerHTTPResponse", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxRes, ok := args[0].(*LoxHTTPResponse); ok {
			return NewLoxBufReaderType(loxRes.res.Body, loxBufReaderHTTPRes), nil
		}
		return argMustBeType(in.callToken, "readerHTTPResponse", "HTTP response object")
	})
	bufioFunc("readerHTTPResponseSize", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxHTTPResponse); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bufio.readerHTTPResponseSize' must be an HTTP response object.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bufio.readerHTTPResponseSize' must be an integer.")
		}
		size := args[1].(int64)
		if size < 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Integer argument to 'bufio.readerHTTPResponseSize' cannot be negative.")
		}
		loxRes := args[0].(*LoxHTTPResponse)
		return NewLoxBufReaderSizeType(
			loxRes.res.Body,
			int(size),
			loxBufReaderHTTPRes,
		), nil
	})
	bufioFunc("readerSize", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bufio.readerSize' must be a file.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bufio.readerSize' must be an integer.")
		}
		loxFile := args[0].(*LoxFile)
		if !loxFile.isRead() {
			return nil, loxerror.RuntimeError(in.callToken,
				"File argument to 'bufio.readerSize' must be in read mode.")
		}
		size := args[1].(int64)
		if size < 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Integer argument to 'bufio.readerSize' cannot be negative.")
		}
		return NewLoxBufReaderSize(loxFile.file, int(size)), nil
	})
	bufioFunc("readerStdin", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxBufReaderStdin(), nil
	})
	bufioFunc("readerStdinSize", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if size, ok := args[0].(int64); ok {
			if size < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'bufio.readerStdinSize' cannot be negative.")
			}
			return NewLoxBufReaderStdinSize(int(size)), nil
		}
		return argMustBeTypeAn(in.callToken, "readerStdinSize", "integer")
	})
	bufioFunc("readerString", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxBufReader(strings.NewReader(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "readerString", "string")
	})
	bufioFunc("readerStringSize", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bufio.readerStringSize' must be a string.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bufio.readerStringSize' must be an integer.")
		}
		size := args[1].(int64)
		if size < 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Integer argument to 'bufio.readerStringSize' cannot be negative.")
		}
		loxStr := args[0].(*LoxString)
		return NewLoxBufReaderSize(
			strings.NewReader(loxStr.str),
			int(size),
		), nil
	})
	bufioFunc("scanner", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxFile, ok := args[0].(*LoxFile); ok {
			if !loxFile.isRead() {
				return nil, loxerror.RuntimeError(in.callToken,
					"File argument to 'bufio.scanner' must be in read mode.")
			}
			return NewLoxBufScanner(loxFile.file), nil
		}
		return argMustBeType(in.callToken, "scanner", "file")
	})
	bufioFunc("scannerConn", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxConn, ok := args[0].(*LoxConnection); ok {
			return NewLoxBufScanner(loxConn.conn), nil
		}
		return argMustBeType(in.callToken, "scannerConn", "connection object")
	})
	bufioFunc("scannerConnFunc", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxConnection); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bufio.scannerConnFunc' must be a connection object.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bufio.scannerConnFunc' must be an integer.")
		}
		funcTypeNum := args[1].(int64)
		splitFunc, ok := splitFuncsMap[funcTypeNum]
		if !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Invalid integer argument to 'bufio.scannerConnFunc'.")
		}
		loxConn := args[0].(*LoxConnection)
		return NewLoxBufScannerSplitFunc(loxConn.conn, splitFunc), nil
	})
	bufioFunc("scannerFunc", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bufio.scannerFunc' must be a file.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bufio.scannerFunc' must be an integer.")
		}
		loxFile := args[0].(*LoxFile)
		if !loxFile.isRead() {
			return nil, loxerror.RuntimeError(in.callToken,
				"File argument to 'bufio.scannerFunc' must be in read mode.")
		}
		funcTypeNum := args[1].(int64)
		splitFunc, ok := splitFuncsMap[funcTypeNum]
		if !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Invalid integer argument to 'bufio.scannerFunc'.")
		}
		return NewLoxBufScannerSplitFunc(loxFile.file, splitFunc), nil
	})
	bufioFunc("scannerHTTPResponse", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxRes, ok := args[0].(*LoxHTTPResponse); ok {
			return NewLoxBufScannerType(loxRes.res.Body, loxBufReaderHTTPRes), nil
		}
		return argMustBeType(in.callToken, "scannerHTTPResponse", "HTTP response object")
	})
	bufioFunc("scannerHTTPResponseFunc", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxHTTPResponse); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bufio.scannerHTTPResponseFunc' must be an HTTP response object.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bufio.scannerHTTPResponseFunc' must be an integer.")
		}
		funcTypeNum := args[1].(int64)
		splitFunc, ok := splitFuncsMap[funcTypeNum]
		if !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Invalid integer argument to 'bufio.scannerHTTPResponseFunc'.")
		}
		loxRes := args[0].(*LoxHTTPResponse)
		return NewLoxBufScannerSplitFuncType(
			loxRes.res.Body,
			splitFunc,
			loxBufReaderHTTPRes,
		), nil
	})
	bufioFunc("scannerStdin", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxBufScannerStdin(), nil
	})
	bufioFunc("scannerStdinFunc", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			splitFunc, ok := splitFuncsMap[num]
			if !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Invalid integer argument to 'bufio.scannerStdinFunc'.")
			}
			return NewLoxBufScannerStdinSplitFunc(splitFunc), nil
		}
		return argMustBeTypeAn(in.callToken, "scannerStdinFunc", "integer")
	})
	bufioFunc("scannerString", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxBufScanner(strings.NewReader(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "scannerString", "string")
	})
	bufioFunc("scannerStringFunc", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bufio.scannerStringFunc' must be a string.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bufio.scannerStringFunc' must be an integer.")
		}
		funcTypeNum := args[1].(int64)
		splitFunc, ok := splitFuncsMap[funcTypeNum]
		if !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Invalid integer argument to 'bufio.scannerStringFunc'.")
		}
		loxStr := args[0].(*LoxString)
		return NewLoxBufScannerSplitFunc(
			strings.NewReader(loxStr.str),
			splitFunc,
		), nil
	})
	bufioFunc("writer", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxFile, ok := args[0].(*LoxFile); ok {
			if !loxFile.isWrite() && !loxFile.isAppend() {
				return nil, loxerror.RuntimeError(in.callToken,
					"File argument to 'bufio.writer' must be in write or append mode.")
			}
			return NewLoxBufWriter(loxFile.file), nil
		}
		return argMustBeType(in.callToken, "writer", "file")
	})
	bufioFunc("writerConn", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxConn, ok := args[0].(*LoxConnection); ok {
			return NewLoxBufWriter(loxConn.conn), nil
		}
		return argMustBeType(in.callToken, "writerConn", "connection object")
	})
	bufioFunc("writerConnSize", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxConnection); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bufio.writerConnSize' must be a connection object.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bufio.writerConnSize' must be an integer.")
		}
		size := args[1].(int64)
		if size < 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Integer argument to 'bufio.writerConnSize' cannot be negative.")
		}
		loxConn := args[0].(*LoxConnection)
		return NewLoxBufWriterSize(loxConn.conn, int(size)), nil
	})
	bufioFunc("writerSize", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bufio.writerSize' must be a file.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bufio.writerSize' must be an integer.")
		}
		loxFile := args[0].(*LoxFile)
		if !loxFile.isWrite() && !loxFile.isAppend() {
			return nil, loxerror.RuntimeError(in.callToken,
				"File argument to 'bufio.writerSize' must be in write or append mode.")
		}
		size := args[1].(int64)
		if size < 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Integer argument to 'bufio.writerSize' cannot be negative.")
		}
		return NewLoxBufWriterSize(loxFile.file, int(size)), nil
	})
	bufioFunc("writerStderr", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxBufWriterStderr(), nil
	})
	bufioFunc("writerStderrSize", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if size, ok := args[0].(int64); ok {
			if size < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'bufio.writerStderrSize' cannot be negative.")
			}
			return NewLoxBufWriterStderrSize(int(size)), nil
		}
		return argMustBeTypeAn(in.callToken, "writerStderrSize", "integer")
	})
	bufioFunc("writerStdout", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxBufWriterStdout(), nil
	})
	bufioFunc("writerStdoutSize", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if size, ok := args[0].(int64); ok {
			if size < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'bufio.writerStdoutSize' cannot be negative.")
			}
			return NewLoxBufWriterStdoutSize(int(size)), nil
		}
		return argMustBeTypeAn(in.callToken, "writerStdoutSize", "integer")
	})
	bufioFunc("writerStringBuilder", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if sb, ok := args[0].(*LoxStringBuilder); ok {
			return NewLoxBufWriter(&sb.builder), nil
		}
		return argMustBeType(in.callToken, "writerStringBuilder", "stringbuilder")
	})
	bufioFunc("writerStringBuilderSize", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxStringBuilder); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bufio.writerStringBuilderSize' must be a stringbuilder.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bufio.writerStringBuilderSize' must be an integer.")
		}
		size := args[1].(int64)
		if size < 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Integer argument to 'bufio.writerStringBuilderSize' cannot be negative.")
		}
		sb := args[0].(*LoxStringBuilder)
		return NewLoxBufWriterSize(&sb.builder, int(size)), nil
	})

	i.globals.Define(className, bufioClass)
}
