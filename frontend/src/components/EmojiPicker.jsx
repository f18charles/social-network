import React, { useState } from "react";
import { BsEmojiSmile } from "react-icons/bs";

const EmojiPicker = ({ onSelectEmoji }) => {
  const [showPicker, setShowPicker] = useState(false);
  const emojis = [
    "😀", "😂", "🥰", "😍", "😎", "🤔", "😮", "😢", "😡", "👍", "👎",
    "🎉", "🔥", "❤️", "✨", "💯", "👏", "🙌", "🎂", "🚀", "💡", "🌟",
    "🍕", "🍺", "☕", "🐱", "🐶", "🌈", "☀️", "✈️", "🎮", "🎵"
  ];

  const handleEmojiClick = (emoji, event) => {
    event.stopPropagation();
    onSelectEmoji?.(emoji);
    setShowPicker(false);
  };

  return (
    <div className="emoji-picker-container" style={{ position: "relative", display: "inline-block" }}>
      <button
        type="button"
        className="emoji-picker-trigger"
        onClick={(e) => {
          e.stopPropagation();
          setShowPicker(!showPicker);
        }}
        title="Insert Emoji"
        style={{
          background: "none",
          border: "none",
          cursor: "pointer",
          padding: "4px 8px",
          display: "flex",
          alignItems: "center",
          color: "var(--text-secondary, #65676b)"
        }}
      >
        <BsEmojiSmile size={20} />
      </button>
      {showPicker && (
        <>
          <div
            className="emoji-picker-overlay"
            onClick={() => setShowPicker(false)}
            style={{
              position: "fixed",
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              zIndex: 999
            }}
          />
          <div
            className="emoji-picker-popover"
            style={{
              position: "absolute",
              bottom: "100%",
              left: 0,
              marginBottom: "8px",
              background: "var(--bg-primary, #ffffff)",
              border: "1px solid var(--border-color, #e4e6eb)",
              borderRadius: "8px",
              boxShadow: "0 4px 12px rgba(0, 0, 0, 0.15)",
              padding: "8px",
              width: "220px",
              display: "grid",
              gridTemplateColumns: "repeat(6, 1fr)",
              gap: "6px",
              zIndex: 1000
            }}
          >
            {emojis.map((emoji) => (
              <button
                key={emoji}
                type="button"
                className="emoji-item-btn"
                onClick={(e) => handleEmojiClick(emoji, e)}
                style={{
                  background: "none",
                  border: "none",
                  fontSize: "20px",
                  cursor: "pointer",
                  padding: "4px",
                  borderRadius: "4px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  transition: "background 0.2s"
                }}
                onMouseOver={(e) => (e.currentTarget.style.background = "var(--hover-bg, #f2f3f5)")}
                onMouseOut={(e) => (e.currentTarget.style.background = "none")}
              >
                {emoji}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
};

export default EmojiPicker;
