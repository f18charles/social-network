// Package devdata owns explicit developer fixture data for local E2E runs.
package devdata

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/storage"
)

const (
	// SharedPassword is the password assigned to every seeded E2E user.
	SharedPassword = "Password123!"

	fixtureEmailDomain = "@example.test"
	fixtureEmailPrefix = "e2e."
	fixtureMediaPrefix = "e2e-fixture-"
)

// Options configures fixture seed and teardown operations.
type Options struct {
	DB           *sql.DB
	AppEnv       string
	ImageDir     string
	AvatarDir    string
	CreatedAt    time.Time
	MediaFetcher MediaFetcher
}

// Status summarizes fixture rows currently present in the database.
type Status struct {
	Users    int
	Posts    int
	Comments int
}

// MediaFetcher downloads one fixture media URL and returns its file bytes.
type MediaFetcher func(url string) ([]byte, error)

type fixtureUser struct {
	ID         string
	Email      string
	FirstName  string
	LastName   string
	Nickname   string
	AboutMe    string
	DOB        string
	IsPublic   bool
	AvatarName string
	AvatarURL  string
}

type fixturePost struct {
	ID         string
	UserIndex  int
	Content    string
	ImageName  string
	ImageURL   string
	Privacy    string
	MinutesAgo int
}

type fixtureComment struct {
	ID         string
	PostID     string
	UserIndex  int
	ParentID   string
	Content    string
	ImageName  string
	ImageURL   string
	MinutesAgo int
}

type fixtureVote struct {
	TargetID   string
	UserIndex  int
	Vote       string
	MinutesAgo int
}

var fixtureUsers = []fixtureUser{
	{ID: "71000000-0000-0000-0000-000000000001", Email: "e2e.alex@example.test", FirstName: "Alex", LastName: "Rivers", Nickname: "alex_e2e", AboutMe: "E2E fixture user focused on outdoor posts.", DOB: "1994-03-12", IsPublic: true, AvatarName: "e2e-fixture-avatar-alex.jpg", AvatarURL: "https://images.pexels.com/photos/220453/pexels-photo-220453.jpeg?auto=compress&cs=tinysrgb&w=256&h=256&fit=crop"},
	{ID: "71000000-0000-0000-0000-000000000002", Email: "e2e.bianca@example.test", FirstName: "Bianca", LastName: "Stone", Nickname: "bianca_e2e", AboutMe: "E2E fixture user focused on food and travel.", DOB: "1991-09-24", IsPublic: true, AvatarName: "e2e-fixture-avatar-bianca.jpg", AvatarURL: "https://images.pexels.com/photos/774909/pexels-photo-774909.jpeg?auto=compress&cs=tinysrgb&w=256&h=256&fit=crop"},
	{ID: "71000000-0000-0000-0000-000000000003", Email: "e2e.chidi@example.test", FirstName: "Chidi", LastName: "Okafor", Nickname: "chidi_e2e", AboutMe: "E2E fixture user focused on engineering notes.", DOB: "1989-11-08", IsPublic: false, AvatarName: "e2e-fixture-avatar-chidi.jpg", AvatarURL: "https://images.pexels.com/photos/2379004/pexels-photo-2379004.jpeg?auto=compress&cs=tinysrgb&w=256&h=256&fit=crop"},
	{ID: "71000000-0000-0000-0000-000000000004", Email: "e2e.dina@example.test", FirstName: "Dina", LastName: "Patel", Nickname: "dina_e2e", AboutMe: "E2E fixture user focused on design journals.", DOB: "1996-01-19", IsPublic: true, AvatarName: "e2e-fixture-avatar-dina.jpg", AvatarURL: "https://images.pexels.com/photos/415829/pexels-photo-415829.jpeg?auto=compress&cs=tinysrgb&w=256&h=256&fit=crop"},
	{ID: "71000000-0000-0000-0000-000000000005", Email: "e2e.elias@example.test", FirstName: "Elias", LastName: "Morgan", Nickname: "elias_e2e", AboutMe: "E2E fixture user focused on music updates.", DOB: "1992-07-30", IsPublic: false, AvatarName: "e2e-fixture-avatar-elias.jpg", AvatarURL: "https://images.pexels.com/photos/614810/pexels-photo-614810.jpeg?auto=compress&cs=tinysrgb&w=256&h=256&fit=crop"},
}

