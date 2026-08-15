import { Suspense, lazy, useState } from "react";
import { BsEmojiSmile } from "react-icons/bs";

// emoji-picker-react bundles its full emoji dataset; code-split it into its
// own chunk so it's only downloaded when someone actually opens the picker.
const EmojiPickerReact = lazy(() => import("emoji-picker-react"));

/**
 * EmojiPicker renders a trigger button that opens a full searchable,
 * categorized emoji picker (search box, category tabs, skin tone selector,
 * recently-used emojis). Selecting an emoji calls onSelectEmoji(emoji) with
 * the plain unicode character, same as before.
 */
const EmojiPicker = ({ onSelectEmoji }) => {
  const [showPicker, setShowPicker] = useState(false);

  const handleEmojiClick = (emojiData, event) => {
    event?.stopPropagation?.();
    onSelectEmoji?.(emojiData.emoji);
    setShowPicker(false);
  };

  return (
    <div className="emoji-picker-container" style={{ position: "relative", display: "inline-block" }}>
      <button
        type="button"
        className="emoji-picker-trigger"
        onClick={(e) => {
          e.stopPropagation();
          setShowPicker((prev) => !prev);
        }}
        title="Insert Emoji"
        aria-label="Insert Emoji"
        aria-expanded={showPicker}
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
            onClick={(e) => e.stopPropagation()}
            style={{
              position: "absolute",
              bottom: "100%",
              left: 0,
              marginBottom: "8px",
              zIndex: 1000,
              boxShadow: "0 4px 12px rgba(0, 0, 0, 0.15)",
              borderRadius: "8px",
              overflow: "hidden",
              width: 320,
              minHeight: 380,
              background: "var(--bg-primary, #ffffff)"
            }}
          >
            <Suspense
              fallback={
                <div
                  style={{
                    width: 320,
                    height: 380,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    color: "var(--text-secondary, #65676b)",
                    fontSize: 13
                  }}
                >
                  Loading emoji…
                </div>
              }
            >
              <EmojiPickerReact
                onEmojiClick={handleEmojiClick}
                autoFocusSearch
                lazyLoadEmojis
                width={320}
                height={380}
                previewConfig={{ showPreview: false }}
                searchPlaceHolder="Search emoji"
              />
            </Suspense>
          </div>
        </>
      )}
    </div>
  );
};

export default EmojiPicker;
