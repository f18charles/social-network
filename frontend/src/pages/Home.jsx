import React from "react";
import { posts } from "../assets/posts-data.js";
import { comments } from "../assets/comments-data.js";
import Post from "../components/Post.jsx";
import "../styles/home.css";
import UpcomingEvent from "../components/UpcomingEvent.jsx";

function Home() {
  const posts = [];

  return (
    <div className="home-container">
      <div className="posts">
        {posts.map((it, idx) => {
          return <Post key={idx} post={it} />;
        })}
      </div>
    </div>
  );
}

export default Home;
