import { useNavigate } from "react-router";
import avatar from "../../assets/user.svg";
import "../../styles/usercard.css";

const toDisplayName = (user) =>
  `${user.first_name || ""} ${user.last_name || ""}`.trim() ||
  user.nickname ||
  "user";

const UserCard = ({ user, actions, onClick }) => {
  const navigate = useNavigate();

  const handleClick = () => {
    if (onClick) {
      onClick(user);
    } else {
      navigate(`/user/${user.id}`);
    }
  };

  return (
    <div className="user-card">
      <img
        src={user.avatar || avatar}
        alt={`${toDisplayName(user)}'s avatar`}
        onClick={handleClick}
        className="user-card__avatar"
      />
      <div onClick={handleClick} className="user-card__info">
        <div className="user-card__name">{toDisplayName(user)}</div>
        {user.nickname && (
          <div className="user-card__nickname">@{user.nickname}</div>
        )}
      </div>
      {actions && <div className="user-card__actions">{actions}</div>}
    </div>
  );
};

export default UserCard;
