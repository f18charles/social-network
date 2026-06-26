import { useState, useEffect } from "react";
import { posts } from "../assets/posts-data.js";
import Post from "../components/Post.jsx";
import "../styles/home.css";
import NewPost from "../components/NewPost.jsx";

function Home() {
  const initialPosts = posts || [];
  const [Allposts, setAllposts] = useState(initialPosts);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // helper to map backend post DTO to Post component shape
  function mapPostPayload(payload) {
    return {
      id: payload.id,
      author: payload.author
        ? {
            id: payload.author.id,
            name:
              payload.author.nickname ||
              `${payload.author.first_name || ""} ${payload.author.last_name || ""}`.trim() ||
              payload.author.email || undefined,
            first_name: payload.author.first_name,
            last_name: payload.author.last_name,
            nickname: payload.author.nickname,
            avatar: payload.author.avatar,
          }
        : null,
      content: payload.content,
      image_url: payload.image_url || null,
      privacy: payload.privacy || "public",
      like_count: payload.like_count || 0,
      comment_count: payload.comment_count || 0,
      created_at: payload.created_at || new Date().toISOString(),
    };
  }

  async function fetchPosts() {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch("/api/posts", { credentials: "include" });
      if (!res.ok) throw new Error(`Failed to fetch posts (${res.status})`);
      const json = await res.json();
      const data = json && json.data ? json.data : [];
      const mapped = Array.isArray(data) ? data.map(mapPostPayload) : [];
      setAllposts(mapped);
    } catch (err) {
      console.error(err);
      setError(err.message || "Failed to load posts");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchPosts();
  }, []);

  function handleNewPost(created) {
    if (!created) return;

    // Backend may return an envelope or the raw post object. normalize to the post payload.
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
        {Allposts?.map((it, idx) => {
          return <Post key={idx} post={it} />;
        })}
      </div>
    </div>
  );
}

export default Home;
