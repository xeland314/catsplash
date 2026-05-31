package state

import (
	"time"
)

// Reaper represents the session expiration engine.
type Reaper struct {
	db             *DB
	sessionTimeout int
	idleTimeout    int
	onExpire       func(mac, ip string) error
}

func NewReaper(db *DB, sessionTimeout, idleTimeout int, onExpire func(mac, ip string) error) *Reaper {
	return &Reaper{
		db:             db,
		sessionTimeout: sessionTimeout,
		idleTimeout:    idleTimeout,
		onExpire:       onExpire,
	}
}

// Start runs the reaper periodically.
func (r *Reaper) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		r.RunOnce()
	}
}

// RunOnce checks for expired sessions and invokes onExpire.
func (r *Reaper) RunOnce() {
	clients, err := r.db.ListAuthenticated()
	if err != nil {
		return
	}

	now := time.Now().Unix()
	for _, c := range clients {
		expired := false

		// Check absolute timeout
		if r.sessionTimeout > 0 && now-c.ConnectedAt >= int64(r.sessionTimeout) {
			expired = true
		}

		// Check idle timeout
		if !expired && r.idleTimeout > 0 && now-c.LastSeen >= int64(r.idleTimeout) {
			expired = true
		}

		if expired {
			if r.onExpire != nil {
				r.onExpire(c.MAC, c.IP)
			}
			r.db.Deauthenticate(c.MAC)
		}
	}
}
