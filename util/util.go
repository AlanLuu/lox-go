package util

import (
	"os"
	"runtime"
	"strconv"

	"github.com/mattn/go-isatty"
)

var (
	DisableLoxCode  = false
	InteractiveMode = false
	UnsafeMode      = false
)

func CountBraces(s string) (int, int) {
	var quoteChr rune = 0
	var prevChr rune = 0
	leftBraceCount := 0
	rightBraceCount := 0
	for _, current := range s {
		switch current {
		case '"', '\'':
			if quoteChr == 0 {
				quoteChr = current
			} else if prevChr != '\\' && quoteChr == current {
				quoteChr = 0
			}
			prevChr = current
			continue
		}
		if quoteChr == 0 {
			switch current {
			case '{':
				leftBraceCount++
			case '}':
				rightBraceCount++
			}
		}
		prevChr = current
	}
	return leftBraceCount, rightBraceCount
}

func FloatIsInt(f float64) bool {
	return f == float64(int64(f))
}

func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func FormatFloatZero(f float64) string {
	if FloatIsInt(f) {
		return FormatFloat(f) + ".0"
	}
	return FormatFloat(f)
}

func IntOrFloat(f float64) any {
	if FloatIsInt(f) {
		return int64(f)
	}
	return f
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func StdinFromTerminal() bool {
	fd := os.Stdin.Fd()
	return InteractiveMode && (isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd))
}
