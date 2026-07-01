import { useState } from "react";
import "../styles/post.css";
import { Dislike, Like } from "./Reactions";
import avatar from "../assets/user.svg";
import { MdPublic } from "react-icons/md";
import { useNavigate } from "react-router";

const Post = ({ post }) => {
  const [likePost, setLikePost] = useState(false);
  const [dislikePost, setDislikePost] = useState(false);
  const [renderedAt] = useState(() => Date.now());

  const navigate = useNavigate();

  function like() {
    setDislikePost(false);
    setLikePost((prev) => !prev);
  }

  function dislike() {
    setLikePost(false);
    setDislikePost((prev) => !prev);
  }

  const DateFormatter = (datestring, now) => {
    const date = new Date(datestring);
    const diffInMs = now - date.getTime(); // Positive if date is in the past

    // Millisecond constants
    const ONE_MINUTE = 60000;
    const ONE_HOUR = 3600000;
    const ONE_DAY = 86400000;
    const ONE_MONTH = 2592000000; // 30 days
    const ONE_YEAR = 31536000000; // 365 days

    // Handle future dates safely
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

  const openPost = (event, post) => {
    event.stopPropagation();
    navigate(`/post/${post.id}`, {
      state: post,
    });
  };

  const authorName = post?.author
    ? post.author.nickname ||
      `${post.author.first_name || ""} ${post.author.last_name || ""}`.trim() ||
      post.author.name
    : "Unknown User";

  return (
    <div className="post-container" onClick={(e) => openPost(e, post)}>
      <div className="top-bar">
        <div className="post-header">
          <img
            src={post?.author?.avatar ? post.author.avatar : avatar}
            alt="avatar"
            className="profile-photo"
            onClick={(e) => {
              e.stopPropagation();
              if (post?.author?.id) {
                navigate(`/user/${post.author.id}`);
              }
            }}
            style={{ cursor: "pointer" }}
          />

          <div className="post-bio">
            <h5
              onClick={(e) => {
                e.stopPropagation();
                if (post?.author?.id) {
                  navigate(`/user/${post.author.id}`);
                }
              }}
              style={{ cursor: "pointer" }}
            >
              {authorName}
            </h5>
            <small>{DateFormatter(post?.created_at, renderedAt)}</small>
          </div>
        </div>
        {String(post?.privacy).toLowerCase() == "public" && (
          <div className="visibility">
            <MdPublic />
            <span>public</span>
          </div>
        )}
      </div>
      <div className="post-body">
        <p>{post?.content}</p>
        {post["image_url"] ? (
          <img
            className="post-image"
            src={post["image_url"]}
            alt="post-image"
          />
        ) : (
          <></>
        )}
      </div>
      <div className="reaction-count">
        <div className="center">
          <Like /> {post["like_count"]}
        </div>
        <div>{post["comment_count"]} Comments</div>
      </div>
      <div className="post-footer">
        <div>
          <Like like={like} isActive={likePost} />
        </div>
        <div>
          <Dislike dislike={dislike} isActive={dislikePost} />
        </div>
        <div>
          <p className="reaction-button">Comment</p>
        </div>
        <div>
          <p className="reaction-button">Share</p>
        </div>
      </div>
    </div>
  );
};

export default Post;
