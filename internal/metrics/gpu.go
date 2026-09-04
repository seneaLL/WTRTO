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
