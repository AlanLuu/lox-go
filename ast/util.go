package ast

import (
	"fmt"
	"math/big"

	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
)

func assertIterErr(iter interfaces.Iterator) interfaces.IteratorErr {
	return convertIface[interfaces.IteratorErr](iter)
}

func checkValidBigint(num *big.Int) (int64, error) {
	if !num.IsInt64() {
		return 0, loxerror.Error(
			fmt.Sprintf(
				"bigint index value '%v' is out of range.",
				bigint.String(num),
			),
		)
	}
	return num.Int64(), nil
}

func convertIface[T any](iface any) T {
	result, ok := iface.(T)
	if !ok {
		var t T
		return t
	}
	return result
}

func convertNegIndex(l int64, num int64) int64 {
	if num < 0 {
		num += l
	}
	return num
}

func convertPosIndex(l int64, num int64) int64 {
	if num > l {
		num = l
	}
	return num
}

func convertSliceIndex(l int64, num int64) int64 {
	return convertPosIndex(l, convertNegIndex(l, num))
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
