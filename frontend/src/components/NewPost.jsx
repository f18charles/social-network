import { useEffect, useMemo, useRef, useState } from "react";
import "../styles/newpost.css";
import avatar from "../assets/user.svg";
import { apiFetch } from "../utils/api.js";
import { logger } from "../utils/logger.js";

const getUserName = (user) =>
  user?.nickname ||
  `${user?.first_name || ""} ${user?.last_name || ""}`.trim() ||
  user?.email ||
  "Follower";

/**
 * NewPost renders the shared multipart post form for create and edit flows.
 */
export default function NewPost({
  mode = "create",
  post = null,
  groupId = null,
  onCreate,
  onUpdate,
  onCancel,
}) {
  const isEdit = mode === "edit";
  const isGroupPost = Boolean(groupId || post?.group_id);
  const [content, setContent] = useState(post?.content || "");
  const [privacy, setPrivacy] = useState(post?.privacy || "public");
  const [file, setFile] = useState(null);
  const [preview, setPreview] = useState(post?.image_url || null);
  const [removeImage, setRemoveImage] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [followers, setFollowers] = useState([]);
  const [audienceIDs, setAudienceIDs] = useState([]);
  const objectUrlRef = useRef(null);
  const fileInputRef = useRef(null);

  const submitLabel = useMemo(() => {
    if (loading) return isEdit ? "Saving..." : "Posting...";
    return isEdit ? "Save" : "Post";
  }, [isEdit, loading]);

  useEffect(() => {
    if (privacy !== "private" || isGroupPost) return undefined;

    let active = true;
    apiFetch("/api/followers/followers")
      .then((data) => {
        if (active) setFollowers(Array.isArray(data) ? data : []);
      })
      .catch((err) => {
        logger.error("Failed to load post audience followers", err);
        if (active) setError("Unable to load followers for private audience.");
      });

    return () => {
      active = false;
    };
  }, [isGroupPost, privacy]);

  useEffect(
    () => () => {
      if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current);
    },
    []
  );

  const resetObjectPreview = () => {
    if (objectUrlRef.current) {
      URL.revokeObjectURL(objectUrlRef.current);
      objectUrlRef.current = null;
    }
  };

  const handleFile = (event) => {
    const selected = event.target.files?.[0] || null;
    resetObjectPreview();
    setFile(selected);
    setRemoveImage(false);
    if (selected) {
      const objectUrl = URL.createObjectURL(selected);
      objectUrlRef.current = objectUrl;
      setPreview(objectUrl);
    }
  };

  const clearImage = () => {
    resetObjectPreview();
    setFile(null);
    setPreview(null);
    setRemoveImage(Boolean(isEdit && post?.image_url));
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const resetForm = () => {
    resetObjectPreview();
    setContent("");
    setFile(null);
    setPreview(null);
    setPrivacy("public");
    setRemoveImage(false);
    setAudienceIDs([]);
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  async function handleSubmit(event) {
    event.preventDefault();
    setError(null);

    const trimmed = content.trim();
    if (!trimmed && !file && !(isEdit && post?.image_url && !removeImage)) {
      setError("Write something or choose an image.");
      return;
    }
    if (!isGroupPost && privacy === "private" && audienceIDs.length === 0) {
      setError("Choose at least one follower for a private post.");
      return;
    }

    const form = new FormData();
    form.append("content", content);
    if (!isGroupPost) {
      form.append("privacy", privacy);
      audienceIDs.forEach((id) => form.append("audience_ids", id));
    }
    if (groupId) form.append("group_id", groupId);
    if (file) form.append("image", file);
    if (isEdit && removeImage) form.append("remove_image", "true");

    setLoading(true);
    try {
      const saved = await apiFetch(isEdit ? `/api/posts/${post.id}` : "/api/posts", {
        method: isEdit ? "PATCH" : "POST",
        body: form,
      });

      if (isEdit) {
        onUpdate?.(saved);
      } else {
        resetForm();
        onCreate?.(saved);
      }
    } catch (err) {
      logger.error("Failed to submit post", err);
      setError(err.message || "Unable to save post.");
    } finally {
      setLoading(false);
    }
  }

  const toggleAudience = (id) => {
    setAudienceIDs((current) =>
      current.includes(id)
        ? current.filter((selectedID) => selectedID !== id)
        : [...current, id]
    );
  };

  return (
    <form className="newpost" onSubmit={handleSubmit}>
      <div className="np-top">
        <img src={post?.author?.avatar || avatar} alt="" className="np-avatar" />
        <textarea
          aria-label={isEdit ? "Edit post content" : "Post content"}
          placeholder={isGroupPost ? "Share with this group" : "What's on your mind?"}
          value={content}
          onChange={(event) => setContent(event.target.value)}
          rows={3}
        />
      </div>

      {preview && (
        <div className="np-preview">
          <img src={preview} alt="Selected post attachment preview" />
          <button type="button" onClick={clearImage}>
            Remove
          </button>
        </div>
      )}

      {!isGroupPost ? (
        <div className="np-privacy">
          <label>
            <span>Privacy</span>
            <select
              value={privacy}
              onChange={(event) => {
                setPrivacy(event.target.value);
                setAudienceIDs([]);
              }}
            >
              <option value="public">Public</option>
              <option value="almost_private">Followers</option>
              <option value="private">Private</option>
            </select>
          </label>
          {privacy === "private" ? (
            <fieldset className="np-audience">
              <legend>Audience</legend>
              {followers.length === 0 ? (
                <p>No accepted followers available.</p>
              ) : (
                followers.map((follower) => (
                  <label key={follower.id}>
                    <input
                      type="checkbox"
                      checked={audienceIDs.includes(follower.id)}
                      onChange={() => toggleAudience(follower.id)}
                    />
                    <span>{getUserName(follower)}</span>
                  </label>
                ))
              )}
            </fieldset>
          ) : null}
        </div>
      ) : null}

      <div className="np-controls">
        <label className="np-file">
          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/gif"
            onChange={handleFile}
          />
          Add image
        </label>

        <div className="np-actions">
          {isEdit && onCancel ? (
            <button type="button" className="np-secondary" onClick={onCancel}>
              Cancel
            </button>
          ) : null}
          <button
            type="submit"
            disabled={loading || (!isEdit && !content.trim() && !file)}
          >
            {submitLabel}
          </button>
        </div>
      </div>
      {error ? <div className="error">{error}</div> : null}
    </form>
  );
}
