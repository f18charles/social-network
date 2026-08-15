import { afterEach, describe, expect, it, vi } from "vitest";
import { apiFetch, ApiError } from "./api";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("apiFetch", () => {
  it("includes credentials, serializes JSON, and unwraps success data", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          status: "success",
          message: "Created",
          data: { id: "post-1" },
          errors: null,
        }),
        {
          status: 201,
          headers: { "Content-Type": "application/json" },
        }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      apiFetch("/api/posts", {
        method: "POST",
        body: { content: "Hello" },
      })
    ).resolves.toEqual({ id: "post-1" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/posts",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ content: "Hello" }),
        credentials: "include",
      })
    );
    expect(fetchMock.mock.calls[0][1].headers.get("Content-Type")).toBe(
      "application/json"
    );
  });

  it("passes FormData through without setting a content type", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(null, {
        status: 204,
      })
    );
    vi.stubGlobal("fetch", fetchMock);
    const body = new FormData();
    body.append("content", "Post with an image");

    await expect(
      apiFetch("/api/posts", {
        method: "POST",
        body,
      })
    ).resolves.toBeNull();

    const request = fetchMock.mock.calls[0][1];
    expect(request.body).toBe(body);
    expect(request.headers.has("Content-Type")).toBe(false);
  });

  it("normalizes non-success responses as ApiError", async () => {
    const responseBody = {
      status: "error",
      message: "Validation failed",
      data: null,
      errors: { content: "is required" },
    };
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify(responseBody), {
          status: 422,
          statusText: "Unprocessable Content",
          headers: { "Content-Type": "application/json" },
        })
      )
    );

    const request = apiFetch("/api/posts", {
      method: "POST",
      body: {},
    });

    await expect(request).rejects.toMatchObject({
      name: "ApiError",
      message: "Validation failed",
      status: 422,
      data: responseBody,
    });
    await expect(request).rejects.toBeInstanceOf(ApiError);
  });

  it("resolves relative and absolute URLs correctly with VITE_API_URL", () => {
    import("./api").then(({ resolveApiUrl }) => {
      expect(resolveApiUrl("/api/users")).toBe("/api/users");
      expect(resolveApiUrl("https://example.com/api/users")).toBe(
        "https://example.com/api/users"
      );
    });
  });
});
