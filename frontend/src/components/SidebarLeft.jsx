import { MdOutlineEvent, MdOutlineGroup } from "react-icons/md";
import { AiOutlineMessage } from "react-icons/ai";
import { IoIosNotificationsOutline } from "react-icons/io";
import { BiGroup, BiHome, BiUser } from "react-icons/bi";
import { useNavigate } from "react-router";
import { useAuth } from "../context/auth/useAuth";

const SidebarLeft = () => {
  const navigate = useNavigate();
  const { unreadNotifications } = useAuth();

  return (
    <aside className="sidebar">
      <ul>
        <li className="links" onClick={() => navigate("/")}>
          <BiHome /> Home
        </li>
        <li className="links" onClick={() => navigate("/profile")}>
          <BiUser /> Profile
        </li>
        <li className="links" onClick={() => navigate("/friends")}>
          <BiGroup />
          Friends
        </li>
        <li className="links" onClick={() => navigate("/groups")}>
          <MdOutlineGroup />
          Groups
        </li>
        <li className="links" onClick={() => navigate("/messages")}>
          <AiOutlineMessage />
          Messages
        </li>
        <li className="links" onClick={() => navigate("/notifications")}>
          <IoIosNotificationsOutline />
          Notifications
          {unreadNotifications > 0 && (
            <span className="sidebar-badge">{unreadNotifications}</span>
          )}
        </li>
        <li className="links" onClick={() => navigate("/events")}>
          <MdOutlineEvent />
          Events
        </li>
      </ul>
    </aside>
  );
};

export default SidebarLeft;
