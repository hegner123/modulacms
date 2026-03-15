import { describe, it, expect } from 'vitest'
import type { JsonValue } from './json.js'

describe('JsonValue', () => {
  it('accepts string', () => {
    const v: JsonValue = "hello"
    expect(v).toBe("hello")
  })

  it('accepts number', () => {
    const v: JsonValue = 42
    expect(v).toBe(42)
  })

  it('accepts boolean', () => {
    const v: JsonValue = true
    expect(v).toBe(true)
  })

  it('accepts null', () => {
    const v: JsonValue = null
    expect(v).toBeNull()
  })

  it('accepts array of primitives', () => {
    const v: JsonValue = [1, "two", true, null]
    expect(v).toEqual([1, "two", true, null])
  })

  it('accepts nested object', () => {
    const v: JsonValue = {
      name: "test",
      count: 5,
      active: false,
      tags: ["a", "b"],
      nested: {
        deep: {
          value: 42
        }
      }
    }
    expect(v).toHaveProperty('nested.deep.value', 42)
  })

  it('accepts array of objects', () => {
    const v: JsonValue = [
      { id: 1, name: "first" },
      { id: 2, name: "second" },
    ]
    expect(Array.isArray(v)).toBe(true)
    expect(v).toHaveLength(2)
  })

  it('accepts empty object', () => {
    const v: JsonValue = {}
    expect(v).toEqual({})
  })

  it('accepts empty array', () => {
    const v: JsonValue = []
    expect(v).toEqual([])
  })

  it('round-trips through JSON.stringify and JSON.parse', () => {
    const original: JsonValue = {
      str: "hello",
      num: 42,
      bool: true,
      nil: null,
      arr: [1, 2, 3],
      obj: { nested: "value" },
    }
    const serialized = JSON.stringify(original)
    const parsed: JsonValue = JSON.parse(serialized) as JsonValue
    expect(parsed).toEqual(original)
  })
})
