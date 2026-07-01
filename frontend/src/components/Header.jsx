import { useNavigate } from "react-router";
import { IoArrowRedo } from "react-icons/io5";
import "../styles/header.css";
import { FaPlusCircle } from "react-icons/fa";
import { AiOutlineMessage } from "react-icons/ai";
import { IoIosNotificationsOutline } from "react-icons/io";
import { MdOutlineGroup } from "react-icons/md";
import avatar from "../assets/user.svg";
import LogoutButton from "./LogoutButton";
import { useAuth } from "../context/auth/useAuth";

const Header = () => {
  const navigate = useNavigate();
  const { unreadNotifications } = useAuth();

  return (
    <div className="header">
      <div className="pretitle">
        <IoArrowRedo />
        <strong>CoreConnect</strong>
        <input
          type="text"
          className="search"
          placeholder="Search CoreConnect..."
        />
      </div>
      <div className="icons">
        <button
          type="button"
          className="icon-button"
          title="Explore groups"
          onClick={() => navigate("/groups")}
        >
          <FaPlusCircle size={24} />
        </button>
        <button
          type="button"
          className="icon-button"
          title="Messages"
          onClick={() => navigate("/messages")}
        >
          <AiOutlineMessage size={24} />
        </button>
        <button
          type="button"
          className="icon-button"
          title="Notifications"
          onClick={() => navigate("/notifications")}
          aria-label="View notifications"
        >
          <IoIosNotificationsOutline size={24} />
          {unreadNotifications > 0 && (
            <span className="header-badge">{unreadNotifications}</span>
          )}
        </button>
        <button
          type="button"
          className="icon-button"
          title="Browse groups"
          onClick={() => navigate("/groups")}
        >
          <MdOutlineGroup size={24} />
        </button>
        <img
          src={avatar}
          className="profile-photo"
          alt="Your profile"
          onClick={() => navigate("/profile")}
          title="Your profile"
        />
        <LogoutButton />
      </div>
    </div>
  );
};

export default Header;
