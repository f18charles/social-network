import { useState, useEffect } from "react";
import Post from "../components/Post.jsx";
import "../styles/home.css";
import NewPost from "../components/NewPost.jsx";
import { apiFetch } from "../utils/api.js";
import { logger } from "../utils/logger.js";

/**
 * mapPostPayload adapts a backend post DTO to the Post component shape.
 */
function mapPostPayload(payload) {
  if (payload?.deleted) return { id: payload.id, deleted: true };

  return {
    id: payload.id,
    author: payload.author
      ? {
          id: payload.author.id,
          name:
            payload.author.nickname ||
            `${payload.author.first_name || ""} ${payload.author.last_name || ""}`.trim() ||
            payload.author.email ||
            undefined,
          first_name: payload.author.first_name,
          last_name: payload.author.last_name,
          nickname: payload.author.nickname,
          avatar: payload.author.avatar,
        }
      : null,
    group_id: payload.group_id || null,
    content: payload.content,
    image_url: payload.image_url || null,
    privacy: payload.privacy || "public",
    like_count: payload.like_count || 0,
    dislike_count: payload.dislike_count || 0,
    comment_count: payload.comment_count || 0,
    viewer_vote: payload.viewer_vote || "none",
    created_at: payload.created_at || new Date().toISOString(),
    updated_at: payload.updated_at || null,
  };
}

function Home() {
  const [Allposts, setAllposts] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    let active = true;

    async function loadPosts() {
      setLoading(true);
      setError(null);
      try {
        const response = await apiFetch("/api/posts");
        const data = Array.isArray(response) ? response : response?.data || [];
        const mapped = Array.isArray(data) ? data.map(mapPostPayload) : [];
        if (active) setAllposts(mapped);
      } catch (err) {
        logger.error("Failed to load home feed", err);
        if (active) setError("Failed to load posts");
      } finally {
        if (active) setLoading(false);
      }
    }

    loadPosts();

    return () => {
      active = false;
    };
  }, []);

  function handlePostChange(updated) {
    if (!updated?.id) return;
    setAllposts((prev) =>
      prev.map((post) => (post.id === updated.id ? mapPostPayload(updated) : post))
    );
  }

  function handleNewPost(created) {
    if (!created) return;

    // Backend may return an envelope or the raw post object. Normalize to the post payload.
    const payload = created.data ? created.data : created;
    if (!payload) return;

    const mapped = mapPostPayload(payload);

    setAllposts((prev) => [mapped, ...prev]);
  }

  return (
    <div className="home-container">
      <div className="posts">
        <NewPost onCreate={handleNewPost} />
        {loading && <div>Loading posts...</div>}
        {error && <div className="error">{error}</div>}
        {Allposts?.map((it) => {
          return (
            <Post
              key={it.id}
              post={it}
              onPostChange={handlePostChange}
              onPostDeleted={handlePostChange}
            />
          );
        })}
      </div>
    </div>
  );
}

export default Home;
