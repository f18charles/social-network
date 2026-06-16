import { useState } from "react";
import "../styles/post.css";
import { Dislike, Like } from "./Reactions";
import avatar from "../assets/user.svg";

const Post = ({ post }) => {
  const [likePost, setLikePost] = useState(false);
  const [dislikePost, setDislikePost] = useState(false);

  function like() {
    setDislikePost(false);
    setLikePost((prev) => !prev);
  }

  function dislike() {
    setLikePost(false);
    setDislikePost((prev) => !prev);
  }

  return (
    <div className="post-container">
      <div className="post-header">
        <img
          src={post?.author?.avatar ? post.author.avatar : avatar}
          alt="avatar"
          className="profile-photo"
        />
        <div className="post-bio">
          <h5>{post?.author?.name}</h5>
          <small>16:06</small>
        </div>
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
        <div className="reaction-container">
          <Like like={like} isActive={likePost} />
          <Dislike dislike={dislike} isActive={dislikePost} />
        </div>
        <div>
          <p className="links">Comment</p>
        </div>
        <div>
          <p className="links">Share</p>
        </div>
      </div>
    </div>
  );
};

export default Post;
