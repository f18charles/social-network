/**
 * Row of clickable stats shown under the profile header.
 *
 * @param {{
 *   postCount: number,
 *   followerCount: number,
 *   followingCount: number,
 *   onShowFollowers?: () => void,
 *   onShowFollowing?: () => void,
 * }} props
 */
const ProfileStats = ({
  postCount = 0,
  followerCount = 0,
  followingCount = 0,
  onShowFollowers,
  onShowFollowing,
}) => {
  return (
    <div className="profile-stats">
      <button type="button" className="profile-stats__item" disabled>
        <span className="profile-stats__count">{postCount}</span>
        <span className="profile-stats__label">Posts</span>
      </button>
      <button
        type="button"
        className="profile-stats__item"
        onClick={onShowFollowers}
      >
        <span className="profile-stats__count">{followerCount}</span>
        <span className="profile-stats__label">Followers</span>
      </button>
      <button
        type="button"
        className="profile-stats__item"
        onClick={onShowFollowing}
      >
        <span className="profile-stats__count">{followingCount}</span>
        <span className="profile-stats__label">Following</span>
      </button>
    </div>
  );
};

export default ProfileStats;
