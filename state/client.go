package state

import (
	"database/sql"
	"time"
)

const (
	StatePending       = "pending"
	StateAuthenticated = "authenticated"
)

type Client struct {
	MAC           string
	IP            string
	State         string
	ConnectedAt   int64
	LastSeen      int64
	BytesIn       int64
	BytesOut      int64
	MaxBytes      int64
	DownloadSpeed string
	UploadSpeed   string
}

// UpsertClient inserts or updates a client record.
func (db *DB) UpsertClient(mac, ip string) error {
	_, err := db.Conn.Exec(`
		INSERT INTO clients (mac, ip, last_seen)
		VALUES (?, ?, ?)
		ON CONFLICT(mac) DO UPDATE SET ip = ?, last_seen = ?`,
		mac, ip, time.Now().Unix(), ip, time.Now().Unix())
	return err
}

// GetClient retrieves a client by MAC address.
func (db *DB) GetClient(mac string) (*Client, error) {
	c := &Client{}
	var connAt, lastSeen sql.NullInt64
	err := db.Conn.QueryRow(`
		SELECT mac, ip, state, connected_at, last_seen, bytes_in, bytes_out, max_bytes, download_speed, upload_speed
		FROM clients WHERE mac = ?`, mac).
		Scan(&c.MAC, &c.IP, &c.State, &connAt, &lastSeen, &c.BytesIn, &c.BytesOut, &c.MaxBytes, &c.DownloadSpeed, &c.UploadSpeed)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.ConnectedAt = connAt.Int64
	c.LastSeen = lastSeen.Int64
	return c, nil
}

// ListAuthenticated returns all clients in 'authenticated' state.
func (db *DB) ListAuthenticated() ([]*Client, error) {
	rows, err := db.Conn.Query(`
		SELECT mac, ip, state, connected_at, last_seen, bytes_in, bytes_out, max_bytes, download_speed, upload_speed
		FROM clients WHERE state = ?`, StateAuthenticated)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*Client
	for rows.Next() {
		c := &Client{}
		var connAt, lastSeen sql.NullInt64
		if err := rows.Scan(&c.MAC, &c.IP, &c.State, &connAt, &lastSeen, &c.BytesIn, &c.BytesOut, &c.MaxBytes, &c.DownloadSpeed, &c.UploadSpeed); err != nil {
			return nil, err
		}
		c.ConnectedAt = connAt.Int64
		c.LastSeen = lastSeen.Int64
		clients = append(clients, c)
	}
	return clients, nil
}

// ListAll returns all clients in the database.
func (db *DB) ListAll() ([]*Client, error) {
	rows, err := db.Conn.Query(`
		SELECT mac, ip, state, connected_at, last_seen, bytes_in, bytes_out, max_bytes, download_speed, upload_speed
		FROM clients`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*Client
	for rows.Next() {
		c := &Client{}
		var connAt, lastSeen sql.NullInt64
		if err := rows.Scan(&c.MAC, &c.IP, &c.State, &connAt, &lastSeen, &c.BytesIn, &c.BytesOut, &c.MaxBytes, &c.DownloadSpeed, &c.UploadSpeed); err != nil {
			return nil, err
		}
		c.ConnectedAt = connAt.Int64
		c.LastSeen = lastSeen.Int64
		clients = append(clients, c)
	}
	return clients, nil
}

// UpdateTraffic updates the bytes_in and bytes_out counters for a client.
func (db *DB) UpdateTraffic(mac string, bytesIn, bytesOut int64) error {
	_, err := db.Conn.Exec(`
		UPDATE clients
		SET bytes_in = ?, bytes_out = ?
		WHERE mac = ?`,
		bytesIn, bytesOut, mac)
	return err
}

// UpdateBandwidth updates the bandwidth limits for a client.
func (db *DB) UpdateBandwidth(mac, downloadSpeed, uploadSpeed string) error {
	_, err := db.Conn.Exec(`
		UPDATE clients
		SET download_speed = ?, upload_speed = ?
		WHERE mac = ?`,
		downloadSpeed, uploadSpeed, mac)
	return err
}

// UpdateMaxBytes updates the max_bytes limit for a client.
func (db *DB) UpdateMaxBytes(mac string, maxBytes int64) error {
	_, err := db.Conn.Exec(`
		UPDATE clients
		SET max_bytes = ?
		WHERE mac = ?`,
		maxBytes, mac)
	return err
}
