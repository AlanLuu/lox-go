package ast

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"github.com/joho/godotenv"
)

type LoxDotenv struct {
	envs             map[string]string
	overload         bool
	panicOnSetEnvErr bool
	prevEnvs         []string
	methods          map[string]*struct{ ProtoLoxCallable }
}

func NewLoxDotenv() *LoxDotenv {
	dotenv, _ := NewLoxDotenvFromMap(make(map[string]string), false)
	return dotenv
}

func NewLoxDotenvFromBytes(bytes []byte) (*LoxDotenv, error) {
	envs, err := godotenv.UnmarshalBytes(bytes)
	if err != nil {
		return nil, err
	}
	if _, ok := envs[""]; ok {
		return nil, loxerror.Error(
			"dotenv buffer cannot have an empty string key.",
		)
	}
	return NewLoxDotenvFromMap(envs, false)
}

func NewLoxDotenvFromMap(envs map[string]string, checkErrs bool) (*LoxDotenv, error) {
	if checkErrs {
		for key := range envs {
			if key == "" {
				return nil, loxerror.Error(
					"dotenv map cannot have an empty string key.",
				)
			}
			if strings.Contains(key, "=") {
				return nil, loxerror.Error(
					"dotenv map cannot have a key with the '=' character.",
				)
			}
		}
	}
	return &LoxDotenv{
		envs:             envs,
		overload:         false,
		panicOnSetEnvErr: true,
		prevEnvs:         nil,
		methods:          make(map[string]*struct{ ProtoLoxCallable }),
	}, nil
}

func NewLoxDotenvFromString(str string) (*LoxDotenv, error) {
	envs, err := godotenv.Unmarshal(str)
	if err != nil {
		return nil, err
	}
	if _, ok := envs[""]; ok {
		return nil, loxerror.Error(
			"dotenv string cannot have an empty string key.",
		)
	}
	return NewLoxDotenvFromMap(envs, false)
}

func (l *LoxDotenv) activate() {
	if l.activated() || len(l.envs) == 0 {
		return
	}
	environ := os.Environ()
	l.prevEnvs = environ
	os.Clearenv()
	for key, value := range l.envs {
		if err := os.Setenv(key, value); err != nil {
			l.maybePanic(err)
		}
	}
}

func (l *LoxDotenv) activated() bool {
	return l.prevEnvs != nil
}

func (l *LoxDotenv) copyGlobalEnv() {
	environ := os.Environ()
	for _, env := range environ {
		envSplit := strings.Split(env, "=")
		l.setEnv(envSplit[0], envSplit[1])
	}
}

func (l *LoxDotenv) copyGlobalEnvForce() {
	environ := os.Environ()
	for _, env := range environ {
		envSplit := strings.Split(env, "=")
		l.setEnvForce(envSplit[0], envSplit[1])
	}
}

func (l *LoxDotenv) copyIntoGlobalEnv() {
	for key, value := range l.envs {
		_, ok := os.LookupEnv(key)
		if !ok {
			if err := os.Setenv(key, value); err != nil {
				l.maybePanic(err)
			}
		}
	}
}

func (l *LoxDotenv) copyIntoGlobalEnvForce() {
	for key, value := range l.envs {
		if err := os.Setenv(key, value); err != nil {
			l.maybePanic(err)
		}
	}
}

func (l *LoxDotenv) deactivate() {
	if !l.activated() {
		return
	}
	os.Clearenv()
	for _, env := range l.prevEnvs {
		envSplit := strings.Split(env, "=")
		if err := os.Setenv(envSplit[0], envSplit[1]); err != nil {
			l.maybePanic(err)
		}
	}
	l.prevEnvs = nil
}

func (l *LoxDotenv) deleteEnv(key string) bool {
	_, success := l.envs[key]
	if success {
		delete(l.envs, key)
		if l.activated() {
			if err := os.Unsetenv(key); err != nil {
				l.maybePanic(err)
			}
		}
	}
	return success
}

