import React, { useState } from "react";
import { MdAddReaction } from "react-icons/md";
import "../styles/emoji-reactions.css";

const EmojiReactions = ({
  reactions = [],
  onReact,
  targetType = "message",
}) => {
  const [showPicker, setShowPicker] = useState(false);
  const commonEmojis = ["👍", "❤️", "😂", "😮", "😢", "🙏"];

  const handleEmojiClick = (emoji, event) => {
    event?.stopPropagation();
    onReact?.(emoji);
    setShowPicker(false);
  };

  return (
    <div className="emoji-reactions-container" onClick={(e) => e.stopPropagation()}>
      <div className="reactions-list">
        {reactions.map((r) => (
          <button
            key={r.emoji}
            type="button"
            className={`reaction-badge ${r.user_reacted ? "active" : ""}`}
            onClick={(e) => handleEmojiClick(r.emoji, e)}
            title={`${r.count} reactions`}
          >
            <span className="emoji-char">{r.emoji}</span>
            <span className="emoji-count">{r.count}</span>
          </button>
        ))}
        
        {targetType === "message" && (
          <div className="add-reaction-wrapper">
            <button
              type="button"
              className="add-reaction-trigger"
              onClick={(e) => {
                e.stopPropagation();
                setShowPicker(!showPicker);
              }}
              title="Add reaction"
            >
              <MdAddReaction size={18} />
            </button>
            {showPicker && (
              <div className="mini-emoji-picker">
                {commonEmojis.map((emoji) => (
                  <button
                    key={emoji}
                    type="button"
                    className="picker-emoji-btn"
                    onClick={(e) => handleEmojiClick(emoji, e)}
                  >
                    {emoji}
                  </button>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default EmojiReactions;
