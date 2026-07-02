import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { AuthContext } from "../context/auth/auth-context";
import { SocketContext } from "../context/socket/socket-context";

const defaultAuth = {
  currentUser: null,
  isAuthenticated: false,
  isLoading: false,
  login: async () => null,
  logout: async () => {},
  refresh: async () => null,
  refreshUnreadNotifications: async () => 0,
};

const defaultSocket = {
  isConnected: false,
  subscribe: () => () => {},
};

export const renderWithProviders = (
  ui,
  { route = "/", auth = {}, socket = {}, router = true } = {}
) => {
  const authValue = { ...defaultAuth, ...auth };
  const socketValue = { ...defaultSocket, ...socket };
  const content = router ? <MemoryRouter initialEntries={[route]}>{ui}</MemoryRouter> : ui;
  const tree = (
    <AuthContext.Provider value={authValue}>
      <SocketContext.Provider value={socketValue}>{content}</SocketContext.Provider>
    </AuthContext.Provider>
  );

  return render(tree);
};
