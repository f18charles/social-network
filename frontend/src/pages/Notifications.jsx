import { useCallback, useEffect, useState } from "react";
import { apiFetch } from "../utils/api";
import { useAuth } from "../context/auth/useAuth";

const Notifications = () => {
  const [notifications, setNotifications] = useState([]);
  const { refreshUnreadNotifications } = useAuth();

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

  return (
    <div style={{ padding: "20px", maxWidth: "600px", margin: "0 auto" }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "20px",
        }}
      >
        <h2 style={{ margin: 0 }}>Notifications</h2>
        {notifications.some((n) => !n.is_read) && (
          <button
            onClick={handleMarkAllAsRead}
            style={{
              backgroundColor: "transparent",
              color: "#667eea",
              border: "1px solid #667eea",
              padding: "8px 16px",
              borderRadius: "6px",
              fontWeight: "bold",
              cursor: "pointer",
            }}
          >
            Mark all as read
          </button>
        )}
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: "15px" }}>
        {notifications.map((n) => (
          <div
            key={n.id}
            style={{
              backgroundColor: n.is_read ? "#181818" : "#222222",
              border: n.is_read ? "1px solid #2e2e2e" : "1px solid #667eea",
              borderRadius: "8px",
              padding: "15px 20px",
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              transition: "all 0.3s ease",
            }}
          >
            <div>
              <p
                style={{
                  margin: "0 0 5px 0",
                  color: "white",
                  fontSize: "0.95rem",
                }}
              >
                {n.message}
              </p>
              <small style={{ color: "#888" }}>
                {new Date(n.created_at).toLocaleString()}
              </small>
            </div>

            {!n.is_read && (
              <button
                onClick={() => handleMarkAsRead(n.id)}
                style={{
                  backgroundColor: "#667eea",
                  color: "white",
                  border: "none",
                  padding: "6px 12px",
                  borderRadius: "4px",
                  cursor: "pointer",
                  fontSize: "0.85rem",
                  fontWeight: "bold",
                }}
              >
                Mark Read
              </button>
            )}
          </div>
        ))}

        {notifications.length === 0 && (
          <div style={{ color: "#888", textAlign: "center", padding: "40px" }}>
            You have no notifications yet.
          </div>
        )}
      </div>
    </div>
  );
};

export default Notifications;