func (l *LoxDotenv) envBytes() []byte {
	return []byte(l.envString())
}

func (l *LoxDotenv) envString() string {
	//This method never returns an error
	envString, _ := godotenv.Marshal(l.envs)
	return envString
}

func (l *LoxDotenv) maybePanic(arg any) {
	if l.panicOnSetEnvErr {
		panic(arg)
	}
}

func (l *LoxDotenv) setEnvMap(envs map[string]string, checkErrs bool) error {
	if checkErrs {
		for key := range envs {
			if key == "" {
				return loxerror.Error(
					"dotenv map cannot have an empty string key.",
				)
			}
			if strings.Contains(key, "=") {
				return loxerror.Error(
					"dotenv map cannot have a key with the '=' character.",
				)
			}
		}
	}
	for key, value := range envs {
		l.setEnv(key, value)
	}
	return nil
}

func (l *LoxDotenv) setEnvMapForce(envs map[string]string, checkErrs bool) error {
	if checkErrs {
		for key := range envs {
			if key == "" {
				return loxerror.Error(
					"dotenv map cannot have an empty string key.",
				)
			}
			if strings.Contains(key, "=") {
				return loxerror.Error(
					"dotenv map cannot have a key with the '=' character.",
				)
			}
		}
	}
	for key, value := range envs {
		l.setEnvForce(key, value)
	}
	return nil
}

func (l *LoxDotenv) setEnv(key string, value string) bool {
	if _, ok := l.envs[key]; !ok || l.overload {
		l.envs[key] = value
		if l.activated() {
			if err := os.Setenv(key, value); err != nil {
				l.maybePanic(err)
			}
		}
		return true
	}
	return false
}

func (l *LoxDotenv) setEnvForce(key string, value string) {
	l.envs[key] = value
	if l.activated() {
		if err := os.Setenv(key, value); err != nil {
			l.maybePanic(err)
		}
	}
}

