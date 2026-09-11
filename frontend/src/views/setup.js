import { appendLog } from '../main.js';

export function renderSetup() {
    const container = document.createElement('div');
    container.style.cssText = `
        display: flex; flex-direction: column; width: 100vw; height: 100vh;
        background: #06080d; position: relative; overflow: hidden;
        font-family: "Plus Jakarta Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
        color: #f1f5f9; user-select: none;
    `;

    // ============= TOP: ARCADE AREA (80% OF SCREEN) =============
    const arcadeArea = document.createElement('div');
    arcadeArea.id = 'setup-arcade-area';
    arcadeArea.style.cssText = `
        flex: 1; display: flex; flex-direction: column; position: relative;
        overflow: hidden; min-height: 280px;
        background: radial-gradient(circle at 50% 30%, #0d121f 0%, #06080d 85%);
    `;

    arcadeArea.innerHTML = `
        <!-- Floating Top Bar inside Arcade -->
        <div style="position: absolute; top: 14px; left: 0; right: 0; z-index: 30; display: flex; justify-content: space-between; align-items: center; padding: 0 28px; pointer-events: none;">
            <!-- Left: Studio Brand & Subtext -->
            <div style="display: flex; align-items: center; gap: 10px; pointer-events: auto;">
                <div style="display: flex; align-items: center; gap: 8px;">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                        <polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5 12 2" stroke="url(#setup-rift-logo-grad)" stroke-width="2.2" stroke-linejoin="round"/>
                        <circle cx="12" cy="12" r="3.5" fill="#38bdf8"/>
                        <defs>
                            <linearGradient id="setup-rift-logo-grad" x1="2" y1="2" x2="22" y2="22" gradientUnits="userSpaceOnUse">
                                <stop stop-color="#38bdf8"/>
                                <stop offset="1" stop-color="#818cf8"/>
                            </linearGradient>
                        </defs>
                    </svg>
                    <span style="font-weight: 800; font-size: 1rem; letter-spacing: -0.4px; color: #fff;">RIFT <span style="font-size: 0.76rem; font-weight: 600; color: rgba(255,255,255,0.4); margin-left: 2px;">ARCADE</span></span>
                </div>
                <span style="height: 12px; width: 1px; background: rgba(255,255,255,0.12);"></span>
                <span style="font-size: 0.7rem; font-family: var(--font-mono); color: rgba(255,255,255,0.45); letter-spacing: 0.02em;">PLAY WHILE RUNTIMES INITIALIZE</span>
            </div>

            <!-- Center: Polished Segmented Game Switcher -->
            <div style="display: flex; background: rgba(15, 20, 32, 0.8); backdrop-filter: blur(16px); -webkit-backdrop-filter: blur(16px); border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 20px; padding: 3px; gap: 4px; pointer-events: auto; box-shadow: 0 4px 20px rgba(0,0,0,0.45);">
                <button id="btn-arcade-golf" style="background: rgba(255,255,255,0.14); color: #fff; border: none; padding: 5px 14px; border-radius: 14px; font-size: 0.74rem; font-weight: 600; cursor: pointer; transition: all 0.2s; display: flex; align-items: center; gap: 6px;">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M4 22V2l14 6-14 6"/></svg>
                    Mini Golf
                </button>
                <button id="btn-arcade-snake" style="background: transparent; color: rgba(255,255,255,0.45); border: none; padding: 5px 14px; border-radius: 14px; font-size: 0.74rem; font-weight: 600; cursor: pointer; transition: all 0.2s; display: flex; align-items: center; gap: 6px;">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
                    Cyber Snake
                </button>
            </div>

            <!-- Right: Real-time Runtime Readiness Status -->
            <div style="display: flex; align-items: center; gap: 10px; pointer-events: auto;">
                <div id="top-status-indicator" style="display: flex; align-items: center; gap: 6px; font-size: 0.72rem; font-family: var(--font-mono); color: #94a3b8; background: rgba(15,20,32,0.75); backdrop-filter: blur(12px); padding: 4px 12px; border-radius: 14px; border: 1px solid rgba(255,255,255,0.08);">
                    <span id="top-status-dot" style="width: 6px; height: 6px; border-radius: 50%; background: #fbbf24;"></span>
                    <span id="top-status-text">Verifying Engines...</span>
                </div>
            </div>
        </div>

        <!-- Dynamic Game Stage Container (Takes 100% of the arcade area) -->
        <div id="arcade-game-stage" style="flex: 1; width: 100%; height: 100%; position: relative; display: flex; align-items: center; justify-content: center;"></div>
    `;
    container.appendChild(arcadeArea);

    // ============= BOTTOM: DOCKED TRANSLATION ENGINES PANEL (~20% OF SCREEN) =============
    const panel = document.createElement('div');
    panel.style.cssText = `
        background: rgba(8, 11, 18, 0.95); backdrop-filter: blur(24px); -webkit-backdrop-filter: blur(24px);
        border-top: 1px solid rgba(255, 255, 255, 0.08); padding: 12px 24px 14px;
        display: flex; flex-direction: column; gap: 8px; z-index: 40;
        box-shadow: 0 -8px 30px rgba(0,0,0,0.5);
    `;

    // Header row
    const headerRow = document.createElement('div');
    headerRow.style.cssText = 'display: flex; justify-content: space-between; align-items: center;';
    headerRow.innerHTML = `
        <div style="display: flex; align-items: center; gap: 10px; cursor: pointer;" id="panel-toggle">
            <div style="width: 22px; height: 22px; border-radius: 5px; background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.12); display: flex; align-items: center; justify-content: center;">
                <svg id="panel-chevron" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.7)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="transition: transform 0.2s; transform: rotate(0deg);"><polyline points="6 9 12 15 18 9"></polyline></svg>
            </div>
            <div>
                <div style="color: #fff; font-size: 0.85rem; font-weight: 700; letter-spacing: -0.2px; margin: 0 0 1px;">
                    System Compatibility & Translation Engines
                </div>
                <div style="color: rgba(255,255,255,0.4); font-family: var(--font-mono); font-size: 0.68rem;">
                    Apple Silicon (ARM64) · Isolated Sandboxed Volumes in ~/.rift/engines/
                </div>
            </div>
        </div>
        <div style="display: flex; gap: 8px; align-items: center;">
            <button id="setup-skip-btn" style="background: transparent; color: rgba(255,255,255,0.45); border: 1px solid rgba(255,255,255,0.12); padding: 5px 12px; border-radius: 6px; font-size: 0.74rem; font-weight: 500; cursor: pointer; transition: all 0.2s;">
                Skip
            </button>
            <button id="setup-install-all-btn" style="background: linear-gradient(135deg, #38bdf8, #2563eb); color: white; border: none; padding: 6px 18px; border-radius: 6px; font-size: 0.78rem; font-weight: 700; cursor: pointer; transition: all 0.2s; box-shadow: 0 4px 14px rgba(37,99,235,0.3);">
                Install All
            </button>
        </div>
    `;
    panel.appendChild(headerRow);

    // ============= 5 Translation Engine Cards Row =============
    const cardsContainer = document.createElement('div');
    cardsContainer.id = 'setup-cards-container';
    cardsContainer.style.cssText = 'display: flex; gap: 8px; width: 100%;';

    const engines = [
        { 
            id: 'rosetta', 
            label: 'Rosetta 2', 
            sub: 'Apple Silicon JIT',
            desc: 'x86_64 binary instruction translation', 
            iconSvg: `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#34d399" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><path d="M12 3v3m0 12v3M3 12h3m12 0h3m-4.5-4.5-2 2m-5 5-2 2m0-9 2 2m5 5 2 2"/></svg>` 
        },
        { 
            id: 'wine', 
            label: 'Wine Staging 11', 
            sub: 'Windows WoW64',
            desc: '64-on-64 Windows NT runtime capsule', 
            iconSvg: `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#f43f5e" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 22h8m-4-7v7M7 2h10l1 7a5 5 0 0 1-5 5 5 5 0 0 1-5-5L7 2z"/></svg>` 
        },
        { 
            id: 'dxvk', 
            label: 'DXVK-macOS', 
            sub: 'DirectX 9/11 → Metal',
            desc: 'Direct3D draw call translation over MoltenVK', 
            iconSvg: `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21 16-9 5-9-5V8l9-5 9 5v8z"/><path d="m3.27 6.96 8.73 4.97 8.73-4.97M12 22.08V12"/></svg>` 
        },
        { 
            id: 'gptk', 
            label: 'Game Porting Toolkit', 
            sub: 'DirectX 12 → Metal 3',
            desc: 'Apple D3DMetal translation framework', 
            iconSvg: `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#a78bfa" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="4" width="16" height="16" rx="3"/><path d="M9 9h6v6H9zM9 1v3m6-3v3M9 20v3m6-3v3M1 9h3m-3 6h3M20 9h3m-3 6h3"/></svg>` 
        },
        { 
            id: 'steamcmd', 
            label: 'SteamCMD', 
            sub: 'Valve Depot Engine',
            desc: 'Authenticated depot & package sync runtime', 
            iconSvg: `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#60a5fa" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>` 
        },
    ];

    const cardEls = {};
    for (const eng of engines) {
        const card = document.createElement('div');
        card.id = `engine-card-${eng.id}`;
        card.style.cssText = `
            flex: 1; min-width: 0; background: rgba(255,255,255,0.025); border: 1px solid rgba(255,255,255,0.07);
            border-radius: 8px; padding: 8px 11px; display: flex; flex-direction: column; justify-content: space-between; gap: 4px;
            transition: border-color 0.25s, background 0.25s;
        `;
        card.innerHTML = `
            <div style="display: flex; justify-content: space-between; align-items: center;">
                <div style="display: flex; align-items: center; gap: 6px;">
                    <div style="display: flex; align-items: center; justify-content: center; width: 20px; height: 20px; border-radius: 5px; background: rgba(255,255,255,0.04);">
                        ${eng.iconSvg}
                    </div>
                    <span style="color: white; font-weight: 700; font-size: 0.76rem; letter-spacing: -0.1px;">${eng.label}</span>
                </div>
                <span id="status-${eng.id}" style="font-size: 0.66rem; font-family: var(--font-mono); font-weight: 600; padding: 2px 6px; border-radius: 4px; background: rgba(255,255,255,0.05); color: rgba(255,255,255,0.4); border: 1px solid rgba(255,255,255,0.06);">
                    Checking...
                </span>
            </div>
            <div style="color: rgba(255,255,255,0.45); font-size: 0.67rem; line-height: 1.25; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
                ${eng.desc}
            </div>
            <div id="progress-wrap-${eng.id}" style="display: none; margin-top: 2px;">
                <div style="display: flex; justify-content: space-between; font-size: 0.62rem; font-family: var(--font-mono); margin-bottom: 2px;">
                    <span id="stage-text-${eng.id}" style="color: #38bdf8; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 80%;">Downloading...</span>
                    <span id="stage-percent-${eng.id}" style="color: #fff; font-weight: 700;">0%</span>
                </div>
                <div style="height: 3px; background: rgba(255,255,255,0.08); border-radius: 2px; overflow: hidden;">
                    <div id="progress-fill-${eng.id}" style="height: 100%; width: 0%; background: linear-gradient(90deg, #38bdf8, #818cf8); transition: width 0.3s ease;"></div>
                </div>
            </div>
            <button id="btn-${eng.id}" style="display: none; margin-top: 3px; background: rgba(56,189,248,0.1); border: 1px solid rgba(56,189,248,0.25); color: #38bdf8; padding: 4px 0; border-radius: 4px; font-size: 0.7rem; font-weight: 600; cursor: pointer; width: 100%; transition: all 0.2s;">
                Auto-Install
            </button>
        `;
        cardsContainer.appendChild(card);
        cardEls[eng.id] = card;
    }
    panel.appendChild(cardsContainer);
    container.appendChild(panel);

    // Toggle bottom panel collapse/expand
    setTimeout(() => {
        const toggleBtn = container.querySelector('#panel-toggle');
        const chevron = container.querySelector('#panel-chevron');
        let isExpanded = true;
        if (toggleBtn && chevron) {
            toggleBtn.addEventListener('click', () => {
                isExpanded = !isExpanded;
                cardsContainer.style.display = isExpanded ? 'flex' : 'none';
                chevron.style.transform = isExpanded ? 'rotate(0deg)' : 'rotate(-90deg)';
            });
        }
    }, 0);

    // ============= ARCADE SWITCHER LOGIC =============
    let currentArcade = 'golf'; // 'golf' or 'snake'

    function mountGolf() {
        if (window.snakeCleanup) window.snakeCleanup();
        const stage = container.querySelector('#arcade-game-stage');
        if (!stage) return;
        stage.innerHTML = `
            <canvas id="golfCanvas" style="background: transparent; cursor: crosshair; width: 100%; height: 100%; display: block;"></canvas>
            <div id="game-overlay" style="position: absolute; text-align: center; pointer-events: none; top: 11%; left: 0; right: 0;">
                <h1 style="color: white; font-size: 1.5rem; font-weight: 800; margin: 0; letter-spacing: -0.4px;">
                    RIFT <span style="color: rgba(255,255,255,0.35); font-weight: 600;">GOLF</span>
                </h1>
                <p style="color: rgba(255,255,255,0.45); margin: 3px 0 0; font-size: 0.74rem;">
                    Drag & release to sink the putt while runtimes initialize.
                </p>
                <div style="margin-top: 6px; font-family: var(--font-mono); color: #38bdf8; font-size: 0.78rem; font-weight: 700; letter-spacing: 0.04em;">
                    HOLE <span id="level-display">1</span>
                </div>
            </div>
        `;
        setTimeout(() => initGolf(container), 0);
    }

    function mountSnake() {
        if (window.golfCleanup) window.golfCleanup();
        const stage = container.querySelector('#arcade-game-stage');
        if (!stage) return;
        stage.innerHTML = `
            <div style="width: 100%; height: 100%; position: relative; display: flex; flex-direction: column; overflow: hidden;">
                <!-- Snake HUD -->
                <div style="position: absolute; top: 12px; right: 28px; z-index: 25; display: flex; align-items: center; gap: 18px;">
                    <div id="active-power-hud" style="display: none; padding: 4px 12px; border-radius: 14px; font-size: 0.7rem; font-weight: 700; letter-spacing: 0.5px;"></div>
                    <div style="display: flex; flex-direction: column; align-items: flex-end;">
                        <span style="font-size: 0.62rem; color: rgba(255,255,255,0.4); text-transform: uppercase; font-weight: 700;">Score</span>
                        <span id="snake-score" style="font-size: 1.1rem; font-weight: 800; color: #fff; line-height: 1;">0</span>
                    </div>
                    <div style="display: flex; flex-direction: column; align-items: flex-end;">
                        <span style="font-size: 0.62rem; color: rgba(255,255,255,0.4); text-transform: uppercase; font-weight: 700;">Multiplier</span>
                        <span id="snake-combo" style="font-size: 1.1rem; font-weight: 800; color: #38bdf8; line-height: 1;">1x</span>
                    </div>
                    <div style="display: flex; flex-direction: column; align-items: flex-end;">
                        <span style="font-size: 0.62rem; color: rgba(255,255,255,0.4); text-transform: uppercase; font-weight: 700;">Best</span>
                        <span id="snake-high" style="font-size: 1.1rem; font-weight: 800; color: #34d399; line-height: 1;">0</span>
                    </div>
                    <button id="btn-sound-toggle" style="background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.12); color: #fff; width: 28px; height: 28px; border-radius: 6px; cursor: pointer; display: flex; align-items: center; justify-content: center;" title="Sound">
                        <svg id="sound-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"></polygon><path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"></path></svg>
                    </button>
                </div>

                <canvas id="snakeCanvas" style="flex: 1; width: 100%; height: 100%; display: block;"></canvas>

                <!-- Legend -->
                <div style="position: absolute; bottom: 8px; left: 50%; transform: translateX(-50%); z-index: 15; display: flex; align-items: center; gap: 14px; background: rgba(15,18,26,0.85); backdrop-filter: blur(8px); border: 1px solid rgba(255,255,255,0.08); padding: 4px 14px; border-radius: 20px; font-size: 0.68rem; color: rgba(255,255,255,0.7);">
                    <div style="display: flex; align-items: center; gap: 4px;"><span style="width: 7px; height: 7px; border-radius: 50%; background: #00ffcc; box-shadow: 0 0 6px #00ffcc;"></span> Food (+50)</div>
                    <span style="color: rgba(255,255,255,0.15);">|</span>
                    <div style="display: flex; align-items: center; gap: 4px;"><span style="width: 7px; height: 7px; border-radius: 50%; background: #ffd700; box-shadow: 0 0 6px #ffd700;"></span> ⚡ Overdrive</div>
                    <span style="color: rgba(255,255,255,0.15);">|</span>
                    <div style="display: flex; align-items: center; gap: 4px;"><span style="width: 7px; height: 7px; border-radius: 50%; background: #38bdf8; box-shadow: 0 0 6px #38bdf8;"></span> ⏳ Slow-Mo</div>
                    <span style="color: rgba(255,255,255,0.15);">|</span>
                    <div style="display: flex; align-items: center; gap: 4px;"><span style="width: 7px; height: 7px; border-radius: 50%; background: #f43f5e; box-shadow: 0 0 6px #f43f5e;"></span> 💎 Gem (+500)</div>
                    <span style="color: rgba(255,255,255,0.15);">|</span>
                    <span style="color: rgba(255,255,255,0.4);">WASD / Arrows to steer</span>
                </div>

                <!-- Snake Start/Pause Overlay -->
                <div id="snake-overlay" style="position: absolute; inset: 0; background: rgba(7,9,14,0.85); backdrop-filter: blur(10px); display: flex; flex-direction: column; align-items: center; justify-content: center; z-index: 30;">
                    <div style="text-align: center; max-width: 380px; width: 85%; padding: 22px; border-radius: 12px; background: #0f121a; border: 1px solid rgba(255,255,255,0.1); box-shadow: 0 15px 40px rgba(0,0,0,0.6);">
                        <h2 id="overlay-title" style="margin: 0 0 6px 0; font-size: 1.4rem; font-weight: 800; color: #fff; letter-spacing: -0.4px;">CYBER SNAKE</h2>
                        <p id="overlay-desc" style="margin: 0 0 16px 0; font-size: 0.78rem; color: rgba(255,255,255,0.6); line-height: 1.4;">
                            Collect data cores, chain combos, and navigate the grid while runtimes configure.
                        </p>
                        <button id="btn-start-snake" style="background: #38bdf8; color: #080808; border: none; padding: 9px 22px; border-radius: 6px; font-size: 0.82rem; font-weight: 700; letter-spacing: 0.02em; cursor: pointer; transition: transform 0.15s, background 0.15s;">
                            START GAME
                        </button>
                    </div>
                </div>
            </div>
        `;
        setTimeout(() => initSetupSnake(container), 0);
    }

    setTimeout(() => {
        const btnGolf = container.querySelector('#btn-arcade-golf');
        const btnSnake = container.querySelector('#btn-arcade-snake');

        if (btnGolf) {
            btnGolf.addEventListener('click', () => {
                if (currentArcade === 'golf') return;
                currentArcade = 'golf';
                btnGolf.style.background = 'rgba(255,255,255,0.14)';
                btnGolf.style.color = '#fff';
                if (btnSnake) {
                    btnSnake.style.background = 'transparent';
                    btnSnake.style.color = 'rgba(255,255,255,0.45)';
                }
                mountGolf();
            });
        }

        if (btnSnake) {
            btnSnake.addEventListener('click', () => {
                if (currentArcade === 'snake') return;
                currentArcade = 'snake';
                btnSnake.style.background = 'rgba(255,255,255,0.14)';
                btnSnake.style.color = '#fff';
                if (btnGolf) {
                    btnGolf.style.background = 'transparent';
                    btnGolf.style.color = 'rgba(255,255,255,0.45)';
                }
                mountSnake();
            });
        }

        // Initial default: Mini Golf taking the 80% screen
        mountGolf();
    }, 0);

    // ============= ENGINE VERIFICATION & INSTALL LOGIC =============
    setTimeout(async () => {
        let engineStatus = { wine: false, dxvk: false, gptk: false, rosetta: true, steamcmd: false };
        const engineResolvers = {};

        try {
            const raw = await window.go.rift.App.CheckEnginesStatus();
            engineStatus = JSON.parse(raw);
        } catch (e) {
            appendLog(`[Setup] Failed to inspect engines: ${e}`, 'log-error');
        }

        function updateOverallReadiness() {
            const readyCount = engines.filter(e => Boolean(engineStatus[e.id])).length;
            const totalCount = engines.length;
            const topText = document.getElementById('top-status-text');
            const topDot = document.getElementById('top-status-dot');

            if (topText && topDot) {
                if (readyCount === totalCount) {
                    topText.textContent = '5 of 5 Engines Ready';
                    topDot.style.background = '#34d399';
                } else {
                    topText.textContent = `${readyCount}/${totalCount} Engines Active`;
                    topDot.style.background = '#fbbf24';
                }
            }

            checkAllEnginesReady();
        }

        function updateCardUI(engId, isReady) {
            const statusEl = document.getElementById(`status-${engId}`);
            const btn = document.getElementById(`btn-${engId}`);
            const card = document.getElementById(`engine-card-${engId}`);
            const progWrap = document.getElementById(`progress-wrap-${engId}`);

            if (isReady) {
                if (statusEl) {
                    statusEl.innerHTML = `
                        <span style="display: inline-flex; align-items: center; gap: 4px; color: #34d399;">
                            <span style="width: 5px; height: 5px; border-radius: 50%; background: #34d399;"></span>
                            Ready
                        </span>
                    `;
                    statusEl.style.background = 'rgba(52,211,153,0.1)';
                    statusEl.style.borderColor = 'rgba(52,211,153,0.25)';
                }
                if (card) {
                    card.style.borderColor = 'rgba(52,211,153,0.22)';
                    card.style.background = 'rgba(255,255,255,0.035)';
                }
                if (btn) btn.style.display = 'none';
                if (progWrap) progWrap.style.display = 'none';
            } else {
                if (statusEl) {
                    statusEl.innerHTML = `
                        <span style="display: inline-flex; align-items: center; gap: 4px; color: #f87171;">
                            <span style="width: 5px; height: 5px; border-radius: 50%; background: #ef4444;"></span>
                            Missing
                        </span>
                    `;
                    statusEl.style.background = 'rgba(239,68,68,0.1)';
                    statusEl.style.borderColor = 'rgba(239,68,68,0.25)';
                }
                if (card) {
                    card.style.borderColor = 'rgba(255,255,255,0.08)';
                }
                if (btn) btn.style.display = 'block';
            }
        }

        for (const eng of engines) {
            updateCardUI(eng.id, Boolean(engineStatus[eng.id]));
        }
        updateOverallReadiness();

        function checkAllEnginesReady() {
            const allReady = engines.every(eng => Boolean(engineStatus[eng.id]));
            const primaryBtn = document.getElementById('setup-install-all-btn');

            if (allReady) {
                if (primaryBtn) {
                    primaryBtn.textContent = '✓ All Systems Ready — Enter RIFT →';
                    primaryBtn.style.background = 'linear-gradient(135deg, #34d399, #059669)';
                    primaryBtn.style.color = '#000';
                    primaryBtn.style.fontWeight = '800';
                    primaryBtn.style.boxShadow = '0 4px 18px rgba(52,211,153,0.4)';
                    primaryBtn.disabled = false;
                    primaryBtn.style.opacity = '1';
                    primaryBtn.onclick = () => {
                        if (window.golfCleanup) window.golfCleanup();
                        if (window.snakeCleanup) window.snakeCleanup();
                        window.hasPassedSetup = true;
                        window.localStorage.setItem('hasPassedSetup', 'true');
                        window.localStorage.setItem('hasLoggedIn', 'true');
                        window.location.hash = '/library';
                    };
                }
                return true;
            } else {
                if (primaryBtn) {
                    primaryBtn.textContent = 'Install All';
                    primaryBtn.style.background = 'linear-gradient(135deg, #38bdf8, #2563eb)';
                    primaryBtn.style.color = '#fff';
                    primaryBtn.style.fontWeight = '700';
                    primaryBtn.style.boxShadow = '0 4px 14px rgba(37,99,235,0.3)';
                    primaryBtn.disabled = false;
                    primaryBtn.onclick = installAllEngines;
                }
                return false;
            }
        }

        // Listen for engine install progress from backend
        if (window.runtime && window.runtime.EventsOn) {
            window.runtime.EventsOn('engine_install_progress', (data) => {
                const comp = data.component;
                const statusEl = document.getElementById(`status-${comp}`);
                const progWrap = document.getElementById(`progress-wrap-${comp}`);
                const progFill = document.getElementById(`progress-fill-${comp}`);
                const stageText = document.getElementById(`stage-text-${comp}`);
                const stagePercent = document.getElementById(`stage-percent-${comp}`);
                const btn = document.getElementById(`btn-${comp}`);

                if (data.status === 'downloading') {
                    if (progWrap) progWrap.style.display = 'block';
                    if (progFill) progFill.style.width = `${Math.max(5, data.percent || 0)}%`;
                    if (stagePercent) stagePercent.textContent = `${data.percent || 0}%`;
                    if (stageText && data.stage) stageText.textContent = data.stage;
                    if (btn) btn.style.display = 'none';

                    if (statusEl) {
                        statusEl.innerHTML = `
                            <span style="display: inline-flex; align-items: center; gap: 4px; color: #fbbf24;">
                                <span style="width: 5px; height: 5px; border-radius: 50%; background: #fbbf24;"></span>
                                ${data.percent || 0}%
                            </span>
                        `;
                        statusEl.style.background = 'rgba(251,191,36,0.1)';
                        statusEl.style.borderColor = 'rgba(251,191,36,0.3)';
                    }
                } else if (data.status === 'done') {
                    if (progWrap) progWrap.style.display = 'none';
                    engineStatus[comp] = true;
                    updateCardUI(comp, true);
                    updateOverallReadiness();
                    appendLog(`[Setup] ✓ ${comp} verified and ready.`, 'log-success');

                    if (engineResolvers[comp]) {
                        engineResolvers[comp].resolve(true);
                        delete engineResolvers[comp];
                    }
                } else if (data.status === 'failed') {
                    if (progWrap) progWrap.style.display = 'none';
                    if (statusEl) {
                        statusEl.innerHTML = `
                            <span style="display: inline-flex; align-items: center; gap: 4px; color: #ef4444;">
                                <span style="width: 5px; height: 5px; border-radius: 50%; background: #ef4444;"></span>
                                Failed
                            </span>
                        `;
                        statusEl.style.background = 'rgba(239,68,68,0.1)';
                        statusEl.style.borderColor = 'rgba(239,68,68,0.3)';
                    }
                    if (btn) {
                        btn.style.display = 'block';
                        btn.disabled = false;
                        btn.textContent = 'Retry';
                        btn.style.opacity = '1';
                    }
                    appendLog(`[Setup] ✗ ${comp} installation failed: ${data.error}`, 'log-error');

                    if (engineResolvers[comp]) {
                        engineResolvers[comp].reject(new Error(data.error));
                        delete engineResolvers[comp];
                    }
                }
            });
        }

        async function installEngineComponent(component) {
            const btn = document.getElementById(`btn-${component}`);
            if (btn) {
                btn.disabled = true;
                btn.textContent = 'Installing...';
                btn.style.opacity = '0.6';
            }

            const statusEl = document.getElementById(`status-${component}`);
            if (statusEl) {
                statusEl.innerHTML = `<span style="color: #fbbf24;">Starting...</span>`;
            }

            const progWrap = document.getElementById(`progress-wrap-${component}`);
            if (progWrap) progWrap.style.display = 'block';

            return new Promise(async (resolve, reject) => {
                engineResolvers[component] = { resolve, reject };
                try {
                    const msg = await window.go.rift.App.InstallEngine(component);
                    appendLog(`[Setup] ${msg}`, 'log-info');
                } catch (err) {
                    appendLog(`[Setup] ${component} install error: ${err}`, 'log-error');
                    if (btn) {
                        btn.disabled = false;
                        btn.textContent = 'Retry';
                        btn.style.opacity = '1';
                    }
                    if (statusEl) {
                        statusEl.innerHTML = `<span style="color: #ef4444;">Error</span>`;
                    }
                    if (progWrap) progWrap.style.display = 'none';
                    delete engineResolvers[component];
                    reject(err);
                }
            });
        }

        // Wire individual card buttons
        for (const eng of engines) {
            const btn = document.getElementById(`btn-${eng.id}`);
            if (btn) {
                btn.addEventListener('click', () => installEngineComponent(eng.id));
            }
        }

        // Install All handler
        async function installAllEngines() {
            if (checkAllEnginesReady()) {
                const primaryBtn = document.getElementById('setup-install-all-btn');
                if (primaryBtn) primaryBtn.onclick();
                return;
            }

            const primaryBtn = document.getElementById('setup-install-all-btn');
            if (primaryBtn) {
                primaryBtn.disabled = true;
                primaryBtn.textContent = 'Installing Runtimes...';
                primaryBtn.style.opacity = '0.7';
            }

            for (const eng of engines) {
                if (!engineStatus[eng.id]) {
                    try {
                        await installEngineComponent(eng.id);
                    } catch (e) {
                        appendLog(`[Setup] ${eng.label} encountered an issue: ${e}`, 'log-warn');
                    }
                }
            }

            updateOverallReadiness();
        }

        // Skip button handler
        const skipBtn = document.getElementById('setup-skip-btn');
        if (skipBtn) {
            skipBtn.addEventListener('click', () => {
                const proceed = window.confirm("WARNING: If you skip engine installation, Windows games will NOT launch. Are you sure you want to skip?");
                if (!proceed) return;
                
                if (window.golfCleanup) window.golfCleanup();
                if (window.snakeCleanup) window.snakeCleanup();
                window.hasPassedSetup = true;
                window.localStorage.setItem('hasPassedSetup', 'true');
                window.localStorage.setItem('hasLoggedIn', 'true');
                window.location.hash = '/library';
            });
        }

        // Initial check: if all 5 are ready, smooth prompt
        if (checkAllEnginesReady()) {
            appendLog('[Setup] All translation engines verified on disk.', 'log-system');
        }
    }, 100);

    return container;
}

