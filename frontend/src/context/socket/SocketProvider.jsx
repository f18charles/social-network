import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { SocketContext } from "./socket-context";
import { useAuth } from "../auth/useAuth";

/**
 * Opens a single WebSocket connection for the whole app once the user is
 * authenticated, and fans incoming messages out to subscribers by type.
 *
 * @param {{children: import("react").ReactNode}} props
 * @returns {import("react").JSX.Element}
 */

export const SocketProvider = ({ children }) => {
  const { isAuthenticated, refreshUnreadNotifications } = useAuth();
  const socketRef = useRef(null);
  const emitterRef = useRef(new EventTarget());
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    if (!isAuthenticated) {
      socketRef.current?.close();
      socketRef.current = null;
      setIsConnected(false);
      return;
    }

    const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${wsProtocol}//${window.location.hostname}:8080/api/ws`;
    const socket = new WebSocket(wsUrl);
    socketRef.current = socket;

    socket.onopen = () => {
      setIsConnected(true);
      console.log("WebSocket connected");
    };

    socket.onclose = () => {
      setIsConnected(false);
      console.log("WebSocket disconnected");
    };

    socket.onerror = (err) => {
      console.error("WebSocket error:", err);
    };

    socket.onmessage = (event) => {
      try {
        const wsMsg = JSON.parse(event.data);
        emitterRef.current.dispatchEvent(
          new CustomEvent(wsMsg.type, { detail: wsMsg.payload })
        );

        // Keep the unread badge in sync regardless of which page you're on
        if (wsMsg.type === "notification") {
          refreshUnreadNotifications();
        }
      } catch (err) {
        console.error("Error parsing WebSocket message:", err);
      }
    };

    return () => {
      socket.close();
    };
  }, [isAuthenticated, refreshUnreadNotifications]);

  const subscribe = useCallback((type, handler) => {
    const listener = (event) => handler(event.detail);
    emitterRef.current.addEventListener(type, listener);
    
    return () => {
      emitterRef.current.removeEventListener(type, listener);
    };
  }, []);

  const value = useMemo(
    () => ({ isConnected, subscribe }),
    [isConnected, subscribe]
  );

  return (
    <SocketContext.Provider value={value}>
      {children}
    </SocketContext.Provider>
  );
};