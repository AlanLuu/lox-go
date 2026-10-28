package bf

func Defopts() *Options {
	return &Options{
		ArrSize:              DefaultArrSize,
		ProgramSize:          DefaultProgramSize,
		LoopTimes:            DefaultLoopTimes,
		ArrHasNegAddress:     false,
		ArrHasStaticSize:     false,
		ProgramHasStaticSize: false,
		ArrBlock:             false,
		ArrWrapAround:        false,
		SepInput:             false,
		NoOptimizations:      false,
		OnEOF:                0,
		Reader:               nil,
		Writer:               nil,
	}
}
