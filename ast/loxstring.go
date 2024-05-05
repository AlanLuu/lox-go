package ast

type LoxString struct {
	str   string
	quote byte
}

func EmptyLoxString() *LoxString {
	return &LoxString{
		str:   "",
		quote: '\'',
	}
}

func (l *LoxString) NewLoxString(str string) *LoxString {
	return &LoxString{
		str:   str,
		quote: l.quote,
	}
}
