import { useNavigate } from "react-router";
import avatarFallback from "../assets/user.svg";

/**
 * getAuthorDisplayName returns the best available display name for public author DTOs.
 */
const getAuthorDisplayName = (author) =>
  author?.nickname ||
  `${author?.first_name || ""} ${author?.last_name || ""}`.trim() ||
  author?.name ||
  "Anonymous";

/**
 * AuthorMeta renders a clickable avatar/name pair with consistent profile navigation.
 */
const AuthorMeta = ({
  author,
  timestamp,
  size = "default",
  className = "",
}) => {
  const navigate = useNavigate();
  const authorName = getAuthorDisplayName(author);
  const openProfile = (event) => {
    event.stopPropagation();
    if (author?.id) {
      navigate(`/user/${author.id}`);
    }
  };
  const clickable = Boolean(author?.id);

  return (
    <div className={`author-meta author-meta--${size} ${className}`.trim()}>
      <button
        type="button"
        className="author-meta__avatar-button"
        onClick={openProfile}
        disabled={!clickable}
        aria-label={clickable ? `Open ${authorName}'s profile` : undefined}
      >
        <img
          src={author?.avatar || avatarFallback}
          alt={`${authorName} avatar`}
          className="profile-photo author-meta__avatar"
        />
      </button>
      <div className="author-meta__text">
        <button
          type="button"
          className="author-meta__name"
          onClick={openProfile}
          disabled={!clickable}
        >
          {authorName}
        </button>
        {timestamp ? <small>{timestamp}</small> : null}
      </div>
    </div>
  );
};

export default AuthorMeta;
