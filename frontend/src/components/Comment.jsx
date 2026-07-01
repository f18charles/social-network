import "../styles/comment.css";
import avatar from "../assets/user.svg";
import { Dislike, Like } from "./Reactions";

const Comment = ({
  comment,
  depth = 1,
  onReplyClick,
  getReplyInput,
  getIsReplyOpen,
  onReplyInputChange,
  onSubmitReply,
  onVote,
  onToggleReplies,
  getShowReplies,
}) => {
  const authorName = comment?.author
    ? (comment.author.nickname || `${comment.author.first_name || ""} ${comment.author.last_name || ""}`.trim())
    : comment?.name;

  const canReply = depth < 4;
  const isReplyOpen = getIsReplyOpen?.(comment.id);
  const replyInput = getReplyInput?.(comment.id) || "";

  return (
    <div
      className={`comment-thread${depth > 1 ? " comment-thread--nested" : ""}`}
      style={{ marginLeft: depth > 1 ? `${(depth - 1) * 20}px` : 0 }}
    >
      <div id="comment-container">
        <img
          src={comment?.author?.avatar ? comment.author.avatar : avatar}
          alt="avatar"
          className="profile-photo"
        />
        <div className="comment-body">
          <div className="comment-details">
            <strong>{authorName || "Anonymous"}</strong>
            <p>{comment?.content}</p>
          </div>
          <div className="comment-footer">
            {comment?.time}
            <div className="comment-reaction">
              <span className="comment-like">
                <Like
                  like={() => onVote?.(comment.id, "like")}
                  isActive={comment?.viewer_vote === "like"}
                  size={18}
                  className="comment-reaction-button"
                />
                <span style={{ marginLeft: "6px", fontSize: "0.9rem" }}>{comment?.like_count || 0}</span>
              </span>
              <span className="comment-like">
                <Dislike
                  dislike={() => onVote?.(comment.id, "dislike")}
                  isActive={comment?.viewer_vote === "dislike"}
                  size={18}
                  className="comment-reaction-button"
                />
                <span style={{ marginLeft: "6px", fontSize: "0.9rem" }}>{comment?.dislike_count || 0}</span>
              </span>
              {canReply && (
                <span
                  className="comment-reply"
                  onClick={() => onReplyClick?.(comment.id)}
                  style={{ cursor: "pointer" }}
                >
                  Reply
                </span>
              )}
              {comment?.replies?.length > 0 && (
                <span
                  className="comment-reply"
                  onClick={() => onToggleReplies?.(comment.id)}
                  style={{ cursor: "pointer" }}
                >
                  {getShowReplies?.(comment.id) ? `Hide replies (${comment.replies.length})` : `View replies (${comment.replies.length})`}
                </span>
              )}
            </div>
          </div>
        </div>
      </div>

      {canReply && isReplyOpen && (
        <form
          onSubmit={(event) => onSubmitReply?.(event, comment.id)}
          style={{ marginLeft: "calc(20px + 36px)", marginTop: "0.75rem" }}
        >
          <textarea
            value={replyInput}
            onChange={(e) => onReplyInputChange?.(comment.id, e.target.value)}
            placeholder="Write a reply..."
            rows={2}
            style={{
              width: "100%",
              padding: "10px",
              borderRadius: "10px",
              border: "1px solid #444",
              backgroundColor: "#1f1f1f",
              color: "#fff",
              resize: "vertical",
            }}
          />
          <button
            type="submit"
            disabled={!replyInput?.trim()}
            style={{
              marginTop: "8px",
              padding: "8px 16px",
              borderRadius: "10px",
              border: "none",
              background: "linear-gradient(135deg, #11998e, #38ef7d)",
              color: "white",
              cursor: replyInput?.trim() ? "pointer" : "not-allowed",
            }}
          >
            Post Reply
          </button>
        </form>
      )}

      {comment?.replies?.length > 0 && getShowReplies?.(comment.id) &&
        comment.replies.map((reply) => (
          <Comment
            key={reply.id}
            comment={reply}
            depth={reply.depth}
            onReplyClick={onReplyClick}
            getReplyInput={getReplyInput}
            getIsReplyOpen={getIsReplyOpen}
            onReplyInputChange={onReplyInputChange}
            onSubmitReply={onSubmitReply}
            onVote={onVote}
            onToggleReplies={onToggleReplies}
            getShowReplies={getShowReplies}
          />
        ))}
    </div>
  );
};

export default Comment;
