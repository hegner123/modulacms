/**
 * <mcms-publish-button> -- Publish action button with dropdown.
 *
 * When status is "draft": button says "Publish". Dropdown shows "Unpublish" disabled.
 * When status is "published": button says "Update". Dropdown shows "Unpublish" enabled.
 *
 * Attributes:
 *   status        - Current status: "published" or "draft"
 *   publish-url   - URL for publish/update action (POST)
 *   unpublish-url - URL for unpublish action (POST)
 *   target        - HTMX swap target (default: #main-content)
 */
class McmsPublishButton extends HTMLElement {
    connectedCallback() {
        this._status = this.getAttribute('status') || 'draft';
        this._publishUrl = this.getAttribute('publish-url') || '';
        this._unpublishUrl = this.getAttribute('unpublish-url') || '';
        this._target = this.getAttribute('target') || '#main-content';
        this._open = false;
        this._render();
        this._bind();
    }

    disconnectedCallback() {
        if (this._outsideClick) {
            document.removeEventListener('click', this._outsideClick);
        }
    }

    _render() {
        var isPublished = this._status === 'published';
        var isDraft = !isPublished;
        var buttonLabel = isPublished ? 'Update' : 'Publish';
        var chevronSvg = '<svg class="size-5 text-white" viewBox="0 0 20 20" fill="currentColor"><path d="M5.22 8.22a.75.75 0 0 1 1.06 0L10 11.94l3.72-3.72a.75.75 0 1 1 1.06 1.06l-4.25 4.25a.75.75 0 0 1-1.06 0L5.22 9.28a.75.75 0 0 1 0-1.06Z" clip-rule="evenodd" fill-rule="evenodd" /></svg>';

        // Publish option: disabled when already draft (primary action handles it)
        var publishClass = isDraft
            ? 'flex w-full items-center gap-x-2 px-4 py-3 text-sm text-gray-500 pointer-events-none'
            : 'flex w-full items-center gap-x-2 px-4 py-3 text-sm text-gray-300 hover:bg-white/5 hover:text-white';

        // Unpublish option: disabled when already draft
        var unpublishClass = isDraft
            ? 'flex w-full items-center gap-x-2 px-4 py-3 text-sm text-gray-500 pointer-events-none'
            : 'flex w-full items-center gap-x-2 px-4 py-3 text-sm text-gray-300 hover:bg-white/5 hover:text-white';

        var html = '<div class="relative inline-flex">'
            + '<div class="inline-flex divide-x divide-[var(--color-primary-hover)] rounded-md">'
            + '<button type="button" class="inline-flex items-center gap-x-1.5 rounded-l-md bg-[var(--color-primary)] px-3 py-2 text-sm font-semibold text-white hover:bg-[var(--color-primary-hover)]" data-publish-action>'
            + buttonLabel
            + '</button>'
            + '<button type="button" class="inline-flex items-center rounded-r-md bg-[var(--color-primary)] p-2 hover:bg-[var(--color-primary-hover)]" data-publish-toggle aria-label="More options">'
            + chevronSvg
            + '</button>'
            + '</div>'
            + '<div class="absolute right-0 top-full z-20 mt-2 hidden w-56 origin-top-right divide-y divide-white/10 overflow-hidden rounded-md bg-gray-800 shadow-lg ring-1 ring-white/10" data-publish-menu>'
            + '<button type="button" class="' + publishClass + '" data-option-publish>'
            + '<svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5"/></svg>'
            + 'Publish'
            + (isPublished ? '<span class="ml-auto text-xs text-gray-500">Current</span>' : '')
            + '</button>'
            + '<button type="button" class="' + unpublishClass + '" data-option-unpublish>'
            + '<svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"/></svg>'
            + 'Unpublish'
            + (isDraft ? '<span class="ml-auto text-xs text-gray-500">Current</span>' : '')
            + '</button>'
            + '</div>'
            + '</div>';

        // Confirmation dialog
        html += '<mcms-dialog data-publish-dialog>'
            + '<div class="p-6">'
            + '<p class="text-sm text-gray-400" data-publish-dialog-msg></p>'
            + '</div>'
            + '<div class="flex justify-end gap-x-3 border-t border-white/10 bg-white/5 px-6 py-3">'
            + '<button type="button" class="rounded-md bg-white/10 px-3 py-2 text-sm font-semibold text-white shadow-xs hover:bg-white/20" data-publish-cancel>Cancel</button>'
            + '<button type="button" class="rounded-md bg-[var(--color-primary)] px-3 py-2 text-sm font-semibold text-white shadow-xs hover:bg-[var(--color-primary-hover)]" data-publish-confirm>Confirm</button>'
            + '</div>'
            + '</mcms-dialog>';

        this.innerHTML = html;
    }

