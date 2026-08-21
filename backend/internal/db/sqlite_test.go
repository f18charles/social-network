package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const migrationsDir = "migrations"

func TestInitDBAppliesPostingMigrations(t *testing.T) {
	db := openMigratedTestDB(t)

	for _, table := range []string{
		"users",
		"sessions",
		"followers",
		"groups",
		"group_members",
		"posts",
		"post_audiences",
		"comments",
		"post_votes",
		"comment_votes",
	} {
		if !sqliteObjectExists(t, db, "table", table) {
			t.Fatalf("expected table %s to exist", table)
		}
	}

	for _, index := range []string{
		"idx_posts_created_at",
		"idx_posts_user_created",
		"idx_posts_group_created",
		"idx_comments_post_created",
		"idx_comments_parent_created",
		"idx_post_audiences_user",
		"idx_post_votes_user",
		"idx_comment_votes_user",
	} {
		if !sqliteObjectExists(t, db, "index", index) {
			t.Fatalf("expected index %s to exist", index)
		}
	}
}

func TestPostingMigrationsDownCleanly(t *testing.T) {
	db := openMigratedTestDB(t)

	applyMigrationFile(t, db, "000005_create_posting_tables.down.sql")
	applyMigrationFile(t, db, "000004_create_group_contract_tables.down.sql")

	for _, table := range []string{
		"posts",
		"post_audiences",
		"comments",
		"post_votes",
		"comment_votes",
		"groups",
		"group_members",
	} {
		if sqliteObjectExists(t, db, "table", table) {
			t.Fatalf("expected table %s to be removed", table)
		}
	}
}

