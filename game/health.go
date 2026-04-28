package game

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// How often the monitor checks for stale heartbeats.
	healthCheckInterval = 2 * time.Second

	// A goroutine is considered stuck if its heartbeat
	// has not been updated within this duration.
	healthStaleThreshold = 5 * time.Second
)

////////////////////////////////////////////////////////////////////////////////

// Heartbeat tracks the last time a goroutine reported
// that it was alive. Updated atomically each iteration
// of the goroutine's main loop.
type Heartbeat struct {
	name    string
	lastNs  atomic.Int64
	alerted atomic.Bool
}

////////////////////////////////////////////////////////////////////////////////

// Beat updates the heartbeat to the current time.
func (h *Heartbeat) Beat() {

	h.lastNs.Store(time.Now().UnixNano())
	h.alerted.Store(false)
}

////////////////////////////////////////////////////////////////////////////////

// HealthMonitor watches a set of heartbeats and logs
// when any goroutine stops reporting.
type HealthMonitor struct {
	mu         sync.Mutex
	heartbeats []*Heartbeat
}

////////////////////////////////////////////////////////////////////////////////

// NewHealthMonitor creates an empty monitor.
func NewHealthMonitor() *HealthMonitor {

	return &HealthMonitor{}
}

////////////////////////////////////////////////////////////////////////////////

// Register adds a named heartbeat and returns it. The
// owning goroutine calls `Beat` on each loop iteration.
func (m *HealthMonitor) Register(name string) *Heartbeat {

	hb := &Heartbeat{name: name}
	hb.lastNs.Store(time.Now().UnixNano())

	m.mu.Lock()
	m.heartbeats = append(m.heartbeats, hb)
	m.mu.Unlock()

	return hb
}

////////////////////////////////////////////////////////////////////////////////

// Start launches the monitoring goroutine.
func (m *HealthMonitor) Start(group *errgroup.Group, ctx context.Context) {

	group.Go(func() error {
		logger.Dbg("starting health monitor")

		for {
			select {
			case <-ctx.Done():
				logger.Dbg("stopping health monitor")
				return nil
			case <-time.After(healthCheckInterval):
			}

			now := time.Now().UnixNano()

			m.mu.Lock()
			for _, hb := range m.heartbeats {
				lastNs := hb.lastNs.Load()
				age := time.Duration(now - lastNs)

				if age > healthStaleThreshold && !hb.alerted.Load() {
					hb.alerted.Store(true)
					logger.Err("goroutine stuck",
						logger.String("name", hb.name),
						logger.Duration("stale_for", age),
					)
				}
			}
			m.mu.Unlock()
		}
	})
}
