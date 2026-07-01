import { useEffect, useState } from "react";
import UpcomingEvent from "./UpcomingEvent";
import { apiFetch } from "../utils/api";
import "../styles/sidebar-right.css";

const SidebarRight = () => {
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchUpcomingEvents = async () => {
      try {
        const groupList = await apiFetch("/api/groups");
        const acceptedGroups = (groupList || []).filter((group) => group.status === "accepted");

        const fetchedEvents = [];
        for (const group of acceptedGroups) {
          try {
            const groupEvents = await apiFetch(`/api/groups/${group.id}/events`);
            if (Array.isArray(groupEvents)) {
              fetchedEvents.push(
                ...groupEvents.map((event) => ({
                  ...event,
                  groupTitle: group.title,
                }))
              );
            }
          } catch (err) {
            console.error(`Failed to load events for group ${group.id}:`, err);
          }
        }

        const sortedEvents = fetchedEvents
          .filter((event) => event.event_date)
          .sort((a, b) => new Date(a.event_date) - new Date(b.event_date));

        setEvents(sortedEvents.slice(0, 3));
        setError(null);
      } catch (err) {
        console.error("Unable to load upcoming events:", err);
        setError(err?.message || "Unable to load upcoming events");
      } finally {
        setLoading(false);
      }
    };

    fetchUpcomingEvents();

    const handleEventsUpdated = () => {
      setLoading(true);
      fetchUpcomingEvents();
    };

    window.addEventListener("eventsUpdated", handleEventsUpdated);
    return () => {
      window.removeEventListener("eventsUpdated", handleEventsUpdated);
    };
  }, []);

  return (
    <div className="quick-links ">
      <div className="upcoming-events card">
        <strong>Upcoming Events</strong>
        {loading && <p style={{ marginTop: "1rem", color: "#888" }}>Loading events...</p>}
        {error && <p style={{ marginTop: "1rem", color: "#e74c3c" }}>{error}</p>}
        {!loading && !error && events.length === 0 && (
          <p style={{ marginTop: "1rem", color: "#888" }}>No upcoming events yet.</p>
        )}
        {!loading && !error && events.map((event) => (
          <UpcomingEvent key={event.id} event={event} />
        ))}
      </div>
    </div>
  );
};

export default SidebarRight;
