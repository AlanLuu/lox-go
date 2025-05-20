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

type SysinfoResult struct {
	Uptime    int64
	Loads     [3]uint64
	Totalram  uint64
	Freeram   uint64
	Sharedram uint64
	Bufferram uint64
	Totalswap uint64
	Freeswap  uint64
	Procs     uint16
	Pad       uint16
	Totalhigh uint64
	Freehigh  uint64
	Unit      uint32
}

func Sysinfo() (SysinfoResult, error) {
	buf := unix.Sysinfo_t{}
	err := unix.Sysinfo(&buf)
	if err != nil {
		return SysinfoResult{}, nil
	}
	return SysinfoResult{
		Uptime:    buf.Uptime,
		Loads:     buf.Loads,
		Totalram:  buf.Totalram,
		Freeram:   buf.Freeram,
		Sharedram: buf.Sharedram,
		Bufferram: buf.Bufferram,
		Totalswap: buf.Totalswap,
		Freeswap:  buf.Freeswap,
		Procs:     buf.Procs,
		Pad:       buf.Pad,
		Totalhigh: buf.Totalhigh,
		Freehigh:  buf.Freehigh,
		Unit:      buf.Unit,
	}, nil
}
