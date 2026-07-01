import { useCallback, useEffect, useRef, useState } from "react";
import { apiFetch } from "../utils/api";
import { useSocket } from "../context/socket/useSocket";
import { useAuth } from "../context/auth/useAuth";
import "../styles/RegisterForm.css";

const Messages = () => {
  const { isConnected, subscribe } = useSocket();
  const { currentUser } = useAuth();
  const [conversations, setConversations] = useState([]);
  const [activeChat, setActiveChat] = useState(null);
  const [messages, setMessages] = useState([]);
  const [inputText, setInputText] = useState("");
  const messagesEndRef = useRef(null);
  const activeChatRef = useRef(activeChat);

  // Keep ref in sync
  useEffect(() => {
    activeChatRef.current = activeChat;
  }, [activeChat]);

  const fetchConversations = useCallback(async () => {
    try {
      const data = await apiFetch("/api/conversations");
      setConversations(data || []);
    } catch (err) {
      console.error("Failed to fetch conversations", err);
    }
  }, []);

  const fetchMessages = useCallback(async (chat) => {
    try {
      const targetId = chat.type === "dm" ? chat.thread_id : chat.group_id;
      const data = await apiFetch(
        `/api/messages?type=${chat.type}&target_id=${targetId}`
      );
      setMessages((data || []).reverse());
    } catch (err) {
      console.error("Failed to fetch messages", err);
    }
  }, []);

  // WebSocket subscription: appends incoming messages to the active chat
  // and keeps the conversation list's last-message preview in sync.
  useEffect(() => {
    if (!isConnected) return;

    const unsubscribe = subscribe("chat", (payload) => {
      const currentChat = activeChatRef.current;
      if (currentChat) {
        const isCurrentDM =
          currentChat.type === "dm" &&
          payload.dm_thread_id === currentChat.thread_id;
        const isCurrentGroup =
          currentChat.type === "group" &&
          payload.group_id === currentChat.group_id;

        if (isCurrentDM || isCurrentGroup) {
          setMessages((prev) => {
            // The sender also receives their own broadcast back over the
            // socket, so guard against adding the same message twice.
            if (prev.some((m) => m.id === payload.id)) {
              return prev;
            }
            return [...prev, payload];
          });
        }
      }

      void fetchConversations();
    });

    return unsubscribe;
  }, [isConnected, subscribe, fetchConversations]);

  // Initial load
  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchConversations();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [fetchConversations]);

  // Fetch messages when active chat changes
  useEffect(() => {
    if (activeChat) {
      const timer = window.setTimeout(() => {
        void fetchMessages(activeChat);
      }, 0);
      return () => window.clearTimeout(timer);
    }
  }, [activeChat, fetchMessages]);

  // Scroll to bottom on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSendMessage = async (e) => {
    e.preventDefault();
    if (!inputText.trim() || !activeChat) return;

    const payload = {
      content: inputText,
    };

    if (activeChat.type === "dm") {
      payload.dm_thread_id = activeChat.thread_id;
    } else {
      payload.group_id = activeChat.group_id;
    }

    try {
      await apiFetch("/api/messages", {
        method: "POST",
        body: payload,
      });
      
      setInputText("");
      await fetchConversations();
    } catch (err) {
      alert("Failed to send message: " + err.message);
    }
  };

  return (
    <div
      style={{
        display: "flex",
        height: "calc(100vh - 80px)",
        border: "1px solid #2e2e2e",
        borderRadius: "12px",
        overflow: "hidden",
        backgroundColor: "#121212",
      }}
    >
      {/* Left Conversations Sidebar */}
      <div
        style={{
          width: "300px",
          borderRight: "1px solid #2e2e2e",
          display: "flex",
          flexDirection: "column",
          backgroundColor: "#1e1e1e",
        }}
      >
        <h3
          style={{
            padding: "20px",
            margin: 0,
            borderBottom: "1px solid #2e2e2e",
            color: "white",
          }}
        >
          Chats {!isConnected && '🔴'}
        </h3>
        <div style={{ flex: 1, overflowY: "auto" }}>
          {conversations.map((c, idx) => {
            const isSelected =
              activeChat &&
              ((c.type === "dm" && c.thread_id === activeChat.thread_id) ||
                (c.type === "group" && c.group_id === activeChat.group_id));
            return (
              <div
                key={idx}
                onClick={() => setActiveChat(c)}
                style={{
                  padding: "15px 20px",
                  borderBottom: "1px solid #2e2e2e",
                  cursor: "pointer",
                  backgroundColor: isSelected ? "#2e2e2e" : "transparent",
                  transition: "background-color 0.2s ease",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    marginBottom: "5px",
                  }}
                >
                  <strong style={{ color: "white" }}>{c.target_name}</strong>
                  {c.type === "group" && (
                    <span
                      style={{
                        fontSize: "0.75rem",
                        backgroundColor: "#764ba2",
                        color: "white",
                        padding: "2px 6px",
                        borderRadius: "4px",
                      }}
                    >
                      group
                    </span>
                  )}
                </div>
                <div
                  style={{
                    color: "#aaa",
                    fontSize: "0.85rem",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    overflow: "hidden",
                  }}
                >
                  {c.last_message || "No messages yet."}
                </div>
              </div>
            );
          })}
          {conversations.length === 0 && (
            <div
              style={{
                color: "#888",
                textAlign: "center",
                padding: "40px 20px",
              }}
            >
              No active chats. Start messaging by following someone!
            </div>
          )}
        </div>
      </div>

      {/* Right Messages Area */}
      <div
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          backgroundColor: "#121212",
        }}
      >
        {activeChat ? (
          <>
            {/* Header */}
            <div
              style={{
                padding: "20px",
                borderBottom: "1px solid #2e2e2e",
                backgroundColor: "#1e1e1e",
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
              }}
            >
              <h3 style={{ margin: 0, color: "white" }}>
                {activeChat.target_name}
              </h3>
              <span style={{ color: '#888', fontSize: '0.8rem' }}>
                {isConnected ? '🟢 Online' : '🔴 Offline'}
              </span>
            </div>

            {/* Message List */}
            <div
              style={{
                flex: 1,
                padding: "20px",
                overflowY: "auto",
                display: "flex",
                flexDirection: "column",
                gap: "10px",
              }}
            >
              {messages.map((m) => {
                const isOwnMessage = m.sender_id === currentUser?.id;
                return (
                  <div
                    key={m.id}
                    style={{
                      alignSelf: isOwnMessage ? "flex-end" : "flex-start",
                      backgroundColor: isOwnMessage ? "#667eea" : "#2e2e2e",
                      color: "white",
                      padding: "10px 15px",
                      borderRadius: "12px",
                      maxWidth: "70%",
                      wordBreak: "break-word",
                    }}
                  >
                    <p style={{ margin: 0 }}>{m.content}</p>
                    <small
                      style={{
                        display: "block",
                        fontSize: "0.7rem",
                        color: "rgba(255,255,255,0.6)",
                        marginTop: "5px",
                        textAlign: "right",
                      }}
                    >
                      {new Date(m.created_at).toLocaleTimeString([], {
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </small>
                  </div>
                );
              })}
              <div ref={messagesEndRef} />
            </div>

            {/* Input Form */}
            <form
              onSubmit={handleSendMessage}
              style={{
                padding: "20px",
                borderTop: "1px solid #2e2e2e",
                backgroundColor: "#1e1e1e",
                display: "flex",
                gap: "10px",
              }}
            >
              <input
                type="text"
                placeholder="Type your message..."
                value={inputText}
                onChange={(e) => setInputText(e.target.value)}
                style={{
                  flex: 1,
                  padding: "12px",
                  borderRadius: "8px",
                  border: "1px solid #444",
                  backgroundColor: "#2e2e2e",
                  color: "white",
                  outline: "none",
                }}
              />
              <button
                type="submit"
                style={{
                  background: "linear-gradient(135deg, #667eea, #764ba2)",
                  color: "white",
                  border: "none",
                  padding: "0 20px",
                  borderRadius: "8px",
                  fontWeight: "bold",
                  cursor: "pointer",
                }}
              >
                Send
              </button>
            </form>
          </>
        ) : (
          <div
            style={{
              flex: 1,
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
              color: "#888",
            }}
          >
            Select a conversation to start chatting.
          </div>
        )}
      </div>
    </div>
  );
};

export default Messages;