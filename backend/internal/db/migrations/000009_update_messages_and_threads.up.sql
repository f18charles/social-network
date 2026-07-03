-- Disable foreign keys temporarily to prevent constraint violation during recreate
PRAGMA foreign_keys = OFF;

-- Recreate dm_threads table
ALTER TABLE dm_threads RENAME TO old_dm_threads;

CREATE TABLE dm_threads (
    id TEXT PRIMARY KEY,
    user1_id TEXT, -- Nullable
    user2_id TEXT, -- Nullable
    last_message_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    FOREIGN KEY (user1_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (user2_id) REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE (user1_id, user2_id)
);

INSERT INTO dm_threads (id, user1_id, user2_id, last_message_at)
SELECT id, user1_id, user2_id, last_message_at FROM old_dm_threads;

DROP TABLE old_dm_threads;

-- Recreate messages table
ALTER TABLE messages RENAME TO old_messages;

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    sender_id TEXT, -- Nullable
    dm_thread_id TEXT,
    group_id TEXT,
    content TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT, -- New column for soft deletion
    message_type TEXT NOT NULL DEFAULT 'user', -- 'user' or 'system'
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (dm_thread_id) REFERENCES dm_threads(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

INSERT INTO messages (id, sender_id, dm_thread_id, group_id, content, created_at)
SELECT id, sender_id, dm_thread_id, group_id, content, created_at FROM old_messages;

DROP TABLE old_messages;

CREATE INDEX idx_messages_dm_thread ON messages(dm_thread_id);
CREATE INDEX idx_messages_group ON messages(group_id);

-- Re-enable foreign keys
PRAGMA foreign_keys = ON;
