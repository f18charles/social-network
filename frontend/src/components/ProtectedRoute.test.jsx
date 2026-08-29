import { screen } from "@testing-library/react";
import { Route, Routes } from "react-router";
import { describe, expect, it } from "vitest";
import ProtectedRoute from "./ProtectedRoute";
import { renderWithProviders } from "../test/render";

const ProtectedRoutes = () => (
  <Routes>
    <Route
      path="/private"
      element={
        <ProtectedRoute>
          <p>Private dashboard</p>
        </ProtectedRoute>
      }
    />
    <Route path="/welcome" element={<p>Welcome screen</p>} />
  </Routes>
);

describe("ProtectedRoute", () => {
  it("shows loading while auth state is being resolved", () => {
    renderWithProviders(<ProtectedRoutes />, {
      route: "/private",
      auth: { isLoading: true },
    });

    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it("redirects unauthenticated users to welcome", async () => {
    renderWithProviders(<ProtectedRoutes />, {
      route: "/private",
      auth: { isAuthenticated: false, isLoading: false },
    });

    expect(await screen.findByText("Welcome screen")).toBeInTheDocument();
  });

  it("renders protected content for authenticated users", () => {
    renderWithProviders(<ProtectedRoutes />, {
      route: "/private",
      auth: { isAuthenticated: true, currentUser: { id: "user-1" } },
    });

    expect(screen.getByText("Private dashboard")).toBeInTheDocument();
  });
});
