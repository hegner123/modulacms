let counter = 0;

/** Set or remove aria-invalid on an element. */
export function setAriaInvalid(el: HTMLElement, invalid: boolean): void {
  if (invalid) {
    el.setAttribute("aria-invalid", "true");
  } else {
    el.removeAttribute("aria-invalid");
  }
}

/** Set or remove aria-describedby on an element. */
export function setAriaDescribedBy(
  el: HTMLElement,
  id: string | null,
): void {
  if (id) {
    el.setAttribute("aria-describedby", id);
  } else {
    el.removeAttribute("aria-describedby");
  }
}

/**
 * Announce a message to screen readers via a temporary live region.
 * The region is inserted, populated, and removed after a short delay.
 */
export function announceToScreenReader(message: string): void {
  const region = document.createElement("div");
  region.setAttribute("role", "status");
  region.setAttribute("aria-live", "polite");
  region.setAttribute("aria-atomic", "true");
  region.style.position = "absolute";
  region.style.width = "1px";
  region.style.height = "1px";
  region.style.overflow = "hidden";
  region.style.clip = "rect(0 0 0 0)";
  region.style.whiteSpace = "nowrap";
  document.body.appendChild(region);

  // Delay text insertion so the live region is registered first
  requestAnimationFrame(() => {
    region.textContent = message;
  });

  setTimeout(() => {
    region.remove();
  }, 3000);
}

/** Generate a unique ID with the given prefix and an incrementing counter. */
export function generateId(prefix: string): string {
  counter++;
  return `${prefix}${counter}`;
}
