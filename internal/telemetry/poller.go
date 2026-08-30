package telemetry

import (
	"context"
	"errors"
	"time"
)

type Snapshot struct {
	Time       time.Time
	Indicators *Indicators
	State      State
	Connected  bool
	Err        error
}

type Poller struct {
	client   *Client
	interval time.Duration
}

func NewPoller(client *Client, interval time.Duration) *Poller {
	if interval <= 0 {
		interval = 60 * time.Millisecond
	}

	return &Poller{client: client, interval: interval}
}

func (p *Poller) Run(ctx context.Context) <-chan Snapshot {
	out := make(chan Snapshot)
	go func() {
		defer close(out)
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				snap := p.poll(ctx)
				select {
				case out <- snap:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}

func (p *Poller) poll(ctx context.Context) Snapshot {
	snap := Snapshot{Time: time.Now()}

	ind, err := p.client.Indicators(ctx)
	switch {
	case err == nil:
		snap.Indicators = ind
		snap.Connected = true
	case errors.Is(err, ErrGameNotRunning):
	default:
		snap.Err = err
	}

	st, err := p.client.State(ctx)
	switch {
	case err == nil:
		snap.State = st
		snap.Connected = true
	case errors.Is(err, ErrGameNotRunning):
	default:
		if snap.Err == nil {
			snap.Err = err
		}
	}

	return snap
}
