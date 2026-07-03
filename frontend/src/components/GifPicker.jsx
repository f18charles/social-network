import React, { useState } from "react";
import { MdGif } from "react-icons/md";

const GifPicker = ({ onSelectGif }) => {
  const [showPicker, setShowPicker] = useState(false);

  // Curated static set of popular GIFs using public, reliable URLs
  const staticGifs = [
    { name: "Thumbs Up", url: "https://media.giphy.com/media/tIeCLkB8geYtW/giphy.gif" },
    { name: "Laughing", url: "https://media.giphy.com/media/26n6Gx9moCgs1pUuk/giphy.gif" },
    { name: "Celebration", url: "https://media.giphy.com/media/kyLYXonQpkUsCxZIKa/giphy.gif" },
    { name: "Surprised", url: "https://media.giphy.com/media/l0HlO3BJ8LALPW4sE/giphy.gif" },
    { name: "Sad", url: "https://media.giphy.com/media/9Y5BbDSkSTiY8/giphy.gif" },
    { name: "Wow", url: "https://media.giphy.com/media/3o7527pa7qs9kCG78A/giphy.gif" },
    { name: "Dancing", url: "https://media.giphy.com/media/13CoXDiaCcC9R6/giphy.gif" },
    { name: "Cat Sleep", url: "https://media.giphy.com/media/12PA1eI8FBqEUM/giphy.gif" }
  ];

  const handleGifClick = (gifUrl, event) => {
    event.stopPropagation();
    onSelectGif?.(gifUrl);
    setShowPicker(false);
  };

  return (
    <div className="gif-picker-container" style={{ position: "relative", display: "inline-block" }}>
      <button
        type="button"
        className="gif-picker-trigger"
        onClick={(e) => {
          e.stopPropagation();
          setShowPicker(!showPicker);
        }}
        title="Insert GIF"
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
        <MdGif size={28} style={{ margin: "-4px 0" }} />
      </button>
      {showPicker && (
        <>
          <div
            className="gif-picker-overlay"
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
            className="gif-picker-popover"
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
              width: "280px",
              maxHeight: "300px",
              overflowY: "auto",
              display: "grid",
              gridTemplateColumns: "repeat(2, 1fr)",
              gap: "8px",
              zIndex: 1000
            }}
          >
            {staticGifs.map((gif) => (
              <button
                key={gif.url}
                type="button"
                className="gif-item-btn"
                onClick={(e) => handleGifClick(gif.url, e)}
                style={{
                  background: "none",
                  border: "none",
                  cursor: "pointer",
                  padding: "0",
                  borderRadius: "4px",
                  overflow: "hidden",
                  width: "100%",
                  height: "90px",
                  display: "block"
                }}
              >
                <img
                  src={gif.url}
                  alt={gif.name}
                  style={{
                    width: "100%",
                    height: "100%",
                    objectFit: "cover",
                    borderRadius: "4px",
                    display: "block"
                  }}
                />
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
};

export default GifPicker;
