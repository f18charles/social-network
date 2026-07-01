import { useEffect, useState } from "react";
import { useLocation, useParams } from "react-router";
import Post from "../components/Post.jsx";
import "../styles/post-detail.css";
import Comment from "../components/Comment.jsx";
import { apiFetch } from "../utils/api.js";

const incrementReplyCount = (comment) =>
  comment?.deleted ? comment : { ...comment, replies_count: (comment?.replies_count || 0) + 1 };

const appendReply = (comments, parentID, reply) =>
  comments.map((comment) => {
    if (comment.id === parentID) {
      return incrementReplyCount({
        ...comment,
        replies: [...(comment.replies || []), reply],
      });
    }

    if (Array.isArray(comment.replies) && comment.replies.length > 0) {
      const nextReplies = appendReply(comment.replies, parentID, reply);
      if (nextReplies !== comment.replies) {
        const changed = nextReplies.some((item, index) => item !== comment.replies[index]);
        return changed ? incrementReplyCount({ ...comment, replies: nextReplies }) : comment;
      }
    }

    return comment;
  });

const updateCommentVote = (comments, commentID, summary) =>
  comments.map((comment) => {
    if (comment.id === commentID) {
      return {
        ...comment,
        like_count: summary?.like_count ?? comment.like_count ?? 0,
        dislike_count: summary?.dislike_count ?? comment.dislike_count ?? 0,
        viewer_vote: summary?.viewer_vote || "none",
      };
    }

    if (Array.isArray(comment.replies) && comment.replies.length > 0) {
      return { ...comment, replies: updateCommentVote(comment.replies, commentID, summary) };
    }

    return comment;
  });

/**
 * CommentComposer creates a multipart comment or reply submission payload.
 */
const CommentComposer = ({ onSubmit, placeholder = "Write a comment" }) => {
  const [content, setContent] = useState("");
  const [image, setImage] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState(null);

  const handleSubmit = async (event) => {
    event.preventDefault();
    const trimmed = content.trim();
    if (!trimmed && !image) {
      setError("Write a comment or choose an image.");
      return;
    }

    const formData = new FormData();
    if (trimmed) formData.append("content", trimmed);
    if (image) formData.append("image", image);

    setIsSubmitting(true);
    setError(null);
    try {
      await onSubmit(formData);
      setContent("");
      setImage(null);
      event.currentTarget.reset();
    } catch (err) {
      setError(err?.message || "Unable to create comment.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form className="comment-composer" onSubmit={handleSubmit}>
      <textarea
        value={content}
        placeholder={placeholder}
        onChange={(event) => setContent(event.target.value)}
        rows={3}
      />
      <div className="comment-composer-actions">
        <input
          type="file"
          accept="image/jpeg,image/png,image/gif"
          onChange={(event) => setImage(event.target.files?.[0] || null)}
        />
        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Posting..." : "Post"}
        </button>
      </div>
      {error ? <div className="error">{error}</div> : null}
    </form>
  );
};

/**
 * PostDetail renders a single post and its database-backed comment thread.
 */
const PostDetail = () => {
  const { id } = useParams();
  const location = useLocation();
  const [post, setPost] = useState(location.state || null);
  const [comments, setComments] = useState([]);
  const [loading, setLoading] = useState(!location.state);
  const [error, setError] = useState(null);

  useEffect(() => {
    let active = true;

    async function loadPostDetail() {
      setLoading(true);
      setError(null);
      try {
        const [postData, commentData] = await Promise.all([
          apiFetch(`/api/posts/${id}`),
          apiFetch(`/api/posts/${id}/comments`),
        ]);
        if (!active) return;
        setPost(postData);
        setComments(Array.isArray(commentData) ? commentData : []);
      } catch (err) {
        if (!active) return;
        setError(err?.message || "Unable to load post.");
      } finally {
        if (active) setLoading(false);
      }
    }

    if (id) {
      loadPostDetail();
    }

    return () => {
      active = false;
    };
  }, [id]);

  const createComment = async (formData, parentCommentID = null) => {
    if (parentCommentID) formData.append("parent_comment_id", parentCommentID);
    const created = await apiFetch(`/api/posts/${id}/comments`, {
      method: "POST",
      body: formData,
    });

    if (parentCommentID) {
      setComments((current) => appendReply(current, parentCommentID, created));
    } else {
      setComments((current) => [...current, created]);
      setPost((current) =>
        current ? { ...current, comment_count: (current.comment_count || 0) + 1 } : current
      );
    }
  };

  const voteComment = async (commentID, vote) => {
    const target = findComment(comments, commentID);
    const summary =
      target?.viewer_vote === vote
        ? await apiFetch(`/api/comments/${commentID}/vote`, { method: "DELETE" })
        : await apiFetch(`/api/comments/${commentID}/vote`, {
            method: "PUT",
            body: { vote },
          });
    setComments((current) => updateCommentVote(current, commentID, summary));
  };

  const renderComposer = (onSubmit, placeholder) => (
    <CommentComposer onSubmit={onSubmit} placeholder={placeholder} />
  );

  return (
    <div className="post-detail">
      {loading && <div>Loading post...</div>}
      {error && <div className="error">{error}</div>}
      {post && <Post post={post} onPostChange={setPost} />}
      <div className="comments card">
        <h3>Comments</h3>
        {post && !post.deleted ? (
          <CommentComposer onSubmit={(formData) => createComment(formData)} />
        ) : null}
        {comments.length === 0 && !loading ? <p>No comments yet.</p> : null}
        {comments.map((comment) => (
          <Comment
            comment={comment}
            postId={id}
            key={comment.id}
            isPostDeleted={post?.deleted}
            onCreateReply={(parentID, formData) => createComment(formData, parentID)}
            onVote={voteComment}
            renderComposer={renderComposer}
          />
        ))}
      </div>
    </div>
  );
};

const findComment = (comments, commentID) => {
  for (const comment of comments) {
    if (comment.id === commentID) return comment;
    const found = Array.isArray(comment.replies) ? findComment(comment.replies, commentID) : null;
    if (found) return found;
  }
  return null;
};

export default PostDetail;
