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
    await waitFor(() =>
      expect(dislike).toHaveAttribute("aria-pressed", "true")
    );
    expect(like).toHaveAttribute("aria-pressed", "false");
    expect(screen.queryByText("Post details opened")).not.toBeInTheDocument();
    expect(apiFetch).toHaveBeenCalledWith("/api/posts/post-1/vote", {
      method: "PUT",
      body: { vote: "dislike" },
    });
  });
  it("renders accessible privacy badges for each privacy mode", () => {
    renderWithProviders(
      <>
        <Post post={{ ...post, id: "public-post", privacy: "public" }} />
        <Post post={{ ...post, id: "followers-post", privacy: "almost_private" }} />
        <Post post={{ ...post, id: "private-post", privacy: "private" }} />
      </>
    );

    expect(screen.getByLabelText("Public post")).toBeInTheDocument();
    expect(screen.getByLabelText("Followers post")).toBeInTheDocument();
    expect(screen.getByLabelText("Private post")).toBeInTheDocument();
  });

  it("shows keyboard-accessible edit and delete controls only to the author", async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce({ id: "post-1", deleted: true });
    const onPostChange = vi.fn();

    renderWithProviders(<Post post={{ ...post, author: { id: "user-1", name: "Amina Njeri" } }} onPostChange={onPostChange} />, {
      auth: { currentUser: { id: "user-1" }, isAuthenticated: true },
    });

    expect(screen.getByRole("button", { name: "Edit post" })).toBeInTheDocument();
    const deleteButton = screen.getByRole("button", { name: "Delete post" });
    expect(deleteButton).toBeInTheDocument();

    fireEvent.click(deleteButton);

    await waitFor(() =>
      expect(apiFetch).toHaveBeenCalledWith("/api/posts/post-1", { method: "DELETE" })
    );
    expect(onPostChange).toHaveBeenCalledWith({ id: "post-1", deleted: true });
  });

  it("hides author controls for other users", () => {
    renderWithProviders(<Post post={{ ...post, author: { id: "user-1", name: "Amina Njeri" } }} />, {
      auth: { currentUser: { id: "user-2" }, isAuthenticated: true },
    });

    expect(screen.queryByRole("button", { name: "Edit post" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Delete post" })).not.toBeInTheDocument();
  });

  it("renders deleted post tombstones without profile controls", () => {
    renderWithProviders(
      <Routes>
        <Route path="/" element={<Post post={{ id: "deleted-post", deleted: true }} />} />
      </Routes>
    );

    expect(screen.getByText("Deleted user")).toBeInTheDocument();
    expect(
      screen.getByText("This post is no longer available.")
    ).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Open Deleted user's profile/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /^Like this post/i })).not.toBeInTheDocument();
  });

});
