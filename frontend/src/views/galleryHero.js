import { store } from '../state.js';
import { escapeHTML } from '../main.js';

let cachedHeroGames = null;
let cachedHeroHTML = '';
let lastGameIdsHash = '';

/**
 * Builds an exhibition pool of EXACTLY 18 items strictly from user-owned games.
 * If user has fewer than 4 games, returns empty (no wall).
 * If user has between 4 and 17 games, cycles through their owned games.
 * If user has 18+, takes top 18 (prioritizing installed titles).
 */
function buildExhibitionPool(userGames) {
    const rawGames = Array.isArray(userGames) ? userGames : [];
    
    // 1. Normalize and deduplicate user owned games
    const ownedMap = new Map();

    rawGames.forEach(g => {
        if (!g || !g.name) return;
        const isInst = !!(g.isInstalled || g.installed || g.status === 'ready' || g.status === 'running');
        const appid = g.appID || g.appid || '';

        const item = {
            id: g.id || `game-${Math.random()}`,
            name: g.name,
            cover: g.cover || (appid ? `https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/${appid}/library_600x900.jpg` : ''),
            heroCover: g.heroCover || g.cover || (appid ? `https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/${appid}/library_hero.jpg` : ''),
            backend: g.backend || (g.isMacNative ? 'macOS Native' : 'D3DMetal (GPTK)'),
            isUserOwned: true,
            isInstalled: isInst,
            badgeText: isInst ? 'INSTALLED' : 'OWNED'
        };
        ownedMap.set(g.id, item);
    });

    const ownedList = Array.from(ownedMap.values());
    if (ownedList.length < 4) {
        return [];
    }

    // 2. If user already has 18+ games, prioritize installed games then return 18
    if (ownedList.length >= 18) {
        const sorted = [...ownedList].sort((a, b) => (b.isInstalled ? 1 : 0) - (a.isInstalled ? 1 : 0));
        return sorted.slice(0, 18);
    }

    // 3. For 4 to 17 games, cycle through user-owned games to fill the 18 frames cleanly.
    // Strictly user games — zero unowned showcase games injected!
    const pool = [];
    for (let i = 0; i < 18; i++) {
        pool.push(ownedList[i % ownedList.length]);
    }
    return pool;
}

export function renderGalleryHero(games) {
    const rawGames = Array.isArray(games) ? games : [];

    // Strictly enforce minimum 4 games threshold
    if (rawGames.length < 4) {
        cachedHeroGames = null;
        cachedHeroHTML = '';
        lastGameIdsHash = '';
        return '';
    }

    // Hash based on user game IDs and installed status
    const currentHash = rawGames.map(g => `${g.id}:${g.isInstalled || g.installed || g.status}`).join(',');
    if (cachedHeroGames && cachedHeroHTML && currentHash === lastGameIdsHash) {
        return cachedHeroHTML;
    }

    const selectedGames = buildExhibitionPool(rawGames);
    if (selectedGames.length === 0) {
        cachedHeroGames = null;
        cachedHeroHTML = '';
        lastGameIdsHash = '';
        return '';
    }

    cachedHeroGames = selectedGames;
    lastGameIdsHash = currentHash;
    
    const leftGames = selectedGames.slice(0, 9);
    const rightGames = selectedGames.slice(9, 18);

    const renderWall = (wallGames, wallClass) => {
        return `
            <div class="gallery-wall ${wallClass}">
                ${wallGames.map((game, i) => {
                    const isLandscape = i % 3 === 0; 
                    const imgUrl = isLandscape ? (game.heroCover || game.cover) : (game.cover || game.heroCover);
                    const shapeClass = isLandscape ? 'frame-landscape' : 'frame-portrait';
                    const badgeClass = 'plaque-badge-owned';
                    
                    return `
                        <div class="gallery-frame ${shapeClass}" style="animation-delay: ${Math.random() * 500}ms;" data-id="${escapeHTML(game.id)}">
                            ${game.isInstalled ? '<div class="frame-status-dot" title="Installed on Device"></div>' : ''}
                            <div class="frame-border">
                                <img src="${imgUrl}" alt="${escapeHTML(game.name)}" loading="lazy" />
                            </div>
                            <div class="frame-plaque">
                                <div class="plaque-header">
                                    <span class="plaque-title" title="${escapeHTML(game.name)}">${escapeHTML(game.name)}</span>
                                    <span class="plaque-badge ${badgeClass}">${game.badgeText}</span>
                                </div>
                                <span class="plaque-engine">${escapeHTML(game.backend)}</span>
                            </div>
                        </div>
                    `;
                }).join('')}
            </div>
        `;
    };

    cachedHeroHTML = `
        <section class="gallery-hero-scene">
            <div class="gallery-room">
                ${renderWall(leftGames, 'wall-left')}
                <div class="gallery-corner-crease"></div>
                ${renderWall(rightGames, 'wall-right')}
                <div class="gallery-floor"></div>
                
                <!-- Room Lighting -->
                <div class="gallery-spotlight light-left"></div>
                <div class="gallery-spotlight light-right"></div>
            </div>
            
            <!-- Global Action Bar overlaid at the bottom of the scene -->
            <div class="gallery-curator-bar">
                <h1 class="curator-title"><span class="text-serif-italic">The</span> <span class="text-sans">Archive</span></h1>
                <p class="curator-subtitle">Curated exhibition of ${rawGames.length} owned ${rawGames.length === 1 ? 'title' : 'titles'} across 18 translation frames.</p>
            </div>
        </section>
    `;
    return cachedHeroHTML;
}

