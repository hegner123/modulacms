// config.js — Block type configuration and constants

export const BLOCK_TYPE_CONFIG = {
        text: { label: 'Text', canHaveChildren: false },
        heading: { label: 'Heading', canHaveChildren: false },
        image: { label: 'Image', canHaveChildren: false },
        container: { label: 'Container', canHaveChildren: true },
};

export function getTypeConfig(type) {
        if (BLOCK_TYPE_CONFIG[type]) return BLOCK_TYPE_CONFIG[type];
        // Fallback for database-defined types: assume they can have children
        return { label: type, canHaveChildren: true };
}

export const MAX_DEPTH = 8;
