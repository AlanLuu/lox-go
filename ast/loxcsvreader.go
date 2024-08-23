package ast

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxCSVReader struct {
	reader  *csv.Reader
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxCSVReader(reader io.Reader) *LoxCSVReader {
	return NewLoxCSVReaderDelimiter(reader, ',')
}

func NewLoxCSVReaderDelimiter(reader io.Reader, delimiter rune) *LoxCSVReader {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = delimiter
	csvReader.FieldsPerRecord = -1
	return &LoxCSVReader{
		reader:  csvReader,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxCSVReader) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	csvReaderFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native csv reader fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	switch methodName {
	case "read":
		return csvReaderFunc(0, func(in *Interpreter, _ list.List[any]) (any, error) {
			fields, err := l.reader.Read()
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			fieldsList := list.NewListCap[any](int64(len(fields)))
			for _, field := range fields {
				fieldsList.Add(NewLoxStringQuote(field))
			}
			return NewLoxList(fieldsList), nil
		})
	case "readAll":
		return csvReaderFunc(0, func(in *Interpreter, _ list.List[any]) (any, error) {
			allFields, err := l.reader.ReadAll()
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			allFieldsList := list.NewListCap[any](int64(len(allFields)))
			for _, fields := range allFields {
				fieldsList := list.NewListCap[any](int64(len(fields)))
				for _, field := range fields {
					fieldsList.Add(NewLoxStringQuote(field))
				}
				allFieldsList.Add(NewLoxList(fieldsList))
			}
			return NewLoxList(allFieldsList), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "CSV readers have no property called '"+methodName+"'.")
}

func (l *LoxCSVReader) String() string {
	return fmt.Sprintf("<csv reader at %p>", l)
}

func (l *LoxCSVReader) Type() string {
	return "csv reader"
}
