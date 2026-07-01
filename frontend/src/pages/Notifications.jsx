import { useCallback, useEffect, useState } from "react";
import { apiFetch } from "../utils/api";
import { useAuth } from "../context/auth/useAuth";
import { useSocket } from "../context/socket/useSocket";
import "../styles/notifications.css";

// Per-type icon and label, used both for the list and to visually separate
// notifications from chat messages elsewhere in the app.
const NOTIFICATION_META = {
  follow_request: { icon: "👤", label: "Follow request" },
  group_invite: { icon: "👥", label: "Group invite" },
  group_request: { icon: "🙋", label: "Join request" },
  event_created: { icon: "📅", label: "New event" },
  event_invite: { icon: "🎟️", label: "Event invite" },
};

const DEFAULT_META = { icon: "🔔", label: "Notification" };

/**
 * Resolves the accept/decline request for a single notification's action
 * buttons. Returns null for notification types that aren't actionable
 * (the caller just renders no buttons in that case).
 *
 * @param {object} notification
 * @param {"accept"|"decline"} action
 * @returns {{url: string, body: object}|null}
 */
const buildActionRequest = (notification, action) => {
  switch (notification.type) {
    case "follow_request":
      return {
        url: `/api/followers/${action === "accept" ? "accept" : "reject"}`,
        body: { follower_id: notification.source_id },
      };
    case "group_invite":
      return {
        url: `/api/groups/${notification.source_id}/respond`,
        body: { action: action === "accept" ? "accept" : "reject" },
      };
    case "group_request":
      return {
        url: `/api/groups/${notification.group_id}/respond`,
        body: {
          user_id: notification.source_id,
          action: action === "accept" ? "accept" : "reject",
        },
      };
    default:
      return null;
  }
};

const Notifications = () => {
  const [notifications, setNotifications] = useState([]);
  const [actioningId, setActioningId] = useState(null);
  const { refreshUnreadNotifications } = useAuth();
  const { subscribe } = useSocket();

  const fetchNotifications = useCallback(async () => {
    try {
      const data = await apiFetch("/api/notifications");
      setNotifications(data || []);
    } catch (err) {
      console.error("Failed to fetch notifications", err);
    }
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchNotifications();
      void refreshUnreadNotifications();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [fetchNotifications, refreshUnreadNotifications]);

  // Prepend newly-pushed notifications without waiting for a refetch.
  useEffect(() => {
    const unsubscribe = subscribe("notification", (payload) => {
      setNotifications((prev) => {
        if (prev.some((n) => n.id === payload.id)) {
          return prev;
        }
        return [payload, ...prev];
      });
    });

    return unsubscribe;
  }, [subscribe]);

  const handleMarkAsRead = async (nId) => {
    try {
      await apiFetch(`/api/notifications/${nId}/read`, { method: "POST" });
      await Promise.all([fetchNotifications(), refreshUnreadNotifications()]);
    } catch (err) {
      alert("Failed to mark as read: " + err.message);
    }
  };

  const handleMarkAllAsRead = async () => {
    try {
      await apiFetch("/api/notifications/read/all", { method: "POST" });
      await Promise.all([fetchNotifications(), refreshUnreadNotifications()]);
    } catch (err) {
      alert("Failed to mark all as read: " + err.message);
    }
  };

  const handleAction = async (notification, action) => {
    const request = buildActionRequest(notification, action);
    if (!request) return;

    setActioningId(notification.id);
    try {
      await apiFetch(request.url, { method: "POST", body: request.body });
      // The request has been resolved either way, so it's no longer
      // actionable — mark it read and drop it from the unread state.
      await apiFetch(`/api/notifications/${notification.id}/read`, {
        method: "POST",
      });
      await Promise.all([fetchNotifications(), refreshUnreadNotifications()]);
    } catch (err) {
      alert(`Failed to ${action} request: ` + err.message);
    } finally {
      setActioningId(null);
    }
  };

  const isActionable = (n) =>
    !n.is_read &&
    (n.type === "follow_request" ||
      n.type === "group_invite" ||
      n.type === "group_request");

  return (
    <div className="notifications-page">
      <div className="notifications-header">
        <h2>Notifications</h2>
        {notifications.some((n) => !n.is_read) && (
          <button className="notifications-mark-all" onClick={handleMarkAllAsRead}>
            Mark all as read
          </button>
        )}
      </div>

      <div className="notifications-list">
        {notifications.map((n) => {
          const meta = NOTIFICATION_META[n.type] || DEFAULT_META;
          const busy = actioningId === n.id;

          return (
            <div
              key={n.id}
              className={`notifications-card notifications-card--${n.type} ${
                n.is_read ? "notifications-card--read" : "notifications-card--unread"
              }`}
            >
              <span className="notifications-card-icon" aria-hidden="true">
                {meta.icon}
              </span>

              <div className="notifications-card-body">
                <span className="notifications-card-type">{meta.label}</span>
                <p className="notifications-card-message">{n.message}</p>
                <small className="notifications-card-time">
                  {new Date(n.created_at).toLocaleString()}
                </small>
              </div>

              <div className="notifications-card-actions">
                {isActionable(n) ? (
                  <>
                    <button
                      className="notifications-action notifications-action--accept"
                      disabled={busy}
                      onClick={() => handleAction(n, "accept")}
                    >
                      Accept
                    </button>
                    <button
                      className="notifications-action notifications-action--decline"
                      disabled={busy}
                      onClick={() => handleAction(n, "decline")}
                    >
                      Decline
                    </button>
                  </>
                ) : (
                  !n.is_read && (
                    <button
                      className="notifications-action"
                      onClick={() => handleMarkAsRead(n.id)}
                    >
                      Mark Read
                    </button>
                  )
                )}
              </div>
            </div>
          );
        })}

        {notifications.length === 0 && (
          <div className="notifications-empty">You have no notifications yet.</div>
        )}
      </div>
    </div>
  );
};

export default Notifications;