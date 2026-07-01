import { screen } from "@testing-library/react";
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
    author: { first_name: "Bianca", last_name: "Stone" },
    content: "Top level database comment.",
    image_url: null,
    created_at: "2026-06-30T10:00:00Z",
    replies: [
      {
        id: "reply-1",
        author: { first_name: "Chidi", last_name: "Okafor" },
        content: "Nested database reply.",
        image_url: "/uploads/images/e2e-fixture-reply.gif",
        created_at: "2026-06-30T10:05:00Z",
        replies: [],
      },
    ],
  },
];

describe("PostDetail", () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset();
  });

  it("loads and renders the API post and nested comments", async () => {
    vi.mocked(apiFetch).mockImplementation((url) => {
      if (url === "/api/posts/post-1") return Promise.resolve(post);
      if (url === "/api/posts/post-1/comments")
        return Promise.resolve(comments);
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
    expect(screen.getByText("Nested database reply.")).toBeInTheDocument();
    expect(screen.getByAltText("comment attachment")).toHaveAttribute(
      "src",
      "/uploads/images/e2e-fixture-reply.gif"
    );
    expect(apiFetch).toHaveBeenCalledWith("/api/posts/post-1");
    expect(apiFetch).toHaveBeenCalledWith("/api/posts/post-1/comments");
  });
});
