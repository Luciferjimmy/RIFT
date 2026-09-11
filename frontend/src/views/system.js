import { store } from '../state.js';

export function renderSystemView() {
    const el = document.createElement('div');
    el.className = 'view-container';
    el.innerHTML = `
        <div class="system-dashboard fade-in">
            <h1 class="view-title">System Diagnostics</h1>
            <p class="view-subtitle">Real-time hardware telemetry and storage management.</p>
            
            <div class="sys-grid">
                <!-- Hardware Telemetry -->
                <div class="sys-card sys-panel-main">
                    <h2>Live Hardware Stream</h2>
                    <div class="sys-metric-list">
                        <div class="sys-metric-item">
                            <span class="sys-metric-label">CPU CORE</span>
                            <span class="sys-metric-val" id="tele-cpu">--%</span>
                            <div class="sys-bar-bg"><div class="sys-bar-fill" id="bar-cpu"></div></div>
                        </div>
                        <div class="sys-metric-item">
                            <span class="sys-metric-label">MEMORY</span>
                            <span class="sys-metric-val" id="tele-ram">-- GB</span>
                            <div class="sys-bar-bg"><div class="sys-bar-fill" id="bar-ram"></div></div>
                        </div>
                        <div class="sys-metric-item">
                            <span class="sys-metric-label">SWAP FILE</span>
                            <span class="sys-metric-val" id="tele-swap">-- MB</span>
                        </div>
                        <div class="sys-metric-item">
                            <span class="sys-metric-label">NET DOWN</span>
                            <span class="sys-metric-val" id="tele-net-dl">-- MB/s</span>
                        </div>
                        <div class="sys-metric-item">
                            <span class="sys-metric-label">NET UP</span>
                            <span class="sys-metric-val" id="tele-net-ul">-- MB/s</span>
                        </div>
                    </div>
                </div>

                <!-- Active Game Session -->
                <div class="sys-card">
                    <h2>Active Translation Layer</h2>
                    <div id="active-game-panel" class="sys-empty-state">
                        <div class="pulse-dot" style="background: rgba(255,255,255,0.2);"></div>
                        <p>No active game session detected.</p>
                    </div>
                    <div id="active-game-metrics" style="display: none;">
                        <div class="sys-metric-item">
                            <span class="sys-metric-label">TRANSLATION CPU LOAD</span>
                            <span class="sys-metric-val highlight-val" id="tele-game-cpu">--%</span>
                        </div>
                        <div class="sys-metric-item">
                            <span class="sys-metric-label">GAME & WINE MEMORY</span>
                            <span class="sys-metric-val highlight-val" id="tele-game-ram">-- MB</span>
                        </div>
                        <div class="sys-metric-item" style="margin-top: 20px; width: 100%;">
                            <canvas id="telemetryChart" height="100"></canvas>
                        </div>
                        <div class="sys-metric-item" style="margin-top: 10px;">
                            <span class="sys-metric-label">POST-MORTEM BUFFER</span>
                            <span class="sys-metric-val" style="color: #34d399;">RECORDING (15m)</span>
                        </div>
                    </div>
                </div>

                <!-- Playtime Stats -->
                <div class="sys-card sys-panel-wide">
                    <h2>Playtime Statistics</h2>
                    <div style="display: flex; gap: 30px; margin-top: 15px;">
                        <div style="flex: 0 0 150px;">
                            <div class="sys-metric-label">GLOBAL TOTAL</div>
                            <div class="sys-metric-val" id="tele-playtime-total" style="font-size: 2rem; color: #60a5fa; margin-top: 5px;">-- hrs</div>
                        </div>
                        <div id="playtime-banners" style="flex: 1; display: flex; flex-direction: column; gap: 8px; max-height: 200px; overflow-y: auto;">
                            <!-- banners injected here -->
                        </div>
                    </div>
                </div>

                <!-- Silicon Identity & Hardware Truth -->
                <div class="sys-card sys-panel-wide">
                    <h2>Apple Silicon Architecture & Truth</h2>
                    <div class="sys-metric-list" style="display: grid; grid-template-columns: repeat(auto-fit, minmax(260px, 1fr)); gap: 14px; margin-top: 14px;">
                        <div class="sys-metric-item" style="flex-direction: column; align-items: flex-start; gap: 4px; padding: 14px 16px; background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.06); border-radius: 10px;">
                            <span class="sys-metric-label">PROCESSOR</span>
                            <span class="sys-metric-val" id="hw-chip" style="font-size: 0.95rem; color: #fff;">Apple Silicon</span>
                            <span style="font-size: 0.72rem; color: rgba(255,255,255,0.4); font-family: var(--font-sans);">Ready to defy physics</span>
                        </div>
                        <div class="sys-metric-item" style="flex-direction: column; align-items: flex-start; gap: 4px; padding: 14px 16px; background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.06); border-radius: 10px;">
                            <span class="sys-metric-label">GPU CORES</span>
                            <span class="sys-metric-val" id="hw-gpu" style="font-size: 0.95rem; color: #38bdf8;">-- Cores</span>
                            <span style="font-size: 0.72rem; color: rgba(255,255,255,0.4); font-family: var(--font-sans);">Finally being used for something other than Figma</span>
                        </div>
                        <div class="sys-metric-item" style="flex-direction: column; align-items: flex-start; gap: 4px; padding: 14px 16px; background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.06); border-radius: 10px;">
                            <span class="sys-metric-label">THERMAL STATE</span>
                            <span class="sys-metric-val" id="hw-thermal" style="font-size: 0.95rem; color: #34d399;">Nominal</span>
                            <span id="hw-thermal-sub" style="font-size: 0.72rem; color: rgba(255,255,255,0.4); font-family: var(--font-sans);">Nominal (Chilly & quiet. For now.)</span>
                        </div>
                        <div class="sys-metric-item" style="flex-direction: column; align-items: flex-start; gap: 4px; padding: 14px 16px; background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.06); border-radius: 10px;">
                            <span class="sys-metric-label">POWER STATUS</span>
                            <span class="sys-metric-val" id="hw-power" style="font-size: 0.95rem; color: #fbbf24;">Wall Power</span>
                            <span id="hw-power-sub" style="font-size: 0.72rem; color: rgba(255,255,255,0.4); font-family: var(--font-sans);">Wall Power (Unlimited power)</span>
                        </div>
                        <div class="sys-metric-item" style="flex-direction: column; align-items: flex-start; gap: 4px; padding: 14px 16px; background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.06); border-radius: 10px;">
                            <span class="sys-metric-label">ROSETTA 2 RUNTIME</span>
                            <span class="sys-metric-val" id="hw-rosetta" style="font-size: 0.95rem; color: #a78bfa;">Installed</span>
                            <span style="font-size: 0.72rem; color: rgba(255,255,255,0.4); font-family: var(--font-sans);">The unsung hero doing Apple's heavy lifting</span>
                        </div>
                        <div class="sys-metric-item" style="flex-direction: column; align-items: flex-start; gap: 4px; padding: 14px 16px; background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.06); border-radius: 10px;">
                            <span class="sys-metric-label">GRAPHICS API</span>
                            <span class="sys-metric-val" id="hw-metal" style="font-size: 0.95rem; color: #fff;">Apple Metal</span>
                            <span style="font-size: 0.72rem; color: rgba(255,255,255,0.4); font-family: var(--font-sans);">DirectX draw calls translated in real-time</span>
                        </div>
                    </div>
                </div>

                <!-- Storage Manager -->
                <div class="sys-card sys-panel-wide">
                    <h2>Storage Breakdown</h2>
                    <div class="storage-manager">
                        <div class="storage-stats">
                            <div class="storage-item">
                                <span class="storage-label">IMPORTED GAMES</span>
                                <span class="storage-val" id="store-games">-- GB</span>
                            </div>
                            <div class="storage-item">
                                <span class="storage-label">TRANSLATION ENGINES</span>
                                <span class="storage-val" id="store-engines">-- GB</span>
                            </div>
                            <div class="storage-item">
                                <span class="storage-label">CACHE & TEMP</span>
                                <span class="storage-val" id="store-cache">-- MB</span>
                            </div>
                        </div>
                        
                        <div class="storage-bar">
                            <div class="storage-chunk games-chunk" id="chunk-games"></div>
                            <div class="storage-chunk engines-chunk" id="chunk-engines"></div>
                            <div class="storage-chunk cache-chunk" id="chunk-cache"></div>
                        </div>

                        <div class="storage-actions">
                            <button class="sys-btn sys-btn-outline" id="btn-purge-cache">
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                                PURGE CACHE
                            </button>
                            <button class="sys-btn sys-btn-outline" id="btn-refresh-storage">
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"></polyline><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path></svg>
                                REFRESH
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `;

    // Hook up dynamic telemetry
    let unsub = null;
    if (window.runtime && window.runtime.EventsOn) {
        unsub = window.runtime.EventsOn("sys-stats", (stats) => {
        const tCpu = el.querySelector('#tele-cpu');
        const tRam = el.querySelector('#tele-ram');
        const tSwap = el.querySelector('#tele-swap');
        const tDl = el.querySelector('#tele-net-dl');
        const tUl = el.querySelector('#tele-net-ul');

        if (tCpu) tCpu.innerText = stats.cpu;
        if (tRam) tRam.innerText = stats.ram;
        if (tSwap) tSwap.innerText = stats.swap;
        if (tDl) tDl.innerText = stats.netDl;
        if (tUl) tUl.innerText = stats.netUl;

        // Custom SVG bars
        const bCpu = el.querySelector('#bar-cpu');
        const bRam = el.querySelector('#bar-ram');
        if (bCpu && stats.cpu !== '--%') {
            bCpu.style.width = stats.cpu;
        }
        if (bRam && stats.ram !== '-- GB') {
            // Very hacky parse "X.X / Y.Y GB"
            const parts = stats.ram.split('/');
            if (parts.length === 2) {
                const used = parseFloat(parts[0]);
                const total = parseFloat(parts[1]);
                if (total > 0) {
                    bRam.style.width = ((used / total) * 100) + '%';
                }
            }
        }

        // Active game metrics
        const agPanel = el.querySelector('#active-game-panel');
        const agMetrics = el.querySelector('#active-game-metrics');
        if (stats.gameCpu !== "--%" && stats.gameCpu !== "0.0%") {
            if (agPanel) agPanel.style.display = 'none';
            if (agMetrics) agMetrics.style.display = 'block';
            
            const titles = el.querySelectorAll('h2');
            if(titles.length > 1) titles[1].innerText = "Active Translation Layer";

            const recStatus = el.querySelector('.sys-metric-val[style*="color: #34d399"]');
            if (recStatus) recStatus.innerText = "RECORDING (15m)";
        } else {
            if (!window.hasSeenGameData) {
                if (agPanel) agPanel.style.display = 'flex';
                if (agMetrics) agMetrics.style.display = 'none';
            } else {
                // Post-mortem state
                const titles = el.querySelectorAll('h2');
                if(titles.length > 1) titles[1].innerText = "Previous Session (Recorded)";
                
                const recStatus = el.querySelector('.sys-metric-val[style*="color: #34d399"]');
                if (recStatus) recStatus.innerText = "SAVED TO CLOUD";
            }
        }
        });
    }

    const refreshStorage = async () => {
        const stats = await window.go.rift.App.GetStorageStats();
        
        const gGB = (stats.totalGamesSize / (1024*1024*1024)).toFixed(1);
        const eGB = (stats.totalEngineSize / (1024*1024*1024)).toFixed(1);
        const cMB = (stats.totalCacheSize / (1024*1024)).toFixed(1);

        el.querySelector('#store-games').innerText = gGB + ' GB';
        el.querySelector('#store-engines').innerText = eGB + ' GB';
        el.querySelector('#store-cache').innerText = cMB + ' MB';

        // Very basic width setting for visual sparkline
        const total = stats.totalGamesSize + stats.totalEngineSize + stats.totalCacheSize + 1; // +1 to avoid div0
        el.querySelector('#chunk-games').style.width = ((stats.totalGamesSize / total) * 100) + '%';
        el.querySelector('#chunk-engines').style.width = ((stats.totalEngineSize / total) * 100) + '%';
        el.querySelector('#chunk-cache').style.width = ((stats.totalCacheSize / total) * 100) + '%';
    };

    // Load initial storage
    refreshStorage();

    el.querySelector('#btn-refresh-storage').addEventListener('click', refreshStorage);
    el.querySelector('#btn-purge-cache').addEventListener('click', async () => {
        const btn = el.querySelector('#btn-purge-cache');
        btn.innerHTML = 'PURGING...';
        await window.go.rift.App.PurgeCache();
        setTimeout(async () => {
            await refreshStorage();
            btn.innerHTML = 'PURGE CACHE';
        }, 1000);
    });

    // Initialize Active Runtime Log
    const fetchLogs = async () => {
        try {
            const logText = await window.go.rift.App.GetSanitizedLogs();
            const logViewer = el.querySelector('#sys-log-viewer');
            if (logViewer) {
                const wasAtBottom = logViewer.scrollHeight - logViewer.scrollTop === logViewer.clientHeight;
                logViewer.textContent = logText;
                if (wasAtBottom) {
                    logViewer.scrollTop = logViewer.scrollHeight;
                }
            }
        } catch(e) {
            console.warn("Failed to fetch logs:", e);
        }
    };
    fetchLogs();
    const logInterval = setInterval(fetchLogs, 2000);

    const btnClearLogs = el.querySelector('#btn-clear-logs');
    if (btnClearLogs) {
        btnClearLogs.addEventListener('click', () => {
            const logViewer = el.querySelector('#sys-log-viewer');
            if (logViewer) logViewer.textContent = 'Cleared.';
        });
    }

    // Clean up event listener when leaving view
    el._cleanup = () => {
        if (typeof unsub === 'function') unsub();
        clearInterval(logInterval);
        if (typeof el._telemetryUnsub === 'function') el._telemetryUnsub();
    };

    // Initialize Chart.js
    setTimeout(() => {
        const ctx = el.querySelector('#telemetryChart');
        if (ctx && window.Chart) {
            const teleChart = new window.Chart(ctx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [
                        { label: 'CPU %', data: [], borderColor: '#34d399', tension: 0.4 },
                        { label: 'RAM MB', data: [], borderColor: '#60a5fa', tension: 0.4 }
                    ]
                },
                options: {
                    responsive: true,
                    animation: false,
                    scales: {
                        x: { display: false },
                        y: { beginAtZero: true }
                    },
                    plugins: { legend: { display: true, labels: { color: '#888' } } }
                }
            });

            if (window.runtime && window.runtime.EventsOn) {
                el._telemetryUnsub = window.runtime.EventsOn("TelemetryUpdate", (data) => {
                window.hasSeenGameData = true;
                const now = new Date().toLocaleTimeString();
                teleChart.data.labels.push(now);
                teleChart.data.datasets[0].data.push(data.cpu);
                teleChart.data.datasets[1].data.push(data.ram);

                // Keep only last 180 points (15 minutes at 5s interval)
                if (teleChart.data.labels.length > 180) {
                    teleChart.data.labels.shift();
                    teleChart.data.datasets[0].data.shift();
                    teleChart.data.datasets[1].data.shift();
                }
                teleChart.update();
                
                const gc = el.querySelector('#tele-game-cpu');
                const gr = el.querySelector('#tele-game-ram');
                if (gc) gc.innerText = data.cpu + '%';
                if (gr) gr.innerText = data.ram + ' MB';
                });
            }
            // Fetch initial data
            (async () => {
                try {
                    const ptStr = await window.go.rift.App.GetPlaytimeStats();
                    const pt = JSON.parse(ptStr);
                    const hours = (pt.total / 3600).toFixed(1);
                    const telePt = el.querySelector('#tele-playtime-total');
                    if (telePt) telePt.innerText = hours + ' hrs';
                    
                    const bannersContainer = el.querySelector('#playtime-banners');
                    if (bannersContainer && pt.per_game) {
                        const gamesMap = {};
                        store.state.games.forEach(g => { gamesMap[g.id] = g; });
                        
                        let html = '';
                        Object.entries(pt.per_game).sort((a, b) => b[1] - a[1]).forEach(([gameId, secs]) => {
                            const g = gamesMap[gameId];
                            if (!g) return;
                            const h = (secs / 3600).toFixed(1);
                            html += `
                                <div style="display: flex; align-items: center; justify-content: space-between; background: rgba(255,255,255,0.03); padding: 8px 12px; border-radius: 6px; border: 1px solid rgba(255,255,255,0.05);">
                                    <div style="display: flex; align-items: center; gap: 10px;">
                                        <img src="${g.cover}" style="width: 24px; height: 24px; object-fit: cover; border-radius: 4px;">
                                        <span style="font-family: var(--font-sans); font-size: 0.85rem; font-weight: 500; color: white;">${g.name}</span>
                                    </div>
                                    <span style="font-family: var(--font-mono); font-size: 0.8rem; color: rgba(255,255,255,0.6);">${h} hrs</span>
                                </div>
                            `;
                        });
                        bannersContainer.innerHTML = html;
                    }
                    
                    const pmStr = await window.go.rift.App.GetLastSessionTelemetry();
                    const pm = JSON.parse(pmStr);
                    if (pm && pm.length > 0) {
                        window.hasSeenGameData = true;
                        let totalCpu = 0, totalRam = 0;
                        pm.forEach(data => {
                            let cpuVal = data.cpu || 0;
                            // If raw CPU was recorded in per-core format (> 100%), normalize by 8 cores
                            if (cpuVal > 100) {
                                cpuVal = cpuVal / 8;
                            }
                            if (cpuVal > 100) cpuVal = 100;

                            totalCpu += cpuVal;
                            totalRam += (data.ram || 0);
                            const d = new Date(data.timestamp * 1000);
                            teleChart.data.labels.push(d.toLocaleTimeString());
                            teleChart.data.datasets[0].data.push(Number(cpuVal.toFixed(1)));
                            teleChart.data.datasets[1].data.push(Number((data.ram || 0).toFixed(0)));
                        });
                        teleChart.update();
                        
                        let avgCpuVal = totalCpu / pm.length;
                        if (avgCpuVal > 100) avgCpuVal = avgCpuVal / 8;
                        if (avgCpuVal > 100) avgCpuVal = 100;

                        const avgCpu = avgCpuVal.toFixed(1);
                        const avgRam = (totalRam / pm.length).toFixed(0);
                        const cpuEl = el.querySelector('#tele-game-cpu');
                        const ramEl = el.querySelector('#tele-game-ram');
                        if (cpuEl) cpuEl.innerText = `${avgCpu}% (Avg)`;
                        if (ramEl) ramEl.innerText = `${avgRam} MB (Avg)`;

                        const agPanel = el.querySelector('#active-game-panel');
                        const agMetrics = el.querySelector('#active-game-metrics');
                        if (agPanel) agPanel.style.display = 'none';
                        if (agMetrics) agMetrics.style.display = 'block';
                        
                        const titles = el.querySelectorAll('h2');
                        if(titles.length > 1) titles[1].innerText = "Previous Session (Recorded)";
                        
                        const recStatus = el.querySelector('.sys-metric-val[style*="color: #34d399"]');
                        if (recStatus) recStatus.innerText = "SAVED TO CLOUD";
                    }
                } catch(e) {
                    console.error("Failed to load initial stats", e);
                }

                // Populate Apple Silicon Architecture & Truth
                try {
                    const specs = await window.go.rift.App.GetSystemSpecs();
                    if (specs) {
                        const chipEl = el.querySelector('#hw-chip');
                        const gpuEl = el.querySelector('#hw-gpu');
                        const thermalEl = el.querySelector('#hw-thermal');
                        const thermalSubEl = el.querySelector('#hw-thermal-sub');
                        const powerEl = el.querySelector('#hw-power');
                        const powerSubEl = el.querySelector('#hw-power-sub');
                        const rosettaEl = el.querySelector('#hw-rosetta');
                        const metalEl = el.querySelector('#hw-metal');

                        if (chipEl && specs.Chip) chipEl.textContent = specs.Chip;
                        if (gpuEl && specs.GPUCores) gpuEl.textContent = `${specs.GPUCores} Cores`;
                        
                        const thermalStr = (specs.ThermalState || 'nominal').toLowerCase();
                        const isNominal = thermalStr === 'nominal';
                        if (thermalEl) {
                            thermalEl.textContent = specs.ThermalState ? specs.ThermalState.charAt(0).toUpperCase() + specs.ThermalState.slice(1) : 'Nominal';
                            thermalEl.style.color = isNominal ? '#34d399' : '#f87171';
                        }
                        if (thermalSubEl) {
                            thermalSubEl.textContent = isNominal ? 'Nominal (Chilly & quiet. For now.)' : 'Warming up — fans might wake up';
                        }

                        const onBattery = specs.OnBattery;
                        if (powerEl) {
                            powerEl.textContent = onBattery ? 'On Battery' : 'Wall Power';
                            powerEl.style.color = onBattery ? '#fbbf24' : '#34d399';
                        }
                        if (powerSubEl) {
                            powerSubEl.textContent = onBattery ? 'On Battery (Bold move, hope you have a charger nearby)' : 'Wall Power (Unlimited power)';
                        }

                        if (rosettaEl) {
                            rosettaEl.textContent = specs.Rosetta2 ? 'Installed' : 'Installed (Native Silicon)';
                        }
                        if (metalEl && specs.MetalSupport) {
                            metalEl.textContent = specs.MetalSupport;
                        }
                    }
                } catch(e) {
                    console.error("Failed to fetch system specs", e);
                }
            })();
        }
    }, 100);

    return el;
}
