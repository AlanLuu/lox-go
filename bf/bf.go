package bf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	DefaultArrSize     = 30000
	DefaultProgramSize = 5000
	DefaultLoopTimes   = 1
	DoNothingOnEOF     = 256
	ErrMsgHeader       = "bf: "
	NegativeArrErrMsg  = "array address became negative"
)

const (
	op_END = iota
	op_FORWARD
	op_BACKWARD
	op_FOWARD_BY
	op_INCREMENT
	op_DECREMENT
	op_INCREMENT_BY
	op_OUTPUT
	op_INPUT
	op_JUMP_FORWARD
	op_JUMP_BACKWARD
)

type chunk struct {
	op  int
	val int
	err int
}

type Options struct {
	ArrSize              int
	ProgramSize          int
	LoopTimes            int64
	ArrHasNegAddress     bool
	ArrHasStaticSize     bool
	ProgramHasStaticSize bool
	ArrBlock             bool
	ArrWrapAround        bool
	SepInput             bool
	NoOptimizations      bool
	OnEOF                int16
	sepInput             *bytes.Buffer
	Reader               io.Reader
	Writer               io.Writer
}

func (o *Options) eofInRange() bool {
	return o.OnEOF >= -256 && o.OnEOF <= 255
}

func bfErr(format string, a ...any) error {
	if len(a) == 0 {
		return errors.New(ErrMsgHeader + format)
	}
	return fmt.Errorf(ErrMsgHeader+format, a...)
}

func bfInterpretErr(index int, format string, a ...any) error {
	return bfErr(
		"error at bf character #"+fmt.Sprint(index)+": "+format,
		a...,
	)
}

func bfInterpretErrOptimized(index int, format string, a ...any) error {
	return bfErr(
		"error in optimized character set right before bf character #"+
			fmt.Sprint(index)+
			": "+format,
		a...,
	)
}

func bfParseErr(index int, format string, a ...any) error {
	return bfErr(
		"parse error at bf character #"+fmt.Sprint(index)+": "+format,
		a...,
	)
}

func bfParseErrAtEnd(format string, a ...any) error {
	return bfErr(
		"parse error at end: "+format,
		a...,
	)
}

