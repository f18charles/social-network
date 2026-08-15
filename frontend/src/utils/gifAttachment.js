import { logger } from "./logger.js";

/**
 * Fetches a GIF from a remote URL (e.g. a pick from GifPicker) and wraps it
 * as a File, so it can be attached through the same multipart image upload
 * path used for regular image attachments. On the backend this means the
 * GIF ends up stored in Cloudinary (or local disk in dev) alongside other
 * post/comment images, rather than just being pasted as a raw URL.
 *
 * Returns null if the GIF could not be fetched (e.g. blocked by CORS or a
 * network error), so callers can fall back to simpler behavior.
 */
export async function fetchGifAsFile(gifUrl) {
  try {
    const response = await fetch(gifUrl);
    if (!response.ok) {
      throw new Error(`Unexpected status ${response.status}`);
    }
    const blob = await response.blob();
    const type = blob.type && blob.type.startsWith("image/") ? blob.type : "image/gif";
    const extension = type.split("/")[1] || "gif";
    const filename = `gif-${Date.now()}.${extension}`;
    return new File([blob], filename, { type });
  } catch (err) {
    logger.error("Failed to fetch GIF for attachment", err);
    return null;
  }
}
