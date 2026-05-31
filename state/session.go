package state

import (
	"time"
)

// Authenticate marks a client as authenticated.
func (db *DB) Authenticate(mac, ip string) error {
	now := time.Now().Unix()
	_, err := db.Conn.Exec(`
		UPDATE clients
		SET state = ?, ip = ?, connected_at = ?, last_seen = ?
		WHERE mac = ?`,
		StateAuthenticated, ip, now, now, mac)
	return err
}

// Deauthenticate marks a client as pending.
func (db *DB) Deauthenticate(mac string) error {
	_, err := db.Conn.Exec(`
		UPDATE clients
		SET state = ?, connected_at = NULL
		WHERE mac = ?`,
		StatePending, mac)
	return err
}

// UpdateLastSeen updates the activity timestamp of a client.
func (db *DB) UpdateLastSeen(mac string) error {
	_, err := db.Conn.Exec(`
		UPDATE clients SET last_seen = ? WHERE mac = ?`,
		time.Now().Unix(), mac)
	return err
}
