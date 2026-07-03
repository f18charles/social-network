import { useEffect, useState } from "react";
import Post from "../Post";
import { apiFetch } from "../../utils/api";
import { logger } from "../../utils/logger.js";

/**
 * Fetches and renders the posts belonging to a profile, reusing the
 * existing Post card component used on the feed.
 *
 * @param {{ userId: string }} props
 */
const ProfilePosts = ({ userId }) => {
  const [posts, setPosts] = useState([]);
  const [status, setStatus] = useState("loading"); // loading | error | ready
  const [error, setError] = useState("");

  useEffect(() => {
    if (!userId) return;

    let isActive = true;

    apiFetch(`/api/users/${userId}/posts`)
      .then((result) => {
        if (!isActive) return;
        const list = Array.isArray(result) ? result : result?.data || [];
        setPosts(list);
        setStatus("ready");
      })
      .catch((err) => {
        logger.error("Failed to load profile posts", err, { userId });
        if (!isActive) return;
        setError("Unable to load posts.");
        setStatus("error");
      });

    return () => {
      isActive = false;
    };
  }, [userId]);

  if (status === "loading") {
    return (
      <div className="profile-posts">
        <div className="profile-skeleton profile-skeleton--row" />
        <div className="profile-skeleton profile-skeleton--row" />
      </div>
    );
  }

  if (status === "error") {
    return <div className="profile-state profile-state--error">{error}</div>;
  }

  const handlePostChange = (updatedPost) => {
    if (!updatedPost?.id) return;
    setPosts((current) =>
      current.map((post) => (post.id === updatedPost.id ? updatedPost : post))
    );
  };

  if (posts.length === 0) {
    return <div className="profile-state">No posts yet. Share something!</div>;
  }

  return (
    <div className="profile-posts">
      {posts.map((post) => (
        <Post
          key={post.id}
          post={post}
          onPostChange={handlePostChange}
          onPostDeleted={handlePostChange}
        />
      ))}
    </div>
  );
};

export default ProfilePosts;
