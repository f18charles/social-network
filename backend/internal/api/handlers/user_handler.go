package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/services"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/storage"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/utils"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(us services.UserService) *UserHandler {
	return &UserHandler{userService: us}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req models.CreateUserRequest
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
		err := json.NewDecoder(r.Body).Decode(&req)
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

	err := json.NewDecoder(r.Body).Decode(&req)
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

	_ = utils.SendSuccess(w, http.StatusOK, "Login successful", map[string]string{
		"token": session.ID.String(),
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

	_ = utils.SendSuccess(w, http.StatusOK, "User retrieved successfully", models.UserResponse{
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

	users, err := h.userService.ListPublicUsers(query, currentUser.ID)
	if err != nil {
		_ = utils.SendError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var response []*models.UserResponse
	for _, u := range users {
		response = append(response, &models.UserResponse{
			ID:          u.ID,
			Email:       u.Email,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			DateOfBirth: u.DOB.Format("2006-01-02"),
			Avatar:      u.Avatar,
			Nickname:    u.Nickname,
			AboutMe:     u.AboutMe,
			IsPublic:    u.IsPublic,
			CreatedAt:   u.CreatedAt,
		})
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Users retrieved successfully", response)
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

	var req models.UpdateUserRequest
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
		err := json.NewDecoder(r.Body).Decode(&req)
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
