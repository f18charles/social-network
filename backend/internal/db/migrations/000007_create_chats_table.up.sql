CREATE TABLE dm_threads (
    id TEXT PRIMARY KEY,
    user1_id TEXT NOT NULL,
    user2_id TEXT NOT NULL,
    last_message_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    FOREIGN KEY (user1_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user2_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (user1_id, user2_id)
);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    sender_id TEXT NOT NULL,
    dm_thread_id TEXT,
    group_id TEXT,
    content TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (dm_thread_id) REFERENCES dm_threads(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE INDEX idx_messages_dm_thread ON messages(dm_thread_id);
CREATE INDEX idx_messages_group ON messages(group_id);
