package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/services"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/storage"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/utils"
)

// PostHandler handles authenticated post feed endpoints.
type PostHandler struct {
	postService services.PostService
}

// multipartFormMemory leaves room for form fields; storage.SaveImage enforces
// the 5 MiB image payload limit.
const multipartFormMemory = 10 << 20

// NewPostHandler creates a handler for post feed endpoints.
func NewPostHandler(ps services.PostService) *PostHandler {
	return &PostHandler{postService: ps}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		_ = utils.SendError(w, http.StatusBadRequest, "Content-Type must be multipart/form-data", nil)
		return
	}

	err := r.ParseMultipartForm(multipartFormMemory)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Failed to parse multipart form", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	privacyStr := strings.TrimSpace(r.FormValue("privacy"))
	groupIDStr := strings.TrimSpace(r.FormValue("group_id"))

	var rawAudience []string
	if vals, ok := r.Form["audience_ids"]; ok {
		rawAudience = vals
	}
	if len(rawAudience) == 1 {
		val := strings.TrimSpace(rawAudience[0])
		if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
			var parsed []string
			if err := json.Unmarshal([]byte(val), &parsed); err == nil {
				rawAudience = parsed
			}
		} else if strings.Contains(val, ",") {
			rawAudience = strings.Split(val, ",")
		}
	}

	var audienceIDs []uuid.UUID
	for _, idStr := range rawAudience {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		parsedID, err := uuid.FromString(idStr)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid input: malformed audience identifier", map[string]string{"audience_ids": "has an invalid format"})
			return
		}
		audienceIDs = append(audienceIDs, parsedID)
	}

	var groupID *uuid.UUID
	if groupIDStr != "" {
		gID, err := uuid.FromString(groupIDStr)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid input: malformed group identifier", map[string]string{"group_id": "has an invalid format"})
			return
		}
		groupID = &gID
	}

	imageFile, _, err := r.FormFile("image")
	hasImage := err == nil
	if hasImage {
		defer imageFile.Close()
	}

	if content == "" && !hasImage {
		_ = utils.SendError(w, http.StatusBadRequest, "Either content or image is required", map[string]string{"content": "is empty and no image uploaded"})
		return
	}

	privacy := models.PostPrivacy(privacyStr)
	if privacyStr == "" {
		privacy = models.PostPrivacyPublic
	} else if privacy != models.PostPrivacyPublic &&
		privacy != models.PostPrivacyAlmostPrivate &&
		privacy != models.PostPrivacyPrivate {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid privacy value", map[string]string{"privacy": "must be public, almost_private, or private"})
		return
	}

	var savedImagePath *string
	var success bool
	defer func() {
		if !success && savedImagePath != nil {
			_ = storage.DeleteImage(*savedImagePath)
		}
	}()

	if hasImage {
		path, err := storage.SaveImage(imageFile)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Failed to save image", map[string]string{"image": err.Error()})
			return
		}
		savedImagePath = &path
	}

	req := &models.CreatePostRequest{
		Content:     content,
		Privacy:     privacy,
		GroupID:     groupID,
		AudienceIDs: audienceIDs,
		ImageURL:    savedImagePath,
	}

	response, err := h.postService.CreatePost(r.Context(), req, currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrForbidden):
			_ = utils.SendError(w, http.StatusForbidden, "Forbidden: you do not have permission to post to this group or access is denied", nil)
		case errors.Is(err, services.ErrNotFollower):
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid audience: all members must be accepted followers", nil)
		case errors.Is(err, services.ErrInvalidPrivacy):
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid privacy value", nil)
		default:
			_ = utils.SendError(w, http.StatusInternalServerError, "Internal server error", nil)
		}
		return
	}

	success = true
	_ = utils.SendSuccess(w, http.StatusCreated, "Post created successfully", response)
}

func (h *PostHandler) GetSinglePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	postID := r.PathValue("id")

	if _, err := uuid.FromString(postID); err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "shared_validation_error: malformed id", nil)
		return
	}

	viewerID := h.extractViewerIDFromContext(r)

	payload, err := h.postService.GetSinglePost(ctx, postID, viewerID)
	if err != nil {
		if errors.Is(err, services.ErrPostNotFound) {
			_ = utils.SendError(w, http.StatusNotFound, "Post not found", nil)
			return
		}
		if errors.Is(err, services.ErrPostForbidden) {
			_ = utils.SendError(w, http.StatusForbidden, "You do not have access to this post", nil)
			return
		}
		_ = utils.SendError(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}
	_ = utils.SendSuccess(w, http.StatusOK, "Post retrieved successfully", payload)
}

