package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/services"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/storage"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/utils"
)

type UserHandler struct {
	userService     services.UserService
	followerService services.FollowerService
	postService     services.PostService
}

func NewUserHandler(us services.UserService, fs services.FollowerService, ps services.PostService) *UserHandler {
	return &UserHandler{
		userService:     us,
		followerService: fs,
		postService:     ps,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req dto.CreateUserRequest
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Parse multipart form (10 MB limit)
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Failed to parse multipart form", nil)
			return
		}

		req.Email = r.FormValue("email")
		req.Password = r.FormValue("password")
		req.FirstName = r.FormValue("first_name")
		req.LastName = r.FormValue("last_name")
		req.DateOfBirth = r.FormValue("date_of_birth")
		req.Nickname = r.FormValue("nickname")
		req.AboutMe = r.FormValue("about_me")
		req.IsPublic = r.FormValue("is_public") == "true"

		// Handle Avatar upload
		file, _, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()

			req.Avatar, err = storage.SaveAvatar(file)
			if err != nil {
				utils.SendError(w, http.StatusInternalServerError, "Failed to save image", nil)
				return
			}
		}
	} else {
		// Handle JSON
		err := utils.DecodeJSON(r, &req)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid request body", nil)
			return
		}
	}

	userResponse, err := h.userService.Register(&req)
	if err != nil {
		if req.Avatar != "" {
			_ = storage.DeleteImage(req.Avatar)
		}

		_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Auto-login: establish session for the new user
	session, err := h.userService.Login(req.Email, req.Password)
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    session.ID.String(),
			Expires:  session.ExpiresAt,
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})
	}

	_ = utils.SendSuccess(w, http.StatusCreated, "User registered successfully", userResponse)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := utils.DecodeJSON(r, &req)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	session, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		_ = utils.SendError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	// Set HttpOnly session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.ID.String(),
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	_ = utils.SendSuccess(w, http.StatusOK, "Login successful", dto.LoginResponse{
		Token: session.ID.String(),
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		_ = h.userService.Logout(cookie.Value)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	_ = utils.SendSuccess(w, http.StatusOK, "Logout successful", nil)
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := h.userService.Authenticate(cookie.Value)
	if err != nil {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "User retrieved successfully", dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DOB.Format("2006-01-02"),
		Avatar:      user.Avatar,
		Nickname:    user.Nickname,
		AboutMe:     user.AboutMe,
		IsPublic:    user.IsPublic,
		CreatedAt:   user.CreatedAt,
	})
}

// DeleteMe deletes the authenticated account and clears the session cookie.
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	user, err := h.userService.Authenticate(cookie.Value)
	if err != nil {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	if err := h.userService.DeleteAccount(user.ID); err != nil {
		_ = utils.SendError(w, http.StatusInternalServerError, "Failed to delete account", nil)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	_ = utils.SendSuccess(w, http.StatusOK, "Account deleted successfully", nil)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	currentUser, err := h.userService.Authenticate(cookie.Value)
	if err != nil {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Get the target user ID from URL
	userID := r.PathValue("id")
	if userID == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	targetID, err := uuid.FromString(userID)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	// Get user data
	user, err := h.userService.GetByID(targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = utils.SendError(w, http.StatusNotFound, "User not found", nil)
			return
		}
		_ = utils.SendError(w, http.StatusInternalServerError, "Failed to fetch user", nil)
		return
	}

	// Build base response
	response := dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DOB.Format("2006-01-02"),
		Avatar:      user.Avatar,
		Nickname:    user.Nickname,
		AboutMe:     user.AboutMe,
		IsPublic:    user.IsPublic,
		CreatedAt:   user.CreatedAt,
		// Default values
		IsFollowing:          false,
		IsFollowedBy:         false,
		FollowRequestPending: false,
	}

	// If not viewing own profile, check follow status
	if currentUser.ID != targetID {
		status, err := h.followerService.GetFollowStatus(currentUser.ID, targetID)
		if err == nil {
			switch status {
			case "accepted":
				response.IsFollowing = true
			case "pending":
				response.FollowRequestPending = true
			}
		}
		reverseStatus, err := h.followerService.GetFollowStatus(targetID, currentUser.ID)
		if err == nil && reverseStatus == "accepted" {
			response.IsFollowedBy = true
		}
	}

	_ = utils.SendSuccess(w, http.StatusOK, "User retrieved successfully", response)
}