func interpret(chunks []chunk, opts *Options) error {
	loopTimes := opts.LoopTimes
	if loopTimes == 0 {
		return nil
	}
	arr := make([]uint8, opts.ArrSize)
	var negArr []uint8
	arrNegAddr := opts.ArrHasNegAddress
	arrBlock := opts.ArrBlock
	arrWrapAround := opts.ArrWrapAround
	hasOpts := !opts.NoOptimizations
	eofInRange := opts.eofInRange()
	chunksLen := len(chunks)
	var errFunc func(index int, format string, a ...any) error
	if opts.NoOptimizations {
		errFunc = bfInterpretErr
	} else {
		errFunc = bfInterpretErrOptimized
	}
	var loopTimesCount int64 = 0
	var b [1]byte
	ip, index := 0, 0
loop_begin:
	for ; ip >= 0 && ip < chunksLen && (arrNegAddr || index >= 0); ip++ {
		switch chunks[ip].op {
		case op_END:
			if loopTimes < 0 {
				for i := 0; i < len(arr); i++ {
					arr[i] = 0
				}
				if negArr != nil {
					for i := 0; i < len(negArr); i++ {
						negArr[i] = 0
					}
				}
				b[0] = 0
				ip, index = 0, 0
				goto loop_begin
			} else if loopTimes > 1 && loopTimesCount < loopTimes-1 {
				loopTimesCount++
				for i := 0; i < len(arr); i++ {
					arr[i] = 0
				}
				if negArr != nil {
					for i := 0; i < len(negArr); i++ {
						negArr[i] = 0
					}
				}
				b[0] = 0
				ip, index = 0, 0
				goto loop_begin
			}
			return nil
		case op_FORWARD:
			index++
		case op_BACKWARD:
			index--
		case op_FOWARD_BY:
			index += chunks[ip].val
		case op_INCREMENT:
			if index < 0 {
				if negArr == nil {
					negArr = make([]uint8, opts.ArrSize)
					for ^index >= len(negArr) {
						negArr = append(negArr, 0)
						negArr = negArr[:cap(negArr)]
					}
				}
				negArr[^index]++
			} else {
				arr[index]++
			}
		case op_DECREMENT:
			if index < 0 {
				if negArr == nil {
					negArr = make([]uint8, opts.ArrSize)
					for ^index >= len(negArr) {
						negArr = append(negArr, 0)
						negArr = negArr[:cap(negArr)]
					}
				}
				negArr[^index]--
			} else {
				arr[index]--
			}
		case op_INCREMENT_BY:
			if index < 0 {
				if negArr == nil {
					negArr = make([]uint8, opts.ArrSize)
					for ^index >= len(negArr) {
						negArr = append(negArr, 0)
						negArr = negArr[:cap(negArr)]
					}
				}
				negArr[^index] += uint8(chunks[ip].val)
			} else {
				arr[index] += uint8(chunks[ip].val)
			}
		case op_OUTPUT:
			if index < 0 {
				if negArr == nil {
					negArr = make([]uint8, opts.ArrSize)
					for ^index >= len(negArr) {
						negArr = append(negArr, 0)
						negArr = negArr[:cap(negArr)]
					}
				}
				b[0] = negArr[^index]
			} else {
				b[0] = arr[index]
			}
			var err error
			if opts.Writer != nil {
				_, err = opts.Writer.Write(b[:])
			} else {
				_, err = os.Stdout.Write(b[:])
			}
			if err != nil {
				return errFunc(
					chunks[ip].err,
					"failed to write byte at array position #%v",
					index+1,
				)
			}
		case op_INPUT:
			var err error
			if opts.Reader != nil {
				_, err = opts.Reader.Read(b[:])
			} else if opts.sepInput != nil {
				b[0], err = opts.sepInput.ReadByte()
			} else {
				_, err = os.Stdin.Read(b[:])
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					if index < 0 {
						if negArr == nil {
							negArr = make([]uint8, opts.ArrSize)
							for ^index >= len(negArr) {
								negArr = append(negArr, 0)
								negArr = negArr[:cap(negArr)]
							}
						}
						if eofInRange {
							negArr[^index] = uint8(opts.OnEOF)
						}
					} else if eofInRange {
						arr[index] = uint8(opts.OnEOF)
					}
				} else {
					return errFunc(
						chunks[ip].err,
						"failed to read byte into array position #%v",
						index+1,
					)
				}
			} else {
				if index < 0 {
					if negArr == nil {
						negArr = make([]uint8, opts.ArrSize)
						for ^index >= len(negArr) {
							negArr = append(negArr, 0)
							negArr = negArr[:cap(negArr)]
						}
					}
					negArr[^index] = b[0]
				} else {
					arr[index] = b[0]
				}
			}
		case op_JUMP_FORWARD:
			if index < 0 {
				if negArr == nil {
					negArr = make([]uint8, opts.ArrSize)
					for ^index >= len(negArr) {
						negArr = append(negArr, 0)
						negArr = negArr[:cap(negArr)]
					}
				}
				if hasOpts && ip+2 < chunksLen &&
					(chunks[ip+1].op == op_DECREMENT ||
						(chunks[ip+1].op == op_INCREMENT_BY &&
							chunks[ip+1].val == -1)) &&
					chunks[ip+2].op == op_JUMP_BACKWARD { //Detect "[-]"
					negArr[^index] = 0
					ip += 2
				} else if negArr[^index] == 0 {
					ip = chunks[ip].val
				}
			} else if hasOpts && ip+2 < chunksLen &&
				(chunks[ip+1].op == op_DECREMENT ||
					(chunks[ip+1].op == op_INCREMENT_BY &&
						chunks[ip+1].val == -1)) &&
				chunks[ip+2].op == op_JUMP_BACKWARD { //Detect "[-]"
				arr[index] = 0
				ip += 2
			} else if arr[index] == 0 {
				ip = chunks[ip].val
			}
		case op_JUMP_BACKWARD:
			if index < 0 {
				if negArr == nil {
					negArr = make([]uint8, opts.ArrSize)
					for ^index >= len(negArr) {
						negArr = append(negArr, 0)
						negArr = negArr[:cap(negArr)]
					}
				}
				if negArr[^index] != 0 {
					ip = chunks[ip].val
				}
			} else if arr[index] != 0 {
				ip = chunks[ip].val
			}
		default: // should never happen
			return errFunc(
				chunks[ip].err,
				"unknown opcode number '%v'",
				chunks[ip].op,
			)
		}
		if arrBlock {
			if index >= len(arr) {
				index = len(arr) - 1
			} else if index < 0 {
				index = 0
			}
			continue
		}
		if arrWrapAround {
			if index >= len(arr) {
				index = 0
			} else if index < 0 {
				index = len(arr) - 1
			}
			continue
		}
		if index >= len(arr) {
			if opts.ArrHasStaticSize {
				return errFunc(
					chunks[ip].err,
					"array address exceeded array size of %v",
					opts.ArrSize,
				)
			}
			for index >= len(arr) {
				arr = append(arr, 0)
				arr = arr[:cap(arr)]
			}
		} else if index < 0 && negArr != nil {
			for ^index >= len(negArr) {
				negArr = append(negArr, 0)
				negArr = negArr[:cap(negArr)]
			}
		}
	}
	if !arrNegAddr && index < 0 {
		if ip <= 0 {
			return errFunc(1, NegativeArrErrMsg)
		}
		return errFunc(chunks[ip-1].err, NegativeArrErrMsg)
	}
	if loopTimes < 0 {
		for i := 0; i < len(arr); i++ {
			arr[i] = 0
		}
		if negArr != nil {
			for i := 0; i < len(negArr); i++ {
				negArr[i] = 0
			}
		}
		b[0] = 0
		ip, index = 0, 0
		goto loop_begin
	} else if loopTimes > 1 && loopTimesCount < loopTimes-1 {
		loopTimesCount++
		for i := 0; i < len(arr); i++ {
			arr[i] = 0
		}
		if negArr != nil {
			for i := 0; i < len(negArr); i++ {
				negArr[i] = 0
			}
		}
		b[0] = 0
		ip, index = 0, 0
		goto loop_begin
	}
	return nil
}

