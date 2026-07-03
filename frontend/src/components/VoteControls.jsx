import { BiSolidDislike, BiSolidLike, BiSolidHeart, BiHeart } from "react-icons/bi";

/**
 * VoteControls renders like/dislike counts and delegates vote changes to callers.
 *
 * @param {{
 *   likes?: number,
 *   dislikes?: number,
 *   hearts?: number,
 *   currentVote?: "like" | "dislike" | "love" | "none",
 *   targetType?: string,
 *   isDisabled?: boolean,
 *   isMutating?: boolean,
 *   onVote?: (vote: "like" | "dislike" | "love") => void,
 *   showHeart?: boolean,
 * }} props
 */
export const VoteControls = ({
  likes = 0,
  dislikes = 0,
  hearts = 0,
  currentVote = "none",
  targetType = "item",
  isDisabled = false,
  isMutating = false,
  onVote = () => {},
  showHeart = false,
}) => {
  const disabled = isDisabled || isMutating;

  const handleVote = (event, vote) => {
    event.stopPropagation();
    onVote(vote);
  };

  return (
    <div className="vote-controls">
      <button
        type="button"
        onClick={(event) => handleVote(event, "like")}
        disabled={disabled}
        aria-pressed={currentVote === "like"}
        aria-label={`Like this ${targetType}. Current likes: ${likes}`}
        className="vote-controls__button vote-controls__button--like"
      >
        <BiSolidLike aria-hidden="true" size={20} />
        <span>{likes}</span>
      </button>

      <button
        type="button"
        onClick={(event) => handleVote(event, "dislike")}
        disabled={disabled}
        aria-pressed={currentVote === "dislike"}
        aria-label={`Dislike this ${targetType}. Current dislikes: ${dislikes}`}
        className="vote-controls__button vote-controls__button--dislike"
      >
        <BiSolidDislike aria-hidden="true" size={20} />
        <span>{dislikes}</span>
      </button>

      {showHeart && (
        <button
          type="button"
          onClick={(event) => handleVote(event, "love")}
          disabled={disabled}
          aria-pressed={currentVote === "love"}
          aria-label={`Love this ${targetType}. Current hearts: ${hearts}`}
          className="vote-controls__button vote-controls__button--love"
        >
          {currentVote === "love" ? (
            <BiSolidHeart aria-hidden="true" size={20} style={{ color: "#ff4b4b" }} />
          ) : (
            <BiHeart aria-hidden="true" size={20} />
          )}
          <span>{hearts}</span>
        </button>
      )}
    </div>
  );
};

