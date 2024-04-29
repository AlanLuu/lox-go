package equatable

type Equatable interface {
	Equals(obj any) bool
}