var fixturePosts = []fixturePost{
	{ID: "72000000-0000-0000-0000-000000000001", UserIndex: 0, Content: "Morning planning notes for the E2E social feed.", Privacy: "public", MinutesAgo: 1},
	{ID: "72000000-0000-0000-0000-000000000002", UserIndex: 0, Content: "JPEG trail image attached for media rendering checks.", ImageName: "e2e-fixture-alex-trail.jpg", ImageURL: "https://images.pexels.com/photos/1365425/pexels-photo-1365425.jpeg?auto=compress&cs=tinysrgb&w=900", Privacy: "public", MinutesAgo: 2},
	{ID: "72000000-0000-0000-0000-000000000003", UserIndex: 0, Content: "Follower-only checklist for visibility testing.", Privacy: "almost_private", MinutesAgo: 3},
	{ID: "72000000-0000-0000-0000-000000000004", UserIndex: 0, Content: "Private PNG board shared with selected followers.", ImageName: "e2e-fixture-alex-board.png", ImageURL: "https://raw.githubusercontent.com/github/explore/main/topics/go/go.png", Privacy: "private", MinutesAgo: 4},
	{ID: "72000000-0000-0000-0000-000000000005", UserIndex: 1, Content: "Coffee tasting notes with no image for text-only coverage.", Privacy: "public", MinutesAgo: 5},
	{ID: "72000000-0000-0000-0000-000000000006", UserIndex: 1, Content: "Animated GIF sample for feed media coverage.", ImageName: "e2e-fixture-bianca-cafe.gif", ImageURL: "https://media.giphy.com/media/3oEjI6SIIHBdRxXI40/giphy.gif", Privacy: "public", MinutesAgo: 6},
	{ID: "72000000-0000-0000-0000-000000000007", UserIndex: 1, Content: "Almost private dinner plan visible to followers.", Privacy: "almost_private", MinutesAgo: 7},
	{ID: "72000000-0000-0000-0000-000000000008", UserIndex: 1, Content: "Private JPEG menu preview.", ImageName: "e2e-fixture-bianca-menu.jpg", ImageURL: "https://images.pexels.com/photos/1640774/pexels-photo-1640774.jpeg?auto=compress&cs=tinysrgb&w=900", Privacy: "private", MinutesAgo: 8},
	{ID: "72000000-0000-0000-0000-000000000009", UserIndex: 2, Content: "Refactoring notes for repository integration tests.", Privacy: "public", MinutesAgo: 9},
	{ID: "72000000-0000-0000-0000-000000000010", UserIndex: 2, Content: "PNG terminal snapshot for upload rendering.", ImageName: "e2e-fixture-chidi-terminal.png", ImageURL: "https://raw.githubusercontent.com/github/explore/main/topics/javascript/javascript.png", Privacy: "public", MinutesAgo: 10},
	{ID: "72000000-0000-0000-0000-000000000011", UserIndex: 2, Content: "Follower-only API notes.", Privacy: "almost_private", MinutesAgo: 11},
	{ID: "72000000-0000-0000-0000-000000000012", UserIndex: 2, Content: "Private GIF debugging loop.", ImageName: "e2e-fixture-chidi-debug.gif", ImageURL: "https://media.giphy.com/media/26tn33aiTi1jkl6H6/giphy.gif", Privacy: "private", MinutesAgo: 12},
	{ID: "72000000-0000-0000-0000-000000000013", UserIndex: 3, Content: "Design review summary without media.", Privacy: "public", MinutesAgo: 13},
	{ID: "72000000-0000-0000-0000-000000000014", UserIndex: 3, Content: "JPEG moodboard tile for image coverage.", ImageName: "e2e-fixture-dina-moodboard.jpg", ImageURL: "https://images.pexels.com/photos/6444/pencil-typography-black-design.jpg?auto=compress&cs=tinysrgb&w=900", Privacy: "public", MinutesAgo: 14},
	{ID: "72000000-0000-0000-0000-000000000015", UserIndex: 3, Content: "Almost private prototype feedback.", Privacy: "almost_private", MinutesAgo: 15},
	{ID: "72000000-0000-0000-0000-000000000016", UserIndex: 3, Content: "Private PNG wireframe capture.", ImageName: "e2e-fixture-dina-wireframe.png", ImageURL: "https://raw.githubusercontent.com/github/explore/main/topics/react/react.png", Privacy: "private", MinutesAgo: 16},
	{ID: "72000000-0000-0000-0000-000000000017", UserIndex: 4, Content: "Playlist notes as text-only content.", Privacy: "public", MinutesAgo: 17},
	{ID: "72000000-0000-0000-0000-000000000018", UserIndex: 4, Content: "GIF stage lighting media sample.", ImageName: "e2e-fixture-elias-stage.gif", ImageURL: "https://media.giphy.com/media/l0HlHFRbmaZtBRhXG/giphy.gif", Privacy: "public", MinutesAgo: 18},
	{ID: "72000000-0000-0000-0000-000000000019", UserIndex: 4, Content: "Follower-only rehearsal notes.", Privacy: "almost_private", MinutesAgo: 19},
	{ID: "72000000-0000-0000-0000-000000000020", UserIndex: 4, Content: "Private JPEG album draft.", ImageName: "e2e-fixture-elias-album.jpg", ImageURL: "https://images.pexels.com/photos/164879/pexels-photo-164879.jpeg?auto=compress&cs=tinysrgb&w=900", Privacy: "private", MinutesAgo: 20},
}

