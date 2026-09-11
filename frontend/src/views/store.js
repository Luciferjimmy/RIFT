import { store } from '../state.js';
import { appendLog } from '../main.js';
import { GetCoverArtMap } from '../../wailsjs/go/rift/App.js';

// Local client storage cache for cover art
let coverArtCache = null;
function getLocalCoverCache() {
    if (coverArtCache) return coverArtCache;
    try {
        const stored = window.localStorage.getItem('rift_cover_map');
        if (stored) {
            coverArtCache = JSON.parse(stored);
            return coverArtCache;
        }
    } catch (e) {}
    coverArtCache = {};
    return coverArtCache;
}

async function getCoverMap(appIDs) {
    if (!Array.isArray(appIDs) || appIDs.length === 0) return getLocalCoverCache();
    const local = getLocalCoverCache();
    const missing = appIDs.filter(id => !local[id]);
    if (missing.length === 0) {
        return local;
    }
    try {
        const raw = await GetCoverArtMap(JSON.stringify(missing));
        const parsed = JSON.parse(raw);
        Object.assign(local, parsed);
        try {
            window.localStorage.setItem('rift_cover_map', JSON.stringify(local));
        } catch (e) {}
        return local;
    } catch (e) {
        return local;
    }
}

// Royal SVG icons (inline, small, elegant)
const royalIcons = {
    crown:   `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><path d="M2 18h20M4 18l2-10 4 4 2-6 2 6 4-4 2 10"/></svg>`,
    sword:   `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><path d="M12 2v14M8 12l4 4 4-4M7 20h10"/></svg>`,
    shield:  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><path d="M12 2l8 4v6c0 5.5-3.8 9.7-8 11-4.2-1.3-8-5.5-8-11V6l8-4z"/></svg>`,
    compass: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><circle cx="12" cy="12" r="10"/><polygon points="16.2,7.8 14,14 7.8,16.2 10,10" fill="currentColor" opacity="0.4"/></svg>`,
    bolt:    `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><polygon points="13,2 3,14 12,14 11,22 21,10 12,10"/></svg>`,
    eye:     `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8S1 12 1 12z"/><circle cx="12" cy="12" r="3"/></svg>`,
    flame:   `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><path d="M12 22c-4-2-7-6-7-10 0-3 2-5 3-7 1 2 3 3 4 3 0-3 1-6 3-8 1 3 4 6 4 12 0 4-3 8-7 10z"/></svg>`,
    gem:     `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><polygon points="6,3 18,3 22,9 12,21 2,9"/><line x1="2" y1="9" x2="22" y2="9"/><line x1="12" y1="21" x2="6" y2="9"/><line x1="12" y1="21" x2="18" y2="9"/></svg>`,
};

// Dinner Party Guests — mapped directly to the 8 genre rooms
const exhibits = [
    { id: 'open-world', name: 'Open World Collection', img: 'left_side_female_1.png', genre: 'Open World', icon: royalIcons.compass },
    { id: 'rpg', name: 'RPG Archive', img: 'left_side_female_2.png', genre: 'RPG', icon: royalIcons.crown },
    { id: 'souls', name: 'Souls & Hardcore', img: 'man1.png', genre: 'Souls', icon: royalIcons.flame },
    { id: 'adventure', name: 'Adventure Wing', img: 'left_side_female3.png', genre: 'Adventure', icon: royalIcons.compass },
    { id: 'sci-fi', name: 'Sci-Fi & Cyberpunk', img: 'right_side_female_1.png', genre: 'Sci-Fi', icon: royalIcons.bolt },
    { id: 'thriller', name: 'Thriller & Mystery', img: 'man3.png', genre: 'Thriller', icon: royalIcons.eye },
    { id: 'action', name: 'Action Vault', img: 'right_side_female_2.png', genre: 'Action', icon: royalIcons.sword },
    { id: 'indie', name: 'Indie & Experimental', img: 'man2.png', genre: 'Indie', icon: royalIcons.gem },
];

