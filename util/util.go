package util

import (
	"os"
	"runtime"
	"runtime/debug"
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

func GitInfo() (hash string, modified bool) {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		foundHash := false
		foundModified := false
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				foundHash = true
				hash = setting.Value
			case "vcs.modified":
				foundModified = true
				modified = setting.Value != "false"
			}
			if foundHash && foundModified {
				return
			}
		}
	}
	return
}

func GitInfoShort() (hash string, modified bool) {
	const upper = 8
	if hash, modified = GitInfo(); len(hash) > upper {
		hash = hash[:upper]
	}
	return
}

func IntOrFloat(f float64) any {
	if FloatIsInt(f) {
		return int64(f)
	}
	return f
}

func IsAndroid() bool {
	return runtime.GOOS == "android"
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func IsLinuxOrAndroid() bool {
	return IsLinux() || IsAndroid()
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func StdinFromTerminal() bool {
	fd := os.Stdin.Fd()
	return InteractiveMode && (isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd))
}
