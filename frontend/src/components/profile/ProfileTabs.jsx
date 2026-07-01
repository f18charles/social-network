/**
 * Simple two/three-way tab switcher used on the profile page.
 *
 * @param {{
 *   tabs: { id: string, label: string }[],
 *   activeTab: string,
 *   onChange: (id: string) => void,
 * }} props
 */
const ProfileTabs = ({ tabs, activeTab, onChange }) => {
  return (
    <div className="profile-tabs" role="tablist">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          type="button"
          role="tab"
          aria-selected={activeTab === tab.id}
          className={`profile-tabs__btn ${
            activeTab === tab.id ? "is-active" : ""
          }`}
          onClick={() => onChange(tab.id)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
};

export default ProfileTabs;
