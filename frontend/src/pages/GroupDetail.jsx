import { useEffect, useState, useCallback } from "react";
import { Link, useParams, useNavigate } from "react-router";
import NewPost from "../components/NewPost.jsx";
import Post from "../components/Post.jsx";
import { apiFetch } from "../utils/api.js";
import { useAuth } from "../context/auth/useAuth.js";
import avatarFallback from "../assets/user.svg";
import "../styles/home.css";
import "../styles/group_detail.css";

const getListData = (response) =>
  Array.isArray(response) ? response : response?.data || [];

/**
 * GroupDetail renders the accepted-member post feed for a single group,
 * along with membership actions, member roles list, and invite options.
 */
export default function GroupDetail() {
  const { groupId } = useParams();
  const navigate = useNavigate();
  const { currentUser } = useAuth();
  const [group, setGroup] = useState(null);
  const [posts, setPosts] = useState([]);
  const [members, setMembers] = useState([]);
  const [invitableFriends, setInvitableFriends] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);

  const fetchMembers = useCallback(async () => {
    try {
      const data = await apiFetch(`/api/groups/${groupId}/members`);
      setMembers(data || []);
    } catch (err) {
      console.error("Failed to fetch members", err);
    }
  }, [groupId]);

  const fetchInvitableFriends = useCallback(async () => {
    if (!currentUser?.id) return;
    try {
      const [followers, following] = await Promise.all([
        apiFetch(`/api/followers/followers?user_id=${currentUser.id}`),
        apiFetch(`/api/followers/following?user_id=${currentUser.id}`),
      ]);

      const followersList = getListData(followers);
      const followingList = getListData(following);

      const uniqueFriends = {};
      [...followersList, ...followingList].forEach((u) => {
        if (u.id !== currentUser.id) {
          uniqueFriends[u.id] = u;
        }
      });

      const memberIds = new Set(members.map((m) => m.id));
      const filtered = Object.values(uniqueFriends).filter((f) => !memberIds.has(f.id));

      setInvitableFriends(filtered);
    } catch (err) {
      console.error("Failed to fetch invitable friends", err);
    }
  }, [currentUser?.id, members]);

  const loadGroupData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const groupData = await apiFetch(`/api/groups/${groupId}`);
      setGroup(groupData);

      const isAcceptedMember = groupData?.status === "accepted" || groupData?.is_member;
      if (isAcceptedMember) {
        const postData = await apiFetch(`/api/posts?group_id=${groupId}`);
        setPosts(getListData(postData));
        await fetchMembers();
      } else {
        setPosts([]);
        setMembers([]);
      }
    } catch (err) {
      console.error("Failed to load group detail", err);
      setError("Unable to load group.");
    } finally {
      setLoading(false);
    }
  }, [groupId, fetchMembers]);

  useEffect(() => {
    if (groupId) {
      void loadGroupData();
    }
  }, [groupId, loadGroupData]);

  useEffect(() => {
    if (group && (group.status === "accepted" || group.is_member)) {
      void fetchInvitableFriends();
    }
  }, [group, members, fetchInvitableFriends]);

  const handleJoin = async () => {
    setActionLoading(true);
    try {
      await apiFetch(`/api/groups/${groupId}/join`, { method: "POST" });
      await loadGroupData();
    } catch (err) {
      alert("Failed to join group: " + err.message);
    } finally {
      setActionLoading(false);
    }
  };

  const handleLeave = async () => {
    if (window.confirm("Are you sure you want to leave this group?")) {
      setActionLoading(true);
      try {
        await apiFetch(`/api/groups/${groupId}/leave`, { method: "POST" });
        await loadGroupData();
      } catch (err) {
        alert("Failed to leave group: " + err.message);
      } finally {
        setActionLoading(false);
      }
    }
  };

  const handleRespondInvitation = async (action) => {
    setActionLoading(true);
    try {
      await apiFetch(`/api/groups/${groupId}/respond`, {
        method: "POST",
        body: { user_id: currentUser.id, action },
      });
      await loadGroupData();
    } catch (err) {
      alert("Failed to respond to invitation: " + err.message);
    } finally {
      setActionLoading(false);
    }
  };

  const handleInvite = async (friendId) => {
    try {
      await apiFetch(`/api/groups/${groupId}/invite`, {
        method: "POST",
        body: { user_id: friendId },
      });
      alert("Invitation sent successfully!");
      setInvitableFriends((prev) => prev.filter((f) => f.id !== friendId));
    } catch (err) {
      alert("Failed to send invitation: " + err.message);
    }
  };

  const handleUpdateRole = async (memberId, action) => {
    try {
      await apiFetch(`/api/groups/${groupId}/members/${memberId}/role/${action}`, {
        method: "POST",
      });
      await fetchMembers();
    } catch (err) {
      alert(`Failed to ${action} member: ` + err.message);
    }
  };

  const isAcceptedMember = group?.status === "accepted" || group?.is_member;
  const currentUserRole = members.find((m) => m.id === currentUser?.id)?.role;
  const isOwnerOrAdmin = currentUserRole === "owner" || currentUserRole === "admin";
  const isOwner = currentUserRole === "owner";

  const handleCreate = (created) => {
    if (!created) return;
    setPosts((current) => [created, ...current]);
  };

  const handlePostChange = (updatedPost) => {
    if (!updatedPost?.id) return;
    setPosts((current) =>
      current.map((post) => (post.id === updatedPost.id ? updatedPost : post))
    );
  };

  return (
    <div className="group-detail-container">
      <div className="group-detail-main">
        <div className="posts">
          <Link to="/groups" className="group-back-link">
            Back to groups
          </Link>
          {loading ? <div>Loading group...</div> : null}
          {error ? <div className="error">{error}</div> : null}
          {group ? (
            <header className="group-feed-header">
              {group.avatar ? <img src={group.avatar} alt="" /> : null}
              <div>
                <h2>{group.title}</h2>
                <p>{group.description || "No description provided."}</p>
              </div>
            </header>
          ) : null}

          {isAcceptedMember ? (
            <NewPost groupId={groupId} onCreate={handleCreate} />
          ) : group && !loading ? (
            <div className="profile-state">
              Only accepted members can view or post in this group.
            </div>
          ) : null}

          {isAcceptedMember && posts.length === 0 && !loading ? (
            <div className="profile-state">No group posts yet.</div>
          ) : null}
          {isAcceptedMember
            ? posts.map((post) => (
                <Post
                  key={post.id}
                  post={post}
                  onPostChange={handlePostChange}
                  onPostDeleted={handlePostChange}
                />
              ))
            : null}
        </div>
      </div>

      {group && !loading && (
        <aside className="group-detail-sidebar">
          {/* Membership Actions Section */}
          <div className="group-detail-section">
            <h4>Membership</h4>
            {!isAcceptedMember ? (
              group.status === "invited" ? (
                <div style={{ display: "flex", gap: "10px" }}>
                  <button
                    type="button"
                    className="group-action-btn group-action-btn--primary"
                    disabled={actionLoading}
                    onClick={() => handleRespondInvitation("accept")}
                  >
                    Accept Invite
                  </button>
                  <button
                    type="button"
                    className="group-action-btn group-action-btn--danger"
                    disabled={actionLoading}
                    onClick={() => handleRespondInvitation("reject")}
                  >
                    Reject
                  </button>
                </div>
              ) : group.status === "pending" ? (
                <span className="group-member-role">Request Pending</span>
              ) : (
                <button
                  type="button"
                  className="group-action-btn group-action-btn--primary"
                  disabled={actionLoading}
                  onClick={handleJoin}
                >
                  Join Group
                </button>
              )
            ) : (
              <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
                <span className="group-member-role group-member-role--owner" style={{ textAlign: "center" }}>
                  {currentUserRole || "member"}
                </span>
                {currentUserRole !== "owner" && (
                  <button
                    type="button"
                    className="group-action-btn group-action-btn--danger"
                    disabled={actionLoading}
                    onClick={handleLeave}
                  >
                    Leave Group
                  </button>
                )}
              </div>
            )}
          </div>

          {/* Members List Section */}
          {isAcceptedMember && (
            <div className="group-detail-section">
              <h4>Members ({members.length})</h4>
              <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
                {members.map((member) => {
                  const isSelf = member.id === currentUser?.id;
                  return (
                    <div key={member.id} className="group-member-item">
                      <div
                        className="group-member-info"
                        onClick={() => {
                          if (member.id === currentUser?.id) {
                            navigate("/profile");
                          } else {
                            navigate(`/user/${member.id}`);
                          }
                        }}
                        style={{ cursor: "pointer" }}
                      >
                        <img
                          src={member.avatar || avatarFallback}
                          alt=""
                          className="group-member-avatar"
                        />
                        <span className="group-member-name">
                          {member.first_name} {member.last_name}
                        </span>
                        <span
                          className={`group-member-role ${
                            member.role === "owner"
                              ? "group-member-role--owner"
                              : member.role === "admin"
                              ? "group-member-role--admin"
                              : ""
                          }`}
                        >
                          {member.role}
                        </span>
                      </div>
                      {/* Admin update action buttons */}
                      {isOwnerOrAdmin && !isSelf && member.role !== "owner" && (
                        <div>
                          {member.role === "member" && (
                            <button
                              type="button"
                              className="group-action-btn"
                              onClick={() => handleUpdateRole(member.id, "promote")}
                            >
                              Promote
                            </button>
                          )}
                          {member.role === "admin" && isOwner && (
                            <button
                              type="button"
                              className="group-action-btn group-action-btn--danger"
                              onClick={() => handleUpdateRole(member.id, "demote")}
                            >
                              Demote
                            </button>
                          )}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Invite Section */}
          {isAcceptedMember && invitableFriends.length > 0 && (
            <div className="group-detail-section">
              <h4>Invite Friends</h4>
              <div className="group-friend-list">
                {invitableFriends.map((friend) => (
                  <div key={friend.id} className="group-friend-item">
                    <div
                      className="group-member-info"
                      onClick={() => navigate(`/user/${friend.id}`)}
                      style={{ cursor: "pointer" }}
                    >
                      <img
                        src={friend.avatar || avatarFallback}
                        alt=""
                        className="group-member-avatar"
                      />
                      <span className="group-member-name">
                        {friend.first_name} {friend.last_name}
                      </span>
                    </div>
                    <button
                      type="button"
                      className="group-action-btn group-action-btn--primary"
                      onClick={() => handleInvite(friend.id)}
                    >
                      Invite
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}
        </aside>
      )}
    </div>
  );
}