var fixtureComments = []fixtureComment{
	{ID: "73000000-0000-0000-0000-000000000001", PostID: "72000000-0000-0000-0000-000000000005", UserIndex: 0, Content: "Alex commenting on Bianca's public coffee post.", MinutesAgo: 21},
	{ID: "73000000-0000-0000-0000-000000000002", PostID: "72000000-0000-0000-0000-000000000009", UserIndex: 1, Content: "Bianca commenting on Chidi's engineering post.", MinutesAgo: 22},
	{ID: "73000000-0000-0000-0000-000000000003", PostID: "72000000-0000-0000-0000-000000000013", UserIndex: 2, Content: "Chidi commenting on Dina's design review.", ImageName: "e2e-fixture-comment-chidi.png", ImageURL: "https://raw.githubusercontent.com/github/explore/main/topics/sqlite/sqlite.png", MinutesAgo: 23},
	{ID: "73000000-0000-0000-0000-000000000004", PostID: "72000000-0000-0000-0000-000000000017", UserIndex: 3, Content: "Dina commenting on Elias's playlist notes.", MinutesAgo: 24},
	{ID: "73000000-0000-0000-0000-000000000005", PostID: "72000000-0000-0000-0000-000000000001", UserIndex: 4, Content: "Elias commenting on Alex's planning notes.", MinutesAgo: 25},
	{ID: "73000000-0000-0000-0000-000000000006", PostID: "72000000-0000-0000-0000-000000000001", UserIndex: 1, Content: "Bianca asks Alex a follow-up on Alex's post.", MinutesAgo: 26},
	{ID: "73000000-0000-0000-0000-000000000007", PostID: "72000000-0000-0000-0000-000000000001", UserIndex: 0, ParentID: "73000000-0000-0000-0000-000000000006", Content: "Alex replies on their own post.", MinutesAgo: 27},
	{ID: "73000000-0000-0000-0000-000000000008", PostID: "72000000-0000-0000-0000-000000000005", UserIndex: 2, Content: "Chidi asks Bianca about the tasting method.", MinutesAgo: 28},
	{ID: "73000000-0000-0000-0000-000000000009", PostID: "72000000-0000-0000-0000-000000000005", UserIndex: 1, ParentID: "73000000-0000-0000-0000-000000000008", Content: "Bianca replies on their own post.", ImageName: "e2e-fixture-reply-bianca.gif", ImageURL: "https://media.giphy.com/media/13HgwGsXF0aiGY/giphy.gif", MinutesAgo: 29},
	{ID: "73000000-0000-0000-0000-000000000010", PostID: "72000000-0000-0000-0000-000000000009", UserIndex: 3, Content: "Dina asks Chidi for the repo link.", MinutesAgo: 30},
	{ID: "73000000-0000-0000-0000-000000000011", PostID: "72000000-0000-0000-0000-000000000009", UserIndex: 2, ParentID: "73000000-0000-0000-0000-000000000010", Content: "Chidi replies on their own post.", MinutesAgo: 31},
	{ID: "73000000-0000-0000-0000-000000000012", PostID: "72000000-0000-0000-0000-000000000013", UserIndex: 4, Content: "Elias asks Dina about the design constraints.", MinutesAgo: 32},
	{ID: "73000000-0000-0000-0000-000000000013", PostID: "72000000-0000-0000-0000-000000000013", UserIndex: 3, ParentID: "73000000-0000-0000-0000-000000000012", Content: "Dina replies on their own post.", MinutesAgo: 33},
	{ID: "73000000-0000-0000-0000-000000000014", PostID: "72000000-0000-0000-0000-000000000017", UserIndex: 0, Content: "Alex asks Elias about the next set list.", MinutesAgo: 34},
	{ID: "73000000-0000-0000-0000-000000000015", PostID: "72000000-0000-0000-0000-000000000017", UserIndex: 4, ParentID: "73000000-0000-0000-0000-000000000014", Content: "Elias replies on their own post.", MinutesAgo: 35},
	{ID: "73000000-0000-0000-0000-000000000016", PostID: "72000000-0000-0000-0000-000000000009", UserIndex: 0, ParentID: "73000000-0000-0000-0000-000000000002", Content: "Alex replies to Bianca on Chidi's post.", MinutesAgo: 36},
	{ID: "73000000-0000-0000-0000-000000000017", PostID: "72000000-0000-0000-0000-000000000013", UserIndex: 1, ParentID: "73000000-0000-0000-0000-000000000003", Content: "Bianca replies to Chidi on Dina's post.", MinutesAgo: 37},
	{ID: "73000000-0000-0000-0000-000000000018", PostID: "72000000-0000-0000-0000-000000000017", UserIndex: 2, ParentID: "73000000-0000-0000-0000-000000000004", Content: "Chidi replies to Dina on Elias's post.", MinutesAgo: 38},
	{ID: "73000000-0000-0000-0000-000000000019", PostID: "72000000-0000-0000-0000-000000000001", UserIndex: 3, ParentID: "73000000-0000-0000-0000-000000000005", Content: "Dina replies to Elias on Alex's post.", MinutesAgo: 39},
	{ID: "73000000-0000-0000-0000-000000000020", PostID: "72000000-0000-0000-0000-000000000005", UserIndex: 4, ParentID: "73000000-0000-0000-0000-000000000001", Content: "Elias replies to Alex on Bianca's post.", MinutesAgo: 40},
}

