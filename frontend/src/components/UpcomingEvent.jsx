import "../styles/upcoming-events.css";

const UpcomingEvent = ({ event }) => {
  const eventDate = event.event_date ? new Date(event.event_date) : null;
  const month = eventDate
    ? eventDate.toLocaleString("default", { month: "short" }).toUpperCase()
    : "";
  const date = eventDate ? eventDate.getDate() : "";
  const time = eventDate
    ? eventDate.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
    : "";
  const subtitle = event.groupTitle || event.description || "Upcoming event";

  return (
    <div className="event card">
      <div className="date">
        <h4>{month}</h4>
        <strong>{date}</strong>
      </div>
      <div className="details">
        <strong>{event.title}</strong>
        <p>
          {time} - {subtitle}
        </p>
      </div>
    </div>
  );
};

export default UpcomingEvent;
