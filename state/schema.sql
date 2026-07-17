CREATE TABLE IF NOT EXISTS clients (
    mac TEXT PRIMARY KEY,
    ip  TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'pending',
    connected_at INTEGER,
    last_seen INTEGER,
    consent_given INTEGER DEFAULT 0,
    consent_timestamp INTEGER DEFAULT NULL,
    session_token TEXT DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS audit_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp   INTEGER NOT NULL,
    action      TEXT    NOT NULL,
    subject_mac TEXT    NOT NULL,
    requester_ip TEXT   NOT NULL,
    details     TEXT    DEFAULT ''
);
