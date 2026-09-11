import { renderLibrary } from './views/library.js';
import { renderSystemView } from './views/system.js';
import { renderStore } from './views/store.js';
import { renderGameDetail } from './views/game.js';
import { renderDownloads } from './views/downloads.js';
import { renderSettings } from './views/settings.js';
import { renderConfig } from './views/config.js';
import { refreshGalleryBanners } from './views/galleryHero.js';
import { renderProfile } from './views/profile.js';
import { renderSetup } from './views/setup.js';
import { renderLogin } from './views/login.js';
import { renderIntro } from './intro/intro.js';
import { renderVerify } from './views/verify.js';
import { renderGolf } from './views/golf.js';
import { renderSnake } from './views/snake.js';
import { appendLog } from './main.js';
import { store } from './state.js';
import { renderSectionSkeleton } from './utils/skeletons.js';

const routes = {
    '/intro':     renderIntro,
    '/library':   renderLibrary,
    '/store':     renderStore,
    '/game':      async (id) => await renderGameDetail(id),
    '/downloads': renderDownloads,
    '/settings':  renderSettings,
    '/config':    renderConfig,
    '/profile':   renderProfile,
    '/login':     renderLogin,
    '/arcade/golf': renderGolf,
    '/arcade/snake': renderSnake,
    '/verify':    renderVerify,
    '/setup':     renderSetup,
    '/setup-preview': renderSetup,
    '/system':    () => renderSystemView(),
};

const CACHED_ROUTES = ['/library', '/store', '/downloads', '/settings', '/config', '/profile', '/system'];
const viewCache = new Map();

/**
 * Invalidates cached route DOM elements so they re-render on next visit.
 * If route is omitted, clears all cached views.
 */
export function invalidateViewCache(path) {
    if (!path) {
        viewCache.forEach(el => el.remove());
        viewCache.clear();
        return;
    }
    const el = viewCache.get(path);
    if (el) {
        el.remove();
        viewCache.delete(path);
    }
}

function renderPlaceholder(title, description) {
    const el = document.createElement('div');
    el.innerHTML = `
        <div class="section-header">
            <h1>${title}</h1>
        </div>
        <div class="empty-state">
            <div class="empty-icon">◇</div>
            <h2>${title}</h2>
            <p>${description}</p>
        </div>
    `;
    return el;
}

