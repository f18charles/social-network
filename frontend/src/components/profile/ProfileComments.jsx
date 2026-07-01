import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import AuthorMeta from "../AuthorMeta.jsx";
import { apiFetch } from "../../utils/api";
import { logger } from "../../utils/logger.js";

const getListData = (response) =>
  Array.isArray(response) ? response : response?.data || [];

/**
 * ProfileComments renders comments and replies authored by the profile owner.
 */
const ProfileComments = ({ userId }) => {
  const [comments, setComments] = useState([]);
  const [status, setStatus] = useState("loading");
  const [error, setError] = useState("");
  const navigate = useNavigate();

  useEffect(() => {
    if (!userId) return;
    let isActive = true;

    async function loadComments() {
      setStatus("loading");
      setError("");
      try {
        const result = await apiFetch(`/api/users/${userId}/comments`);
        if (!isActive) return;
        setComments(getListData(result));
        setStatus("ready");
      } catch (err) {
        logger.error("Failed to load profile comments", err, { userId });
        if (!isActive) return;
        setError("Unable to load comments.");
        setStatus("error");
      }
    }

    loadComments();
    return () => {
      isActive = false;
    };
  }, [userId]);

  const openComment = (comment) => {
    if (!comment?.post_id || !comment?.id) return;
    navigate(`/post/${comment.post_id}?comment_id=${comment.id}`);
  };

  if (status === "loading") {
    return (
      <div className="profile-comments">
        <div className="profile-skeleton profile-skeleton--row" />
        <div className="profile-skeleton profile-skeleton--row" />
      </div>
    );
  }

  if (status === "error") {
    return <div className="profile-state profile-state--error">{error}</div>;
  }

  if (comments.length === 0) {
    return <div className="profile-state">No comments yet.</div>;
  }

  return (
    <div className="profile-comments">
      {comments.map((comment) => (
        <div
          role="button"
          tabIndex={0}
          key={comment.id}
          className="profile-comment-card"
          onClick={() => openComment(comment)}
          onKeyDown={(event) => {
            if (event.key === "Enter" || event.key === " ") {
              event.preventDefault();
              openComment(comment);
            }
          }}
        >
          <AuthorMeta
            author={comment.author}
            timestamp={new Date(comment.created_at).toLocaleString()}
            size="compact"
          />
          <span className="profile-comment-card__content">
            {comment.content}
          </span>
          <span className="profile-comment-card__meta">
            {comment.parent_comment_id ? "Reply" : "Comment"} -{" "}
            {comment.replies_count || 0} replies
          </span>
        </div>
      ))}
    </div>
  );
};

export default ProfileComments;
