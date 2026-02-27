// cache.js — Datatype fetch cache (shared across all <block-editor> instances)

const _dtCache = {
        data: null,       // Array of {id, label, type} from API
        fetchedAt: 0,     // timestamp ms
        ttl: 5 * 60 * 1000, // 5 minutes
        pending: null,    // in-flight promise to deduplicate concurrent fetches
};

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
                                return { id: dt.datatype_id, label: dt.label, type: dt.type };
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
