package util

import (
	"os"
	"strconv"
)

var InteractiveMode = false

func CountBraces(s string) (int, int) {
	inString := false
	leftBraceCount := 0
	rightBraceCount := 0
	for _, current := range s {
		switch current {
		case '"', '\'':
			inString = !inString
			continue
		}
		if !inString {
			switch current {
			case '{':
				leftBraceCount++
			case '}':
				rightBraceCount++
			}
		}
	}
	return leftBraceCount, rightBraceCount
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
