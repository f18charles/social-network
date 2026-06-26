import React, { useState } from "react";
import "../styles/newpost.css";
import avatar from "../assets/user.svg";

export default function NewPost({ onCreate }) {
  const [content, setContent] = useState("");
  const [privacy, setPrivacy] = useState("public");
  const [file, setFile] = useState(null);
  const [preview, setPreview] = useState(null);
  const [loading, setLoading] = useState(false);
  const token = localStorage.getItem("token");

  async function handleSubmit(e) {
    e.preventDefault();
    if (!content.trim() && !file) return;

    setLoading(true);
    try {
      const form = new FormData();
      form.append("content", content);
      form.append("privacy", privacy);
      if (file) form.append("image", file);

      const res = await fetch("/api/posts", {
        method: "POST",
        // send cookies (server uses session cookie) and let browser handle auth
        credentials: "include",
        body: form,
      });

      if (!res.ok) {
        throw new Error(`Failed to create post (${res.status})`);
      }

      const json = await res.json();
      // backend uses envelope: { status, message, data, errors }
      const created = json && json.data ? json.data : null;
      setContent("");
      setFile(null);
      setPreview(null);
      setPrivacy("public");
      if (typeof onCreate === "function" && created) onCreate(created);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }

  function handleFile(e) {
    const f = e.target.files[0];
    if (!f) return;
    setFile(f);
    setPreview(URL.createObjectURL(f));
  }

  return (
    <form className="newpost" onSubmit={handleSubmit}>
      <div className="np-top">
        <img src={avatar} alt="me" className="np-avatar" />
        <textarea
          placeholder="What's on your mind?"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          rows={3}
        />
      </div>

      {preview && (
        <div className="np-preview">
          <img src={preview} alt="preview" />
          <button
            type="button"
            onClick={() => {
              setFile(null);
              setPreview(null);
            }}
          >
            Remove
          </button>
        </div>
      )}

      <div className="np-controls">
        <label className="np-file">
          <input type="file" accept="image/*" onChange={handleFile} />
          Add image
        </label>

        <select value={privacy} onChange={(e) => setPrivacy(e.target.value)}>
          <option value="public">Public</option>
          <option value="private">Private</option>
        </select>

        <button type="submit" disabled={loading || (!content.trim() && !file)}>
          {loading ? "Posting..." : "Post"}
        </button>
      </div>
    </form>
  );
}
