import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router";
import { useAuth } from "../context/useAuth";
import { apiFetch, ApiError } from "../utils/api";
import ProfileHeader from "../components/profile/ProfileHeader";
import ProfileStats from "../components/profile/ProfileStats";
import ProfileTabs from "../components/profile/ProfileTabs";
import ProfileAbout from "../components/profile/ProfileAbout";
import ProfilePosts from "../components/profile/ProfilePosts";
import FollowListModal from "../components/profile/FollowList";
import ProfileUpdateForm from "../components/ProfileUpdateForm";
import FollowAction from "../components/follow/FollowAction";
import "../styles/profile.css";
import avatarFallback from "../assets/user.svg"

const TABS = [
  { id: "posts", label: "Posts" },
  { id: "about", label: "About" },
];

const Profile = () => {
  const { userId } = useParams();
  const navigate = useNavigate();
  const { currentUser, refresh } = useAuth();
  
  // Determine if we're viewing own profile
  const isOwnProfile = !userId || userId === currentUser?.id;
  
  // State for own profile
  const [activeTab, setActiveTab] = useState("posts");
  const [isEditing, setIsEditing] = useState(false);
  const [followModal, setFollowModal] = useState(null);
  const [avatarError, setAvatarError] = useState("");
  
  // State for other user profile
  const [user, setUser] = useState(null);
  const [followStatus, setFollowStatus] = useState("unfollowed");
  
  // Shared state
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [followerCount, setFollowerCount] = useState(0);
  const [followingCount, setFollowingCount] = useState(0);
  const [postCount, setPostCount] = useState(0);

  // If viewing own profile but /user/:id route was used, redirect to /profile
  useEffect(() => {
    if (userId && currentUser && userId === currentUser.id) {
      navigate("/profile", { replace: true });
    }
  }, [userId, currentUser, navigate]);

  // Load own profile data
  const loadOwnProfile = async () => {
    if (!currentUser) return;
    
    try {
      const [followers, following, posts] = await Promise.all([
        apiFetch(`/api/followers/followers?user_id=${currentUser.id}`),
        apiFetch(`/api/followers/following?user_id=${currentUser.id}`),
        apiFetch(`/api/users/${currentUser.id}/posts`),
      ]);
      
      setFollowerCount(
        Array.isArray(followers) ? followers.length : (followers?.data || []).length
      );
      setFollowingCount(
        Array.isArray(following) ? following.length : (following?.data || []).length
      );
      const list = Array.isArray(posts) ? posts : posts?.data || [];
      setPostCount(list.length);
    } catch {
      setFollowerCount(0);
      setFollowingCount(0);
      setPostCount(0);
    }
  };

  // Load other user profile data
  const loadUserProfile = async (targetUserId) => {
    try {
      const [userData, postsData, followersData, followingData] = await Promise.all([
        apiFetch(`/api/users/${targetUserId}`),
        apiFetch(`/api/users/${targetUserId}/posts`),
        apiFetch(`/api/followers/followers?user_id=${targetUserId}`),
        apiFetch(`/api/followers/following?user_id=${targetUserId}`),
      ]);

      setUser(userData);
      setPostCount(
        Array.isArray(postsData) ? postsData.length : postsData?.data?.length || 0
      );
      setFollowerCount(
        Array.isArray(followersData) ? followersData.length : followersData?.data?.length || 0
      );
      setFollowingCount(
        Array.isArray(followingData) ? followingData.length : followingData?.data?.length || 0
      );

      // Determine follow status
      if (userData?.is_following) {
        setFollowStatus("following");
      } else if (userData?.follow_request_pending) {
        setFollowStatus("requested");
      } else {
        setFollowStatus("unfollowed");
      }
    } catch (err) {
      throw err;
    }
  };

  // Main data loading effect
  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      setError("");

      try {
        if (isOwnProfile) {
          await loadOwnProfile();
        } else if (userId) {
          await loadUserProfile(userId);
        } else {
          // No userId and not own profile? Shouldn't happen, but fallback to own
          await loadOwnProfile();
        }
      } catch (err) {
        setError(
          err instanceof ApiError ? err.message : "Failed to load profile data"
        );
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [userId, currentUser, isOwnProfile]);

  // Handle avatar change (own profile only)
  const handleAvatarChange = async (file) => {
    setAvatarError("");
    const data = new FormData();
    data.append("avatar", file);

    try {
      await apiFetch("/api/users/update", { method: "PATCH", body: data });
      await refresh();
      await loadOwnProfile();
    } catch (err) {
      setAvatarError(
        err instanceof ApiError ? err.message : "Unable to update avatar."
      );
    }
  };

  // Handle follow status change (other user profile only)
  const handleFollowStatusChange = (newStatus) => {
    setFollowStatus(newStatus);
    if (newStatus === "following") {
      setFollowerCount((prev) => prev + 1);
    } else if (newStatus === "unfollowed") {
      setFollowerCount((prev) => Math.max(0, prev - 1));
    }
  };

  // Loading state
  if (loading) {
    return (
      <div className="profile-page">
        <div className="profile-page__inner">
          <div className="profile-skeleton profile-skeleton--header" />
          <div className="profile-skeleton profile-skeleton--row" />
          <div className="profile-skeleton profile-skeleton--row" />
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="profile-page">
        <div className="profile-page__inner">
          <div className="profile-state profile-state--error">{error}</div>
        </div>
      </div>
    );
  }

  // For other user profiles, if user not found
  if (!isOwnProfile && !user) {
    return (
      <div className="profile-page">
        <div className="profile-page__inner">
          <div className="profile-state profile-state--error">User not found</div>
        </div>
      </div>
    );
  }

  // Get the user data to display (either currentUser or the fetched user)
  const displayUser = isOwnProfile ? currentUser : user;
  if (!displayUser) {
    return (
      <div className="profile-page">
        <div className="profile-page__inner">
          <div className="profile-skeleton profile-skeleton--header" />
        </div>
      </div>
    );
  }

  // Render the unified profile
  return (
    <div className="profile-page">
      <div className="profile-page__inner">
        {/* Profile Header */}
        {isOwnProfile ? (
          <ProfileHeader
            user={displayUser}
            isOwnProfile={true}
            onEdit={() => setIsEditing(true)}
            onAvatarChange={handleAvatarChange}
          />
        ) : (
          <div className="profile-header">
            <div className="profile-header__cover" />
            <div className="profile-header__body">
              <div className="profile-header__avatar-wrap">
                <img
                  src={displayUser.avatar || avatarFallback}
                  alt={`${displayUser.first_name}'s avatar`}
                  className="profile-header__avatar"
                />
              </div>
              <div className="profile-header__meta">
                <div className="profile-header__names">
                  <h1>
                    {`${displayUser.first_name || ""} ${displayUser.last_name || ""}`.trim() ||
                      displayUser.nickname ||
                      "Unnamed User"}
                  </h1>
                  {displayUser.nickname && (
                    <p className="profile-header__nickname">@{displayUser.nickname}</p>
                  )}
                  <div className="profile-header__badges">
                    <span
                      className={`profile-badge ${displayUser.is_public ? "" : "is-private"}`}
                    >
                      {displayUser.is_public ? "Public" : "Private"}
                    </span>
                  </div>
                </div>
                <div className="profile-header__actions">
                  <FollowAction
                    targetUserId={displayUser.id}
                    initialStatus={followStatus}
                    isPrivate={!displayUser.is_public}
                    onStatusChange={handleFollowStatusChange}
                  />
                </div>
              </div>
            </div>
            <p
              className={`profile-header__bio ${displayUser.about_me ? "" : "profile-header__bio--empty"}`}
            >
              {displayUser.about_me || "This user hasn't written a bio yet."}
            </p>
          </div>
        )}

        {/* Avatar error message (own profile only) */}
        {isOwnProfile && avatarError && (
          <div className="profile-state profile-state--error">{avatarError}</div>
        )}

        {/* Stats */}
        <div className="profile-stats-card">
          {isOwnProfile ? (
            <ProfileStats
              postCount={postCount}
              followerCount={followerCount}
              followingCount={followingCount}
              onShowFollowers={() => setFollowModal("followers")}
              onShowFollowing={() => setFollowModal("following")}
            />
          ) : (
            <div className="profile-stats">
              <div className="profile-stats__item">
                <span className="profile-stats__count">{postCount}</span>
                <span className="profile-stats__label">Posts</span>
              </div>
              <button
                type="button"
                className="profile-stats__item"
                onClick={() => setFollowModal("followers")}
              >
                <span className="profile-stats__count">{followerCount}</span>
                <span className="profile-stats__label">Followers</span>
              </button>
              <button
                type="button"
                className="profile-stats__item"
                onClick={() => setFollowModal("following")}
              >
                <span className="profile-stats__count">{followingCount}</span>
                <span className="profile-stats__label">Following</span>
              </button>
            </div>
          )}
        </div>

        {/* Tabs & Content (only for own profile) */}
        {isOwnProfile ? (
          <>
            <ProfileTabs tabs={TABS} activeTab={activeTab} onChange={setActiveTab} />
            
            {activeTab === "posts" ? (
              <ProfilePosts userId={displayUser.id} />
            ) : (
              <ProfileAbout user={displayUser} isOwnProfile={true} />
            )}
          </>
        ) : (
          // For other user profiles, show their posts directly
          <div className="profile-posts">
            {postCount === 0 ? (
              <div className="profile-state">This user hasn't posted anything yet.</div>
            ) : (
              <ProfilePosts userId={displayUser.id} />
            )}
          </div>
        )}
      </div>

      {/* Modals */}
      {isEditing && (
        <div className="profile-edit-overlay">
          <ProfileUpdateForm onClose={() => setIsEditing(false)} />
        </div>
      )}

      {followModal && (
        <FollowListModal
          userId={displayUser.id}
          type={followModal}
          onClose={() => setFollowModal(null)}
        />
      )}
    </div>
  );
};

export default Profile;