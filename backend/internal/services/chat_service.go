package services

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/config"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/repositories"
)

// ChatService handles DM/group messaging and the WebSocket hub that delivers it live.
type ChatService interface {
	SendMessage(senderID uuid.UUID, req dto.SendMessageRequest) (*dto.MessageResponse, error)
	GetMessages(viewerID uuid.UUID, targetType string, targetID uuid.UUID, limit, offset int) ([]*dto.MessageResponse, error)
	GetConversations(userID uuid.UUID) ([]*dto.ConversationResponse, error)
	ListDMCandidates(userID uuid.UUID, limit int) ([]*dto.UserResponse, error)
	OpenDM(senderID, recipientID uuid.UUID) (*dto.ConversationResponse, error)
	HandleWS(w http.ResponseWriter, r *http.Request, userID uuid.UUID)
	SetMessageReaction(messageID, userID uuid.UUID, emoji string) ([]*models.MessageReactionSummary, error)
	DeleteMessageReaction(messageID, userID uuid.UUID, emoji string) ([]*models.MessageReactionSummary, error)
	GetMessageReactionSummary(messageID, viewerID uuid.UUID) ([]*models.MessageReactionSummary, error)
	DeleteMessage(messageID, senderID uuid.UUID) error
	DeleteAllMessagesInChat(chatID uuid.UUID, chatType string, senderID uuid.UUID) error
}

// chatService is the default ChatService implementation and WebSocket hub.
type chatService struct {
	messageRepo    repositories.MessageRepository
	followerRepo   repositories.FollowersRepository
	membershipRepo repositories.GroupMembershipRepository
	userRepo       repositories.UserRepository
	groupRepo      repositories.GroupRepository
	reactionRepo   repositories.MessageReactionRepository

	// WebSocket Hub
	clients    map[uuid.UUID]map[*wsClient]bool
	register   chan *wsClient
	unregister chan *wsClient
	mu         sync.RWMutex
}

// wsClient is one connected WebSocket client (one browser tab/session).
type wsClient struct {
	userID uuid.UUID
	conn   *websocket.Conn
	send   chan []byte
	chat   *chatService
}

// upgrader configures the HTTP->WebSocket upgrade.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

// checkOrigin allows same-origin (or no-Origin) WebSocket connections only.
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	return origin == config.App.AllowedOrigin
}

// NewChatService builds a ChatService and starts its WebSocket hub goroutine.
func NewChatService(
	mr repositories.MessageRepository,
	fr repositories.FollowersRepository,
	gmr repositories.GroupMembershipRepository,
	ur repositories.UserRepository,
	gr repositories.GroupRepository,
	ns NotificationService,
	rr repositories.MessageReactionRepository,
) ChatService {
	s := &chatService{
		messageRepo:    mr,
		followerRepo:   fr,
		membershipRepo: gmr,
		userRepo:       ur,
		groupRepo:      gr,
		reactionRepo:   rr,
		clients:        make(map[uuid.UUID]map[*wsClient]bool),
		register:       make(chan *wsClient),
		unregister:     make(chan *wsClient),
	}

	// Register notification push handler
	ns.RegisterPushHandler(s.PushPayload)

	// Run Hub loop
	go s.runHub()

	return s
}

