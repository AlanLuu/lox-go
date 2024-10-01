package interfaces

type Capacity interface {
	Capacity() int64
}

type Equatable interface {
	Equals(obj any) bool
}

type Iterator interface {
	HasNext() bool
	Next() any
}

type Iterable interface {
	Iterator() Iterator
}

type LazyType interface {
	LazyTypeEval() error
}

type Length interface {
	Length() int64
}

type ReverseIterable interface {
	Iterable
	ReverseIterator() Iterator
}

type Type interface {
	Type() string
}
