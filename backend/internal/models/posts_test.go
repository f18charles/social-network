package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

func TestMapDeletedPostResponseOmitsRetainedData(t *testing.T) {
	id := uuid.Must(uuid.FromString("6ff02ee4-d77a-4ff0-a946-e60e74bb9c53"))
	authorID := uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1"))
	deletedAt := time.Date(2026, 6, 16, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 16, 8, 30, 0, 0, time.UTC)
	imageURL := "/uploads/images/hike.gif"

	dto, err := MapPostResponse(&PostWithAuthor{
		Post: Post{
			ID:           id,
			UserID:       &authorID,
			Content:      "retained content that must not leak",
			ImageURL:     &imageURL,
			Privacy:      PostPrivacyPrivate,
			CommentCount: 4,
			LikeCount:    7,
			DislikeCount: 1,
			CreatedAt:    time.Date(2026, 6, 16, 8, 15, 0, 0, time.UTC),
			UpdatedAt:    &updatedAt,
			DeletedAt:    &deletedAt,
		},
		Author: &PublicUser{
			ID:        authorID,
			FirstName: "Amina",
			LastName:  "Njeri",
		},
		ViewerVote: ViewerVoteLike,
	})
	if err != nil {
		t.Fatalf("MapPostResponse returned error: %v", err)
	}

	assertJSONEqual(t, dto, `{"id":"6ff02ee4-d77a-4ff0-a946-e60e74bb9c53","deleted":true}`)
	assertJSONOmits(t, dto, "author", "content", "image_url", "privacy", "comment_count", "like_count", "dislike_count", "viewer_vote", "created_at", "updated_at")
}

func TestMapActivePostResponseIncludesOpenAPIFields(t *testing.T) {
	postID := uuid.Must(uuid.FromString("8a6de4c1-2f4a-4c52-a38f-61d76c9e7d11"))
	authorID := uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1"))
	nickname := "amina"
	avatar := "/uploads/avatars/amina.png"
	imageURL := "/uploads/images/hike.gif"
	createdAt := time.Date(2026, 6, 16, 8, 15, 0, 0, time.UTC)

	dto, err := MapPostResponse(&PostWithAuthor{
		Post: Post{
			ID:           postID,
			UserID:       &authorID,
			Content:      "First hike of the season was worth the early alarm.",
			ImageURL:     &imageURL,
			Privacy:      PostPrivacyPublic,
			CommentCount: 2,
			LikeCount:    14,
			DislikeCount: 0,
			CreatedAt:    createdAt,
		},
		Author: &PublicUser{
			ID:        authorID,
			FirstName: "Amina",
			LastName:  "Njeri",
			Nickname:  &nickname,
			Avatar:    &avatar,
		},
		ViewerVote: ViewerVoteLike,
	})
	if err != nil {
		t.Fatalf("MapPostResponse returned error: %v", err)
	}

	assertJSONEqual(t, dto, `{"id":"8a6de4c1-2f4a-4c52-a38f-61d76c9e7d11","deleted":false,"author":{"id":"6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1","first_name":"Amina","last_name":"Njeri","nickname":"amina","avatar":"/uploads/avatars/amina.png"},"group_id":null,"content":"First hike of the season was worth the early alarm.","image_url":"/uploads/images/hike.gif","privacy":"public","comment_count":2,"like_count":14,"dislike_count":0,"viewer_vote":"like","created_at":"2026-06-16T08:15:00Z","updated_at":null}`)
}

