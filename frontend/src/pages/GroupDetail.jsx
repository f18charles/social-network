import { useEffect, useState } from "react";
import { Link, useParams } from "react-router";
import NewPost from "../components/NewPost.jsx";
import Post from "../components/Post.jsx";
import { apiFetch } from "../utils/api.js";
import { logger } from "../utils/logger.js";
import "../styles/home.css";

const getListData = (response) =>
  Array.isArray(response) ? response : response?.data || [];

/**
 * GroupDetail renders the accepted-member post feed for a single group.
 */
export default function GroupDetail() {
  const { groupId } = useParams();
  const [group, setGroup] = useState(null);
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    let active = true;

    async function loadGroup() {
      setLoading(true);
      setError(null);
      try {
        const groupData = await apiFetch(`/api/groups/${groupId}`);
        if (!active) return;
        setGroup(groupData);

        if (groupData?.status === "accepted" || groupData?.is_member) {
          const postData = await apiFetch(`/api/posts?group_id=${groupId}`);
          if (active) setPosts(getListData(postData));
        } else {
          setPosts([]);
        }
      } catch (err) {
        logger.error("Failed to load group feed", err, { groupId });
        if (active) setError("Unable to load group.");
      } finally {
        if (active) setLoading(false);
      }
    }

    if (groupId) loadGroup();

    return () => {
      active = false;
    };
  }, [groupId]);

  const isAcceptedMember = group?.status === "accepted" || group?.is_member;

  const handleCreate = (created) => {
    if (!created) return;
    setPosts((current) => [created, ...current]);
  };

  const handlePostChange = (updatedPost) => {
    if (!updatedPost?.id) return;
    setPosts((current) =>
      current.map((post) => (post.id === updatedPost.id ? updatedPost : post))
    );
  };

  return (
    <div className="home-container">
      <div className="posts">
        <Link to="/groups" className="group-back-link">
          Back to groups
        </Link>
        {loading ? <div>Loading group...</div> : null}
        {error ? <div className="error">{error}</div> : null}
        {group ? (
          <header className="group-feed-header">
            {group.avatar ? <img src={group.avatar} alt="" /> : null}
            <div>
              <h2>{group.title}</h2>
              <p>{group.description || "No description provided."}</p>
            </div>
          </header>
        ) : null}

        {isAcceptedMember ? (
          <NewPost groupId={groupId} onCreate={handleCreate} />
        ) : group && !loading ? (
          <div className="profile-state">Only accepted members can view or post in this group.</div>
        ) : null}

        {isAcceptedMember && posts.length === 0 && !loading ? (
          <div className="profile-state">No group posts yet.</div>
        ) : null}
        {isAcceptedMember
          ? posts.map((post) => (
              <Post
                key={post.id}
                post={post}
                onPostChange={handlePostChange}
                onPostDeleted={handlePostChange}
              />
            ))
          : null}
      </div>
    </div>
  );
}
