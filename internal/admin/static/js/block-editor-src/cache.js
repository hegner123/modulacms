// cache.js — Datatype fetch cache (shared across all <block-editor> instances)

const _dtCache = {
        data: null,       // Array of {id, parentId, name, label, type} from API
        fetchedAt: 0,     // timestamp ms
        ttl: 5 * 60 * 1000, // 5 minutes
        pending: null,    // in-flight promise to deduplicate concurrent fetches
};

// System types that are not selectable blocks
var SYSTEM_TYPES = { '_root': true, '_nested_root': true, '_system_log': true };

export function fetchDatatypes() {
        var now = Date.now();
        if (_dtCache.data && (now - _dtCache.fetchedAt) < _dtCache.ttl) {
                return Promise.resolve(_dtCache.data);
        }
        if (_dtCache.pending) return _dtCache.pending;

        _dtCache.pending = fetch('/admin/api/datatypes', { credentials: 'same-origin' })
                .then(function(res) {
                        if (!res.ok) throw new Error('Failed to fetch datatypes: ' + res.status);
                        return res.json();
                })
                .then(function(datatypes) {
                        _dtCache.data = datatypes.map(function(dt) {
                                return { id: dt.datatype_id, parentId: dt.parent_id || null, name: dt.name, label: dt.label, type: dt.type };
                        });
                        _dtCache.fetchedAt = Date.now();
                        _dtCache.pending = null;
                        return _dtCache.data;
                })
                .catch(function(err) {
                        _dtCache.pending = null;
                        throw err;
                });

        return _dtCache.pending;
}

/**
 * Fetch datatypes and group them into categories for the block picker.
 *
 * Categories:
 *   1. Root type children (if rootDatatypeId provided)
 *   2. "Collections" — children of _collection datatypes
 *   3. "Global" — _global datatypes and their children
 *
 * Returns { categories: [{ name, items: [{ id, label, type, depth }] }] }
 */
export function fetchDatatypesGrouped(rootDatatypeId) {
        return fetchDatatypes().then(function(datatypes) {
                // Build parent -> children map
                var childrenOf = {};
                var byId = {};
                for (var i = 0; i < datatypes.length; i++) {
                        var dt = datatypes[i];
                        byId[dt.id] = dt;
                        var pid = dt.parentId || '_none';
                        if (!childrenOf[pid]) childrenOf[pid] = [];
                        childrenOf[pid].push(dt);
                }

                // Recursively collect children as flat list with depth
                function collectChildren(parentId, baseDepth) {
                        var result = [];
                        var kids = childrenOf[parentId];
                        if (!kids) return result;
                        for (var j = 0; j < kids.length; j++) {
                                var kid = kids[j];
                                if (SYSTEM_TYPES[kid.type]) continue;
                                result.push({ id: kid.id, name: kid.name, label: kid.label, type: kid.type, depth: baseDepth });
                                var grandchildren = collectChildren(kid.id, baseDepth + 1);
                                for (var k = 0; k < grandchildren.length; k++) {
                                        result.push(grandchildren[k]);
                                }
                        }
                        return result;
                }

                var categories = [];

                // Category 1: Root type children
                if (rootDatatypeId && byId[rootDatatypeId]) {
                        var rootDt = byId[rootDatatypeId];
                        var rootItems = collectChildren(rootDatatypeId, 0);
                        if (rootItems.length > 0) {
                                categories.push({ name: rootDt.label, items: rootItems });
                        }
                }

                // Category 2: Collections — find _collection parents, collect their children
                var collectionItems = [];
                for (var ci = 0; ci < datatypes.length; ci++) {
                        if (datatypes[ci].type === '_collection') {
                                var kids2 = collectChildren(datatypes[ci].id, 0);
                                for (var ck = 0; ck < kids2.length; ck++) {
                                        collectionItems.push(kids2[ck]);
                                }
                        }
                }
                if (collectionItems.length > 0) {
                        categories.push({ name: 'Collections', items: collectionItems });
                }

                // Category 3: Global — _global datatypes themselves + their children
                var globalItems = [];
                for (var gi = 0; gi < datatypes.length; gi++) {
                        var gdt = datatypes[gi];
                        if (gdt.type === '_global') {
                                globalItems.push({ id: gdt.id, name: gdt.name, label: gdt.label, type: gdt.type, depth: 0 });
                                var gkids = collectChildren(gdt.id, 1);
                                for (var gk = 0; gk < gkids.length; gk++) {
                                        globalItems.push(gkids[gk]);
                                }
                        }
                }
                if (globalItems.length > 0) {
                        categories.push({ name: 'Global', items: globalItems });
                }

                return { categories: categories };
        });
}