func (l *LoxDotenv) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	dotenvFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native dotenv object fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'dotenv object.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "activate":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.activate()
			return nil, nil
		})
	case "clear":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.activated() {
				os.Clearenv()
			}
			clear(l.envs)
			return nil, nil
		})
	case "containsEnv":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				_, success := l.envs[loxStr.str]
				return success, nil
			}
			return argMustBeType("string")
		})
	case "copyGlobalEnv":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.copyGlobalEnv()
			return l, nil
		})
	case "copyGlobalEnvForce":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.copyGlobalEnvForce()
			return l, nil
		})
	case "copyIntoGlobalEnv":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.copyIntoGlobalEnv()
			return l, nil
		})
	case "copyIntoGlobalEnvForce":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.copyIntoGlobalEnvForce()
			return l, nil
		})
	case "deactivate":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.deactivate()
			return nil, nil
		})
	case "deleteEnv":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.deleteEnv(loxStr.str)
				return l, nil
			}
			return argMustBeType("string")
		})
	case "deleteEnvBool":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return l.deleteEnv(loxStr.str), nil
			}
			return argMustBeType("string")
		})
	case "envBuf":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			envBytes := l.envBytes()
			buffer := EmptyLoxBufferCap(int64(len(envBytes)))
			for _, b := range envBytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "envStr":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.envString()), nil
		})
	case "exec":
		return dotenvFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"Expected at least 1 argument but got 0.")
			}
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dotenv object.exec' must be a string.")
			}
			cmdStr := args[0].(*LoxString).str
			var cmd *exec.Cmd
			if argsLen > 1 {
				cmdArgs := make([]string, 0, argsLen-1)
				for i := 1; i < argsLen; i++ {
					switch arg := args[i].(type) {
					case *LoxString:
						cmdArgs = append(cmdArgs, arg.str)
					default:
						cmdArgs = nil
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								"List element at index %v in the second argument of 'dotenv object.exec' must be a string.",
								i,
							),
						)
					}
				}
				cmd = exec.Command(cmdStr, cmdArgs...)
			} else {
				cmd = exec.Command(cmdStr)
			}
			if !l.activated() {
				l.activate()
				defer l.deactivate()
			}
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "execList":
		return dotenvFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dotenv object.execList' must be a string.")
			}
			if _, ok := args[1].(*LoxList); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dotenv object.execList' must be a list.")
			}
			secondList := args[1].(*LoxList).elements
			cmdArgs := make([]string, 0, len(secondList))
			for i, element := range secondList {
				switch element := element.(type) {
				case *LoxString:
					cmdArgs = append(cmdArgs, element.str)
				default:
					cmdArgs = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"List element at index %v in the second argument of 'dotenv object.execList' must be a string.",
							i,
						),
					)
				}
			}
			if !l.activated() {
				l.activate()
				defer l.deactivate()
			}
			cmdStr := args[0].(*LoxString).str
			cmd := exec.Command(cmdStr, cmdArgs...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "getEnv":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str, success := l.envs[loxStr.str]
				if !success {
					return nil, nil
				}
				return NewLoxStringQuote(str), nil
			}
			return argMustBeType("string")
		})
	case "isActivated", "activated":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.activated(), nil
		})
	case "isOverload":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.overload, nil
		})
	case "loadEnv":
		return dotenvFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"Expected at least 1 argument but got 0.")
			}
			filenames := list.NewListCap[string](int64(argsLen))
			for i := 0; i < argsLen; i++ {
				switch arg := args[i].(type) {
				case *LoxString:
					filenames.Add(arg.str)
				default:
					filenames = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument number %v in 'dotenv object.loadEnv' must be a string.",
							i+1,
						),
					)
				}
			}
			envs, err := godotenv.Read(filenames...)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			if _, ok := envs[""]; ok {
				return nil, loxerror.RuntimeError(name,
					"dotenv object.loadEnv: keys cannot be empty strings.")
			}
			l.setEnvMap(envs, false)
			return l, nil
		})
	case "loadEnvBuf":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxBuffer, ok := args[0].(*LoxBuffer); ok {
				envBytes := make([]byte, 0, len(loxBuffer.elements))
				for _, element := range loxBuffer.elements {
					envBytes = append(envBytes, byte(element.(int64)))
				}
				envs, err := godotenv.UnmarshalBytes(envBytes)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				if _, ok := envs[""]; ok {
					return nil, loxerror.RuntimeError(name,
						"dotenv object.loadEnvBuf: keys cannot be empty strings.")
				}
				l.setEnvMap(envs, false)
				return l, nil
			}
			return argMustBeType("buffer")
		})
	case "loadEnvBufForce":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxBuffer, ok := args[0].(*LoxBuffer); ok {
				envBytes := make([]byte, 0, len(loxBuffer.elements))
				for _, element := range loxBuffer.elements {
					envBytes = append(envBytes, byte(element.(int64)))
				}
				envs, err := godotenv.UnmarshalBytes(envBytes)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				if _, ok := envs[""]; ok {
					return nil, loxerror.RuntimeError(name,
						"dotenv object.loadEnvBufForce: keys cannot be empty strings.")
				}
				l.setEnvMapForce(envs, false)
				return l, nil
			}
			return argMustBeType("buffer")
		})
	case "loadEnvDict":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDict, ok := args[0].(*LoxDict); ok {
				const errMsg = "Dictionary argument to 'dotenv object.loadEnvDict' must only have strings."
				keysSet := make(map[string]bool)
				defer func() {
					keysSet = nil
				}()
				it := loxDict.Iterator()
				for it.HasNext() {
					pair := it.Next().(*LoxList).elements
					var key, value string
					switch pairKey := pair[0].(type) {
					case *LoxString:
						key = pairKey.str
					default:
						return nil, loxerror.RuntimeError(name, errMsg)
					}
					if key == "" {
						for key, value := range keysSet {
							if value {
								l.deleteEnv(key)
							}
						}
						return nil, loxerror.RuntimeError(name,
							"dotenv object.loadEnvDict: dictionary argument cannot have a key that is an empty string.")
					}
					if strings.Contains(key, "=") {
						for key, value := range keysSet {
							if value {
								l.deleteEnv(key)
							}
						}
						return nil, loxerror.RuntimeError(name,
							"dotenv object.loadEnvDict: dictionary argument cannot have a key with the '=' character.")
					}
					switch pairValue := pair[1].(type) {
					case *LoxString:
						value = pairValue.str
					default:
						return nil, loxerror.RuntimeError(name, errMsg)
					}
					keysSet[key] = l.setEnv(key, value)
				}
				return l, nil
			}
			return argMustBeType("dictionary")
		})
	case "loadEnvDictForce":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDict, ok := args[0].(*LoxDict); ok {
				const errMsg = "Dictionary argument to 'dotenv object.loadEnvDictForce' must only have strings."
				type keysSetVal struct {
					originalValue string
					modified      bool
				}
				keysSet := make(map[string]keysSetVal)
				defer func() {
					keysSet = nil
				}()
				it := loxDict.Iterator()
				for it.HasNext() {
					pair := it.Next().(*LoxList).elements
					var key, value string
					switch pairKey := pair[0].(type) {
					case *LoxString:
						key = pairKey.str
					default:
						return nil, loxerror.RuntimeError(name, errMsg)
					}
					if key == "" {
						for key, value := range keysSet {
							if value.modified {
								l.setEnvForce(key, value.originalValue)
							} else {
								l.deleteEnv(key)
							}
						}
						return nil, loxerror.RuntimeError(name,
							"dotenv object.loadEnvDictForce: dictionary argument cannot have a key that is an empty string.")
					}
					if strings.Contains(key, "=") {
						for key, value := range keysSet {
							if value.modified {
								l.setEnvForce(key, value.originalValue)
							} else {
								l.deleteEnv(key)
							}
						}
						return nil, loxerror.RuntimeError(name,
							"dotenv object.loadEnvDictForce: dictionary argument cannot have a key with the '=' character.")
					}
					switch pairValue := pair[1].(type) {
					case *LoxString:
						value = pairValue.str
					default:
						return nil, loxerror.RuntimeError(name, errMsg)
					}
					if !l.setEnv(key, value) {
						keysSet[key] = keysSetVal{l.envs[key], true}
						l.setEnvForce(key, value)
					} else {
						keysSet[key] = keysSetVal{"", false}
					}
				}
				return l, nil
			}
			return argMustBeType("dictionary")
		})
	case "loadEnvForce":
		return dotenvFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"Expected at least 1 argument but got 0.")
			}
			filenames := list.NewListCap[string](int64(argsLen))
			for i := 0; i < argsLen; i++ {
				switch arg := args[i].(type) {
				case *LoxString:
					filenames.Add(arg.str)
				default:
					filenames = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument number %v in 'dotenv object.loadEnvForce' must be a string.",
							i+1,
						),
					)
				}
			}
			envs, err := godotenv.Read(filenames...)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			if _, ok := envs[""]; ok {
				return nil, loxerror.RuntimeError(name,
					"dotenv object.loadEnvForce: keys cannot be empty strings.")
			}
			l.setEnvMapForce(envs, false)
			return l, nil
		})
	case "loadEnvStr":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				envs, err := godotenv.Unmarshal(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				if _, ok := envs[""]; ok {
					return nil, loxerror.RuntimeError(name,
						"dotenv object.loadEnvStr: keys cannot be empty strings.")
				}
				l.setEnvMap(envs, false)
				return l, nil
			}
			return argMustBeType("string")
		})
	case "loadEnvStrForce":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				envs, err := godotenv.Unmarshal(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				if _, ok := envs[""]; ok {
					return nil, loxerror.RuntimeError(name,
						"dotenv object.loadEnvStrForce: keys cannot be empty strings.")
				}
				l.setEnvMapForce(envs, false)
				return l, nil
			}
			return argMustBeType("string")
		})
	case "mustGetEnv":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := loxStr.str
				result, success := l.envs[str]
				if !success {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"dotenv object.mustGetEnv: unknown env variable '%v'.",
							str,
						),
					)
				}
				return NewLoxStringQuote(result), nil
			}
			return argMustBeType("string")
		})
	case "overload", "setOverload":
		return dotenvFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if overload, ok := args[0].(bool); ok {
				l.overload = overload
				return l, nil
			}
			return argMustBeType("boolean")
		})
	case "panicOnSetEnvErr":
		return dotenvFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			switch argsLen {
			case 0:
				return l.panicOnSetEnvErr, nil
			case 1:
				if panicOnSetEnvErr, ok := args[0].(bool); ok {
					l.panicOnSetEnvErr = panicOnSetEnvErr
					return l, nil
				}
				return argMustBeType("boolean")
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
		})
	case "printEnvStr":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			envStr := l.envString()
			if envStr != "" {
				fmt.Println(envStr)
			} else {
				fmt.Println("''")
			}
			return nil, nil
		})
	case "setEnv":
		return dotenvFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dotenv object.setEnv' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dotenv object.setEnv' must be a string.")
			}
			first := args[0].(*LoxString).str
			if first == "" {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dotenv object.setEnv' cannot be an empty string.")
			}
			if strings.Contains(first, "=") {
				return nil, loxerror.RuntimeError(name,
					"First string argument to 'dotenv object.setEnv' cannot contain the character '='.")
			}
			second := args[1].(*LoxString).str
			l.setEnv(first, second)
			return l, nil
		})
	case "setEnvBool":
		return dotenvFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dotenv object.setEnvBool' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dotenv object.setEnvBool' must be a string.")
			}
			first := args[0].(*LoxString).str
			if first == "" {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dotenv object.setEnvBool' cannot be an empty string.")
			}
			if strings.Contains(first, "=") {
				return nil, loxerror.RuntimeError(name,
					"First string argument to 'dotenv object.setEnvBool' cannot contain the character '='.")
			}
			second := args[1].(*LoxString).str
			return l.setEnv(first, second), nil
		})
	case "setEnvForce":
		return dotenvFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dotenv object.setEnvForce' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'dotenv object.setEnvForce' must be a string.")
			}
			first := args[0].(*LoxString).str
			if first == "" {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'dotenv object.setEnvForce' cannot be an empty string.")
			}
			if strings.Contains(first, "=") {
				return nil, loxerror.RuntimeError(name,
					"First string argument to 'dotenv object.setEnvForce' cannot contain the character '='.")
			}
			second := args[1].(*LoxString).str
			l.setEnvForce(first, second)
			return l, nil
		})
	case "toDict", "envDict":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			dict := EmptyLoxDict()
			for key, value := range l.envs {
				dict.setKeyValue(NewLoxStringQuote(key), NewLoxStringQuote(value))
			}
			return dict, nil
		})
	case "toList", "envList":
		return dotenvFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pairs := list.NewList[any]()
			it := l.Iterator()
			for it.HasNext() {
				pairs.Add(it.Next())
			}
			return NewLoxList(pairs), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Dotenv objects have no property called '"+methodName+"'.")
}

func (l *LoxDotenv) Iterator() interfaces.Iterator {
	pairs := list.NewListCap[*LoxList](int64(len(l.envs)))
	for key, value := range l.envs {
		pair := list.NewListCap[any](2)
		pair.Add(NewLoxStringQuote(key))
		pair.Add(NewLoxStringQuote(value))
		pairs.Add(NewLoxList(pair))
	}
	return &LoxDictIterator{pairs, 0}
}

func (l *LoxDotenv) String() string {
	if l.activated() {
		return fmt.Sprintf("<activated dotenv object at %p>", l)
	}
	return fmt.Sprintf("<dotenv object at %p>", l)
}

func (l *LoxDotenv) Type() string {
	return "dotenv"
}
