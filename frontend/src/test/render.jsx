import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { AuthContext } from "../context/auth-context";

const defaultAuth = {
  currentUser: null,
  isAuthenticated: false,
  isLoading: false,
  login: async () => null,
  logout: async () => {},
  refresh: async () => null,
  refreshUnreadNotifications: async () => 0,
};

export const renderWithProviders = (
  ui,
  { route = "/", auth = {}, router = true } = {}
) => {
  const authValue = { ...defaultAuth, ...auth };
  const tree = (
    <AuthContext.Provider value={authValue}>
      {router ? <MemoryRouter initialEntries={[route]}>{ui}</MemoryRouter> : ui}
    </AuthContext.Provider>
  );

  return render(tree);
};
