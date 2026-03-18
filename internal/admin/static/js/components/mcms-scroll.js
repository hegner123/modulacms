/**
 * @module mcms-scroll
 * @description Custom scrollbar component that wraps content in a hidden-scrollbar
 * viewport and renders a themed track and thumb overlay.
 *
 * Supports auto-hide behavior (default) where the scrollbar fades in on scroll or
 * hover and fades out after 1 second of inactivity. The `visible` attribute keeps the
 * scrollbar always visible. The thumb is draggable, and clicking the track jumps to
 * the corresponding scroll position.
 *
 * The component uses a `ResizeObserver` to recalculate the thumb size and position
 * when the viewport dimensions change. The scrollbar track is hidden entirely when
 * content fits within the viewport.
 *
 * This is a Light DOM web component -- child content is moved into an internal
 * viewport div, and the scrollbar track/thumb are appended as siblings.
 *
 * @example
 * <!-- Auto-hiding scrollbar (default) -->
 * <mcms-scroll style="height: 400px;">
 *   <div>Long scrollable content here...</div>
 * </mcms-scroll>
 *
 * @example
 * <!-- Always-visible scrollbar -->
 * <mcms-scroll visible style="height: 300px;">
 *   <ul>
 *     <li>Item 1</li>
 *     <li>Item 2</li>
 *     <!-- ... many items ... -->
 *   </ul>
 * </mcms-scroll>
 *
 * @example
 * <!-- Programmatic scroll control -->
 * <mcms-scroll id="my-scroll" style="height: 500px;">
 *   <div id="target-element">Target</div>
 *   <!-- content -->
 * </mcms-scroll>
 * <script>
 *   const scroller = document.getElementById('my-scroll');
 *   scroller.scrollToTop();
 *   scroller.scrollToBottom();
 *   scroller.scrollTo(200);
 *   scroller.scrollBy(50);
 *   scroller.scrollIntoView(document.getElementById('target-element'));
 * </script>
 */

/**
 * Custom scrollbar web component with themed track/thumb, drag support, and auto-hide.
 *
 * Moves all child content into an internal viewport div that has native scrolling with
 * `scrollbar-none` (hidden browser scrollbar). A custom track and thumb overlay are
 * rendered alongside the viewport, sized and positioned proportionally to the content.
 *
 * @extends HTMLElement
 *
 * @attr {boolean} visible - When present, the scrollbar track is always visible instead
 *   of fading out after 1 second of inactivity. By default, the scrollbar auto-hides.
 */
class McmsScroll extends HTMLElement {
    constructor() {
        super();

        /**
         * The viewport div that holds the scrollable content.
         * @type {HTMLDivElement|null}
         * @private
         */
        this._viewport = null;

        /**
         * The scrollbar track element (container for the thumb).
         * @type {HTMLDivElement|null}
         * @private
         */
        this._track = null;

        /**
         * The scrollbar thumb element (draggable indicator).
         * @type {HTMLDivElement|null}
         * @private
         */
        this._thumb = null;

        /**
         * ResizeObserver watching the viewport for dimension changes.
         * @type {ResizeObserver|null}
         * @private
         */
        this._resizeObserver = null;

        /**
         * Timer ID for the auto-hide delay. Cleared when the user hovers or drags.
         * @type {number|null}
         * @private
         */
        this._hideTimer = null;

        /**
         * Whether the user is currently dragging the scrollbar thumb.
         * @type {boolean}
         * @private
         */
        this._dragging = false;

        /**
         * The clientY coordinate at the start of a thumb drag operation.
         * @type {number}
         * @private
         */
        this._dragStartY = 0;

        /**
         * The viewport's scrollTop value at the start of a thumb drag operation.
         * @type {number}
         * @private
         */
        this._dragStartScrollTop = 0;

        // Bound handlers for cleanup
        this._onScroll = this._onScroll.bind(this);
        this._onPointerDown = this._onPointerDown.bind(this);
        this._onPointerMove = this._onPointerMove.bind(this);
        this._onPointerUp = this._onPointerUp.bind(this);
        this._onTrackClick = this._onTrackClick.bind(this);
        this._onResize = this._onResize.bind(this);
        this._onPointerEnter = this._onPointerEnter.bind(this);
        this._onPointerLeave = this._onPointerLeave.bind(this);
    }

