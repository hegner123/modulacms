/** JSON-serializable value. Prevents passing functions, symbols, or circular structures.
 * Note: Date objects are NOT JsonValue — call .toISOString() before passing dates as request bodies. */
export type JsonValue = string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue }
