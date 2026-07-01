import { useContext } from "react";
import { AuthContext } from "./auth-context";

/**
 * Returns the shared authentication state and actions.
 *
 * @returns {import("./auth-context").AuthContextValue} Authentication context.
 * @throws {Error} When called outside an AuthProvider.
 */
export const useAuth = () => {
  const context = useContext(AuthContext);

  if (context === null) {
    throw new Error("useAuth must be used within an AuthProvider");
  }

  return context;
};
