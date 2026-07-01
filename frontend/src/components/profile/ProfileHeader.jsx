import { useRef } from "react";
import { FiCamera, FiLock, FiGlobe } from "react-icons/fi";
import avatarFallback from "../../assets/user.svg";

/**
 * Displays the cover, avatar, name, privacy badge, bio and the primary
 * action (Edit profile / Follow / Unfollow) for a profile.
 *
 * @param {{
 *   user: object,
 *   isOwnProfile: boolean,
 *   isFollowing?: boolean,
 *   followLoading?: boolean,
 *   onEdit?: () => void,
 *   onAvatarChange?: (file: File) => void,
 *   onFollow?: () => void,
 *   onUnfollow?: () => void,
 * }} props
 */
const ProfileHeader = ({
  user,
  isOwnProfile,
  isFollowing = false,
  followLoading = false,
  onEdit,
  onAvatarChange,
  onFollow,
  onUnfollow,
}) => {
  const fileInputRef = useRef(null);

  const displayName =
    `${user?.first_name || ""} ${user?.last_name || ""}`.trim() ||
    user?.nickname ||
    "Unnamed User";

  const handleAvatarClick = () => {
    if (isOwnProfile && fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  const handleFileSelected = (event) => {
    const file = event.target.files?.[0];
    if (file && onAvatarChange) {
      onAvatarChange(file);
    }
  };

  return (
    <div className="profile-header">
      <div className="profile-header__cover" />

      <div className="profile-header__body">
        <div className="profile-header__avatar-wrap">
          <img
            src={user?.avatar || avatarFallback}
            alt={`${displayName}'s avatar`}
            className="profile-header__avatar"
          />
          {isOwnProfile && (
            <>
              <button
                type="button"
                className="profile-header__avatar-edit"
                onClick={handleAvatarClick}
                aria-label="Change avatar"
              >
                <FiCamera size={14} />
              </button>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleFileSelected}
                style={{ display: "none" }}
              />
            </>
          )}
        </div>

        <div className="profile-header__meta">
          <div className="profile-header__names">
            <h1>{displayName}</h1>
            {user?.nickname && (
              <p className="profile-header__nickname">@{user.nickname}</p>
            )}
            <div className="profile-header__badges">
              <span
                className={`profile-badge ${user?.is_public ? "" : "is-private"}`}
              >
                {user?.is_public ? (
                  <>
                    <FiGlobe size={12} /> Public
                  </>
                ) : (
                  <>
                    <FiLock size={12} /> Private
                  </>
                )}
              </span>
            </div>
          </div>

          <div className="profile-header__actions">
            {isOwnProfile ? (
              <button
                type="button"
                className="profile-btn profile-btn--primary"
                onClick={onEdit}
              >
                Edit profile
              </button>
            ) : isFollowing ? (
              <button
                type="button"
                className="profile-btn profile-btn--following"
                onClick={onUnfollow}
                disabled={followLoading}
              >
                Following
              </button>
            ) : (
              <button
                type="button"
                className="profile-btn profile-btn--primary"
                onClick={onFollow}
                disabled={followLoading}
              >
                Follow
              </button>
            )}
          </div>
        </div>
      </div>

      <p
        className={`profile-header__bio ${
          user?.about_me ? "" : "profile-header__bio--empty"
        }`}
      >
        {user?.about_me || "This user hasn't written a bio yet."}
      </p>
    </div>
  );
};

export default ProfileHeader;
