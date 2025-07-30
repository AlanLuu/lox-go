//go:build linux && (386 || arm || mips || mipsle)

package linuxsyscalls

import "golang.org/x/sys/unix"

type SysinfoResult struct {
	Uptime    int32
	Loads     [3]uint32
	Totalram  uint32
	Freeram   uint32
	Sharedram uint32
	Bufferram uint32
	Totalswap uint32
	Freeswap  uint32
	Procs     uint16
	Pad       uint16
	Totalhigh uint32
	Freehigh  uint32
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
