import { useState } from "react";
import "../styles/post.css";
import { MdPublic } from "react-icons/md";
import { useNavigate } from "react-router";
import AuthorMeta from "./AuthorMeta.jsx";
import { VoteControls } from "./VoteControls";
import { apiFetch } from "../utils/api.js";
import { logger } from "../utils/logger.js";

const formatPostTime = (datestring, now) => {
  const date = new Date(datestring);
  const diffInMs = now - date.getTime();

  const ONE_MINUTE = 60000;
  const ONE_HOUR = 3600000;
  const ONE_DAY = 86400000;
  const ONE_MONTH = 2592000000;
  const ONE_YEAR = 31536000000;

  if (diffInMs < 0) return "In the future";
  if (diffInMs < ONE_HOUR)
    return `${Math.floor(diffInMs / ONE_MINUTE)} minutes ago`;
  if (diffInMs < ONE_DAY) return `${Math.floor(diffInMs / ONE_HOUR)} hours ago`;
  if (diffInMs < ONE_MONTH) return `${Math.floor(diffInMs / ONE_DAY)} days ago`;
  if (diffInMs < ONE_YEAR)
    return `${Math.floor(diffInMs / ONE_MONTH)} months ago`;
  return `${Math.floor(diffInMs / ONE_YEAR)} years ago`;
};

/**
 * Post renders a post summary with API-backed vote controls.
 */
const Post = ({ post, onPostChange }) => {
  const [voteOverride, setVoteOverride] = useState(null);
  const [isVoting, setIsVoting] = useState(false);
  const [voteError, setVoteError] = useState(null);
  const [renderedAt] = useState(() => Date.now());
  const navigate = useNavigate();

  const localPost =
    voteOverride?.id === post?.id ? { ...post, ...voteOverride } : post;

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
          ? await apiFetch(`/api/posts/${localPost.id}/vote`, {
              method: "DELETE",
            })
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
      logger.error("Failed to update post vote", err, {
        postId: localPost.id,
        vote,
      });
      setVoteError("Unable to update vote.");
    } finally {
      setIsVoting(false);
    }
  };

  const openPost = (event, selectedPost) => {
    event.stopPropagation();
    if (!selectedPost?.id) return;
    navigate(`/post/${selectedPost.id}`, { state: selectedPost });
  };

  if (localPost?.deleted) {
    return (
      <div
        className="post-container post-container--deleted"
        onClick={(event) => openPost(event, localPost)}
      >
        <div className="top-bar">
          <AuthorMeta
            author={{ name: "Deleted user" }}
            className="post-header"
          />
        </div>
        <div className="post-body post-body--deleted">
          <p>This post is no longer available.</p>
        </div>
      </div>
    );
  }

  return (
    <div
      className="post-container"
      onClick={(event) => openPost(event, localPost)}
    >
      <div className="top-bar">
        <AuthorMeta
          author={localPost?.author}
          timestamp={formatPostTime(localPost?.created_at, renderedAt)}
          className="post-header"
        />
        {String(localPost?.privacy).toLowerCase() === "public" && (
          <div className="visibility">
            <MdPublic />
            <span>public</span>
          </div>
        )}
      </div>
      <div className="post-body">
        <p>{localPost?.content}</p>
        {localPost?.image_url ? (
          <img
            className="post-image"
            src={localPost.image_url}
            alt="post attachment"
          />
        ) : null}
      </div>
      <div className="reaction-count">
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
