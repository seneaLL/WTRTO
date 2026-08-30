package metrics

import (
	"sync"
	"time"
)

const HistorySize = 90

type Sampler struct {
	FPS     *Ring
	AppCPU  *Ring
	AppMem  *Ring
	WTCPU   *Ring
	WTMem   *Ring
	GPUUtil *Ring

	ownPIDs    []int32
	appTracker *ProcTracker
	wtTracker  *ProcTracker

	mu            sync.RWMutex
	wtRunning     bool
	wtPID         int32
	gpuAvailable  bool
	gpuMemUsedMB  float64
	gpuMemTotalMB float64
}

func NewSampler(ownPIDs []int32) *Sampler {
	return &Sampler{
		FPS:        NewRing(HistorySize),
		AppCPU:     NewRing(HistorySize),
		AppMem:     NewRing(HistorySize),
		WTCPU:      NewRing(HistorySize),
		WTMem:      NewRing(HistorySize),
		GPUUtil:    NewRing(HistorySize),
		ownPIDs:    ownPIDs,
		appTracker: NewProcTracker(),
		wtTracker:  NewProcTracker(),
	}
}

func (s *Sampler) RecordFPS(fps float64) {
	s.FPS.Push(float32(fps))
}

func (s *Sampler) Run(stop <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var wtPID int32
	var haveWT bool

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if cpu, mem, ok := s.appTracker.Sample(s.ownPIDs); ok {
				s.AppCPU.Push(float32(cpu))
				s.AppMem.Push(float32(mem))
			}

			if !haveWT {
				if pid, found := FindWarThunderProcess(); found {
					wtPID = pid
					haveWT = true
				}
			}

			if haveWT {
				if cpu, mem, ok := s.wtTracker.Sample([]int32{wtPID}); ok {
					s.WTCPU.Push(float32(cpu))
					s.WTMem.Push(float32(mem))
				} else {
					haveWT = false
				}
			}

			s.mu.Lock()
			s.wtRunning = haveWT
			s.wtPID = wtPID
			s.mu.Unlock()

			gpu := SampleGPU()
			s.mu.Lock()
			s.gpuAvailable = gpu.Available
			s.gpuMemUsedMB = gpu.MemUsedMB
			s.gpuMemTotalMB = gpu.MemTotalMB
			s.mu.Unlock()

			if gpu.Available {
				s.GPUUtil.Push(float32(gpu.UtilPercent))
			}
		}
	}
}

func (s *Sampler) WTRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.wtRunning
}

func (s *Sampler) WTPID() (int32, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.wtPID, s.wtRunning
}

func (s *Sampler) GPUAvailable() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.gpuAvailable
}

func (s *Sampler) GPUMem() (usedMB, totalMB float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.gpuMemUsedMB, s.gpuMemTotalMB
}