func parse(text string, opts *Options) ([]chunk, error) {
	chunks := make([]chunk, opts.ProgramSize)
	bracketsStack := make([]int, 0, 10)
	foundBfChar := false
	ptrCount := 0
	incCount := 0
	i := 0
	charPos := 1
	for _, c := range text {
		if opts.sepInput != nil {
			opts.sepInput.WriteRune(c)
			continue
		}
		if opts.ProgramHasStaticSize && i >= opts.ProgramSize {
			return nil, bfInterpretErr(
				charPos,
				"program size of %v exceeded",
				opts.ProgramSize,
			)
		}
		if opts.NoOptimizations {
			switch c {
			case '>':
				chunks[i].op = op_FORWARD
				foundBfChar = true
				goto foundBfChar_label
			case '<':
				chunks[i].op = op_BACKWARD
				foundBfChar = true
				goto foundBfChar_label
			case '+':
				chunks[i].op = op_INCREMENT
				foundBfChar = true
				goto foundBfChar_label
			case '-':
				chunks[i].op = op_DECREMENT
				foundBfChar = true
				goto foundBfChar_label
			}
		} else {
			switch c {
			case '>', '<', '+', '-':
				switch c {
				case '>', '<':
					if incCount != 0 {
						chunks[i].op = op_INCREMENT_BY
						chunks[i].val = incCount
						chunks[i].err = charPos
						i++
						if i >= len(chunks) && !opts.ProgramHasStaticSize {
							chunks = append(chunks, chunk{})
							chunks = chunks[:cap(chunks)]
						}
						incCount = 0
					}
					if c == '>' {
						ptrCount++
					} else {
						ptrCount--
					}
				default:
					if ptrCount != 0 {
						chunks[i].op = op_FOWARD_BY
						chunks[i].val = ptrCount
						chunks[i].err = charPos
						i++
						if i >= len(chunks) && !opts.ProgramHasStaticSize {
							chunks = append(chunks, chunk{})
							chunks = chunks[:cap(chunks)]
						}
						ptrCount = 0
					}
					if c == '+' {
						incCount++
					} else {
						incCount--
					}
				}
				charPos++
				continue
			default:
				if ptrCount != 0 {
					chunks[i].op = op_FOWARD_BY
					chunks[i].val = ptrCount
					chunks[i].err = charPos
					i++
					if i >= len(chunks) && !opts.ProgramHasStaticSize {
						chunks = append(chunks, chunk{})
						chunks = chunks[:cap(chunks)]
					}
					ptrCount = 0
				}
				if incCount != 0 {
					chunks[i].op = op_INCREMENT_BY
					chunks[i].val = incCount
					chunks[i].err = charPos
					i++
					if i >= len(chunks) && !opts.ProgramHasStaticSize {
						chunks = append(chunks, chunk{})
						chunks = chunks[:cap(chunks)]
					}
					incCount = 0
				}
			}
		}
		switch c {
		case '.':
			chunks[i].op = op_OUTPUT
			foundBfChar = true
		case ',':
			chunks[i].op = op_INPUT
			foundBfChar = true
		case '[':
			chunks[i].op = op_JUMP_FORWARD
			foundBfChar = true
			chunks[i].val = i
			bracketsStack = append(bracketsStack, i)
		case ']':
			if len(bracketsStack) == 0 {
				return nil, bfParseErr(
					charPos,
					"mismatched closing brackets",
				)
			}
			chunks[i].op = op_JUMP_BACKWARD
			foundBfChar = true
			last := len(bracketsStack) - 1
			ii := bracketsStack[last]
			bracketsStack = bracketsStack[:last]
			chunks[i].val = ii
			chunks[ii].val = i
		case '!':
			if opts.SepInput {
				opts.sepInput = new(bytes.Buffer)
			}
		}
	foundBfChar_label:
		if foundBfChar {
			chunks[i].err = charPos
			charPos++
			i++
			if i >= len(chunks) && !opts.ProgramHasStaticSize {
				chunks = append(chunks, chunk{})
				chunks = chunks[:cap(chunks)]
			}
			foundBfChar = false
		}
	}
	if len(bracketsStack) > 0 {
		return nil, bfParseErrAtEnd("mismatched opening brackets")
	}
	if !opts.NoOptimizations {
		if ptrCount != 0 {
			chunks[i].op = op_FOWARD_BY
			chunks[i].val = ptrCount
			chunks[i].err = charPos
			i++
			if i >= len(chunks) && !opts.ProgramHasStaticSize {
				chunks = append(chunks, chunk{})
				chunks = chunks[:cap(chunks)]
			}
		}
		if incCount != 0 {
			chunks[i].op = op_INCREMENT_BY
			chunks[i].val = incCount
			chunks[i].err = charPos
			i++
			if i >= len(chunks) && !opts.ProgramHasStaticSize {
				chunks = append(chunks, chunk{})
				chunks = chunks[:cap(chunks)]
			}
		}
	}
	return chunks, nil
}

