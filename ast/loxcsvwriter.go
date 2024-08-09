package ast

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
)

type LoxCSVWriter struct {
	writer  *csv.Writer
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxCSVWriter(writer io.Writer) *LoxCSVWriter {
	return NewLoxCSVWriterDelimiter(writer, ',')
}

func NewLoxCSVWriterDelimiter(writer io.Writer, delimiter rune) *LoxCSVWriter {
	csvWriter := csv.NewWriter(writer)
	csvWriter.Comma = delimiter
	if util.IsWindows() {
		csvWriter.UseCRLF = true
	}
	return &LoxCSVWriter{
		writer:  csvWriter,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxCSVWriter) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	csvWriterFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native csv writer fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'csv writer.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "bufferedWrite":
		return csvWriterFunc(1, func(in *Interpreter, args list.List[any]) (any, error) {
			if loxList, ok := args[0].(*LoxList); ok {
				records := list.NewList[string]()
				for _, element := range loxList.elements {
					var record string
					switch element := element.(type) {
					case *LoxString:
						record = element.str
					case fmt.Stringer:
						record = element.String()
					default:
						record = fmt.Sprint(element)
					}
					records.Add(record)
				}
				err := l.writer.Write(records)
				if err != nil {
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
				return nil, nil
			}
			return argMustBeType("list")
		})
	case "flush":
		return csvWriterFunc(0, func(in *Interpreter, _ list.List[any]) (any, error) {
			l.writer.Flush()
			err := l.writer.Error()
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		})
	case "write":
		return csvWriterFunc(1, func(in *Interpreter, args list.List[any]) (any, error) {
			if loxList, ok := args[0].(*LoxList); ok {
				record := list.NewList[string]()
				for _, element := range loxList.elements {
					var value string
					switch element := element.(type) {
					case *LoxString:
						value = element.str
					case fmt.Stringer:
						value = element.String()
					default:
						value = fmt.Sprint(element)
					}
					record.Add(value)
				}
				err := l.writer.Write(record)
				if err != nil {
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
				l.writer.Flush()
				err = l.writer.Error()
				if err != nil {
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
				return nil, nil
			}
			return argMustBeType("list")
		})
	case "writeAll":
		return csvWriterFunc(1, func(in *Interpreter, args list.List[any]) (any, error) {
			if loxList, ok := args[0].(*LoxList); ok {
				records := [][]string{}
				for _, outer := range loxList.elements {
					switch outer := outer.(type) {
					case interfaces.Iterable:
						record := []string{}
						it := outer.Iterator()
						for it.HasNext() {
							inner := it.Next()
							var value string
							switch inner := inner.(type) {
							case *LoxString:
								value = inner.str
							case fmt.Stringer:
								value = inner.String()
							default:
								value = fmt.Sprint(inner)
							}
							record = append(record, value)
						}
						records = append(records, record)
					default:
						records = nil
						return nil, loxerror.RuntimeError(in.callToken,
							"List argument to 'csv writer.writeAll' must only contain iterables.")
					}
				}
				err := l.writer.WriteAll(records)
				if err != nil {
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
				return nil, nil
			}
			return argMustBeType("list")
		})
	}
	return nil, loxerror.RuntimeError(name, "CSV writers have no property called '"+methodName+"'.")
}

func (l *LoxCSVWriter) String() string {
	return fmt.Sprintf("<csv writer at %p>", l)
}

func (l *LoxCSVWriter) Type() string {
	return "csv writer"
}
