package metrics

import "sync"

type Ring struct {
	mu     sync.Mutex
	buf    []float32
	cap    int
	head   int
	filled int
}

func NewRing(capacity int) *Ring {
	if capacity <= 0 {
		capacity = 1
	}

	return &Ring{buf: make([]float32, capacity), cap: capacity}
}

func (r *Ring) Push(v float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.head] = v
	r.head = (r.head + 1) % r.cap
	if r.filled < r.cap {
		r.filled++
	}
}

func (r *Ring) Values() []float32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]float32, r.filled)
	start := (r.head - r.filled + r.cap) % r.cap
	for i := 0; i < r.filled; i++ {
		out[i] = r.buf[(start+i)%r.cap]
	}

	return out
}

func (r *Ring) Last() float32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.filled == 0 {
		return 0
	}
	idx := (r.head - 1 + r.cap) % r.cap

	return r.buf[idx]
}

func Stats(values []float32) (min, max, avg float32) {
	if len(values) == 0 {
		return 0, 0, 0
	}
	min, max = values[0], values[0]
	var sum float32
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}

	return min, max, sum / float32(len(values))
}
