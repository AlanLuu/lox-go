package interfaces

type Equatable interface {
	Equals(obj any) bool
}

type Length interface {
	Length() int64
}

type Type interface {
	Type() string
}
