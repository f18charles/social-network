import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import NewPost from "./NewPost";
import { renderWithProviders } from "../test/render";
import { apiFetch } from "../utils/api.js";

vi.mock("../utils/api.js", () => ({
  apiFetch: vi.fn(),
}));

describe("NewPost", () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset();
    vi.spyOn(console, "error").mockImplementation(() => {});
    vi.stubGlobal("URL", {
      ...URL,
      createObjectURL: vi.fn(() => "blob:preview"),
      revokeObjectURL: vi.fn(),
    });
  });

  it("keeps submit disabled until content or an image is provided", () => {
    renderWithProviders(<NewPost />);

    expect(screen.getByRole("button", { name: "Post" })).toBeDisabled();

    fireEvent.change(screen.getByPlaceholderText("What's on your mind?"), {
      target: { value: "A useful update" },
    });

    expect(screen.getByRole("button", { name: "Post" })).toBeEnabled();
  });

  it("submits multipart posts with selected privacy and resets on success", async () => {
    const onCreate = vi.fn();
    vi.mocked(apiFetch).mockResolvedValueOnce({
      id: "post-1",
      content: "A useful update",
    });

    renderWithProviders(<NewPost onCreate={onCreate} />);

    fireEvent.change(screen.getByPlaceholderText("What's on your mind?"), {
      target: { value: "A useful update" },
    });
    fireEvent.change(screen.getByRole("combobox"), {
      target: { value: "almost_private" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Post" }));

    await waitFor(() =>
      expect(onCreate).toHaveBeenCalledWith({
        id: "post-1",
        content: "A useful update",
      })
    );
    const [, request] = vi.mocked(apiFetch).mock.calls[0];
    expect(apiFetch).toHaveBeenCalledWith("/api/posts", expect.any(Object));
    expect(request).toMatchObject({ method: "POST" });
    expect(request.body).toBeInstanceOf(FormData);
    expect(request.body.get("content")).toBe("A useful update");
    expect(request.body.get("privacy")).toBe("almost_private");
    expect(screen.getByPlaceholderText("What's on your mind?")).toHaveValue("");
    expect(screen.getByRole("combobox")).toHaveValue("public");
  });

  it("requires and submits private audience ids from accepted followers", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce([{ id: "user-2", nickname: "bee" }])
      .mockResolvedValueOnce({ id: "post-2" });

    renderWithProviders(<NewPost onCreate={vi.fn()} />);

    fireEvent.change(screen.getByPlaceholderText("What's on your mind?"), {
      target: { value: "Private update" },
    });
    fireEvent.change(screen.getByRole("combobox"), {
      target: { value: "private" },
    });

    expect(await screen.findByLabelText("bee")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Post" }));
    expect(screen.getByText("Choose at least one follower for a private post.")).toBeInTheDocument();

    fireEvent.click(screen.getByLabelText("bee"));
    fireEvent.click(screen.getByRole("button", { name: "Post" }));

    await waitFor(() => expect(apiFetch).toHaveBeenCalledWith("/api/posts", expect.any(Object)));
    const [, request] = vi.mocked(apiFetch).mock.calls[1];
    expect(request.body.getAll("audience_ids")).toEqual(["user-2"]);
  });

  it("sends group_id without ordinary privacy controls in group mode", async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce({ id: "post-3", group_id: "group-1" });

    renderWithProviders(<NewPost groupId="group-1" onCreate={vi.fn()} />);

    fireEvent.change(screen.getByPlaceholderText("Share with this group"), {
      target: { value: "Group update" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Post" }));

    await waitFor(() => expect(apiFetch).toHaveBeenCalledWith("/api/posts", expect.any(Object)));
    const [, request] = vi.mocked(apiFetch).mock.calls[0];
    expect(request.body.get("group_id")).toBe("group-1");
    expect(request.body.has("privacy")).toBe(false);
  });

  it("preserves input when the API rejects submission", async () => {
    vi.mocked(apiFetch).mockRejectedValueOnce(new Error("No permission"));

    renderWithProviders(<NewPost />);

    fireEvent.change(screen.getByPlaceholderText("What's on your mind?"), {
      target: { value: "Do not lose this" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Post" }));

    expect(await screen.findByText("No permission")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("What's on your mind?")).toHaveValue("Do not lose this");
  });
});