var fixturePostVotes = []fixtureVote{
	{TargetID: "72000000-0000-0000-0000-000000000001", UserIndex: 1, Vote: "like", MinutesAgo: 41},
	{TargetID: "72000000-0000-0000-0000-000000000001", UserIndex: 2, Vote: "like", MinutesAgo: 42},
	{TargetID: "72000000-0000-0000-0000-000000000002", UserIndex: 3, Vote: "like", MinutesAgo: 43},
	{TargetID: "72000000-0000-0000-0000-000000000006", UserIndex: 0, Vote: "like", MinutesAgo: 44},
	{TargetID: "72000000-0000-0000-0000-000000000009", UserIndex: 4, Vote: "like", MinutesAgo: 45},
	{TargetID: "72000000-0000-0000-0000-000000000014", UserIndex: 2, Vote: "dislike", MinutesAgo: 46},
	{TargetID: "72000000-0000-0000-0000-000000000018", UserIndex: 1, Vote: "like", MinutesAgo: 47},
}

var fixtureCommentVotes = []fixtureVote{
	{TargetID: "73000000-0000-0000-0000-000000000001", UserIndex: 1, Vote: "like", MinutesAgo: 48},
	{TargetID: "73000000-0000-0000-0000-000000000003", UserIndex: 3, Vote: "like", MinutesAgo: 49},
	{TargetID: "73000000-0000-0000-0000-000000000007", UserIndex: 4, Vote: "like", MinutesAgo: 50},
	{TargetID: "73000000-0000-0000-0000-000000000009", UserIndex: 0, Vote: "like", MinutesAgo: 51},
	{TargetID: "73000000-0000-0000-0000-000000000016", UserIndex: 2, Vote: "dislike", MinutesAgo: 52},
	{TargetID: "73000000-0000-0000-0000-000000000020", UserIndex: 3, Vote: "like", MinutesAgo: 53},
}