    _bind() {
        var self = this;
        var isPublished = this._status === 'published';
        var actionBtn = this.querySelector('[data-publish-action]');
        var toggle = this.querySelector('[data-publish-toggle]');
        var menu = this.querySelector('[data-publish-menu]');
        var dialog = this.querySelector('[data-publish-dialog]');
        var dialogMsg = this.querySelector('[data-publish-dialog-msg]');
        var cancelBtn = this.querySelector('[data-publish-cancel]');
        var confirmBtn = this.querySelector('[data-publish-confirm]');
        var pendingAction = null;

        // Primary button: Publish (draft) or Update (published)
        actionBtn.addEventListener('click', function() {
            pendingAction = { url: self._publishUrl };
            dialogMsg.textContent = isPublished
                ? 'Update this content? This creates a new public snapshot.'
                : 'Publish this content? This creates a public snapshot.';
            dialog.open();
        });

        // Dropdown toggle
        toggle.addEventListener('click', function(e) {
            e.stopPropagation();
            self._open = !self._open;
            menu.classList.toggle('hidden', !self._open);
        });

        // Publish option (only active when currently draft — but primary button already handles this,
        // so this is for explicitly choosing "Publish" from the dropdown when published/update state)
        var publishOpt = this.querySelector('[data-option-publish]');
        if (publishOpt && isPublished) {
            // When published, "Publish" in dropdown means re-publish/update (same as primary button)
            publishOpt.addEventListener('click', function() {
                menu.classList.add('hidden'); self._open = false;
                pendingAction = { url: self._publishUrl };
                dialogMsg.textContent = 'Update this content? This creates a new public snapshot.';
                dialog.open();
            });
        }

        // Unpublish option (only active when published)
        var unpublishOpt = this.querySelector('[data-option-unpublish]');
        if (unpublishOpt && isPublished) {
            unpublishOpt.addEventListener('click', function() {
                menu.classList.add('hidden'); self._open = false;
                pendingAction = { url: self._unpublishUrl };
                dialogMsg.textContent = 'Unpublish this content? It will no longer be publicly visible.';
                dialog.open();
            });
        }

        // Close dropdown on outside click
        this._outsideClick = function(e) {
            if (!self.contains(e.target)) {
                menu.classList.add('hidden');
                self._open = false;
            }
        };
        document.addEventListener('click', this._outsideClick);

        // Cancel dialog
        cancelBtn.addEventListener('click', function() {
            dialog.close();
            pendingAction = null;
        });

        // Confirm — fire HTMX POST
        confirmBtn.addEventListener('click', function() {
            if (!pendingAction) return;
            dialog.close();
            if (typeof htmx !== 'undefined') {
                htmx.ajax('POST', pendingAction.url, {
                    target: self._target,
                    swap: 'innerHTML',
                    headers: { 'X-CSRF-Token': self._getCSRF() }
                });
            }
            pendingAction = null;
        });
    }

    _getCSRF() {
        var meta = document.querySelector('meta[name="csrf-token"]');
        return meta ? meta.content : '';
    }
}

customElements.define('mcms-publish-button', McmsPublishButton);
