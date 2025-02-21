package ast

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/AlanLuu/lox/util"
	"github.com/joho/godotenv"
)

func (i *Interpreter) defineDotenvFuncs() {
	className := "dotenv"
	dotenvClass := NewLoxClass(className, nil, false)
	dotenvFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native dotenv fn %v at %p>", name, &s)
		}
		dotenvClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'dotenv.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	dotenvFunc("dictToEnv", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxDict, ok := args[0].(*LoxDict); ok {
			const errMsg = "Dictionary argument to 'dotenv.dictToEnv' must only have strings."
			envMap := map[string]string{}
			it := loxDict.Iterator()
			for it.HasNext() {
				pair := it.Next().(*LoxList).elements
				var key, value string
				switch pairKey := pair[0].(type) {
				case *LoxString:
					key = pairKey.str
				default:
					return nil, loxerror.RuntimeError(in.callToken, errMsg)
				}
				if key == "" {
					envMap = nil
					return nil, loxerror.RuntimeError(in.callToken,
						"Dictionary argument to 'dotenv.dictToEnv' cannot have a key that is an empty string.")
				}
				if strings.Contains(key, "=") {
					envMap = nil
					return nil, loxerror.RuntimeError(in.callToken,
						"Dictionary argument to 'dotenv.dictToEnv' cannot have a key with the '=' character.")
				}
				switch pairValue := pair[1].(type) {
				case *LoxString:
					value = pairValue.str
				default:
					return nil, loxerror.RuntimeError(in.callToken, errMsg)
				}
				envMap[key] = value
			}
			envStr, err := godotenv.Marshal(envMap)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxStringQuote(envStr), nil
		}
		return argMustBeType(in.callToken, "dictToEnv", "dictionary")
	})
	dotenvFunc("dictToEnvBuf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxDict, ok := args[0].(*LoxDict); ok {
			const errMsg = "Dictionary argument to 'dotenv.dictToEnvBuf' must only have strings."
			envMap := map[string]string{}
			it := loxDict.Iterator()
			for it.HasNext() {
				pair := it.Next().(*LoxList).elements
				var key, value string
				switch pairKey := pair[0].(type) {
				case *LoxString:
					key = pairKey.str
				default:
					return nil, loxerror.RuntimeError(in.callToken, errMsg)
				}
				if key == "" {
					envMap = nil
					return nil, loxerror.RuntimeError(in.callToken,
						"Dictionary argument to 'dotenv.dictToEnvBuf' cannot have a key that is an empty string.")
				}
				if strings.Contains(key, "=") {
					envMap = nil
					return nil, loxerror.RuntimeError(in.callToken,
						"Dictionary argument to 'dotenv.dictToEnvBuf' cannot have a key with the '=' character.")
				}
				switch pairValue := pair[1].(type) {
				case *LoxString:
					value = pairValue.str
				default:
					return nil, loxerror.RuntimeError(in.callToken, errMsg)
				}
				envMap[key] = value
			}
			envStr, err := godotenv.Marshal(envMap)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			envBytes := []byte(envStr)
			buffer := EmptyLoxBufferCap(int64(len(envBytes)))
			for _, b := range envBytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
			return buffer, nil
		}
		return argMustBeType(in.callToken, "dictToEnvBuf", "dictionary")
	})
	dotenvFunc("dictToEnvFile", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 1 && argsLen != 2 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(*LoxDict); !ok {
			var msg string
			if argsLen == 2 {
				msg = "First argument to 'dotenv.dictToEnvFile' must be a dictionary."
			} else {
				msg = "Argument to 'dotenv.dictToEnvFile' must be a dictionary."
			}
			return nil, loxerror.RuntimeError(in.callToken, msg)
		}
		if argsLen == 2 {
			switch arg := args[1].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(in.callToken,
						"File argument to 'dotenv.dictToEnvFile' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'dotenv.dictToEnvFile' must be a file or string.")
			}
		}

		loxDict := args[0].(*LoxDict)
		const errMsg = "Dictionary argument to 'dotenv.dictToEnvFile' must only have strings."
		envMap := map[string]string{}
		it := loxDict.Iterator()
		for it.HasNext() {
			pair := it.Next().(*LoxList).elements
			var key, value string
			switch pairKey := pair[0].(type) {
			case *LoxString:
				key = pairKey.str
			default:
				return nil, loxerror.RuntimeError(in.callToken, errMsg)
			}
			if key == "" {
				envMap = nil
				return nil, loxerror.RuntimeError(in.callToken,
					"Dictionary argument to 'dotenv.dictToEnvFile' cannot have a key that is an empty string.")
			}
			if strings.Contains(key, "=") {
				envMap = nil
				return nil, loxerror.RuntimeError(in.callToken,
					"Dictionary argument to 'dotenv.dictToEnvFile' cannot have a key with the '=' character.")
			}
			switch pairValue := pair[1].(type) {
			case *LoxString:
				value = pairValue.str
			default:
				return nil, loxerror.RuntimeError(in.callToken, errMsg)
			}
			envMap[key] = value
		}

		envStr, envStrErr := godotenv.Marshal(envMap)
		if envStrErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, envStrErr.Error())
		}

		envBytes := []byte(envStr)
		if r, _ := utf8.DecodeLastRuneInString(envStr); r != utf8.RuneError && r != '\n' {
			if util.IsWindows() {
				envBytes = append(envBytes, '\r', '\n')
			} else {
				envBytes = append(envBytes, '\n')
			}
		}
		var err error
		if argsLen == 2 {
			switch arg := args[1].(type) {
			case *LoxFile:
				_, err = arg.file.Write(envBytes)
			case *LoxString:
				err = os.WriteFile(arg.str, envBytes, 0666)
			}
		} else {
			err = os.WriteFile(".env", envBytes, 0666)
		}

		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	dotenvFunc("dictToEnvFileNoNewlineEnd", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 1 && argsLen != 2 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(*LoxDict); !ok {
			var msg string
			if argsLen == 2 {
				msg = "First argument to 'dotenv.dictToEnvFileNoNewlineEnd' must be a dictionary."
			} else {
				msg = "Argument to 'dotenv.dictToEnvFileNoNewlineEnd' must be a dictionary."
			}
			return nil, loxerror.RuntimeError(in.callToken, msg)
		}
		if argsLen == 2 {
			switch arg := args[1].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(in.callToken,
						"File argument to 'dotenv.dictToEnvFileNoNewlineEnd' must be in write or append mode.")
				}
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'dotenv.dictToEnvFileNoNewlineEnd' must be a file or string.")
			}
		}

		loxDict := args[0].(*LoxDict)
		const errMsg = "Dictionary argument to 'dotenv.dictToEnvFileNoNewlineEnd' must only have strings."
		envMap := map[string]string{}
		it := loxDict.Iterator()
		for it.HasNext() {
			pair := it.Next().(*LoxList).elements
			var key, value string
			switch pairKey := pair[0].(type) {
			case *LoxString:
				key = pairKey.str
			default:
				return nil, loxerror.RuntimeError(in.callToken, errMsg)
			}
			if key == "" {
				envMap = nil
				return nil, loxerror.RuntimeError(in.callToken,
					"Dictionary argument to 'dotenv.dictToEnvFileNoNewlineEnd' cannot have a key that is an empty string.")
			}
			if strings.Contains(key, "=") {
				envMap = nil
				return nil, loxerror.RuntimeError(in.callToken,
					"Dictionary argument to 'dotenv.dictToEnvFileNoNewlineEnd' cannot have a key with the '=' character.")
			}
			switch pairValue := pair[1].(type) {
			case *LoxString:
				value = pairValue.str
			default:
				return nil, loxerror.RuntimeError(in.callToken, errMsg)
			}
			envMap[key] = value
		}

		envStr, envStrErr := godotenv.Marshal(envMap)
		if envStrErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, envStrErr.Error())
		}

		envBytes := []byte(envStr)
		var err error
		if argsLen == 2 {
			switch arg := args[1].(type) {
			case *LoxFile:
				_, err = arg.file.Write(envBytes)
			case *LoxString:
				err = os.WriteFile(arg.str, envBytes, 0666)
			}
		} else {
			err = os.WriteFile(".env", envBytes, 0666)
		}

		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	dotenvFunc("exec", 4, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxList); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'dotenv.exec' must be a list.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'dotenv.exec' must be a string.")
		}
		if _, ok := args[2].(*LoxList); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'dotenv.exec' must be a list.")
		}
		if _, ok := args[3].(bool); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Fourth argument to 'dotenv.exec' must be a boolean.")
		}

		firstList := args[0].(*LoxList).elements
		filenames := make([]string, 0, len(firstList))
		for i, element := range firstList {
			switch element := element.(type) {
			case *LoxString:
				filenames = append(filenames, element.str)
			default:
				filenames = nil
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"List element at index %v in the first argument of 'dotenv.exec' must be a string.",
						i,
					),
				)
			}
		}

		secondList := args[2].(*LoxList).elements
		cmdArgs := make([]string, 0, len(secondList))
		for i, element := range secondList {
			switch element := element.(type) {
			case *LoxString:
				cmdArgs = append(cmdArgs, element.str)
			default:
				cmdArgs = nil
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"List element at index %v in the third argument of 'dotenv.exec' must be a string.",
						i,
					),
				)
			}
		}

		cmd := args[1].(*LoxString).str
		overload := args[3].(bool)
		err := godotenv.Exec(filenames, cmd, cmdArgs, overload)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	dotenvFunc("load", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var err error
		argsLen := len(args)
		switch argsLen {
		case 0:
			err = godotenv.Load()
		default:
			filenames := make([]string, 0, argsLen)
			for i := 0; i < argsLen; i++ {
				switch arg := args[i].(type) {
				case *LoxString:
					filenames = append(filenames, arg.str)
				default:
					filenames = nil
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"Argument number %v in 'dotenv.load' must be a string.",
							i+1,
						),
					)
				}
			}
			err = godotenv.Load(filenames...)
		}
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	dotenvFunc("new", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		switch argsLen {
		case 0:
			return NewLoxDotenv(), nil
		case 1:
			switch arg := args[0].(type) {
			case *LoxBuffer:
				envBytes := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					envBytes = append(envBytes, byte(element.(int64)))
				}
				dotenv, err := NewLoxDotenvFromBytes(envBytes)
				if err != nil {
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
				return dotenv, nil
			case *LoxDict:
				const errMsg = "Dictionary argument to 'dotenv.new' must only have strings."
				envs := map[string]string{}
				it := arg.Iterator()
				for it.HasNext() {
					pair := it.Next().(*LoxList).elements
					var key, value string
					switch pairKey := pair[0].(type) {
					case *LoxString:
						key = pairKey.str
					default:
						return nil, loxerror.RuntimeError(in.callToken, errMsg)
					}
					if key == "" {
						return nil, loxerror.RuntimeError(in.callToken,
							"Dictionary argument to 'dotenv.new' cannot have a key that is an empty string.")
					}
					if strings.Contains(key, "=") {
						return nil, loxerror.RuntimeError(in.callToken,
							"Dictionary argument to 'dotenv.new' cannot have a key with the '=' character.")
					}
					switch pairValue := pair[1].(type) {
					case *LoxString:
						value = pairValue.str
					default:
						return nil, loxerror.RuntimeError(in.callToken, errMsg)
					}
					envs[key] = value
				}
				dotenv, _ := NewLoxDotenvFromMap(envs, false)
				return dotenv, nil
			case *LoxString:
				dotenv, err := NewLoxDotenvFromString(arg.str)
				if err != nil {
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
				return dotenv, nil
			default:
				return argMustBeType(in.callToken, "new", "buffer, dictionary, or string")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
	})
	dotenvFunc("overload", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var err error
		argsLen := len(args)
		switch argsLen {
		case 0:
			err = godotenv.Overload()
		default:
			filenames := make([]string, 0, argsLen)
			for i := 0; i < argsLen; i++ {
				switch arg := args[i].(type) {
				case *LoxString:
					filenames = append(filenames, arg.str)
				default:
					filenames = nil
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"Argument number %v in 'dotenv.overload' must be a string.",
							i+1,
						),
					)
				}
			}
			err = godotenv.Overload(filenames...)
		}
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return nil, nil
	})
	dotenvFunc("parse", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var envMap map[string]string
		var err error

		switch arg := args[0].(type) {
		case *LoxBuffer:
			bytes := make([]byte, 0, len(arg.elements))
			for _, element := range arg.elements {
				bytes = append(bytes, byte(element.(int64)))
			}
			envMap, err = godotenv.UnmarshalBytes(bytes)
		case *LoxString:
			envMap, err = godotenv.Unmarshal(arg.str)
		default:
			return argMustBeType(in.callToken, "parse", "buffer or string")
		}

		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}

		dict := EmptyLoxDict()
		for key, value := range envMap {
			if key == "" {
				dict = nil
				return nil, loxerror.RuntimeError(in.callToken,
					"dotenv.parse: keys cannot be empty strings.")
			}
			if strings.Contains(key, "=") {
				dict = nil
				return nil, loxerror.RuntimeError(in.callToken,
					"dotenv.parse: keys cannot contain the character '='.")
			}
			dict.setKeyValue(NewLoxStringQuote(key), NewLoxStringQuote(value))
		}
		return dict, nil
	})
	dotenvFunc("read", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		filenames := make([]string, 0, argsLen)
		for i := 0; i < argsLen; i++ {
			switch arg := args[i].(type) {
			case *LoxString:
				filenames = append(filenames, arg.str)
			default:
				filenames = nil
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"Argument number %v in 'dotenv.read' must be a string.",
						i+1,
					),
				)
			}
		}
		envMap, err := godotenv.Read(filenames...)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		dict := EmptyLoxDict()
		for key, value := range envMap {
			if key == "" {
				dict = nil
				return nil, loxerror.RuntimeError(in.callToken,
					"dotenv.read: keys cannot be empty strings.")
			}
			if strings.Contains(key, "=") {
				dict = nil
				return nil, loxerror.RuntimeError(in.callToken,
					"dotenv.read: keys cannot contain the character '='.")
			}
			dict.setKeyValue(NewLoxStringQuote(key), NewLoxStringQuote(value))
		}
		return dict, nil
	})

	i.globals.Define(className, dotenvClass)
}