func TestMapCommentTreeTombstoneOmitsRetainedDataAndKeepsReplies(t *testing.T) {
	postID := uuid.Must(uuid.FromString("8a6de4c1-2f4a-4c52-a38f-61d76c9e7d11"))
	parentID := uuid.Must(uuid.FromString("623b1e60-babc-44db-a4a1-0c3d071a80a3"))
	replyID := uuid.Must(uuid.FromString("a1ac5c54-6cb6-41bc-a88f-55ad4fb3a6d2"))
	authorID := uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1"))
	deletedAt := time.Date(2026, 6, 16, 8, 25, 0, 0, time.UTC)
	nickname := "amina"
	avatar := "/uploads/avatars/amina.png"

	tree, err := MapCommentTree([]*CommentWithAuthor{
		{
			Comment: Comment{
				ID:           parentID,
				PostID:       postID,
				UserID:       &authorID,
				Content:      "retained comment that must not leak",
				LikeCount:    3,
				CreatedAt:    time.Date(2026, 6, 16, 8, 20, 0, 0, time.UTC),
				DeletedAt:    &deletedAt,
				DislikeCount: 1,
			},
			Author: &PublicUser{
				ID:        authorID,
				FirstName: "Amina",
				LastName:  "Njeri",
				Nickname:  &nickname,
				Avatar:    &avatar,
			},
			ViewerVote: ViewerVoteDislike,
		},
		{
			Comment: Comment{
				ID:              replyID,
				PostID:          postID,
				UserID:          &authorID,
				ParentCommentID: &parentID,
				Content:         "That view is unreal.",
				CreatedAt:       time.Date(2026, 6, 16, 8, 22, 0, 0, time.UTC),
			},
			Author: &PublicUser{
				ID:        authorID,
				FirstName: "Amina",
				LastName:  "Njeri",
				Nickname:  &nickname,
				Avatar:    &avatar,
			},
			ViewerVote: ViewerVoteNone,
		},
	})
	if err != nil {
		t.Fatalf("MapCommentTree returned error: %v", err)
	}

	assertJSONEqual(t, tree, `[{"id":"623b1e60-babc-44db-a4a1-0c3d071a80a3","deleted":true,"replies":[{"id":"a1ac5c54-6cb6-41bc-a88f-55ad4fb3a6d2","deleted":false,"post_id":"8a6de4c1-2f4a-4c52-a38f-61d76c9e7d11","parent_comment_id":"623b1e60-babc-44db-a4a1-0c3d071a80a3","author":{"id":"6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1","first_name":"Amina","last_name":"Njeri","nickname":"amina","avatar":"/uploads/avatars/amina.png"},"content":"That view is unreal.","image_url":null,"like_count":0,"dislike_count":0,"viewer_vote":"none","created_at":"2026-06-16T08:22:00Z","updated_at":null,"replies":[]}]}]`)
	assertJSONOmits(t, tree, "retained comment that must not leak", `"author":{"id":"623b1e60-babc-44db-a4a1-0c3d071a80a3"`, "privacy")
}

func TestMapCommentTreeLeafRepliesSerializeAsEmptyArray(t *testing.T) {
	postID := uuid.Must(uuid.FromString("8a6de4c1-2f4a-4c52-a38f-61d76c9e7d11"))
	commentID := uuid.Must(uuid.FromString("a1ac5c54-6cb6-41bc-a88f-55ad4fb3a6d2"))
	authorID := uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1"))

	tree, err := MapCommentTree([]*CommentWithAuthor{
		{
			Comment: Comment{
				ID:        commentID,
				PostID:    postID,
				UserID:    &authorID,
				Content:   "Leaf comment",
				CreatedAt: time.Date(2026, 6, 16, 8, 22, 0, 0, time.UTC),
			},
			Author: &PublicUser{
				ID:        authorID,
				FirstName: "Amina",
				LastName:  "Njeri",
			},
			ViewerVote: ViewerVoteNone,
		},
	})
	if err != nil {
		t.Fatalf("MapCommentTree returned error: %v", err)
	}

	body, err := json.Marshal(tree)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(body), `"replies":[]`) {
		t.Fatalf("expected leaf replies to serialize as [], got %s", body)
	}
	if strings.Contains(string(body), `"replies":null`) {
		t.Fatalf("expected no null replies, got %s", body)
	}
}

func assertJSONEqual(t *testing.T, got any, want string) {
	t.Helper()
	body, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if string(body) != want {
		t.Fatalf("unexpected JSON\n got: %s\nwant: %s", body, want)
	}
}

func assertJSONOmits(t *testing.T, got any, omitted ...string) {
	t.Helper()
	body, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	for _, value := range omitted {
		if strings.Contains(string(body), value) {
			t.Fatalf("JSON leaked %q: %s", value, body)
		}
	}
}
