import { fireEvent, screen } from "@testing-library/react";
import { Route, Routes } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import ProfileComments from "./ProfileComments.jsx";
import { renderWithProviders } from "../../test/render.jsx";
import { apiFetch } from "../../utils/api";

vi.mock("../../utils/api", () => ({
  apiFetch: vi.fn(),
}));

describe("ProfileComments", () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset();
  });

  it("opens the post with the selected comment id", async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce({
      status: "success",
      data: [
        {
          id: "comment-1",
          post_id: "post-1",
          parent_comment_id: null,
          author: { id: "user-1", first_name: "Amina", last_name: "Njeri" },
          content: "Profile activity comment.",
          created_at: "2026-06-30T10:00:00Z",
          replies_count: 2,
        },
      ],
      pagination: { limit: 20, offset: 0, has_more: false },
    });

    renderWithProviders(
      <Routes>
        <Route path="/profile" element={<ProfileComments userId="user-1" />} />
        <Route path="/post/:id" element={<p>Post opened</p>} />
      </Routes>,
      { route: "/profile" }
    );

    fireEvent.click(await screen.findByText("Profile activity comment."));

    expect(await screen.findByText("Post opened")).toBeInTheDocument();
    expect(apiFetch).toHaveBeenCalledWith("/api/users/user-1/comments");
  });
});
