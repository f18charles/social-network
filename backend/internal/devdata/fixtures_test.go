package devdata

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/db"
)

func TestSeedCreatesDeterministicE2EFixtures(t *testing.T) {
	database, imageDir, avatarDir := newFixtureDB(t)
	opts := testOptions(database, imageDir, avatarDir)

	if err := Seed(opts); err != nil {
		t.Fatalf("Seed returned error: %v", err)
	}

	status, err := Inspect(database)
	if err != nil {
		t.Fatalf("Inspect returned error: %v", err)
	}
	if status.Users != 6 || status.Posts != 23 || status.Comments != 30 || status.Groups != 1 || status.GroupMessages != 110 {
		t.Fatalf("status = %+v, want 6 users, 23 posts, 30 comments, 1 group, 110 group messages", status)
	}

	for idx, user := range fixtureUsers {
		assertPasswordWorks(t, database, user.Email)
		if idx < 5 {
			assertUserParticipation(t, database, user.ID)
		}
	}

	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE image_url IS NULL`, 11)
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE image_url LIKE '%.jpg'`, 4)
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE image_url LIKE '%.png'`, 3)
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE image_url LIKE '%.gif'`, 3)
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE privacy = 'public'`, 13)
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE privacy = 'almost_private'`, 5)
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE privacy = 'private'`, 5)
	assertScalar(t, database, `SELECT COUNT(*) FROM followers WHERE status = 'accepted'`, 30)
	assertScalar(t, database, `SELECT COUNT(*) FROM post_audiences`, 25)
	assertScalar(t, database, `SELECT COUNT(*) FROM users WHERE avatar LIKE '/uploads/avatars/e2e-fixture-%'`, 6)
	assertScalar(t, database, `SELECT COUNT(*) FROM post_votes`, len(fixturePostVotes))
	assertScalar(t, database, `SELECT COUNT(*) FROM comment_votes`, len(fixtureCommentVotes))
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE like_count > 0`, 1)
	assertScalar(t, database, `SELECT COUNT(*) FROM comments WHERE like_count > 0`, 1)
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE deleted_at IS NOT NULL AND user_id IS NULL`, 1)
	assertScalar(t, database, `SELECT COUNT(*) FROM comments WHERE deleted_at IS NOT NULL AND user_id IS NULL`, 3)
	assertScalar(t, database, `SELECT COUNT(*) FROM comments WHERE parent_comment_id IS NOT NULL`, 17)
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE dislike_count > 0`, 1)
	assertScalar(t, database, `SELECT COUNT(*) FROM comments WHERE dislike_count > 0`, 1)
	assertScalar(t, database, `SELECT COUNT(*) FROM groups WHERE id = '75000000-0000-0000-0000-000000000001'`, 1)
	assertScalar(t, database, `SELECT COUNT(*) FROM group_members WHERE group_id = '75000000-0000-0000-0000-000000000001' AND status = 'accepted'`, 3)
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE user_id = '71000000-0000-0000-0000-000000000006' AND group_id IS NULL`, 0)
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE user_id = '71000000-0000-0000-0000-000000000006' AND group_id = '75000000-0000-0000-0000-000000000001'`, 1)
	assertScalarEquals(t, database, `SELECT COUNT(*) FROM messages WHERE group_id = '75000000-0000-0000-0000-000000000001' AND id LIKE '74000000-%'`, 110)

	for _, name := range []string{"e2e-fixture-alex-trail.jpg", "e2e-fixture-alex-board.png", "e2e-fixture-bianca-cafe.gif"} {
		if _, err := os.Stat(filepath.Join(imageDir, name)); err != nil {
			t.Fatalf("expected media %s to exist: %v", name, err)
		}
	}
	for _, name := range []string{"e2e-fixture-avatar-alex.jpg", "e2e-fixture-avatar-bianca.jpg", "e2e-fixture-avatar-noor.jpg"} {
		if _, err := os.Stat(filepath.Join(avatarDir, name)); err != nil {
			t.Fatalf("expected avatar %s to exist: %v", name, err)
		}
	}
}

