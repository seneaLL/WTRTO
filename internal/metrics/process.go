package metrics

import (
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

type ProcTracker struct {
	procs map[int32]*process.Process
}

func NewProcTracker() *ProcTracker {
	return &ProcTracker{procs: make(map[int32]*process.Process)}
}

func (t *ProcTracker) Sample(pids []int32) (cpuPercent float64, memMB float64, ok bool) {
	for _, pid := range pids {
		p, exists := t.procs[pid]
		if !exists {
			np, err := process.NewProcess(pid)
			if err != nil {
				continue
			}
			t.procs[pid] = np
			p = np
		}

		c, err := p.Percent(0)
		if err != nil {
			delete(t.procs, pid)
			continue
		}

		mem, err := p.MemoryInfo()
		if err != nil {
			continue
		}

		cpuPercent += c
		memMB += float64(mem.RSS) / (1024 * 1024)
		ok = true
	}

	return cpuPercent, memMB, ok
}

func FindWarThunderProcess() (int32, bool) {
	for _, name := range warThunderProcessNames() {
		if pid, ok := FindProcessByName(name); ok {
			return pid, true
		}
	}

	return 0, false
}

func warThunderProcessNames() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"aces.exe"}
	case "linux":
		return []string{"aces.exe", "aces"}
	default:
		return []string{"aces"}
	}
}

func FindProcessByName(substr string) (int32, bool) {
	procs, err := process.Processes()
	if err != nil {
		return 0, false
	}

	substr = strings.ToLower(substr)
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(name), substr) {
			return p.Pid, true
		}
	}

	return 0, false
}