    /**
     * Lifecycle callback invoked when the element is inserted into the DOM.
     *
     * Captures existing child nodes, creates the viewport and moves children into it,
     * creates the track and thumb elements, appends everything to the component,
     * binds all event handlers (scroll, pointer, resize), and performs an initial
     * thumb size/position calculation.
     */
    connectedCallback() {
        // Capture existing children before restructuring
        var children = Array.from(this.childNodes);

        // Create viewport
        this._viewport = document.createElement('div');
        this._viewport.className = 'overflow-y-scroll h-full scrollbar-none';

        // Move children into viewport
        for (var i = 0; i < children.length; i++) {
            this._viewport.appendChild(children[i]);
        }

        // Create track and thumb
        this._track = document.createElement('div');
        this._track.className = 'absolute top-0 right-0 bottom-0 w-2 opacity-0 transition-opacity duration-200';

        this._thumb = document.createElement('div');
        this._thumb.className = 'absolute w-full rounded-full bg-[var(--color-text-dim)] opacity-50 hover:opacity-75 transition-opacity';
        this._track.appendChild(this._thumb);

        // Append to component
        this.appendChild(this._viewport);
        this.appendChild(this._track);

        // Bind events
        this._viewport.addEventListener('scroll', this._onScroll, { passive: true });
        this._thumb.addEventListener('pointerdown', this._onPointerDown);
        this._track.addEventListener('click', this._onTrackClick);
        this.addEventListener('pointerenter', this._onPointerEnter);
        this.addEventListener('pointerleave', this._onPointerLeave);

        // Watch for size changes
        this._resizeObserver = new ResizeObserver(this._onResize);
        this._resizeObserver.observe(this._viewport);

        // Initial calculation
        this._updateThumb();
        console.log('[mcms-scroll] connected', {
            viewportHeight: this._viewport.clientHeight,
            scrollHeight: this._viewport.scrollHeight,
            trackHeight: this._track.clientHeight,
            children: this._viewport.childElementCount,
            componentHeight: this.clientHeight,
            componentStyle: window.getComputedStyle(this).height,
            viewportStyle: window.getComputedStyle(this._viewport).height,
            overflow: window.getComputedStyle(this._viewport).overflowY
        });
    }

    /**
     * Lifecycle callback invoked when the element is removed from the DOM.
     *
     * Cleans up the ResizeObserver, clears the auto-hide timer, removes all event
     * listeners from the viewport, thumb, track, and document to prevent memory leaks
     * and orphaned handlers.
     */
    disconnectedCallback() {
        if (this._resizeObserver) {
            this._resizeObserver.disconnect();
            this._resizeObserver = null;
        }
        if (this._hideTimer) {
            clearTimeout(this._hideTimer);
            this._hideTimer = null;
        }
        if (this._viewport) {
            this._viewport.removeEventListener('scroll', this._onScroll);
        }
        if (this._thumb) {
            this._thumb.removeEventListener('pointerdown', this._onPointerDown);
        }
        if (this._track) {
            this._track.removeEventListener('click', this._onTrackClick);
        }
        this.removeEventListener('pointerenter', this._onPointerEnter);
        this.removeEventListener('pointerleave', this._onPointerLeave);
        // Release any active pointer capture
        document.removeEventListener('pointermove', this._onPointerMove);
        document.removeEventListener('pointerup', this._onPointerUp);
    }

    /**
     * Recalculates the thumb's height and vertical position based on the viewport's
     * scroll dimensions. Hides the track entirely when content fits without scrolling.
     *
     * The thumb height is proportional to the ratio of viewport height to scroll height,
     * with a minimum of 30px. The thumb position is proportional to the current scroll
     * offset within the scrollable range.
     *
     * @private
     */
    _updateThumb() {
        if (!this._viewport || !this._track || !this._thumb) return;

        var viewportHeight = this._viewport.clientHeight;
        var scrollHeight = this._viewport.scrollHeight;

        // Hide track if content fits
        if (scrollHeight <= viewportHeight) {
            this._track.classList.add('hidden');
            return;
        }
        this._track.classList.remove('hidden');

        var trackHeight = this._track.clientHeight;
        var ratio = viewportHeight / scrollHeight;
        var thumbHeight = Math.max(30, Math.round(trackHeight * ratio));
        var scrollTop = this._viewport.scrollTop;
        var maxScroll = scrollHeight - viewportHeight;
        var maxThumbTop = trackHeight - thumbHeight;
        var thumbTop = maxScroll > 0 ? Math.round((scrollTop / maxScroll) * maxThumbTop) : 0;

        this._thumb.style.height = thumbHeight + 'px';
        this._thumb.style.top = thumbTop + 'px';
    }

    /**
     * Shows the scrollbar track by adding a full-opacity override class, then
     * schedules the auto-hide timer via `_resetHideTimer()`.
     *
     * @private
     */
    _showTrack() {
        if (!this._track) return;
        this._track.classList.add('!opacity-100');
        this._resetHideTimer();
    }

