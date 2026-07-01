import { BiSolidDislike, BiSolidLike } from "react-icons/bi";

/**
 * VoteControls renders like/dislike counts and delegates vote changes to callers.
 *
 * @param {{
 *   likes?: number,
 *   dislikes?: number,
 *   currentVote?: "like" | "dislike" | "none",
 *   targetType?: string,
 *   isDisabled?: boolean,
 *   isMutating?: boolean,
 *   onVote?: (vote: "like" | "dislike") => void,
 * }} props
 */
export const VoteControls = ({
  likes = 0,
  dislikes = 0,
  currentVote = "none",
  targetType = "item",
  isDisabled = false,
  isMutating = false,
  onVote = () => {},
}) => {
  const disabled = isDisabled || isMutating;

  const handleVote = (event, vote) => {
    event.stopPropagation();
    onVote(vote);
  };

  return (
    <div className="flex items-center space-x-2 vote-controls">
      <button
        type="button"
        onClick={(event) => handleVote(event, "like")}
        disabled={disabled}
        aria-pressed={currentVote === "like"}
        aria-label={`Like this ${targetType}. Current likes: ${likes}`}
        className={`flex items-center p-2 rounded transition-colors ${
          currentVote === "like"
            ? "text-blue-600 bg-blue-50"
            : "text-gray-500 hover:bg-gray-100"
        } disabled:opacity-50`}
      >
        <BiSolidLike aria-hidden="true" size={24} className="mr-1" />
        <span className="text-sm font-medium">{likes}</span>
      </button>

      <button
        type="button"
        onClick={(event) => handleVote(event, "dislike")}
        disabled={disabled}
        aria-pressed={currentVote === "dislike"}
        aria-label={`Dislike this ${targetType}. Current dislikes: ${dislikes}`}
        className={`flex items-center p-2 rounded transition-colors ${
          currentVote === "dislike"
            ? "text-red-600 bg-red-50"
            : "text-gray-500 hover:bg-gray-100"
        } disabled:opacity-50`}
      >
        <BiSolidDislike aria-hidden="true" size={24} className="mr-1" />
        <span className="text-sm font-medium">{dislikes}</span>
      </button>
    </div>
  );
};
