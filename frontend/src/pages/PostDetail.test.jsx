import { fireEvent, screen, waitFor } from "@testing-library/react";
import { Route, Routes } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import PostDetail from "./PostDetail";
import { renderWithProviders } from "../test/render";
import { apiFetch } from "../utils/api.js";

vi.mock("../utils/api.js", () => ({
  apiFetch: vi.fn(),
}));

const post = {
  id: "post-1",
  author: {
    id: "user-1",
    first_name: "Amina",
    last_name: "Njeri",
    nickname: "amina",
  },
  content: "Database backed post content.",
  image_url: "/uploads/images/e2e-fixture-sample.jpg",
  privacy: "public",
  created_at: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
  like_count: 4,
  comment_count: 2,
};

const comments = [
  {
    id: "comment-1",
    author: { id: "user-2", first_name: "Bianca", last_name: "Stone" },
    content: "Top level database comment.",
    image_url: null,
    created_at: "2026-06-30T10:00:00Z",
    replies_count: 1,
    replies: [],
  },
];

const repliesEnvelope = {
  status: "success",
  data: [
    {
      id: "reply-1",
      author: { id: "user-3", first_name: "Chidi", last_name: "Okafor" },
      content: "Nested database reply.",
      image_url: "/uploads/images/e2e-fixture-reply.gif",
      created_at: "2026-06-30T10:05:00Z",
      replies_count: 0,
      replies: [],
    },
  ],
  pagination: { limit: 10, offset: 0, has_more: false },
};

describe("PostDetail", () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset();
  });

  it("loads folded comments and fetches replies on demand", async () => {
    vi.mocked(apiFetch).mockImplementation((url) => {
      if (url === "/api/posts/post-1") return Promise.resolve(post);
      if (url === "/api/posts/post-1/comments")
        return Promise.resolve(comments);
      if (url === "/api/comments/comment-1/replies?limit=10&offset=0")
        return Promise.resolve(repliesEnvelope);
      return Promise.reject(new Error(`Unexpected URL ${url}`));
    });

    renderWithProviders(
      <Routes>
        <Route path="/post/:id" element={<PostDetail />} />
      </Routes>,
      { route: "/post/post-1" }
    );

    expect(
      await screen.findByText("Database backed post content.")
    ).toBeInTheDocument();
    expect(screen.getByText("Top level database comment.")).toBeInTheDocument();
    expect(
      screen.queryByText("Nested database reply.")
    ).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /View replies/i }));

    expect(
      await screen.findByText("Nested database reply.")
    ).toBeInTheDocument();
    expect(screen.getByAltText("comment attachment")).toHaveAttribute(
      "src",
      "/uploads/images/e2e-fixture-reply.gif"
    );
    await waitFor(() =>
      expect(apiFetch).toHaveBeenCalledWith(
        "/api/comments/comment-1/replies?limit=10&offset=0"
      )
    );
    expect(apiFetch).toHaveBeenCalledWith("/api/posts/post-1");
    expect(apiFetch).toHaveBeenCalledWith("/api/posts/post-1/comments");
  });
  it("renders deleted comment tombstones with active nested replies", async () => {
    vi.mocked(apiFetch).mockImplementation((url) => {
      if (url === "/api/posts/post-1") return Promise.resolve(post);
      if (url === "/api/posts/post-1/comments")
        return Promise.resolve([
          {
            id: "deleted-comment-1",
            deleted: true,
            replies_count: 1,
            replies: [
              {
                id: "active-reply-1",
                author: { id: "user-3", first_name: "Chidi", last_name: "Okafor" },
                content: "Still visible below the tombstone.",
                image_url: null,
                created_at: "2026-06-30T10:05:00Z",
                replies_count: 0,
                replies: [],
              },
            ],
          },
        ]);
      return Promise.reject(new Error(`Unexpected URL ${url}`));
    });

    renderWithProviders(
      <Routes>
        <Route path="/post/:id" element={<PostDetail />} />
      </Routes>,
      { route: "/post/post-1" }
    );

    expect(
      await screen.findByText("This comment is no longer available.")
    ).toBeInTheDocument();
    expect(
      screen.getByText("Still visible below the tombstone.")
    ).toBeInTheDocument();
  });

});
