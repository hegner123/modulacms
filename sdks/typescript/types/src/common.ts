/**
 * Common primitive type aliases used across ModulaCMS entities.
 *
 * @module common
 */

/**
 * A ULID (Universally Unique Lexicographically Sortable Identifier) string.
 * All entity primary keys in ModulaCMS use this format.
 *
 * @example "01HXK4N2F8QZJV3K7M1Y9ABCDE"
 */
export type ULID = string;

/**
 * An ISO 8601 / RFC 3339 UTC timestamp string as returned by the API.
 *
 * @example "2026-01-30T12:00:00Z"
 */
export type Timestamp = string;

/** A string value that may be null when the API field is not set. */
export type NullableString = string | null;

/** A numeric value that may be null when the API field is not set. */
export type NullableNumber = number | null;