func (h *UserHandler) SearchPublicUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	query := r.URL.Query().Get("query")
	searchType := r.URL.Query().Get("type")
	if searchType == "" {
		searchType = "all"
	}

	var userResponses []*dto.UserResponse = []*dto.UserResponse{}
	var postResponse *dto.PostListResponse

	// 1. Search Users if type is "all" or "users"
	if searchType == "all" || searchType == "users" {
		users, err := h.userService.ListAllUsers(query, currentUser.ID)
		if err == nil {
			for _, u := range users {
				isFollowing := false
				followRequestPending := false
				if currentUser.ID != u.ID {
					status, err := h.followerService.GetFollowStatus(currentUser.ID, u.ID)
					if err == nil {
						switch status {
						case "accepted":
							isFollowing = true
						case "pending":
							followRequestPending = true
						}
					}
				}
				userResponses = append(userResponses, &dto.UserResponse{
					ID:                   u.ID,
					Email:                u.Email,
					FirstName:            u.FirstName,
					LastName:             u.LastName,
					DateOfBirth:          u.DOB.Format("2006-01-02"),
					Avatar:               u.Avatar,
					Nickname:             u.Nickname,
					AboutMe:              u.AboutMe,
					IsPublic:             u.IsPublic,
					CreatedAt:            u.CreatedAt,
					IsFollowing:          isFollowing,
					FollowRequestPending: followRequestPending,
				})
			}
		}
	}

	// 2. Search Posts if type is "all" or "posts"
	if searchType == "all" || searchType == "posts" {
		var err error
		postResponse, err = h.postService.SearchPosts(query, currentUser.ID, 50, 0)
		if err != nil {
			postResponse = &dto.PostListResponse{
				Status:  "success",
				Message: "No posts retrieved",
				Data:    []dto.PostResponse{},
			}
		}
	}

	// 3. Return results based on searchType
	if searchType == "users" {
		_ = utils.SendSuccess(w, http.StatusOK, "Users search results", userResponses)
	} else if searchType == "posts" {
		_ = utils.SendSuccess(w, http.StatusOK, "Posts search results", postResponse.Data)
	} else {
		type combinedResponse struct {
			Users []*dto.UserResponse `json:"users"`
			Posts []dto.PostResponse  `json:"posts"`
		}
		postsList := []dto.PostResponse{}
		if postResponse != nil {
			postsList = postResponse.Data
		}
		_ = utils.SendSuccess(w, http.StatusOK, "Combined search results", combinedResponse{
			Users: userResponses,
			Posts: postsList,
		})
	}
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := h.userService.Authenticate(cookie.Value)
	if err != nil {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req dto.UpdateUserRequest
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Failed to parse multipart form", nil)
			return
		}

		req.Email = r.FormValue("email")
		req.CurrentPassword = r.FormValue("current_password")
		req.NewPassword = r.FormValue("new_password")
		req.FirstName = r.FormValue("first_name")
		req.LastName = r.FormValue("last_name")
		req.DateOfBirth = r.FormValue("date_of_birth")
		req.Nickname = r.FormValue("nickname")
		req.AboutMe = r.FormValue("about_me")
		req.IsPublic = r.FormValue("is_public") == "true"

		file, _, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			req.Avatar, err = storage.SaveAvatar(file)
			if err != nil {
				utils.SendError(w, http.StatusInternalServerError, "Failed to save image", nil)
				return
			}
		}
	} else {
		err := utils.DecodeJSON(r, &req)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid request body", nil)
			return
		}
	}

	updatedUser, err := h.userService.Update(user.ID, &req)
	if err != nil {
		if req.Avatar != "" {
			_ = storage.DeleteImage(req.Avatar)
		}

		_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Profile updated successfully", updatedUser)
}
