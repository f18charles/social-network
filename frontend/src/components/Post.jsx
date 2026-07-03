import { useState } from "react";
import "../styles/post.css";
import { MdDelete, MdEdit, MdLock, MdPeople, MdPublic } from "react-icons/md";
import { useNavigate } from "react-router";
import AuthorMeta from "./AuthorMeta.jsx";
import NewPost from "./NewPost.jsx";
import { VoteControls } from "./VoteControls";
import { useAuth } from "../context/auth/useAuth.js";
import { apiFetch } from "../utils/api.js";
import { logger } from "../utils/logger.js";

const privacyMeta = {
  public: { label: "Public", icon: MdPublic },
  almost_private: { label: "Followers", icon: MdPeople },
  private: { label: "Private", icon: MdLock },
};

const formatPostTime = (datestring, now) => {
  const date = new Date(datestring);
  const diffInMs = now - date.getTime();

  const ONE_MINUTE = 60000;
  const ONE_HOUR = 3600000;
  const ONE_DAY = 86400000;
  const ONE_MONTH = 2592000000;
  const ONE_YEAR = 31536000000;

  if (Number.isNaN(date.getTime())) return "";
  if (diffInMs < 0) return "In the future";
  if (diffInMs < ONE_HOUR)
    return `${Math.floor(diffInMs / ONE_MINUTE)} minutes ago`;
  if (diffInMs < ONE_DAY) return `${Math.floor(diffInMs / ONE_HOUR)} hours ago`;
  if (diffInMs < ONE_MONTH) return `${Math.floor(diffInMs / ONE_DAY)} days ago`;
  if (diffInMs < ONE_YEAR)
    return `${Math.floor(diffInMs / ONE_MONTH)} months ago`;
  return `${Math.floor(diffInMs / ONE_YEAR)} years ago`;
};

const getPrivacyMeta = (post) => {
  if (post?.group_id) return { label: "Group", icon: MdPeople };
  return privacyMeta[String(post?.privacy || "public").toLowerCase()] || privacyMeta.public;
};

/**
 * Post renders a post summary with privacy, owner controls, and vote controls.
 */
const Post = ({ post, onPostChange, onPostDeleted }) => {
  const [voteOverride, setVoteOverride] = useState(null);
  const [isVoting, setIsVoting] = useState(false);
  const [voteError, setVoteError] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [actionError, setActionError] = useState(null);
  const [renderedAt] = useState(() => Date.now());
  const navigate = useNavigate();
  const { currentUser } = useAuth();

  const localPost =
    voteOverride?.id === post?.id ? { ...post, ...voteOverride } : post;
  const isOwner =
    Boolean(currentUser?.id) && String(currentUser.id) === String(localPost?.author?.id);
  const canMutate = isOwner && !localPost?.deleted;

  const updatePost = (nextPost) => {
    setVoteOverride({
      id: nextPost.id,
      like_count: nextPost.like_count,
      dislike_count: nextPost.dislike_count,
      heart_count: nextPost.heart_count,
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
        heart_count: summary?.heart_count ?? localPost.heart_count ?? 0,
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

  const stopCardAction = (event) => event.stopPropagation();

  const handleDelete = async (event) => {
    stopCardAction(event);
    if (!canMutate || isDeleting) return;

    setIsDeleting(true);
    setActionError(null);
    try {
      const deleted = await apiFetch(`/api/posts/${localPost.id}`, {
        method: "DELETE",
      });
      onPostDeleted?.(deleted);
      onPostChange?.(deleted);
    } catch (err) {
      logger.error("Failed to delete post", err, { postId: localPost.id });
      setActionError("Unable to delete post.");
    } finally {
      setIsDeleting(false);
    }
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

  if (isEditing) {
    return (
      <div className="post-container post-container--editing" onClick={stopCardAction}>
        <NewPost
          mode="edit"
          post={localPost}
          onUpdate={(updated) => {
            setIsEditing(false);
            onPostChange?.(updated);
          }}
          onCancel={() => setIsEditing(false)}
        />
      </div>
    );
  }

  const meta = getPrivacyMeta(localPost);
  const PrivacyIcon = meta.icon;

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
        <div className="post-toolbar">
          <div
            className="visibility"
            aria-label={`${meta.label} post`}
            title={`${meta.label} post`}
          >
            <PrivacyIcon aria-hidden="true" />
            <span>{meta.label}</span>
          </div>
          {canMutate ? (
            <div className="post-actions" onClick={stopCardAction}>
              <button
                type="button"
                aria-label="Edit post"
                title="Edit post"
                onClick={() => setIsEditing(true)}
              >
                <MdEdit aria-hidden="true" />
              </button>
              <button
                type="button"
                aria-label="Delete post"
                title="Delete post"
                onClick={handleDelete}
                disabled={isDeleting}
              >
                <MdDelete aria-hidden="true" />
              </button>
            </div>
          ) : null}
        </div>
      </div>
      <div className="post-body">
        <p>{localPost?.content}</p>
        {localPost?.updated_at ? <small>Edited</small> : null}
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
      <div className="post-footer" onClick={stopCardAction}>
        <VoteControls
          likes={localPost?.like_count || 0}
          dislikes={localPost?.dislike_count || 0}
          hearts={localPost?.heart_count || 0}
          currentVote={localPost?.viewer_vote || "none"}
          targetType="post"
          isDisabled={localPost?.deleted}
          isMutating={isVoting}
          onVote={handleVote}
          showHeart={true}
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
      {actionError ? <div className="error">{actionError}</div> : null}
    </div>
  );
};

export default Post;
