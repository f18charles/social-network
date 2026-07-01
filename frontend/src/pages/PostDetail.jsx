import { useEffect, useState } from "react";
import { useLocation, useParams } from "react-router";
import Post from "../components/Post.jsx";
import Comment from "../components/Comment.jsx";
import { useAuth } from "../context/useAuth.js";
import { apiFetch } from "../utils/api";
import "../styles/post-detail.css";

const PostDetail = () => {
  const location = useLocation();
  const { id } = useParams();
  const { isAuthenticated } = useAuth();
  const [post, setPost] = useState(location.state || null);
  const [comments, setComments] = useState([]);
  const [commentInput, setCommentInput] = useState("");
  const [replyInputs, setReplyInputs] = useState({});
  const [openReplyTargets, setOpenReplyTargets] = useState({});
  const [visibleReplies, setVisibleReplies] = useState({});
  const [loadingComments, setLoadingComments] = useState(true);
  const [commentError, setCommentError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const DateFormatter = (datestring, now) => {
    const date = new Date(datestring);
    const diffInMs = now - date.getTime();

    const ONE_MINUTE = 60000;
    const ONE_HOUR = 3600000;
    const ONE_DAY = 86400000;
    const ONE_MONTH = 2592000000;
    const ONE_YEAR = 31536000000;

    if (diffInMs < 0) {
      return "In the future";
    }

    switch (true) {
      case diffInMs < ONE_HOUR:
        return `${Math.floor(diffInMs / ONE_MINUTE)} minutes ago`;
      case diffInMs < ONE_DAY:
        return `${Math.floor(diffInMs / ONE_HOUR)} hours ago`;
      case diffInMs < ONE_MONTH:
        return `${Math.floor(diffInMs / ONE_DAY)} days ago`;
      case diffInMs < ONE_YEAR:
        return `${Math.floor(diffInMs / ONE_MONTH)} months ago`;
      case diffInMs > ONE_YEAR:
        return `${Math.floor(diffInMs / ONE_YEAR)} years ago`;
      default:
        const yearsAgo = Math.floor(diffInMs / ONE_YEAR);
        return `${yearsAgo} ${yearsAgo === 1 ? "year" : "years"} ago`;
    }
  };

  const normalizeComments = (commentList, depth = 1) =>
    (commentList || []).map((comment) => ({
      ...comment,
      depth,
      time: DateFormatter(comment.created_at || comment.createdAt, Date.now()),
      replies: normalizeComments(comment.replies || [], Math.min(depth + 1, 4)),
    }));

  const handleReplyClick = (commentId) => {
    setOpenReplyTargets((prev) => ({
      ...prev,
      [commentId]: !prev[commentId],
    }));
  };

  const handleToggleReplies = (commentId) => {
    setVisibleReplies((prev) => ({
      ...prev,
      [commentId]: !prev[commentId],
    }));
  };

  const handleReplyInputChange = (commentId, value) => {
    setReplyInputs((prev) => ({
      ...prev,
      [commentId]: value,
    }));
  };

  const findCommentByID = (commentList, targetId) => {
    for (const comment of commentList) {
      if (comment.id === targetId) {
        return comment;
      }
      if (comment.replies?.length > 0) {
        const found = findCommentByID(comment.replies, targetId);
        if (found) {
          return found;
        }
      }
    }
    return null;
  };

  const getReplyInput = (commentId) => replyInputs[commentId] || "";
  const getIsReplyOpen = (commentId) => Boolean(openReplyTargets[commentId]);
  const getShowReplies = (commentId) => Boolean(visibleReplies[commentId]);

  const handleSubmitReply = async (event, parentId) => {
    event.preventDefault();
    if (!parentId || !id) {
      return;
    }

    const content = (replyInputs[parentId] || "").trim();
    if (!content) {
      return;
    }

    setSubmitting(true);
    setCommentError("");

    try {
      const formData = new FormData();
      formData.append("content", content);
      formData.append("parent_comment_id", parentId);

      await apiFetch(`/api/posts/${id}/comments`, {
        method: "POST",
        body: formData,
      });

      setReplyInputs((prev) => ({
        ...prev,
        [parentId]: "",
      }));
      setOpenReplyTargets((prev) => ({
        ...prev,
        [parentId]: false,
      }));
      await fetchComments();

      if (post) {
        setPost((prev) => ({
          ...prev,
          comment_count: (prev?.comment_count || 0) + 1,
        }));
      }
    } catch (err) {
      console.error("Failed to submit reply:", err);
      setCommentError(err.message || "Failed to submit reply");
    } finally {
      setSubmitting(false);
    }
  };

  const fetchPost = async () => {
    if (post || !id) {
      return;
    }

    try {
      const data = await apiFetch(`/api/posts/${id}`);
      setPost(data);
    } catch (err) {
      console.error("Failed to load post:", err);
    }
  };

  const handleVoteComment = async (commentId, vote) => {
    if (!commentId || !id) {
      return;
    }

    const comment = findCommentByID(comments, commentId);
    const sameVote = comment?.viewer_vote === vote;
    const method = sameVote ? "DELETE" : "POST";

    try {
      await apiFetch(`/api/comments/${commentId}/vote`, {
        method,
        body: method === "POST" ? { vote } : undefined,
      });
      await fetchComments();
    } catch (err) {
      console.error("Failed to vote comment:", err);
      setCommentError(err.message || "Unable to update comment vote");
    }
  };

  const fetchComments = async () => {
    if (!isAuthenticated || !id) {
      setComments([]);
      setLoadingComments(false);
      return;
    }

    setLoadingComments(true);
    setCommentError("");

    try {
      const data = await apiFetch(`/api/posts/${id}/comments`);
      setComments(normalizeComments(data || []));
    } catch (err) {
      console.error("Failed to load comments:", err);
      setCommentError(err.message || "Unable to load comments");
    } finally {
      setLoadingComments(false);
    }
  };

  useEffect(() => {
    fetchPost();
  }, [id]);

  useEffect(() => {
    fetchComments();
  }, [id, isAuthenticated]);

  const handleSubmitComment = async (event) => {
    event.preventDefault();
    if (!commentInput.trim() || !id) {
      return;
    }

    setSubmitting(true);
    setCommentError("");

    try {
      const formData = new FormData();
      formData.append("content", commentInput.trim());

      const newComment = await apiFetch(`/api/posts/${id}/comments`, {
        method: "POST",
        body: formData,
      });

      setCommentInput("");
      setComments((prev) => [
        {
          ...newComment,
          time: DateFormatter(newComment.created_at || newComment.createdAt, Date.now()),
        },
        ...prev,
      ]);

      if (post) {
        setPost((prev) => ({
          ...prev,
          comment_count: (prev?.comment_count || 0) + 1,
        }));
      }
    } catch (err) {
      console.error("Failed to submit comment:", err);
      setCommentError(err.message || "Failed to submit comment");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="post-detail">
      <Post post={post} />
      <div className="comments card">
        <h3>Comments</h3>

        {isAuthenticated ? (
          <form onSubmit={handleSubmitComment} style={{ marginBottom: "1rem" }}>
            <textarea
              value={commentInput}
              onChange={(e) => setCommentInput(e.target.value)}
              placeholder="Write a comment..."
              rows={3}
              style={{
                width: "100%",
                padding: "12px",
                borderRadius: "10px",
                border: "1px solid #444",
                backgroundColor: "#1f1f1f",
                color: "#fff",
                resize: "vertical",
              }}
            />
            <button
              type="submit"
              disabled={submitting || !commentInput.trim()}
              style={{
                marginTop: "10px",
                padding: "10px 18px",
                borderRadius: "10px",
                border: "none",
                background: "linear-gradient(135deg, #11998e, #38ef7d)",
                color: "white",
                cursor: submitting ? "not-allowed" : "pointer",
              }}
            >
              {submitting ? "Posting..." : "Post Comment"}
            </button>
          </form>
        ) : (
          <p style={{ color: "#888", marginBottom: "1rem" }}>
            Log in to view and post comments.
          </p>
        )}

        {commentError && (
          <p style={{ color: "#e74c3c" }}>{commentError}</p>
        )}

        {loadingComments ? (
          <p style={{ color: "#888" }}>Loading comments...</p>
        ) : comments.length === 0 ? (
          <p style={{ color: "#888" }}>No comments yet. Be the first to comment!</p>
        ) : (
          comments.map((comment) => (
            <Comment
              key={comment.id}
              comment={comment}
              depth={comment.depth}
              onReplyClick={handleReplyClick}
              getReplyInput={getReplyInput}
              getIsReplyOpen={getIsReplyOpen}
              onReplyInputChange={handleReplyInputChange}
              onSubmitReply={handleSubmitReply}
              onVote={handleVoteComment}
              onToggleReplies={handleToggleReplies}
              getShowReplies={getShowReplies}
            />
          ))
        )}
      </div>
    </div>
  );
};

export default PostDetail;
