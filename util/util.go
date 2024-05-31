package util

import (
	"os"
	"strconv"
)

var InteractiveMode = false

func CharCount(s string, c rune) int {
	inString := false
	count := 0
	for _, current := range s {
		switch current {
		case '"', '\'':
			inString = !inString
		}
		if !inString && current == c {
			count++
		}
	}
	return count
}

func FloatIsInt(f float64) bool {
	return f == float64(int64(f))
}

func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func IntOrFloat(f float64) any {
	if FloatIsInt(f) {
		return int64(f)
	}
	return f
}

func StdinFromTerminal() bool {
	stat, _ := os.Stdin.Stat()
	return InteractiveMode && (stat.Mode()&os.ModeCharDevice) != 0
}
