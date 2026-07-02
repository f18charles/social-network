import { useEffect, useState, useCallback } from "react";
import { useSearchParams, useNavigate } from "react-router";
import { apiFetch, ApiError } from "../utils/api";
import Post from "../components/Post";
import avatarFallback from "../assets/user.svg";
import "../styles/search.css";

const Search = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const query = searchParams.get("query") || "";
  const activeFilter = searchParams.get("type") || "all";

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [users, setUsers] = useState([]);
  const [posts, setPosts] = useState([]);

  const performSearch = useCallback(async () => {
    if (!query.trim()) {
      setUsers([]);
      setPosts([]);
      return;
    }

    setLoading(true);
    setError("");

    try {
      // Fetch results from unified search endpoint
      const response = await apiFetch(
        `/api/users/search?query=${encodeURIComponent(query)}&type=${activeFilter}`
      );

      if (activeFilter === "users") {
        setUsers(response || []);
        setPosts([]);
      } else if (activeFilter === "posts") {
        setPosts(response || []);
        setUsers([]);
      } else {
        // "all" filter returns { users: [...], posts: [...] }
        setUsers(response?.users || []);
        setPosts(response?.posts || []);
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Search failed");
    } finally {
      setLoading(false);
    }
  }, [query, activeFilter]);

  useEffect(() => {
    performSearch();
  }, [performSearch]);

  const handleFilterChange = (filterType) => {
    const params = new URLSearchParams(searchParams);
    params.set("type", filterType);
    setSearchParams(params);
  };

  const handleUserClick = (userId) => {
    navigate(`/user/${userId}`);
  };

  return (
    <div className="search-page">
      <div className="search-page__header">
        <h1>Search Results</h1>
        {query && (
          <p className="search-page__query-text">
            Showing results for &ldquo;<strong>{query}</strong>&rdquo;
          </p>
        )}
      </div>

      {/* Filter Tabs */}
      <div className="search-filters">
        <button
          type="button"
          className={`search-filter-btn ${activeFilter === "all" ? "is-active" : ""}`}
          onClick={() => handleFilterChange("all")}
        >
          All Results
        </button>
        <button
          type="button"
          className={`search-filter-btn ${activeFilter === "users" ? "is-active" : ""}`}
          onClick={() => handleFilterChange("users")}
        >
          People
        </button>
        <button
          type="button"
          className={`search-filter-btn ${activeFilter === "posts" ? "is-active" : ""}`}
          onClick={() => handleFilterChange("posts")}
        >
          Posts
        </button>
      </div>

      {loading && (
        <div className="search-loading">
          <div className="search-spinner" />
          <p>Searching network...</p>
        </div>
      )}

      {error && <div className="search-error">{error}</div>}

      {!loading && !error && (
        <div className="search-results">
          {/* Users Section */}
          {activeFilter !== "posts" && users.length > 0 && (
            <section className="search-section">
              <h2>People</h2>
              <div className="search-users-grid">
                {users.map((user) => (
                  <div
                    key={user.id}
                    className="search-user-card"
                    onClick={() => handleUserClick(user.id)}
                  >
                    <img
                      src={user.avatar || avatarFallback}
                      alt={`${user.first_name}'s avatar`}
                      className="search-user-card__avatar"
                    />
                    <div className="search-user-card__info">
                      <h3>
                        {`${user.first_name || ""} ${user.last_name || ""}`.trim() ||
                          "Unnamed User"}
                      </h3>
                      {user.nickname && (
                        <p className="search-user-card__nickname">@{user.nickname}</p>
                      )}
                      {user.about_me && (
                        <p className="search-user-card__bio">{user.about_me}</p>
                      )}
                    </div>
                    <div className="search-user-card__badges">
                      <span className={`profile-badge ${user.is_public ? "" : "is-private"}`}>
                        {user.is_public ? "Public" : "Private"}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </section>
          )}

          {/* Posts Section */}
          {activeFilter !== "users" && posts.length > 0 && (
            <section className="search-section">
              <h2>Posts</h2>
              <div className="search-posts-list">
                {posts.map((post) => (
                  <Post key={post.id} post={post} />
                ))}
              </div>
            </section>
          )}

          {/* Empty State */}
          {users.length === 0 && posts.length === 0 && query && (
            <div className="search-empty">
              <h3>No results found</h3>
              <p>Try checking your spelling or searching for a different keyword.</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default Search;