func Execute(text string, opts *Options) error {
	if opts != nil {
		if opts.ArrSize < 0 {
			return bfErr("array size cannot be negative")
		}
		if opts.ProgramSize < 0 {
			return bfErr("program size cannot be negative")
		}
		if opts.Reader != nil && opts.SepInput {
			return bfErr("cannot enable SepInput with non-nil reader")
		}
		if opts.ArrHasStaticSize && opts.ArrHasNegAddress {
			return bfErr(
				"cannot enable both negative address and " +
					"static size flags at the same time",
			)
		}
		if opts.ArrSize == 0 {
			opts.ArrSize = DefaultArrSize
		}
		if opts.ProgramSize == 0 {
			opts.ProgramSize = DefaultProgramSize
		}
		if opts.ArrBlock {
			if opts.ArrHasNegAddress {
				return bfErr(
					"cannot enable both array block and " +
						"negative address flags at the same time",
				)
			}
			if opts.ArrHasStaticSize {
				return bfErr(
					"cannot enable both array block and " +
						"static size flags at the same time",
				)
			}
			if opts.ArrWrapAround {
				return bfErr(
					"cannot enable both array block and " +
						"wraparound flags at the same time",
				)
			}
		}
		if opts.ArrWrapAround {
			if opts.ArrHasNegAddress {
				return bfErr(
					"cannot enable both wraparound and " +
						"negative address flags at the same time",
				)
			}
			if opts.ArrHasStaticSize {
				return bfErr(
					"cannot enable both wraparound and " +
						"static size flags at the same time",
				)
			}
			if opts.ArrBlock {
				return bfErr(
					"cannot enable both wraparound and " +
						"array block flags at the same time",
				)
			}
		}
	} else {
		opts = Defopts()
	}
	chunks, err := parse(text, opts)
	if err != nil {
		return err
	}
	return interpret(chunks, opts)
}
