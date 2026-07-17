CREATE TABLE IF NOT EXISTS clients (
    mac TEXT PRIMARY KEY,
    ip  TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'pending',
    connected_at INTEGER,
    last_seen INTEGER,
    consent_given INTEGER DEFAULT 0,
    consent_timestamp INTEGER DEFAULT NULL
);