    /**
     * Resets the auto-hide timer. After 1 second of inactivity (no scrolling, hovering,
     * or dragging), removes the opacity override to fade the track out. The timer is
     * suppressed when the `visible` attribute is present, the pointer is hovering over
     * the component, or the thumb is being dragged.
     *
     * @private
     */
    _resetHideTimer() {
        if (this._hideTimer) clearTimeout(this._hideTimer);
        if (this.hasAttribute('visible') || this._hovering || this._dragging) return;
        this._hideTimer = setTimeout(function() {
            if (this._track && !this._dragging && !this._hovering) {
                this._track.classList.remove('!opacity-100');
            }
        }.bind(this), 1000);
    }

    /**
     * Handles viewport scroll events. Updates the thumb position and shows the track.
     *
     * @private
     */
    _onScroll() {
        console.log('[mcms-scroll] scroll', { scrollTop: this._viewport.scrollTop, scrollHeight: this._viewport.scrollHeight, clientHeight: this._viewport.clientHeight });
        this._updateThumb();
        this._showTrack();
    }

    /**
     * Handles pointer entering the component. Sets the hovering flag and shows the track.
     *
     * @private
     */
    _onPointerEnter() {
        console.log('[mcms-scroll] pointerenter');
        this._hovering = true;
        this._showTrack();
    }

    /**
     * Handles pointer leaving the component. Clears the hovering flag and starts the
     * auto-hide timer.
     *
     * @private
     */
    _onPointerLeave() {
        console.log('[mcms-scroll] pointerleave');
        this._hovering = false;
        this._resetHideTimer();
    }

    /**
     * Handles pointer down on the scrollbar thumb to begin a drag operation.
     *
     * Records the starting Y coordinate and scroll position, sets the dragging flag,
     * applies full opacity to both thumb and track, and attaches document-level
     * pointermove/pointerup listeners for tracking the drag across the entire page.
     * Only responds to the primary (left) mouse button.
     *
     * @param {PointerEvent} e - The pointer down event.
     * @private
     */
    _onPointerDown(e) {
        console.log('[mcms-scroll] pointerdown on thumb', { button: e.button, clientY: e.clientY });
        if (e.button !== 0) return;
        e.preventDefault();
        e.stopPropagation();

        this._dragging = true;
        this._dragStartY = e.clientY;
        this._dragStartScrollTop = this._viewport.scrollTop;
        this._thumb.classList.add('!opacity-100');
        this._track.classList.add('!opacity-100');

        document.addEventListener('pointermove', this._onPointerMove);
        document.addEventListener('pointerup', this._onPointerUp);
    }

    /**
     * Handles pointer move during a thumb drag operation.
     *
     * Converts the pixel delta from the drag start position into a proportional scroll
     * offset based on the ratio of thumb travel distance to maximum scroll distance.
     * Updates the viewport's `scrollTop` accordingly.
     *
     * @param {PointerEvent} e - The pointer move event.
     * @private
     */
    _onPointerMove(e) {
        if (!this._dragging) return;
        console.log('[mcms-scroll] pointermove (dragging)', { clientY: e.clientY, deltaY: e.clientY - this._dragStartY });

        var deltaY = e.clientY - this._dragStartY;
        var trackHeight = this._track.clientHeight;
        var viewportHeight = this._viewport.clientHeight;
        var scrollHeight = this._viewport.scrollHeight;
        var thumbHeight = Math.max(30, Math.round(trackHeight * (viewportHeight / scrollHeight)));
        var maxThumbTop = trackHeight - thumbHeight;

        console.log('[mcms-scroll] drag calc', { trackHeight: trackHeight, thumbHeight: thumbHeight, maxThumbTop: maxThumbTop, scrollHeight: scrollHeight, viewportHeight: viewportHeight });

        if (maxThumbTop <= 0) return;

        var scrollRatio = deltaY / maxThumbTop;
        var maxScroll = scrollHeight - viewportHeight;
        this._viewport.scrollTop = this._dragStartScrollTop + (scrollRatio * maxScroll);
    }

    /**
     * Handles pointer up to end a thumb drag operation.
     *
     * Clears the dragging flag, removes the forced opacity from the thumb, detaches
     * document-level pointermove/pointerup listeners, and starts the auto-hide timer.
     *
     * @private
     */
    _onPointerUp() {
        console.log('[mcms-scroll] pointerup');
        this._dragging = false;
        this._thumb.classList.remove('!opacity-100');
        document.removeEventListener('pointermove', this._onPointerMove);
        document.removeEventListener('pointerup', this._onPointerUp);
        this._resetHideTimer();
    }

