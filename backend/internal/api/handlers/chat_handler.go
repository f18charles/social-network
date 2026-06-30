package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/services"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/utils"
)

type ChatHandler struct {
	chatService services.ChatService
}

func NewChatHandler(cs services.ChatService) *ChatHandler {
	return &ChatHandler{chatService: cs}
}

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid request. content is required.", nil)
		return
	}

	msg, err := h.chatService.SendMessage(currentUser.ID, req)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusCreated, "Message sent successfully", msg)
}

func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	targetType := r.URL.Query().Get("type") // 'dm' or 'group'
	targetIDStr := r.URL.Query().Get("target_id")

	if targetType == "" || targetIDStr == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "type and target_id are required query parameters", nil)
		return
	}

	targetID, err := uuid.FromString(targetIDStr)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid target_id format", nil)
		return
	}

	limit := 20
	offset := 0

	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if val, err := strconv.Atoi(lStr); err == nil && val > 0 {
			limit = val
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if val, err := strconv.Atoi(oStr); err == nil && val >= 0 {
			offset = val
		}
	}

	messages, err := h.chatService.GetMessages(currentUser.ID, targetType, targetID, limit, offset)
	if err != nil {
		_ = utils.SendError(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Messages returned.", messages)
}

func (h *ChatHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	conversations, err := h.chatService.GetConversations(currentUser.ID)
	if err != nil {
		_ = utils.SendError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Conversations returned.", conversations)
}

func (h *ChatHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.chatService.HandleWS(w, r, currentUser.ID)
}
