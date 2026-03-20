/**
 * Dispatch a CustomEvent that bubbles and crosses shadow DOM boundaries.
 * Returns the result of dispatchEvent (false if preventDefault was called
 * on a cancelable event).
 */
export function dispatch<T = undefined>(
  el: HTMLElement,
  name: string,
  detail?: T,
  options?: { cancelable?: boolean },
): boolean {
  const event = new CustomEvent(name, {
    detail: detail as unknown,
    bubbles: true,
    composed: true,
    cancelable: options?.cancelable ?? false,
  });
  return el.dispatchEvent(event);
}