// SendMessage validates and stores a DM/group message, then broadcasts it live.
func (s *chatService) SendMessage(senderID uuid.UUID, req dto.SendMessageRequest) (*dto.MessageResponse, error) {
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		return nil, errors.New("message content is empty")
	}
	if len([]rune(req.Content)) > 2000 {
		return nil, errors.New("message content exceeds 2000 characters")
	}
	if req.ChatID != "" {
		if req.ChatType == "group" {
			req.GroupID = &req.ChatID
		} else if req.ChatType == "dm" {
			req.DMThreadID = &req.ChatID
		} else {
			return nil, errors.New("invalid chat_type")
		}
	}

	var dmThreadID *uuid.UUID
	var groupID *uuid.UUID

	if req.GroupID != nil && *req.GroupID != "" {
		gID, err := uuid.FromString(*req.GroupID)
		if err != nil {
			return nil, errors.New("invalid group_id")
		}
		// Check group membership
		isMember, err := s.membershipRepo.IsAcceptedGroupMember(gID, senderID)
		if err != nil || !isMember {
			return nil, errors.New("unauthorized: must be group member to post messages")
		}
		groupID = &gID
	} else {
		// DM logic
		var recipientID uuid.UUID
		if req.DMThreadID != nil && *req.DMThreadID != "" {
			tID, err := uuid.FromString(*req.DMThreadID)
			if err != nil {
				return nil, errors.New("invalid dm_thread_id")
			}
			t, err := s.messageRepo.GetDMThreadByID(tID)
			if err != nil {
				return nil, errors.New("thread not found")
			}
			dmThreadID = &tID
			if t.User1ID != nil && *t.User1ID == senderID {
				if t.User2ID == nil {
					return nil, errors.New("cannot message a deleted account")
				}
				recipientID = *t.User2ID
			} else if t.User2ID != nil && *t.User2ID == senderID {
				if t.User1ID == nil {
					return nil, errors.New("cannot message a deleted account")
				}
				recipientID = *t.User1ID
			} else {
				return nil, errors.New("unauthorized: not a participant in this conversation thread")
			}
		} else if req.RecipientID != nil && *req.RecipientID != "" {
			rID, err := uuid.FromString(*req.RecipientID)
			if err != nil {
				return nil, errors.New("invalid recipient_id")
			}
			recipientID = rID

			// Get or create thread
			thread, err := s.messageRepo.GetOrCreateDMThread(senderID, recipientID)
			if err != nil {
				return nil, err
			}
			dmThreadID = &thread.ID
		} else {
			return nil, errors.New("must supply recipient_id, dm_thread_id, or group_id")
		}

		// Verify follower relationship (at least one must follow the other)
		status1, err1 := s.followerRepo.GetStatus(senderID, recipientID)
		status2, err2 := s.followerRepo.GetStatus(recipientID, senderID)

		if (err1 != nil || status1 != "accepted") && (err2 != nil || status2 != "accepted") {
			return nil, errors.New("unauthorized: must follow or be followed by the user to message them")
		}
	}

	msgID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	m := &models.Message{
		ID:          msgID,
		SenderID:    &senderID,
		DMThreadID:  dmThreadID,
		GroupID:     groupID,
		Content:     req.Content,
		CreatedAt:   time.Now(),
		MessageType: "user",
		Reactions:   []*models.MessageReactionSummary{},
	}

	if err := s.messageRepo.CreateMessage(m); err != nil {
		return nil, err
	}

	// Broadcast message to recipients (echo client_message_id back to the sender only)
	s.broadcastMessage(m, req.ClientMessageID)

	return dto.MapMessageResponse(m), nil
}

// GetMessages returns paginated messages for a DM thread or group, with reactions attached.
func (s *chatService) GetMessages(viewerID uuid.UUID, targetType string, targetID uuid.UUID, limit, offset int) ([]*dto.MessageResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var messages []*models.Message
	var err error
	if targetType == "group" {
		// Check membership
		isMember, err := s.membershipRepo.IsAcceptedGroupMember(targetID, viewerID)
		if err != nil || !isMember {
			return nil, errors.New("unauthorized: must be group member to view messages")
		}
		messages, err = s.messageRepo.ListMessagesByGroup(targetID, limit, offset)
	} else if targetType == "dm" {
		t, err := s.messageRepo.GetDMThreadByID(targetID)
		if err != nil {
			return nil, err
		}
		// Verify viewer is in thread
		isParticipant := (t.User1ID != nil && *t.User1ID == viewerID) || (t.User2ID != nil && *t.User2ID == viewerID)
		if !isParticipant {
			return nil, errors.New("unauthorized: not a participant in this conversation thread")
		}
		messages, err = s.messageRepo.ListMessagesByThread(targetID, limit, offset)
	} else {
		return nil, errors.New("invalid targetType: must be 'dm' or 'group'")
	}
	if err != nil {
		return nil, err
	}
	for _, m := range messages {
		reactions, err := s.reactionRepo.GetMessageReactionSummary(m.ID, viewerID)
		if err == nil {
			m.Reactions = reactions
		} else {
			m.Reactions = []*models.MessageReactionSummary{}
		}
	}
	return dto.MapMessageResponses(messages), nil
}

// GetConversations returns a user's DM and group conversation list.
func (s *chatService) GetConversations(userID uuid.UUID) ([]*dto.ConversationResponse, error) {
	conversations, err := s.messageRepo.ListConversations(userID)
	if err != nil {
		return nil, err
	}
	return dto.MapConversationResponses(conversations), nil
}

// HandleWS upgrades an HTTP request to a WebSocket connection and registers the client.
func (s *chatService) HandleWS(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade failed for user %s: %v", userID, err)
		return
	}

	client := &wsClient{
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, 256),
		chat:   s,
	}

	s.register <- client

	// Start read/write loops
	go client.writePump()
	go client.readPump()
}

