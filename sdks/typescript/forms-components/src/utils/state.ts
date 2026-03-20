/**
 * Lightweight reactive state container.
 * Holds a single value and notifies subscribers on change.
 */
export class State<T> {
  private value: T;
  private listeners: Set<(value: T) => void> = new Set();

  constructor(initial: T) {
    this.value = initial;
  }

  /** Return the current value. */
  get(): T {
    return this.value;
  }

  /**
   * Set the next value. Accepts a direct value or an updater function
   * that receives the previous value and returns the next.
   * Notifies all subscribers after the value changes.
   */
  set(next: T | ((prev: T) => T)): void {
    if (typeof next === "function") {
      this.value = (next as (prev: T) => T)(this.value);
    } else {
      this.value = next;
    }
    for (const fn of this.listeners) {
      fn(this.value);
    }
  }

  /**
   * Subscribe to value changes. Returns an unsubscribe function.
   */
  subscribe(fn: (value: T) => void): () => void {
    this.listeners.add(fn);
    return () => {
      this.listeners.delete(fn);
    };
  }
}
