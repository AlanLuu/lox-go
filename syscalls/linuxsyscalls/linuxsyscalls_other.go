//go:build !linux

package linuxsyscalls

import (
	"runtime"

	"github.com/AlanLuu/lox/loxerror"
)

const (
	GRND_INSECURE = -1
	GRND_NONBLOCK
	GRND_RANDOM
)

func unsupported(name string) error {
	osName := runtime.GOOS
	return loxerror.Error("'os." + name + "' is unsupported on " + osName + ".")
}

func Fallocate(fd int, mode uint32, off int64, len int64) error {
	return unsupported("fallocate")
}

func Getrandom(buf []byte, flags int) (int, error) {
	return -1, unsupported("getrandom")
}

func Setresgid(rgid int, egid int, sgid int) error {
	return unsupported("setresgid")
}

func Setresuid(ruid int, euid int, suid int) error {
	return unsupported("setresuid")
}
