package loxsignal

type LoopSignal struct{}

func (l LoopSignal) String() string {
	return "Lox loop signal"
}

func (l LoopSignal) Signal() {}