func TestTeardownRemovesOnlyFixtureData(t *testing.T) {
	database, imageDir, avatarDir := newFixtureDB(t)
	opts := testOptions(database, imageDir, avatarDir)

	if err := Seed(opts); err != nil {
		t.Fatalf("Seed returned error: %v", err)
	}
	_, err := database.Exec(`
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, dob, is_public,
			follower_count, following_count, created_at
		)
		VALUES ('99999999-0000-0000-0000-000000000001', 'real.user@example.test', 'hash', 'Real', 'User', '1990-01-01', 1, 0, 0, '2026-06-30T12:00:00Z');
		INSERT INTO posts (id, user_id, content, privacy, created_at)
		VALUES ('99999999-0000-0000-0000-000000000101', '99999999-0000-0000-0000-000000000001', 'real post', 'public', '2026-06-30T12:01:00Z');
		INSERT INTO posts (id, user_id, content, privacy, created_at)
		VALUES ('88888888-0000-0000-0000-000000000101', '71000000-0000-0000-0000-000000000001', 'extra fixture-owned post', 'public', '2026-06-30T12:02:00Z');
		INSERT INTO comments (id, post_id, user_id, content, created_at)
		VALUES ('88888888-0000-0000-0000-000000000201', '88888888-0000-0000-0000-000000000101', '99999999-0000-0000-0000-000000000001', 'real comment under fixture post', '2026-06-30T12:03:00Z');
		INSERT INTO comments (id, post_id, user_id, parent_comment_id, content, created_at)
		VALUES ('88888888-0000-0000-0000-000000000202', '88888888-0000-0000-0000-000000000101', '99999999-0000-0000-0000-000000000001', '88888888-0000-0000-0000-000000000201', 'real descendant under fixture post', '2026-06-30T12:04:00Z');
		INSERT INTO comments (id, post_id, user_id, content, created_at)
		VALUES ('88888888-0000-0000-0000-000000000203', '99999999-0000-0000-0000-000000000101', '71000000-0000-0000-0000-000000000002', 'fixture comment under real post', '2026-06-30T12:05:00Z');
		INSERT INTO groups (id, creator_id, title, description, created_at)
		VALUES ('88888888-0000-0000-0000-000000000301', '71000000-0000-0000-0000-000000000001', 'Fixture extra group', 'cleanup coverage', '2026-06-30T12:06:00Z');
		INSERT INTO group_members (group_id, user_id, status)
		VALUES ('88888888-0000-0000-0000-000000000301', '99999999-0000-0000-0000-000000000001', 'accepted');
		INSERT INTO events (id, group_id, creator_id, title, event_date, created_at)
		VALUES ('88888888-0000-0000-0000-000000000401', '88888888-0000-0000-0000-000000000301', '71000000-0000-0000-0000-000000000001', 'Fixture event', '2026-07-01T12:00:00Z', '2026-06-30T12:07:00Z');
		INSERT INTO event_rsvps (event_id, user_id, status)
		VALUES ('88888888-0000-0000-0000-000000000401', '99999999-0000-0000-0000-000000000001', 'going');
		INSERT INTO messages (id, sender_id, group_id, content, created_at)
		VALUES ('88888888-0000-0000-0000-000000000501', '71000000-0000-0000-0000-000000000001', '88888888-0000-0000-0000-000000000301', 'fixture group message', '2026-06-30T12:08:00Z');
		INSERT INTO notifications (id, user_id, type, source_id, group_id, created_at)
		VALUES ('88888888-0000-0000-0000-000000000601', '99999999-0000-0000-0000-000000000001', 'group_request', '71000000-0000-0000-0000-000000000001', '88888888-0000-0000-0000-000000000301', '2026-06-30T12:09:00Z')`)
	if err != nil {
		t.Fatalf("insert teardown coverage rows: %v", err)
	}

	if err := Teardown(opts); err != nil {
		t.Fatalf("Teardown returned error: %v", err)
	}

	status, err := Inspect(database)
	if err != nil {
		t.Fatalf("Inspect returned error: %v", err)
	}
	if status != (Status{}) {
		t.Fatalf("status = %+v, want empty fixture status", status)
	}
	assertScalar(t, database, `SELECT COUNT(*) FROM users WHERE email = 'real.user@example.test'`, 1)
	assertScalar(t, database, `SELECT COUNT(*) FROM posts WHERE id = '99999999-0000-0000-0000-000000000101'`, 1)
	assertScalarEquals(t, database, `SELECT COUNT(*) FROM comments WHERE post_id = '99999999-0000-0000-0000-000000000101'`, 0)
	assertScalarEquals(t, database, `SELECT COUNT(*) FROM groups WHERE id = '88888888-0000-0000-0000-000000000301'`, 0)
	assertScalarEquals(t, database, `SELECT COUNT(*) FROM notifications WHERE id = '88888888-0000-0000-0000-000000000601'`, 0)

	matches, err := filepath.Glob(filepath.Join(imageDir, "e2e-fixture-*"))
	if err != nil {
		t.Fatalf("glob fixture media: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("fixture media remained after teardown: %v", matches)
	}
	matches, err = filepath.Glob(filepath.Join(avatarDir, "e2e-fixture-*"))
	if err != nil {
		t.Fatalf("glob fixture avatars: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("fixture avatars remained after teardown: %v", matches)
	}
}

func TestSeedAndTeardownRejectProduction(t *testing.T) {
	database, imageDir, avatarDir := newFixtureDB(t)
	opts := testOptions(database, imageDir, avatarDir)
	opts.AppEnv = "production"

	if err := Seed(opts); err == nil {
		t.Fatal("Seed in production returned nil error")
	}
	if err := Teardown(opts); err == nil {
		t.Fatal("Teardown in production returned nil error")
	}
}

func newFixtureDB(t *testing.T) (*sql.DB, string, string) {
	t.Helper()
	tempDir := t.TempDir()
	database, err := db.InitDB(filepath.Join(tempDir, "devdata.db"), filepath.Join("..", "db", "migrations"))
	if err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database, filepath.Join(tempDir, "uploads", "images"), filepath.Join(tempDir, "uploads", "avatars")
}

func testOptions(database *sql.DB, imageDir, avatarDir string) Options {
	return Options{
		DB:        database,
		AppEnv:    "development",
		ImageDir:  imageDir,
		AvatarDir: avatarDir,
		CreatedAt: time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC),
		MediaFetcher: func(string) ([]byte, error) {
			return []byte("fixture media bytes"), nil
		},
	}
}

func assertPasswordWorks(t *testing.T, database *sql.DB, email string) {
	t.Helper()
	var hash string
	if err := database.QueryRow(`SELECT password_hash FROM users WHERE email = ?`, email).Scan(&hash); err != nil {
		t.Fatalf("select password hash for %s: %v", email, err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(SharedPassword)); err != nil {
		t.Fatalf("password for %s does not match SharedPassword: %v", email, err)
	}
}

func assertUserParticipation(t *testing.T, database *sql.DB, userID string) {
	t.Helper()
	assertScalar(t, database, `
		SELECT COUNT(*)
		FROM comments c
		JOIN posts p ON p.id = c.post_id
		WHERE c.user_id = ? AND c.parent_comment_id IS NULL AND p.user_id <> ?`, 1, userID, userID)
	assertScalar(t, database, `
		SELECT COUNT(*)
		FROM comments r
		JOIN posts p ON p.id = r.post_id
		WHERE r.user_id = ? AND r.parent_comment_id IS NOT NULL AND p.user_id = ?`, 1, userID, userID)
	assertScalar(t, database, `
		SELECT COUNT(*)
		FROM comments r
		JOIN comments parent ON parent.id = r.parent_comment_id
		JOIN posts p ON p.id = r.post_id
		WHERE r.user_id = ? AND parent.user_id <> ? AND p.user_id <> ?`, 1, userID, userID, userID)
}

func assertScalar(t *testing.T, database *sql.DB, query string, min int, args ...any) {
	t.Helper()
	var got int
	if err := database.QueryRow(query, args...).Scan(&got); err != nil {
		t.Fatalf("query %q returned error: %v", query, err)
	}
	if got < min {
		t.Fatalf("query %q = %d, want at least %d", query, got, min)
	}
}

func assertScalarEquals(t *testing.T, database *sql.DB, query string, want int, args ...any) {
	t.Helper()
	var got int
	if err := database.QueryRow(query, args...).Scan(&got); err != nil {
		t.Fatalf("query %q returned error: %v", query, err)
	}
	if got != want {
		t.Fatalf("query %q = %d, want %d", query, got, want)
	}
}
