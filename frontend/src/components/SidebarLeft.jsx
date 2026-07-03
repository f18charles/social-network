import { MdOutlineEvent, MdOutlineGroup, MdChevronLeft, MdChevronRight } from "react-icons/md";
import { AiOutlineMessage } from "react-icons/ai";
import { IoIosNotificationsOutline } from "react-icons/io";
import { BiGroup, BiHome, BiUser } from "react-icons/bi";
import { useNavigate } from "react-router";
import { useAuth } from "../context/auth/useAuth";

const SidebarLeft = ({ isCollapsed, onToggleCollapse }) => {
  const navigate = useNavigate();
  const { unreadNotifications } = useAuth();

  return (
    <aside className={`sidebar ${isCollapsed ? "sidebar--collapsed" : ""}`}>
      <button
        type="button"
        className="sidebar-toggle-btn"
        onClick={onToggleCollapse}
        title={isCollapsed ? "Expand sidebar" : "Collapse sidebar"}
      >
        {isCollapsed ? <MdChevronRight size={20} /> : <MdChevronLeft size={20} />}
      </button>
      <ul>
        <li className="links" onClick={() => navigate("/")} title="Home">
          <BiHome />
          {!isCollapsed && <span>Home</span>}
        </li>
        <li className="links" onClick={() => navigate("/profile")} title="Profile">
          <BiUser />
          {!isCollapsed && <span>Profile</span>}
        </li>
        <li className="links" onClick={() => navigate("/friends")} title="Friends">
          <BiGroup />
          {!isCollapsed && <span>Friends</span>}
        </li>
        <li className="links" onClick={() => navigate("/groups")} title="Groups">
          <MdOutlineGroup />
          {!isCollapsed && <span>Groups</span>}
        </li>
        <li className="links" onClick={() => navigate("/messages")} title="Messages">
          <AiOutlineMessage />
          {!isCollapsed && <span>Messages</span>}
        </li>
        <li className="links" onClick={() => navigate("/notifications")} title="Notifications">
          <IoIosNotificationsOutline />
          {!isCollapsed && <span>Notifications</span>}
          {unreadNotifications > 0 && (
            <span className="sidebar-badge">{unreadNotifications}</span>
          )}
        </li>
        <li className="links" onClick={() => navigate("/events")} title="Events">
          <MdOutlineEvent />
          {!isCollapsed && <span>Events</span>}
        </li>
      </ul>
    </aside>
  );
};

export default SidebarLeft;
