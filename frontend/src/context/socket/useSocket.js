import { useContext } from "react";
import { SocketContext } from "./socket-context";

/**
 * Returns the shared WebSocket connection state and subscribe function.
 *
 * @returns {import("./socket-context").SocketContextValue}
 * @throws {Error} When called outside a SocketProvider.
 */
export const useSocket = () => {
  const context = useContext(SocketContext);

  if (context === null) {
    throw new Error("useSocket must be used within a SocketProvider");
  }

  return context;
};