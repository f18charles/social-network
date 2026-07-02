import { useState } from "react";
import "../styles/comment.css";
import { MdEdit } from "react-icons/md";
import AuthorMeta from "./AuthorMeta.jsx";
import CommentComposer from "./CommentComposer.jsx";
import { VoteControls } from "./VoteControls";
import { useAuth } from "../context/auth/useAuth.js";
import { apiFetch } from "../utils/api.js";
import { logger } from "../utils/logger.js";

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
  onUpdateComment,
  onVote,
  onLoadReplies,
  loadingReplies = {},
  focusedCommentId = null,
}) => {
  const [isReplying, setIsReplying] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editError, setEditError] = useState(null);
  const { currentUser } = useAuth();
  const replies = Array.isArray(comment?.replies) ? comment.replies : [];
  const repliesCount = comment?.replies_count || 0;
  const hasMoreReplies = repliesCount > replies.length;
  const isLoadingReplies = Boolean(loadingReplies[comment?.id]);
  const isOwner =
    Boolean(currentUser?.id) && String(currentUser.id) === String(comment?.author?.id);
  const canInteract = !isPostDeleted && !comment?.deleted;

  const handleReplyCreated = async (formData) => {
    await onCreateReply?.(comment.id, formData);
    setIsReplying(false);
  };

  const handleEdit = async (formData) => {
    setEditError(null);
    try {
      const updated = await apiFetch(`/api/comments/${comment.id}`, {
        method: "PATCH",
        body: formData,
      });
      onUpdateComment?.(updated);
      setIsEditing(false);
    } catch (err) {
      logger.error("Failed to update comment", err, { commentId: comment.id });
      setEditError(err.message || "Unable to update comment.");
      throw err;
    }
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
            onUpdateComment={onUpdateComment}
            onVote={onVote}
            onLoadReplies={onLoadReplies}
            loadingReplies={loadingReplies}
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
          {isEditing ? (
            <CommentComposer
              initialComment={comment}
              onSubmit={handleEdit}
              onCancel={() => setIsEditing(false)}
              placeholder="Edit comment"
              submitLabel="Save"
            />
          ) : (
            <div className="comment-details">
              <p>{comment?.content}</p>
              {comment?.updated_at ? <small>Edited</small> : null}
              {comment?.image_url ? (
                <img
                  className="comment-image"
                  src={comment.image_url}
                  alt="comment attachment"
                />
              ) : null}
            </div>
          )}
          <div className="comment-footer">
            <span>{formatCommentTime(comment)}</span>
            <span>{repliesCount} replies</span>
            {canInteract ? (
              <button
                type="button"
                className="comment-reply"
                onClick={() => setIsReplying((value) => !value)}
              >
                Reply
              </button>
            ) : null}
            {canInteract && isOwner ? (
              <button
                type="button"
                className="comment-edit"
                aria-label="Edit comment"
                onClick={() => setIsEditing(true)}
              >
                <MdEdit aria-hidden="true" />
              </button>
            ) : null}
          </div>
          {canInteract ? (
            <VoteControls
              likes={comment?.like_count || 0}
              dislikes={comment?.dislike_count || 0}
              currentVote={comment?.viewer_vote || "none"}
              targetType="comment"
              onVote={(vote) => onVote?.(comment.id, vote)}
            />
          ) : null}
          {isReplying ? (
            <CommentComposer
              onSubmit={handleReplyCreated}
              placeholder="Write a reply"
              submitLabel="Reply"
            />
          ) : null}
          {editError ? <div className="error">{editError}</div> : null}
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
          onUpdateComment={onUpdateComment}
          onVote={onVote}
          onLoadReplies={onLoadReplies}
          loadingReplies={loadingReplies}
          focusedCommentId={focusedCommentId}
        />
      ))}
    </div>
  );
};

export default Comment;