func (h *PostHandler) extractViewerIDFromContext(r *http.Request) *string {
	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		return nil
	}
	userIDStr := currentUser.ID.String()
	return &userIDStr
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		_ = utils.SendError(w, http.StatusBadRequest, "Content-Type must be multipart/form-data", nil)
		return
	}

	err := r.ParseMultipartForm(multipartFormMemory)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Failed to parse multipart form", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	postIDStr := r.PathValue("id")
	if _, err := uuid.FromString(postIDStr); err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "shared_validation_error: malformed id", nil)
		return
	}

	var contentPtr *string
	if r.MultipartForm != nil {
		if vals, ok := r.MultipartForm.Value["content"]; ok && len(vals) > 0 {
			val := vals[0]
			contentPtr = &val
		}
	}

	var privacyPtr *models.PostPrivacy
	if r.MultipartForm != nil {
		if vals, ok := r.MultipartForm.Value["privacy"]; ok && len(vals) > 0 {
			val := models.PostPrivacy(strings.TrimSpace(vals[0]))
			privacyPtr = &val
		}
	}

	var audienceIDs []uuid.UUID
	if r.MultipartForm != nil {
		if vals, ok := r.MultipartForm.Value["audience_ids"]; ok {
			var rawAudience []string = vals
			if len(rawAudience) == 1 {
				val := strings.TrimSpace(rawAudience[0])
				if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
					var parsed []string
					if err := json.Unmarshal([]byte(val), &parsed); err == nil {
						rawAudience = parsed
					}
				} else if strings.Contains(val, ",") {
					rawAudience = strings.Split(val, ",")
				}
			}

			for _, idStr := range rawAudience {
				idStr = strings.TrimSpace(idStr)
				if idStr == "" {
					continue
				}
				parsedID, err := uuid.FromString(idStr)
				if err != nil {
					_ = utils.SendError(w, http.StatusBadRequest, "Invalid input: malformed audience identifier", map[string]string{"audience_ids": "has an invalid format"})
					return
				}
				audienceIDs = append(audienceIDs, parsedID)
			}
		}
	}

	removeImage := false
	if r.MultipartForm != nil {
		if vals, ok := r.MultipartForm.Value["remove_image"]; ok && len(vals) > 0 {
			removeImage, _ = strconv.ParseBool(vals[0])
		}
	}

	imageFile, _, err := r.FormFile("image")
	hasImage := err == nil
	if hasImage {
		defer imageFile.Close()
	}

	var savedImagePath *string
	var success bool
	defer func() {
		if !success && savedImagePath != nil {
			_ = storage.DeleteImage(*savedImagePath)
		}
	}()

	if hasImage {
		path, err := storage.SaveImage(imageFile)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Failed to save image", map[string]string{"image": err.Error()})
			return
		}
		savedImagePath = &path
	}

	req := &models.UpdatePostRequest{
		Content:     contentPtr,
		Privacy:     privacyPtr,
		AudienceIDs: audienceIDs,
		ImageURL:    savedImagePath,
		RemoveImage: removeImage,
	}

	response, err := h.postService.UpdatePost(r.Context(), postIDStr, req, currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrPostNotFound):
			_ = utils.SendError(w, http.StatusNotFound, "Post not found", nil)
		case errors.Is(err, services.ErrForbidden):
			_ = utils.SendError(w, http.StatusForbidden, "Forbidden: you do not have permission to edit this post", nil)
		case errors.Is(err, services.ErrPostOrCommentDeleted):
			_ = utils.SendError(w, http.StatusConflict, "Post or selected parent comment is deleted", nil)
		case errors.Is(err, services.ErrNotFollower):
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid audience: all members must be accepted followers", nil)
		default:
			_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		}
		return
	}

	success = true
	_ = utils.SendSuccess(w, http.StatusOK, "Post updated successfully", response)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	postIDStr := r.PathValue("id")
	if _, err := uuid.FromString(postIDStr); err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "shared_validation_error: malformed id", nil)
		return
	}

	response, err := h.postService.DeletePost(r.Context(), postIDStr, currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrPostNotFound):
			_ = utils.SendError(w, http.StatusNotFound, "Post not found", nil)
		case errors.Is(err, services.ErrForbidden):
			_ = utils.SendError(w, http.StatusForbidden, "Forbidden: you do not have permission to delete this post", nil)
		default:
			_ = utils.SendError(w, http.StatusInternalServerError, "Internal server error", nil)
		}
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Post deleted successfully", response)
}

