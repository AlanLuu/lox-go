package ast

import "fmt"

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

func StringIndexMustBeWholeNum(index any) string {
	return fmt.Sprintf("String index '%v' must be a whole number.", index)
}

func StringIndexOutOfRange(index int64) string {
	return fmt.Sprintf("String index %v out of range.", index)
}

func (l *LoxString) NewLoxString(str string) *LoxString {
	return &LoxString{
		str:   str,
		quote: l.quote,
	}
}

func (l *LoxString) String() string {
	return l.str
}
