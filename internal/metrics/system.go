package metrics

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

func SampleSystemCPU() (float64, bool) {
	pcts, err := cpu.Percent(0, false)
	if err != nil || len(pcts) == 0 {
		return 0, false
	}

	return pcts[0], true
}

type MemSample struct {
	UsedMB  float64
	TotalMB float64
}

func SampleSystemMem() (MemSample, bool) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return MemSample{}, false
	}

	const mb = 1024 * 1024

	return MemSample{
		UsedMB:  float64(vm.Used) / mb,
		TotalMB: float64(vm.Total) / mb,
	}, true
}
