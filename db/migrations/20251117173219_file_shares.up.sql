CREATE TABLE file_shares(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL REFERENCES files(id),
    shared_with_id INTEGER NOT NULL REFERENCES users(id),
    permission TEXT NOT NULL CHECK ( permission IN ('read', 'write') ),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    UNIQUE (file_id, shared_with_id)
);

CREATE INDEX idx_file_shares_shared_with ON file_shares(shared_with_id);
CREATE INDEX idx_file_shares_expires ON file_shares(expires_at) WHERE expires_at IS NOT NULL;