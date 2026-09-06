package metrics

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type GPUSample struct {
	Available   bool
	UtilPercent float64
	MemUsedMB   float64
	MemTotalMB  float64
}

var nvidiaSMIPath = resolveNvidiaSMI()

func resolveNvidiaSMI() string {
	if p, err := exec.LookPath("nvidia-smi"); err == nil {
		return p
	}
	if runtime.GOOS == "windows" {
		fallback := `C:\Program Files\NVIDIA Corporation\NVSMI\nvidia-smi.exe`
		if _, err := exec.LookPath(fallback); err == nil {
			return fallback
		}
	}

	return ""
}

func SampleGPU() GPUSample {
	if nvidiaSMIPath == "" {
		return GPUSample{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, nvidiaSMIPath,
		"--query-gpu=utilization.gpu,memory.used,memory.total",
		"--format=csv,noheader,nounits")
	hideWindow(cmd)

	out, err := cmd.Output()
	if err != nil {
		return GPUSample{}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return GPUSample{}
	}

	parts := strings.Split(lines[0], ",")
	if len(parts) != 3 {
		return GPUSample{}
	}

	util, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	used, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	total, err3 := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return GPUSample{}
	}

	return GPUSample{Available: true, UtilPercent: util, MemUsedMB: used, MemTotalMB: total}
}

func SampleProcessGPUUtil(pids []int32) map[int32]float64 {
	result := make(map[int32]float64)
	if nvidiaSMIPath == "" || len(pids) == 0 {
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, nvidiaSMIPath, "pmon", "-c", "1", "-s", "u")
	hideWindow(cmd)

	out, err := cmd.Output()
	if err != nil {
		return result
	}

	want := make(map[int32]bool, len(pids))
	for _, p := range pids {
		want[p] = true
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		pid64, err := strconv.ParseInt(fields[1], 10, 32)
		if err != nil {
			continue
		}

		pid := int32(pid64)
		if !want[pid] {
			continue
		}

		sm := 0.0
		if fields[3] != "-" {
			sm, _ = strconv.ParseFloat(fields[3], 64)
		}
		result[pid] += sm
	}

	return result
}
