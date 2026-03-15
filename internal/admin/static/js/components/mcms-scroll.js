// <mcms-scroll> — Custom scrollbar component (Light DOM)
//
// Wraps content in a hidden-scrollbar viewport and renders a themed
// track + thumb overlay. Supports auto-hide (default) and always-visible
// via the `visible` attribute. Handles drag, track-click, and resize.
class McmsScroll extends HTMLElement {
    constructor() {
        super();
        this._viewport = null;
        this._track = null;
        this._thumb = null;
        this._resizeObserver = null;
        this._hideTimer = null;
        this._dragging = false;
        this._dragStartY = 0;
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

    _showTrack() {
        if (!this._track) return;
        this._track.classList.add('!opacity-100');
        this._resetHideTimer();
    }

    _resetHideTimer() {
        if (this._hideTimer) clearTimeout(this._hideTimer);
        if (this.hasAttribute('visible') || this._hovering || this._dragging) return;
        this._hideTimer = setTimeout(function() {
            if (this._track && !this._dragging && !this._hovering) {
                this._track.classList.remove('!opacity-100');
            }
        }.bind(this), 1000);
    }

    _onScroll() {
        console.log('[mcms-scroll] scroll', { scrollTop: this._viewport.scrollTop, scrollHeight: this._viewport.scrollHeight, clientHeight: this._viewport.clientHeight });
        this._updateThumb();
        this._showTrack();
    }

    _onPointerEnter() {
        console.log('[mcms-scroll] pointerenter');
        this._hovering = true;
        this._showTrack();
    }

    _onPointerLeave() {
        console.log('[mcms-scroll] pointerleave');
        this._hovering = false;
        this._resetHideTimer();
    }

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

    _onPointerUp() {
        console.log('[mcms-scroll] pointerup');
        this._dragging = false;
        this._thumb.classList.remove('!opacity-100');
        document.removeEventListener('pointermove', this._onPointerMove);
        document.removeEventListener('pointerup', this._onPointerUp);
        this._resetHideTimer();
    }

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

    _onResize() {
        this._updateThumb();
    }

    // --- Public API ---

    scrollTo(top) {
        if (!this._viewport) return;
        this._viewport.scrollTop = top;
    }

    scrollBy(delta) {
        if (!this._viewport) return;
        this._viewport.scrollTop += delta;
    }

    scrollToTop() {
        this.scrollTo(0);
    }

    scrollToBottom() {
        if (!this._viewport) return;
        this._viewport.scrollTop = this._viewport.scrollHeight;
    }

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

    get scrollTop() {
        return this._viewport ? this._viewport.scrollTop : 0;
    }

    get scrollHeight() {
        return this._viewport ? this._viewport.scrollHeight : 0;
    }

    get clientHeight() {
        return this._viewport ? this._viewport.clientHeight : 0;
    }
}

customElements.define('mcms-scroll', McmsScroll);
