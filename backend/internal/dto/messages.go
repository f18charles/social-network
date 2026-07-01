package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

// SendMessageRequest is the API payload for sending a direct or group message.
type SendMessageRequest struct {
	Content     string  `json:"content"`
	DMThreadID  *string `json:"dm_thread_id,omitempty"`
	RecipientID *string `json:"recipient_id,omitempty"`
	GroupID     *string `json:"group_id,omitempty"`
}

// MessageResponse is the API representation of a chat message.
type MessageResponse struct {
	ID         uuid.UUID  `json:"id"`
	SenderID   uuid.UUID  `json:"sender_id"`
	DMThreadID *uuid.UUID `json:"dm_thread_id,omitempty"`
	GroupID    *uuid.UUID `json:"group_id,omitempty"`
	Content    string     `json:"content"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ConversationResponse is the API representation of a chat conversation.
type ConversationResponse struct {
	ThreadID      *uuid.UUID `json:"thread_id,omitempty"`
	GroupID       *uuid.UUID `json:"group_id,omitempty"`
	Type          string     `json:"type"`
	TargetName    string     `json:"target_name"`
	TargetAvatar  string     `json:"target_avatar,omitempty"`
	LastMessage   string     `json:"last_message"`
	LastMessageAt time.Time  `json:"last_message_at"`
}

// WSMessage represents a message wrapper sent over WebSocket.
type WSMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// MapMessageResponse maps a message domain model to an API DTO.
func MapMessageResponse(message *models.Message) *MessageResponse {
	if message == nil {
		return nil
	}
	return &MessageResponse{
		ID:         message.ID,
		SenderID:   message.SenderID,
		DMThreadID: message.DMThreadID,
		GroupID:    message.GroupID,
		Content:    message.Content,
		CreatedAt:  message.CreatedAt,
	}
}

// MapMessageResponses maps message domain models to API DTOs.
func MapMessageResponses(messages []*models.Message) []*MessageResponse {
	response := make([]*MessageResponse, 0, len(messages))
	for _, message := range messages {
		response = append(response, MapMessageResponse(message))
	}
	return response
}

// MapConversationResponse maps a conversation read model to an API DTO.
func MapConversationResponse(conversation *models.Conversation) *ConversationResponse {
	if conversation == nil {
		return nil
	}
	return &ConversationResponse{
		ThreadID:      conversation.ThreadID,
		GroupID:       conversation.GroupID,
		Type:          conversation.Type,
		TargetName:    conversation.TargetName,
		TargetAvatar:  conversation.TargetAvatar,
		LastMessage:   conversation.LastMessage,
		LastMessageAt: conversation.LastMessageAt,
	}
}

// MapConversationResponses maps conversation read models to API DTOs.
func MapConversationResponses(conversations []*models.Conversation) []*ConversationResponse {
	response := make([]*ConversationResponse, 0, len(conversations))
	for _, conversation := range conversations {
		response = append(response, MapConversationResponse(conversation))
	}
	return response
}
