import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AuthProvider } from "./AuthContext";
import { useAuth } from "./useAuth";
import { apiFetch } from "../../utils/api";

vi.mock("../utils/api", async (importOriginal) => {
  const original = await importOriginal();

  return {
    ...original,
    apiFetch: vi.fn(),
  };
});

const AuthStatus = () => {
  const { currentUser, isAuthenticated, isLoading, logout, refresh } =
    useAuth();

  if (isLoading) {
    return <p>Loading session</p>;
  }

  return (
    <>
      <p>
        {isAuthenticated
          ? `Signed in as ${currentUser.username}`
          : "Signed out"}
      </p>
      <button type="button" onClick={refresh}>
        Refresh
      </button>
      <button type="button" onClick={logout}>
        Log out
      </button>
    </>
  );
};

describe("AuthProvider", () => {
  beforeEach(() => {
    apiFetch.mockReset();
  });

  it("loads the current user and exposes an authenticated session", async () => {
    apiFetch
      .mockResolvedValueOnce({ id: 1, username: "ada" })
      .mockResolvedValueOnce([]);

    render(
      <AuthProvider>
        <AuthStatus />
      </AuthProvider>
    );

    expect(screen.getByText("Loading session")).toBeInTheDocument();
    expect(await screen.findByText("Signed in as ada")).toBeInTheDocument();
    expect(apiFetch).toHaveBeenCalledWith("/api/users/me");
    await waitFor(() => {
      expect(apiFetch).toHaveBeenCalledWith("/api/notifications");
    });
  });

  it("refreshes the current user and clears it after logout", async () => {
    apiFetch
      .mockResolvedValueOnce({ id: 1, username: "ada" })
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce({ id: 1, username: "grace" })
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce(null);

    render(
      <AuthProvider>
        <AuthStatus />
      </AuthProvider>
    );

    expect(await screen.findByText("Signed in as ada")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    expect(await screen.findByText("Signed in as grace")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Log out" }));
    expect(await screen.findByText("Signed out")).toBeInTheDocument();
    expect(apiFetch).toHaveBeenNthCalledWith(2, "/api/notifications");
    expect(apiFetch).toHaveBeenNthCalledWith(4, "/api/notifications");
    expect(apiFetch).toHaveBeenNthCalledWith(5, "/api/users/logout", {
      method: "POST",
    });
  });
});
