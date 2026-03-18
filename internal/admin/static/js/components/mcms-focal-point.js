/**
 * @module mcms-focal-point
 * @description Focal point picker for images that allows users to click on an image
 * to set a focal point coordinate, displayed as a crosshair marker overlay.
 *
 * The focal point is stored as normalized coordinates (0-1 range on both axes) in
 * hidden form inputs (`focal_x` and `focal_y`) within a target form. This enables
 * responsive image cropping systems to preserve the most important part of an image
 * regardless of the output dimensions.
 *
 * The marker automatically repositions when the image container resizes, using a
 * `ResizeObserver`. A "Clear" button removes the focal point and empties the hidden
 * inputs.
 *
 * This is a Light DOM web component -- all rendered elements are direct children of the
 * host element, making them accessible to global CSS and form handling.
 *
 * @example
 * <!-- Basic focal point picker -->
 * <form id="media-form">
 *   <mcms-focal-point
 *       src="/uploads/photo.jpg"
 *       alt="Product photo"
 *       focal-x="0.35"
 *       focal-y="0.6"
 *   ></mcms-focal-point>
 * </form>
 *
 * @example
 * <!-- With custom form target -->
 * <form id="edit-image">
 *   <mcms-focal-point
 *       src="/uploads/banner.jpg"
 *       alt="Banner image"
 *       form-id="edit-image"
 *   ></mcms-focal-point>
 * </form>
 */

/**
 * Focal point picker web component for images.
 *
 * Renders an image with a crosshair marker overlay at the selected focal point.
 * Click anywhere on the image to set the focal point; the component updates hidden
 * form inputs with the normalized (0-1) coordinates. A "Clear" button removes the
 * focal point entirely. The marker tracks image resizes via `ResizeObserver`.
 *
 * @extends HTMLElement
 *
 * @attr {string} src - URL of the image to display.
 * @attr {string} alt - Alt text for the image element.
 * @attr {string} focal-x - Initial X coordinate of the focal point, normalized to
 *   the 0-1 range where 0 is the left edge and 1 is the right edge. Leave empty
 *   or omit to indicate no focal point is set.
 * @attr {string} focal-y - Initial Y coordinate of the focal point, normalized to
 *   the 0-1 range where 0 is the top edge and 1 is the bottom edge. Leave empty
 *   or omit to indicate no focal point is set.
 * @attr {string} form-id - The `id` of the form element where hidden inputs
 *   (`focal_x` and `focal_y`) are created or updated. Default: `"media-form"`.
 */
class McmsFocalPoint extends HTMLElement {
    /**
     * Lifecycle callback invoked when the element is inserted into the DOM.
     *
     * Parses initial focal point coordinates from attributes, renders the image
     * container with marker and clear button, ensures hidden inputs exist in the
     * target form, and binds click, clear, and resize event handlers. If initial
     * coordinates are provided, positions the marker once the image loads.
     */
    connectedCallback() {
        this._focalX = this._parseAttr('focal-x');
        this._focalY = this._parseAttr('focal-y');
        this._formId = this.getAttribute('form-id') || 'media-form';
        this._render();
        this._bindEvents();
    }

    /**
     * Lifecycle callback invoked when the element is removed from the DOM.
     *
     * Disconnects the `ResizeObserver` that tracks image size changes to prevent
     * memory leaks.
     */
    disconnectedCallback() {
        if (this._resizeObserver) {
            this._resizeObserver.disconnect();
        }
    }

    /**
     * Parses a numeric attribute value, returning `null` for empty, missing, or
     * non-numeric values.
     *
     * @param {string} name - The attribute name to read.
     * @returns {number|null} The parsed float value, or `null` if the attribute is
     *   absent, empty, or not a valid number.
     * @private
     */
    _parseAttr(name) {
        const val = this.getAttribute(name);
        if (val === null || val === '') return null;
        const n = parseFloat(val);
        return isNaN(n) ? null : n;
    }

    /**
     * Builds the component's inner HTML: an image element with crosshair cursor,
     * an absolutely positioned marker (initially hidden), and a "Clear" button
     * (initially hidden). Stores references to key elements and creates hidden
     * form inputs in the target form.
     *
     * If initial focal point coordinates exist, waits for the image to load before
     * positioning the marker.
     *
     * @private
     */
    _render() {
        const src = this.getAttribute('src') || '';
        const alt = this.getAttribute('alt') || '';

        this.innerHTML = `
            <div class="relative" data-focal-container>
                <img
                    src="${src}"
                    alt="${alt}"
                    class="w-full rounded-lg object-contain cursor-crosshair"
                    data-focal-img
                    draggable="false"
                />
                <div class="absolute pointer-events-none hidden" data-focal-marker style="transform:translate(-50%,-50%)">
                    <div class="size-6 rounded-full border-2 border-white bg-blue-500/50 flex items-center justify-center">
                        <div class="size-2 rounded-full bg-white"></div>
                    </div>
                </div>
                <button
                    type="button"
                    class="absolute top-2 right-2 hidden rounded-md bg-gray-900/70 px-2 py-1 text-xs font-medium text-white hover:bg-gray-900/90"
                    data-focal-clear
                >Clear</button>
            </div>
        `;

        this._container = this.querySelector('[data-focal-container]');
        this._img = this.querySelector('[data-focal-img]');
        this._marker = this.querySelector('[data-focal-marker]');
        this._clearBtn = this.querySelector('[data-focal-clear]');

        // Ensure hidden inputs exist in the target form
        this._ensureInputs();

        // Position marker if focal point is set
        if (this._focalX !== null && this._focalY !== null) {
            this._waitForImageThen(() => this._updateMarker());
        }
    }

