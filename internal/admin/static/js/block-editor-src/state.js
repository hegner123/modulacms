// state.js — State factory

export function createState() {
        return {
                blocks: {},
                rootId: null,
                selectedBlockId: null,
                dirty: false,
        };
}
