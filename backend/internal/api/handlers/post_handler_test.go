package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/services"
)

func TestPostHandlerFeedReturnsTopLevelPagination(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	service := &fakePostService{
		homeResponse: samplePostListResponse(t, 20, 0, true),
	}
	handler := NewPostHandler(service)
	request := authenticatedRequest(http.MethodGet, "/api/posts", viewerID)
	recorder := httptest.NewRecorder()

	handler.Feed(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := response["pagination"]; !ok {
		t.Fatalf("expected top-level pagination, got %#v", response)
	}
	if _, ok := response["data"].([]any); !ok {
		t.Fatalf("expected top-level data array, got %#v", response["data"])
	}
}

func TestPostHandlerFeedRejectsInvalidGroupID(t *testing.T) {
	handler := NewPostHandler(&fakePostService{})
	request := authenticatedRequest(http.MethodGet, "/api/posts?group_id=not-a-uuid", uuid.Must(uuid.NewV4()))
	recorder := httptest.NewRecorder()

	handler.Feed(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestPostHandlerFeedMapsForbiddenGroupFeed(t *testing.T) {
	handler := NewPostHandler(&fakePostService{groupErr: services.ErrForbidden})
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000001"))
	request := authenticatedRequest(http.MethodGet, "/api/posts?group_id="+groupID.String(), uuid.Must(uuid.NewV4()))
	recorder := httptest.NewRecorder()

	handler.Feed(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestPostHandlerProfilePostsParsesPathAndMapsForbidden(t *testing.T) {
	handler := NewPostHandler(&fakePostService{profileErr: services.ErrForbidden})
	profileID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	request := authenticatedRequest(http.MethodGet, "/api/users/"+profileID.String()+"/posts", uuid.Must(uuid.NewV4()))
	request.SetPathValue("id", profileID.String())
	recorder := httptest.NewRecorder()

	handler.ProfilePosts(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestPostHandlerRejectsInvalidPagination(t *testing.T) {
	handler := NewPostHandler(&fakePostService{homeErr: services.ErrInvalidPagination})
	request := authenticatedRequest(http.MethodGet, "/api/posts?limit=101", uuid.Must(uuid.NewV4()))
	recorder := httptest.NewRecorder()

	handler.Feed(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func authenticatedRequest(method, target string, userID uuid.UUID) *http.Request {
	request := httptest.NewRequest(method, target, nil)
	user := &models.User{ID: userID, Email: "viewer@example.com"}
	ctx := context.WithValue(request.Context(), middleware.UserContextKey, user)
	return request.WithContext(ctx)
}

type fakePostService struct {
	homeResponse          *dto.PostListResponse
	profileResponse       *dto.PostListResponse
	groupResponse         *dto.PostListResponse
	createResponse        dto.PostResponse
	commentsResponse      *dto.CommentListResponse
	createCommentResponse dto.CommentResponse
	updateResponse        dto.PostResponse
	deleteResponse        dto.PostResponse
	updateCommentResponse dto.CommentResponse
	deleteCommentResponse dto.CommentResponse
	homeErr               error
	profileErr            error
	groupErr              error
	createErr             error
	commentsErr           error
	createCommentErr      error
	updateErr             error
	deleteErr             error
	updateCommentErr      error
	deleteCommentErr      error
}

func (s *fakePostService) CreatePost(ctx context.Context, req *dto.CreatePostRequest, authorID uuid.UUID) (dto.PostResponse, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	return s.createResponse, nil
}

func (s *fakePostService) GetSinglePost(ctx context.Context, postID string, viewerID *string) (dto.PostResponse, error) {
	return nil, nil
}

func (s *fakePostService) GetCommentsByPost(ctx context.Context, postID string, viewerID uuid.UUID, limit, offset int) (*dto.CommentListResponse, error) {
	if s.commentsErr != nil {
		return nil, s.commentsErr
	}
	return s.commentsResponse, nil
}

func (s *fakePostService) CreateComment(ctx context.Context, req *dto.CreateCommentRequest, authorID uuid.UUID) (dto.CommentResponse, error) {
	if s.createCommentErr != nil {
		return nil, s.createCommentErr
	}
	return s.createCommentResponse, nil
}

func (s *fakePostService) UpdatePost(ctx context.Context, postID string, req *dto.UpdatePostRequest, authorID uuid.UUID) (dto.PostResponse, error) {
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	return s.updateResponse, nil
}

func (s *fakePostService) DeletePost(ctx context.Context, postID string, authorID uuid.UUID) (dto.PostResponse, error) {
	if s.deleteErr != nil {
		return nil, s.deleteErr
	}
	return s.deleteResponse, nil
}

func (s *fakePostService) UpdateComment(ctx context.Context, commentID string, req *dto.UpdateCommentRequest, authorID uuid.UUID) (dto.CommentResponse, error) {
	if s.updateCommentErr != nil {
		return nil, s.updateCommentErr
	}
	return s.updateCommentResponse, nil
}

func (s *fakePostService) DeleteComment(ctx context.Context, commentID string, authorID uuid.UUID) (dto.CommentResponse, error) {
	if s.deleteCommentErr != nil {
		return nil, s.deleteCommentErr
	}
	return s.deleteCommentResponse, nil
}

func (s *fakePostService) GetHomeFeed(viewerID uuid.UUID, limit, offset int) (*dto.PostListResponse, error) {
	if s.homeErr != nil {
		return nil, s.homeErr
	}
	return s.homeResponse, nil
}

func (s *fakePostService) GetProfilePosts(profileUserID, viewerID uuid.UUID, limit, offset int) (*dto.PostListResponse, error) {
	if s.profileErr != nil {
		return nil, s.profileErr
	}
	return s.profileResponse, nil
}

func (s *fakePostService) GetGroupFeed(groupID, viewerID uuid.UUID, limit, offset int) (*dto.PostListResponse, error) {
	if s.groupErr != nil {
		return nil, s.groupErr
	}
	return s.groupResponse, nil
}

func samplePostListResponse(t *testing.T, limit, offset int, hasMore bool) *dto.PostListResponse {
	t.Helper()
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000009"))
	postID := uuid.Must(uuid.FromString("aaaaaaaa-0000-0000-0000-000000000001"))
	post, err := dto.MapPostResponse(&models.PostWithAuthor{
		Post: models.Post{
			ID:        postID,
			UserID:    &authorID,
			Content:   "post",
			Privacy:   models.PostPrivacyPublic,
			CreatedAt: time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC),
		},
		Author: &models.PublicUser{
			ID:        authorID,
			FirstName: "Amina",
			LastName:  "Njeri",
		},
	})
	if err != nil {
		t.Fatalf("MapPostResponse returned error: %v", err)
	}
	return &dto.PostListResponse{
		Status:  "success",
		Message: "Posts returned.",
		Data:    []dto.PostResponse{post},
		Errors:  nil,
		Pagination: dto.Pagination{
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		},
	}
}

var _ services.PostService = (*fakePostService)(nil)

func TestPostHandlerCreatePost(t *testing.T) {
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))

	t.Run("Create post with invalid privacy rejected", func(t *testing.T) {
		handler := NewPostHandler(&fakePostService{})
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("content", "hello")
		_ = writer.WriteField("privacy", "invalid-privacy")
		_ = writer.Close()

		request := httptest.NewRequest(http.MethodPost, "/api/posts", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		user := &models.User{ID: authorID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))

		recorder := httptest.NewRecorder()
		handler.CreatePost(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("Create post with empty content and no image rejected", func(t *testing.T) {
		handler := NewPostHandler(&fakePostService{})
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("content", "   ")
		_ = writer.Close()

		request := httptest.NewRequest(http.MethodPost, "/api/posts", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		user := &models.User{ID: authorID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))

		recorder := httptest.NewRecorder()
		handler.CreatePost(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("Create post success", func(t *testing.T) {
		postResponse, err := dto.MapPostResponse(&models.PostWithAuthor{
			Post: models.Post{
				ID:        uuid.Must(uuid.NewV4()),
				UserID:    &authorID,
				Content:   "hello",
				Privacy:   models.PostPrivacyPublic,
				CreatedAt: time.Now(),
			},
			Author: &models.PublicUser{
				ID:        authorID,
				FirstName: "Amina",
				LastName:  "Njeri",
			},
		})
		if err != nil {
			t.Fatalf("failed to map post: %v", err)
		}

		service := &fakePostService{
			createResponse: postResponse,
		}
		handler := NewPostHandler(service)

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("content", "hello")
		_ = writer.WriteField("privacy", "public")
		_ = writer.Close()

		request := httptest.NewRequest(http.MethodPost, "/api/posts", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		user := &models.User{ID: authorID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))

		recorder := httptest.NewRecorder()
		handler.CreatePost(recorder, request)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
		}

		var respEnvelope map[string]any
		if err := json.NewDecoder(recorder.Body).Decode(&respEnvelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if respEnvelope["status"] != "success" {
			t.Fatalf("expected status success, got %v", respEnvelope["status"])
		}
		data, ok := respEnvelope["data"].(map[string]any)
		if !ok {
			t.Fatalf("expected data object inside response, got %T", respEnvelope["data"])
		}
		if data["content"] != "hello" {
			t.Fatalf("expected content 'hello', got %v", data["content"])
		}
	})
}

func TestPostHandlerGetComments(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	postID := uuid.Must(uuid.FromString("bbbbbbbb-0000-0000-0000-000000000001"))

	t.Run("GetComments success", func(t *testing.T) {
		commentResponse := &dto.CommentListResponse{
			Status:  "success",
			Message: "Comments returned.",
			Data: []dto.CommentResponse{
				&dto.ActiveCommentResponse{
					ID:              uuid.Must(uuid.NewV4()),
					Deleted:         false,
					PostID:          postID,
					ParentCommentID: nil,
					Author: dto.PublicUserResponse{
						ID:        viewerID,
						FirstName: "John",
						LastName:  "Doe",
					},
					Content:      "Nice post!",
					LikeCount:    0,
					DislikeCount: 0,
					ViewerVote:   models.ViewerVoteNone,
					CreatedAt:    time.Now(),
					Replies:      []dto.CommentResponse{},
				},
			},
			Pagination: dto.Pagination{
				Limit:   10,
				Offset:  0,
				HasMore: false,
			},
		}

		service := &fakePostService{
			commentsResponse: commentResponse,
		}
		handler := NewPostHandler(service)

		request := authenticatedRequest(http.MethodGet, "/api/posts/"+postID.String()+"/comments", viewerID)
		request.SetPathValue("id", postID.String())
		recorder := httptest.NewRecorder()

		handler.GetComments(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
		}

		var respEnvelope map[string]any
		if err := json.NewDecoder(recorder.Body).Decode(&respEnvelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if respEnvelope["status"] != "success" {
			t.Fatalf("expected status success, got %v", respEnvelope["status"])
		}
		data, ok := respEnvelope["data"].([]any)
		if !ok {
			t.Fatalf("expected data to be an array, got %T", respEnvelope["data"])
		}
		if len(data) != 1 {
			t.Fatalf("expected 1 comment, got %d", len(data))
		}
	})

	t.Run("GetComments rejects invalid post ID", func(t *testing.T) {
		handler := NewPostHandler(&fakePostService{})
		request := authenticatedRequest(http.MethodGet, "/api/posts/not-a-uuid/comments", viewerID)
		request.SetPathValue("id", "not-a-uuid")
		recorder := httptest.NewRecorder()

		handler.GetComments(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("GetComments rejects invalid pagination parameter", func(t *testing.T) {
		handler := NewPostHandler(&fakePostService{})
		request := authenticatedRequest(http.MethodGet, "/api/posts/"+postID.String()+"/comments?limit=invalid", viewerID)
		request.SetPathValue("id", postID.String())
		recorder := httptest.NewRecorder()

		handler.GetComments(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("GetComments post not found returns 404", func(t *testing.T) {
		service := &fakePostService{
			commentsErr: services.ErrPostNotFound,
		}
		handler := NewPostHandler(service)
		request := authenticatedRequest(http.MethodGet, "/api/posts/"+postID.String()+"/comments", viewerID)
		request.SetPathValue("id", postID.String())
		recorder := httptest.NewRecorder()

		handler.GetComments(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
		}
	})

	t.Run("GetComments post forbidden returns 403", func(t *testing.T) {
		service := &fakePostService{
			commentsErr: services.ErrPostForbidden,
		}
		handler := NewPostHandler(service)
		request := authenticatedRequest(http.MethodGet, "/api/posts/"+postID.String()+"/comments", viewerID)
		request.SetPathValue("id", postID.String())
		recorder := httptest.NewRecorder()

		handler.GetComments(recorder, request)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
		}
	})
}

func TestPostHandlerCreateComment(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	postID := uuid.Must(uuid.FromString("bbbbbbbb-0000-0000-0000-000000000001"))
	commentID := uuid.Must(uuid.FromString("11111111-0000-0000-0000-000000000001"))

	t.Run("CreateComment success", func(t *testing.T) {
		commentResponse, _ := dto.MapCommentResponse(&models.CommentWithAuthor{
			Comment: models.Comment{
				ID:        commentID,
				PostID:    postID,
				UserID:    &viewerID,
				Content:   "Looks great!",
				CreatedAt: time.Now(),
			},
			Author: &models.PublicUser{
				ID:        viewerID,
				FirstName: "John",
				LastName:  "Doe",
			},
			ViewerVote: models.ViewerVoteNone,
		}, []dto.CommentResponse{})

		service := &fakePostService{
			createCommentResponse: commentResponse,
		}
		handler := NewPostHandler(service)

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("content", "Looks great!")
		_ = writer.Close()

		request := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/comments", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		user := &models.User{ID: viewerID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))
		request.SetPathValue("id", postID.String())

		recorder := httptest.NewRecorder()
		handler.CreateComment(recorder, request)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
		}

		var respEnvelope map[string]any
		if err := json.NewDecoder(recorder.Body).Decode(&respEnvelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if respEnvelope["status"] != "success" {
			t.Fatalf("expected status success, got %v", respEnvelope["status"])
		}
		data, ok := respEnvelope["data"].(map[string]any)
		if !ok {
			t.Fatalf("expected data object, got %T", respEnvelope["data"])
		}
		if data["content"] != "Looks great!" {
			t.Fatalf("expected content 'Looks great!', got %v", data["content"])
		}
	})

	t.Run("CreateComment rejects empty content and no image", func(t *testing.T) {
		handler := NewPostHandler(&fakePostService{})

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("content", "   ")
		_ = writer.Close()

		request := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/comments", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		user := &models.User{ID: viewerID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))
		request.SetPathValue("id", postID.String())

		recorder := httptest.NewRecorder()
		handler.CreateComment(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("CreateComment post or parent deleted returns 409", func(t *testing.T) {
		service := &fakePostService{
			createCommentErr: services.ErrPostOrCommentDeleted,
		}
		handler := NewPostHandler(service)

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("content", "looks great!")
		_ = writer.Close()

		request := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/comments", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		user := &models.User{ID: viewerID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))
		request.SetPathValue("id", postID.String())

		recorder := httptest.NewRecorder()
		handler.CreateComment(recorder, request)

		if recorder.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
		}
	})

	t.Run("CreateComment cross-post parent returns 400", func(t *testing.T) {
		service := &fakePostService{
			createCommentErr: services.ErrCrossPostParent,
		}
		handler := NewPostHandler(service)

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("content", "looks great!")
		_ = writer.WriteField("parent_comment_id", commentID.String())
		_ = writer.Close()

		request := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/comments", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		user := &models.User{ID: viewerID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))
		request.SetPathValue("id", postID.String())

		recorder := httptest.NewRecorder()
		handler.CreateComment(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})
}

func TestPostHandlerUpdatePost(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	postID := uuid.Must(uuid.FromString("bbbbbbbb-0000-0000-0000-000000000001"))

	t.Run("UpdatePost success", func(t *testing.T) {
		postResponse, _ := dto.MapPostResponse(&models.PostWithAuthor{
			Post: models.Post{
				ID:      postID,
				UserID:  &viewerID,
				Content: "Updated Content",
				Privacy: models.PostPrivacyPublic,
			},
			Author: &models.PublicUser{
				ID:        viewerID,
				FirstName: "John",
				LastName:  "Doe",
			},
			ViewerVote: models.ViewerVoteNone,
		})

		service := &fakePostService{
			updateResponse: postResponse,
		}
		handler := NewPostHandler(service)

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("content", "Updated Content")
		_ = writer.Close()

		request := httptest.NewRequest(http.MethodPatch, "/api/posts/"+postID.String(), &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		user := &models.User{ID: viewerID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))
		request.SetPathValue("id", postID.String())

		recorder := httptest.NewRecorder()
		handler.UpdatePost(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
		}

		var respEnvelope map[string]any
		if err := json.NewDecoder(recorder.Body).Decode(&respEnvelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if respEnvelope["status"] != "success" {
			t.Fatalf("expected status success, got %v", respEnvelope["status"])
		}
		data, ok := respEnvelope["data"].(map[string]any)
		if !ok {
			t.Fatalf("expected data object, got %T", respEnvelope["data"])
		}
		if data["content"] != "Updated Content" {
			t.Fatalf("expected content 'Updated Content', got %v", data["content"])
		}
	})

	t.Run("UpdatePost forbidden returns 403", func(t *testing.T) {
		service := &fakePostService{
			updateErr: services.ErrForbidden,
		}
		handler := NewPostHandler(service)

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("content", "Updated Content")
		_ = writer.Close()

		request := httptest.NewRequest(http.MethodPatch, "/api/posts/"+postID.String(), &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		user := &models.User{ID: viewerID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))
		request.SetPathValue("id", postID.String())

		recorder := httptest.NewRecorder()
		handler.UpdatePost(recorder, request)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
		}
	})
}

func TestPostHandlerDeletePost(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	postID := uuid.Must(uuid.FromString("bbbbbbbb-0000-0000-0000-000000000001"))

	t.Run("DeletePost success", func(t *testing.T) {
		service := &fakePostService{
			deleteResponse: &dto.DeletedPostResponse{
				ID:      postID,
				Deleted: true,
			},
		}
		handler := NewPostHandler(service)

		request := httptest.NewRequest(http.MethodDelete, "/api/posts/"+postID.String(), nil)
		user := &models.User{ID: viewerID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))
		request.SetPathValue("id", postID.String())

		recorder := httptest.NewRecorder()
		handler.DeletePost(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
		}

		var respEnvelope map[string]any
		if err := json.NewDecoder(recorder.Body).Decode(&respEnvelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if respEnvelope["status"] != "success" {
			t.Fatalf("expected status success, got %v", respEnvelope["status"])
		}
		data, ok := respEnvelope["data"].(map[string]any)
		if !ok {
			t.Fatalf("expected data object, got %T", respEnvelope["data"])
		}
		if data["deleted"] != true {
			t.Fatalf("expected deleted true, got %v", data["deleted"])
		}
	})

	t.Run("DeletePost forbidden returns 403", func(t *testing.T) {
		service := &fakePostService{
			deleteErr: services.ErrForbidden,
		}
		handler := NewPostHandler(service)

		request := httptest.NewRequest(http.MethodDelete, "/api/posts/"+postID.String(), nil)
		user := &models.User{ID: viewerID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))
		request.SetPathValue("id", postID.String())

		recorder := httptest.NewRecorder()
		handler.DeletePost(recorder, request)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
		}
	})
}

func TestPostHandlerUpdateComment(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	commentID := uuid.Must(uuid.FromString("11111111-0000-0000-0000-000000000001"))

	t.Run("UpdateComment success", func(t *testing.T) {
		service := &fakePostService{
			updateCommentResponse: &dto.ActiveCommentResponse{
				ID:      commentID,
				Content: "Updated content",
			},
		}
		handler := NewPostHandler(service)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("content", "Updated content")
		_ = writer.Close()

		request := httptest.NewRequest(http.MethodPatch, "/api/comments/"+commentID.String(), body)
		request.Header.Set("Content-Type", writer.FormDataContentType())

		user := &models.User{ID: viewerID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))
		request.SetPathValue("id", commentID.String())

		recorder := httptest.NewRecorder()
		handler.UpdateComment(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
		}
	})

	t.Run("UpdateComment not allowed method returns 405", func(t *testing.T) {
		handler := NewPostHandler(&fakePostService{})
		request := httptest.NewRequest(http.MethodPost, "/api/comments/"+commentID.String(), nil)
		recorder := httptest.NewRecorder()
		handler.UpdateComment(recorder, request)

		if recorder.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestPostHandlerDeleteComment(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	commentID := uuid.Must(uuid.FromString("11111111-0000-0000-0000-000000000001"))

	t.Run("DeleteComment success", func(t *testing.T) {
		service := &fakePostService{
			deleteCommentResponse: &dto.DeletedCommentResponse{
				ID:      commentID,
				Deleted: true,
			},
		}
		handler := NewPostHandler(service)

		request := httptest.NewRequest(http.MethodDelete, "/api/comments/"+commentID.String(), nil)
		user := &models.User{ID: viewerID, Email: "viewer@example.com"}
		request = request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))
		request.SetPathValue("id", commentID.String())

		recorder := httptest.NewRecorder()
		handler.DeleteComment(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
		}
	})
}
