package util

import "os"

var InteractiveMode = false

func FloatIsInt(f float64) bool {
	return f == float64(int64(f))
}

func StdinFromTerminal() bool {
	stat, _ := os.Stdin.Stat()
	return InteractiveMode && (stat.Mode()&os.ModeCharDevice) != 0
}
