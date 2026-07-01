import { useState } from "react";
import "../styles/comment.css";
import avatar from "../assets/user.svg";
import { VoteControls } from "./VoteControls";

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
const Comment = ({
  comment,
  postId,
  isPostDeleted = false,
  onCreateReply,
  onVote,
  renderComposer,
}) => {
  const [isReplying, setIsReplying] = useState(false);
  const authorName = comment?.author
    ? comment.author.nickname ||
      `${comment.author.first_name || ""} ${comment.author.last_name || ""}`.trim()
    : comment?.name;
  const replies = Array.isArray(comment?.replies) ? comment.replies : [];

  const handleReplyCreated = async (formData) => {
    await onCreateReply?.(comment.id, formData);
    setIsReplying(false);
  };

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
          <Comment
            comment={reply}
            key={reply.id}
            postId={postId}
            isPostDeleted={isPostDeleted}
            onCreateReply={onCreateReply}
            onVote={onVote}
            renderComposer={renderComposer}
          />
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
            <span>{formatCommentTime(comment)}</span>
            <span>{comment?.replies_count || 0} replies</span>
            {!isPostDeleted ? (
              <button
                type="button"
                className="comment-reply"
                onClick={() => setIsReplying((value) => !value)}
              >
                Reply
              </button>
            ) : null}
          </div>
          {!isPostDeleted ? (
            <VoteControls
              likes={comment?.like_count || 0}
              dislikes={comment?.dislike_count || 0}
              currentVote={comment?.viewer_vote || "none"}
              targetType="comment"
              onVote={(vote) => onVote?.(comment.id, vote)}
            />
          ) : null}
          {isReplying && renderComposer ? renderComposer(handleReplyCreated, "Write a reply") : null}
        </div>
      </div>
      {replies.map((reply) => (
        <Comment
          comment={reply}
          key={reply.id}
          postId={postId}
          isPostDeleted={isPostDeleted}
          onCreateReply={onCreateReply}
          onVote={onVote}
          renderComposer={renderComposer}
        />
      ))}
    </div>
  );
};

export default Comment;
