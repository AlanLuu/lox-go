package linuxsyscalls

import (
	"syscall"

	"golang.org/x/sys/unix"
)

const (
	GRND_INSECURE = unix.GRND_INSECURE
	GRND_NONBLOCK = unix.GRND_NONBLOCK
	GRND_RANDOM   = unix.GRND_RANDOM
)

func Fallocate(fd int, mode uint32, off int64, len int64) error {
	return unix.Fallocate(fd, mode, off, len)
}

func Getrandom(buf []byte, flags int) (int, error) {
	return unix.Getrandom(buf, flags)
}

func Setresgid(rgid int, egid int, sgid int) error {
	return syscall.Setresgid(rgid, egid, sgid)
}

func Setresuid(ruid int, euid int, suid int) error {
	return syscall.Setresuid(ruid, euid, suid)
}
