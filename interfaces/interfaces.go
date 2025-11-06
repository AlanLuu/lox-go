package interfaces

import "math/big"

type BigLength interface {
	BigLength() *big.Int
}

type Capacity interface {
	Capacity() int64
}

type Equatable interface {
	Equals(obj any) bool
}

type Index interface {
	Index(element any) (any, error)
}

type IndexSlice interface {
	IndexSlice(first, second any) (any, error)
}

type IndexInt interface {
	IndexInt(index int64) (any, error)
}

type IndexIntSlice interface {
	IndexIntSlice(first, second int64) (any, error)
}

type IndexBigInt interface {
	IndexBigInt(index *big.Int) (any, error)
}

type IndexBigIntSlice interface {
	IndexBigIntSlice(first, second *big.Int) (any, error)
}

type Iterator interface {
	HasNext() bool
	Next() any
}

type IteratorErr interface {
	Iterator
	HasNextErr() (bool, error)
	NextErr() (any, error)
}

type Iterable interface {
	Iterator() Iterator
}

type IterableErr interface {
	IteratorErr() IteratorErr
}

type LazyType interface {
	LazyTypeEval() error
}

type Length interface {
	Length() int64
}

type RandChoose interface {
	Length
	IndexInt
}

type RandChooseBig interface {
	BigLength
	IndexBigInt
}

type ReverseIterable interface {
	Iterable
	ReverseIterator() Iterator
}

type ReverseIterableErr interface {
	IterableErr
	ReverseIteratorErr() IteratorErr
}

type Type interface {
	Type() string
}
