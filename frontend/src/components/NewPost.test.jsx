import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import NewPost from "./NewPost";
import { renderWithProviders } from "../test/render";

describe("NewPost", () => {
  beforeEach(() => {
    global.fetch = vi.fn();
    vi.spyOn(console, "error").mockImplementation(() => {});
    vi.stubGlobal("URL", {
      ...URL,
      createObjectURL: vi.fn(() => "blob:preview"),
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
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: { id: "post-1", content: "A useful update" } }),
    });

    renderWithProviders(<NewPost onCreate={onCreate} />);

    fireEvent.change(screen.getByPlaceholderText("What's on your mind?"), {
      target: { value: "A useful update" },
    });
    fireEvent.change(screen.getByRole("combobox"), {
      target: { value: "almost_private" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Post" }));

    await waitFor(() => expect(onCreate).toHaveBeenCalledWith({ id: "post-1", content: "A useful update" }));
    const [, request] = vi.mocked(fetch).mock.calls[0];
    expect(request).toMatchObject({ method: "POST", credentials: "include" });
    expect(request.body).toBeInstanceOf(FormData);
    expect(request.body.get("content")).toBe("A useful update");
    expect(request.body.get("privacy")).toBe("almost_private");
    expect(screen.getByPlaceholderText("What's on your mind?")).toHaveValue("");
    expect(screen.getByRole("combobox")).toHaveValue("public");
  });
});
