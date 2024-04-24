package ast

type LoxInstance struct {
	class LoxClass
}

func (i LoxInstance) String() string {
	return i.class.name + " instance"
}
