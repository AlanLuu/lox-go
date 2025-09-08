package ast

import (
	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
)

func assertIterErr(iter interfaces.Iterator) interfaces.IteratorErr {
	return convertIface[interfaces.IteratorErr](iter)
}

func convertIface[T any](iface any) T {
	result, ok := iface.(T)
	if !ok {
		var t T
		return t
	}
	return result
}

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

func reverseUint8[T ~uint8](b T) T {
	//https://stackoverflow.com/a/2602885
	b = (b&0xF0)>>4 | (b&0x0F)<<4
	b = (b&0xCC)>>2 | (b&0x33)<<2
	b = (b&0xAA)>>1 | (b&0x55)<<1
	return b
}
