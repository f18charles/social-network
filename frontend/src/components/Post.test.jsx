import { fireEvent, screen, waitFor } from "@testing-library/react";
import { Route, Routes } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Post from "./Post";
import { renderWithProviders } from "../test/render";
import { apiFetch } from "../utils/api.js";

vi.mock("../utils/api.js", () => ({
  apiFetch: vi.fn(),
}));

const post = {
  id: "post-1",
  author: { name: "Amina Njeri" },
  content: "A careful post about testing behavior.",
  privacy: "public",
  created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
  like_count: 2,
  dislike_count: 0,
  comment_count: 3,
  viewer_vote: "none",
};

describe("Post", () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset();
  });

  it("opens the post when the post content is selected", async () => {
    renderWithProviders(
      <Routes>
        <Route path="/" element={<Post post={post} />} />
        <Route path="/post/:id" element={<p>Post details opened</p>} />
      </Routes>
    );

    fireEvent.click(screen.getByText(post.content));

    expect(await screen.findByText("Post details opened")).toBeInTheDocument();
  });

  it("updates reactions through the API without opening the post", async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce({
      like_count: 3,
      dislike_count: 0,
      viewer_vote: "like",
    });
    vi.mocked(apiFetch).mockResolvedValueOnce({
      like_count: 2,
      dislike_count: 1,
      viewer_vote: "dislike",
    });

    renderWithProviders(
      <Routes>
        <Route path="/" element={<Post post={post} />} />
        <Route path="/post/:id" element={<p>Post details opened</p>} />
      </Routes>
    );

    const like = screen.getByRole("button", { name: /^Like this post/i });
    const dislike = screen.getByRole("button", { name: /^Dislike this post/i });

    fireEvent.click(like);
    await waitFor(() => expect(like).toHaveAttribute("aria-pressed", "true"));
    expect(screen.queryByText("Post details opened")).not.toBeInTheDocument();
    expect(apiFetch).toHaveBeenCalledWith("/api/posts/post-1/vote", {
      method: "PUT",
      body: { vote: "like" },
    });

    fireEvent.click(dislike);
    await waitFor(() => expect(dislike).toHaveAttribute("aria-pressed", "true"));
    expect(like).toHaveAttribute("aria-pressed", "false");
    expect(screen.queryByText("Post details opened")).not.toBeInTheDocument();
    expect(apiFetch).toHaveBeenCalledWith("/api/posts/post-1/vote", {
      method: "PUT",
      body: { vote: "dislike" },
    });
  });
});
