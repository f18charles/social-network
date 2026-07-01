import { useCallback, useEffect, useState } from "react";
import { apiFetch } from "../utils/api";

const Events = () => {
  const [groups, setGroups] = useState([]);
  const [events, setEvents] = useState([]);
  const [selectedGroup, setSelectedGroup] = useState("");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [eventDate, setEventDate] = useState("");
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [loading, setLoading] = useState(false);

  const fetchGroupsAndEvents = useCallback(async () => {
    try {
      const groupList = await apiFetch("/api/groups");
      const acceptedGroups = (groupList || []).filter(
        (g) => g.status === "accepted"
      );
      setGroups(acceptedGroups);

      // Fetch events for all accepted groups
      const allEvents = [];
      for (const group of acceptedGroups) {
        try {
          const groupEvents = await apiFetch(`/api/groups/${group.id}/events`);
          if (groupEvents) {
            allEvents.push(
              ...groupEvents.map((e) => ({ ...e, groupTitle: group.title }))
            );
          }
        } catch (err) {
          console.error(`Failed to fetch events for group ${group.id}:`, err);
        }
      }
      setEvents(allEvents);
    } catch (err) {
      console.error(err);
    }
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchGroupsAndEvents();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [fetchGroupsAndEvents]);

  const handleCreateEvent = async (e) => {
    e.preventDefault();
    if (!selectedGroup || !title || !eventDate) return;
    setLoading(true);
    try {
      // Parse event date to RFC3339 format
      const formattedDate = new Date(eventDate).toISOString();
      await apiFetch(`/api/groups/${selectedGroup}/events`, {
        method: "POST",
        body: { title, description, event_date: formattedDate },
      });
      setTitle("");
      setDescription("");
      setEventDate("");
      setShowCreateModal(false);
      await fetchGroupsAndEvents();
    } catch (err) {
      alert(err.message || "Failed to create event");
    } finally {
      setLoading(false);
    }
  };

  const handleRSVP = async (eventId, status) => {
    try {
      await apiFetch(`/api/events/${eventId}/rsvp`, {
        method: "POST",
        body: { status },
      });
      await fetchGroupsAndEvents();
    } catch (err) {
      alert(err.message || "Failed to send RSVP");
    }
  };

  return (
    <div style={{ padding: "20px", maxWidth: "800px", margin: "0 auto" }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "20px",
        }}
      >
        <h2 style={{ margin: 0 }}>Group Events</h2>
        {groups.length > 0 && (
          <button
            onClick={() => {
              setSelectedGroup(groups[0]?.id || "");
              setShowCreateModal(true);
            }}
            style={{
              background: "linear-gradient(135deg, #11998e, #38ef7d)",
              color: "white",
              border: "none",
              padding: "10px 20px",
              borderRadius: "8px",
              fontWeight: "bold",
              cursor: "pointer",
            }}
          >
            Create Event
          </button>
        )}
      </div>

      {showCreateModal && (
        <div
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: "rgba(0,0,0,0.5)",
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: "#1e1e1e",
              padding: "30px",
              borderRadius: "12px",
              width: "450px",
              color: "white",
            }}
          >
            <h3 style={{ marginTop: 0 }}>Create Group Event</h3>
            <form onSubmit={handleCreateEvent}>
              <div style={{ marginBottom: "15px" }}>
                <label style={{ display: "block", marginBottom: "5px" }}>
                  Select Group
                </label>
                <select
                  value={selectedGroup}
                  onChange={(e) => setSelectedGroup(e.target.value)}
                  style={{
                    width: "100%",
                    padding: "10px",
                    borderRadius: "6px",
                    border: "1px solid #444",
                    backgroundColor: "#2e2e2e",
                    color: "white",
                  }}
                >
                  {groups.map((g) => (
                    <option key={g.id} value={g.id}>
                      {g.title}
                    </option>
                  ))}
                </select>
              </div>
              <div style={{ marginBottom: "15px" }}>
                <label style={{ display: "block", marginBottom: "5px" }}>
                  Event Title
                </label>
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  style={{
                    width: "100%",
                    padding: "10px",
                    borderRadius: "6px",
                    border: "1px solid #444",
                    backgroundColor: "#2e2e2e",
                    color: "white",
                  }}
                  required
                />
              </div>
              <div style={{ marginBottom: "15px" }}>
                <label style={{ display: "block", marginBottom: "5px" }}>
                  Description
                </label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  style={{
                    width: "100%",
                    padding: "10px",
                    borderRadius: "6px",
                    border: "1px solid #444",
                    backgroundColor: "#2e2e2e",
                    color: "white",
                    height: "80px",
                  }}
                />
              </div>
              <div style={{ marginBottom: "20px" }}>
                <label style={{ display: "block", marginBottom: "5px" }}>
                  Event Date & Time
                </label>
                <input
                  type="datetime-local"
                  value={eventDate}
                  onChange={(e) => setEventDate(e.target.value)}
                  style={{
                    width: "100%",
                    padding: "10px",
                    borderRadius: "6px",
                    border: "1px solid #444",
                    backgroundColor: "#2e2e2e",
                    color: "white",
                  }}
                  required
                />
              </div>
              <div
                style={{
                  display: "flex",
                  justifyContent: "flex-end",
                  gap: "10px",
                }}
              >
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  style={{
                    padding: "8px 16px",
                    borderRadius: "6px",
                    border: "1px solid #555",
                    backgroundColor: "transparent",
                    color: "white",
                    cursor: "pointer",
                  }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={loading}
                  style={{
                    padding: "8px 16px",
                    borderRadius: "6px",
                    border: "none",
                    background: "linear-gradient(135deg, #11998e, #38ef7d)",
                    color: "white",
                    cursor: "pointer",
                  }}
                >
                  {loading ? "Creating..." : "Create Event"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: "20px" }}>
        {events.map((event) => (
          <div
            key={event.id}
            style={{
              backgroundColor: "#1e1e1e",
              borderRadius: "12px",
              padding: "20px",
              border: "1px solid #2e2e2e",
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <div>
              <span
                style={{
                  fontSize: "0.8rem",
                  color: "#11998e",
                  textTransform: "uppercase",
                  fontWeight: "bold",
                }}
              >
                {event.groupTitle}
              </span>
              <h3 style={{ margin: "5px 0 10px 0", color: "#fff" }}>
                {event.title}
              </h3>
              <p
                style={{
                  color: "#aaa",
                  fontSize: "0.95rem",
                  margin: "0 0 10px 0",
                }}
              >
                {event.description}
              </p>
              <small style={{ color: "#888" }}>
                📅 {new Date(event.event_date).toLocaleString()}
              </small>
              <div style={{ marginTop: "10px" }}>
                <span style={{ color: "#2ecc71", marginRight: "15px" }}>
                  ✔ {event.going_count} Going
                </span>
                <span style={{ color: "#e74c3c" }}>
                  ❌ {event.not_going_count} Not Going
                </span>
              </div>
            </div>

            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "10px",
                width: "120px",
              }}
            >
              <button
                onClick={() => handleRSVP(event.id, "going")}
                style={{
                  backgroundColor:
                    event.user_rsvp === "going" ? "#2ecc71" : "#2c3e50",
                  color: "white",
                  border: "none",
                  padding: "8px",
                  borderRadius: "6px",
                  cursor: "pointer",
                  fontWeight: "bold",
                }}
              >
                Going
              </button>
              <button
                onClick={() => handleRSVP(event.id, "not_going")}
                style={{
                  backgroundColor:
                    event.user_rsvp === "not_going" ? "#e74c3c" : "#2c3e50",
                  color: "white",
                  border: "none",
                  padding: "8px",
                  borderRadius: "6px",
                  cursor: "pointer",
                  fontWeight: "bold",
                }}
              >
                Not Going
              </button>
            </div>
          </div>
        ))}

        {events.length === 0 && (
          <div style={{ color: "#888", textAlign: "center", padding: "40px" }}>
            No upcoming events. Make sure you are in a group to view or create
            events!
          </div>
        )}
      </div>
    </div>
  );
};

export default Events;
