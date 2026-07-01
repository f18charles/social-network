import { useState } from "react";
import { apiFetch } from "../../utils/api";
import "../../styles/follow-action.css";

const FollowAction = ({
  targetUserId,
  initialStatus = "unfollowed",
  isPrivate = false,
  onStatusChange,
}) => {
  const [status, setStatus] = useState(initialStatus); // "unfollowed" | "following" | "requested"
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleClick = async () => {
    setLoading(true);
    setError("");

    try {
      if (status === "unfollowed") {
        // Send follow request
        const body = { following_id: targetUserId };
        await apiFetch("/api/followers/follow", {
          method: "POST",
          body,
        });
        const newStatus = isPrivate ? "requested" : "following";
        setStatus(newStatus);
        onStatusChange?.(newStatus);
      } else if (status === "following") {
        // Unfollow
        await apiFetch("/api/followers/unfollow", {
          method: "POST",
          body: { following_id: targetUserId },
        });
        setStatus("unfollowed");
        onStatusChange?.("unfollowed");
      } else if (status === "requested") {
        // Cancel follow request
        await apiFetch("/api/followers/cancel", {
          method: "POST",
          body: { following_id: targetUserId },
        });
        setStatus("unfollowed");
        onStatusChange?.("unfollowed");
      }
    } catch (err) {
      setError(err.message || "Failed to update follow status");
    } finally {
      setLoading(false);
    }
  };

  const getButtonConfig = () => {
    switch (status) {
      case "following":
        return {
          label: loading ? "..." : "Following",
          className: "follow-action--following",
          hoverLabel: "Unfollow",
        };
      case "requested":
        return {
          label: loading ? "..." : "Requested",
          className: "follow-action--requested",
          hoverLabel: "Cancel Request",
        };
      default:
        return {
          label: loading ? "..." : "Follow",
          className: "follow-action--follow",
          hoverLabel: null,
        };
    }
  };

  const config = getButtonConfig();

  return (
    <div className="follow-action">
      <button
        className={`follow-action__btn ${config.className}`}
        onClick={handleClick}
        disabled={loading}
        data-hover-label={config.hoverLabel}
      >
        {loading ? "..." : config.label}
      </button>
      {error && <div className="follow-action__error">{error}</div>}
    </div>
  );
};

export default FollowAction;
