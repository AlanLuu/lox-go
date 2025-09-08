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
