package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"golang.org/x/sys/windows"
)

func (i *Interpreter) defineWindowsFuncs() {
	className := "windows"
	windowsClass := NewLoxClass(className, nil, false)
	windowsFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native windows fn %v at %p>", name, &s)
		}
		windowsClass.classProperties[name] = s
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'windows.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	windowsFunc("computerName", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		version, err := windows.ComputerName()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxStringQuote(version), nil
	})
	windowsFunc("getFileType", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			fileType, err := windows.GetFileType(windows.Handle(arg))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return int64(fileType), nil
		}
		return argMustBeTypeAn(in.callToken, "getFileType", "integer")
	})
	windowsFunc("getFileTypeStr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			fileType, err := windows.GetFileType(windows.Handle(arg))
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxString(map[uint32]string{
				0x0002: "CHAR",
				0x0001: "DISK",
				0x0003: "PIPE",
				0x8000: "REMOTE",
				0x0000: "UNKNOWN",
			}[fileType], '\''), nil
		}
		return argMustBeTypeAn(in.callToken, "getFileTypeStr", "integer")
	})
	windowsFunc("getLastError", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		err := windows.GetLastError()
		if err == nil {
			return nil, nil
		}
		return NewLoxError(err), nil
	})
	windowsFunc("getLogicalDrives", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		mask, err := windows.GetLogicalDrives()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return int64(mask), nil
	})
	windowsFunc("getMaximumProcessorCount", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return int64(windows.GetMaximumProcessorCount(windows.ALL_PROCESSOR_GROUPS)), nil
	})
	windowsFunc("getSystemDirectory", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		dir, err := windows.GetSystemDirectory()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxStringQuote(dir), nil
	})
	windowsFunc("getSystemWindowsDirectory", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		dir, err := windows.GetSystemWindowsDirectory()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxStringQuote(dir), nil
	})
	windowsFunc("listDrives", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		mask, err := windows.GetLogicalDrives()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		drives := list.NewList[any]()
		var num uint32 = 1
		var letter rune = 'A'
		for {
			if mask&num != 0 {
				drives.Add(NewLoxString(string(letter)+":\\\\", '\''))
			}
			if num == 1<<31 || num<<1 > mask || letter >= 'Z' {
				break
			}
			num <<= 1
			letter++
		}
		return NewLoxList(drives), nil
	})
	windowsClass.classProperties["stderr"] = int64(windows.Stderr)
	windowsClass.classProperties["stdin"] = int64(windows.Stdin)
	windowsClass.classProperties["stdout"] = int64(windows.Stdout)

	i.globals.Define(className, windowsClass)
}
