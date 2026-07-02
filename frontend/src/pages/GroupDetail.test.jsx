import { screen } from "@testing-library/react";
import { Route, Routes } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import GroupDetail from "./GroupDetail.jsx";
import { renderWithProviders } from "../test/render";
import { apiFetch } from "../utils/api.js";

vi.mock("../utils/api.js", () => ({
  apiFetch: vi.fn(),
}));

vi.mock("../components/NewPost.jsx", () => ({
  default: ({ groupId }) => <form aria-label="group composer">Composer for {groupId}</form>,
}));

vi.mock("../components/Post.jsx", () => ({
  default: ({ post }) => <article>{post.content}</article>,
}));

describe("GroupDetail", () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset();
  });

  it("renders an accepted member group feed with composer", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce({ id: "group-1", title: "Writers", status: "accepted" })
      .mockResolvedValueOnce({ data: [{ id: "post-1", content: "Group-only post" }] });

    renderWithProviders(
      <Routes>
        <Route path="/groups/:groupId" element={<GroupDetail />} />
      </Routes>,
      { route: "/groups/group-1" }
    );

    expect(await screen.findByText("Writers")).toBeInTheDocument();
    expect(screen.getByRole("form", { name: "group composer" })).toHaveTextContent("group-1");
    expect(screen.getByText("Group-only post")).toBeInTheDocument();
    expect(apiFetch).toHaveBeenCalledWith("/api/posts?group_id=group-1");
  });

  it("hides composer and feed for non-members", async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce({
      id: "group-1",
      title: "Writers",
      status: "none",
    });

    renderWithProviders(
      <Routes>
        <Route path="/groups/:groupId" element={<GroupDetail />} />
      </Routes>,
      { route: "/groups/group-1" }
    );

    expect(await screen.findByText("Writers")).toBeInTheDocument();
    expect(screen.queryByRole("form", { name: "group composer" })).not.toBeInTheDocument();
    expect(screen.getByText("Only accepted members can view or post in this group.")).toBeInTheDocument();
    expect(apiFetch).not.toHaveBeenCalledWith("/api/posts?group_id=group-1");
  });
});