export function initRouter() {
    const container = document.getElementById('app-content');
    if (!container) return;

    async function handleRoute() {
        let hash = window.location.hash.slice(1) || '/library';
        
        const [path, queryString] = hash.split('?');
        const params = new URLSearchParams(queryString || '');
        const id = params.get('id');

        // Instantly reset scroll position to top on every route transition
        if (container) {
            container.scrollTo({ top: 0, left: 0, behavior: 'instant' });
            container.scrollTop = 0;
        }
        window.scrollTo({ top: 0, left: 0, behavior: 'instant' });

        appendLog(`[Router] Navigate → ${path}${id ? `?id=${id}` : ''}`, 'log-system');

        // OOBE Check on initial load
        const seenIntro = window.localStorage.getItem('hasSeenIntro');
        let hasLoggedIn = window.localStorage.getItem('hasLoggedIn');

        // Force logout if backend lost the session file
        try {
            const backendLoggedIn = await window.go.rift.App.IsLoggedIn();
            if (!backendLoggedIn && hasLoggedIn) {
                window.localStorage.removeItem('hasLoggedIn');
                hasLoggedIn = null;
            }
        } catch (e) {}

        if (!seenIntro && path !== '/intro') {
            hash = '/intro';
            window.location.hash = hash;
            window.hasPassedSetup = false;
        } else if (!hasLoggedIn && path !== '/login' && path !== '/intro') {
            hash = '/login';
            window.location.hash = hash;
            window.hasPassedSetup = false;
        } else if (window.hasPassedSetup === undefined && hasLoggedIn) {
            try {
                const isReady = await window.go.rift.App.CheckSystemDependencies();
                const engineRaw = await window.go.rift.App.CheckEnginesStatus();
                const engines = JSON.parse(engineRaw);
                const savedPassedSetup = window.localStorage.getItem('hasPassedSetup') === 'true';

                // Route to setup if system deps or wine are missing, unless already passed setup
                if ((!isReady || !engines.wine) && !savedPassedSetup) {
                    window.hasPassedSetup = false;
                    hash = '/setup';
                    window.location.hash = '/setup';
                } else {
                    window.hasPassedSetup = true;
                }
            } catch (e) {
                console.error("Dependency check failed", e);
                window.hasPassedSetup = true;
            }
        } else if (window.hasPassedSetup === false && !['/setup', '/setup-preview', '/intro', '/login'].includes(hash)) {
            // Force user to stay on setup/intro/login if they try to navigate away
            hash = !seenIntro ? '/intro' : (!hasLoggedIn ? '/login' : '/setup');
            window.location.hash = hash;
        }

        // Offline Mode protection for Store
        if (path === '/store' && store.state.offlineMode) {
            if (window.showRiftNotification) {
                window.showRiftNotification('Store Unavailable', 'You are currently in Offline Mode. Please connect to the internet to browse the store.', 'warn');
            }
            window.location.hash = '/library';
            return;
        }

        // Cleanup snake arcade view if needed
        if (window.snakeCleanup && path !== '/arcade/snake') {
            try { window.snakeCleanup(); } catch(e) {}
        }

        // Handle Setup Fullscreen Override
        if (['/setup', '/setup-preview', '/intro', '/login', '/arcade/golf', '/arcade/snake'].includes(path)) {
            document.body.classList.add('setup-active');
        } else {
            document.body.classList.remove('setup-active');
        }

        // Update sidebar
        document.querySelectorAll('.nav-item').forEach(el => {
            el.classList.toggle('active', el.dataset.route === path);
        });

        // Hide notification dot if on downloads page
        if (path === '/downloads') {
            const dlDot = document.getElementById('nav-dl-dot');
            if (dlDot) dlDot.style.display = 'none';
        }

        const renderer = routes[path] || routes['/library'];

        // Persistent Cache Handling for Instant (<2ms) Navigation
        const isCacheable = CACHED_ROUTES.includes(path) && !queryString;

        if (isCacheable) {
            const dyn = document.getElementById('view-dynamic');
            if (dyn) dyn.style.display = 'none';

            if (viewCache.has(path)) {
                // Instantly toggle display
                viewCache.forEach((el, route) => {
                    el.style.display = (route === path) ? 'block' : 'none';
                });

                const targetView = viewCache.get(path);

                // For Library: refresh the 18 banner images in-place without touching 3D layout or styles
                if (path === '/library') {
                    refreshGalleryBanners(targetView);
                }
                return;
            }

            // Not yet cached: render once and store in viewCache
            viewCache.forEach(el => { el.style.display = 'none'; });

            const safeId = 'view-' + path.replace(/[^a-zA-Z0-9]/g, '-');
            const wrapper = document.createElement('div');
            wrapper.className = 'route-view view-enter';
            wrapper.id = safeId;
            wrapper.style.width = '100%';
            wrapper.style.minHeight = '100%';

            renderSectionSkeleton(wrapper, path);
            container.appendChild(wrapper);

            const content = await renderer(id);
            wrapper.innerHTML = '';
            if (typeof content === 'string') {
                wrapper.innerHTML = content;
            } else if (content instanceof HTMLElement) {
                wrapper.appendChild(content);
            }

            viewCache.set(path, wrapper);
            return;
        }

        // Dynamic / Uncached Route (Game Details, Setup, Login, Arcade)
        viewCache.forEach(el => { el.style.display = 'none'; });

        let dynWrapper = document.getElementById('view-dynamic');
        if (!dynWrapper) {
            dynWrapper = document.createElement('div');
            dynWrapper.id = 'view-dynamic';
            dynWrapper.style.width = '100%';
            dynWrapper.style.minHeight = '100%';
            container.appendChild(dynWrapper);
        }
        dynWrapper.style.display = 'block';

        const prevView = dynWrapper.firstElementChild;
        if (prevView && typeof prevView._cleanup === 'function') {
            try { prevView._cleanup(); } catch (err) { console.error("View cleanup failed:", err); }
        }

        if (!['/setup', '/setup-preview', '/intro', '/login', '/arcade/golf', '/arcade/snake'].includes(path)) {
            renderSectionSkeleton(dynWrapper, path);
        }

        const content = await renderer(id);
        dynWrapper.innerHTML = '';
        if (typeof content === 'string') {
            dynWrapper.innerHTML = content;
        } else if (content instanceof HTMLElement) {
            dynWrapper.appendChild(content);
        }

        // Ensure container is scrolled to top on dynamic view load
        if (container) {
            container.scrollTo({ top: 0, left: 0, behavior: 'instant' });
            container.scrollTop = 0;
            requestAnimationFrame(() => {
                container.scrollTop = 0;
                window.scrollTo({ top: 0, left: 0, behavior: 'instant' });
            });
        }

        dynWrapper.classList.remove('view-enter');
        void dynWrapper.offsetWidth; // Force reflow
        dynWrapper.classList.add('view-enter');
    }

    window.addEventListener('hashchange', handleRoute);

    // Sidebar clicks
    document.querySelectorAll('.nav-item[data-route]').forEach(el => {
        el.addEventListener('click', () => {
            window.location.hash = el.dataset.route;
        });
    });

    // Initial
    handleRoute();
}
