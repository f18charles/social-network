import { useCallback, useEffect, useRef, useState } from "react";
import { useLocation } from "react-router";
import { apiFetch } from "../utils/api";
import { useSocket } from "../context/socket/useSocket";
import { useAuth } from "../context/auth/useAuth";
import "../styles/messages.css";

const Messages = () => {
  const location = useLocation();
  const { isConnected, subscribe, send } = useSocket();
  const { currentUser } = useAuth();
  const [conversations, setConversations] = useState([]);
  const [dmCandidates, setDMCandidates] = useState([]);
  const [activeChat, setActiveChat] = useState(null);
  const [messages, setMessages] = useState([]);
  const [inputText, setInputText] = useState("");
  const messagesEndRef = useRef(null);
  const activeChatRef = useRef(activeChat);

  useEffect(() => {
    activeChatRef.current = activeChat;
  }, [activeChat]);

  const fetchConversations = useCallback(async () => {
    try {
      const data = await apiFetch("/api/chats");
      setConversations(data || []);
    } catch (err) {
      console.error("Failed to fetch conversations", err);
    }
  }, []);

  const fetchDMCandidates = useCallback(async () => {
    try {
      const data = await apiFetch("/api/chats/dm-candidates?limit=10");
      setDMCandidates(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error("Failed to fetch DM candidates", err);
    }
  }, []);

  const fetchMessages = useCallback(async (chat) => {
    try {
      const targetId = chat.type === "dm" ? chat.thread_id : chat.group_id;
      const data = await apiFetch(
        `/api/chats/${targetId}/messages?chat_type=${chat.type}&limit=100`
      );
      setMessages(data || []);
    } catch (err) {
      console.error("Failed to fetch messages", err);
    }
  }, []);

  useEffect(() => {
    if (!isConnected) return undefined;

    const unsubscribe = subscribe("message.created", (payload) => {
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
      void fetchDMCandidates();
    });

    return unsubscribe;
  }, [isConnected, subscribe, fetchConversations, fetchDMCandidates]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchConversations();
      void fetchDMCandidates();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [fetchConversations, fetchDMCandidates]);

  useEffect(() => {
    const selected = location.state?.selectedConversation;
    if (!selected) return undefined;

    const timer = window.setTimeout(() => {
      setActiveChat(selected);
    }, 0);
    return () => window.clearTimeout(timer);
  }, [location.state]);

  useEffect(() => {
    if (!activeChat) return undefined;
    const timer = window.setTimeout(() => {
      void fetchMessages(activeChat);
    }, 0);
    return () => window.clearTimeout(timer);
  }, [activeChat, fetchMessages]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSendMessage = async (event) => {
    event.preventDefault();
    if (!inputText.trim() || !activeChat) return;

    const payload = { content: inputText };
    if (activeChat.type === "dm") {
      payload.dm_thread_id = activeChat.thread_id;
    } else {
      payload.group_id = activeChat.group_id;
    }

    try {
      const sent = send({
        type: "message.send",
        chat_id:
          activeChat.type === "dm" ? activeChat.thread_id : activeChat.group_id,
        chat_type: activeChat.type,
        content: inputText,
        client_message_id: `client-${Date.now()}`,
      });
      if (!sent) {
        await apiFetch("/api/messages", {
          method: "POST",
          body: payload,
        });
        await fetchMessages(activeChat);
      }

      setInputText("");
      await fetchConversations();
      await fetchDMCandidates();
    } catch (err) {
      alert("Failed to send message: " + err.message);
    }
  };

  const handleOpenDM = async (user) => {
    try {
      const conversation = await apiFetch("/api/chats/dm", {
        method: "POST",
        body: { recipient_id: user.id },
      });
      setActiveChat(conversation);
      setMessages([]);
      await fetchConversations();
      await fetchDMCandidates();
    } catch (err) {
      alert(err.message || "Unable to start direct message.");
    }
  };

  return (
    <div className="messages-container">
      <div className="messages-sidebar">
        <h3 className="messages-sidebar-header">
          Chats {!isConnected && "Offline"}
        </h3>
        <div className="messages-conversation-list">
          {conversations.map((conversation) => {
            const key =
              conversation.type === "dm"
                ? `dm-${conversation.thread_id}`
                : `group-${conversation.group_id}`;
            const isSelected =
              activeChat &&
              ((conversation.type === "dm" &&
                conversation.thread_id === activeChat.thread_id) ||
                (conversation.type === "group" &&
                  conversation.group_id === activeChat.group_id));

            return (
              <div
                key={key}
                onClick={() => setActiveChat(conversation)}
                className={`messages-conversation-item ${
                  isSelected ? "messages-conversation-item--selected" : ""
                }`}
              >
                <div className="messages-conversation-item-header">
                  <strong>{conversation.target_name}</strong>
                  {conversation.type === "group" && (
                    <span className="messages-group-badge">group</span>
                  )}
                </div>
                <div className="messages-conversation-item-preview">
                  {conversation.last_message || "No messages yet."}
                </div>
              </div>
            );
          })}
          {conversations.length === 0 && (
            <div className="messages-empty-state">No active chats yet.</div>
          )}
        </div>

        <div className="messages-starter-list">
          <h4>Start a DM</h4>
          {dmCandidates.length === 0 ? (
            <div className="messages-empty-state messages-empty-state--compact">
              No people available.
            </div>
          ) : (
            dmCandidates.map((user) => {
              const name =
                `${user.first_name || ""} ${user.last_name || ""}`.trim() ||
                user.nickname ||
                user.email;
              return (
                <button
                  type="button"
                  key={user.id}
                  className="messages-starter-item"
                  onClick={() => handleOpenDM(user)}
                >
                  <span>{name}</span>
                  {user.nickname && <small>@{user.nickname}</small>}
                </button>
              );
            })
          )}
        </div>
      </div>

      <div className="messages-chat-area">
        {activeChat ? (
          <>
            <div className="messages-chat-header">
              <h3>{activeChat.target_name}</h3>
              <span className="messages-connection-status">
                {isConnected ? "Online" : "Offline"}
              </span>
            </div>

            <div className="messages-list">
              {messages.map((message) => {
                const isOwnMessage = message.sender_id === currentUser?.id;
                return (
                  <div
                    key={message.id}
                    className={`messages-message ${
                      isOwnMessage
                        ? "messages-message--own"
                        : "messages-message--other"
                    }`}
                  >
                    <p>{message.content}</p>
                    <small>
                      {new Date(message.created_at).toLocaleTimeString([], {
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </small>
                  </div>
                );
              })}
              <div ref={messagesEndRef} />
            </div>

            <form className="messages-input-form" onSubmit={handleSendMessage}>
              <input
                type="text"
                placeholder="Type your message..."
                value={inputText}
                onChange={(event) => setInputText(event.target.value)}
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
