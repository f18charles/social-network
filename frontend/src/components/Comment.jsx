import "../styles/comment.css";
import avatar from "../assets/user.svg";
import { Like } from "./Reactions";

const formatCommentTime = (comment) => {
  if (comment?.time) return comment.time;
  if (!comment?.created_at) return "";

  const created = new Date(comment.created_at);
  if (Number.isNaN(created.getTime())) return "";

  return created.toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

/**
 * Comment renders a comment or reply from the post comments API tree.
 */
const Comment = ({ comment }) => {
  const authorName = comment?.author
    ? comment.author.nickname ||
      `${comment.author.first_name || ""} ${comment.author.last_name || ""}`.trim()
    : comment?.name;
  const replies = Array.isArray(comment?.replies) ? comment.replies : [];

  if (comment?.deleted) {
    return (
      <div className="comment-thread">
        <div className="comment-container">
          <div className="comment-body">
            <div className="comment-details">
              <strong>Deleted comment</strong>
              <p>This comment was deleted.</p>
            </div>
          </div>
        </div>
        {replies.map((reply) => (
          <Comment comment={reply} key={reply.id} />
        ))}
      </div>
    );
  }

  return (
    <div className="comment-thread">
      <div className="comment-container">
        <img
          src={comment?.author?.avatar ? comment.author.avatar : avatar}
          alt="avatar"
          className="profile-photo"
        />
        <div className="comment-body">
          <div className="comment-details">
            <strong>{authorName || "Anonymous"}</strong>
            <p>{comment?.content}</p>
            {comment?.image_url ? (
              <img
                className="comment-image"
                src={comment.image_url}
                alt="comment attachment"
              />
            ) : null}
          </div>
          <div className="comment-footer">
            {formatCommentTime(comment)}
            <div className="comment-reaction">
              <span className="comment-like">
                <Like />
              </span>{" "}
              <span className="comment-reply">Reply</span>
            </div>
          </div>
        </div>
      </div>
      {replies.map((reply) => (
        <Comment comment={reply} key={reply.id} />
      ))}
    </div>
  );
};

export default Comment;
