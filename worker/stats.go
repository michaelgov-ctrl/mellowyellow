package worker

import (
	"fmt"

	"github.com/c9s/goprocinfo/linux"
)

type Stats struct {
	MemStats  *linux.MemInfo
	DiskStats *linux.Disk
	CpuStats  *linux.CPUStat
	LoadStats *linux.LoadAvg
	TaskCount int
	Errors    []error
}

func (s *Stats) MemTotalKb() uint64 {
	return s.MemStats.MemTotal
}

func (s *Stats) MemAvailableKb() uint64 {
	return s.MemStats.MemAvailable
}

func (s *Stats) MemUsedKb() uint64 {
	return s.MemTotalKb() - s.MemAvailableKb()
}

func (s *Stats) MemUsedPercent() uint64 {
	return s.MemAvailableKb() / s.MemTotalKb()
}

func (s *Stats) DiskTotal() uint64 {
	return s.DiskStats.All
}

func (s *Stats) DiskFree() uint64 {
	return s.DiskStats.Free
}

func (s *Stats) DiskUsed() uint64 {
	return s.DiskStats.Used
}

func (s *Stats) CpuUsage() float64 {
	idle := s.CpuStats.Idle + s.CpuStats.IOWait
	nonIdle := s.CpuStats.User + s.CpuStats.Nice + s.CpuStats.System + s.CpuStats.IRQ + s.CpuStats.SoftIRQ + s.CpuStats.Steal
	total := idle + nonIdle

	if total == 0 {
		return 0.00
	}

	return (float64(total) - float64(idle)) / float64(total)
}

func GetStats() *Stats {
	s := Stats{Errors: []error{}}

	m, err := GetMemoryInfo()
	if err != nil {
		s.Errors = append(s.Errors, err)
	}
	s.MemStats = m

	d, err := GetDiskInfo()
	if err != nil {
		s.Errors = append(s.Errors, err)
	}
	s.DiskStats = d

	stats, err := GetCpuStats()
	if err != nil {
		s.Errors = append(s.Errors, err)
	}
	s.CpuStats = stats

	l, err := GetLoadAvg()
	if err != nil {
		s.Errors = append(s.Errors, err)
	}
	s.LoadStats = l

	return &s
}

func GetMemoryInfo() (*linux.MemInfo, error) {
	p := "/proc/meminfo"
	m, err := linux.ReadMemInfo(p)
	if err != nil {
		return &linux.MemInfo{}, fmt.Errorf("error reading from %s", p)
	}

	return m, nil
}

func GetDiskInfo() (*linux.Disk, error) {
	p := "/"
	d, err := linux.ReadDisk(p)
	if err != nil {
		return &linux.Disk{}, fmt.Errorf("error reading from %s", p)
	}

	return d, nil
}

func GetCpuStats() (*linux.CPUStat, error) {
	p := "/proc/stat"
	s, err := linux.ReadStat(p)
	if err != nil {
		return &linux.CPUStat{}, fmt.Errorf("error reading from %s", p)
	}

	return &s.CPUStatAll, nil
}

func GetLoadAvg() (*linux.LoadAvg, error) {
	p := "/proc/loadavg"
	l, err := linux.ReadLoadAvg(p)
	if err != nil {
		return &linux.LoadAvg{}, fmt.Errorf("error reading from %s", p)
	}

	return l, nil
}
