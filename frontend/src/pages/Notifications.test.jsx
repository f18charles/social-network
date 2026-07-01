import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Notifications from "./Notifications";
import { renderWithProviders } from "../test/render";
import { apiFetch } from "../utils/api";

vi.mock("../utils/api", () => ({
  apiFetch: vi.fn(),
}));

describe("Notifications", () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset();
    vi.spyOn(window, "alert").mockImplementation(() => {});
  });

  it("loads notifications and marks one as read", async () => {
    const refreshUnreadNotifications = vi.fn();
    vi.mocked(apiFetch)
      .mockResolvedValueOnce([
        { id: "n-1", message: "Ada requested to follow you.", is_read: false, created_at: "2026-07-01T00:00:00Z" },
      ])
      .mockResolvedValueOnce({})
      .mockResolvedValueOnce([
        { id: "n-1", message: "Ada requested to follow you.", is_read: true, created_at: "2026-07-01T00:00:00Z" },
      ]);

    renderWithProviders(<Notifications />, {
      auth: { refreshUnreadNotifications },
    });

    fireEvent.click(await screen.findByRole("button", { name: "Mark Read" }));

    await waitFor(() =>
      expect(apiFetch).toHaveBeenCalledWith("/api/notifications/n-1/read", { method: "POST" })
    );
    expect(refreshUnreadNotifications).toHaveBeenCalled();
  });

  it("marks all unread notifications as read", async () => {
    const refreshUnreadNotifications = vi.fn();
    vi.mocked(apiFetch)
      .mockResolvedValueOnce([
        { id: "n-1", message: "Ada requested to follow you.", is_read: false, created_at: "2026-07-01T00:00:00Z" },
      ])
      .mockResolvedValueOnce({})
      .mockResolvedValueOnce([]);

    renderWithProviders(<Notifications />, {
      auth: { refreshUnreadNotifications },
    });

    fireEvent.click(await screen.findByRole("button", { name: "Mark all as read" }));

    await waitFor(() =>
      expect(apiFetch).toHaveBeenCalledWith("/api/notifications/read/all", { method: "POST" })
    );
  });

  it("shows an empty state", async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce([]);

    renderWithProviders(<Notifications />);

    expect(await screen.findByText("You have no notifications yet.")).toBeInTheDocument();
  });
});
