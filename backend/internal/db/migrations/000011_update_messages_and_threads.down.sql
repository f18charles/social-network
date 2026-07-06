PRAGMA foreign_keys = OFF;

-- Revert messages table
ALTER TABLE messages RENAME TO old_messages;

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

-- Note: Null sender_ids are filtered or ignored if we are reverting
INSERT INTO messages (id, sender_id, dm_thread_id, group_id, content, created_at)
SELECT id, COALESCE(sender_id, '00000000-0000-0000-0000-000000000000'), dm_thread_id, group_id, content, created_at 
FROM old_messages;

DROP TABLE old_messages;

CREATE INDEX idx_messages_dm_thread ON messages(dm_thread_id);
CREATE INDEX idx_messages_group ON messages(group_id);

-- Revert dm_threads table
ALTER TABLE dm_threads RENAME TO old_dm_threads;

CREATE TABLE dm_threads (
    id TEXT PRIMARY KEY,
    user1_id TEXT NOT NULL,
    user2_id TEXT NOT NULL,
    last_message_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    FOREIGN KEY (user1_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user2_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (user1_id, user2_id)
);

INSERT INTO dm_threads (id, user1_id, user2_id, last_message_at)
SELECT id, COALESCE(user1_id, '00000000-0000-0000-0000-000000000000'), COALESCE(user2_id, '00000000-0000-0000-0000-000000000000'), last_message_at 
FROM old_dm_threads;

DROP TABLE old_dm_threads;

PRAGMA foreign_keys = ON;
