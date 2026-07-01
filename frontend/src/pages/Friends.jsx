import { useCallback, useEffect, useState } from "react";
import { useAuth } from "../context/useAuth";
import { apiFetch, ApiError } from "../utils/api";
import FollowRequestsList from "../components/follow/FollowRequestsList";
import UserCard from "../components/user/UserCard";
import FollowAction from "../components/follow/FollowAction";
import "../styles/friends.css";

/**
 * Friends page - displays pending follow requests and discoverable users.
 * Uses modern components (FollowRequestsList, UserCard, FollowAction) with
 * fallback to legacy implementation for compatibility.
 */
const Friends = () => {
  const { currentUser } = useAuth();
  const [query, setQuery] = useState("");
  const [actionMessage, setActionMessage] = useState("");
  const [error, setError] = useState("");
  const [requestCount, setRequestCount] = useState(0);
  const [mutuals, setMutuals] = useState([]);
  const [notFollowedBack, setNotFollowedBack] = useState([]);
  const [strangers, setStrangers] = useState([]);

  const loadPage = useCallback(async () => {
    setError("");

    try {
      const [pending, followers, following, suggestions] = await Promise.all([
        apiFetch("/api/followers/pending"),
        apiFetch(`/api/followers/followers?user_id=${currentUser.id}`),
        apiFetch(`/api/followers/following?user_id=${currentUser.id}`),
        apiFetch("/api/users/search"),
      ]);

      const followingIds = new Set((following || []).map((u) => u.id));
      const followerIds = new Set((followers || []).map((u) => u.id));

      // People who follow you AND you follow back
      setMutuals((followers || []).filter((u) => followingIds.has(u.id)));

      // People who follow you but you haven't followed back
      setNotFollowedBack(
        (followers || []).filter((u) => !followingIds.has(u.id))
      );

      // People from search who don't follow you and you don't follow them
      setStrangers(
        (suggestions || []).filter(
          (u) =>
            !followerIds.has(u.id) &&
            !followingIds.has(u.id) &&
            u.id !== currentUser.id
        )
      );
      setRequestCount(
        (Array.isArray(pending) ? pending : pending || []).length
      );
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : "Unable to load friends data."
      );
    }
  }, [currentUser]);

  useEffect(() => {
    if (!currentUser) return undefined;
    const timer = window.setTimeout(() => {
      void loadPage();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [currentUser, loadPage]);

  // Load/search suggestions
  const loadSuggestions = async (searchTerm = "") => {
    setError("");
    try {
      const results = await apiFetch(
        `/api/users/search${searchTerm ? `?query=${encodeURIComponent(searchTerm)}` : ""}`
      );
      setStrangers(Array.isArray(results) ? results : results || []);
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : "Unable to load suggestions."
      );
    }
  };

  const handleSearchSubmit = async (event) => {
    event.preventDefault();
    await loadSuggestions(query);
  };

  // --- New Handler (for FollowAction component) ---
  const handleFollowStatusChange = async (userId, newStatus) => {
    setActionMessage(
      newStatus === "following"
        ? "Followed successfully!"
        : newStatus === "unfollowed"
          ? "Unfollowed successfully!"
          : "Follow request sent!"
    );
    await loadPage();
  };

  if (!currentUser) {
    return <div className="friends-page">Loading friends...</div>;
  }

  return (
    <div className="friends-page">
      <div className="friends-header">
        <h1>Friends</h1>
        <p>
          Find public profiles, send friend requests, and accept pending follow
          requests.
        </p>
      </div>

      {error && (
        <div className="profile-state profile-state--error">{error}</div>
      )}
      {actionMessage && <div className="profile-state">{actionMessage}</div>}

      {/* Pending Requests Section */}
      <section className="friends-section">
        <div className="friends-section__header">
          <h2>
            Pending requests{" "}
            {requestCount > 0 && (
              <span className="friends-badge">{requestCount}</span>
            )}
          </h2>
        </div>

        <FollowRequestsList onRequestCountChange={setRequestCount} />
      </section>

      {/* Discover Profiles Section */}
      <section className="friends-section">
        <div className="friends-section__header friends-section__header--search">
          <h2>Discover other profiles</h2>
          <form className="friends-search" onSubmit={handleSearchSubmit}>
            <input
              type="search"
              value={query}
              placeholder="Search by username, first name, or last name"
              onChange={(event) => setQuery(event.target.value)}
              className="friends-search__input"
            />
            <button type="submit" className="profile-btn profile-btn--primary">
              Search
            </button>
          </form>
        </div>

        <div className="friends-list">
          {strangers.map((user) => (
            <UserCard
              key={user.id}
              user={user}
              actions={
                <FollowAction
                  targetUserId={user.id}
                  initialStatus="unfollowed"
                  isPrivate={!user.is_public}
                  onStatusChange={(status) =>
                    handleFollowStatusChange(user.id, status)
                  }
                />
              }
            />
          ))}
        </div>
      </section>

      {/* mutuals friends list */}
      <section className="friends-section">
        <h2>
          Friends <span className="friends-badge">{mutuals.length}</span>
        </h2>
        {mutuals.length === 0 ? (
          <div className="profile-state">No mutual follows yet.</div>
        ) : (
          <div className="friends-list">
            {mutuals.map((user) => (
              <UserCard
                key={user.id}
                user={user}
                actions={
                  <FollowAction
                    targetUserId={user.id}
                    initialStatus="following"
                    isPrivate={!user.is_public}
                    onStatusChange={(status) =>
                      handleFollowStatusChange(user.id, status)
                    }
                  />
                }
              />
            ))}
          </div>
        )}
      </section>

      {/* not followed back profiles list */}
      <section className="friends-section">
        <h2>
          Follow back{" "}
          <span className="friends-badge">{notFollowedBack.length}</span>
        </h2>
        {notFollowedBack.length === 0 ? (
          <div className="profile-state">
            You follow everyone who follows you.
          </div>
        ) : (
          <div className="friends-list">
            {notFollowedBack.map((user) => (
              <UserCard
                key={user.id}
                user={user}
                actions={
                  <FollowAction
                    targetUserId={user.id}
                    initialStatus="unfollowed"
                    isPrivate={!user.is_public}
                    onStatusChange={(status) =>
                      handleFollowStatusChange(user.id, status)
                    }
                  />
                }
              />
            ))}
          </div>
        )}
      </section>
    </div>
  );
};

export default Friends;
