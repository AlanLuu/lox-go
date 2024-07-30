package ast

import "fmt"

type LoxRangeDictSetKey struct {
	start int64
	stop  int64
	step  int64
}

func (l LoxRangeDictSetKey) String() string {
	if l.step == 1 {
		return fmt.Sprintf("range(%v, %v)", l.start, l.stop)
	}
	return fmt.Sprintf("range(%v, %v, %v)", l.start, l.stop, l.step)
}
