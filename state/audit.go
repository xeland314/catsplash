package state

import "time"

// Audit action constants for LOPDP traceability.
const (
	AuditDataAccess   = "data_access"
	AuditDataDeletion = "data_deletion"
	AuditAuthSuccess  = "auth_success"
	AuditAuthDenied   = "auth_denied"
)

// AuditEvent represents a single LOPDP audit trail entry.
type AuditEvent struct {
	ID           int64
	Timestamp    int64
	Action       string
	SubjectMAC   string
	RequesterIP  string
	Details      string
}

// LogAuditEvent writes an entry to the audit_log table.
// subject_mac is stored as a masked (SHA-256 truncated) value, never in plaintext.
func (db *DB) LogAuditEvent(action, subjectMAC, requesterIP, details string) error {
	_, err := db.Conn.Exec(`
		INSERT INTO audit_log (timestamp, action, subject_mac, requester_ip, details)
		VALUES (?, ?, ?, ?, ?)`,
		time.Now().Unix(), action, subjectMAC, requesterIP, details)
	return err
}

// ListAuditEvents returns the most recent audit events, limited to n entries.
func (db *DB) ListAuditEvents(n int) ([]*AuditEvent, error) {
	if n <= 0 {
		n = 100
	}
	rows, err := db.Conn.Query(`
		SELECT id, timestamp, action, subject_mac, requester_ip, details
		FROM audit_log
		ORDER BY id DESC
		LIMIT ?`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*AuditEvent
	for rows.Next() {
		e := &AuditEvent{}
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.Action, &e.SubjectMAC, &e.RequesterIP, &e.Details); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

// PurgeOldAuditEvents deletes audit entries older than maxAge.
func (db *DB) PurgeOldAuditEvents(maxAge time.Duration) (int64, error) {
	cutoff := time.Now().Add(-maxAge).Unix()
	result, err := db.Conn.Exec(`DELETE FROM audit_log WHERE timestamp < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
