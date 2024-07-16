//go:build !windows

package syscalls

import "syscall"

func Mkfifo(path string, mode uint32) error {
	return syscall.Mkfifo(path, mode)
}
