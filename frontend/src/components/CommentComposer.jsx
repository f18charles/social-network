import { useEffect, useRef, useState } from "react";
import "../styles/comment.css";
import { logger } from "../utils/logger.js";

/**
 * CommentComposer builds multipart comment payloads for create, reply, and edit flows.
 */
export default function CommentComposer({
  initialComment = null,
  onSubmit,
  onCancel,
  placeholder = "Write a comment",
  submitLabel = "Post",
}) {
  const [content, setContent] = useState(initialComment?.content || "");
  const [image, setImage] = useState(null);
  const [preview, setPreview] = useState(initialComment?.image_url || null);
  const [removeImage, setRemoveImage] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState(null);
  const objectUrlRef = useRef(null);
  const fileInputRef = useRef(null);

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

  const resetForm = () => {
    resetObjectPreview();
    setContent("");
    setImage(null);
    setPreview(null);
    setRemoveImage(false);
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const handleImage = (event) => {
    const selected = event.target.files?.[0] || null;
    resetObjectPreview();
    setImage(selected);
    setRemoveImage(false);
    if (selected) {
      const objectUrl = URL.createObjectURL(selected);
      objectUrlRef.current = objectUrl;
      setPreview(objectUrl);
    }
  };

  const clearImage = () => {
    resetObjectPreview();
    setImage(null);
    setPreview(null);
    setRemoveImage(Boolean(initialComment?.image_url));
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    if (isSubmitting) return;

    const trimmed = content.trim();
    if (!trimmed && !image && !(initialComment?.image_url && !removeImage)) {
      setError("Write a comment or choose an image.");
      return;
    }

    const formData = new FormData();
    if (trimmed || initialComment) formData.append("content", content);
    if (image) formData.append("image", image);
    if (removeImage) formData.append("remove_image", "true");

    setIsSubmitting(true);
    setError(null);
    try {
      await onSubmit(formData);
      if (!initialComment) resetForm();
    } catch (err) {
      logger.error("Failed to submit comment", err);
      setError(err.message || "Unable to save comment.");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleKeyDown = (event) => {
    if (event.key !== "Enter" || event.shiftKey) return;
    event.preventDefault();
    event.currentTarget.form?.requestSubmit();
  };

  return (
    <form className="comment-composer" onSubmit={handleSubmit}>
      <textarea
        aria-label={placeholder}
        value={content}
        placeholder={placeholder}
        onChange={(event) => setContent(event.target.value)}
        onKeyDown={handleKeyDown}
        rows={3}
      />
      {preview ? (
        <div className="comment-composer-preview">
          <img src={preview} alt="Selected comment attachment preview" />
          <button type="button" onClick={clearImage}>
            Remove
          </button>
        </div>
      ) : null}
      <div className="comment-composer-actions">
        <input
          ref={fileInputRef}
          type="file"
          accept="image/jpeg,image/png,image/gif"
          onChange={handleImage}
        />
        <div className="comment-composer-buttons">
          {onCancel ? (
            <button type="button" onClick={onCancel}>
              Cancel
            </button>
          ) : null}
          <button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Saving..." : submitLabel}
          </button>
        </div>
      </div>
      {error ? <div className="error">{error}</div> : null}
    </form>
  );
}
