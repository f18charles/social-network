import { useState } from "react";
import { useNavigate } from "react-router";
import { FiMessageCircle } from "react-icons/fi";
import { apiFetch } from "../../utils/api";

/**
 * Button that opens or creates a direct-message conversation for an eligible user.
 * The backend enforces the accepted follow relationship required to message.
 */
const DMAction = ({ userId, disabled = false, className = "profile-btn" }) => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const handleClick = async (event) => {
    event.stopPropagation();
    if (!userId || disabled || loading) return;

    setLoading(true);
    try {
      const conversation = await apiFetch("/api/chats/dm", {
        method: "POST",
        body: { recipient_id: userId },
      });
      navigate("/messages", { state: { selectedConversation: conversation } });
    } catch (err) {
      alert(err.message || "Unable to open direct message.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <button
      type="button"
      className={className}
      onClick={handleClick}
      disabled={disabled || loading}
      title="Message"
      aria-label="Message user"
    >
      <FiMessageCircle size={16} />
      <span>{loading ? "Opening" : "Message"}</span>
    </button>
  );
};

export default DMAction;