// Feed returns the authenticated user's home feed or a group feed.
func (h *PostHandler) Feed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	limit, offset, err := parseFeedPagination(r)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid pagination", map[string]string{"pagination": err.Error()})
		return
	}

	groupIDParam := r.URL.Query().Get("group_id")
	if groupIDParam != "" {
		groupID, err := uuid.FromString(groupIDParam)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid input", map[string]string{"group_id": "has an invalid format"})
			return
		}
		response, err := h.postService.GetGroupFeed(groupID, currentUser.ID, limit, offset)
		h.writeFeedResponse(w, response, err)
		return
	}

	response, err := h.postService.GetHomeFeed(currentUser.ID, limit, offset)
	h.writeFeedResponse(w, response, err)
}

// ProfilePosts returns posts visible on the selected user's profile.
func (h *PostHandler) ProfilePosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	profileID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid input", map[string]string{"id": "has an invalid format"})
		return
	}

	limit, offset, err := parseFeedPagination(r)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid pagination", map[string]string{"pagination": err.Error()})
		return
	}

	response, err := h.postService.GetProfilePosts(profileID, currentUser.ID, limit, offset)
	h.writeFeedResponse(w, response, err)
}

func (h *PostHandler) writeFeedResponse(w http.ResponseWriter, response any, err error) {
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidPagination):
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid pagination", nil)
		case errors.Is(err, services.ErrForbidden):
			_ = utils.SendError(w, http.StatusForbidden, "Forbidden", nil)
		case isNotFoundError(err):
			_ = utils.SendError(w, http.StatusNotFound, "Not found", nil)
		default:
			_ = utils.SendError(w, http.StatusInternalServerError, "Internal server error", nil)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *PostHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	postIDStr := r.PathValue("id")
	if _, err := uuid.FromString(postIDStr); err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "shared_validation_error: malformed id", nil)
		return
	}

	limit, offset, err := parseFeedPagination(r)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid pagination", map[string]string{"pagination": err.Error()})
		return
	}

	response, err := h.postService.GetCommentsByPost(r.Context(), postIDStr, currentUser.ID, limit, offset)
	if err != nil {
		if errors.Is(err, services.ErrPostNotFound) {
			_ = utils.SendError(w, http.StatusNotFound, "Post not found", nil)
			return
		}
		if errors.Is(err, services.ErrPostForbidden) {
			_ = utils.SendError(w, http.StatusForbidden, "You do not have access to this post's comments", nil)
			return
		}
		_ = utils.SendError(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *PostHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		_ = utils.SendError(w, http.StatusBadRequest, "Content-Type must be multipart/form-data", nil)
		return
	}

	err := r.ParseMultipartForm(multipartFormMemory)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Failed to parse multipart form", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	postIDStr := r.PathValue("id")
	postID, err := uuid.FromString(postIDStr)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "shared_validation_error: malformed id", nil)
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	parentIDStr := strings.TrimSpace(r.FormValue("parent_comment_id"))

	var parentCommentID *uuid.UUID
	if parentIDStr != "" {
		pID, err := uuid.FromString(parentIDStr)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid input: malformed parent comment identifier", map[string]string{"parent_comment_id": "has an invalid format"})
			return
		}
		parentCommentID = &pID
	}

	imageFile, _, err := r.FormFile("image")
	hasImage := err == nil
	if hasImage {
		defer imageFile.Close()
	}

	if content == "" && !hasImage {
		_ = utils.SendError(w, http.StatusBadRequest, "Either content or image is required", map[string]string{"content": "is empty and no image uploaded"})
		return
	}

	var savedImagePath *string
	var success bool
	defer func() {
		if !success && savedImagePath != nil {
			_ = storage.DeleteImage(*savedImagePath)
		}
	}()

	if hasImage {
		path, err := storage.SaveImage(imageFile)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Failed to save image", map[string]string{"image": err.Error()})
			return
		}
		savedImagePath = &path
	}

	req := &models.CreateCommentRequest{
		PostID:          postID,
		ParentCommentID: parentCommentID,
		Content:         content,
		ImageURL:        savedImagePath,
	}

	response, err := h.postService.CreateComment(r.Context(), req, currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrPostNotFound):
			_ = utils.SendError(w, http.StatusNotFound, "Post not found", nil)
		case errors.Is(err, services.ErrPostForbidden):
			_ = utils.SendError(w, http.StatusForbidden, "You do not have access to this post", nil)
		case errors.Is(err, services.ErrPostOrCommentDeleted):
			_ = utils.SendError(w, http.StatusConflict, "Post or selected parent comment is deleted", nil)
		case errors.Is(err, services.ErrCrossPostParent):
			_ = utils.SendError(w, http.StatusBadRequest, "Parent comment belongs to a different post", nil)
		case errors.Is(err, services.ErrCommentNotFound):
			_ = utils.SendError(w, http.StatusBadRequest, "Parent comment not found", nil)
		default:
			_ = utils.SendError(w, http.StatusInternalServerError, "Internal server error", nil)
		}
		return
	}

	success = true
	_ = utils.SendSuccess(w, http.StatusCreated, "Comment created successfully", response)
}

