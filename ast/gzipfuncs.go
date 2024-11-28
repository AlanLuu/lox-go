package ast

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

const (
	GZIP_USE_BUFFER = 1 + iota
)

func defineGzipFields(gzipClass *LoxClass) {
	gzipFields := map[string]int64{
		"bestCompression":    gzip.BestCompression,
		"bestSpeed":          gzip.BestSpeed,
		"defaultCompression": gzip.DefaultCompression,
		"huffmanOnly":        gzip.HuffmanOnly,
		"noCompression":      gzip.NoCompression,
		"USE_BUFFER":         GZIP_USE_BUFFER,
	}
	for key, value := range gzipFields {
		gzipClass.classProperties[key] = value
	}
}

func (i *Interpreter) defineGzipFuncs() {
	className := "gzip"
	gzipClass := NewLoxClass(className, nil, false)
	gzipFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native gzip fn %v at %p>", name, &s)
		}
		gzipClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'gzip.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	defineGzipFields(gzipClass)
	gzipFunc("buffer", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 1 && argsLen != 2 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
		var data []byte
		switch arg := args[0].(type) {
		case *LoxBuffer:
			data = make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				data = append(data, byte(element.(int64)))
			}
		case *LoxFile:
			if !arg.isRead() {
				return nil, loxerror.RuntimeError(in.callToken,
					"File argument to 'gzip.buffer' must be in read mode.")
			}
			var readErr error
			data, readErr = io.ReadAll(arg.file)
			if readErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, readErr.Error())
			}
		case *LoxString:
			data = []byte(arg.str)
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'gzip.buffer' must be a buffer, file, or string.")
		}
		var compressionLevel int = gzip.DefaultCompression
		if argsLen == 2 {
			if arg, ok := args[1].(int64); ok {
				compressionLevel = int(arg)
			} else {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'gzip.buffer' must be an integer.")
			}
		}
		bytesBuffer := new(bytes.Buffer)
		gzipWriter, err := gzip.NewWriterLevel(bytesBuffer, compressionLevel)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		_, err = gzipWriter.Write(data)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		err = gzipWriter.Close()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		compressed := bytesBuffer.Bytes()
		buffer := EmptyLoxBufferCap(int64(len(compressed)))
		for _, b := range compressed {
			addErr := buffer.add(int64(b))
			if addErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
			}
		}
		return buffer, nil
	})
	gzipFunc("reader", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var reader io.Reader
		switch arg := args[0].(type) {
		case *LoxBuffer:
			data := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				data = append(data, byte(element.(int64)))
			}
			reader = bytes.NewBuffer(data)
		case *LoxFile:
			if !arg.isRead() {
				return nil, loxerror.RuntimeError(in.callToken,
					"Cannot create gzip reader for file not in read mode.")
			}
			reader = arg.file
		default:
			return argMustBeType(in.callToken, "reader", "buffer or file")
		}
		gzipReader, err := NewLoxGZIPReader(reader)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return gzipReader, nil
	})
	gzipFunc("write", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 2 && argsLen != 3 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(*LoxFile); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'gzip.write' must be a file.")
		}
		switch args[1].(type) {
		case *LoxBuffer:
		case *LoxFile:
		case *LoxString:
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'gzip.write' must be a buffer, file, or string.")
		}
		var compressionLevel int = gzip.DefaultCompression
		if argsLen == 3 {
			if arg, ok := args[2].(int64); ok {
				compressionLevel = int(arg)
			} else {
				return nil, loxerror.RuntimeError(in.callToken,
					"Third argument to 'gzip.write' must be an integer.")
			}
		}
		loxFile := args[0].(*LoxFile)
		if !loxFile.isWrite() && !loxFile.isAppend() {
			return nil, loxerror.RuntimeError(in.callToken,
				"First file argument to 'gzip.write' must be in write or append mode.")
		}
		var data []byte
		switch arg := args[1].(type) {
		case *LoxBuffer:
			data = make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				data = append(data, byte(element.(int64)))
			}
		case *LoxFile:
			if !arg.isRead() {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second file argument to 'gzip.write' must be in read mode.")
			}
			var readErr error
			data, readErr = io.ReadAll(arg.file)
			if readErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, readErr.Error())
			}
		case *LoxString:
			data = []byte(arg.str)
		}
		gzipWriter, err := gzip.NewWriterLevel(loxFile.file, compressionLevel)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		_, err = gzipWriter.Write(data)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		err = gzipWriter.Close()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	gzipFunc("writer", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *LoxFile:
			if !arg.isWrite() && !arg.isAppend() {
				return nil, loxerror.RuntimeError(in.callToken,
					"Cannot create gzip writer for file not in write or append mode.")
			}
			return NewLoxGZIPWriter(arg.file), nil
		case int64:
			switch arg {
			case GZIP_USE_BUFFER:
				return NewLoxGZIPWriterBytes(new(bytes.Buffer)), nil
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"Integer argument to 'gzip.writer' must be equal to the field 'gzip.USE_BUFFER'.")
			}
		default:
			return argMustBeType(in.callToken, "writer", "file or the field 'gzip.USE_BUFFER'")
		}
	})
	gzipFunc("writerLevel", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *LoxFile:
		case int64:
			switch arg {
			case GZIP_USE_BUFFER:
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"First integer argument to 'gzip.writerLevel' must be equal to the field 'gzip.USE_BUFFER'.")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'gzip.writerLevel' must be a file or the field 'gzip.USE_BUFFER'.")
		}
		var compressionLevel int = gzip.DefaultCompression
		switch arg := args[1].(type) {
		case int64:
			compressionLevel = int(arg)
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'gzip.writerLevel' must be an integer.")
		}
		var gzipWriter *LoxGZIPWriter
		var err error
		switch arg := args[0].(type) {
		case *LoxFile:
			if !arg.isWrite() && !arg.isAppend() {
				return nil, loxerror.RuntimeError(in.callToken,
					"Cannot create gzip writer for file not in write or append mode.")
			}
			gzipWriter, err = NewLoxGZIPWriterLevel(arg.file, compressionLevel)
		case int64:
			switch arg {
			case GZIP_USE_BUFFER:
				gzipWriter, err = NewLoxGZIPWriterBytesLevel(new(bytes.Buffer), compressionLevel)
			}
		}
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return gzipWriter, nil
	})

	i.globals.Define(className, gzipClass)
}
