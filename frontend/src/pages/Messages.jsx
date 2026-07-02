import { useCallback, useEffect, useRef, useState } from "react";
import { apiFetch } from "../utils/api";
import { useSocket } from "../context/socket/useSocket";
import { useAuth } from "../context/auth/useAuth";
import "../styles/messages.css";

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

  // WebSocket subscription
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
      await fetchMessages(activeChat);
      await fetchConversations();
    } catch (err) {
      alert("Failed to send message: " + err.message);
    }
  };

  return (
    <div className="messages-container">
      {/* Left Conversations Sidebar */}
      <div className="messages-sidebar">
        <h3 className="messages-sidebar-header">
          Chats {!isConnected && '🔴'}
        </h3>
        <div className="messages-conversation-list">
          {conversations.map((c, idx) => {
            const isSelected =
              activeChat &&
              ((c.type === "dm" && c.thread_id === activeChat.thread_id) ||
                (c.type === "group" && c.group_id === activeChat.group_id));
            return (
              <div
                key={idx}
                onClick={() => setActiveChat(c)}
                className={`messages-conversation-item ${
                  isSelected ? "messages-conversation-item--selected" : ""
                }`}
              >
                <div className="messages-conversation-item-header">
                  <strong>{c.target_name}</strong>
                  {c.type === "group" && (
                    <span className="messages-group-badge">group</span>
                  )}
                </div>
                <div className="messages-conversation-item-preview">
                  {c.last_message || "No messages yet."}
                </div>
              </div>
            );
          })}
          {conversations.length === 0 && (
            <div className="messages-empty-state">
              No active chats. Start messaging by following someone!
            </div>
          )}
        </div>
      </div>

      {/* Right Messages Area */}
      <div className="messages-chat-area">
        {activeChat ? (
          <>
            {/* Header */}
            <div className="messages-chat-header">
              <h3>{activeChat.target_name}</h3>
              <span className="messages-connection-status">
                {isConnected ? '🟢 Online' : '🔴 Offline'}
              </span>
            </div>

            {/* Message List */}
            <div className="messages-list">
              {messages.map((m) => {
                const isOwnMessage = m.sender_id === currentUser?.id;
                return (
                  <div
                    key={m.id}
                    className={`messages-message ${
                      isOwnMessage
                        ? "messages-message--own"
                        : "messages-message--other"
                    }`}
                  >
                    <p>{m.content}</p>
                    <small>
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
            <form className="messages-input-form" onSubmit={handleSendMessage}>
              <input
                type="text"
                placeholder="Type your message..."
                value={inputText}
                onChange={(e) => setInputText(e.target.value)}
                className="messages-input"
              />
              <button type="submit" className="messages-send-button">
                Send
              </button>
            </form>
          </>
        ) : (
          <div className="messages-placeholder">
            Select a conversation to start chatting.
          </div>
        )}
      </div>
    </div>
  );
};

export default Messages;