func parseFeedPagination(r *http.Request) (int, int, error) {
	limit, err := parseOptionalInt(r.URL.Query().Get("limit"))
	if err != nil {
		return 0, 0, err
	}
	offset, err := parseOptionalInt(r.URL.Query().Get("offset"))
	if err != nil {
		return 0, 0, err
	}
	return limit, offset, nil
}

func parseOptionalInt(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func isNotFoundError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}

func (h *PostHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		_ = utils.SendError(w, http.StatusBadRequest, "Content-Type must be multipart/form-data", nil)
		return
	}

	err := r.ParseMultipartForm(multipartFormMemory)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Failed to parse multipart form", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	commentIDStr := r.PathValue("id")
	if _, err := uuid.FromString(commentIDStr); err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "shared_validation_error: malformed id", nil)
		return
	}

	var contentPtr *string
	if r.MultipartForm != nil {
		if vals, ok := r.MultipartForm.Value["content"]; ok && len(vals) > 0 {
			val := vals[0]
			contentPtr = &val
		}
	}

	removeImage := false
	if r.MultipartForm != nil {
		if vals, ok := r.MultipartForm.Value["remove_image"]; ok && len(vals) > 0 {
			removeImage, _ = strconv.ParseBool(vals[0])
		}
	}

	imageFile, _, err := r.FormFile("image")
	hasImage := err == nil
	if hasImage {
		defer imageFile.Close()
	}

	var savedImagePath *string
	var success bool
	defer func() {
		if !success && savedImagePath != nil {
			_ = storage.DeleteImage(*savedImagePath)
		}
	}()

	if hasImage {
		path, err := storage.SaveImage(imageFile)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Failed to save image", map[string]string{"image": err.Error()})
			return
		}
		savedImagePath = &path
	}

	req := &models.UpdateCommentRequest{
		Content:     contentPtr,
		ImageURL:    savedImagePath,
		RemoveImage: removeImage,
	}

	response, err := h.postService.UpdateComment(r.Context(), commentIDStr, req, currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCommentNotFound):
			_ = utils.SendError(w, http.StatusNotFound, "Comment not found", nil)
		case errors.Is(err, services.ErrForbidden):
			_ = utils.SendError(w, http.StatusForbidden, "Forbidden: you do not have permission to edit this comment", nil)
		case errors.Is(err, services.ErrPostOrCommentDeleted):
			_ = utils.SendError(w, http.StatusConflict, "Post or comment is deleted", nil)
		case errors.Is(err, services.ErrPostNotFound):
			_ = utils.SendError(w, http.StatusNotFound, "Post not found", nil)
		default:
			_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		}
		return
	}

	success = true
	_ = utils.SendSuccess(w, http.StatusOK, "Comment updated successfully", response)
}

func (h *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	commentIDStr := r.PathValue("id")
	if _, err := uuid.FromString(commentIDStr); err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "shared_validation_error: malformed id", nil)
		return
	}

	response, err := h.postService.DeleteComment(r.Context(), commentIDStr, currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCommentNotFound):
			_ = utils.SendError(w, http.StatusNotFound, "Comment not found", nil)
		case errors.Is(err, services.ErrForbidden):
			_ = utils.SendError(w, http.StatusForbidden, "Forbidden: you do not have permission to delete this comment", nil)
		default:
			_ = utils.SendError(w, http.StatusInternalServerError, "Internal server error", nil)
		}
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Comment deleted successfully", response)
}
