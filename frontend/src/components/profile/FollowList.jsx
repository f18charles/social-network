import { useEffect, useState } from "react";
import { FiX } from "react-icons/fi";
import avatarFallback from "../../assets/user.svg";
import { apiFetch } from "../../utils/api";
import { useAuth } from "../../context/auth/useAuth";
import DMAction from "../chat/DMAction";

/**
 * Modal that lists either the followers or the following of a given user.
 *
 * @param {{
 *   userId: string,
 *   type: "followers" | "following",
 *   onClose: () => void,
 * }} props
 */
const FollowListModal = ({ userId, type, onClose }) => {
  const { currentUser } = useAuth();
  const [people, setPeople] = useState([]);
  const [status, setStatus] = useState("loading");
  const [error, setError] = useState("");

  const title = type === "following" ? "Following" : "Followers";
  const endpoint =
    type === "following"
      ? "/api/followers/following"
      : "/api/followers/followers";

  useEffect(() => {
    if (!userId) return;

    let isActive = true;

    apiFetch(`${endpoint}?user_id=${encodeURIComponent(userId)}`)
      .then((result) => {
        if (!isActive) return;
        setPeople(Array.isArray(result) ? result : result?.data || []);
        setStatus("ready");
      })
      .catch((err) => {
        if (!isActive) return;
        setError(err?.message || `Unable to load ${title.toLowerCase()}.`);
        setStatus("error");
      });

    return () => {
      isActive = false;
    };
  }, [userId, endpoint, title]);

  const handleOverlayClick = (event) => {
    if (event.target === event.currentTarget) {
      onClose();
    }
  };

  return (
    <div className="profile-modal-overlay" onClick={handleOverlayClick}>
      <div className="profile-modal" role="dialog" aria-label={title}>
        <div className="profile-modal__header">
          <h3>{title}</h3>
          <button
            type="button"
            className="profile-modal__close"
            onClick={onClose}
            aria-label="Close"
          >
            <FiX size={18} />
          </button>
        </div>

        <div className="profile-modal__body">
          {status === "loading" && (
            <>
              <div className="profile-skeleton profile-skeleton--row" />
              <div className="profile-skeleton profile-skeleton--row" />
            </>
          )}

          {status === "error" && (
            <div className="profile-state profile-state--error">{error}</div>
          )}

          {status === "ready" && people.length === 0 && (
            <div className="profile-state">Nobody here yet.</div>
          )}

          {status === "ready" && people.length > 0 && (
            <ul className="profile-modal__list">
              {people.map((person) => {
                const name =
                  `${person.first_name || ""} ${person.last_name || ""}`.trim() ||
                  person.nickname ||
                  "Unnamed User";

                return (
                  <li key={person.id} className="profile-modal__list-item">
                    <img
                      src={person.avatar || avatarFallback}
                      alt={`${name}'s avatar`}
                      className="profile-modal__avatar"
                    />
                    <div>
                      <p className="profile-modal__name">{name}</p>
                      {person.nickname && (
                        <p className="profile-modal__handle">
                          @{person.nickname}
                        </p>
                      )}
                    </div>
                    {person.id !== currentUser?.id && (
                      <DMAction
                        userId={person.id}
                        className="profile-btn profile-btn--compact"
                      />
                    )}
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
};

export default FollowListModal;