// Seed replaces the fixture dataset with a fresh deterministic copy.
func Seed(opts Options) error {
	if err := validateOptions(opts); err != nil {
		return err
	}
	if isProduction(opts.AppEnv) {
		return errors.New("refusing to seed developer data in production")
	}
	if err := Teardown(opts); err != nil {
		return err
	}
	if err := writeFixtureMedia(imageDir(opts), avatarDir(opts), mediaFetcher(opts)); err != nil {
		return err
	}
	if err := seedRows(opts); err != nil {
		_ = deleteFixtureMedia(imageDir(opts))
		_ = deleteFixtureMedia(avatarDir(opts))
		return err
	}
	return nil
}

// Teardown removes all fixture-owned rows and media files.
func Teardown(opts Options) error {
	if err := validateOptions(opts); err != nil {
		return err
	}
	if isProduction(opts.AppEnv) {
		return errors.New("refusing to teardown developer data in production")
	}

	tx, err := opts.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ids := userIDs()
	postIDs := fixturePostIDs()
	commentIDs := fixtureCommentIDs()

	if err := execIn(tx, "DELETE FROM comment_votes WHERE comment_id", commentIDs); err != nil {
		return err
	}
	if err := execIn(tx, "DELETE FROM post_votes WHERE post_id", postIDs); err != nil {
		return err
	}
	if err := execIn(tx, "DELETE FROM comments WHERE id", commentIDs); err != nil {
		return err
	}
	if err := execIn(tx, "DELETE FROM post_audiences WHERE post_id", postIDs); err != nil {
		return err
	}
	if err := execIn(tx, "DELETE FROM post_audiences WHERE user_id", ids); err != nil {
		return err
	}
	if err := execIn(tx, "DELETE FROM posts WHERE id", postIDs); err != nil {
		return err
	}
	if err := execIn(tx, "DELETE FROM followers WHERE follower_id", ids); err != nil {
		return err
	}
	if err := execIn(tx, "DELETE FROM followers WHERE followee_id", ids); err != nil {
		return err
	}
	if err := execIn(tx, "DELETE FROM sessions WHERE user_id", ids); err != nil {
		return err
	}
	if err := execIn(tx, "DELETE FROM users WHERE id", ids); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	if err := deleteFixtureMedia(imageDir(opts)); err != nil {
		return err
	}
	return deleteFixtureMedia(avatarDir(opts))
}

