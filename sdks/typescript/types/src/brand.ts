/**
 * Nominal (branded) type utility that adds a compile-time tag to a base type.
 * Prevents accidental assignment between structurally identical but semantically
 * different types (e.g. `UserID` vs `RoleID`). The brand is erased at runtime
 * and carries zero cost.
 *
 * @typeParam T - The underlying primitive type.
 * @typeParam B - A unique string literal that distinguishes this brand.
 *
 * @example
 * ```ts
 * type OrderID = Brand<string, 'OrderID'>
 * const id = 'abc' as OrderID
 * ```
 */
export type Brand<T, B extends string> = T & { readonly __brand: B }
