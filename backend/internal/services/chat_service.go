package services

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/repositories"
)

type ChatService interface {
	SendMessage(senderID uuid.UUID, req dto.SendMessageRequest) (*dto.MessageResponse, error)
	GetMessages(viewerID uuid.UUID, targetType string, targetID uuid.UUID, limit, offset int) ([]*dto.MessageResponse, error)
	GetConversations(userID uuid.UUID) ([]*dto.ConversationResponse, error)
	HandleWS(w http.ResponseWriter, r *http.Request, userID uuid.UUID)
}

type chatService struct {
	messageRepo    repositories.MessageRepository
	followerRepo   repositories.FollowersRepository
	membershipRepo repositories.GroupMembershipRepository
	userRepo       repositories.UserRepository
	groupRepo      repositories.GroupRepository

	// WebSocket Hub
	clients    map[uuid.UUID]map[*wsClient]bool
	register   chan *wsClient
	unregister chan *wsClient
	mu         sync.RWMutex
}

type wsClient struct {
	userID uuid.UUID
	conn   *websocket.Conn
	send   chan []byte
	chat   *chatService
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow development origin
	},
}

func NewChatService(
	mr repositories.MessageRepository,
	fr repositories.FollowersRepository,
	gmr repositories.GroupMembershipRepository,
	ur repositories.UserRepository,
	gr repositories.GroupRepository,
	ns NotificationService,
) ChatService {
	s := &chatService{
		messageRepo:    mr,
		followerRepo:   fr,
		membershipRepo: gmr,
		userRepo:       ur,
		groupRepo:      gr,
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

func (s *chatService) SendMessage(senderID uuid.UUID, req dto.SendMessageRequest) (*dto.MessageResponse, error) {
	if req.Content == "" {
		return nil, errors.New("message content is empty")
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
			if t.User1ID == senderID {
				recipientID = t.User2ID
			} else {
				recipientID = t.User1ID
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
		ID:         msgID,
		SenderID:   senderID,
		DMThreadID: dmThreadID,
		GroupID:    groupID,
		Content:    req.Content,
		CreatedAt:  time.Now(),
	}

	if err := s.messageRepo.CreateMessage(m); err != nil {
		return nil, err
	}

	// Broadcast message to recipients
	s.broadcastMessage(m)

	return dto.MapMessageResponse(m), nil
}

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
		if t.User1ID != viewerID && t.User2ID != viewerID {
			return nil, errors.New("unauthorized: not a participant in this conversation thread")
		}
		messages, err = s.messageRepo.ListMessagesByThread(targetID, limit, offset)
	} else {
		return nil, errors.New("invalid targetType: must be 'dm' or 'group'")
	}
	if err != nil {
		return nil, err
	}
	return dto.MapMessageResponses(messages), nil
}

func (s *chatService) GetConversations(userID uuid.UUID) ([]*dto.ConversationResponse, error) {
	conversations, err := s.messageRepo.ListConversations(userID)
	if err != nil {
		return nil, err
	}
	return dto.MapConversationResponses(conversations), nil
}

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

func (s *chatService) broadcastMessage(m *models.Message) {
	wsMsg := dto.WSMessage{
		Type:    "chat",
		Payload: dto.MapMessageResponse(m),
	}

	if m.GroupID != nil {
		// Group message: broadcast to all online group members
		members, err := s.membershipRepo.ListGroupMembers(*m.GroupID)
		if err == nil {
			for _, mb := range members {
				s.PushPayload(mb.ID, wsMsg)
			}
		}
	} else if m.DMThreadID != nil {
		// DM: push to sender and recipient
		t, err := s.messageRepo.GetDMThreadByID(*m.DMThreadID)
		if err == nil {
			s.PushPayload(t.User1ID, wsMsg)
			s.PushPayload(t.User2ID, wsMsg)
		}
	}
}

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
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// We only push downstream, ignore upstream WebSocket payloads for RESTful cleanliness
	}
}

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

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Add queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
