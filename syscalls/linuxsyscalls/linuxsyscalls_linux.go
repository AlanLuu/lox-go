package linuxsyscalls

import "syscall"

func Setresgid(rgid int, egid int, sgid int) error {
	return syscall.Setresgid(rgid, egid, sgid)
}

func Setresuid(ruid int, euid int, suid int) error {
	return syscall.Setresuid(ruid, euid, suid)
}
