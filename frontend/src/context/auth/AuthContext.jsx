import { useCallback, useEffect, useMemo, useState } from "react";
import { apiFetch, ApiError } from "../../utils/api";
import { AuthContext } from "./auth-context";

/**
 * Provides shared authentication state and actions to descendant components.
 *
 * @param {{children: import("react").ReactNode}} props Provider properties.
 * @returns {import("react").JSX.Element} Authentication context provider.
 */
export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [unreadNotifications, setUnreadNotifications] = useState(0);

  const refreshUnreadNotifications = useCallback(async () => {
    try {
      const notifications = await apiFetch("/api/notifications");
      const unreadCount = Array.isArray(notifications)
        ? notifications.filter((notification) => !notification.is_read).length
        : 0;

      setUnreadNotifications(unreadCount);
      return unreadCount;
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        setUnreadNotifications(0);
        return 0;
      }

      console.error("Failed to refresh notifications", error);
      setUnreadNotifications(0);
      return 0;
    }
  }, []);

  const refresh = useCallback(async () => {
    setIsLoading(true);

    try {
      const user = await apiFetch("/api/users/me");
      setCurrentUser(user);
      await refreshUnreadNotifications();
      return user;
    } catch (error) {
      setCurrentUser(null);
      setUnreadNotifications(0);

      if (error instanceof ApiError && error.status === 401) {
        return null;
      }

      throw error;
    } finally {
      setIsLoading(false);
    }
  }, [refreshUnreadNotifications]);

  const login = useCallback(
    async (credentials) => {
      setIsLoading(true);

      try {
        await apiFetch("/api/users/login", {
          method: "POST",
          body: credentials,
        });
        return await refresh();
      } catch (error) {
        setCurrentUser(null);
        setIsLoading(false);
        throw error;
      }
    },
    [refresh]
  );

  const logout = useCallback(async () => {
    setIsLoading(true);

    try {
      await apiFetch("/api/users/logout", { method: "POST" });
      setCurrentUser(null);
      setUnreadNotifications(0);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    let isActive = true;

    const loadCurrentUser = async () => {
      try {
        const user = await apiFetch("/api/users/me");

        if (isActive) {
          setCurrentUser(user);
          await refreshUnreadNotifications(user);
        }
      } catch {
        if (isActive) {
          setCurrentUser(null);
          setUnreadNotifications(0);
        }
      } finally {
        if (isActive) {
          setIsLoading(false);
        }
      }
    };

    loadCurrentUser();

    return () => {
      isActive = false;
    };
  }, [refreshUnreadNotifications]);

  const value = useMemo(
    () => ({
      currentUser,
      isAuthenticated: currentUser !== null,
      isLoading,
      unreadNotifications,
      refreshUnreadNotifications,
      login,
      logout,
      refresh,
    }),
    [
      currentUser,
      isLoading,
      unreadNotifications,
      login,
      logout,
      refresh,
      refreshUnreadNotifications,
    ]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
