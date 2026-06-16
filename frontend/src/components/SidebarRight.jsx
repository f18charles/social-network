import React from "react";
import UpcomingEvent from "./UpcomingEvent";
import "../styles/sidebar-right.css"

const SidebarRight = () => {
  const events = [
    {
      id: "evt_001",
      month: "JUL",
      date: "18",
      title: "Tech Innovators Summit 2026",
      time: "09:00 AM",
      location: "Nairobi Startup Hub",
    },
    {
      id: "evt_002",
      month: "AUG",
      date: "05",
      title: "React & Go Developer Meetup",
      time: "06:30 PM",
      location: "Mombasa Tech Space",
    },
    {
      id: "evt_003",
      month: "SEP",
      date: "12",
      title: "Open Source Hackathon",
      time: "08:00 AM",
      location: "Kisumu Innovation Center",
    },
  ];

  return (
    <div className="quick-links ">
      <div classname="upcoming-events card">
        <strong>Upcoming Events</strong>
        {events.map((it) => (
          <UpcomingEvent event={it} />
        ))}
      </div>
      <div classname="contacts card">
        <strong>Contacts</strong>
      </div>
    </div>
  );
};

export default SidebarRight;