// ============= GOLF GAME ENGINE =============
function initGolf(container) {
    const canvas = container.querySelector('#golfCanvas');
    if (!canvas) return;
    canvas.style.setProperty('--wails-draggable', 'no-drag');
    const ctx = canvas.getContext('2d');
    const levelDisplay = container.querySelector('#level-display');

    let golfAnimFrame = null;

    function resize() {
        if (!canvas.parentElement) return;
        canvas.width = canvas.parentElement.clientWidth;
        canvas.height = canvas.parentElement.clientHeight || 300;
        updateCamera();
    }
    window.addEventListener('resize', resize);
    
    let scale = 1, ox = 0, oy = 0;
    const TILE_SIZE = 40, GRAVITY = -0.4, FRICTION = 0.96, RESTITUTION = 0.55;
    let level = 1, isDragging = false, dragStart = { x: 0, y: 0 }, dragCurrent = { x: 0, y: 0 }, strokes = 0;
    let grid = [], hole = { q: 0, r: 0 }, startPos = { q: 0, r: 0 };
    const ball = { x: 0, y: 0, z: 200, vx: 0, vy: 0, vz: 0, radius: 4, state: 'falling' };

    function updateCamera() {
        if (!grid || grid.length === 0) return;
        const size = grid.length;
        const gridW = 2 * size * TILE_SIZE * Math.cos(Math.PI / 6);
        const gridH = 2 * size * TILE_SIZE * Math.sin(Math.PI / 6);
        
        const safeTop = Math.min(canvas.height * 0.3, 140);
        const safeH = canvas.height - safeTop - 30; 
        const safeW = canvas.width - 40;
        
        scale = Math.min(1, safeW / gridW, safeH / gridH);
        ox = canvas.width / 2;
        oy = safeTop + safeH / 2 - ((size * TILE_SIZE) * Math.sin(Math.PI / 6) * scale);
    }

    function generateLevel() {
        strokes = 0;
        const size = Math.min(6 + level * 2, 14);
        grid = [];
        for (let i = 0; i < size; i++) {
            grid[i] = [];
            for (let j = 0; j < size; j++) {
                if (Math.random() > 0.2 || (i===1&&j===1) || (i===size-2&&j===size-2)) {
                    grid[i][j] = Math.random() > 0.8 && !(i===1&&j===1) && !(i===size-2&&j===size-2) ? 2 : 1;
                } else {
                    grid[i][j] = 0;
                }
            }
        }
        startPos = { q: 1, r: 1 };
        hole = { q: size - 2, r: size - 2 };
        for (let i = 0; i < size; i++) {
            for (let j = 0; j < size; j++) {
                if (grid[i][j] === 2) {
                    if (Math.abs(i - startPos.q) <= 1 && Math.abs(j - startPos.r) <= 1) grid[i][j] = 1;
                    if (Math.abs(i - hole.q) <= 2 && Math.abs(j - hole.r) <= 2) grid[i][j] = 1;
                }
            }
        }
        grid[startPos.q][startPos.r] = 1;
        grid[hole.q][hole.r] = 1;
        ball.x = startPos.q * TILE_SIZE + TILE_SIZE/2;
        ball.y = startPos.r * TILE_SIZE + TILE_SIZE/2;
        ball.z = 200; ball.vx = 0; ball.vy = 0; ball.vz = 0;
        if (levelDisplay) levelDisplay.textContent = level;
        updateCamera();
    }

    function project(x, y, z) {
        return { 
            x: ((x - y) * Math.cos(Math.PI / 6)) * scale + ox, 
            y: ((x + y) * Math.sin(Math.PI / 6) - z) * scale + oy 
        };
    }

    function unproject(sx, sy) {
        sx = (sx - ox) / scale;
        sy = (sy - oy) / scale;
        const cos = Math.cos(Math.PI / 6), sin = Math.sin(Math.PI / 6);
        return { x: (sy / sin + sx / cos) / 2, y: (sy / sin - sx / cos) / 2 };
    }

    function getElevation(x, y) {
        const q = Math.floor(x / TILE_SIZE), r = Math.floor(y / TILE_SIZE);
        if (q < 0 || r < 0 || q >= grid.length || r >= grid[0].length) return -1000;
        const t = grid[q][r];
        if (t === 0) return -1000;
        return t === 2 ? TILE_SIZE : 0;
    }

    function update() {
        ball.vz += GRAVITY;
        ball.x += ball.vx; ball.y += ball.vy; ball.z += ball.vz;
        const el = getElevation(ball.x, ball.y);
        if (ball.z <= el + ball.radius) {
            ball.z = el + ball.radius;
            if (ball.vz < -1) { ball.vz = -ball.vz * RESTITUTION; }
            else { ball.vz = 0; ball.state = 'rolling'; }
            ball.vx *= FRICTION; ball.vy *= FRICTION;
        } else { ball.state = 'falling'; }

        // wall collision
        const nx = ball.x + ball.vx, ny = ball.y + ball.vy;
        if (getElevation(nx + Math.sign(ball.vx) * ball.radius, ball.y) > ball.z) { ball.vx *= -RESTITUTION; }
        if (getElevation(ball.x, ny + Math.sign(ball.vy) * ball.radius) > ball.z) { ball.vy *= -RESTITUTION; }

        const hx = hole.q * TILE_SIZE + TILE_SIZE/2, hy = hole.r * TILE_SIZE + TILE_SIZE/2;
        if (ball.z <= ball.radius + 2 && Math.sqrt((ball.x-hx)**2 + (ball.y-hy)**2) < 10 && Math.sqrt(ball.vx**2+ball.vy**2) < 3) {
            level++; generateLevel(); return;
        }
        if (ball.z < -200) { ball.x = startPos.q * TILE_SIZE + TILE_SIZE/2; ball.y = startPos.r * TILE_SIZE + TILE_SIZE/2; ball.z = 200; ball.vx = 0; ball.vy = 0; ball.vz = 0; }
        if (ball.state === 'rolling' && Math.sqrt(ball.vx**2+ball.vy**2) < 0.1) { ball.vx = 0; ball.vy = 0; ball.state = 'stationary'; }
    }

    function drawBlock(qx, qy, z, h, type) {
        const x = qx * TILE_SIZE, y = qy * TILE_SIZE;
        const p1 = project(x, y, z+h), p2 = project(x+TILE_SIZE, y, z+h), p3 = project(x+TILE_SIZE, y+TILE_SIZE, z+h), p4 = project(x, y+TILE_SIZE, z+h);
        const b1 = project(x, y, z), b2 = project(x+TILE_SIZE, y, z), b3 = project(x+TILE_SIZE, y+TILE_SIZE, z), b4 = project(x, y+TILE_SIZE, z);
        ctx.fillStyle = type === 2 ? '#1a1a1a' : '#222';
        ctx.beginPath(); ctx.moveTo(p4.x,p4.y); ctx.lineTo(p3.x,p3.y); ctx.lineTo(b3.x,b3.y); ctx.lineTo(b4.x,b4.y); ctx.fill();
        ctx.strokeStyle = 'rgba(255,255,255,0.15)'; ctx.stroke();
        ctx.fillStyle = type === 2 ? '#222' : '#2a2a2a';
        ctx.beginPath(); ctx.moveTo(p3.x,p3.y); ctx.lineTo(p2.x,p2.y); ctx.lineTo(b2.x,b2.y); ctx.lineTo(b3.x,b3.y); ctx.fill();
        ctx.strokeStyle = 'rgba(255,255,255,0.15)'; ctx.stroke();
        ctx.fillStyle = type === 2 ? '#2a2a2a' : type === 3 ? '#0e121a' : '#141820';
        ctx.beginPath(); ctx.moveTo(p1.x,p1.y); ctx.lineTo(p2.x,p2.y); ctx.lineTo(p3.x,p3.y); ctx.lineTo(p4.x,p4.y); ctx.fill();
        ctx.strokeStyle = 'rgba(255,255,255,0.25)'; ctx.stroke();
        if (type === 3) {
            const c = project(x+TILE_SIZE/2, y+TILE_SIZE/2, z+h);
            ctx.beginPath(); ctx.ellipse(c.x, c.y, TILE_SIZE*0.3*scale, TILE_SIZE*0.15*scale, 0, 0, Math.PI*2);
            ctx.fillStyle = '#000'; ctx.fill();
            ctx.strokeStyle = '#38bdf8'; ctx.lineWidth = 1.5; ctx.stroke();
        }
    }

    function draw() {
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        const queue = [];
        for (let i = 0; i < grid.length; i++) {
            for (let j = 0; j < grid[i].length; j++) {
                if (grid[i][j] === 0) continue;
                queue.push({ type: 'block', depth: i*TILE_SIZE + j*TILE_SIZE, q: i, r: j, z: -10, h: grid[i][j] === 2 ? TILE_SIZE : 10, bt: (i===hole.q&&j===hole.r) ? 3 : grid[i][j] });
            }
        }
        queue.push({ type: 'ball', depth: ball.x + ball.y });
        queue.sort((a, b) => a.depth - b.depth);

        for (const item of queue) {
            if (item.type === 'block') {
                drawBlock(item.q, item.r, item.z, item.h, item.bt);
            } else {
                const el = getElevation(ball.x, ball.y);
                if (ball.z >= 0 && el >= 0) {
                    const sh = project(ball.x, ball.y, el);
                    const sc = Math.max(0, 1 - (ball.z - el)/100);
                    ctx.beginPath(); ctx.ellipse(sh.x, sh.y, ball.radius*scale*2*sc, ball.radius*scale*sc, 0, 0, Math.PI*2);
                    ctx.fillStyle = 'rgba(0,0,0,0.5)'; ctx.fill();
                }
                const pb = project(ball.x, ball.y, ball.z);
                ctx.beginPath(); ctx.arc(pb.x, pb.y, ball.radius*scale*1.5, 0, Math.PI*2);
                ctx.fillStyle = '#fff'; ctx.shadowColor = '#38bdf8'; ctx.shadowBlur = 10; ctx.fill(); ctx.shadowBlur = 0;
            }
        }

        if (isDragging) {
            const pb = project(ball.x, ball.y, ball.z);
            ctx.beginPath(); ctx.moveTo(pb.x, pb.y);
            const ws = unproject(dragStart.x, dragStart.y), wc = unproject(dragCurrent.x, dragCurrent.y);
            let dx = ws.x - wc.x, dy = ws.y - wc.y, pvx = dx*0.05, pvy = dy*0.05, pvz = Math.min(Math.sqrt(pvx*pvx+pvy*pvy)*0.3, 10);
            let sx = ball.x, sy = ball.y, sz = ball.z;
            for (let s = 0; s < 20; s++) { pvz += GRAVITY; sx += pvx; sy += pvy; sz += pvz; if (sz < 0) break; const pp = project(sx, sy, sz); ctx.lineTo(pp.x, pp.y); }
            ctx.strokeStyle = '#38bdf8'; ctx.lineWidth = 1.5; ctx.setLineDash([4,4]); ctx.stroke(); ctx.setLineDash([]);
        }
    }

    function loop() { update(); draw(); golfAnimFrame = requestAnimationFrame(loop); }

    generateLevel(); resize(); loop();

    const onMouseDown = (e) => {
        if (ball.state !== 'stationary' && ball.z <= ball.radius + 1) return;
        const r = canvas.getBoundingClientRect();
        dragStart.x = e.clientX - r.left; dragStart.y = e.clientY - r.top;
        isDragging = true; dragCurrent = { ...dragStart };
    };

    const onMouseMove = (e) => {
        if (!isDragging) return;
        const r = canvas.getBoundingClientRect();
        dragCurrent.x = e.clientX - r.left; dragCurrent.y = e.clientY - r.top;
    };

    const onMouseUp = () => {
        if (!isDragging) return;
        isDragging = false;
        const ws = unproject(dragStart.x, dragStart.y), wc = unproject(dragCurrent.x, dragCurrent.y);
        ball.vx = (ws.x - wc.x) * 0.05; ball.vy = (ws.y - wc.y) * 0.05;
        ball.vz = Math.min(Math.sqrt(ball.vx**2 + ball.vy**2) * 0.4, 8);
        ball.state = 'falling';
    };

    canvas.addEventListener('mousedown', onMouseDown);
    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);

    window.golfCleanup = () => {
        if (golfAnimFrame) cancelAnimationFrame(golfAnimFrame);
        window.removeEventListener('resize', resize);
        window.removeEventListener('mousemove', onMouseMove);
        window.removeEventListener('mouseup', onMouseUp);
    };
}

