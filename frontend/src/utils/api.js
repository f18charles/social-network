"use strict";

/**
 * Error thrown when an API request fails.
 */
export class ApiError extends Error {
  /**
   * @param {string} message Human-readable error message suitable for rendering.
   * @param {object} [details] Additional response details.
   * @param {number} [details.status=0] HTTP status code, or 0 for network errors.
   * @param {string} [details.statusText=""] HTTP status text.
   * @param {*} [details.data=null] Parsed response body returned by the API.
   */
  constructor(message, { status = 0, statusText = "", data = null } = {}) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.statusText = statusText;
    this.data = data;
  }
}

/**
 * Parses a response body according to its content type.
 *
 * @param {Response} response Fetch response to parse.
 * @returns {Promise<*>} Parsed JSON, response text, or null for an empty body.
 */
const parseResponse = async (response) => {
  if (response.status === 204) {
    return null;
  }

  const contentType = response.headers.get("content-type") || "";

  if (contentType.includes("application/json")) {
    return response.json();
  }

  const text = await response.text();
  return text || null;
};

/**
 * Extracts a renderable error message from an API response.
 *
 * @param {*} data Parsed response body.
 * @param {Response} response Failed fetch response.
 * @returns {string} Human-readable error message.
 */
const getErrorMessage = (data, response) => {
  if (typeof data === "string" && data.trim()) {
    return data.trim();
  }

  if (data && typeof data === "object") {
    return data.message || data.error || response.statusText;
  }

  return response.statusText || "Request failed";
};

/**
 * Resolves the target API URL, prepending VITE_API_URL if configured for cross-origin deployments.
 *
 * @param {string|URL} endpoint
 * @returns {string|URL}
 */
export const resolveApiUrl = (endpoint) => {
  if (typeof endpoint !== "string" || endpoint.startsWith("http://") || endpoint.startsWith("https://")) {
    return endpoint;
  }
  const baseUrl = import.meta.env.VITE_API_URL || "";
  if (baseUrl) {
    const cleanBase = baseUrl.replace(/\/+$/, "");
    const cleanEndpoint = endpoint.startsWith("/") ? endpoint : `/${endpoint}`;
    return `${cleanBase}${cleanEndpoint}`;
  }
  return endpoint;
};

/**
 * Sends an API request with session cookies and parses the response.
 *
 * Plain object bodies are serialized as JSON. FormData bodies are passed through
 * unchanged so the browser can set the multipart content type and boundary.
 *
 * @param {string|URL} url API endpoint.
 * @param {RequestInit & {body?: *}} [options={}] Fetch options.
 * Standard success envelopes are unwrapped to return their `data` payload.
 *
 * @returns {Promise<*>} Response payload, parsed body, or null when empty.
 * @throws {ApiError} When the network request or API response fails.
 */
export const apiFetch = async (url, options = {}) => {
  const { body, headers: customHeaders, ...requestOptions } = options;
  const headers = new Headers(customHeaders);
  const isFormData = body instanceof FormData;
  let requestBody = body;

  if (body != null && !isFormData) {
    if (!headers.has("Content-Type")) {
      headers.set("Content-Type", "application/json");
    }

    if (typeof body !== "string") {
      requestBody = JSON.stringify(body);
    }
  }

  const targetUrl = resolveApiUrl(url);
  let response;

  try {
    response = await fetch(targetUrl, {
      ...requestOptions,
      headers,
      body: requestBody,
      credentials: "include",
    });
  } catch (error) {
    throw new ApiError(error.message || "Unable to connect to the server");
  }

  let data;

  try {
    data = await parseResponse(response);
  } catch {
    throw new ApiError("The server returned an invalid response", {
      status: response.status,
      statusText: response.statusText,
    });
  }

  if (!response.ok) {
    throw new ApiError(getErrorMessage(data, response), {
      status: response.status,
      statusText: response.statusText,
      data,
    });
  }

  if (
    data &&
    typeof data === "object" &&
    data.status === "success" &&
    Object.prototype.hasOwnProperty.call(data, "data") &&
    !Object.prototype.hasOwnProperty.call(data, "pagination")
  ) {
    return data.data;
  }

  return data;
};