// runHub is the single goroutine that owns s.clients, processing register/unregister events.
func (s *chatService) runHub() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			if _, exists := s.clients[client.userID]; !exists {
				s.clients[client.userID] = make(map[*wsClient]bool)
			}
			s.clients[client.userID][client] = true
			s.mu.Unlock()

		case client := <-s.unregister:
			s.mu.Lock()
			if clientsMap, exists := s.clients[client.userID]; exists {
				if _, existsClient := clientsMap[client]; existsClient {
					delete(clientsMap, client)
					close(client.send)
				}
				if len(clientsMap) == 0 {
					delete(s.clients, client.userID)
				}
			}
			s.mu.Unlock()
		}
	}
}

// PushPayload sends a JSON payload to every open connection for a user.
func (s *chatService) PushPayload(userID uuid.UUID, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if clientsMap, exists := s.clients[userID]; exists {
		for client := range clientsMap {
			select {
			case client.send <- data:
			default:
				go func(c *wsClient) {
					s.unregister <- c
					c.conn.Close()
				}(client)
			}
		}
	}
}

// broadcastMessage pushes a new/updated message to its group members or DM
// participants. clientMessageID (if any) is echoed back only to the sender's
// own copy, so their UI can reconcile it with the optimistic message it
// already rendered instead of showing a duplicate.
func (s *chatService) broadcastMessage(m *models.Message, clientMessageID string) {
	message := dto.MapMessageResponse(m)
	wsMsg := dto.WSMessage{
		Type:    "message.created",
		Payload: message,
		Data:    message,
	}
	senderWsMsg := wsMsg
	senderWsMsg.ClientMessageID = clientMessageID

	pushTo := func(userID uuid.UUID) {
		if m.SenderID != nil && userID == *m.SenderID {
			s.PushPayload(userID, senderWsMsg)
		} else {
			s.PushPayload(userID, wsMsg)
		}
	}

	if m.GroupID != nil {
		// Group message: broadcast to all online group members
		members, err := s.membershipRepo.ListGroupMembers(*m.GroupID)
		if err == nil {
			for _, mb := range members {
				pushTo(mb.ID)
			}
		}
	} else if m.DMThreadID != nil {
		// DM: push to sender and recipient
		t, err := s.messageRepo.GetDMThreadByID(*m.DMThreadID)
		if err == nil {
			if t.User1ID != nil {
				pushTo(*t.User1ID)
			}
			if t.User2ID != nil {
				pushTo(*t.User2ID)
			}
		}
	}
}

// readPump reads incoming frames from the client until the connection closes.
func (c *wsClient) readPump() {
	defer func() {
		c.chat.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(4096)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		c.handleIncoming(raw)
	}
}

