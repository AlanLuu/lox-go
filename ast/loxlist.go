package ast

import (
	"github.com/AlanLuu/lox/list"
)

type LoxList struct {
	elements list.List[Expr]
}
