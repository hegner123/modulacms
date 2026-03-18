/**
 * mcms.settings — Admin panel UI preference persistence.
 *
 * Thin wrapper over localStorage with namespaced keys.
 * All UI preference reads/writes go through this module.
 * Internals can be swapped to server-sync later without changing call sites.
 *
 * Usage:
 *   mcms.settings.get('media.gridSize', '3')
 *   mcms.settings.set('media.gridSize', '2')
 *   mcms.settings.remove('media.gridSize')
 *   mcms.settings.getAll()  // { 'media.gridSize': '2', ... }
 */

(function() {
    var PREFIX = 'mcms.';

    var settings = {
        /**
         * Get a setting value.
         * @param {string} key - Dot-separated key (e.g. 'media.gridSize')
         * @param {string} [defaultValue] - Fallback if not set
         * @returns {string|null}
         */
        get: function(key, defaultValue) {
            try {
                var val = localStorage.getItem(PREFIX + key);
                return val !== null ? val : (defaultValue !== undefined ? defaultValue : null);
            } catch (e) {
                return defaultValue !== undefined ? defaultValue : null;
            }
        },

        /**
         * Set a setting value.
         * @param {string} key - Dot-separated key
         * @param {string} value - Value to store
         */
        set: function(key, value) {
            try {
                localStorage.setItem(PREFIX + key, String(value));
            } catch (e) {
                // Storage full or unavailable — fail silently
            }
        },

        /**
         * Remove a setting.
         * @param {string} key - Dot-separated key
         */
        remove: function(key) {
            try {
                localStorage.removeItem(PREFIX + key);
            } catch (e) {}
        },

        /**
         * Get all settings as a plain object.
         * @returns {Object}
         */
        getAll: function() {
            var result = {};
            try {
                for (var i = 0; i < localStorage.length; i++) {
                    var k = localStorage.key(i);
                    if (k && k.indexOf(PREFIX) === 0) {
                        result[k.substring(PREFIX.length)] = localStorage.getItem(k);
                    }
                }
            } catch (e) {}
            return result;
        }
    };

    // Expose globally
    window.mcms = window.mcms || {};
    window.mcms.settings = settings;
})();
