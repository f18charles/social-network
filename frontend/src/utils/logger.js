const REDACTED = "[redacted]";
const MAX_DEPTH = 4;
const SENSITIVE_KEY_PATTERN =
  /(password|passcode|token|secret|cookie|authorization|session|email|content|body|message|image|avatar|file)/i;

const isPlainObject = (value) =>
  Object.prototype.toString.call(value) === "[object Object]";

/**
 * sanitizeForLog returns a redacted clone of values before they reach console logs.
 * It avoids logging user content, credentials, cookies, files, and raw request bodies.
 */
export const sanitizeForLog = (value, depth = 0) => {
  if (value == null) return value;
  if (depth > MAX_DEPTH) return "[max-depth]";
  if (value instanceof Error) return sanitizeError(value);
  if (value instanceof FormData) return "[form-data]";
  if (typeof File !== "undefined" && value instanceof File) return "[file]";
  if (typeof Blob !== "undefined" && value instanceof Blob) return "[blob]";
  if (typeof value === "function") return "[function]";
  if (typeof value !== "object") return value;
  if (Array.isArray(value)) {
    return value.slice(0, 20).map((item) => sanitizeForLog(item, depth + 1));
  }
  if (!isPlainObject(value)) {
    return Object.prototype.toString.call(value);
  }

  return Object.fromEntries(
    Object.entries(value).map(([key, entry]) => [
      key,
      SENSITIVE_KEY_PATTERN.test(key)
        ? REDACTED
        : sanitizeForLog(entry, depth + 1),
    ])
  );
};

/**
 * sanitizeError keeps diagnostic error fields without exposing custom payload data.
 */
export const sanitizeError = (error) => {
  if (!error) return null;
  return {
    name: error.name || "Error",
    message: error.message || "",
    status: typeof error.status === "number" ? error.status : undefined,
    statusText: error.statusText || undefined,
    stack: import.meta.env.DEV ? error.stack : undefined,
  };
};

/**
 * logger centralizes frontend diagnostics and always redacts contextual data.
 */
export const logger = {
  error(message, error, context = {}) {
    console.error(message, sanitizeError(error), sanitizeForLog(context));
  },
  warn(message, context = {}) {
    console.warn(message, sanitizeForLog(context));
  },
};

/**
 * installGlobalErrorLogging captures uncaught frontend errors without rendering raw messages.
 */
export const installGlobalErrorLogging = () => {
  window.addEventListener("error", (event) => {
    logger.error(
      "Unhandled frontend error",
      event.error || new Error(event.message),
      {
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
      }
    );
  });

  window.addEventListener("unhandledrejection", (event) => {
    const reason =
      event.reason instanceof Error
        ? event.reason
        : new Error(String(event.reason));
    logger.error("Unhandled frontend promise rejection", reason);
  });
};
