//go:build !windows

package syscalls

import "syscall"

func Mkfifo(path string, mode uint32) error {
	return syscall.Mkfifo(path, mode)
}

func Setegid(egid int) error {
	return syscall.Setegid(egid)
}

func Seteuid(euid int) error {
	return syscall.Seteuid(euid)
}

func Setgid(gid int) error {
	return syscall.Setgid(gid)
}

func Setuid(uid int) error {
	return syscall.Setuid(uid)
}
