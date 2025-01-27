package health

import (
	"sync/atomic"
)

const (
	degraded uint32 = 0
	healthy  uint32 = 1
)

// Health represents a system health status
type Health struct {
	v *atomic.Uint32
}

// New returns a [Health] object whose initial
// status is "unhealthy".
func New() *Health {
	value := new(atomic.Uint32)
	result := &Health{
		v: value,
	}

	return result
}

// Degrade demotes the health status to unhealthy.
func (h *Health) Degrade() {
	h.v.Store(degraded)
}

// Restore promotes the health status to healthy.
func (h *Health) Restore() {
	h.v.Store(healthy)
}

// Status returns the health status;
// true for healthy, otherwise false.
func (h *Health) Status() bool {
	return h.v.Load() == healthy
}
