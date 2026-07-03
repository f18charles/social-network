import { useState } from "react";
import { useNavigate } from "react-router";
import { IoArrowRedo } from "react-icons/io5";
import "../styles/header.css";
import { FaPlusCircle } from "react-icons/fa";
import { AiOutlineMessage } from "react-icons/ai";
import { IoIosNotificationsOutline } from "react-icons/io";
import { MdOutlineGroup } from "react-icons/md";
import avatar from "../assets/user.svg";
import { useAuth } from "../context/auth/useAuth";

const Header = ({ onToggleMobileMenu }) => {
  const navigate = useNavigate();
  const { currentUser, unreadNotifications, logout } = useAuth();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  return (
    <div className="header">
      <div className="pretitle">
        <button
          type="button"
          className="mobile-menu-toggle"
          onClick={onToggleMobileMenu}
          aria-label="Toggle navigation menu"
        >
          ☰
        </button>
        <IoArrowRedo />
        <strong>CoreConnect</strong>
        <input
          type="text"
          className="search"
          placeholder="Search CoreConnect..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              const term = searchQuery.trim();
              if (term) {
                navigate(`/search?query=${encodeURIComponent(term)}`);
              }
            }
          }}
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
        <div className="profile-photo-container">
          <img
            src={currentUser?.avatar || avatar}
            className="profile-photo"
            alt="Your profile"
            onClick={() => setIsMenuOpen(!isMenuOpen)}
            title="Your profile"
          />
          {isMenuOpen && (
            <>
              <div className="avatar-dropdown-backdrop" onClick={() => setIsMenuOpen(false)} />
              <div className="avatar-dropdown">
                <button
                  type="button"
                  className="dropdown-item"
                  onClick={() => {
                    setIsMenuOpen(false);
                    navigate("/profile");
                  }}
                >
                  My Profile
                </button>
                <button
                  type="button"
                  className="dropdown-item"
                  onClick={() => {
                    setIsMenuOpen(false);
                    navigate("/profile/edit");
                  }}
                >
                  Edit Profile
                </button>
                <button
                  type="button"
                  className="dropdown-item"
                  onClick={async () => {
                    setIsMenuOpen(false);
                    await logout();
                    navigate("/login");
                  }}
                >
                  Sign Out
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default Header;
