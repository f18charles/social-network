import { useState } from "react";
import "../styles/post.css";
import { Like } from "./Reactions";
import avatar from "../assets/user.svg";
import { MdPublic } from "react-icons/md";
import { useNavigate } from "react-router";
import { VoteControls } from "./VoteControls";
import { apiFetch } from "../utils/api.js";

/**
 * Post renders a post summary with API-backed vote controls.
 */
const Post = ({ post, onPostChange }) => {
  const [voteOverride, setVoteOverride] = useState(null);
  const [isVoting, setIsVoting] = useState(false);
  const [voteError, setVoteError] = useState(null);
  const [renderedAt] = useState(() => Date.now());

  const navigate = useNavigate();


  const localPost = voteOverride?.id === post?.id ? { ...post, ...voteOverride } : post;

  const updatePost = (nextPost) => {
    setVoteOverride({
      id: nextPost.id,
      like_count: nextPost.like_count,
      dislike_count: nextPost.dislike_count,
      viewer_vote: nextPost.viewer_vote,
    });
    onPostChange?.(nextPost);
  };

  const handleVote = async (vote) => {
    if (!localPost?.id || localPost?.deleted) return;

    setIsVoting(true);
    setVoteError(null);
    try {
      const currentVote = localPost.viewer_vote || "none";
      const summary =
        currentVote === vote
          ? await apiFetch(`/api/posts/${localPost.id}/vote`, { method: "DELETE" })
          : await apiFetch(`/api/posts/${localPost.id}/vote`, {
              method: "PUT",
              body: { vote },
            });

      updatePost({
        ...localPost,
        like_count: summary?.like_count ?? localPost.like_count ?? 0,
        dislike_count: summary?.dislike_count ?? localPost.dislike_count ?? 0,
        viewer_vote: summary?.viewer_vote || "none",
      });
    } catch (err) {
      setVoteError(err?.message || "Unable to update vote.");
    } finally {
      setIsVoting(false);
    }
  };

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

      default: {
        const yearsAgo = Math.floor(diffInMs / ONE_YEAR);
        return `${yearsAgo} ${yearsAgo === 1 ? "year" : "years"} ago`;
      }
    }
  };

  const openPost = (event, selectedPost) => {
    event.stopPropagation();
    if (!selectedPost?.id) return;
    navigate(`/post/${selectedPost.id}`, {
      state: selectedPost,
    });
  };

  const authorName = localPost?.author
    ? localPost.author.nickname ||
      `${localPost.author.first_name || ""} ${localPost.author.last_name || ""}`.trim() ||
      localPost.author.name
    : "Unknown User";

  return (
    <div className="post-container" onClick={(e) => openPost(e, localPost)}>
      <div className="top-bar">
        <div className="post-header">
          <img
            src={localPost?.author?.avatar ? localPost.author.avatar : avatar}
            alt="avatar"
            className="profile-photo"
            onClick={(e) => {
              e.stopPropagation();
              if (localPost?.author?.id) {
                navigate(`/user/${localPost.author.id}`);
              }
            }}
            style={{ cursor: "pointer" }}
          />

          <div className="post-bio">
            <h5
              onClick={(e) => {
                e.stopPropagation();
                if (localPost?.author?.id) {
                  navigate(`/user/${localPost.author.id}`);
                }
              }}
              style={{ cursor: "pointer" }}
            >
              {authorName}
            </h5>
            <small>{DateFormatter(localPost?.created_at, renderedAt)}</small>
          </div>
        </div>
        {String(localPost?.privacy).toLowerCase() == "public" && (
          <div className="visibility">
            <MdPublic />
            <span>public</span>
          </div>
        )}
      </div>
      <div className="post-body">
        <p>{localPost?.content}</p>
        {localPost?.image_url ? (
          <img className="post-image" src={localPost.image_url} alt="post-image" />
        ) : null}
      </div>
      <div className="reaction-count">
        <div className="center">
          <Like /> {localPost?.like_count || 0}
        </div>
        <div>{localPost?.comment_count || 0} Comments</div>
      </div>
      <div className="post-footer">
        <VoteControls
          likes={localPost?.like_count || 0}
          dislikes={localPost?.dislike_count || 0}
          currentVote={localPost?.viewer_vote || "none"}
          targetType="post"
          isDisabled={localPost?.deleted}
          isMutating={isVoting}
          onVote={handleVote}
        />
        <button
          type="button"
          className="reaction-button"
          onClick={(event) => openPost(event, localPost)}
        >
          Comment
        </button>
      </div>
      {voteError ? <div className="error">{voteError}</div> : null}
    </div>
  );
};

export default Post;
