//go:build linux && (amd64 || arm64 || loong64 || mips64 || mips64le || ppc64 || ppc64le || riscv64 || s390x)

package linuxsyscalls

import "golang.org/x/sys/unix"

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
