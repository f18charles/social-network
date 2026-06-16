import React from "react";
import "../styles/upcoming-events.css"

const UpcomingEvent = ({ event }) => {
  return (
    <div className="event card">
      <div className="date">
        <h4>{event.month}</h4>
        <strong>{event.date}</strong>
      </div>
      <div className="details">
        <strong>{event.title}</strong>
        <p>{event?.time} - {event.location}</p>
      </div>
    </div>
  );
};

export default UpcomingEvent;
