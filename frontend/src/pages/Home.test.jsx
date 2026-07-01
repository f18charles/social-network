import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Home from "./Home";
import { renderWithProviders } from "../test/render";

vi.mock("../components/NewPost.jsx", () => ({
  default: ({ onCreate }) => (
    <button type="button" onClick={() => onCreate({ id: "new-post", content: "Created in test", author: { first_name: "Test" } })}>
      Create mocked post
    </button>
  ),
}));

vi.mock("../components/Post.jsx", () => ({
  default: ({ post }) => <article>{post.content}</article>,
}));

describe("Home", () => {
  beforeEach(() => {
    global.fetch = vi.fn();
  });

  it("loads and maps feed posts from the backend envelope", async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        data: [
          {
            id: "post-1",
            content: "Mapped backend post",
            author: { first_name: "Ada", last_name: "Lovelace" },
            created_at: "2026-07-01T00:00:00Z",
          },
        ],
      }),
    });

    renderWithProviders(<Home />);

    expect(await screen.findByText("Mapped backend post")).toBeInTheDocument();
    expect(fetch).toHaveBeenCalledWith("/api/posts", { credentials: "include" });
  });

  it("shows a recoverable error when the feed request fails", async () => {
    vi.mocked(fetch).mockResolvedValueOnce({ ok: false, status: 500 });

    renderWithProviders(<Home />);

    expect(await screen.findByText("Failed to load posts")).toBeInTheDocument();
  });

  it("prepends newly created posts", async () => {
    vi.mocked(fetch).mockResolvedValueOnce({ ok: true, json: async () => ({ data: [] }) });

    renderWithProviders(<Home />);

    await waitFor(() => expect(fetch).toHaveBeenCalled());
    screen.getByRole("button", { name: "Create mocked post" }).click();

    expect(await screen.findByText("Created in test")).toBeInTheDocument();
  });
});