    /**
     * Creates hidden `focal_x` and `focal_y` input elements in the target form
     * if they do not already exist. Sets their initial values from the current
     * focal point coordinates.
     *
     * @private
     */
    _ensureInputs() {
        const form = document.getElementById(this._formId);
        if (!form) return;

        if (!form.querySelector('input[name="focal_x"]')) {
            const ix = document.createElement('input');
            ix.type = 'hidden';
            ix.name = 'focal_x';
            ix.value = this._focalX !== null ? String(this._focalX) : '';
            form.appendChild(ix);
        }
        if (!form.querySelector('input[name="focal_y"]')) {
            const iy = document.createElement('input');
            iy.type = 'hidden';
            iy.name = 'focal_y';
            iy.value = this._focalY !== null ? String(this._focalY) : '';
            form.appendChild(iy);
        }
    }

    /**
     * Binds all interactive event handlers: image click to set the focal point,
     * clear button to reset, and a `ResizeObserver` on the image to reposition
     * the marker when the image dimensions change.
     *
     * Image click coordinates are converted to normalized 0-1 values relative to
     * the image's bounding rect, clamped to ensure they stay within bounds.
     *
     * @private
     */
    _bindEvents() {
        // Click on image to set focal point
        this._img.addEventListener('click', (e) => {
            const rect = this._img.getBoundingClientRect();
            let x = (e.clientX - rect.left) / rect.width;
            let y = (e.clientY - rect.top) / rect.height;

            // Clamp to [0, 1]
            x = Math.max(0, Math.min(1, x));
            y = Math.max(0, Math.min(1, y));

            this._focalX = x;
            this._focalY = y;
            this._updateMarker();
            this._updateInputs();
        });

        // Clear button
        this._clearBtn.addEventListener('click', () => {
            this._focalX = null;
            this._focalY = null;
            this._marker.classList.add('hidden');
            this._clearBtn.classList.add('hidden');
            this._updateInputs();
        });

        // Handle resize
        this._resizeObserver = new ResizeObserver(() => {
            if (this._focalX !== null && this._focalY !== null) {
                this._updateMarker();
            }
        });
        this._resizeObserver.observe(this._img);
    }

    /**
     * Executes a callback once the image has loaded. If the image is already
     * complete (cached), calls the function immediately; otherwise, adds a
     * one-time `load` event listener.
     *
     * @param {Function} fn - The callback to execute after the image loads.
     * @private
     */
    _waitForImageThen(fn) {
        if (this._img.complete && this._img.naturalWidth > 0) {
            fn();
        } else {
            this._img.addEventListener('load', fn, { once: true });
        }
    }

    /**
     * Positions the marker element at the current focal point coordinates.
     *
     * Converts normalized (0-1) focal coordinates to pixel positions relative to
     * the image's rendered dimensions. Shows the marker and clear button. No-ops
     * if focal coordinates are null or the image has zero dimensions (not yet rendered).
     *
     * @private
     */
    _updateMarker() {
        if (this._focalX === null || this._focalY === null) return;

        const w = this._img.clientWidth;
        const h = this._img.clientHeight;
        if (w === 0 || h === 0) return;

        const px = this._focalX * w;
        const py = this._focalY * h;

        this._marker.style.left = px + 'px';
        this._marker.style.top = py + 'px';
        this._marker.classList.remove('hidden');
        this._clearBtn.classList.remove('hidden');
    }

    /**
     * Toggles the visibility of the focal point marker. If the marker is currently
     * hidden, it is shown at the current focal coordinates; if visible, it is hidden.
     * No-ops if no focal point is set.
     *
     * @public
     *
     * @example
     * const picker = document.querySelector('mcms-focal-point');
     * picker.toggleMarker(); // hides the marker if visible, shows if hidden
     */
    toggleMarker() {
        if (this._focalX === null || this._focalY === null) return;
        if (this._marker.classList.contains('hidden')) {
            this._updateMarker();
        } else {
            this._marker.classList.add('hidden');
        }
    }

    /**
     * Synchronizes the hidden form inputs (`focal_x` and `focal_y`) in the target
     * form with the current focal point coordinates. Sets values to empty strings
     * when the focal point is cleared.
     *
     * @private
     */
    _updateInputs() {
        const form = document.getElementById(this._formId);
        if (!form) return;

        const ix = form.querySelector('input[name="focal_x"]');
        const iy = form.querySelector('input[name="focal_y"]');

        if (ix) ix.value = this._focalX !== null ? String(this._focalX) : '';
        if (iy) iy.value = this._focalY !== null ? String(this._focalY) : '';
    }
}

customElements.define('mcms-focal-point', McmsFocalPoint);
