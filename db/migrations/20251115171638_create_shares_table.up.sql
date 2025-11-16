CREATE TABLE shares (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    shared_by_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resource_type TEXT NOT NULL CHECK(resource_type IN ('file', 'folder')),
    resource_id INTEGER NOT NULL,
    permission TEXT NOT NULL DEFAULT 'read' CHECK(permission IN ('read', 'write')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(shared_by_user_id, shared_with_user_id, resource_type, resource_id)
);

CREATE INDEX idx_shares_shared_with_user_id ON shares(shared_with_user_id);
CREATE INDEX idx_shares_shared_by_user_id ON shares(shared_by_user_id);
CREATE INDEX idx_shares_resource ON shares(resource_type, resource_id);