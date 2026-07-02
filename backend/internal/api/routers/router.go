package routers

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/handlers"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/repositories"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/services"
)

// RegisterRoutes configures the application's HTTP ServeMux.
// It maps URL paths to their corresponding handler functions, sets up static file serving for uploads,
// and wraps the router with necessary middleware like CORS and authentication.
// Note that all registered application endpoints fall under the "/api/" path.
func Router(database *sql.DB) http.Handler {
	// initailize repos
	userRepo := repositories.NewUserRepository(database)
	sessionRepo := repositories.NewSessionRepository(database)
	followerRepo := repositories.NewFollowerRepository(database)
	postRepo := repositories.NewPostRepository(database)
	groupMembershipRepo := repositories.NewGroupMembershipRepository(database)
	groupRepo := repositories.NewGroupRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	messageRepo := repositories.NewMessageRepository(database)
	notificationRepo := repositories.NewNotificationRepository(database)
	commentRepo := repositories.NewCommentRepository(database)
	postVoteRepo := repositories.NewPostVoteRepository(database)
	commentVoteRepo := repositories.NewCommentVoteRepository(database)

	//initialize services
	userService := services.NewUserService(userRepo, sessionRepo)
	notificationService := services.NewNotificationService(notificationRepo, userRepo, groupRepo, eventRepo)
	followerService := services.NewFollowerService(followerRepo, userRepo, notificationService)
	groupService := services.NewGroupService(groupRepo, groupMembershipRepo, userRepo, notificationService)
	eventService := services.NewEventService(eventRepo, groupMembershipRepo, notificationService)
	chatService := services.NewChatService(messageRepo, followerRepo, groupMembershipRepo, userRepo, groupRepo, notificationService)
	postService := services.NewPostServiceWithVotes(postRepo, userRepo, followerRepo, groupMembershipRepo, commentRepo, postVoteRepo, commentVoteRepo)

	// initialize handlers
	userHandler := handlers.NewUserHandler(userService, followerService)
	followerHandler := handlers.NewFollowerHandler(followerService, userService)
	postHandler := handlers.NewPostHandler(postService)
	groupHandler := handlers.NewGroupHandler(groupService)
	eventHandler := handlers.NewEventHandler(eventService)
	chatHandler := handlers.NewChatHandler(chatService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	authMiddleware := middleware.Auth(userService)

	// Initialize ServeMux
	mux := http.NewServeMux()

	// Serve static uploads (for avatar files)
	uploadsDir := "./uploads"
	if err := os.MkdirAll(filepath.Join(uploadsDir, "images"), 0o755); err != nil {
		log.Fatalf("Failed to create image uploads directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(uploadsDir, "avatars"), 0o755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", handlers.SafeUploadsHandler(uploadsDir)))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// for media
	fs := http.FileServer(http.Dir("../../uploads/"))
	mux.Handle("/api/uploads/", http.StripPrefix("/api/", fs))

	// Public routes
	mux.HandleFunc("/api/users/register", userHandler.Register)
	mux.HandleFunc("/api/users/login", userHandler.Login)

	// Authenticated routes
	// user & follower routes
	mux.Handle("/api/users/me", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			userHandler.DeleteMe(w, r)
			return
		}
		userHandler.Me(w, r)
	}))
	mux.Handle("/api/users/search", authMiddleware(http.HandlerFunc(userHandler.SearchPublicUsers)))
	mux.Handle("/api/users/update", http.HandlerFunc(userHandler.Update))
	mux.Handle("/api/users/{id}", http.HandlerFunc(userHandler.GetUser))
	mux.HandleFunc("/api/users/logout", userHandler.Logout)

	mux.Handle("/api/followers/follow", authMiddleware(http.HandlerFunc(followerHandler.Follow)))
	mux.Handle("/api/followers/unfollow", authMiddleware(http.HandlerFunc(followerHandler.Unfollow)))
	mux.Handle("/api/followers/accept", authMiddleware(http.HandlerFunc(followerHandler.AcceptFollow)))
	mux.Handle("/api/followers/reject", authMiddleware(http.HandlerFunc(followerHandler.RejectFollow)))
	mux.Handle("/api/followers/followers", authMiddleware(http.HandlerFunc(followerHandler.GetFollowers)))
	mux.Handle("/api/followers/following", authMiddleware(http.HandlerFunc(followerHandler.GetFollowing)))
	mux.Handle("/api/followers/pending", authMiddleware(http.HandlerFunc(followerHandler.GetPendingFollowers)))
	mux.Handle("/api/posts", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			postHandler.CreatePost(w, r)
		} else {
			postHandler.Feed(w, r)
		}
	})))
	mux.Handle("/api/users/{id}/posts", authMiddleware(http.HandlerFunc(postHandler.ProfilePosts)))
	mux.Handle("/api/users/{id}/comments", authMiddleware(http.HandlerFunc(postHandler.ProfileComments)))

	mux.Handle("/api/posts/{id}", authMiddleware(http.HandlerFunc(postHandler.GetSinglePost)))
	mux.Handle("/api/posts/{id}/comments", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			postHandler.CreateComment(w, r)
		} else {
			postHandler.GetComments(w, r)
		}
	})))
	mux.Handle("/api/posts/{id}/vote", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			postHandler.SetPostVote(w, r)
		} else if r.Method == http.MethodDelete {
			postHandler.DeletePostVote(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/comments/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			postHandler.UpdateComment(w, r)
		} else if r.Method == http.MethodDelete {
			postHandler.DeleteComment(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/comments/{id}/vote", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			postHandler.SetCommentVote(w, r)
		} else if r.Method == http.MethodDelete {
			postHandler.DeleteCommentVote(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/comments/{id}/replies", authMiddleware(http.HandlerFunc(postHandler.GetCommentReplies)))
	mux.Handle("/api/comments/{id}/context", authMiddleware(http.HandlerFunc(postHandler.GetCommentContext)))

	// Groups routes
	mux.Handle("/api/groups", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			groupHandler.CreateGroup(w, r)
		} else {
			groupHandler.ListGroups(w, r)
		}
	})))
	mux.Handle("/api/groups/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			groupHandler.UpdateGroup(w, r)
		} else {
			groupHandler.GetGroup(w, r)
		}
	})))
	mux.Handle("/api/groups/{id}/join", authMiddleware(http.HandlerFunc(groupHandler.RequestJoin)))
	mux.Handle("/api/groups/{id}/invite", authMiddleware(http.HandlerFunc(groupHandler.InviteUser)))
	mux.Handle("/api/groups/{id}/respond", authMiddleware(http.HandlerFunc(groupHandler.RespondMembership)))
	mux.Handle("/api/groups/{id}/leave", authMiddleware(http.HandlerFunc(groupHandler.LeaveGroup)))
	mux.Handle("/api/groups/{id}/members", authMiddleware(http.HandlerFunc(groupHandler.ListMembers)))
	mux.Handle("/api/groups/{id}/requests", authMiddleware(http.HandlerFunc(groupHandler.ListPendingRequests)))
	mux.Handle("/api/groups/{id}/invitations", authMiddleware(http.HandlerFunc(groupHandler.ListPendingInvitations)))
	mux.Handle("/api/groups/{id}/members/{userID}/role/{action}", authMiddleware(http.HandlerFunc(groupHandler.UpdateMemberRole)))

	// Events routes
	mux.Handle("/api/groups/{id}/events", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			eventHandler.CreateEvent(w, r)
		} else {
			eventHandler.ListEvents(w, r)
		}
	})))
	mux.Handle("/api/events/{id}/rsvp", authMiddleware(http.HandlerFunc(eventHandler.RespondEvent)))

	// Chat/Messages routes
	mux.Handle("/api/chats", authMiddleware(http.HandlerFunc(chatHandler.GetConversations)))
	mux.Handle("/api/chats/dm", authMiddleware(http.HandlerFunc(chatHandler.OpenDM)))
	mux.Handle("/api/chats/dm-candidates", authMiddleware(http.HandlerFunc(chatHandler.GetDMCandidates)))
	mux.Handle("/api/chats/{id}/messages", authMiddleware(http.HandlerFunc(chatHandler.GetChatMessages)))
	mux.Handle("/api/conversations", authMiddleware(http.HandlerFunc(chatHandler.GetConversations)))
	mux.Handle("/api/messages", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			chatHandler.SendMessage(w, r)
		} else {
			chatHandler.GetMessages(w, r)
		}
	})))
	mux.Handle("/api/ws", authMiddleware(http.HandlerFunc(chatHandler.HandleWS)))

	// Notifications routes
	mux.Handle("/api/notifications", authMiddleware(http.HandlerFunc(notificationHandler.ListNotifications)))
	mux.Handle("/api/notifications/{id}/read", authMiddleware(http.HandlerFunc(notificationHandler.MarkAsRead)))
	mux.Handle("/api/notifications/read/all", authMiddleware(http.HandlerFunc(notificationHandler.MarkAllAsRead)))

	return middleware.CorsMiddleware(mux)
}
