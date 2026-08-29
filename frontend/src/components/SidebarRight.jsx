import { useEffect, useState } from "react";
import { Link } from "react-router";
import UpcomingEvent from "./UpcomingEvent";
import { apiFetch } from "../utils/api";
import "../styles/sidebar-right.css";

const SidebarRight = () => {
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    let isMounted = true;

    const fetchUpcomingEvents = async () => {
      try {
        const groupList = await apiFetch("/api/groups");
        const acceptedGroups = (groupList || []).filter(
          (group) => group.status === "accepted"
        );

        const fetchedEvents = [];
        for (const group of acceptedGroups) {
          try {
            const groupEvents = await apiFetch(
              `/api/groups/${group.id}/events`
            );
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

        const now = new Date();
        const sortedEvents = fetchedEvents
          .filter(
            (event) => event.event_date && new Date(event.event_date) >= now
          )
          .sort((a, b) => new Date(a.event_date) - new Date(b.event_date));

        if (isMounted) {
          setEvents(sortedEvents.slice(0, 3));
          setError(null);
        }
      } catch (err) {
        console.error("Unable to load upcoming events:", err);
        if (isMounted) {
          setError(err?.message || "Unable to load upcoming events");
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    };

    fetchUpcomingEvents();

    const handleEventsUpdated = () => {
      setLoading(true);
      fetchUpcomingEvents();
    };

    window.addEventListener("eventsUpdated", handleEventsUpdated);
    return () => {
      isMounted = false;
      window.removeEventListener("eventsUpdated", handleEventsUpdated);
    };
  }, []);

  return (
    <aside className="quick-links">
      <div className="upcoming-events card">
        <div className="upcoming-events-header">
          <strong>Upcoming Events</strong>
          <Link to="/events" className="upcoming-events-view-all">
            View all
          </Link>
        </div>
        {loading && (
          <p className="upcoming-events-status">Loading events...</p>
        )}
        {error && (
          <p className="upcoming-events-error">{error}</p>
        )}
        {!loading && !error && events.length === 0 && (
          <p className="upcoming-events-status">
            No upcoming events yet.
          </p>
        )}
        {!loading &&
          !error &&
          events.map((event) => <UpcomingEvent key={event.id} event={event} />)}
      </div>
    </aside>
  );
};

export default SidebarRight;