// Curated Catalog for all 8 genre mood rooms (Steam AppIDs for 600x900 box art)
const catalogGames = [
    // 1. OPEN WORLD COLLECTION
    { id: 'red-dead-2', name: 'Red Dead Redemption 2', appid: 1174180, price: '$59.99', compat: 'playable', backend: 'DXVK', genre: 'open-world' },
    { id: 'gtav', name: 'Grand Theft Auto V', appid: 271590, price: '$29.99', compat: 'verified', backend: 'DXVK', genre: 'open-world' },
    { id: 'cyberpunk-2077', name: 'Cyberpunk 2077', appid: 1091500, price: '$59.99', compat: 'verified', backend: 'D3DMetal', genre: 'open-world' },
    { id: 'ghost-of-tsushima', name: 'Ghost of Tsushima', appid: 2215430, price: '$59.99', compat: 'verified', backend: 'D3DMetal', genre: 'open-world' },
    { id: 'horizon', name: 'Horizon Zero Dawn', appid: 1151640, price: '$19.99', compat: 'playable', backend: 'DXVK', genre: 'open-world' },
    { id: 'days-gone', name: 'Days Gone', appid: 1259420, price: '$49.99', compat: 'verified', backend: 'D3DMetal', genre: 'open-world' },
    { id: 'death-stranding', name: 'Death Stranding DC', appid: 1850570, price: '$39.99', compat: 'verified', backend: 'D3DMetal', genre: 'open-world' },
    { id: 'ac-odyssey', name: "Assassin's Creed Odyssey", appid: 812140, price: '$59.99', compat: 'playable', backend: 'DXVK', genre: 'open-world' },

    // 2. THE RPG ARCHIVE
    { id: 'baldurs-gate-3', name: "Baldur's Gate 3", appid: 1086940, price: '$59.99', compat: 'verified', backend: 'D3DMetal', genre: 'rpg' },
    { id: 'skyrim', name: 'Skyrim Special Edition', appid: 489830, price: '$39.99', compat: 'verified', backend: 'DXVK', genre: 'rpg' },
    { id: 'witcher-3-store', name: 'The Witcher 3: Wild Hunt', appid: 292030, price: '$19.99', compat: 'verified', backend: 'DXVK', genre: 'rpg' },
    { id: 'fallout-4', name: 'Fallout 4', appid: 377160, price: '$19.99', compat: 'playable', backend: 'DXVK', genre: 'rpg' },
    { id: 'fnv', name: 'Fallout: New Vegas', appid: 22490, price: '$9.99', compat: 'verified', backend: 'DXVK', genre: 'rpg' },
    { id: 'starfield', name: 'Starfield', appid: 1716740, price: '$69.99', compat: 'playable', backend: 'D3DMetal', genre: 'rpg' },
    { id: 'diablo-iv', name: 'Diablo IV', appid: 2344520, price: '$69.99', compat: 'verified', backend: 'D3DMetal', genre: 'rpg' },
    { id: 'persona-5-royal', name: 'Persona 5 Royal', appid: 1687950, price: '$59.99', compat: 'verified', backend: 'DXVK', genre: 'rpg' },
    { id: 'dragons-dogma-2', name: "Dragon's Dogma 2", appid: 2054970, price: '$69.99', compat: 'playable', backend: 'D3DMetal', genre: 'rpg' },
    { id: 'mass-effect-le', name: 'Mass Effect Legendary Edition', appid: 1328670, price: '$59.99', compat: 'verified', backend: 'DXVK', genre: 'rpg' },

    // 3. SOULS & HARDCORE
    { id: 'elden-ring', name: 'Elden Ring', appid: 1245620, price: '$59.99', compat: 'verified', backend: 'D3DMetal', genre: 'souls' },
    { id: 'sekiro', name: 'Sekiro: Shadows Die Twice', appid: 814380, price: '$59.99', compat: 'verified', backend: 'DXVK', genre: 'souls' },
    { id: 'dark-souls-3', name: 'Dark Souls III', appid: 374320, price: '$59.99', compat: 'verified', backend: 'DXVK', genre: 'souls' },
    { id: 'dark-souls-remastered', name: 'Dark Souls: Remastered', appid: 570940, price: '$39.99', compat: 'verified', backend: 'DXVK', genre: 'souls' },
    { id: 'lies-of-p', name: 'Lies of P', appid: 1627720, price: '$59.99', compat: 'verified', backend: 'D3DMetal', genre: 'souls' },
    { id: 'mortal-shell', name: 'Mortal Shell', appid: 1110910, price: '$29.99', compat: 'verified', backend: 'D3DMetal', genre: 'souls' },
    { id: 'lords-of-the-fallen', name: 'Lords of the Fallen', appid: 1501750, price: '$59.99', compat: 'playable', backend: 'D3DMetal', genre: 'souls' },
    { id: 'nioh-2', name: 'Nioh 2 - Complete Edition', appid: 1325200, price: '$49.99', compat: 'playable', backend: 'DXVK', genre: 'souls' },
    { id: 'remnant-2', name: 'Remnant II', appid: 1282100, price: '$49.99', compat: 'playable', backend: 'D3DMetal', genre: 'souls' },

    // 4. ADVENTURE WING
    { id: 'god-of-war', name: 'God of War', appid: 1593500, price: '$39.99', compat: 'verified', backend: 'D3DMetal', genre: 'adventure' },
    { id: 'spiderman', name: "Spider-Man Remastered", appid: 1817070, price: '$49.99', compat: 'verified', backend: 'D3DMetal', genre: 'adventure' },
    { id: 'uncharted-lot', name: 'Uncharted: Legacy of Thieves', appid: 1659420, price: '$49.99', compat: 'playable', backend: 'D3DMetal', genre: 'adventure' },
    { id: 'stray', name: 'Stray', appid: 1332010, price: '$29.99', compat: 'verified', backend: 'DXVK', genre: 'adventure' },
    { id: 'kena', name: 'Kena: Bridge of Spirits', appid: 1954200, price: '$39.99', compat: 'playable', backend: 'D3DMetal', genre: 'adventure' },
    { id: 'tomb-raider', name: 'Shadow of the Tomb Raider', appid: 750920, price: '$39.99', compat: 'verified', backend: 'DXVK', genre: 'adventure' },
    { id: 'it-takes-two', name: 'It Takes Two', appid: 1426210, price: '$39.99', compat: 'playable', backend: 'DXVK', genre: 'adventure' },
    { id: 'plague-tale-requiem', name: 'A Plague Tale: Requiem', appid: 1452590, price: '$49.99', compat: 'playable', backend: 'D3DMetal', genre: 'adventure' },
    { id: 'detroit-become-human', name: 'Detroit: Become Human', appid: 1222140, price: '$39.99', compat: 'verified', backend: 'DXVK', genre: 'adventure' },

    // 5. SCI-FI & CYBERPUNK
    { id: 'doom-eternal', name: 'DOOM Eternal', appid: 782330, price: '$39.99', compat: 'verified', backend: 'D3DMetal', genre: 'sci-fi' },
    { id: 'control', name: 'Control', appid: 870780, price: '$29.99', compat: 'playable', backend: 'DXVK', genre: 'sci-fi' },
    { id: 'dead-space', name: 'Dead Space (2023)', appid: 1693980, price: '$59.99', compat: 'playable', backend: 'D3DMetal', genre: 'sci-fi' },
    { id: 'no-mans-sky', name: "No Man's Sky", appid: 275850, price: '$59.99', compat: 'verified', backend: 'Metal', genre: 'sci-fi' },
    { id: 'returnal', name: 'Returnal', appid: 1898920, price: '$59.99', compat: 'playable', backend: 'D3DMetal', genre: 'sci-fi' },
    { id: 'titanfall-2', name: 'Titanfall 2', appid: 1237970, price: '$29.99', compat: 'verified', backend: 'DXVK', genre: 'sci-fi' },
    { id: 'halo-mcc', name: 'Halo: Master Chief Collection', appid: 976730, price: '$39.99', compat: 'verified', backend: 'DXVK', genre: 'sci-fi' },
    { id: 'subnautica', name: 'Subnautica', appid: 264710, price: '$29.99', compat: 'verified', backend: 'DXVK', genre: 'sci-fi' },

    // 6. THRILLER & MYSTERY
    { id: 're4', name: 'Resident Evil 4', appid: 2050650, price: '$59.99', compat: 'verified', backend: 'D3DMetal', genre: 'thriller' },
    { id: 're-village', name: 'Resident Evil Village', appid: 1196590, price: '$39.99', compat: 'verified', backend: 'Metal', genre: 'thriller' },
    { id: 're2-remake', name: 'Resident Evil 2', appid: 883710, price: '$39.99', compat: 'verified', backend: 'D3DMetal', genre: 'thriller' },
    { id: 'outlast', name: 'Outlast', appid: 238320, price: '$19.99', compat: 'verified', backend: 'DXVK', genre: 'thriller' },
    { id: 'the-evil-within-2', name: 'The Evil Within 2', appid: 601430, price: '$39.99', compat: 'playable', backend: 'DXVK', genre: 'thriller' },
    { id: 'alien-isolation', name: 'Alien: Isolation', appid: 214490, price: '$39.99', compat: 'verified', backend: 'DXVK', genre: 'thriller' },
    { id: 'soma', name: 'SOMA', appid: 282140, price: '$29.99', compat: 'verified', backend: 'DXVK', genre: 'thriller' },
    { id: 'alan-wake', name: 'Alan Wake Remastered', appid: 108710, price: '$29.99', compat: 'verified', backend: 'DXVK', genre: 'thriller' },
    { id: 'little-nightmares-2', name: 'Little Nightmares II', appid: 860510, price: '$29.99', compat: 'verified', backend: 'DXVK', genre: 'thriller' },

    // 7. THE ACTION VAULT
    { id: 'mh-world', name: 'Monster Hunter: World', appid: 582010, price: '$29.99', compat: 'playable', backend: 'DXVK', genre: 'action' },
    { id: 'mh-rise', name: 'Monster Hunter Rise', appid: 1446780, price: '$39.99', compat: 'verified', backend: 'D3DMetal', genre: 'action' },
    { id: 'dmc5', name: 'Devil May Cry 5', appid: 601150, price: '$29.99', compat: 'verified', backend: 'D3DMetal', genre: 'action' },
    { id: 'armored-core-6', name: 'Armored Core VI', appid: 1888160, price: '$59.99', compat: 'verified', backend: 'D3DMetal', genre: 'action' },
    { id: 'batman-arkham-knight', name: 'Batman: Arkham Knight', appid: 208650, price: '$19.99', compat: 'verified', backend: 'DXVK', genre: 'action' },
    { id: 'mgsv', name: 'Metal Gear Solid V', appid: 287700, price: '$19.99', compat: 'verified', backend: 'DXVK', genre: 'action' },
    { id: 'sifu', name: 'Sifu', appid: 2138710, price: '$39.99', compat: 'verified', backend: 'D3DMetal', genre: 'action' },
    { id: 'ghostrunner', name: 'Ghostrunner', appid: 1139900, price: '$29.99', compat: 'verified', backend: 'DXVK', genre: 'action' },
    { id: 'hi-fi-rush', name: 'Hi-Fi RUSH', appid: 1817230, price: '$29.99', compat: 'verified', backend: 'D3DMetal', genre: 'action' },

    // 8. THE INDIE GARDEN
    { id: 'hades', name: 'Hades', appid: 1145360, price: '$24.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
    { id: 'hollow-knight', name: 'Hollow Knight', appid: 367520, price: '$14.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
    { id: 'dead-cells', name: 'Dead Cells', appid: 588650, price: '$24.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
    { id: 'celeste', name: 'Celeste', appid: 504230, price: '$19.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
    { id: 'stardew', name: 'Stardew Valley', appid: 413150, price: '$14.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
    { id: 'portal-2', name: 'Portal 2', appid: 620, price: '$9.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
    { id: 'cuphead', name: 'Cuphead', appid: 268910, price: '$19.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
    { id: 'ori', name: 'Ori: Will of the Wisps', appid: 1057090, price: '$29.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
    { id: 'tunic', name: 'Tunic', appid: 553420, price: '$29.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
    { id: 'sea-of-stars', name: 'Sea of Stars', appid: 1244090, price: '$34.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
    { id: 'dave-the-diver', name: 'Dave the Diver', appid: 1868140, price: '$19.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
    { id: 'balatro', name: 'Balatro', appid: 2379780, price: '$14.99', compat: 'verified', backend: 'DXVK', genre: 'indie' },
];

const moodRooms = [
    { key: 'open-world', accent: '245, 158, 11', serif: 'The', sans: 'Open World Collection' },
    { key: 'rpg', accent: '168, 85, 247', serif: 'The', sans: 'RPG Archive' },
    { key: 'souls', accent: '239, 68, 68', serif: 'The', sans: 'Souls & Hardcore' },
    { key: 'adventure', accent: '56, 189, 248', serif: 'The', sans: 'Adventure Wing' },
    { key: 'sci-fi', accent: '6, 182, 212', serif: 'The', sans: 'Sci-Fi & Cyberpunk' },
    { key: 'thriller', accent: '129, 140, 248', serif: 'The', sans: 'Thriller & Mystery' },
    { key: 'action', accent: '244, 63, 94', serif: 'The', sans: 'Action Vault' },
    { key: 'indie', accent: '52, 211, 153', serif: 'The', sans: 'Indie Garden' },
];

const clipPaths = {
    rhombus: 'polygon(15% 0%, 100% 0%, 85% 100%, 0% 100%)',
    hexagon: 'polygon(25% 0%, 75% 0%, 100% 50%, 75% 100%, 25% 100%, 0% 50%)',
    circle: 'circle(50% at 50% 50%)',
    triangle: 'polygon(50% 5%, 95% 95%, 5% 95%)',
    pill: 'inset(0 round 999px)',
    shield: 'polygon(50% 0%, 100% 35%, 80% 100%, 20% 100%, 0% 35%)',
    diamond: 'polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%)',
};

export async function renderStore() {
    const container = document.createElement('div');
    container.className = 'view-premium';

    // Synchronous immediate cover map from memory / localStorage (0ms blocking!)
    const coverMap = getLocalCoverCache();
    const catalogAppIDs = catalogGames.map(g => String(g.appid));

    // Asynchronously fetch any missing covers in background without delaying store display
    getCoverMap(catalogAppIDs).then(updatedMap => {
        container.querySelectorAll('.mood-room-img').forEach(img => {
            const appId = img.dataset.appid;
            if (appId && updatedMap[appId]) {
                const newSrc = updatedMap[appId].cover || updatedMap[appId].banner;
                if (newSrc && img.getAttribute('src') !== newSrc) {
                    img.src = newSrc;
                }
            }
        });
    });

    container.innerHTML = `
        <!-- AMBIENT GLOW -->
        <div class="ambient-glow" style="background: radial-gradient(ellipse 80% 50% at 50% 0%, rgba(168, 85, 247, 0.06) 0%, transparent 70%);"></div>

        <!-- SUBTLE BACKGROUND DOT GRID -->
        <div class="exhibition-bg-grid"></div>

        <!-- THE LAST SUPPER HEADER & EXHIBITION (Hides on search) -->
        <div id="exhibition-hero-section">
            <div class="store-header">
                <h1 class="store-canvas-title"><span class="text-serif-italic">The</span> <span class="text-display">Exhibition</span></h1>
                <p class="store-canvas-subtitle">Curated PC artifacts framed as translation profiles. Not a store — a gallery.</p>
            </div>

            <div class="dinner-party-container">
                <div class="dinner-party-scene">
                    <img class="dinner-bg-img" src="/images/dinner/base.png" alt="Dinner Party">
                    
                    ${exhibits.map((ex, i) => `
                        <div class="dinner-guest guest-seat-${i + 1}" data-id="${ex.id}">
                            <div class="guest-label label-seat-${i + 1}">
                                <span class="guest-label-icon">${ex.icon}</span>
                                <span class="guest-label-text">${ex.genre.toUpperCase()}</span>
                            </div>
                            <div class="guest-bust-container">
                                <img class="guest-bust-img" src="/images/dinner/${ex.img}" alt="${ex.name}">
                            </div>
                            <div class="head-splatter"></div>
                        </div>
                    `).join('')}

                    <div class="dinner-scene-overlay">
                        <p class="dinner-scene-quote">"Every seat at the table is a world waiting to be translated."</p>
                    </div>
                </div>
            </div>

            <div class="editorial-block">
                <p class="editorial-quote">"Not a store.<br><em>A gallery of playable artifacts.</em>"</p>
                <span class="editorial-label">SCROLL TO EXPLORE CURATED GENRE ROOMS ↓<br><br>EST. COLLECTION: ${catalogGames.length} ARTIFACTS</span>
            </div>
        </div>

        <!-- FREE GAMES SEARCH (Steam App List) -->
        <section class="lib-section free-games-section">
            <div class="lib-section-head">
                <h2 class="lib-section-title"><span class="text-serif-italic">Free</span> <span class="text-sans">Games Search</span></h2>
                <span class="lib-count">Powered by Steam</span>
            </div>
            <div class="free-games-body">
                <div class="free-games-controls">
                    <div class="free-games-search-wrap">
                        <input type="text" id="steam-app-search-input" placeholder="Search Steam games by name..." class="free-games-input">
                        <span class="free-games-search-icon">⌕</span>
                    </div>
                    <button id="steam-app-search-btn" class="exhibit-enter-btn free-games-btn">Search</button>
                    <button id="steam-free-toggle" class="exhibit-enter-btn free-games-btn free-toggle-btn">Free Only</button>
                </div>
                <div id="steam-app-results" class="free-games-scroll">
                    <!-- Results loaded dynamically -->
                </div>
                <div id="steam-app-status" class="free-games-status"></div>
                <div id="steam-show-more-wrap" class="free-games-show-more" style="display: none;">
                    <button id="steam-show-more-btn" class="exhibit-enter-btn">Show All (<span id="steam-total-count">0</span>)</button>
                </div>
            </div>
        </section>

        <!-- MOOD ROOMS -->
        <div class="mood-rooms-grid" id="mood-rooms-container">
            ${moodRooms.map(room => {
            const roomGames = catalogGames.filter(g => g.genre === room.key);
            return `
            <section class="mood-room" data-genre="${room.key}" style="--room-accent: ${room.accent};">
                <div class="mood-room-header">
                    <div class="mood-room-accent-line"></div>
                    <h2 class="mood-room-title"><span class="text-serif-italic">${room.serif}</span> <span class="text-sans">${room.sans}</span></h2>
                    <span class="lib-count">${roomGames.length} artifacts</span>
                </div>
                <div class="mood-room-scroll">
                    ${roomGames.map((game, i) => {
                        const levelClass = game.compat === 'verified' ? 'native' : 'translated';
                        const levelLabel = game.compat === 'verified' ? '✓ Native' : '◈ Translated';
                        const imgSrc = coverMap[String(game.appid)]?.cover || coverMap[String(game.appid)]?.banner || `https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/${game.appid}/library_600x900.jpg`;
                        return `
                        <div class="mood-card tilt-card stagger-item searchable-card" data-id="${game.id}" data-name="${game.name.toLowerCase()}" data-genre="${game.genre}" style="animation-delay: ${i * 60}ms;">
                            <div class="mood-card-img">
                                <img src="${imgSrc}" alt="${game.name}" loading="lazy" referrerpolicy="no-referrer" class="mood-room-img" data-appid="${game.appid}">
                                <div class="mood-card-gradient"></div>
                                <div class="mood-card-hover">
                                    <button class="mood-card-btn">Acquire</button>
                                </div>
                            </div>
                            <div class="mood-card-info">
                                <span class="lib-card-name">${game.name}</span>
                                <div class="lib-card-footer">
                                    <span class="lib-card-meta">${game.price}</span>
                                    <span class="compat-badge ${levelClass}">${levelLabel} · ${game.backend}</span>
                                </div>
                            </div>
                        </div>
                    `}).join('')}
                </div>
            </section>
            `;
        }).join('')}
        </div>

        <!-- SEARCH NO RESULTS -->
        <div id="search-empty-state" class="empty-state" style="display: none; padding-top: 100px;">
            <div class="empty-icon" style="font-size: 32px; margin-bottom: 20px;">✧</div>
            <h2 style="font-family: var(--font-serif); font-style: italic; font-size: 24px; color: rgba(255,255,255,0.9);">No artifacts found</h2>
            <p style="color: rgba(255,255,255,0.4); font-size: 14px;">Try searching for another title or genre.</p>
        </div>

        <div class="editorial-block editorial-block-end" id="editorial-footer">
            <p class="editorial-quote">"Built for macOS.<br><em>Plays like Windows never left.</em>"</p>
        </div>
    `;

    // --- INTERACTIVITY ---

    // Parallax tilt on mood cards
    container.querySelectorAll('.tilt-card').forEach(card => {
        card.addEventListener('mousemove', (e) => {
            const rect = card.getBoundingClientRect();
            const x = (e.clientX - rect.left) / rect.width - 0.5;
            const y = (e.clientY - rect.top) / rect.height - 0.5;
            card.style.transform = `perspective(600px) rotateY(${x * 5}deg) rotateX(${y * -3}deg) translateY(-4px)`;
        });
        card.addEventListener('mouseleave', () => {
            card.style.transform = `perspective(600px) rotateY(0deg) rotateX(0deg) translateY(0)`;
        });

        // Click to route to Game Detail
        card.addEventListener('click', () => {
            const gameId = card.dataset.id;
            window.location.hash = `/game?id=${gameId}`;
        });
    });

    // Image error fallback for mood-room covers (same Steam CDN fallback chain)
    container.querySelectorAll('.mood-room-img').forEach(img => {
        const appId = img.dataset.appid;
        let fb = 0;
        img.addEventListener('error', function() {
            fb++;
            if (fb === 1) this.src = `https://steamcdn-a.akamaihd.net/steam/apps/${appId}/library_600x900.jpg`;
            else if (fb === 2) this.src = `https://steamcdn-a.akamaihd.net/steam/apps/${appId}/library_600x900_2x.jpg`;
            else if (fb === 3) this.src = `https://steamcdn-a.akamaihd.net/steam/apps/${appId}/header.jpg`;
            else this.parentElement.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100%;background:rgba(255,255,255,0.03);color:rgba(255,255,255,0.15);font-size:2rem;">⌕</div>';
        });
    });

    // Cinematic smooth cubic-eased scroll function
    function cinematicScrollTo(targetElement, duration = 1100) {
        const scrollContainer = document.getElementById('app-content') || document.documentElement;
        const startY = scrollContainer.scrollTop || window.pageYOffset || 0;
        const rect = targetElement.getBoundingClientRect();
        const containerRect = scrollContainer.getBoundingClientRect ? scrollContainer.getBoundingClientRect() : { top: 0 };
        const targetY = startY + (rect.top - containerRect.top) - 20;
        const distance = targetY - startY;
        let startTime = null;

        function easeInOutCubic(t) {
            return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
        }

        function step(timestamp) {
            if (!startTime) startTime = timestamp;
            const elapsed = timestamp - startTime;
            const progress = Math.min(elapsed / duration, 1);
            const currentPosition = startY + distance * easeInOutCubic(progress);
            
            scrollContainer.scrollTop = currentPosition;
            if (scrollContainer === document.documentElement) {
                window.scrollTo(0, currentPosition);
            }

            if (progress < 1) {
                requestAnimationFrame(step);
            }
        }
        requestAnimationFrame(step);
    }

    // Attach smooth-scroll click listeners to all dinner party guests
    container.querySelectorAll('.dinner-guest').forEach(guest => {
        const genreId = guest.getAttribute('data-id');
        guest.addEventListener('click', (e) => {
            e.stopPropagation();
            const targetRoom = container.querySelector(`.mood-room[data-genre="${genreId}"]`);
            if (targetRoom) {
                targetRoom.scrollIntoView({ behavior: 'smooth', block: 'start' });
            }
        });
    });

    // SEARCH LOGIC
    const heroSection = container.querySelector('#exhibition-hero-section');
    const footer = container.querySelector('#editorial-footer');
    const emptyState = container.querySelector('#search-empty-state');
    const allCards = container.querySelectorAll('.searchable-card');
    const allRooms = container.querySelectorAll('.mood-room');

    // Listen to the global search event fired from the header (index.html/main.js)
    const handleGlobalSearch = (e) => {
        const query = (e.detail || '').toLowerCase().trim();
        
        if (query === '') {
            // Restore everything
            if (heroSection) heroSection.style.display = 'block';
            if (footer) footer.style.display = 'block';
            if (emptyState) emptyState.style.display = 'none';
            document.getElementById('mood-rooms-container').style.display = 'block';
            const searchResults = document.getElementById('search-results-container');
            if (searchResults) searchResults.style.display = 'none';
            return;
        }

        // Hide hero exhibition stuff and normal rooms
        if (heroSection) heroSection.style.display = 'none';
        if (footer) footer.style.display = 'none';
        document.getElementById('mood-rooms-container').style.display = 'none';

        let searchResults = document.getElementById('search-results-container');
        if (!searchResults) {
            searchResults = document.createElement('div');
            searchResults.id = 'search-results-container';
            searchResults.className = 'search-results-grid';
            searchResults.style.padding = '40px';
            searchResults.style.display = 'grid';
            searchResults.style.gridTemplateColumns = 'repeat(auto-fill, minmax(200px, 1fr))';
            searchResults.style.gap = '24px';
            
            // Insert it before the empty state
            emptyState.parentNode.insertBefore(searchResults, emptyState);
        }
        
        searchResults.style.display = 'grid';
        searchResults.innerHTML = ''; // clear previous

        let totalMatches = 0;
        
        catalogGames.forEach((game, i) => {
            const name = (game.name || '').toLowerCase();
            const genre = (game.genre || '').toLowerCase();
            if (name.includes(query) || genre.includes(query)) {
                totalMatches++;
                const levelClass = game.compat === 'verified' ? 'native' : 'translated';
                const levelLabel = game.compat === 'verified' ? '✓ Native' : '◈ Translated';
                const gCover = coverMap[String(game.appid)]?.cover || coverMap[String(game.appid)]?.banner || `https://steamcdn-a.akamaihd.net/steam/apps/${game.appid}/library_600x900.jpg`;
                searchResults.innerHTML += `
                    <div class="mood-card tilt-card searchable-card" data-id="${game.id}" style="animation-delay: ${Math.min(i, 10) * 40}ms; opacity: 0; animation: staggerFadeUp 0.6s ease-out forwards;">
                        <div class="mood-card-img">
                            <img src="${gCover}" alt="${game.name}" loading="lazy" referrerpolicy="no-referrer" class="mood-room-img" data-appid="${game.appid}">
                            <div class="mood-card-gradient"></div>
                            <div class="mood-card-hover">
                                <button class="mood-card-btn">Acquire</button>
                            </div>
                        </div>
                        <div class="mood-card-info">
                            <span class="lib-card-name">${game.name}</span>
                            <div class="lib-card-footer">
                                <span class="lib-card-meta">${game.price}</span>
                                <span class="compat-badge ${levelClass}">${levelLabel} · ${game.backend}</span>
                            </div>
                        </div>
                    </div>
                `;
            }
        });

        // Re-attach listeners for the dynamically created cards
        searchResults.querySelectorAll('.tilt-card').forEach(card => {
            card.addEventListener('mousemove', (e) => {
                const rect = card.getBoundingClientRect();
                const x = (e.clientX - rect.left) / rect.width - 0.5;
                const y = (e.clientY - rect.top) / rect.height - 0.5;
                card.style.transform = `perspective(600px) rotateY(${x * 5}deg) rotateX(${y * -3}deg) translateY(-4px)`;
            });
            card.addEventListener('mouseleave', () => {
                card.style.transform = `perspective(600px) rotateY(0deg) rotateX(0deg) translateY(0)`;
            });
            card.addEventListener('click', () => {
                window.location.hash = `/game?id=${card.dataset.id}`;
            });
        });

        // Show empty state if nothing matches
        if (emptyState) emptyState.style.display = totalMatches === 0 ? 'flex' : 'none';
        if (totalMatches === 0) searchResults.style.display = 'none';
    };

    window.addEventListener('global-search', handleGlobalSearch);

    // FREE GAMES SEARCH (Steam App List)
    const searchInput = container.querySelector('#steam-app-search-input');
    const searchBtn = container.querySelector('#steam-app-search-btn');
    const resultsContainer = container.querySelector('#steam-app-results');
    const statusEl = container.querySelector('#steam-app-status');

    // Helper to determine button action/state for store items
    function getStoreButtonState(appId) {
        const lib = store.state.library || [];
        const libGame = lib.find(g => String(g.app_id || g.appid) === String(appId));
        if (libGame) {
            if (libGame.installed || libGame.is_installed) {
                return {
                    textHtml: 'PLAY',
                    action: 'play',
                    gameId: libGame.id,
                    btnStyle: 'background: var(--brand-emerald, #10b981); color: #000;'
                };
            }
            return {
                textHtml: 'DOWNLOAD',
                action: 'download',
                gameId: libGame.id,
                btnStyle: 'background: var(--brand-cyan, #06b6d4); color: #000;'
            };
        }
        return {
            textHtml: '<span class="text-serif-italic" style="text-transform: none; font-weight: normal; font-size: 1.3em; margin-right: 4px;">Add</span><strong style="font-weight: 800; letter-spacing: 0.05em;">TO LIBRARY</strong>',
            action: 'add',
            gameId: `steam-${appId}`,
            btnStyle: ''
        };
    }

    let searchTimeout = null;

    async function performSteamSearch(query) {
        if (!query || query.length < 2) {
            resultsContainer.innerHTML = '';
            statusEl.textContent = '';
            return;
        }

        appendLog(`[Store] Steam search: "${query}"`, 'log-system');
        statusEl.textContent = 'Searching Steam catalog...';
        resultsContainer.innerHTML = '';

        try {
            const raw = await window.go.rift.App.SearchSteamAppList(query);
            statusEl.textContent = '';

            const results = JSON.parse(raw);

            // Check for backend error
            if (results && results.error) {
                appendLog(`[Store] Search error: ${results.error}`, 'log-error');
                resultsContainer.innerHTML = `<div style="grid-column: 1/-1; text-align: center; padding: 40px; color: rgba(255,200,100,0.7); font-family: 'Inter', sans-serif; font-size: 0.85rem;">⚠ ${results.error}</div>`;
                return;
            }

            if (!results || !Array.isArray(results) || results.length === 0) {
                appendLog(`[Store] Search "${query}": 0 results`, 'log-info');
                resultsContainer.innerHTML = `<div style="grid-column: 1/-1; text-align: center; padding: 40px; color: rgba(255,255,255,0.4); font-family: 'Inter', sans-serif; font-size: 0.85rem;">No games found matching "${query}".</div>`;
                return;
            }

            const showFreeOnly = container.querySelector('#steam-free-toggle')?.classList.contains('active');
            let filtered = results.filter(app => app.appid && app.appid > 0); // skip invalid entries
            if (showFreeOnly) {
                filtered = filtered.filter(app => app.price_cents === 0);
            }

            if (filtered.length === 0) {
                resultsContainer.innerHTML = `<div style="grid-column: 1/-1; text-align: center; padding: 40px; color: rgba(255,255,255,0.4); font-family: 'Inter', sans-serif; font-size: 0.85rem;">${showFreeOnly ? 'No free games found' : 'No games found'} matching "${query}".</div>`;
                return;
            }

            const showMoreWrap = container.querySelector('#steam-show-more-wrap');
            const totalCountEl = container.querySelector('#steam-total-count');

            appendLog(`[Store] "${query}": ${filtered.length} matches${showFreeOnly ? ' (free only)' : ''}`, 'log-info');

            const INITIAL_SHOW = 10;
            const showMore = filtered.length > INITIAL_SHOW;

            // Fetch cover art for search results from local DB
            const searchAppIDs = filtered.map(a => String(a.appid));
            const searchCoverMap = await getCoverMap(searchAppIDs);

            // Render only INITIAL_SHOW cards
            const visible = filtered.slice(0, INITIAL_SHOW);
            resultsContainer.innerHTML = visible.map((app, i) => {
                const appId = String(app.appid);
                const name = app.name || 'Unknown';
                const imgUrl = searchCoverMap[appId]?.banner || searchCoverMap[appId]?.cover || coverMap[appId]?.banner || coverMap[appId]?.cover || `https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/${appId}/header.jpg`;
                const isFree = app.price_cents === 0;
                const priceLabel = isFree ? 'Free' : '$' + (app.price_cents / 100).toFixed(2);
                const cardWidth = 340 + (parseInt(appId) % 80);
                const btnState = getStoreButtonState(appId);
                return `
                    <div class="mood-card tilt-card steam-result-card steam-horizontal-card" data-appid="${appId}" style="width: ${cardWidth}px; animation-delay: ${Math.min(i, 10) * 40}ms;">
                        <div class="mood-card-img">
                            <img src="${imgUrl}" data-fallback="${app.image || ''}" alt="${name}" loading="lazy" referrerpolicy="no-referrer" class="steam-result-img">
                            <div class="exhibit-placard-overlay">
                                <button class="exhibit-enter-btn steam-open-btn" style="margin: auto; ${btnState.btnStyle}" data-appid="${appId}" data-name="${name.replace(/"/g, '&quot;')}" data-action="${btnState.action}" data-gameid="${btnState.gameId || ''}">${btnState.textHtml}</button>
                            </div>
                        </div>
                        <div class="mood-card-info">
                            <span class="lib-card-name">${name}</span>
                            <div class="lib-card-footer">
                                <span class="lib-card-meta">Steam · ${priceLabel}</span>
                            </div>
                        </div>
                    </div>
                `;
            }).join('');

            // Show/hide "Show All" button
            if (showMore && showMoreWrap) {
                if (totalCountEl) totalCountEl.textContent = filtered.length;
                showMoreWrap.style.display = '';
                // Store full data + cover map on the wrapper for "Show All"
                showMoreWrap.dataset.fullResults = JSON.stringify(filtered.map(app => ({
                    appid: String(app.appid),
                    name: app.name || 'Unknown',
                    price_cents: app.price_cents,
                    image: app.image
                })));
                showMoreWrap.dataset.coverMap = JSON.stringify(searchCoverMap);
            } else if (showMoreWrap) {
                showMoreWrap.style.display = 'none';
            }
        } catch (err) {
            statusEl.textContent = `Search failed: ${err}`;
            resultsContainer.innerHTML = `<div style="grid-column: 1/-1; text-align: center; padding: 40px; color: rgba(255,100,100,0.6); font-family: 'Inter', sans-serif; font-size: 0.85rem;">Search error: ${err}</div>`;
        }

        // Fix broken images: try fallback URLs before showing placeholder
        resultsContainer.querySelectorAll('.steam-result-img').forEach(img => {
            const appId = img.closest('.steam-result-card')?.dataset.appid;
            let fallbackAttempts = 0;
            
            // Steam CDN returns a 1x1 pixel image instead of 404 for missing assets.
            // Catch this and manually trigger the error fallback chain.
            img.addEventListener('load', function() {
                if (this.naturalWidth <= 1) {
                    this.dispatchEvent(new Event('error'));
                }
            });

            img.addEventListener('error', function onImgError() {
                fallbackAttempts++;
                if (fallbackAttempts === 1) {
                    this.src = `https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/${appId}/capsule_616x353.jpg`;
                } else if (fallbackAttempts === 2 && this.dataset.fallback) {
                    this.src = this.dataset.fallback;
                } else {
                    this.parentElement.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100%;width:100%;background:#0a0a0a;color:rgba(255,255,255,0.15);font-size:2rem;">⌕</div>';
                }
            });
        });

        // Attach tilt-card mouse handlers to match other store cards
        resultsContainer.querySelectorAll('.steam-result-card').forEach(card => {
            card.addEventListener('mousemove', (e) => {
                const rect = card.getBoundingClientRect();
                const x = (e.clientX - rect.left) / rect.width - 0.5;
                const y = (e.clientY - rect.top) / rect.height - 0.5;
                card.style.transform = `perspective(600px) rotateY(${x * 5}deg) rotateX(${y * -3}deg) translateY(-4px)`;
            });
            card.addEventListener('mouseleave', () => {
                card.style.transform = `perspective(600px) rotateY(0deg) rotateX(0deg) translateY(0)`;
            });
        });

        // Attach button action handlers (ADD TO LIBRARY / DOWNLOAD / PLAY)
        resultsContainer.querySelectorAll('.steam-open-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                const appId = btn.dataset.appid;
                const name = btn.dataset.name;
                const action = btn.dataset.action || 'add';
                const gameId = btn.dataset.gameid || `steam-${appId}`;

                if (!appId || appId === '0' || appId === 'undefined') {
                    appendLog(`[Store] Cannot open "${name}" — invalid App ID.`, 'log-error');
                    return;
                }

                if (action === 'play') {
                    appendLog(`[Store] Launching "${name}"...`, 'log-system');
                    await window.go.rift.App.PlayGame(gameId, 'steam', appId);
                    return;
                }

                if (action === 'download') {
                    appendLog(`[Store] Starting download for "${name}"...`, 'log-system');
                    btn.textContent = 'Starting Download…';
                    btn.style.pointerEvents = 'none';
                    btn.style.opacity = '0.7';
                    await window.go.rift.App.AcquireGame(gameId, 'steam', appId);
                    return;
                }

                if (action === 'downloading') {
                    return;
                }

                btn.textContent = 'Adding to Library…';
                btn.style.pointerEvents = 'none';
                btn.style.opacity = '0.7';

                try {
                    const result = await window.go.rift.App.AddGameToLibrary(appId, name);
                    if (!result || !result.success) {
                        btn.textContent = '✗ Failed';
                        btn.style.opacity = '1';
                        btn.style.background = 'rgba(231, 76, 60, 0.4)';
                        btn.style.borderColor = 'rgba(231, 76, 60, 0.6)';
                        btn.style.pointerEvents = 'auto';
                        appendLog(`[Store] Failed to add "${name}" to library`, 'log-error');
                        return;
                    }

                    btn.textContent = 'DOWNLOAD';
                    btn.dataset.action = 'download';
                    btn.dataset.gameid = result.gameId || `steam-${appId}`;
                    btn.style.background = 'rgba(56, 189, 248, 0.25)';
                    btn.style.borderColor = '#38bdf8';
                    btn.style.color = '#38bdf8';
                    btn.style.pointerEvents = 'auto';
                    btn.style.opacity = '1';
                    appendLog(`[Store] Added "${name}" to RIFT Library. Click Download to start.`, 'log-success');
                } catch (err) {
                    btn.textContent = '✗ Failed';
                    btn.style.opacity = '1';
                    btn.style.background = 'rgba(231, 76, 60, 0.4)';
                    btn.style.borderColor = 'rgba(231, 76, 60, 0.6)';
                    btn.style.pointerEvents = 'auto';
                    appendLog(`[Store] Failed to add "${name}": ${err}`, 'log-error');
                }
            });
        });
    }

    if (searchBtn) {
        searchBtn.addEventListener('click', () => {
            performSteamSearch(searchInput.value.trim());
        });
    }

    if (searchInput) {
        searchInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                clearTimeout(searchTimeout);
                performSteamSearch(searchInput.value.trim());
            }
        });
        // Debounced search on input (after 400ms idle)
        searchInput.addEventListener('input', () => {
            clearTimeout(searchTimeout);
            const val = searchInput.value.trim();
            if (val.length >= 2) {
                searchTimeout = setTimeout(() => performSteamSearch(val), 400);
            } else {
                resultsContainer.innerHTML = '';
                statusEl.textContent = '';
            }
        });
    }

    // Free-only filter toggle
    const freeToggle = container.querySelector('#steam-free-toggle');
    if (freeToggle) {
        freeToggle.addEventListener('click', () => {
            freeToggle.classList.toggle('active');
            const isActive = freeToggle.classList.contains('active');
            freeToggle.textContent = isActive ? '✓ Free Only' : 'Free Only';
            freeToggle.style.background = isActive ? 'rgba(46, 204, 113, 0.3)' : 'rgba(46, 204, 113, 0.15)';
            freeToggle.style.borderColor = isActive ? 'rgba(46, 204, 113, 0.6)' : 'rgba(46, 204, 113, 0.3)';
            // Re-filter current results if there's a search active
            if (searchInput.value.trim().length >= 2) {
                performSteamSearch(searchInput.value.trim());
            }
        });
    }

    // "Show All" / "Show Less" toggle
    const showMoreWrap = container.querySelector('#steam-show-more-wrap');
    const showMoreBtn = container.querySelector('#steam-show-more-btn');
    if (showMoreBtn && showMoreWrap) {
        showMoreBtn.addEventListener('click', () => {
            const isShowingAll = showMoreBtn.dataset.showingAll === 'true';
            const fullData = JSON.parse(showMoreWrap.dataset.fullResults || '[]');
            const searchCoverMap = JSON.parse(showMoreWrap.dataset.coverMap || '{}');
            if (!isShowingAll && fullData.length > 0) {
                // Render ALL cards
                resultsContainer.innerHTML = fullData.map((app, i) => {
                    const aid = String(app.appid);
                    const imgUrl = searchCoverMap[aid]?.banner || searchCoverMap[aid]?.cover || coverMap[aid]?.banner || coverMap[aid]?.cover || `https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/${aid}/header.jpg`;
                    const isFree = app.price_cents === 0;
                    const priceLabel = isFree ? 'Free' : '$' + (app.price_cents / 100).toFixed(2);
                    const cardWidth = 340 + (parseInt(aid) % 80);
                    return `
                        <div class="mood-card tilt-card steam-result-card steam-horizontal-card" data-appid="${app.appid}" style="width: ${cardWidth}px;">
                            <div class="mood-card-img">
                                <img src="${imgUrl}" data-fallback="${app.image || ''}" alt="${app.name}" loading="lazy" referrerpolicy="no-referrer" class="steam-result-img">
                                <div class="exhibit-placard-overlay">
                                    <button class="exhibit-enter-btn steam-open-btn" style="margin: auto;" data-appid="${app.appid}" data-name="${app.name.replace(/"/g, '&quot;')}"><span class="text-serif-italic" style="text-transform: none; font-weight: normal; font-size: 1.3em; margin-right: 4px;">Add</span><strong style="font-weight: 800; letter-spacing: 0.05em;">TO LIBRARY</strong></button>
                                </div>
                            </div>
                            <div class="mood-card-info">
                                <span class="lib-card-name">${app.name}</span>
                                <div class="lib-card-footer">
                                    <span class="lib-card-meta">Steam · ${priceLabel}</span>
                                </div>
                            </div>
                        </div>
                    `;
                }).join('');
                showMoreBtn.innerHTML = 'Show Less';
                showMoreBtn.dataset.showingAll = 'true';
                // Re-attach image error handlers and tilt/button handlers
                resultsContainer.querySelectorAll('.steam-result-img').forEach(img => {
                    const appId = img.closest('.steam-result-card')?.dataset.appid;
                    let fb = 0;
                    img.addEventListener('load', function() {
                        if (this.naturalWidth <= 1) {
                            this.dispatchEvent(new Event('error'));
                        }
                    });
                    img.addEventListener('error', function() {
                        fb++;
                        if (fb === 1) this.src = `https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/${appId}/capsule_616x353.jpg`;
                        else if (fb === 2 && this.dataset.fallback) this.src = this.dataset.fallback;
                        else this.parentElement.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100%;width:100%;background:#0a0a0a;color:rgba(255,255,255,0.15);font-size:2rem;">⌕</div>';
                    });
                });
                resultsContainer.querySelectorAll('.steam-result-card').forEach(card => {
                    card.addEventListener('mousemove', (e) => {
                        const rect = card.getBoundingClientRect();
                        const x = (e.clientX - rect.left) / rect.width - 0.5;
                        const y = (e.clientY - rect.top) / rect.height - 0.5;
                        card.style.transform = `perspective(600px) rotateY(${x * 5}deg) rotateX(${y * -3}deg) translateY(-4px)`;
                    });
                    card.addEventListener('mouseleave', () => {
                        card.style.transform = `perspective(600px) rotateY(0deg) rotateX(0deg) translateY(0)`;
                    });
                });
                // No "In Library" pre-marking — every button just opens Steam
                resultsContainer.querySelectorAll('.steam-open-btn').forEach(btn => {
                    btn.addEventListener('click', async (e) => {
                        e.stopPropagation();
                        const appId = btn.dataset.appid;
                        const name = btn.dataset.name;

                        if (!appId || appId === '0' || appId === 'undefined') {
                            appendLog(`[Store] Cannot open "${name}" — invalid App ID.`, 'log-error');
                            return;
                        }

                        btn.textContent = 'Adding to Library…';
                        btn.style.pointerEvents = 'none';
                        btn.style.opacity = '0.7';

                        try {
                            const result = await window.go.rift.App.AddGameToLibrary(appId, name);
                            if (!result || !result.success) {
                                btn.textContent = '✗ Failed';
                                btn.style.opacity = '1';
                                btn.style.background = 'rgba(231, 76, 60, 0.4)';
                                btn.style.borderColor = 'rgba(231, 76, 60, 0.6)';
                                btn.style.pointerEvents = 'auto';
                                appendLog(`[Store] Failed to add "${name}" to library`, 'log-error');
                                return;
                            }

                            btn.textContent = '✓ Added to Library';
                            btn.style.background = 'rgba(46, 204, 113, 0.4)';
                            btn.style.borderColor = 'rgba(46, 204, 113, 0.6)';
                            btn.title = 'Added to RIFT Library. Go to Your Archive to download and play.';
                            appendLog(`[Store] Added "${name}" to RIFT Library.`, 'log-success');
                        } catch (err) {
                            btn.textContent = '✗ Failed';
                            btn.style.opacity = '1';
                            btn.style.background = 'rgba(231, 76, 60, 0.4)';
                            btn.style.borderColor = 'rgba(231, 76, 60, 0.6)';
                            btn.style.pointerEvents = 'auto';
                            appendLog(`[Store] Failed to open "${name}": ${err}`, 'log-error');
                        }
                    });
                });
            } else if (isShowingAll) {
                // Collapse back to INITIAL_SHOW
                const filtered = fullData;
                const INITIAL_SHOW = 10;
                const visible = filtered.slice(0, INITIAL_SHOW);
                resultsContainer.innerHTML = visible.map((app, i) => {
                    const aid = String(app.appid);
                    const imgUrl = searchCoverMap[aid]?.banner || searchCoverMap[aid]?.cover || coverMap[aid]?.banner || coverMap[aid]?.cover || `https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/${aid}/header.jpg`;
                    const isFree = app.price_cents === 0;
                    const priceLabel = isFree ? 'Free' : '$' + (app.price_cents / 100).toFixed(2);
                    const cardWidth = 340 + (parseInt(aid) % 80);
                    return `
                        <div class="mood-card tilt-card steam-result-card steam-horizontal-card" data-appid="${app.appid}" style="width: ${cardWidth}px;">
                            <div class="mood-card-img">
                                <img src="${imgUrl}" data-fallback="${app.image || ''}" alt="${app.name}" loading="lazy" referrerpolicy="no-referrer" class="steam-result-img">
                                <div class="exhibit-placard-overlay">
                                    <button class="exhibit-enter-btn steam-open-btn" style="margin: auto;" data-appid="${app.appid}" data-name="${app.name.replace(/"/g, '&quot;')}"><span class="text-serif-italic" style="text-transform: none; font-weight: normal; font-size: 1.3em; margin-right: 4px;">Add</span><strong style="font-weight: 800; letter-spacing: 0.05em;">TO LIBRARY</strong></button>
                                </div>
                            </div>
                            <div class="mood-card-info">
                                <span class="lib-card-name">${app.name}</span>
                                <div class="lib-card-footer">
                                    <span class="lib-card-meta">Steam · ${priceLabel}</span>
                                </div>
                            </div>
                        </div>
                    `;
                }).join('');
                showMoreBtn.innerHTML = 'Show All (<span id="steam-total-count">' + fullData.length + '</span>)';
                showMoreBtn.dataset.showingAll = 'false';
                // Re-attach handlers for collapsed view
                resultsContainer.querySelectorAll('.steam-result-img').forEach(img => {
                    const appId = img.closest('.steam-result-card')?.dataset.appid;
                    let fb = 0;
                    img.addEventListener('load', function() {
                        if (this.naturalWidth <= 1) {
                            this.dispatchEvent(new Event('error'));
                        }
                    });
                    img.addEventListener('error', function() {
                        fb++;
                        if (fb === 1) this.src = `https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/${appId}/header.jpg`;
                        else if (fb === 2) this.src = `https://steamcdn-a.akamaihd.net/steam/apps/${appId}/header.jpg`;
                        else this.parentElement.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100%;width:100%;background:#111;color:rgba(255,255,255,0.15);font-size:2rem;">⌕</div>';
                    });
                });
                resultsContainer.querySelectorAll('.steam-result-card').forEach(card => {
                    card.addEventListener('mousemove', (e) => {
                        const rect = card.getBoundingClientRect();
                        const x = (e.clientX - rect.left) / rect.width - 0.5;
                        const y = (e.clientY - rect.top) / rect.height - 0.5;
                        card.style.transform = `perspective(600px) rotateY(${x * 5}deg) rotateX(${y * -3}deg) translateY(-4px)`;
                    });
                    card.addEventListener('mouseleave', () => {
                        card.style.transform = `perspective(600px) rotateY(0deg) rotateX(0deg) translateY(0)`;
                    });
                });
                // No "In Library" pre-marking — every button just opens Steam
                resultsContainer.querySelectorAll('.steam-open-btn').forEach(btn => {
                    btn.addEventListener('click', async (e) => {
                        e.stopPropagation();
                        const appId = btn.dataset.appid;
                        const name = btn.dataset.name;

                        if (!appId || appId === '0' || appId === 'undefined') {
                            appendLog(`[Store] Cannot open "${name}" — invalid App ID.`, 'log-error');
                            return;
                        }

                        btn.textContent = 'Adding to Library…';
                        btn.style.pointerEvents = 'none';
                        btn.style.opacity = '0.7';

                        try {
                            const result = await window.go.rift.App.AddGameToLibrary(appId, name);
                            if (!result || !result.success) {
                                btn.textContent = '✗ Failed';
                                btn.style.opacity = '1';
                                btn.style.background = 'rgba(231, 76, 60, 0.4)';
                                btn.style.borderColor = 'rgba(231, 76, 60, 0.6)';
                                btn.style.pointerEvents = 'auto';
                                appendLog(`[Store] Failed to add "${name}" to library`, 'log-error');
                                return;
                            }

                            btn.textContent = '✓ Added to Library';
                            btn.style.background = 'rgba(46, 204, 113, 0.4)';
                            btn.style.borderColor = 'rgba(46, 204, 113, 0.6)';
                            btn.title = 'Added to RIFT Library. Go to Your Archive to download and play.';
                            appendLog(`[Store] Added "${name}" to RIFT Library.`, 'log-success');
                        } catch (err) {
                            btn.textContent = '✗ Failed';
                            btn.style.opacity = '1';
                            btn.style.background = 'rgba(231, 76, 60, 0.4)';
                            btn.style.borderColor = 'rgba(231, 76, 60, 0.6)';
                            btn.style.pointerEvents = 'auto';
                            appendLog(`[Store] Failed to open "${name}": ${err}`, 'log-error');
                        }
                    });
                });
            }
        });
    }

    return container;
}