// Inspect returns counts for currently seeded fixture rows.
func Inspect(db *sql.DB) (Status, error) {
	if db == nil {
		return Status{}, errors.New("database is required")
	}
	var status Status
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email LIKE ?`, fixtureEmailPrefix+"%"+fixtureEmailDomain).Scan(&status.Users); err != nil {
		return Status{}, err
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM posts WHERE id LIKE '72000000-%'`).Scan(&status.Posts); err != nil {
		return Status{}, err
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM comments WHERE id LIKE '73000000-%'`).Scan(&status.Comments); err != nil {
		return Status{}, err
	}
	return status, nil
}

func seedRows(opts Options) error {
	tx, err := opts.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(SharedPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := opts.CreatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	for idx, user := range fixtureUsers {
		createdAt := now.Add(-time.Duration(100+idx) * time.Minute).Format(time.RFC3339)
		_, err := tx.Exec(`
			INSERT INTO users (
				id, email, password_hash, first_name, last_name, dob, avatar,
				nickname, about_me, is_public, follower_count, following_count, created_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?)`,
			user.ID,
			user.Email,
			string(passwordHash),
			user.FirstName,
			user.LastName,
			user.DOB,
			avatarURL(user.AvatarName),
			user.Nickname,
			user.AboutMe,
			boolInt(user.IsPublic),
			createdAt,
		)
		if err != nil {
			return err
		}
	}

	ids := userIDs()
	for _, followerID := range ids {
		for _, followeeID := range ids {
			if followerID == followeeID {
				continue
			}
			_, err := tx.Exec(
				`INSERT INTO followers (follower_id, followee_id, status, created_at) VALUES (?, ?, 'accepted', ?)`,
				followerID,
				followeeID,
				now.Add(-90*time.Minute).Format(time.RFC3339),
			)
			if err != nil {
				return err
			}
		}
	}

	for _, post := range fixturePosts {
		authorID := fixtureUsers[post.UserIndex].ID
		createdAt := now.Add(-time.Duration(post.MinutesAgo) * time.Minute).Format(time.RFC3339)
		_, err := tx.Exec(`
			INSERT INTO posts (
				id, user_id, group_id, content, image_url, privacy,
				comment_count, like_count, dislike_count, created_at, updated_at, deleted_at
			)
			VALUES (?, ?, NULL, ?, ?, ?, 0, 0, 0, ?, NULL, NULL)`,
			post.ID,
			authorID,
			post.Content,
			mediaURL(post.ImageName),
			post.Privacy,
			createdAt,
		)
		if err != nil {
			return err
		}
		if post.Privacy == "private" {
			for _, audienceID := range ids {
				if audienceID == authorID {
					continue
				}
				if _, err := tx.Exec(`INSERT INTO post_audiences (post_id, user_id) VALUES (?, ?)`, post.ID, audienceID); err != nil {
					return err
				}
			}
		}
	}

	for _, comment := range fixtureComments {
		userID := fixtureUsers[comment.UserIndex].ID
		createdAt := now.Add(-time.Duration(comment.MinutesAgo) * time.Minute).Format(time.RFC3339)
		_, err := tx.Exec(`
			INSERT INTO comments (
				id, post_id, user_id, parent_comment_id, content, image_url,
				like_count, dislike_count, created_at, deleted_at, updated_at
			)
			VALUES (?, ?, ?, ?, ?, ?, 0, 0, ?, NULL, NULL)`,
			comment.ID,
			comment.PostID,
			userID,
			nullableString(comment.ParentID),
			comment.Content,
			mediaURL(comment.ImageName),
			createdAt,
		)
		if err != nil {
			return err
		}
	}

	for _, vote := range fixturePostVotes {
		userID := fixtureUsers[vote.UserIndex].ID
		createdAt := now.Add(-time.Duration(vote.MinutesAgo) * time.Minute).Format(time.RFC3339)
		_, err := tx.Exec(
			`INSERT INTO post_votes (post_id, user_id, vote, created_at, updated_at) VALUES (?, ?, ?, ?, NULL)`,
			vote.TargetID,
			userID,
			vote.Vote,
			createdAt,
		)
		if err != nil {
			return err
		}
	}
	for _, vote := range fixtureCommentVotes {
		userID := fixtureUsers[vote.UserIndex].ID
		createdAt := now.Add(-time.Duration(vote.MinutesAgo) * time.Minute).Format(time.RFC3339)
		_, err := tx.Exec(
			`INSERT INTO comment_votes (comment_id, user_id, vote, created_at, updated_at) VALUES (?, ?, ?, ?, NULL)`,
			vote.TargetID,
			userID,
			vote.Vote,
			createdAt,
		)
		if err != nil {
			return err
		}
	}

	for _, post := range fixturePosts {
		if _, err := tx.Exec(`
			UPDATE posts
			SET
				comment_count = (SELECT COUNT(*) FROM comments WHERE post_id = ?),
				like_count = (SELECT COUNT(*) FROM post_votes WHERE post_id = ? AND vote = 'like'),
				dislike_count = (SELECT COUNT(*) FROM post_votes WHERE post_id = ? AND vote = 'dislike')
			WHERE id = ?`, post.ID, post.ID, post.ID, post.ID); err != nil {
			return err
		}
	}
	for _, comment := range fixtureComments {
		if _, err := tx.Exec(`
			UPDATE comments
			SET
				like_count = (SELECT COUNT(*) FROM comment_votes WHERE comment_id = ? AND vote = 'like'),
				dislike_count = (SELECT COUNT(*) FROM comment_votes WHERE comment_id = ? AND vote = 'dislike')
			WHERE id = ?`, comment.ID, comment.ID, comment.ID); err != nil {
			return err
		}
	}
	for _, id := range ids {
		_, err := tx.Exec(`
			UPDATE users
			SET
				follower_count = (SELECT COUNT(*) FROM followers WHERE followee_id = ? AND status = 'accepted'),
				following_count = (SELECT COUNT(*) FROM followers WHERE follower_id = ? AND status = 'accepted')
			WHERE id = ?`,
			id,
			id,
			id,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func validateOptions(opts Options) error {
	if opts.DB == nil {
		return errors.New("database is required")
	}
	return nil
}

func isProduction(appEnv string) bool {
	return strings.EqualFold(strings.TrimSpace(appEnv), "production")
}

func imageDir(opts Options) string {
	if opts.ImageDir != "" {
		return opts.ImageDir
	}
	return storage.ImageDir
}

func avatarDir(opts Options) string {
	if opts.AvatarDir != "" {
		return opts.AvatarDir
	}
	return storage.AvatarDir
}

func mediaFetcher(opts Options) MediaFetcher {
	if opts.MediaFetcher != nil {
		return opts.MediaFetcher
	}
	return downloadMedia
}

func execIn(tx *sql.Tx, prefix string, values []string) error {
	if len(values) == 0 {
		return nil
	}
	placeholders := make([]string, len(values))
	args := make([]any, len(values))
	for i, value := range values {
		placeholders[i] = "?"
		args[i] = value
	}
	_, err := tx.Exec(prefix+" IN ("+strings.Join(placeholders, ",")+")", args...)
	return err
}

func userIDs() []string {
	ids := make([]string, 0, len(fixtureUsers))
	for _, user := range fixtureUsers {
		ids = append(ids, user.ID)
	}
	return ids
}

func fixturePostIDs() []string {
	ids := make([]string, 0, len(fixturePosts))
	for _, post := range fixturePosts {
		ids = append(ids, post.ID)
	}
	return ids
}

func fixtureCommentIDs() []string {
	ids := make([]string, 0, len(fixtureComments))
	for _, comment := range fixtureComments {
		ids = append(ids, comment.ID)
	}
	return ids
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func mediaURL(name string) any {
	if name == "" {
		return nil
	}
	return storage.ImageURLPrefix + name
}

func avatarURL(name string) any {
	if name == "" {
		return nil
	}
	return storage.AvatarURLPrefix + name
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func writeFixtureMedia(imageDir, avatarDir string, fetch MediaFetcher) error {
	if err := os.MkdirAll(imageDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		return err
	}
	written := map[string]bool{}
	for _, user := range fixtureUsers {
		if err := writeFixtureMediaFile(avatarDir, user.AvatarName, user.AvatarURL, written, fetch); err != nil {
			return err
		}
	}
	for _, post := range fixturePosts {
		if err := writeFixtureMediaFile(imageDir, post.ImageName, post.ImageURL, written, fetch); err != nil {
			return err
		}
	}
	for _, comment := range fixtureComments {
		if err := writeFixtureMediaFile(imageDir, comment.ImageName, comment.ImageURL, written, fetch); err != nil {
			return err
		}
	}
	return nil
}

func writeFixtureMediaFile(dir, name, sourceURL string, written map[string]bool, fetch MediaFetcher) error {
	if name == "" || written[name] {
		return nil
	}
	payload, err := fetch(sourceURL)
	if err != nil {
		return fmt.Errorf("download fixture media %s: %w", name, err)
	}
	if len(payload) == 0 {
		return fmt.Errorf("download fixture media %s: empty response", name)
	}
	if len(payload) > storage.MaxImageSize {
		return fmt.Errorf("download fixture media %s: file too large", name)
	}
	if err := os.WriteFile(filepath.Join(dir, name), payload, 0o644); err != nil {
		return err
	}
	written[name] = true
	return nil
}

func deleteFixtureMedia(dir string) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), fixtureMediaPrefix) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func downloadMedia(sourceURL string) ([]byte, error) {
	client := http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "social-network-devdata/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, storage.MaxImageSize+1))
}
