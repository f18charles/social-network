import { useEffect, useState } from "react";
import { useLocation, useParams } from "react-router";
import Post from "../components/Post.jsx";
import "../styles/post-detail.css";
import Comment from "../components/Comment.jsx";
import { apiFetch } from "../utils/api.js";

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

  return (
    <div className="post-detail">
      {loading && <div>Loading post...</div>}
      {error && <div className="error">{error}</div>}
      {post && <Post post={post} />}
      <div className="comments card">
        <h3>Comments</h3>
        {comments.length === 0 && !loading ? <p>No comments yet.</p> : null}
        {comments.map((comment) => (
          <Comment comment={comment} key={comment.id} />
        ))}
      </div>
    </div>
  );
};

export default PostDetail;
