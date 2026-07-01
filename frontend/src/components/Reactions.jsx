import { BiSolidDislike, BiSolidLike } from "react-icons/bi";

const Like = ({ like, isActive, size = 24, className = "" }) => {
  if (!like) {
    return <BiSolidLike aria-hidden="true" size={size} />;
  }

  const handleClick = (event) => {
    event.stopPropagation();
    like?.();
  };

  return (
    <button
      type="button"
      aria-label="Like"
      aria-pressed={Boolean(isActive)}
      className={`reaction-button ${isActive ? "reaction-like" : ""} ${className}`.trim()}
      onClick={handleClick}
    >
      <BiSolidLike aria-hidden="true" size={size} />
    </button>
  );
};

const Dislike = ({ dislike, isActive, size = 24, className = "" }) => {
  if (!dislike) {
    return <BiSolidDislike aria-hidden="true" size={size} />;
  }

  const handleClick = (event) => {
    event.stopPropagation();
    dislike?.();
  };

  return (
    <button
      type="button"
      aria-label="Dislike"
      aria-pressed={Boolean(isActive)}
      className={`reaction-button ${isActive ? "reaction-dislike" : ""} ${className}`.trim()}
      onClick={handleClick}
    >
      <BiSolidDislike aria-hidden="true" size={size} />
    </button>
  );
};

export { Like, Dislike };
