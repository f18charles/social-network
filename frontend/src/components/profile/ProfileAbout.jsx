import { FiMail, FiCalendar, FiClock, FiFileText } from "react-icons/fi";

/**
 * Read-only "About" card: email, date of birth, member-since date and bio.
 *
 * @param {{ user: object, isOwnProfile: boolean }} props
 */
const ProfileAbout = ({ user, isOwnProfile }) => {
  const formatDate = (value) => {
    if (!value) return null;
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return null;
    return date.toLocaleDateString(undefined, {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const dob = formatDate(user?.date_of_birth || user?.dob);
  const joined = formatDate(user?.created_at);

  return (
    <div className="profile-about">
      <h2>About</h2>

      {isOwnProfile && user?.email && (
        <div className="profile-about__row">
          <FiMail className="profile-about__icon" size={16} />
          <span className="profile-about__label">Email</span>
          <span className="profile-about__value">{user.email}</span>
        </div>
      )}

      {dob && (
        <div className="profile-about__row">
          <FiCalendar className="profile-about__icon" size={16} />
          <span className="profile-about__label">Born</span>
          <span className="profile-about__value">{dob}</span>
        </div>
      )}

      {joined && (
        <div className="profile-about__row">
          <FiClock className="profile-about__icon" size={16} />
          <span className="profile-about__label">Joined</span>
          <span className="profile-about__value">{joined}</span>
        </div>
      )}

      <div className="profile-about__row">
        <FiFileText className="profile-about__icon" size={16} />
        <span className="profile-about__label">Bio</span>
        <span className="profile-about__value">
          {user?.about_me || "No bio yet."}
        </span>
      </div>
    </div>
  );
};

export default ProfileAbout;
