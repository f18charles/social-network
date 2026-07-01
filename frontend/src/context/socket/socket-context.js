import { createContext } from "react";

/**
 * Live WebSocket connection and pub/sub access shared across the app.
 *
 * @typedef {object} SocketContextValue
 * @property {boolean} isConnected Whether the socket is currently open.
 * @property {(type: string, handler: (payload: any) => void) => () => void} subscribe
 * Registers a handler for a given WS message type (e.g. "chat", "notification").
 * Returns an unsubscribe function — call it in your effect cleanup.
 */

/** @type {import("react").Context<SocketContextValue|null>} */
export const SocketContext = createContext(null);