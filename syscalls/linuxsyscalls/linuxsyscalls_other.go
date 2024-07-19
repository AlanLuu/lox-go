//go:build !linux

package linuxsyscalls

import (
	"runtime"

	"github.com/AlanLuu/lox/loxerror"
)

func unsupported(name string) error {
	osName := runtime.GOOS
	return loxerror.Error("'os." + name + "' is unsupported on " + osName + ".")
}

func Setresgid(rgid int, egid int, sgid int) error {
	return unsupported("setresgid")
}

func Setresuid(ruid int, euid int, suid int) error {
	return unsupported("setresuid")
}
