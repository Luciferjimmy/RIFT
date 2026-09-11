import { store } from '../state.js';
import { appendLog, escapeHTML } from '../main.js';
import { renderGalleryHero, setupGalleryInteractivity } from './galleryHero.js';

export function renderLibrary() {
    const container = document.createElement('div');
    container.className = 'view-premium';

    const content = document.createElement('div');
    content.id = 'library-content';
    container.appendChild(content);

    // Inline download progress bar (shown on top of library when a download is active)
    const progressBar = document.createElement('div');
    progressBar.id = 'library-dl-bar';
    progressBar.style.cssText = 'display: none; margin-bottom: 16px; border-radius: 10px; background: rgba(255,255,255,0.03); border: 1px solid rgba(255,255,255,0.08); overflow: hidden;';
    container.insertBefore(progressBar, content);

    function renderProgressBar(dl) {
        if (!dl || !dl.gameId) {
            progressBar.style.display = 'none';
            return;
        }
        const game = store.state.games.find(g => g.id === dl.gameId);
        const name = game ? game.name : dl.gameId;
        const safeName = escapeHTML(name);
        const pct = dl.percent || 0;
        const stage = dl.stage || 'Starting...';
        const safeStage = escapeHTML(stage);
        const speed = typeof dl.speed === 'number' ? `${dl.speed.toFixed(1)} MB/s` : '';

        progressBar.style.display = 'block';
        progressBar.innerHTML = `
            <div style="padding: 16px 20px; display: flex; align-items: center; gap: 16px;">
                <div style="flex: 1; min-width: 0;">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
                        <span style="font-family: 'Plus Jakarta Sans', sans-serif; font-size: 0.85rem; font-weight: 600; color: white;">${safeName}</span>
                        <span style="font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; color: rgba(255,255,255,0.6);">${pct.toFixed(0)}% ${speed ? '· ' + speed : ''}</span>
                    </div>
                    <div style="height: 4px; background: rgba(255,255,255,0.08); border-radius: 4px; overflow: hidden; margin-bottom: 4px;">
                        <div style="height: 100%; width: ${pct}%; background: linear-gradient(90deg, #38bdf8, #818cf8); border-radius: 4px; transition: width 0.4s ease;"></div>
                    </div>
                    <div style="font-family: 'JetBrains Mono', monospace; font-size: 0.7rem; color: rgba(255,255,255,0.4); overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">${safeStage}</div>
                </div>
            </div>
        `;
    }

    // Subscribe to activeDownload for the progress bar
    const unsubDL = store.subscribe('activeDownload', renderProgressBar);

    let searchQuery = '';
    let filterStatus = 'all';

    function getStorefrontPortalsHtml() {
        const isSteamConnected = !!store.state.authStatus?.steamConnected;
        const steamUser = store.state.authStatus?.steamUsername || '';
        return `
            <section class="lib-section" style="padding-top: 40px; border-bottom: 1px solid rgba(255,255,255,0.05); padding-bottom: 40px;">
                <div class="lib-section-head">
                    <h2 class="lib-section-title"><span class="text-serif-italic">Storefront</span> <span class="text-sans">Portals</span></h2>
                </div>
                
                <div style="display: flex; flex-direction: column; gap: 16px; margin-top: 14px;">
                    <!-- STEAM STOREFRONT BANNER -->
                    <div class="runtime-banner tilt-card" style="display: flex; align-items: center; justify-content: space-between; background: linear-gradient(135deg, #171a21 0%, #1b2838 50%, #2a475e 100%); border-radius: 12px; padding: 22px 32px; border: 1px solid rgba(255,255,255,0.1); box-shadow: 0 10px 30px rgba(0,0,0,0.3);">
                        <div style="display: flex; align-items: center; gap: 16px;">
                            <img src="https://store.cloudflare.steamstatic.com/public/shared/images/header/logo_steam.svg" alt="Steam" style="height: 42px; filter: drop-shadow(0 4px 12px rgba(0,0,0,0.5));">
                            ${isSteamConnected ? `
                                <div style="display: flex; flex-direction: column; gap: 3px;">
                                    <span style="font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; color: #66c0f4; letter-spacing: 0.5px;">✓ CONNECTED: ${escapeHTML(steamUser)}</span>
                                    <span style="font-family: 'Plus Jakarta Sans', sans-serif; font-size: 0.8rem; color: rgba(255,255,255,0.7);">Search titles in the Store to install games into your Archive.</span>
                                </div>
                            ` : ''}
                        </div>
                        <button class="exhibit-enter-btn steam-banner-btn" style="margin: 0; transition: all 0.3s ease; background: rgba(52, 152, 219, 0.15); color: #66c0f4; border-color: rgba(102, 192, 244, 0.4); cursor: pointer; display: inline-flex; align-items: center; gap: 8px;">
                            <span>OPEN STEAM WEB STORE</span>
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6M15 3h6v6M10 14L21 3"/></svg>
                        </button>
                    </div>

                    <!-- EPIC GAMES STOREFRONT BANNER -->
                    <div class="runtime-banner tilt-card" style="display: flex; align-items: center; justify-content: space-between; background: linear-gradient(135deg, #141416 0%, #1c1c20 50%, #292930 100%); border-radius: 12px; padding: 22px 32px; border: 1px solid rgba(255,255,255,0.1); box-shadow: 0 10px 30px rgba(0,0,0,0.3);">
                        <div style="display: flex; align-items: center; gap: 14px;">
                            <svg width="34" height="34" viewBox="0 0 24 24" fill="#ffffff" style="filter: drop-shadow(0 4px 12px rgba(0,0,0,0.5));">
                                <path d="M3.537 0C2.165 0 1.66.506 1.66 1.879V18.44a4.262 4.262 0 00.02.433c.031.3.037.59.316.92.027.033.311.245.311.245.153.075.258.13.43.2l8.335 3.491c.433.199.614.276.928.27h.002c.314.006.495-.071.928-.27l8.335-3.492c.172-.07.277-.124.43-.2 0 0 .284-.211.311-.243.28-.33.285-.621.316-.92a4.261 4.261 0 00.02-.434V1.879c0-1.373-.506-1.88-1.878-1.88zm13.366 3.11h.68c1.138 0 1.688.553 1.688 1.696v1.88h-1.374v-1.8c0-.369-.17-.54-.523-.54h-.235c-.367 0-.537.17-.537.539v5.81c0 .369.17.54.537.54h.262c.353 0 .523-.171.523-.54V8.619h1.373v2.143c0 1.144-.562 1.71-1.7 1.71h-.694c-1.138 0-1.7-.566-1.7-1.71V4.82c0-1.144.562-1.709 1.7-1.709zm-12.186.08h3.114v1.274H6.117v2.603h1.648v1.275H6.117v2.774h1.74v1.275h-3.14zm3.816 0h2.198c1.138 0 1.7.564 1.7 1.708v2.445c0 1.144-.562 1.71-1.7 1.71h-.799v3.338h-1.4zm4.53 0h1.4v9.201h-1.4zm-3.13 1.235v3.392h.575c.354 0 .523-.171.523-.54V4.965c0-.368-.17-.54-.523-.54zm-3.74 10.147a1.708 1.708 0 01.591.108 1.745 1.745 0 01.49.299l-.452.546a1.247 1.247 0 00-.308-.195.91.91 0 00-.363-.068.658.658 0 00-.28.06.703.703 0 00-.224.163.783.783 0 00-.151.243.799.799 0 00-.056.299v.008a.852.852 0 00.056.31.7.7 0 00.157.245.736.736 0 00.238.16.774.774 0 00.303.058.79.79 0 00.445-.116v-.339h-.548v-.565H7.37v1.255a2.019 2.019 0 01-.524.307 1.789 1.789 0 01-.683.123 1.642 1.642 0 01-.602-.107 1.46 1.46 0 01-.478-.3 1.371 1.371 0 01-.318-.455 1.438 1.438 0 01-.115-.58v-.008a1.426 1.426 0 01.113-.57 1.449 1.449 0 01.312-.46 1.418 1.418 0 01.474-.309 1.58 1.58 0 01.598-.111 1.708 1.708 0 01.045 0zm11.963.008a2.006 2.006 0 01.612.094 1.61 1.61 0 01.507.277l-.386.546a1.562 1.562 0 00-.39-.205 1.178 1.178 0 00-.388-.07.347.347 0 00-.208.052.154.154 0 00-.07.127v.008a.158.158 0 00.022.084.198.198 0 00.076.066.831.831 0 00.147.06c.062.02.14.04.236.061a3.389 3.389 0 01.43.122 1.292 1.292 0 01.328.17.678.678 0 01.207.24.739.739 0 01.071.337v.008a.865.865 0 01-.081.382.82.82 0 01-.229.285 1.032 1.032 0 01-.353.18 1.606 1.606 0 01-.46.061 2.16 2.16 0 01-.71-.116 1.718 1.718 0 01-.593-.346l.43-.514c.277.223.578.335.9.335a.457.457 0 00.236-.05.157.157 0 00.082-.142v-.008a.15.15 0 00-.02-.077.204.204 0 00-.073-.066.753.753 0 00-.143-.062 2.45 2.45 0 00-.233-.062 5.036 5.036 0 01-.413-.113 1.26 1.26 0 01-.331-.16.72.72 0 01-.222-.243.73.73 0 01-.082-.36v-.008a.863.863 0 01.074-.359.794.794 0 01.214-.283 1.007 1.007 0 01.34-.185 1.423 1.423 0 01.448-.066 2.006 2.006 0 01.025 0zm-9.358.025h.742l1.183 2.81h-.825l-.203-.499H8.623l-.198.498h-.81zm2.197.02h.814l.663 1.08.663-1.08h.814v2.79h-.766v-1.602l-.711 1.091h-.016l-.707-1.083v1.593h-.754zm3.469 0h2.235v.658h-1.473v.422h1.334v.61h-1.334v.442h1.493v.658h-2.255zm-5.3.897l-.315.793h.624zm-1.145 5.19h8.014l-4.09 1.348z"/>
                            </svg>
                            <span style="font-family: 'Plus Jakarta Sans', sans-serif; font-weight: 800; font-size: 1.3rem; letter-spacing: 1px; color: white;">EPIC GAMES</span>
                        </div>
                        <button class="exhibit-enter-btn epic-banner-btn" style="margin: 0; transition: all 0.3s ease; background: rgba(255, 255, 255, 0.08); color: #ffffff; border-color: rgba(255, 255, 255, 0.25); cursor: pointer; display: inline-flex; align-items: center; gap: 8px;">
                            <span>OPEN EPIC WEB STORE</span>
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6M15 3h6v6M10 14L21 3"/></svg>
                        </button>
                    </div>
                </div>
            </section>
        `;
    }

    function render(games) {
        appendLog(`[Library] Render: ${games ? games.length : 0} games, filter="${filterStatus}", search="${searchQuery}"`, 'log-system');
        if (!games || games.length === 0) {
            content.innerHTML = `
                ${getStorefrontPortalsHtml()}
                <section class="lib-section" style="padding-top: 40px;">
                    <div class="empty-state" style="background: rgba(255, 255, 255, 0.02); border: 1px dashed rgba(255, 255, 255, 0.08); border-radius: 12px; padding: 80px 40px; text-align: center; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px;">
                        <div style="width: 48px; height: 48px; border-radius: 50%; background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.08); display: flex; align-items: center; justify-content: center; font-size: 1.2rem; color: rgba(255,255,255,0.4); margin-bottom: 8px;">◈</div>
                        <h2 style="font-family: 'Plus Jakarta Sans', sans-serif; font-weight: 700; font-size: 1.4rem; color: white; margin: 0;">Tim Cook's Dream Mac</h2>
                        <p style="font-family: 'Inter', sans-serif; font-size: 0.9rem; color: rgba(255,255,255,0.6); margin: 0; max-width: 460px; line-height: 1.6;">Completely empty of games. Connect Steam or Epic below and let's fix that mistake.</p>
                    </div>
                </section>
            `;
            setupGalleryInteractivity(content);
            content.querySelectorAll('.steam-banner-btn').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    e.stopPropagation();
                    if (window.runtime && typeof window.runtime.BrowserOpenURL === 'function') {
                        window.runtime.BrowserOpenURL('https://store.steampowered.com');
                    } else {
                        window.open('https://store.steampowered.com', '_blank', 'noopener,noreferrer');
                    }
                });
            });
            content.querySelectorAll('.epic-banner-btn').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    e.stopPropagation();
                    if (window.runtime && typeof window.runtime.BrowserOpenURL === 'function') {
                        window.runtime.BrowserOpenURL('https://store.epicgames.com');
                    } else {
                        window.open('https://store.epicgames.com', '_blank', 'noopener,noreferrer');
                    }
                });
            });
            return;
        }

        try {

        const getCompatBadge = (game) => {
            if (game.isMacNative) return '<span class="compat-badge native" title="A rare Apple Silicon native build. Cherish it." style="background: rgba(46, 204, 113, 0.15); color: #2ecc71; border: 1px solid rgba(46,204,113,0.3); cursor: help;">✓ Mac Native</span>';
            return '<span class="compat-badge translated" title="Built for DirectX. Running on your Mac through sheer determination." style="cursor: help;">◈ Translated</span>';
        };

        const getPlayButtonHTML = (game) => {
            if (game.status === 'running') {
                return '<button class="exhibit-enter-btn hover-action-btn force-quit-btn" style="background: rgba(239, 68, 68, 0.15); border: 1px solid rgba(239, 68, 68, 0.4); color: #ef4444; pointer-events: auto;">FORCE QUIT</button>';
            }
            if (game.status === 'downloading') {
                return '<button class="exhibit-enter-btn hover-action-btn" style="pointer-events: auto;">DOWNLOADING...</button>';
            }
            if (game.status === 'download_interrupted') {
                return '<button class="exhibit-enter-btn hover-action-btn resume-btn" style="background: rgba(245, 158, 11, 0.25); border: 1px solid rgba(245, 158, 11, 0.6); color: #f59e0b; pointer-events: auto; font-weight: 700;">RESUME DOWNLOAD</button>';
            }
            if (game.status === 'ready' || game.isInstalled) {
                return '<button class="exhibit-enter-btn hover-action-btn" style="pointer-events: auto;">PLAY</button>';
            }
            return '<button class="exhibit-enter-btn hover-action-btn" style="pointer-events: auto;">DOWNLOAD</button>';
        };

        const featuredGame = games.find(g => g.status === 'running') || games[0] || {};
        
        const recentGames = games.filter(g => 
            (g.status === 'ready' || g.status === 'running')
        );

        const filteredGames = games.filter(game => {
            const matchesSearch = (game.name || '').toLowerCase().includes(searchQuery.toLowerCase());
            const matchesStatus = filterStatus === 'all' ||
                (filterStatus === 'running' && game.status === 'running') ||
                (filterStatus === 'ready' && (game.status === 'ready' || game.isInstalled)) ||
                (filterStatus === 'not_installed' && (game.status === 'not_installed' || game.status === 'not_downloaded' || game.status === 'download_interrupted'));
            return matchesSearch && matchesStatus;
        }).sort((a, b) => (a.name || '').localeCompare(b.name || ''));

        // Bifurcate: game.platform is "Steam" or "Epic" — set by Go adapters
        const steamGames = filteredGames.filter(g => g.platform === 'Steam');
        const otherGames = filteredGames.filter(g => g.platform !== 'Steam');

        content.innerHTML = `
            ${renderGalleryHero(games)}

            <!-- EDITORIAL BRAND BREAK -->
            <div class="editorial-block">
                <p class="editorial-quote">"Every frame, <em>translated.</em>"</p>
                <span class="editorial-label">RIFT Engine v1.0 · Consolidated Translation Layer</span>
            </div>
            
            ${recentGames.length > 0 ? `
            <section class="lib-section">
                <div class="lib-section-head">
                    <h2 class="lib-section-title"><span class="text-serif-italic">Resume</span> <span class="text-sans">Translation</span></h2>
                </div>
                <div class="marquee-strip">
                    ${recentGames.map((game, i) => {
                        const isSteam = game.platform === 'Steam' || (game.id && game.id.startsWith('steam-'));
                        const btnHTML = getPlayButtonHTML(game);
                        const isInstalled = (game.status === 'ready' || game.status === 'running' || game.isInstalled);
                        
                        const safeName = escapeHTML(game.name || 'Unknown');
                        const safeBackend = escapeHTML(game.backend || 'Wine-Staging');

                        if (isSteam) {
                            const appId = game.appID || game.id.replace('steam-', '');
                            const imgCover = 'https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/' + appId + '/header.jpg';
                            const fallbackCover = 'https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/' + appId + '/capsule_616x353.jpg';
                            const removeOrUninstall = isInstalled
                                ? '<button class="exhibit-enter-btn uninstall-action-btn" style="background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); color: #ef4444; font-size: 0.68rem; padding: 6px 12px; border-radius: 4px; transition: all 0.2s;">UNINSTALL</button>'
                                : '<button class="exhibit-enter-btn remove-action-btn" style="background: rgba(239, 68, 68, 0.15); border: 1px solid rgba(239, 68, 68, 0.4); color: #f87171; font-size: 0.68rem; padding: 6px 12px; border-radius: 4px; transition: all 0.2s;">REMOVE FROM LIBRARY</button>';

                            return '<div class="exhibit-card resume-exhibit-card steam-resume-card tilt-card stagger-item" data-id="' + game.id + '" style="animation-delay: ' + (i * 80) + 'ms;">'
                                + '<div class="exhibit-frame">'
                                + '<img src="' + imgCover + '" data-appid="' + appId + '" alt="' + safeName + '" loading="lazy" class="steam-card-img" onerror="window.handleCoverError(this, \'' + fallbackCover + '\', \'' + safeName + '\')">'
                                + '<div class="exhibit-gradient"></div>'
                                + (game.status === 'running' ? '<div class="lib-card-live">LIVE</div>' : '')
                                + '<div class="exhibit-placard-overlay steam-placard-overlay">'
                                + '<div class="placard-mono-header">[STEAM PROFILE]</div>'
                                + '<div class="placard-divider"></div>'
                                + '<div class="steam-placard-grid">'
                                + '<div class="placard-stat"><span class="lbl">CIPHER:</span> <span class="val">' + safeBackend + '</span></div>'
                                + '<div class="placard-stat"><span class="lbl">EDITION:</span> <span class="val">Win10 64-bit</span></div>'
                                + '<div class="placard-stat"><span class="lbl">MEDIUM:</span> <span class="val">Wine-Staging</span></div>'
                                + '<div class="placard-stat"><span class="lbl">CADENCE:</span> <span class="val">60 FPS</span></div>'
                                + '</div>'
                                + '<div class="steam-placard-actions">'
                                + btnHTML
                                + removeOrUninstall
                                + '</div>'
                                + '</div>'
                                + '</div>'
                                + '<div class="exhibit-info">'
                                + '<h3 class="exhibit-card-title text-serif-italic" style="margin-top: 8px;">' + safeName + '</h3>'
                                + '<div class="exhibit-card-footer">'
                                + '<span class="exhibit-card-engine">' + safeBackend + '</span>'
                                + getCompatBadge(game)
                                + '</div>'
                                + '</div>'
                                + '</div>';
                        } else {
                            return '<div class="exhibit-card resume-exhibit-card epic-resume-card tilt-card stagger-item" data-id="' + game.id + '" style="animation-delay: ' + (i * 80) + 'ms;">'
                                + '<div class="exhibit-frame">'
                                + '<img src="' + (game.cover || '') + '" alt="' + safeName + '" loading="lazy" onerror="window.handleCoverError(this, \'\', \'' + safeName + '\')">'
                                + '<div class="exhibit-gradient"></div>'
                                + (game.status === 'running' ? '<div class="lib-card-live">LIVE</div>' : '')
                                + '<div class="epic-hover-overlay">'
                                + '<div class="epic-hover-title">' + safeName + '</div>'
                                + btnHTML
                                + (isInstalled ? '<button class="exhibit-enter-btn uninstall-action-btn" style="background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); color: #ef4444; font-size: 0.75rem; padding: 8px 16px; pointer-events: auto; transition: all 0.2s;">UNINSTALL</button>' : '')
                                + '</div>'
                                + '</div>'
                                + '<div class="exhibit-info">'
                                + '<h3 class="exhibit-card-title text-serif-italic" style="margin-top: 8px;">' + safeName + '</h3>'
                                + '<div class="exhibit-card-footer">'
                                + '<span class="exhibit-card-engine">' + safeBackend + '</span>'
                                + getCompatBadge(game)
                                + '</div>'
                                + '</div>'
                                + '</div>';
                        }
                    }).join('')}
                </div>
            </section>
            ` : ''}

            <!-- FILTER BAR -->
            <div class="lib-filter-bar">
                <div class="filter-tabs">
                    <button class="filter-tab ${filterStatus === 'all' ? 'active' : ''}" data-status="all">All Exhibits</button>
                    <button class="filter-tab ${filterStatus === 'running' ? 'active' : ''}" data-status="running">Running</button>
                    <button class="filter-tab ${filterStatus === 'ready' ? 'active' : ''}" data-status="ready">Ready</button>
                </div>
                <div class="filter-search-wrap">
                    <span class="search-icon">⌕</span>
                    <input type="text" class="filter-search" placeholder="Filter archive..." value="${searchQuery}">
                </div>
            </div>

            <!-- STOREFRONT PORTALS -->
            ${getStorefrontPortalsHtml()}

            <!-- STEAM ARTIFACTS REGISTRY (LANDSCAPE 16:9 GRID) -->
            ${steamGames.length > 0 ? `
            <section class="lib-section" style="margin-top: 40px;">
                <div class="lib-section-head">
                    <h2 class="lib-section-title"><span class="text-serif-italic">Steam</span> <span class="text-sans">Artifacts</span></h2>
                    <span class="lib-count">${steamGames.length} exhibits</span>
                </div>
                <div class="steam-exhibit-grid">
                    ${steamGames.map((game, i) => {
                        const WineVer = game.isMacNative ? 'macOS App Bundle' : 'Wine-Staging 11.10';
                        const Resolution = 'Adaptive (ProMotion)';
                        const gridBtnHTML = getPlayButtonHTML(game);
                        const safeName = escapeHTML(game.name || 'Unknown');
                        const safeBackend = escapeHTML(game.backend || 'Wine-Staging');
                        const isInstalled = (game.status === 'ready' || game.status === 'running' || game.isInstalled);
                        const appId = game.appID || game.id.replace('steam-', '');
                        const imgCover = 'https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/' + appId + '/header.jpg';
                        const fallbackCover = 'https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/' + appId + '/capsule_616x353.jpg';
                        
                        const removeOrUninstall = isInstalled
                            ? '<button class="exhibit-enter-btn uninstall-action-btn" style="background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); color: #ef4444; font-size: 0.68rem; padding: 6px 12px; border-radius: 4px; transition: all 0.2s;">UNINSTALL</button>'
                            : '<button class="exhibit-enter-btn remove-action-btn" style="background: rgba(239, 68, 68, 0.15); border: 1px solid rgba(239, 68, 68, 0.4); color: #f87171; font-size: 0.68rem; padding: 6px 12px; border-radius: 4px; transition: all 0.2s;">REMOVE FROM LIBRARY</button>';

                        return '<div class="exhibit-card steam-exhibit-card tilt-card stagger-item" data-id="' + game.id + '" style="animation-delay: ' + ((i + 1) * 60) + 'ms;">'
                            + '<div class="exhibit-frame">'
                            + '<img src="' + imgCover + '" data-appid="' + appId + '" alt="' + safeName + '" loading="lazy" class="steam-card-img" onerror="window.handleCoverError(this, \'' + fallbackCover + '\', \'' + safeName + '\')">'
                            + '<div class="exhibit-gradient"></div>'
                            + (game.status === 'running' ? '<div class="lib-card-live">LIVE</div>' : '')
                            + '<div class="exhibit-placard-overlay steam-placard-overlay">'
                            + '<div class="placard-mono-header">[STEAM PROFILE]</div>'
                            + '<div class="placard-divider"></div>'
                            + '<div class="steam-placard-grid">'
                            + '<div class="placard-stat"><span class="lbl">CIPHER:</span> <span class="val">' + safeBackend + '</span></div>'
                            + '<div class="placard-stat"><span class="lbl">EDITION:</span> <span class="val">Win10 64-bit</span></div>'
                            + '<div class="placard-stat"><span class="lbl">MEDIUM:</span> <span class="val">' + WineVer + '</span></div>'
                            + '<div class="placard-stat"><span class="lbl">CADENCE:</span> <span class="val">' + Resolution + '</span></div>'
                            + '</div>'
                            + '<div class="steam-placard-actions">'
                            + gridBtnHTML
                            + removeOrUninstall
                            + '</div>'
                            + '</div>'
                            + '</div>'
                            + '<div class="exhibit-info">'
                            + '<span class="exhibit-num">STEAM NO. 0' + (i + 1) + '</span>'
                            + '<h3 class="exhibit-card-title text-serif-italic">' + safeName + '</h3>'
                            + '<div class="exhibit-card-footer">'
                            + '<span class="exhibit-card-engine">' + safeBackend + '</span>'
                            + getCompatBadge(game)
                            + '</div>'
                            + '</div>'
                            + '</div>';
                    }).join('')}
                </div>
            </section>
            ` : ''}

            <!-- EPIC & CUSTOM ARTIFACTS REGISTRY (PORTRAIT GRID) -->
            <section class="lib-section" style="margin-top: 40px;">
                <div class="lib-section-head">
                    <h2 class="lib-section-title"><span class="text-serif-italic">Epic & Custom</span> <span class="text-sans">Artifacts</span></h2>
                    <span class="lib-count">${otherGames.length} exhibits</span>
                </div>
                <div class="exhibit-grid">
                    <!-- VACANT EXHIBIT CARD -->
                    <div class="exhibit-card-vacant vacant-import-btn stagger-item" style="animation-delay: 0ms; cursor: pointer;">
                        <div class="vacant-frame">
                            <span class="vacant-plus">+</span>
                            <span class="vacant-title">Vacant Exhibit Slot</span>
                            <span class="vacant-subtitle">Import Artifact</span>
                        </div>
                    </div>

                    ${otherGames.map((game, i) => {
                        const WineVer = game.isMacNative ? 'macOS App Bundle' : 'Wine-Staging 11.10';
                        const Resolution = 'Adaptive (ProMotion)';
                        const gridBtnHTML = getPlayButtonHTML(game);
                        const safeName = escapeHTML(game.name || 'Unknown');
                        const safeBackend = escapeHTML(game.backend || 'Wine-Staging');

                        return `
                        <div class="exhibit-card tilt-card stagger-item" data-id="${game.id}" style="animation-delay: ${(i + 1) * 60}ms;">
                            <div class="exhibit-frame">
                                <img src="${game.cover || ''}" alt="${safeName}" loading="lazy" onerror="window.handleCoverError(this, '', '${safeName}')">
                                <div class="exhibit-gradient"></div>
                                ${game.status === 'running' ? '<div class="lib-card-live">LIVE</div>' : ''}
                                
                                <div class="exhibit-placard-overlay">
                                    <div class="placard-mono-header">[TRANSLATION PROFILE]</div>
                                    <div class="placard-divider"></div>
                                    <div class="placard-stat"><span class="lbl">CIPHER:</span> <span class="val">${safeBackend}</span></div>
                                    <div class="placard-stat"><span class="lbl">MEDIUM:</span> <span class="val">${WineVer}</span></div>
                                    <div class="placard-stat"><span class="lbl">EDITION:</span> <span class="val">Win10 64-bit</span></div>
                                    <div class="placard-stat"><span class="lbl">CADENCE:</span> <span class="val">${Resolution}</span></div>
                                    ${gridBtnHTML}
                                    ${game.status === 'not_installed' ? `
                                        <button class="exhibit-enter-btn locate-action-btn" style="margin-top: 8px; background: rgba(255,255,255,0.08); border: 1px solid rgba(255,255,255,0.15); color: white; font-size: 0.68rem; padding: 6px 12px; border-radius: 4px;">LOCATE FILES</button>
                                    ` : `
                                        <button class="exhibit-enter-btn uninstall-action-btn" style="margin-top: 8px; background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); color: #ef4444; font-size: 0.68rem; padding: 6px 12px; border-radius: 4px; transition: all 0.2s;">UNINSTALL</button>
                                    `}
                                </div>
                            </div>
                            <div class="exhibit-info">
                                <span class="exhibit-num">EXHIBIT NO. 0${i + 1}</span>
                                <h3 class="exhibit-card-title text-serif-italic">${safeName}</h3>
                                <div class="exhibit-card-footer">
                                    <span class="exhibit-card-engine">${safeBackend}</span>
                                    ${getCompatBadge(game)}
                                </div>
                            </div>
                        </div>
                    `}).join('')}
                </div>
            </section>

            <!-- EDITORIAL CLOSER -->
            <div class="editorial-block editorial-block-end">
                <p class="editorial-quote">"Your Windows library.<br><em>Native-grade fidelity.</em>"</p>
            </div>
        `;

        } catch (renderErr) {
            appendLog(`[Library RENDER ERROR] ${renderErr.message}`, 'log-error');
            console.error('[Library] Render crashed:', renderErr);
            content.innerHTML = `<div style="color: #ef4444; padding: 40px; font-family: monospace; white-space: pre-wrap;">Library render error:\n${renderErr.message}\n${renderErr.stack}</div>`;
        }


        setupGalleryInteractivity(content);

        // --- INTERACTIVITY ---

        // Search
        content.querySelector('.filter-search')?.addEventListener('input', (e) => {
            searchQuery = e.target.value;
            appendLog(`[Library] Filter text: "${searchQuery}"`, 'log-info');
            render(store.state.games);
            const input = content.querySelector('.filter-search');
            if (input) { input.focus(); input.setSelectionRange(input.value.length, input.value.length); }
        });

        // Filter tabs
        content.querySelectorAll('.filter-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                filterStatus = tab.dataset.status;
                appendLog(`[Library] Filter tab: ${filterStatus}`, 'log-info');
                render(store.state.games);
            });
        });

        // Settings config clicks (Configure Prefix)
        content.querySelectorAll('.config-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                const gameId = btn.dataset.id;
                if (gameId) {
                    const game = store.state.games.find(g => g.id === gameId);
                    if (!game) return;
                    if (game.isMacNative) {
                        appendLog(`[Config] '${game.name}' runs natively on macOS. No translation capsule to configure.`, 'log-info');
                        return;
                    }

                    appendLog(`[Config] Opening Wine configuration tool (winecfg) for '${game.name}'...`, 'log-system');
                    try {
                        const res = await window.go.rift.App.ConfigureGamePrefix(game.id);
                        appendLog(`[Config] ${res}`, 'log-success');
                    } catch (err) {
                        appendLog(`[Config] Failed to open Wine configuration: ${err}`, 'log-error');
                    }
                }
            });
        });

        // Steam Banner Clicks
        content.querySelectorAll('.steam-banner-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                appendLog('[Storefront] Opening Steam Web Store in browser...', 'log-system');
                if (window.runtime && typeof window.runtime.BrowserOpenURL === 'function') {
                    window.runtime.BrowserOpenURL('https://store.steampowered.com');
                } else {
                    window.open('https://store.steampowered.com', '_blank', 'noopener,noreferrer');
                }
            });
        });

        // Epic Banner Clicks
        content.querySelectorAll('.epic-banner-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                appendLog('[Storefront] Opening Epic Games Web Store in browser...', 'log-system');
                if (window.runtime && typeof window.runtime.BrowserOpenURL === 'function') {
                    window.runtime.BrowserOpenURL('https://store.epicgames.com');
                } else {
                    window.open('https://store.epicgames.com', '_blank', 'noopener,noreferrer');
                }
            });
        });

        // Vacant Exhibit Slot Import click
        content.querySelectorAll('.vacant-import-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                appendLog(`[Import] Selecting directory to import local game artifact...`, 'log-system');
                try {
                    const importedGame = await window.go.rift.App.ImportGameFromFolder();
                    if (importedGame && importedGame.id) {
                        appendLog(`[Import] Successfully imported game: ${importedGame.name}`, 'log-success');
                        
                        // Resync library and update state
                        const updatedGames = await window.go.rift.App.SyncLibrary();
                        store.update('games', updatedGames);
                    }
                } catch (err) {
                    appendLog(`[Import] Cancelled or failed importing folder: ${err}`, 'log-error');
                }
            });
        });

        // Locate Files Buttons
        content.querySelectorAll('.locate-action-btn, .locate-hero-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                const card = btn.closest('[data-id]');
                const gameId = card ? card.dataset.id : btn.dataset.id;
                
                if (gameId) {
                    const game = store.state.games.find(g => g.id === gameId);
                    if (!game) return;

                    appendLog(`[Import] Select local installation folder for '${game.name}'...`, 'log-system');
                    try {
                        const res = await window.go.rift.App.LocateAndImportGame(game.id, game.platform, game.appID);
                        appendLog(`[Import] ${res}`, 'log-success');
                        // Sync library and update state
                        const updatedGames = await window.go.rift.App.SyncLibrary();
                        store.update('games', updatedGames);
                    } catch (err) {
                        appendLog(`[Import] Failed to import game files: ${err}`, 'log-error');
                    }
                }
            });
        });

        // Uninstall Buttons
        content.querySelectorAll('.uninstall-action-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                
                // Double-click confirmation logic
                if (btn.innerText !== 'SURE?') {
                    const originalText = btn.innerText;
                    btn.innerText = 'SURE?';
                    btn.style.background = 'rgba(239, 68, 68, 0.4)';
                    setTimeout(() => {
                        btn.innerText = originalText;
                        btn.style.background = 'rgba(239, 68, 68, 0.1)';
                    }, 3000);
                    return;
                }

                const card = btn.closest('[data-id]');
                const gameId = card ? card.dataset.id : null;
                
                if (gameId) {
                    const game = store.state.games.find(g => g.id === gameId);
                    if (!game) return;

                    btn.innerText = 'DELETING...';
                    appendLog(`[RIFT Engine] Uninstalling '${game.name}' and removing all traces...`, 'log-system');
                    try {
                        const res = await window.go.rift.App.UninstallGame(game.id);
                        appendLog(`[RIFT Engine] ${res}`, 'log-success');
                        
                        // Sync library to reflect the uninstalled state
                        const updatedGames = await window.go.rift.App.SyncLibrary();
                        store.update('games', updatedGames);
                    } catch (err) {
                        appendLog(`[RIFT Engine] Failed to uninstall game: ${err}`, 'log-error');
                        btn.innerText = 'ERROR';
                    }
                }
            });
        });

        // Remove from Library Buttons (for uninstalled Steam games)
        content.querySelectorAll('.remove-action-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                const card = btn.closest('[data-id]');
                const gameId = card ? card.dataset.id : null;

                if (gameId) {
                    btn.innerText = 'REMOVING...';
                    appendLog(`[RIFT Engine] Removing '${gameId}' from library...`, 'log-system');
                    try {
                        await window.go.rift.App.RemoveFromLibrary(gameId);
                        appendLog(`[RIFT Engine] Removed from library.`, 'log-success');
                        const updatedGames = await window.go.rift.App.SyncLibrary();
                        store.update('games', updatedGames);
                    } catch (err) {
                        appendLog(`[RIFT Engine] Failed to remove from library: ${err}`, 'log-error');
                        btn.innerText = 'ERROR';
                    }
                }
            });
        });

        // Play and Install buttons
        content.querySelectorAll('.lib-play-btn, .exhibit-enter-btn:not(.steam-banner-btn):not(.uninstall-action-btn):not(.locate-action-btn), .lib-card-play-btn, .hover-action-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                const card = btn.closest('[data-id]');
                const gameId = card ? card.dataset.id : btn.closest('.lib-hero-actions')?.querySelector('.lib-play-btn')?.dataset.id || btn.dataset.id;
                
                if (gameId) {
                    const game = store.state.games.find(g => g.id === gameId);
                    if (!game) return;

                    if (game.status === 'downloading') {
                        return; // Already downloading, ignore click
                    }

                    if (game.status === 'running') {
                        // FORCE QUIT LOGIC
                        btn.innerText = 'KILLING...';
                        appendLog(`[RIFT Engine] Force quitting '${game.name}'...`, 'log-warn');
                        try {
                            const res = await window.go.rift.App.ForceQuitGame(game.id);
                            appendLog(`[RIFT Engine] ${res}`, 'log-success');
                        } catch(err) {
                            appendLog(`[RIFT Engine] Force quit failed: ${err}`, 'log-error');
                        }
                        return;
                    }

                    if (game.status === 'not_downloaded' || game.status === 'not_installed' || game.status === 'download_interrupted' || game.status === 'failed') {
                        if (store.state.offlineMode) {
                            if (window.showRiftNotification) {
                                window.showRiftNotification('Network Disconnected', 'You cannot download games while in Offline Mode. Please connect to the internet and try again.', 'warn');
                            } else {
                                alert('Network Disconnected. Cannot download in Offline Mode.');
                            }
                            return;
                        }

                        // ONE-CLICK ACQUIRE / RESUME: starts AI + prefix + download + post-setup
                        const isResume = game.status === 'download_interrupted';
                        appendLog(`[RIFT Engine] ${isResume ? 'Resuming download for' : 'Acquiring'} '${game.name}'...`, 'log-system');
                        store.update('activeDownload', {
                            gameId: game.id,
                            percent: 0,
                            speed: 0,
                            status: 'starting',
                            stage: isResume ? 'Resuming partial download...' : 'Starting acquisition pipeline...'
                        });
                        const updated = store.state.games.map(g => {
                            if (g.id === gameId) return { ...g, status: 'downloading' };
                            return g;
                        });
                        store.update('games', updated);

                        try {
                            const res = await window.go.rift.App.AcquireGame(game.id, game.platform, game.appID);
                            appendLog(`[RIFT Engine] ${res}`, 'log-success');
                        } catch (err) {
                            appendLog(`[RIFT Engine] Acquisition failed: ${err}`, 'log-error');
                            store.update('activeDownload', null);
                            const resetGames = store.state.games.map(g => {
                                if (g.id === gameId) return { ...g, status: isResume ? 'download_interrupted' : 'not_downloaded' };
                                return g;
                            });
                            store.update('games', resetGames);
                        }
                    } else {
                        // Game is ready/running — use PlayGame which checks if download exists
                        const updated = store.state.games.map(g => {
                            if (g.id === gameId) return { ...g, status: 'running', lastPlayed: 'Playing now' };
                            if (g.status === 'running') return { ...g, status: 'ready', lastPlayed: 'Just now' };
                            return g;
                        });
                        store.update('games', updated);
                        
                        appendLog(`[RIFT Engine] Launching '${game.name}'...`, 'log-system');

                        try {
                            const res = await window.go.rift.App.PlayGame(game.id, game.platform, game.appID);
                            if (res !== "Acquisition started." && res !== "Download started.") {
                                appendLog(`[RIFT Engine] ${res}`, 'log-success');
                            }
                        } catch (err) {
                            appendLog(`[RIFT Engine] Launch failed: ${err}`, 'log-error');
                            const resetGames = store.state.games.map(g => {
                                if (g.id === gameId) return { ...g, status: 'ready' };
                                return g;
                            });
                            store.update('games', resetGames);
                        }
                    }
                }
            });
        });

        // Parallax tilt
        content.querySelectorAll('.tilt-card').forEach(card => {
            card.addEventListener('mousemove', (e) => {
                const rect = card.getBoundingClientRect();
                const x = (e.clientX - rect.left) / rect.width - 0.5;
                const y = (e.clientY - rect.top) / rect.height - 0.5;
                card.style.transform = `perspective(600px) rotateY(${x * 6}deg) rotateX(${y * -4}deg) translateY(-4px)`;
            });
            card.addEventListener('mouseleave', () => {
                card.style.transform = '';
            });
        });
    }

    const handleGlobalSearch = (e) => {
        searchQuery = (e.detail || '').trim();
        render(store.state.games);
        const searchInput = content.querySelector('.filter-search');
        if (searchInput) {
            searchInput.value = searchQuery;
        }
    };
    window.addEventListener('global-search', handleGlobalSearch);

    const unsubGames = store.subscribe('games', render);
    const unsubLoading = store.subscribe('steamLoading', () => render(store.state.games));
    const unsubAuth = store.subscribe('authStatus', () => render(store.state.games));
    
    // Router cleanup hook to prevent memory leaks
    container._cleanup = () => {
        window.removeEventListener('global-search', handleGlobalSearch);
        unsubGames();
        unsubLoading();
        unsubAuth();
        unsubDL();
    };

    return container;
}
