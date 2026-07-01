import { describe, expect, it, vi } from "vitest";
import { logger, sanitizeForLog } from "./logger.js";

describe("logger", () => {
  it("redacts sensitive values from logged context", () => {
    const output = sanitizeForLog({
      postId: "post-1",
      content: "private comment",
      Authorization: "Bearer secret",
      nested: { body: "raw payload", safe: "ok" },
    });

    expect(output).toEqual({
      postId: "post-1",
      content: "[redacted]",
      Authorization: "[redacted]",
      nested: { body: "[redacted]", safe: "ok" },
    });
  });

  it("logs sanitized errors and context", () => {
    const spy = vi.spyOn(console, "error").mockImplementation(() => {});

    logger.error("Failed", new Error("boom"), {
      password: "secret",
      commentId: "comment-1",
    });

    expect(spy).toHaveBeenCalledWith(
      "Failed",
      expect.objectContaining({ name: "Error", message: "boom" }),
      { password: "[redacted]", commentId: "comment-1" }
    );
    spy.mockRestore();
  });
});
