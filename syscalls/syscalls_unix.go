//go:build !windows

package syscalls

import (
	"bytes"
	"syscall"

	"golang.org/x/sys/unix"
)

func Chroot(path string) error {
	return syscall.Chroot(path)
}

func Mkfifo(path string, mode uint32) error {
	return syscall.Mkfifo(path, mode)
}

func Read(fd int, p []byte) (int, error) {
	return syscall.Read(fd, p)
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

type UnameResult struct {
	Sysname  string
	Nodename string
	Release  string
	Version  string
	Machine  string
}

func Uname() (UnameResult, error) {
	buf := unix.Utsname{}
	err := unix.Uname(&buf)
	if err != nil {
		return UnameResult{}, err
	}
	toString := func(arr []byte) string {
		index := bytes.IndexByte(arr, 0)
		if index == 0 {
			return ""
		} else if index > 0 {
			return string(arr[:index])
		}
		return string(arr)
	}
	return UnameResult{
		toString(buf.Sysname[:]),
		toString(buf.Nodename[:]),
		toString(buf.Release[:]),
		toString(buf.Version[:]),
		toString(buf.Machine[:]),
	}, nil
}

func Write(fd int, p []byte) (int, error) {
	return syscall.Write(fd, p)
}
