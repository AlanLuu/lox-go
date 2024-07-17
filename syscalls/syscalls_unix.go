//go:build !windows

package syscalls

import "syscall"

func Mkfifo(path string, mode uint32) error {
	return syscall.Mkfifo(path, mode)
}

func Setgid(gid int) error {
	return syscall.Setgid(gid)
}

func Setuid(uid int) error {
	return syscall.Setuid(uid)
}
