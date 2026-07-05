ALTER TABLE posts ADD COLUMN heart_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE post_votes RENAME TO post_votes_old;

CREATE TABLE post_votes (
    post_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    vote TEXT NOT NULL CHECK (vote IN ('like', 'dislike', 'love')),
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT,

    PRIMARY KEY (post_id, user_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO post_votes (post_id, user_id, vote, created_at, updated_at)
SELECT post_id, user_id, vote, created_at, updated_at FROM post_votes_old;

DROP TABLE post_votes_old;

CREATE INDEX idx_post_votes_user ON post_votes(user_id);
CREATE INDEX idx_post_votes_vote ON post_votes(vote);

CREATE TABLE message_reactions (
    message_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    emoji TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),

    PRIMARY KEY (message_id, user_id, emoji),
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_message_reactions_message ON message_reactions(message_id);
