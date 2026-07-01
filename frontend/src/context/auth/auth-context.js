import { createContext } from "react";

/**
 * Credentials accepted by the login endpoint.
 *
 * @typedef {object} LoginCredentials
 * @property {string} email User email address.
 * @property {string} password User password.
 */

/**
 * Authentication state and actions shared across the application.
 *
 * @typedef {object} AuthContextValue
 * @property {object|null} currentUser Authenticated user's profile.
 * @property {boolean} isAuthenticated Whether a user is authenticated.
 * @property {boolean} isLoading Whether an authentication request is pending.
 * @property {(credentials: LoginCredentials) => Promise<object|null>} login
 * Authenticates credentials and returns the current user.
 * @property {() => Promise<void>} logout Ends the current session.
 * @property {() => Promise<object|null>} refresh Reloads the current user.
 */

/**
 * React context containing the application's shared authentication state.
 *
 * @type {import("react").Context<AuthContextValue|null>}
 */
export const AuthContext = createContext(null);