// writePump writes queued outgoing frames and sends periodic pings to keep the connection alive.
func (c *wsClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

			// Drain any additional queued messages, each as its own frame —
			// the frontend does JSON.parse(event.data) per message and has
			// no framing/delimiter handling of its own.
			n := len(c.send)
			for i := 0; i < n; i++ {
				queued := <-c.send
				_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.conn.WriteMessage(websocket.TextMessage, queued); err != nil {
					return
				}
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleIncoming parses one client WS frame and dispatches it (ack or new message).
func (c *wsClient) handleIncoming(raw []byte) {
	var envelope struct {
		Type            string `json:"type"`
		ChatID          string `json:"chat_id"`
		ChatType        string `json:"chat_type"`
		Content         string `json:"content"`
		ClientMessageID string `json:"client_message_id"`
		MessageID       string `json:"message_id"`
		SenderID        string `json:"sender_id"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		c.sendError("INVALID_JSON", "Message payload is invalid JSON.", "")
		return
	}
	if envelope.Type == "message.received" {
		senderUUID, err := uuid.FromString(envelope.SenderID)
		if err != nil {
			return
		}
		wsMsg := dto.WSMessage{
			Type:            "message.received",
			ClientMessageID: envelope.ClientMessageID,
			Payload: map[string]any{
				"message_id":  envelope.MessageID,
				"receiver_id": c.userID.String(),
				"chat_id":     envelope.ChatID,
				"chat_type":   envelope.ChatType,
			},
			Data: map[string]any{
				"message_id":  envelope.MessageID,
				"receiver_id": c.userID.String(),
				"chat_id":     envelope.ChatID,
				"chat_type":   envelope.ChatType,
			},
		}
		c.chat.PushPayload(senderUUID, wsMsg)
		return
	}
	if envelope.Type != "message.send" {
		c.sendError("UNKNOWN_EVENT", "Unsupported WebSocket event type.", envelope.ClientMessageID)
		return
	}
	_, err := c.chat.SendMessage(c.userID, dto.SendMessageRequest{
		ChatID:          envelope.ChatID,
		ChatType:        envelope.ChatType,
		Content:         envelope.Content,
		ClientMessageID: envelope.ClientMessageID,
	})
	if err != nil {
		c.sendError("CHAT_FORBIDDEN", err.Error(), envelope.ClientMessageID)
		return
	}
	// No separate ack here: SendMessage's broadcastMessage already pushes a
	// "message.created" event back to this same client (as a chat
	// participant), with ClientMessageID attached, so the sender's UI can
	// reconcile its optimistic message. Sending a second copy here caused
	// duplicate messages to appear for the sender.
}

// sendError sends a WS error frame back to the client.
func (c *wsClient) sendError(code, message, clientMessageID string) {
	payload := dto.WSMessage{
		Type: "error",
		Error: map[string]string{
			"code":    code,
			"message": message,
		},
		ClientMessageID: clientMessageID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
		c.chat.unregister <- c
		_ = c.conn.Close()
	}
}

// ListDMCandidates returns users the caller could start a new DM with.
func (s *chatService) ListDMCandidates(userID uuid.UUID, limit int) ([]*dto.UserResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	users, err := s.messageRepo.ListDMCandidates(userID, limit)
	if err != nil {
		return nil, err
	}
	response := make([]*dto.UserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, dto.MustMapUserResponse(user))
	}
	return response, nil
}

// OpenDM gets or creates a DM thread between two mutually-eligible users.
func (s *chatService) OpenDM(senderID, recipientID uuid.UUID) (*dto.ConversationResponse, error) {
	if senderID == recipientID {
		return nil, errors.New("cannot open a DM with yourself")
	}
	status1, err1 := s.followerRepo.GetStatus(senderID, recipientID)
	status2, err2 := s.followerRepo.GetStatus(recipientID, senderID)
	if (err1 != nil || status1 != "accepted") && (err2 != nil || status2 != "accepted") {
		return nil, errors.New("unauthorized: must follow or be followed by the user to message them")
	}
	thread, err := s.messageRepo.GetOrCreateDMThread(senderID, recipientID)
	if err != nil {
		return nil, err
	}
	var otherID uuid.UUID
	if thread.User1ID != nil && *thread.User1ID == senderID {
		if thread.User2ID != nil {
			otherID = *thread.User2ID
		}
	} else if thread.User2ID != nil && *thread.User2ID == senderID {
		if thread.User1ID != nil {
			otherID = *thread.User1ID
		}
	}

	var targetName string = "Deleted User"
	var targetAvatar string = ""
	if otherID != uuid.Nil {
		other, err := s.userRepo.GetUserByID(otherID)
		if err == nil {
			targetName = other.FirstName + " " + other.LastName
			targetAvatar = other.Avatar
		}
	}

	var targetID *uuid.UUID
	if otherID != uuid.Nil {
		targetID = &otherID
	}

	return &dto.ConversationResponse{
		ThreadID:      &thread.ID,
		TargetID:      targetID,
		Type:          "dm",
		TargetName:    targetName,
		TargetAvatar:  targetAvatar,
		LastMessage:   "",
		LastMessageAt: thread.LastMessageAt,
	}, nil
}

// SetMessageReaction adds a reaction and broadcasts the updated summary.
func (s *chatService) SetMessageReaction(messageID, userID uuid.UUID, emoji string) ([]*models.MessageReactionSummary, error) {
	msg, err := s.messageRepo.GetMessageByID(messageID)
	if err != nil {
		return nil, errors.New("message not found")
	}

	err = s.reactionRepo.SetMessageReaction(messageID, userID, emoji)
	if err != nil {
		return nil, err
	}

	summary, err := s.reactionRepo.GetMessageReactionSummary(messageID, userID)
	if err != nil {
		return nil, err
	}

	s.broadcastMessageReactionUpdate(msg, summary)
	return summary, nil
}

// DeleteMessageReaction removes a reaction and broadcasts the updated summary.
func (s *chatService) DeleteMessageReaction(messageID, userID uuid.UUID, emoji string) ([]*models.MessageReactionSummary, error) {
	msg, err := s.messageRepo.GetMessageByID(messageID)
	if err != nil {
		return nil, errors.New("message not found")
	}

	err = s.reactionRepo.DeleteMessageReaction(messageID, userID, emoji)
	if err != nil {
		return nil, err
	}

	summary, err := s.reactionRepo.GetMessageReactionSummary(messageID, userID)
	if err != nil {
		return nil, err
	}

	s.broadcastMessageReactionUpdate(msg, summary)
	return summary, nil
}

// GetMessageReactionSummary returns per-emoji reaction counts for a message.
func (s *chatService) GetMessageReactionSummary(messageID, viewerID uuid.UUID) ([]*models.MessageReactionSummary, error) {
	return s.reactionRepo.GetMessageReactionSummary(messageID, viewerID)
}

// broadcastMessageReactionUpdate pushes a per-viewer reaction summary to each recipient.
func (s *chatService) broadcastMessageReactionUpdate(m *models.Message, summary []*models.MessageReactionSummary) {
	if m.GroupID != nil {
		members, err := s.membershipRepo.ListGroupMembers(*m.GroupID)
		if err == nil {
			for _, mb := range members {
				userSummary, err := s.reactionRepo.GetMessageReactionSummary(m.ID, mb.ID)
				if err == nil {
					wsMsg := dto.WSMessage{
						Type: "message.reaction.updated",
						Payload: map[string]any{
							"message_id": m.ID,
							"reactions":  userSummary,
						},
						Data: map[string]any{
							"message_id": m.ID,
							"reactions":  userSummary,
						},
					}
					s.PushPayload(mb.ID, wsMsg)
				}
			}
		}
	} else if m.DMThreadID != nil {
		t, err := s.messageRepo.GetDMThreadByID(*m.DMThreadID)
		if err == nil {
			var userIDs []uuid.UUID
			if t.User1ID != nil {
				userIDs = append(userIDs, *t.User1ID)
			}
			if t.User2ID != nil {
				userIDs = append(userIDs, *t.User2ID)
			}
			for _, uID := range userIDs {
				userSummary, err := s.reactionRepo.GetMessageReactionSummary(m.ID, uID)
				if err == nil {
					wsMsg := dto.WSMessage{
						Type: "message.reaction.updated",
						Payload: map[string]any{
							"message_id": m.ID,
							"reactions":  userSummary,
						},
						Data: map[string]any{
							"message_id": m.ID,
							"reactions":  userSummary,
						},
					}
					s.PushPayload(uID, wsMsg)
				}
			}
		}
	}
}

// DeleteMessage deletes a message (author only) and broadcasts the removal.
func (s *chatService) DeleteMessage(messageID, senderID uuid.UUID) error {
	msg, err := s.messageRepo.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	if msg.SenderID == nil || *msg.SenderID != senderID {
		return errors.New("unauthorized: cannot delete someone else's message")
	}

	if err := s.messageRepo.DeleteMessage(messageID, senderID); err != nil {
		return err
	}

	now := time.Now()
	msg.DeletedAt = &now
	s.broadcastMessage(msg, "")

	return nil
}

// DeleteAllMessagesInChat deletes all of senderID's messages in a chat and notifies participants.
func (s *chatService) DeleteAllMessagesInChat(chatID uuid.UUID, chatType string, senderID uuid.UUID) error {
	if err := s.messageRepo.DeleteAllMessagesInChat(chatID, chatType, senderID); err != nil {
		return err
	}

	type ClearedPayload struct {
		ChatID   uuid.UUID `json:"chat_id"`
		ChatType string    `json:"chat_type"`
		SenderID uuid.UUID `json:"sender_id"`
	}
	wsMsg := dto.WSMessage{
		Type: "messages.cleared",
		Payload: ClearedPayload{
			ChatID:   chatID,
			ChatType: chatType,
			SenderID: senderID,
		},
		Data: ClearedPayload{
			ChatID:   chatID,
			ChatType: chatType,
			SenderID: senderID,
		},
	}

	if chatType == "group" {
		members, err := s.membershipRepo.ListGroupMembers(chatID)
		if err == nil {
			for _, mb := range members {
				s.PushPayload(mb.ID, wsMsg)
			}
		}
	} else {
		t, err := s.messageRepo.GetDMThreadByID(chatID)
		if err == nil {
			if t.User1ID != nil {
				s.PushPayload(*t.User1ID, wsMsg)
			}
			if t.User2ID != nil {
				s.PushPayload(*t.User2ID, wsMsg)
			}
		}
	}

	return nil
}