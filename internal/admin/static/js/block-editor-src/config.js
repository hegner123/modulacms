// config.js — Block editor constants

export const MAX_DEPTH = 8;

// Capitalize a block type string for display (e.g. "heading" -> "Heading").
export function typeLabel(type) {
        if (!type) return 'Block';
        return type.charAt(0).toUpperCase() + type.slice(1);
}