export function setupGalleryInteractivity(parent = document) {
    const scene = parent.querySelector ? parent.querySelector('.gallery-hero-scene') : document.querySelector('.gallery-hero-scene');
    const room = parent.querySelector ? parent.querySelector('.gallery-room') : document.querySelector('.gallery-room');
    if (!scene || !room) return;

    // Attach click navigation to /game?id= once
    if (!scene.dataset.navBound) {
        scene.dataset.navBound = 'true';
        scene.addEventListener('click', (e) => {
            const frame = e.target.closest('.gallery-frame');
            if (!frame) return;
            const gameId = frame.dataset.id;
            if (gameId) {
                window.location.hash = `#/game?id=${encodeURIComponent(gameId)}`;
            }
        });
    }

    // Attach 3D tilt tracking once
    if (scene.dataset.interactive === 'true') return;
    scene.dataset.interactive = 'true';

    let currentX = 0, currentY = 0;
    let targetX = 0, targetY = 0;
    let isHovering = false;
    let rafId = null;

    const lerp = (start, end, factor) => start + (end - start) * factor;

    const updateTransform = () => {
        if (!isHovering && Math.abs(currentX - targetX) < 0.01 && Math.abs(currentY - targetY) < 0.01) {
            cancelAnimationFrame(rafId);
            rafId = null;
            return;
        }

        currentX = lerp(currentX, targetX, 0.08); 
        currentY = lerp(currentY, targetY, 0.08);
        
        room.style.transform = `rotateX(${currentX}deg) rotateY(${currentY}deg)`;
        rafId = requestAnimationFrame(updateTransform);
    };

    scene.addEventListener('mousemove', (e) => {
        isHovering = true;
        const rect = scene.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        const centerX = rect.width / 2;
        const centerY = rect.height / 2;
        
        targetY = ((x - centerX) / centerX) * 22; 
        targetX = -((y - centerY) / centerY) * 6;

        if (!rafId) {
            rafId = requestAnimationFrame(updateTransform);
        }
    });

    scene.addEventListener('mouseleave', () => {
        isHovering = false;
        targetX = 0;
        targetY = 0;
        if (!rafId) {
            rafId = requestAnimationFrame(updateTransform);
        }
    });
}

/**
 * Shuffles and updates the covers and titles in the 18 hero gallery frames in-place.
 * Leaves the 3D room, lighting, layout, and rotations completely untouched for instant (<1ms) performance.
 */
export function refreshGalleryBanners(parent = document, games = null) {
    const scene = parent.querySelector ? parent.querySelector('.gallery-hero-scene') : document.querySelector('.gallery-hero-scene');
    if (!scene) return;

    const frames = scene.querySelectorAll('.gallery-frame');
    if (!frames || frames.length === 0) return;

    if (!games || games.length === 0) {
        games = (store && store.state && store.state.games) || cachedHeroGames || [];
    }
    if (!games || games.length < 4) return;

    const pool = buildExhibitionPool(games);
    if (pool.length === 0) return;
    const shuffled = [...pool].sort(() => 0.5 - Math.random());

    frames.forEach((frame, i) => {
        const game = shuffled[i % shuffled.length];
        if (!game) return;

        frame.dataset.id = game.id;
        const img = frame.querySelector('img');
        const isLandscape = frame.classList.contains('frame-landscape');
        const imgUrl = isLandscape ? (game.heroCover || game.cover) : (game.cover || game.heroCover);

        if (img && imgUrl && img.getAttribute('src') !== imgUrl) {
            img.src = imgUrl;
            img.alt = game.name || '';
        }

        const titleSpan = frame.querySelector('.plaque-title');
        if (titleSpan && game.name) {
            titleSpan.textContent = game.name;
            titleSpan.title = game.name;
        }

        const engineSpan = frame.querySelector('.plaque-engine');
        if (engineSpan) {
            engineSpan.textContent = game.backend || '';
        }

        const badgeSpan = frame.querySelector('.plaque-badge');
        if (badgeSpan) {
            badgeSpan.textContent = game.badgeText || (game.isInstalled ? 'INSTALLED' : 'OWNED');
            badgeSpan.className = 'plaque-badge plaque-badge-owned';
        }

        // Status dot update
        let dot = frame.querySelector('.frame-status-dot');
        if (game.isInstalled) {
            if (!dot) {
                dot = document.createElement('div');
                dot.className = 'frame-status-dot';
                dot.title = 'Installed on Device';
                frame.prepend(dot);
            }
        } else if (dot) {
            dot.remove();
        }
    });
}

