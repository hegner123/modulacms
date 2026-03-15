// <mcms-media-tree> -- Media folder tree with drag-and-drop support (Light DOM)
// Enables dragging media grid items onto folder links to move them.
// Also handles folder rename via data attributes on rename buttons.
// Dispatches: media-tree:move custom event on successful drop.

(function() {
    // Drag-and-drop: media items onto folders
    document.addEventListener('dragstart', function(e) {
        var item = e.target.closest('[data-media-grid-item][data-media-id]');
        if (!item) return;
        var mediaId = item.getAttribute('data-media-id');
        if (!mediaId) return;
        e.dataTransfer.setData('text/plain', mediaId);
        e.dataTransfer.effectAllowed = 'move';
        item.setAttribute('data-dragging', '');
        item.classList.add('dragging'); // DUAL: data-dragging + class
    });

    document.addEventListener('dragend', function(e) {
        var item = e.target.closest('[data-media-grid-item][data-media-id]');
        if (item) {
            item.removeAttribute('data-dragging');
            item.classList.remove('dragging'); // DUAL: data-dragging + class
        }
        // Clean up all drop targets
        var targets = document.querySelectorAll('[data-drop-target]');
        for (var i = 0; i < targets.length; i++) {
            targets[i].removeAttribute('data-drop-target');
            targets[i].classList.remove('media-folder-drop-target'); // DUAL: data-drop-target + class
        }
    });

    document.addEventListener('dragover', function(e) {
        var folderLink = e.target.closest('[data-folder-link], [data-folder-root]');
        if (!folderLink) return;
        e.preventDefault();
        e.dataTransfer.dropEffect = 'move';
        folderLink.setAttribute('data-drop-target', '');
        folderLink.classList.add('media-folder-drop-target'); // DUAL: data-drop-target + class
    });

    document.addEventListener('dragleave', function(e) {
        var folderLink = e.target.closest('[data-folder-link], [data-folder-root]');
        if (folderLink) {
            folderLink.removeAttribute('data-drop-target');
            folderLink.classList.remove('media-folder-drop-target'); // DUAL: data-drop-target + class
        }
    });

    document.addEventListener('drop', function(e) {
        var folderLink = e.target.closest('[data-folder-link], [data-folder-root]');
        if (!folderLink) return;
        e.preventDefault();
        folderLink.removeAttribute('data-drop-target');
        folderLink.classList.remove('media-folder-drop-target'); // DUAL: data-drop-target + class

        var mediaId = e.dataTransfer.getData('text/plain');
        if (!mediaId) return;

        var folderId = folderLink.getAttribute('data-folder-id') || '';

        // Get CSRF token
        var meta = document.querySelector('meta[name="csrf-token"]');
        var csrfToken = meta ? meta.content : '';

        // Build form data
        var formData = new FormData();
        formData.append('folder_id', folderId);
        formData.append('_csrf', csrfToken);

        fetch('/admin/media/move/' + encodeURIComponent(mediaId), {
            method: 'POST',
            headers: {
                'X-CSRF-Token': csrfToken,
                'HX-Request': 'true'
            },
            body: formData,
            credentials: 'same-origin'
        }).then(function(res) {
            if (res.ok) {
                var toast = document.querySelector('mcms-toast');
                if (toast) toast.show('Media moved to folder', 'success');
                // Remove the dragged item from the grid
                var item = document.getElementById('media-item-' + mediaId);
                if (item) item.remove();
                // Dispatch custom event
                document.dispatchEvent(new CustomEvent('media-tree:move', {
                    bubbles: true,
                    detail: { mediaId: mediaId, folderId: folderId }
                }));
            } else {
                var toast = document.querySelector('mcms-toast');
                if (toast) toast.show('Failed to move media', 'error');
            }
        }).catch(function() {
            var toast = document.querySelector('mcms-toast');
            if (toast) toast.show('Network error while moving media', 'error');
        });
    });

    // Folder rename via data-attribute buttons
    document.addEventListener('click', function(e) {
        var btn = e.target.closest('[data-folder-rename]');
        if (!btn) return;
        var folderId = btn.getAttribute('data-folder-id');
        var currentName = btn.getAttribute('data-folder-name');
        if (!folderId) return;

        var newName = prompt('Rename folder:', currentName || '');
        if (newName === null || newName.trim() === '' || newName.trim() === currentName) return;

        var meta = document.querySelector('meta[name="csrf-token"]');
        var csrfToken = meta ? meta.content : '';

        var formData = new FormData();
        formData.append('name', newName.trim());
        formData.append('_csrf', csrfToken);

        fetch('/admin/media-folders/' + encodeURIComponent(folderId), {
            method: 'POST',
            headers: {
                'X-CSRF-Token': csrfToken,
                'HX-Request': 'true'
            },
            body: formData,
            credentials: 'same-origin'
        }).then(function(res) {
            if (res.ok) {
                return res.text();
            }
            var toast = document.querySelector('mcms-toast');
            if (toast) toast.show('Failed to rename folder', 'error');
            return null;
        }).then(function(html) {
            if (html) {
                var sidebar = document.getElementById('media-folder-sidebar');
                if (sidebar) {
                    var temp = document.createElement('div');
                    temp.innerHTML = html;
                    var newSidebar = temp.firstElementChild;
                    if (newSidebar) {
                        sidebar.replaceWith(newSidebar);
                        // Re-initialize lucide icons in new content
                        if (typeof lucide !== 'undefined') lucide.createIcons();
                        // Process HTMX on new content
                        if (typeof htmx !== 'undefined') htmx.process(newSidebar);
                    }
                }
                var toast = document.querySelector('mcms-toast');
                if (toast) toast.show('Folder renamed', 'success');
            }
        }).catch(function() {
            var toast = document.querySelector('mcms-toast');
            if (toast) toast.show('Network error while renaming folder', 'error');
        });
    });

    // Re-initialize lucide icons after HTMX swaps that include folder tree
    document.body.addEventListener('htmx:afterSwap', function(e) {
        if (e.detail.target && (
            e.detail.target.id === 'media-folder-sidebar' ||
            e.detail.target.id === 'main-content'
        )) {
            if (typeof lucide !== 'undefined') {
                setTimeout(function() { lucide.createIcons(); }, 10);
            }
        }
    });
})();