func TestPostingMigrationsEnforceConstraintsDefaultsAndForeignKeys(t *testing.T) {
	db := openMigratedTestDB(t)

	authorID := "10000000-0000-0000-0000-000000000001"
	viewerID := "10000000-0000-0000-0000-000000000002"
	groupID := "20000000-0000-0000-0000-000000000001"
	postID := "30000000-0000-0000-0000-000000000001"
	parentCommentID := "40000000-0000-0000-0000-000000000001"
	replyCommentID := "40000000-0000-0000-0000-000000000002"

	insertMigrationTestUser(t, db, authorID, "author@example.com")
	insertMigrationTestUser(t, db, viewerID, "viewer@example.com")
	insertMigrationTestGroup(t, db, groupID, authorID)

	expectExecError(t, db, `
		INSERT INTO posts (id, user_id, content, privacy)
		VALUES ('30000000-0000-0000-0000-000000000099', ?, 'bad privacy', 'friends_only')
	`, authorID)

	if _, err := db.Exec(`
		INSERT INTO posts (id, user_id, group_id, content, privacy)
		VALUES (?, ?, ?, 'group post', 'private')
	`, postID, authorID, groupID); err != nil {
		t.Fatalf("insert post: %v", err)
	}

	var likeCount int
	var dislikeCount int
	var createdAt string
	var updatedAt sql.NullString
	var deletedAt sql.NullString
	if err := db.QueryRow(`
		SELECT like_count, dislike_count, created_at, updated_at, deleted_at
		FROM posts
		WHERE id = ?
	`, postID).Scan(&likeCount, &dislikeCount, &createdAt, &updatedAt, &deletedAt); err != nil {
		t.Fatalf("query post defaults: %v", err)
	}
	if likeCount != 0 || dislikeCount != 0 || createdAt == "" || updatedAt.Valid || deletedAt.Valid {
		t.Fatalf("post defaults = likes:%d dislikes:%d created:%q updated:%v deleted:%v", likeCount, dislikeCount, createdAt, updatedAt, deletedAt)
	}

	if _, err := db.Exec(`
		INSERT INTO post_audiences (post_id, user_id)
		VALUES (?, ?)
	`, postID, viewerID); err != nil {
		t.Fatalf("insert post audience: %v", err)
	}

	expectExecError(t, db, `
		INSERT INTO post_votes (post_id, user_id, vote)
		VALUES (?, ?, 'heart')
	`, postID, viewerID)
	if _, err := db.Exec(`
		INSERT INTO post_votes (post_id, user_id, vote)
		VALUES (?, ?, 'like')
	`, postID, viewerID); err != nil {
		t.Fatalf("insert post vote: %v", err)
	}
	expectExecError(t, db, `
		INSERT INTO post_votes (post_id, user_id, vote)
		VALUES (?, ?, 'like')
	`, postID, viewerID)
	expectExecError(t, db, `
		INSERT INTO post_votes (post_id, user_id, vote)
		VALUES (?, ?, 'dislike')
	`, postID, viewerID)

	if _, err := db.Exec(`
		INSERT INTO comments (id, post_id, user_id, content)
		VALUES (?, ?, ?, 'parent')
	`, parentCommentID, postID, authorID); err != nil {
		t.Fatalf("insert parent comment: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO comments (id, post_id, user_id, parent_comment_id, content)
		VALUES (?, ?, ?, ?, 'reply')
	`, replyCommentID, postID, authorID, parentCommentID); err != nil {
		t.Fatalf("insert reply comment: %v", err)
	}

	expectExecError(t, db, `
		DELETE FROM comments
		WHERE id = ?
	`, parentCommentID)

	if _, err := db.Exec(`
		UPDATE comments
		SET content = NULL, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, parentCommentID); err != nil {
		t.Fatalf("soft delete parent comment: %v", err)
	}

	var retainedParent sql.NullString
	if err := db.QueryRow(`
		SELECT parent_comment_id
		FROM comments
		WHERE id = ?
	`, replyCommentID).Scan(&retainedParent); err != nil {
		t.Fatalf("query reply parent: %v", err)
	}
	if !retainedParent.Valid || retainedParent.String != parentCommentID {
		t.Fatalf("reply parent after soft delete = %v, want %s", retainedParent, parentCommentID)
	}

	expectExecError(t, db, `
		INSERT INTO comment_votes (comment_id, user_id, vote)
		VALUES (?, ?, 'heart')
	`, replyCommentID, viewerID)
	if _, err := db.Exec(`
		INSERT INTO comment_votes (comment_id, user_id, vote)
		VALUES (?, ?, 'dislike')
	`, replyCommentID, viewerID); err != nil {
		t.Fatalf("insert comment vote: %v", err)
	}

	if _, err := db.Exec(`
		UPDATE posts
		SET content = NULL, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, postID); err != nil {
		t.Fatalf("soft delete post: %v", err)
	}

	var retainedGroupID string
	var retainedPrivacy string
	if err := db.QueryRow(`
		SELECT group_id, privacy
		FROM posts
		WHERE id = ?
	`, postID).Scan(&retainedGroupID, &retainedPrivacy); err != nil {
		t.Fatalf("query soft-deleted post: %v", err)
	}
	if retainedGroupID != groupID || retainedPrivacy != "private" {
		t.Fatalf("soft-deleted post retained group/privacy = %s/%s, want %s/private", retainedGroupID, retainedPrivacy, groupID)
	}

	if _, err := db.Exec(`DELETE FROM users WHERE id = ?`, authorID); err != nil {
		t.Fatalf("delete author: %v", err)
	}
	assertNullColumn(t, db, "posts", "user_id", "id", postID)
	assertNullColumn(t, db, "comments", "user_id", "id", parentCommentID)

	if _, err := db.Exec(`DELETE FROM users WHERE id = ?`, viewerID); err != nil {
		t.Fatalf("delete voter/audience user: %v", err)
	}
	assertRowCount(t, db, "post_audiences", 0)
	assertRowCount(t, db, "post_votes", 0)
	assertRowCount(t, db, "comment_votes", 0)
}

func TestSQLiteForeignKeysEnabledOnEveryConnection(t *testing.T) {
	db := openMigratedTestDB(t)
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)

	ctx := context.Background()
	conns := make([]*sql.Conn, 0, 4)
	for i := 0; i < 4; i++ {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("open conn %d: %v", i, err)
		}
		conns = append(conns, conn)
	}
	t.Cleanup(func() {
		for _, conn := range conns {
			_ = conn.Close()
		}
	})

	for i, conn := range conns {
		var enabled int
		if err := conn.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&enabled); err != nil {
			t.Fatalf("query foreign_keys on conn %d: %v", i, err)
		}
		if enabled != 1 {
			t.Fatalf("foreign_keys on conn %d = %d, want 1", i, enabled)
		}
	}
}

func openMigratedTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := InitDB(dbPath, migrationsDir, "", "")
	if err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func applyMigrationFile(t *testing.T, db *sql.DB, filename string) {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
	if err != nil {
		t.Fatalf("read migration %s: %v", filename, err)
	}
	if _, err := db.Exec(string(content)); err != nil {
		t.Fatalf("apply migration %s: %v", filename, err)
	}
}

func sqliteObjectExists(t *testing.T, db *sql.DB, objectType, name string) bool {
	t.Helper()

	var exists bool
	if err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM sqlite_master
			WHERE type = ? AND name = ?
		)
	`, objectType, name).Scan(&exists); err != nil {
		t.Fatalf("query sqlite object %s %s: %v", objectType, name, err)
	}
	return exists
}

func insertMigrationTestUser(t *testing.T, db *sql.DB, id, email string) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, dob
		)
		VALUES (?, ?, 'hash', 'Test', 'User', '1998-04-12')
	`, id, email)
	if err != nil {
		t.Fatalf("insert user %s: %v", id, err)
	}
}

func insertMigrationTestGroup(t *testing.T, db *sql.DB, id, creatorID string) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO groups (id, creator_id, title)
		VALUES (?, ?, 'Test Group')
	`, id, creatorID)
	if err != nil {
		t.Fatalf("insert group %s: %v", id, err)
	}
}

func expectExecError(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()

	if _, err := db.Exec(query, args...); err == nil {
		t.Fatalf("expected query to fail: %s", query)
	}
}

func assertNullColumn(t *testing.T, db *sql.DB, table, column, keyColumn, keyValue string) {
	t.Helper()

	var value sql.NullString
	query := "SELECT " + column + " FROM " + table + " WHERE " + keyColumn + " = ?"
	if err := db.QueryRow(query, keyValue).Scan(&value); err != nil {
		t.Fatalf("query %s.%s: %v", table, column, err)
	}
	if value.Valid {
		t.Fatalf("%s.%s = %q, want NULL", table, column, value.String)
	}
}

func assertRowCount(t *testing.T, db *sql.DB, table string, expected int) {
	t.Helper()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if count != expected {
		t.Fatalf("%s row count = %d, want %d", table, count, expected)
	}
}
