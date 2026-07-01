import { useState } from "react";
import { apiFetch } from "../../utils/api";
import avatarFallback from "../../assets/user.svg";
import "../../styles/follow-request-card.css";

const toDisplayName = (user) =>
  `${user.first_name || ""} ${user.last_name || ""}`.trim() ||
  user.nickname ||
  "Unnamed User";

const FollowRequestCard = ({ request, onAccept, onReject }) => {
  const [loading, setLoading] = useState(false);

  const handleAccept = async () => {
    setLoading(true);
    try {
      await apiFetch("/api/followers/accept", {
        method: "POST",
        body: { follower_id: request.id },
      });
      onAccept?.(request.id);
    } catch (err) {
      console.error("Failed to accept request:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleReject = async () => {
    setLoading(true);
    try {
      await apiFetch("/api/followers/reject", {
        method: "POST",
        body: { follower_id: request.id },
      });
      onReject?.(request.id);
    } catch (err) {
      console.error("Failed to reject request:", err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="follow-request-card">
      <div className="follow-request-card__user">
        <img
          src={request.avatar || avatarFallback}
          alt={`${toDisplayName(request)}'s avatar`}
          className="follow-request-card__avatar"
        />
        <div>
          <div className="follow-request-card__name">
            {toDisplayName(request)}
          </div>
          {request.nickname && (
            <div className="follow-request-card__nickname">
              @{request.nickname}
            </div>
          )}
        </div>
      </div>
      <div className="follow-request-card__actions">
        <button
          className="follow-request-card__accept"
          onClick={handleAccept}
          disabled={loading}
        >
          Accept
        </button>
        <button
          className="follow-request-card__reject"
          onClick={handleReject}
          disabled={loading}
        >
          Reject
        </button>
      </div>
    </div>
  );
};

export default FollowRequestCard;
