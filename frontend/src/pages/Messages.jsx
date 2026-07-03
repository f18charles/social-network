import { useCallback, useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router";
import { apiFetch } from "../utils/api";
import { useSocket } from "../context/socket/useSocket";
import { useAuth } from "../context/auth/useAuth";
import avatarFallback from "../assets/user.svg";
import { AiOutlineLoading3Quarters } from "react-icons/ai";
import "../styles/messages.css";

const formatDateMarker = (dateStr) => {
  const date = new Date(dateStr);
  const today = new Date();
  const yesterday = new Date();
  yesterday.setDate(today.getDate() - 1);

  if (date.toDateString() === today.toDateString()) {
    return "Today";
  } else if (date.toDateString() === yesterday.toDateString()) {
    return "Yesterday";
  } else {
    return date.toLocaleDateString([], {
      weekday: "long",
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  }
};

const Messages = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { isConnected, subscribe, send } = useSocket();
  const { currentUser } = useAuth();
  const [conversations, setConversations] = useState([]);
  const [dmCandidates, setDMCandidates] = useState([]);
  const [activeChat, setActiveChat] = useState(null);
  const [messages, setMessages] = useState([]);
  const [hasMore, setHasMore] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [headerMenuOpen, setHeaderMenuOpen] = useState(false);
  const [inputText, setInputText] = useState("");
  const [messageReceipts, setMessageReceipts] = useState({});
  const [activeGroupMembers, setActiveGroupMembers] = useState([]);
  const messagesEndRef = useRef(null);
  const messagesContainerRef = useRef(null);
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

  const fetchMessages = useCallback(async (chat, appendOffset = 0) => {
    try {
      const targetId = chat.type === "dm" ? chat.thread_id : chat.group_id;
      const limit = 20;
      const data = await apiFetch(
        `/api/chats/${targetId}/messages?chat_type=${chat.type}&limit=${limit}&offset=${appendOffset}`
      );
      
      const loadedMessages = data || [];
      if (loadedMessages.length < limit) {
        setHasMore(false);
      } else {
        setHasMore(true);
      }

      if (appendOffset === 0) {
        setMessages(loadedMessages);
        setTimeout(() => {
          if (messagesContainerRef.current) {
            messagesContainerRef.current.scrollTop = messagesContainerRef.current.scrollHeight;
          }
        }, 50);
      } else {
        const container = messagesContainerRef.current;
        const prevScrollHeight = container ? container.scrollHeight : 0;
        const prevScrollTop = container ? container.scrollTop : 0;

        setMessages((prev) => [...loadedMessages, ...prev]);

        setTimeout(() => {
          if (container) {
            const newScrollHeight = container.scrollHeight;
            container.scrollTop = prevScrollTop + (newScrollHeight - prevScrollHeight);
          }
        }, 50);
      }
    } catch (err) {
      console.error("Failed to fetch messages", err);
    }
  }, []);

  useEffect(() => {
    if (!isConnected) return undefined;

    const unsub1 = subscribe("message.created", (payload) => {
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
            const clientMsgId = payload.client_message_id;
            const existingIdx = prev.findIndex(
              (m) => m.id === clientMsgId || m.id === payload.id
            );

            if (existingIdx > -1) {
              const next = [...prev];
              next[existingIdx] = { ...payload, status: "saved" };
              return next;
            }

            return [...prev, { ...payload, status: "saved" }];
          });
        }

        // Send a delivery receipt back to the sender if we received a message from someone else
        if (payload.sender_id !== currentUser?.id) {
          send({
            type: "message.received",
            chat_id: currentChat.type === "dm" ? currentChat.thread_id : currentChat.group_id,
            chat_type: currentChat.type,
            message_id: payload.id,
            sender_id: payload.sender_id,
          });
        }
      }

      void fetchConversations();
      void fetchDMCandidates();
    });

    const unsub2 = subscribe("messages.cleared", (payload) => {
      const currentChat = activeChatRef.current;
      if (currentChat) {
        const targetId = currentChat.type === "dm" ? currentChat.thread_id : currentChat.group_id;
        if (payload.chat_id === targetId && payload.chat_type === currentChat.type) {
          setMessages((prev) => prev.filter((m) => m.sender_id !== payload.sender_id));
        }
      }
      void fetchConversations();
    });

    const unsub3 = subscribe("message.received", (payload) => {
      setMessageReceipts((prev) => {
        const existing = prev[payload.message_id] || [];
        if (existing.includes(payload.receiver_id)) return prev;
        return {
          ...prev,
          [payload.message_id]: [...existing, payload.receiver_id],
        };
      });
    });

    return () => {
      unsub1();
      unsub2();
      unsub3();
    };
  }, [isConnected, subscribe, send, currentUser, fetchConversations, fetchDMCandidates]);

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
    setHasMore(true);
    setLoadingMore(false);
    setHeaderMenuOpen(false);

    // Fetch group members if group chat is selected
    const fetchGroupMembers = async () => {
      if (activeChat.type === "group") {
        try {
          const data = await apiFetch(`/api/groups/${activeChat.group_id}/members`);
          setActiveGroupMembers(data || []);
        } catch (err) {
          console.error("Failed to fetch group members", err);
          setActiveGroupMembers([]);
        }
      } else {
        setActiveGroupMembers([]);
      }
    };
    void fetchGroupMembers();

    const timer = window.setTimeout(() => {
      void fetchMessages(activeChat, 0);
    }, 0);
    return () => window.clearTimeout(timer);
  }, [activeChat, fetchMessages]);

  const handleScroll = async (e) => {
    const container = e.currentTarget;
    if (container.scrollTop === 0 && hasMore && !loadingMore && messages.length > 0) {
      setLoadingMore(true);
      await fetchMessages(activeChat, messages.length);
      setLoadingMore(false);
    }
  };

  const prevMessagesLength = useRef(messages.length);
  useEffect(() => {
    const prevLen = prevMessagesLength.current;
    prevMessagesLength.current = messages.length;

    if (
      messages.length === prevLen + 1 &&
      messages[0]?.id === messages[0]?.id
    ) {
      messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  const sendMessageWithRetry = async (content, existingClientMsgId = null) => {
    const clientMsgId = existingClientMsgId || `client-${Date.now()}`;
    
    const tempMessage = {
      id: clientMsgId,
      sender_id: currentUser?.id,
      content: content,
      created_at: new Date().toISOString(),
      status: "sending",
    };

    setMessages((prev) => {
      const idx = prev.findIndex((m) => m.id === clientMsgId);
      if (idx > -1) {
        const next = [...prev];
        next[idx] = tempMessage;
        return next;
      }
      return [...prev, tempMessage];
    });

    try {
      const sent = send({
        type: "message.send",
        chat_id:
          activeChat.type === "dm" ? activeChat.thread_id : activeChat.group_id,
        chat_type: activeChat.type,
        content: content,
        client_message_id: clientMsgId,
      });

      if (!sent) {
        const payload = { content };
        if (activeChat.type === "dm") {
          payload.dm_thread_id = activeChat.thread_id;
        } else {
          payload.group_id = activeChat.group_id;
        }

        const data = await apiFetch("/api/messages", {
          method: "POST",
          body: payload,
        });

        setMessages((prev) =>
          prev.map((m) =>
            m.id === clientMsgId
              ? { ...m, ...data, status: "saved" }
              : m
          )
        );
      }
    } catch (err) {
      console.error("Failed to send message", err);
      setMessages((prev) =>
        prev.map((m) =>
          m.id === clientMsgId
            ? { ...m, status: "failed" }
            : m
        )
      );
    }
  };

  const handleDeleteMessage = async (messageId) => {
    if (window.confirm("Are you sure you want to delete this message?")) {
      try {
        await apiFetch(`/api/messages/${messageId}`, {
          method: "DELETE",
        });
        setMessages((prev) =>
          prev.map((m) =>
            m.id === messageId
              ? { ...m, deleted_at: new Date().toISOString(), content: "This message is no longer available" }
              : m
          )
        );
      } catch (err) {
        alert("Failed to delete message: " + err.message);
      }
    }
  };

  const handleClearChat = async () => {
    setHeaderMenuOpen(false);
    if (window.confirm("Are you sure you want to delete all your messages in this chat? This action cannot be undone.")) {
      try {
        const targetId = activeChat.type === "dm" ? activeChat.thread_id : activeChat.group_id;
        await apiFetch(`/api/messages?chat_id=${targetId}&chat_type=${activeChat.type}`, {
          method: "DELETE",
        });
        setMessages((prev) => prev.filter((m) => m.sender_id !== currentUser?.id));
      } catch (err) {
        alert("Failed to clear messages: " + err.message);
      }
    }
  };

  const handleSendMessage = async (event) => {
    event.preventDefault();
    if (!inputText.trim() || !activeChat) return;

    const content = inputText.trim();
    setInputText("");
    await sendMessageWithRetry(content);
    await fetchConversations();
    await fetchDMCandidates();
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

  const handleConversationRedirect = (e, conversation) => {
    e.stopPropagation();
    if (conversation.type === "dm" && conversation.target_id) {
      navigate(`/user/${conversation.target_id}`);
    } else if (conversation.type === "group" && conversation.group_id) {
      navigate(`/groups/${conversation.group_id}`);
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
                <img
                  src={conversation.target_avatar || avatarFallback}
                  alt={conversation.target_name}
                  className="messages-conversation-item__avatar"
                  onClick={(e) => handleConversationRedirect(e, conversation)}
                  style={{
                    cursor: "pointer"
                  }}
                />
                <div className="messages-conversation-item-body">
                  <div className="messages-conversation-item-header">
                    <strong
                      onClick={(e) => handleConversationRedirect(e, conversation)}
                      style={{
                        cursor: "pointer"
                      }}
                    >
                      {conversation.target_name}
                    </strong>
                    {conversation.type === "group" && (
                      <span className="messages-group-badge">group</span>
                    )}
                  </div>
                  <div className="messages-conversation-item-preview">
                    {conversation.last_message || "No messages yet."}
                  </div>
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
              <div className="messages-chat-header-info">
                <img
                  src={activeChat.target_avatar || avatarFallback}
                  alt={activeChat.target_name}
                  className="messages-chat-header__avatar"
                  onClick={(e) => handleConversationRedirect(e, activeChat)}
                  style={{
                    cursor: "pointer"
                  }}
                />
                <h3
                  onClick={(e) => handleConversationRedirect(e, activeChat)}
                  style={{
                    cursor: "pointer"
                  }}
                >
                  {activeChat.target_name}
                </h3>
              </div>
              <div className="messages-chat-header-actions">
                <span className="messages-connection-status">
                  {isConnected ? "Online" : "Offline"}
                </span>
                <div className="messages-header-dropdown-container">
                  <button
                    type="button"
                    className="messages-header-more-btn"
                    onClick={() => setHeaderMenuOpen(!headerMenuOpen)}
                    title="Chat settings"
                  >
                    •••
                  </button>
                  {headerMenuOpen && (
                    <>
                      <div className="messages-header-dropdown-backdrop" onClick={() => setHeaderMenuOpen(false)} />
                      <div className="messages-header-dropdown">
                        <button
                          type="button"
                          className="dropdown-item dropdown-item--danger"
                          onClick={handleClearChat}
                        >
                          Clear my messages
                        </button>
                      </div>
                    </>
                  )}
                </div>
              </div>
            </div>

            <div
              className="messages-list"
              ref={messagesContainerRef}
              onScroll={handleScroll}
            >
              {loadingMore && (
                <div className="messages-loading-more">Loading older messages...</div>
              )}
              {messages.map((message, index) => {
                const prevMessage = index > 0 ? messages[index - 1] : null;
                const currentDate = new Date(message.created_at).toDateString();
                const prevDate = prevMessage ? new Date(prevMessage.created_at).toDateString() : null;
                const isDateChanged = currentDate !== prevDate;
                const isOwnMessage = message.sender_id === currentUser?.id;
                const isDeleted = !!message.deleted_at;
                const isSystem = message.message_type === "system";
                
                return (
                  <div key={message.id} style={{ display: "flex", flexDirection: "column", width: "100%" }}>
                    {isDateChanged && !isDeleted && !isSystem && (
                      <div className="messages-date-marker">
                        <span>{formatDateMarker(message.created_at)}</span>
                      </div>
                    )}
                    <div
                      className={`messages-message ${
                        isDeleted
                          ? "messages-message--tombstone"
                          : isSystem
                          ? "messages-message--system"
                          : isOwnMessage
                          ? "messages-message--own"
                          : "messages-message--other"
                      }`}
                    >
                      <p>{message.content}</p>
                      {!isDeleted && !isSystem && (
                        <div className="messages-message-meta">
                          <small>
                            {new Date(message.created_at).toLocaleTimeString([], {
                              hour: "2-digit",
                              minute: "2-digit",
                            })}
                          </small>
                          {isOwnMessage && (
                            <span className="messages-message-status">
                              {message.status === "sending" && (
                                <span className="status-sending" title="Sending" style={{ display: "inline-block", verticalAlign: "middle" }}>
                                  <AiOutlineLoading3Quarters className="animate-spin" size={10} />
                                </span>
                              )}
                              {(message.status === "saved" || message.status === "sent" || !message.status) && (
                                <>
                                  {(() => {
                                    const receipts = messageReceipts[message.id] || [];
                                    const isDM = !activeChat || activeChat.type === "dm";
                                    const otherMembersCount = isDM
                                      ? 1
                                      : activeGroupMembers.length - 1;
                                    const isDelivered =
                                      !message.status || receipts.length >= otherMembersCount;
                                    
                                    if (isDelivered) {
                                      return <span className="status-sent" title="Delivered">✓✓</span>;
                                    } else {
                                      return <span className="status-saved" title="Saved">✓</span>;
                                    }
                                  })()}
                                  <button
                                    type="button"
                                    className="messages-message-delete-btn"
                                    onClick={() => handleDeleteMessage(message.id)}
                                    title="Delete message"
                                  >
                                    Delete
                                  </button>
                                </>
                              )}
                              {message.status === "failed" && (
                                <span className="status-failed">
                                  <span title="Failed to send">⚠️</span>
                                  <button
                                    type="button"
                                    className="messages-retry-button"
                                    onClick={() => sendMessageWithRetry(message.content, message.id)}
                                  >
                                    Retry
                                  </button>
                                </span>
                              )}
                            </span>
                          )}
                        </div>
                      )}
                    </div>
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
