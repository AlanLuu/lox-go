package ast

import "github.com/AlanLuu/lox/list"

func getArgList(callback *LoxFunction, numArgs int) list.List[any] {
	argList := list.NewListLen[any](int64(numArgs))
	callbackArity := callback.arity()
	if callbackArity > numArgs {
		for i := 0; i < callbackArity-numArgs; i++ {
			argList.Add(nil)
		}
	}
	return argList
}

func isValidPortNum[T int | int64](portNum T) bool {
	return portNum >= 0 && portNum <= 65535
}