// ============= CYBER SNAKE ENGINE =============
function initSetupSnake(container) {
    const canvas = container.querySelector('#snakeCanvas');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');

    const scoreEl = container.querySelector('#snake-score');
    const comboEl = container.querySelector('#snake-combo');
    const highEl = container.querySelector('#snake-high');
    const overlay = container.querySelector('#snake-overlay');
    const overlayTitle = container.querySelector('#overlay-title');
    const overlayDesc = container.querySelector('#overlay-desc');
    const btnStart = container.querySelector('#btn-start-snake');
    const btnSound = container.querySelector('#btn-sound-toggle');
    const powerHud = container.querySelector('#active-power-hud');

    let audioCtx = null;
    let soundEnabled = true;

    function initAudio() {
        if (!audioCtx) {
            try { audioCtx = new (window.AudioContext || window.webkitAudioContext)(); } catch (e) {}
        }
    }

    function playTone(freq, type, duration, gainVal = 0.06) {
        if (!soundEnabled || !audioCtx) return;
        try {
            if (audioCtx.state === 'suspended') audioCtx.resume();
            const osc = audioCtx.createOscillator();
            const gain = audioCtx.createGain();
            osc.type = type;
            osc.frequency.setValueAtTime(freq, audioCtx.currentTime);
            gain.gain.setValueAtTime(gainVal, audioCtx.currentTime);
            gain.gain.exponentialRampToValueAtTime(0.0001, audioCtx.currentTime + duration);
            osc.connect(gain);
            gain.connect(audioCtx.destination);
            osc.start();
            osc.stop(audioCtx.currentTime + duration);
        } catch (e) {}
    }

    if (btnSound) {
        btnSound.addEventListener('click', () => {
            soundEnabled = !soundEnabled;
            btnSound.style.opacity = soundEnabled ? '1' : '0.4';
        });
    }

    let highScore = parseInt(localStorage.getItem('rift_snake_highscore') || '0', 10);
    if (highEl) highEl.textContent = highScore.toLocaleString();

    let width = 0, height = 0;
    const GRID_SIZE = 22;
    let cols = 0, rows = 0;

    function resize() {
        if (!canvas.parentElement) return;
        const dpr = window.devicePixelRatio || 1;
        width = canvas.parentElement.clientWidth;
        height = canvas.parentElement.clientHeight || 300;
        canvas.width = width * dpr;
        canvas.height = height * dpr;
        ctx.scale(dpr, dpr);
        cols = Math.max(12, Math.floor(width / GRID_SIZE));
        rows = Math.max(10, Math.floor(height / GRID_SIZE));
    }
    window.addEventListener('resize', resize);
    resize();

    let isRunning = false;
    let isPaused = false;
    let score = 0;
    let combo = 1;
    let comboTimer = 0;
    const COMBO_WINDOW = 180;

    let snake = [];
    let dir = { x: 1, y: 0 };
    let nextDir = { x: 1, y: 0 };
    let food = { x: 0, y: 0, type: 'standard' };
    let powerEffect = null;
    let particles = [];
    let floatingTexts = [];

    function spawnParticles(x, y, color, count = 12) {
        for (let i = 0; i < count; i++) {
            const angle = Math.random() * Math.PI * 2;
            const speed = Math.random() * 3 + 1;
            particles.push({
                x: x * GRID_SIZE + GRID_SIZE / 2,
                y: y * GRID_SIZE + GRID_SIZE / 2,
                vx: Math.cos(angle) * speed,
                vy: Math.sin(angle) * speed,
                color: color,
                size: Math.random() * 2 + 1,
                alpha: 1,
                decay: Math.random() * 0.03 + 0.02
            });
        }
    }

    function addFloatingText(x, y, text, color) {
        floatingTexts.push({
            x: x * GRID_SIZE + GRID_SIZE / 2,
            y: y * GRID_SIZE,
            text: text,
            color: color,
            alpha: 1,
            vy: -1.2
        });
    }

    function placeFood() {
        let valid = false;
        let candidate = { x: 0, y: 0 };
        while (!valid) {
            candidate.x = Math.floor(Math.random() * (cols - 4)) + 2;
            candidate.y = Math.floor(Math.random() * (rows - 4)) + 2;
            valid = !snake.some(seg => seg.x === candidate.x && seg.y === candidate.y);
        }

        const rand = Math.random();
        let type = 'standard';
        if (rand > 0.88) type = 'quantum';
        else if (rand > 0.76) type = 'chrono';
        else if (rand > 0.64) type = 'overdrive';

        food = { x: candidate.x, y: candidate.y, type: type };
    }

    function resetGame() {
        initAudio();
        const startX = Math.floor(cols / 2);
        const startY = Math.floor(rows / 2);
        snake = [
            { x: startX, y: startY },
            { x: startX - 1, y: startY },
            { x: startX - 2, y: startY }
        ];
        dir = { x: 1, y: 0 };
        nextDir = { x: 1, y: 0 };
        score = 0;
        combo = 1;
        comboTimer = 0;
        powerEffect = null;
        particles = [];
        floatingTexts = [];
        if (scoreEl) scoreEl.textContent = '0';
        if (comboEl) comboEl.textContent = '1x';
        if (powerHud) powerHud.style.display = 'none';
        placeFood();
        isRunning = true;
        isPaused = false;
        if (overlay) overlay.style.display = 'none';
        playTone(520, 'sine', 0.1, 0.1);
    }

    function gameOver() {
        isRunning = false;
        playTone(140, 'sawtooth', 0.35, 0.15);
        if (score > highScore) {
            highScore = score;
            localStorage.setItem('rift_snake_highscore', highScore.toString());
            if (highEl) highEl.textContent = highScore.toLocaleString();
        }

        if (powerHud) powerHud.style.display = 'none';
        if (overlayTitle) overlayTitle.textContent = 'GAME OVER';
        if (overlayDesc) overlayDesc.innerHTML = `Score: <strong style="color: #fff;">${score.toLocaleString()}</strong> · Best Multiplier: <strong style="color: #38bdf8;">${combo}x</strong>`;
        if (btnStart) btnStart.textContent = 'PLAY AGAIN';
        if (overlay) overlay.style.display = 'flex';
    }

    function handleKeyDown(e) {
        if (!isRunning && (e.key === ' ' || e.key === 'Enter')) {
            resetGame();
            return;
        }

        if (e.key === ' ' || e.key.toLowerCase() === 'p') {
            isPaused = !isPaused;
            if (isPaused) {
                if (overlayTitle) overlayTitle.textContent = 'PAUSED';
                if (overlayDesc) overlayDesc.textContent = 'Press SPACE to continue.';
                if (btnStart) btnStart.textContent = 'RESUME';
                if (overlay) overlay.style.display = 'flex';
            } else {
                if (overlay) overlay.style.display = 'none';
            }
            return;
        }

        const k = e.key.toLowerCase();
        if ((k === 'arrowup' || k === 'w') && dir.y === 0) nextDir = { x: 0, y: -1 };
        else if ((k === 'arrowdown' || k === 's') && dir.y === 0) nextDir = { x: 0, y: 1 };
        else if ((k === 'arrowleft' || k === 'a') && dir.x === 0) nextDir = { x: -1, y: 0 };
        else if ((k === 'arrowright' || k === 'd') && dir.x === 0) nextDir = { x: 1, y: 0 };
    }
    window.addEventListener('keydown', handleKeyDown);

    if (btnStart) {
        btnStart.addEventListener('click', () => {
            if (isPaused) {
                isPaused = false;
                if (overlay) overlay.style.display = 'none';
            } else {
                resetGame();
            }
        });
    }

    let lastStepTime = 0;
    let animFrame = null;

    function gameLoop(time) {
        animFrame = requestAnimationFrame(gameLoop);

        let baseSpeed = Math.max(70, 115 - Math.floor(score / 250) * 4);
        if (powerEffect && powerEffect.type === 'overdrive') baseSpeed = 55;
        if (powerEffect && powerEffect.type === 'chrono') baseSpeed = 160;

        if (isRunning && !isPaused) {
            if (comboTimer > 0) {
                comboTimer--;
                if (comboTimer <= 0) {
                    combo = 1;
                    if (comboEl) {
                        comboEl.textContent = '1x';
                        comboEl.style.color = '#38bdf8';
                    }
                }
            }

            if (powerEffect) {
                powerEffect.duration--;
                const secondsLeft = (powerEffect.duration / 60).toFixed(1);
                if (powerHud) {
                    powerHud.style.display = 'block';
                    if (powerEffect.type === 'overdrive') {
                        powerHud.style.background = 'rgba(255, 215, 0, 0.15)';
                        powerHud.style.border = '1px solid rgba(255, 215, 0, 0.4)';
                        powerHud.style.color = '#ffd700';
                        powerHud.textContent = `⚡ OVERDRIVE · ${secondsLeft}s`;
                    } else if (powerEffect.type === 'chrono') {
                        powerHud.style.background = 'rgba(56, 189, 248, 0.15)';
                        powerHud.style.border = '1px solid rgba(56, 189, 248, 0.4)';
                        powerHud.style.color = '#38bdf8';
                        powerHud.textContent = `⏳ SLOW-MO · ${secondsLeft}s`;
                    }
                }
                if (powerEffect.duration <= 0) {
                    powerEffect = null;
                    if (powerHud) powerHud.style.display = 'none';
                }
            }

            if (time - lastStepTime > baseSpeed) {
                lastStepTime = time;
                dir = nextDir;
                const head = { x: snake[0].x + dir.x, y: snake[0].y + dir.y };

                if (head.x < 0 || head.x >= cols || head.y < 0 || head.y >= rows) {
                    gameOver();
                    return;
                }

                if (snake.some(seg => seg.x === head.x && seg.y === head.y)) {
                    gameOver();
                    return;
                }

                snake.unshift(head);

                if (head.x === food.x && head.y === food.y) {
                    comboTimer = COMBO_WINDOW;
                    combo = Math.min(6, combo + 1);
                    if (comboEl) {
                        comboEl.textContent = `${combo}x`;
                        comboEl.style.color = combo >= 3 ? '#34d399' : '#38bdf8';
                    }

                    let points = 50 * combo;
                    let particleColor = '#00ffcc';

                    if (food.type === 'overdrive') {
                        powerEffect = { type: 'overdrive', duration: 300 };
                        points = 150 * combo;
                        particleColor = '#ffd700';
                        addFloatingText(food.x, food.y, `⚡ 2X SPEED! +${points}`, '#ffd700');
                        playTone(660, 'triangle', 0.15, 0.1);
                    } else if (food.type === 'chrono') {
                        powerEffect = { type: 'chrono', duration: 300 };
                        points = 100 * combo;
                        particleColor = '#38bdf8';
                        addFloatingText(food.x, food.y, `⏳ SLOW-MO! +${points}`, '#38bdf8');
                        playTone(440, 'sine', 0.25, 0.1);
                    } else if (food.type === 'quantum') {
                        points = 500 * combo;
                        particleColor = '#f43f5e';
                        addFloatingText(food.x, food.y, `💎 +${points} GEM!`, '#f43f5e');
                        playTone(880, 'sine', 0.2, 0.12);
                    } else {
                        addFloatingText(food.x, food.y, `+${points}`, '#00ffcc');
                        playTone(480 + combo * 40, 'sine', 0.08, 0.08);
                    }

                    score += points;
                    if (scoreEl) scoreEl.textContent = score.toLocaleString();
                    spawnParticles(food.x, food.y, particleColor, 14);
                    placeFood();
                } else {
                    snake.pop();
                }
            }
        }

        renderScene();
    }

    function renderScene() {
        ctx.clearRect(0, 0, width, height);

        // Grid lines
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.025)';
        ctx.lineWidth = 1;
        for (let x = 0; x < width; x += GRID_SIZE) {
            ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, height); ctx.stroke();
        }
        for (let y = 0; y < height; y += GRID_SIZE) {
            ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(width, y); ctx.stroke();
        }

        // Render Food
        const fx = food.x * GRID_SIZE + GRID_SIZE / 2;
        const fy = food.y * GRID_SIZE + GRID_SIZE / 2;
        const pulse = Math.sin(Date.now() * 0.008) * 1.5;

        ctx.save();
        if (food.type === 'overdrive') {
            ctx.shadowColor = '#ffd700'; ctx.shadowBlur = 12; ctx.fillStyle = '#ffd700';
            ctx.beginPath(); ctx.arc(fx, fy, (GRID_SIZE / 2 - 2) + pulse, 0, Math.PI * 2); ctx.fill();
            ctx.fillStyle = '#07090e'; ctx.font = 'bold 10px sans-serif'; ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
            ctx.fillText('⚡', fx, fy);
        } else if (food.type === 'chrono') {
            ctx.shadowColor = '#38bdf8'; ctx.shadowBlur = 12; ctx.fillStyle = '#38bdf8';
            ctx.beginPath(); ctx.arc(fx, fy, (GRID_SIZE / 2 - 2) + pulse, 0, Math.PI * 2); ctx.fill();
            ctx.fillStyle = '#07090e'; ctx.font = 'bold 10px sans-serif'; ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
            ctx.fillText('⏳', fx, fy);
        } else if (food.type === 'quantum') {
            ctx.shadowColor = '#f43f5e'; ctx.shadowBlur = 14; ctx.fillStyle = '#f43f5e';
            ctx.beginPath(); ctx.arc(fx, fy, (GRID_SIZE / 2 - 1) + pulse, 0, Math.PI * 2); ctx.fill();
            ctx.fillStyle = '#ffffff'; ctx.font = 'bold 10px sans-serif'; ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
            ctx.fillText('💎', fx, fy);
        } else {
            ctx.shadowColor = '#00ffcc'; ctx.shadowBlur = 10; ctx.fillStyle = '#00ffcc';
            ctx.beginPath(); ctx.arc(fx, fy, (GRID_SIZE / 2 - 3) + pulse, 0, Math.PI * 2); ctx.fill();
            ctx.fillStyle = '#07090e'; ctx.beginPath(); ctx.arc(fx, fy, 2, 0, Math.PI * 2); ctx.fill();
        }
        ctx.restore();

        // Snake
        for (let i = snake.length - 1; i >= 0; i--) {
            const seg = snake[i];
            const px = seg.x * GRID_SIZE;
            const py = seg.y * GRID_SIZE;

            ctx.save();
            if (i === 0) {
                ctx.shadowColor = powerEffect ? (powerEffect.type === 'overdrive' ? '#ffd700' : '#38bdf8') : '#00ffcc';
                ctx.shadowBlur = 12;
                ctx.fillStyle = powerEffect ? (powerEffect.type === 'overdrive' ? '#ffd700' : '#38bdf8') : '#00ffcc';
                ctx.beginPath(); ctx.roundRect(px + 1, py + 1, GRID_SIZE - 2, GRID_SIZE - 2, 4); ctx.fill();

                ctx.fillStyle = '#07090e';
                const eyeOffset = 4;
                if (dir.x === 1) {
                    ctx.fillRect(px + GRID_SIZE - 5, py + eyeOffset, 3, 3);
                    ctx.fillRect(px + GRID_SIZE - 5, py + GRID_SIZE - eyeOffset - 3, 3, 3);
                } else if (dir.x === -1) {
                    ctx.fillRect(px + 2, py + eyeOffset, 3, 3);
                    ctx.fillRect(px + 2, py + GRID_SIZE - eyeOffset - 3, 3, 3);
                } else if (dir.y === 1) {
                    ctx.fillRect(px + eyeOffset, py + GRID_SIZE - 5, 3, 3);
                    ctx.fillRect(px + GRID_SIZE - eyeOffset - 3, py + GRID_SIZE - 5, 3, 3);
                } else {
                    ctx.fillRect(px + eyeOffset, py + 2, 3, 3);
                    ctx.fillRect(px + GRID_SIZE - eyeOffset - 3, py + 2, 3, 3);
                }
            } else {
                const ratio = i / snake.length;
                const alpha = Math.max(0.3, 1 - ratio * 0.6);
                ctx.fillStyle = `rgba(0, 255, 204, ${alpha})`;
                if (powerEffect && powerEffect.type === 'overdrive') {
                    ctx.fillStyle = `rgba(255, 215, 0, ${alpha})`;
                } else if (powerEffect && powerEffect.type === 'chrono') {
                    ctx.fillStyle = `rgba(56, 189, 248, ${alpha})`;
                }
                ctx.beginPath(); ctx.roundRect(px + 2, py + 2, GRID_SIZE - 4, GRID_SIZE - 4, 3); ctx.fill();
            }
            ctx.restore();
        }

        // Particles
        for (let i = particles.length - 1; i >= 0; i--) {
            const p = particles[i];
            p.x += p.vx; p.y += p.vy; p.alpha -= p.decay;
            if (p.alpha <= 0) { particles.splice(i, 1); continue; }
            ctx.save(); ctx.globalAlpha = p.alpha; ctx.fillStyle = p.color;
            ctx.beginPath(); ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2); ctx.fill(); ctx.restore();
        }

        // Floating texts
        for (let i = floatingTexts.length - 1; i >= 0; i--) {
            const ft = floatingTexts[i];
            ft.y += ft.vy; ft.alpha -= 0.025;
            if (ft.alpha <= 0) { floatingTexts.splice(i, 1); continue; }
            ctx.save(); ctx.globalAlpha = ft.alpha; ctx.fillStyle = ft.color;
            ctx.font = 'bold 11px sans-serif'; ctx.textAlign = 'center';
            ctx.fillText(ft.text, ft.x, ft.y); ctx.restore();
        }
    }

    window.snakeCleanup = () => {
        window.removeEventListener('keydown', handleKeyDown);
        window.removeEventListener('resize', resize);
        if (animFrame) cancelAnimationFrame(animFrame);
        if (audioCtx) {
            try { audioCtx.close(); } catch(e) {}
        }
    };

    animFrame = requestAnimationFrame(gameLoop);
}
