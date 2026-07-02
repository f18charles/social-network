import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import CommentComposer from "./CommentComposer.jsx";
import { renderWithProviders } from "../test/render";

describe("CommentComposer", () => {
  beforeEach(() => {
    vi.stubGlobal("URL", {
      ...URL,
      createObjectURL: vi.fn(() => "blob:comment-preview"),
      revokeObjectURL: vi.fn(),
    });
  });

  it("submits with Enter and keeps Shift+Enter as text input", async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    renderWithProviders(<CommentComposer onSubmit={onSubmit} />);

    const textarea = screen.getByLabelText("Write a comment");
    fireEvent.change(textarea, { target: { value: "Line one" } });
    fireEvent.keyDown(textarea, { key: "Enter", shiftKey: true });
    expect(onSubmit).not.toHaveBeenCalled();

    fireEvent.keyDown(textarea, { key: "Enter" });

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    const formData = onSubmit.mock.calls[0][0];
    expect(formData.get("content")).toBe("Line one");
  });

  it("previews media and revokes object URLs on removal", () => {
    const file = new File(["gif"], "reply.gif", { type: "image/gif" });
    renderWithProviders(<CommentComposer onSubmit={vi.fn()} />);

    fireEvent.change(screen.getByLabelText("Write a comment").form.querySelector('input[type="file"]'), {
      target: { files: [file] },
    });

    expect(screen.getByAltText("Selected comment attachment preview")).toHaveAttribute(
      "src",
      "blob:comment-preview"
    );

    fireEvent.click(screen.getByRole("button", { name: "Remove" }));

    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:comment-preview");
    expect(screen.queryByAltText("Selected comment attachment preview")).not.toBeInTheDocument();
  });
});
