package handlers

import (
	"net/http"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
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

	var req dto.SendMessageRequest
	if err := utils.DecodeJSON(r, &req); err != nil || req.Content == "" {
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

	limit := 100
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

// GetDMCandidates returns accepted follow connections that do not have an active DM history.
func (h *ChatHandler) GetDMCandidates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if val, err := strconv.Atoi(raw); err == nil && val > 0 {
			limit = val
		}
	}
	candidates, err := h.chatService.ListDMCandidates(currentUser.ID, limit)
	if err != nil {
		_ = utils.SendError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	_ = utils.SendSuccess(w, http.StatusOK, "DM candidates returned.", candidates)
}

func (h *ChatHandler) OpenDM(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	var req struct {
		RecipientID string `json:"recipient_id"`
	}
	if err := utils.DecodeJSON(r, &req); err != nil || req.RecipientID == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid request. recipient_id is required.", nil)
		return
	}
	recipientID, err := uuid.FromString(req.RecipientID)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid recipient_id format", nil)
		return
	}
	conversation, err := h.chatService.OpenDM(currentUser.ID, recipientID)
	if err != nil {
		_ = utils.SendError(w, http.StatusForbidden, err.Error(), nil)
		return
	}
	_ = utils.SendSuccess(w, http.StatusOK, "DM chat returned.", conversation)
}

func (h *ChatHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	chatID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid chat ID format", nil)
		return
	}
	targetType := r.URL.Query().Get("chat_type")
	if targetType == "" {
		targetType = r.URL.Query().Get("type")
	}
	if targetType == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "chat_type is required", nil)
		return
	}
	limit := 100
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
	messages, err := h.chatService.GetMessages(currentUser.ID, targetType, chatID, limit, offset)
	if err != nil {
		_ = utils.SendError(w, http.StatusForbidden, err.Error(), nil)
		return
	}
	_ = utils.SendSuccess(w, http.StatusOK, "Messages returned.", messages)
}

func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	idStr := r.PathValue("id")
	msgID, err := uuid.FromString(idStr)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid message ID format", nil)
		return
	}

	if err := h.chatService.DeleteMessage(msgID, currentUser.ID); err != nil {
		_ = utils.SendError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Message deleted successfully", nil)
}

func (h *ChatHandler) ClearChatMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	chatIDStr := r.URL.Query().Get("chat_id")
	chatType := r.URL.Query().Get("chat_type")
	if chatIDStr == "" || chatType == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "chat_id and chat_type are required query parameters", nil)
		return
	}

	chatID, err := uuid.FromString(chatIDStr)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid chat_id format", nil)
		return
	}

	if err := h.chatService.DeleteAllMessagesInChat(chatID, chatType, currentUser.ID); err != nil {
		_ = utils.SendError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Chat messages cleared successfully", nil)
}
