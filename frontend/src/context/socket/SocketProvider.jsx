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
      const timer = window.setTimeout(() => {
        setIsConnected(false);
      }, 0);
      return () => window.clearTimeout(timer);
    }

    let wsUrl;
    if (import.meta.env.VITE_WS_URL) {
      wsUrl = import.meta.env.VITE_WS_URL;
    } else if (import.meta.env.VITE_API_URL) {
      const parsed = new URL(import.meta.env.VITE_API_URL, window.location.href);
      const wsProto = parsed.protocol === "https:" ? "wss:" : "ws:";
      wsUrl = `${wsProto}//${parsed.host}/api/ws`;
    } else {
      const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      wsUrl = `${wsProtocol}//${window.location.host}/api/ws`;
    }
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
        let detail = wsMsg.data ?? wsMsg.payload ?? wsMsg.error;
        if (detail && typeof detail === "object" && wsMsg.client_message_id) {
          detail = { ...detail, client_message_id: wsMsg.client_message_id };
        }
        emitterRef.current.dispatchEvent(
          new CustomEvent(wsMsg.type, {
            detail: detail,
          })
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

  const send = useCallback((payload) => {
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      return false;
    }
    socketRef.current.send(JSON.stringify(payload));
    return true;
  }, []);

  const value = useMemo(
    () => ({ isConnected, subscribe, send }),
    [isConnected, subscribe, send]
  );

  return (
    <SocketContext.Provider value={value}>{children}</SocketContext.Provider>
  );
};
