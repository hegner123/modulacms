/**
 * Error thrown by {@link ModulaClient} on non-2xx API responses
 * or client-side validation failures.
 *
 * @example
 * try {
 *   const page = await cms.getPage("about");
 * } catch (err) {
 *   if (err instanceof ModulaError) {
 *     console.log(err.status);        // 404
 *     console.log(err.errorMessage);   // "Not found"
 *     console.log(err.body);           // full response body
 *   }
 * }
 */
export class ModulaError extends Error {
  /** HTTP status code from the API response. 0 for client-side errors (e.g. invalid config). */
  readonly status: number;
  /** Raw response body. May be a parsed JSON object or a plain text string. */
  readonly body: unknown;

  /**
   * @param status - HTTP status code, or 0 for client-side errors.
   * @param body - The response body. If it contains an `error` property, that value is used as the error message.
   */
  constructor(status: number, body: unknown) {
    const message = typeof body === "object" && body !== null && "error" in body
      ? String(body.error)
      : `Request failed with status ${status}`;

    super(message);
    this.name = "ModulaError";
    this.status = status;
    this.body = body;
  }

  /**
   * Convenience accessor that extracts the `error` string from the response body.
   * Returns `undefined` if the body does not contain an `error` property.
   */
  get errorMessage(): string | undefined {
    if (typeof this.body === "object" && this.body !== null && "error" in this.body) {
      return String(this.body.error);
    }
    return undefined;
  }
}
