import { useState } from "react";
import "../styles/comment.css";
import AuthorMeta from "./AuthorMeta.jsx";
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
 * Comment renders a folded comment or reply from the post comments API.
 */
const Comment = ({
  comment,
  depth = 1,
  isPostDeleted = false,
  onCreateReply,
  onVote,
  onLoadReplies,
  loadingReplies = {},
  renderComposer,
  focusedCommentId = null,
}) => {
  const [isReplying, setIsReplying] = useState(false);
  const replies = Array.isArray(comment?.replies) ? comment.replies : [];
  const repliesCount = comment?.replies_count || 0;
  const hasMoreReplies = repliesCount > replies.length;
  const isLoadingReplies = Boolean(loadingReplies[comment?.id]);

  const handleReplyCreated = async (formData) => {
    await onCreateReply?.(comment.id, formData);
    setIsReplying(false);
  };

  const loadMoreReplies = (event) => {
    event.stopPropagation();
    onLoadReplies?.(comment.id, replies.length);
  };

  if (comment?.deleted) {
    return (
      <div
        id={`comment-${comment?.id}`}
        className={`comment-thread${depth > 1 ? " comment-thread--nested" : ""}${focusedCommentId === comment?.id ? " is-focused" : ""}`}
      >
        <div className="comment-container comment-container--deleted">
          <AuthorMeta author={{ name: "Deleted user" }} size="compact" />
          <div className="comment-body">
            <div className="comment-details comment-details--deleted">
              <strong>Deleted user</strong>
              <p>This comment is no longer available.</p>
            </div>
          </div>
        </div>
        {hasMoreReplies ? (
          <button
            type="button"
            className="comment-load-replies"
            onClick={loadMoreReplies}
          >
            {isLoadingReplies
              ? "Loading replies..."
              : `View replies (${repliesCount - replies.length})`}
          </button>
        ) : null}
        {replies.map((reply) => (
          <Comment
            comment={reply}
            depth={depth + 1}
            key={reply.id}
            isPostDeleted={isPostDeleted}
            onCreateReply={onCreateReply}
            onVote={onVote}
            onLoadReplies={onLoadReplies}
            loadingReplies={loadingReplies}
            renderComposer={renderComposer}
            focusedCommentId={focusedCommentId}
          />
        ))}
      </div>
    );
  }

  return (
    <div
      id={`comment-${comment?.id}`}
      data-comment-id={comment?.id}
      className={`comment-thread${depth > 1 ? " comment-thread--nested" : ""}${focusedCommentId === comment?.id ? " is-focused" : ""}`}
    >
      <div className="comment-container">
        <AuthorMeta author={comment?.author} size="compact" />
        <div className="comment-body">
          <div className="comment-details">
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
            <span>{repliesCount} replies</span>
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
          {isReplying && renderComposer
            ? renderComposer(handleReplyCreated, "Write a reply")
            : null}
        </div>
      </div>
      {hasMoreReplies ? (
        <button
          type="button"
          className="comment-load-replies"
          onClick={loadMoreReplies}
        >
          {isLoadingReplies
            ? "Loading replies..."
            : `View replies (${repliesCount - replies.length})`}
        </button>
      ) : null}
      {replies.map((reply) => (
        <Comment
          comment={reply}
          depth={depth + 1}
          key={reply.id}
          isPostDeleted={isPostDeleted}
          onCreateReply={onCreateReply}
          onVote={onVote}
          onLoadReplies={onLoadReplies}
          loadingReplies={loadingReplies}
          renderComposer={renderComposer}
          focusedCommentId={focusedCommentId}
        />
      ))}
    </div>
  );
};

export default Comment;
