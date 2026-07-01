import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import FollowAction from "./FollowAction";
import { renderWithProviders } from "../../test/render";
import { apiFetch } from "../../utils/api";

vi.mock("../../utils/api", () => ({
  apiFetch: vi.fn(),
}));

describe("FollowAction", () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset();
  });

  it("follows public profiles immediately", async () => {
    const onStatusChange = vi.fn();
    vi.mocked(apiFetch).mockResolvedValueOnce({});

    renderWithProviders(
      <FollowAction targetUserId="user-2" onStatusChange={onStatusChange} />
    );

    fireEvent.click(screen.getByRole("button", { name: "Follow" }));

    await waitFor(() =>
      expect(screen.getByRole("button", { name: "Following" })).toBeEnabled()
    );
    expect(apiFetch).toHaveBeenCalledWith("/api/followers/follow", {
      method: "POST",
      body: { following_id: "user-2" },
    });
    expect(onStatusChange).toHaveBeenCalledWith("following");
  });

  it("keeps private follows pending and cancels with the unfollow endpoint", async () => {
    const onStatusChange = vi.fn();
    vi.mocked(apiFetch).mockResolvedValue({});

    renderWithProviders(
      <FollowAction
        targetUserId="user-2"
        initialStatus="unfollowed"
        isPrivate={true}
        onStatusChange={onStatusChange}
      />
    );

    fireEvent.click(screen.getByRole("button", { name: "Follow" }));
    await screen.findByRole("button", { name: "Requested" });

    fireEvent.click(screen.getByRole("button", { name: "Requested" }));
    await screen.findByRole("button", { name: "Follow" });

    expect(apiFetch).toHaveBeenNthCalledWith(1, "/api/followers/follow", {
      method: "POST",
      body: { following_id: "user-2" },
    });
    expect(apiFetch).toHaveBeenNthCalledWith(2, "/api/followers/unfollow", {
      method: "POST",
      body: { following_id: "user-2" },
    });
    expect(onStatusChange).toHaveBeenLastCalledWith("unfollowed");
  });

  it("renders API errors without changing status", async () => {
    vi.mocked(apiFetch).mockRejectedValueOnce(new Error("request already exists"));

    renderWithProviders(<FollowAction targetUserId="user-2" />);

    fireEvent.click(screen.getByRole("button", { name: "Follow" }));

    expect(await screen.findByText("request already exists")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Follow" })).toBeEnabled();
  });
});
