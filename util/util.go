package util

import "os"

var InteractiveMode = false

func StdinFromTerminal() bool {
	stat, _ := os.Stdin.Stat()
	return InteractiveMode && (stat.Mode()&os.ModeCharDevice) != 0
}
