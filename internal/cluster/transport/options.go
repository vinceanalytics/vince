package transport

import "time"

type Option func(m *Manager)

// WithHeartbeatTimeout configures the transport to not wait for than d for
// heartbeat to be executes by remote peer.
func WithHeartbeatTimeout(d time.Duration) Option {
	return func(m *Manager) {
		m.heartbeatTimeout = d
	}
}
