import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Groups from "./Groups";
import { renderWithProviders } from "../test/render";
import { apiFetch } from "../utils/api";

vi.mock("../utils/api", () => ({
  apiFetch: vi.fn(),
}));

describe("Groups", () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset();
  });

  it("creates a group and refreshes the list", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce({ id: "group-1" })
      .mockResolvedValueOnce([{ id: "group-1", title: "Testing Group", description: "Reliable tests", status: "accepted", creator_id: "user-1" }]);

    renderWithProviders(<Groups />, {
      auth: { currentUser: { id: "user-1" }, isAuthenticated: true },
    });

    await screen.findByText("No groups available. Create one to get started!");
    fireEvent.click(screen.getByRole("button", { name: "Create Group" }));
    fireEvent.change(screen.getByLabelText("Title"), { target: { value: "Testing Group" } });
    fireEvent.change(screen.getByLabelText("Description"), { target: { value: "Reliable tests" } });
    fireEvent.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() =>
      expect(apiFetch).toHaveBeenCalledWith("/api/groups", {
        method: "POST",
        body: { title: "Testing Group", description: "Reliable tests" },
      })
    );
    expect(await screen.findByText("Testing Group")).toBeInTheDocument();
  });

  it("submits a join request for non-member groups", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce([{ id: "group-1", title: "Open Group", description: "", status: "none" }])
      .mockResolvedValueOnce({})
      .mockResolvedValueOnce([{ id: "group-1", title: "Open Group", description: "", status: "pending_request" }]);

    renderWithProviders(<Groups />, {
      auth: { currentUser: { id: "user-2" }, isAuthenticated: true },
    });

    fireEvent.click(await screen.findByRole("button", { name: "Join Group" }));

    await waitFor(() =>
      expect(apiFetch).toHaveBeenCalledWith("/api/groups/group-1/join", { method: "POST" })
    );
    expect(await screen.findByText("Request Pending")).toBeInTheDocument();
  });
});