    /**
     * Handles clicks on the scrollbar track (not on the thumb itself).
     *
     * Calculates the click position as a ratio of the track height and scrolls the
     * viewport to the proportional position in the content. Ignores clicks that land
     * directly on the thumb element.
     *
     * @param {MouseEvent} e - The click event on the track.
     * @private
     */
    _onTrackClick(e) {
        // Ignore clicks on the thumb itself
        if (e.target === this._thumb) {
            console.log('[mcms-scroll] track click ignored (on thumb)');
            return;
        }
        e.stopPropagation();
        console.log('[mcms-scroll] track click', { clientY: e.clientY, target: e.target.className });

        var trackRect = this._track.getBoundingClientRect();
        var clickY = e.clientY - trackRect.top;
        var trackHeight = this._track.clientHeight;
        var ratio = clickY / trackHeight;
        var scrollHeight = this._viewport.scrollHeight;
        var viewportHeight = this._viewport.clientHeight;

        console.log('[mcms-scroll] track click calc', { clickY: clickY, trackHeight: trackHeight, ratio: ratio, scrollTo: ratio * (scrollHeight - viewportHeight) });
        this._viewport.scrollTop = ratio * (scrollHeight - viewportHeight);
    }

    /**
     * Handles ResizeObserver notifications. Recalculates thumb size and position
     * when the viewport dimensions change.
     *
     * @private
     */
    _onResize() {
        this._updateThumb();
    }

    // --- Public API ---

    /**
     * Scrolls the viewport to an absolute vertical position.
     *
     * @param {number} top - The target `scrollTop` value in pixels.
     * @public
     *
     * @example
     * document.querySelector('mcms-scroll').scrollTo(150);
     */
    scrollTo(top) {
        if (!this._viewport) return;
        this._viewport.scrollTop = top;
    }

    /**
     * Scrolls the viewport by a relative amount.
     *
     * @param {number} delta - The number of pixels to scroll. Positive values scroll
     *   down, negative values scroll up.
     * @public
     *
     * @example
     * document.querySelector('mcms-scroll').scrollBy(100);  // scroll down 100px
     * document.querySelector('mcms-scroll').scrollBy(-50);  // scroll up 50px
     */
    scrollBy(delta) {
        if (!this._viewport) return;
        this._viewport.scrollTop += delta;
    }

    /**
     * Scrolls the viewport to the very top (scrollTop = 0).
     *
     * @public
     *
     * @example
     * document.querySelector('mcms-scroll').scrollToTop();
     */
    scrollToTop() {
        this.scrollTo(0);
    }

    /**
     * Scrolls the viewport to the very bottom, revealing the last content.
     *
     * @public
     *
     * @example
     * document.querySelector('mcms-scroll').scrollToBottom();
     */
    scrollToBottom() {
        if (!this._viewport) return;
        this._viewport.scrollTop = this._viewport.scrollHeight;
    }

    /**
     * Scrolls the viewport so that a child element is visible within the viewport bounds.
     *
     * If the child is above the viewport, scrolls up to align its top edge with the
     * viewport top. If below, scrolls down to align its bottom edge with the viewport
     * bottom. If already visible, does nothing.
     *
     * @param {HTMLElement} child - The child element to scroll into view. Must be a
     *   descendant of the viewport.
     * @public
     *
     * @example
     * const item = document.getElementById('list-item-42');
     * document.querySelector('mcms-scroll').scrollIntoView(item);
     */
    scrollIntoView(child) {
        if (!this._viewport || !child) return;
        var vRect = this._viewport.getBoundingClientRect();
        var cRect = child.getBoundingClientRect();
        if (cRect.top < vRect.top) {
            this._viewport.scrollTop -= vRect.top - cRect.top;
        } else if (cRect.bottom > vRect.bottom) {
            this._viewport.scrollTop += cRect.bottom - vRect.bottom;
        }
    }

    /**
     * Returns the current vertical scroll offset of the viewport.
     *
     * @type {number}
     * @readonly
     * @public
     */
    get scrollTop() {
        return this._viewport ? this._viewport.scrollTop : 0;
    }

    /**
     * Returns the total scrollable height of the viewport's content.
     *
     * @type {number}
     * @readonly
     * @public
     */
    get scrollHeight() {
        return this._viewport ? this._viewport.scrollHeight : 0;
    }

    /**
     * Returns the visible height of the viewport (the non-scrolling portion).
     *
     * @type {number}
     * @readonly
     * @public
     */
    get clientHeight() {
        return this._viewport ? this._viewport.clientHeight : 0;
    }
}

customElements.define('mcms-scroll', McmsScroll);
