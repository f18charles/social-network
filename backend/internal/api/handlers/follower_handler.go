package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/services"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/utils"
)

type FollowerHandler struct {
	followerService services.FollowerService
	userService     services.UserService
}

func NewFollowerHandler(fs services.FollowerService, us services.UserService) *FollowerHandler {
	return &FollowerHandler{
		followerService: fs,
		userService:     us,
	}
}

func (h *FollowerHandler) Follow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input models.FollowRequestInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil || input.FollowingID == "" {
		utils.ErrorResponse(w, "Invalid input. following_id is required.", http.StatusBadRequest)
		return
	}

	followingUUID, err := uuid.FromString(input.FollowingID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid following_id format.", http.StatusBadRequest)
		return
	}

	_, err = h.followerService.Follow(currentUser.ID, followingUUID)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SuccessResponse(w, map[string]string{"message": "Follow request processed"}, http.StatusAccepted)
}

func (h *FollowerHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var input models.FollowRequestInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil || input.FollowingID == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid input", map[string]string{"following_id": "is required"})
		return
	}

	followingUUID, err := uuid.FromString(input.FollowingID)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid input", map[string]string{"following_id": "has an invalid format"})
		return
	}

	err = h.followerService.Unfollow(currentUser.ID, followingUUID)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Unfollowed successfully", nil)
}

func (h *FollowerHandler) AcceptFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var input models.AcceptRejectFollowInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil || input.FollowerID == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid input", map[string]string{"follower_id": "is required"})
		return
	}

	followerUUID, err := uuid.FromString(input.FollowerID)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid input", map[string]string{"follower_id": "has an invalid format"})
		return
	}

	err = h.followerService.AcceptFollow(followerUUID, currentUser.ID)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Follow request accepted", nil)
}

func (h *FollowerHandler) RejectFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var input models.AcceptRejectFollowInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil || input.FollowerID == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid input", map[string]string{"follower_id": "is required"})
		return
	}

	followerUUID, err := uuid.FromString(input.FollowerID)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid input", map[string]string{"follower_id": "has an invalid format"})
		return
	}

	err = h.followerService.RejectFollow(followerUUID, currentUser.ID)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Follow request rejected", nil)
}

func (h *FollowerHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var targetUserID uuid.UUID
	targetUserIDStr := r.URL.Query().Get("user_id")
	if targetUserIDStr != "" {
		parsed, err := uuid.FromString(targetUserIDStr)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid input", map[string]string{"user_id": "has an invalid format"})
			return
		}
		targetUserID = parsed
	} else {
		targetUserID = currentUser.ID
	}

	// Verify permission if listing another user's followers
	if targetUserID != currentUser.ID {
		err := h.verifyAccess(currentUser.ID, targetUserID)
		if err != nil {
			_ = utils.SendError(w, http.StatusForbidden, err.Error(), nil)
			return
		}
	}

	followers, err := h.followerService.GetFollowers(targetUserID)
	if err != nil {
		_ = utils.SendError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Map to UserResponse to avoid exposing passwords/hashes
	var response []*models.UserResponse
	for _, f := range followers {
		response = append(response, &models.UserResponse{
			ID:          f.ID,
			Email:       f.Email,
			FirstName:   f.FirstName,
			LastName:    f.LastName,
			DateOfBirth: f.DOB.Format("2006-01-02"),
			Avatar:      f.Avatar,
			Nickname:    f.Nickname,
			AboutMe:     f.AboutMe,
			IsPublic:    f.IsPublic,
			CreatedAt:   f.CreatedAt,
		})
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Followers retrieved successfully", response)
}

func (h *FollowerHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var targetUserID uuid.UUID
	targetUserIDStr := r.URL.Query().Get("user_id")
	if targetUserIDStr != "" {
		parsed, err := uuid.FromString(targetUserIDStr)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid input", map[string]string{"user_id": "has an invalid format"})
			return
		}
		targetUserID = parsed
	} else {
		targetUserID = currentUser.ID
	}

	// Verify permission if listing another user's following list
	if targetUserID != currentUser.ID {
		err := h.verifyAccess(currentUser.ID, targetUserID)
		if err != nil {
			_ = utils.SendError(w, http.StatusForbidden, err.Error(), nil)
			return
		}
	}

	following, err := h.followerService.GetFollowing(targetUserID)
	if err != nil {
		_ = utils.SendError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var response []*models.UserResponse
	for _, f := range following {
		response = append(response, &models.UserResponse{
			ID:          f.ID,
			Email:       f.Email,
			FirstName:   f.FirstName,
			LastName:    f.LastName,
			DateOfBirth: f.DOB.Format("2006-01-02"),
			Avatar:      f.Avatar,
			Nickname:    f.Nickname,
			AboutMe:     f.AboutMe,
			IsPublic:    f.IsPublic,
			CreatedAt:   f.CreatedAt,
		})
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Following retrieved successfully", response)
}

func (h *FollowerHandler) verifyAccess(currentUserID, targetUserID uuid.UUID) error {
	targetUser, err := h.userService.GetByID(targetUserID)
	if err != nil {
		return errors.New("user not found")
	}

	if targetUser.IsPublic {
		return nil
	}

	// If the profile is private, currentUserID must follow targetUserID with status = 'accepted'
	status, err := h.followerService.GetFollowStatus(currentUserID, targetUserID)
	if err != nil {
		return err
	}

	if status != string(models.Accepted) {
		return errors.New("profile is private. You must follow this user to view their activity.")
	}

	return nil
}
