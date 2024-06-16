package interfaces

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

type Length interface {
	Length() int64
}

type Type interface {
	Type() string
}
