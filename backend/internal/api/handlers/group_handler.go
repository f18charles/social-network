package handlers

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/services"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/utils"
)

type GroupHandler struct {
	groupService services.GroupService
}

func NewGroupHandler(gs services.GroupService) *GroupHandler {
	return &GroupHandler{groupService: gs}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req dto.CreateGroupRequest
	if err := utils.DecodeJSON(r, &req); err != nil || req.Title == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid request payload. title is required.", nil)
		return
	}

	g, err := h.groupService.CreateGroup(currentUser.ID, req.Title, req.Description)
	if err != nil {
		_ = utils.SendError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusCreated, "Group created successfully", g)
}

func (h *GroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	groups, err := h.groupService.ListGroups(currentUser.ID)
	if err != nil {
		_ = utils.SendError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Groups returned.", groups)
}

func (h *GroupHandler) RequestJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid group ID format", nil)
		return
	}

	err = h.groupService.RequestJoin(groupID, currentUser.ID)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Join request submitted", nil)
}

func (h *GroupHandler) InviteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid group ID format", nil)
		return
	}

	var req dto.InviteUserRequest
	if err := utils.DecodeJSON(r, &req); err != nil || req.UserID == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid request. user_id is required.", nil)
		return
	}

	inviteeID, err := uuid.FromString(req.UserID)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid user_id format", nil)
		return
	}

	err = h.groupService.InviteUser(groupID, currentUser.ID, inviteeID)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "User invited successfully", nil)
}

func (h *GroupHandler) RespondMembership(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid group ID format", nil)
		return
	}

	var req struct {
		UserID string `json:"user_id"`
		Action string `json:"action"` // 'accept', 'reject' / 'decline'
	}
	if err := utils.DecodeJSON(r, &req); err != nil || req.Action == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid request. action is required.", nil)
		return
	}

	targetUserID := currentUser.ID
	if req.UserID != "" {
		parsed, err := uuid.FromString(req.UserID)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid user_id format", nil)
			return
		}
		targetUserID = parsed
	}

	err = h.groupService.RespondToMembership(groupID, targetUserID, currentUser.ID, req.Action)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Response processed successfully", nil)
}

func (h *GroupHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid group ID format", nil)
		return
	}

	members, err := h.groupService.ListMembers(groupID, currentUser.ID)
	if err != nil {
		_ = utils.SendError(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Group members returned.", members)
}

func (h *GroupHandler) ListPendingRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid group ID format", nil)
		return
	}

	requests, err := h.groupService.ListPendingRequests(groupID, currentUser.ID)
	if err != nil {
		_ = utils.SendError(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Pending group requests returned.", requests)
}
