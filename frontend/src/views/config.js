import { store } from '../state.js';
import { appendLog } from '../main.js';

export function renderConfig() {
    const container = document.createElement('div');
    container.className = 'view-premium';
    container.style.width = '100%';
    container.style.height = '100%';
    container.style.overflowY = 'auto';

    // Instant loading state while backend synchronizes
    container.innerHTML = `
        <div class="settings-content-wrapper" style="max-width: 880px; margin: 0 auto; padding: 36px 32px 80px;">
            <div style="border-bottom: 1px solid rgba(255,255,255,0.06); padding-bottom: 24px; margin-bottom: 32px;">
                <h1 style="font-family: 'Plus Jakarta Sans', sans-serif; font-weight: 800; font-size: 2rem; letter-spacing: -0.5px; color: white; margin: 0 0 6px 0;">Configuration</h1>
                <p style="font-family: 'Inter', sans-serif; font-size: 0.88rem; color: rgba(255,255,255,0.45); margin: 0;">Per-game graphics engine, executable selection, and translation parameters.</p>
            </div>
            <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; padding-top: 60px;">
                <div style="width: 36px; height: 36px; border-radius: 50%; border: 2px solid rgba(56, 189, 248, 0.2); border-top-color: #38bdf8; animation: rift-spin 1s linear infinite;"></div>
                <div style="margin-top: 18px; color: rgba(255,255,255,0.4); font-family: 'Inter', sans-serif; font-size: 0.85rem; letter-spacing: 0.5px;">Scanning Game Capsules...</div>
            </div>
            <style>
                @keyframes rift-spin { to { transform: rotate(360deg); } }
            </style>
        </div>
    `;

    async function init() {
        let games = await window.go.rift.App.SyncLibrary();

        if (!games || games.length === 0) {
            if (store && store.state && store.state.games) {
                games = store.state.games;
            }
        }

        // Only show installed games
        if (games) {
            games = games.filter(g => g.status !== 'not_installed' && g.status !== 'not_downloaded');
        }

        if (!games || games.length === 0) {
            container.innerHTML = `
                <div class="settings-content-wrapper" style="max-width: 880px; margin: 0 auto; padding: 36px 32px 80px;">
                    <div style="border-bottom: 1px solid rgba(255,255,255,0.06); padding-bottom: 24px; margin-bottom: 32px;">
                        <h1 style="font-family: 'Plus Jakarta Sans', sans-serif; font-weight: 800; font-size: 2rem; letter-spacing: -0.5px; color: white; margin: 0 0 6px 0;">Configuration</h1>
                        <p style="font-family: 'Inter', sans-serif; font-size: 0.88rem; color: rgba(255,255,255,0.45); margin: 0;">Per-game graphics engine, executable selection, and translation parameters.</p>
                    </div>
                    <div style="text-align: center; padding: 64px 20px; background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.05); border-radius: 12px;">
                        <div style="font-size: 2.5rem; margin-bottom: 12px; opacity: 0.3;">⚙</div>
                        <h2 style="font-family: 'Plus Jakarta Sans', sans-serif; font-weight: 700; font-size: 1.2rem; color: white; margin: 0 0 8px 0;">No Configurable Games Found</h2>
                        <p style="font-family: 'Inter', sans-serif; font-size: 0.85rem; color: rgba(255,255,255,0.4); margin: 0;">Install a game to customize translation environments, binaries, and DirectX parameters.</p>
                    </div>
                </div>
            `;
            return;
        }

        let selectedGameId = games[0].id;
        let runtimeConfig = await window.go.rift.App.GetGameRuntimeConfig(selectedGameId);

        const cardStyle = `
            background: rgba(255, 255, 255, 0.02);
            border: 1px solid rgba(255, 255, 255, 0.06);
            border-radius: 12px;
            padding: 24px 28px;
            margin-bottom: 20px;
        `;

        const rowStyle = `
            display: flex;
            justify-content: space-between;
            align-items: center;
            background: rgba(0, 0, 0, 0.25);
            border: 1px solid rgba(255, 255, 255, 0.05);
            padding: 14px 20px;
            border-radius: 8px;
        `;

        const renderForm = async () => {
            if (!runtimeConfig) {
                runtimeConfig = {
                    engine: "d3dmetal", enhanced_sync: "msync", metal_hud: false, metal_trace: false,
                    dxr_enabled: false, avx_enabled: true, dxvk_async: true, dxvk_hud: "off",
                    retina_mode: false, resolution_scale: "1.0", manual_override: false, launch_args: [],
                    dxvk_frame_rate: 0
                };
            }
            if (!runtimeConfig.engine) runtimeConfig.engine = "d3dmetal";
            if (!runtimeConfig.enhanced_sync) runtimeConfig.enhanced_sync = "msync";
            if (!runtimeConfig.dxvk_hud) runtimeConfig.dxvk_hud = "off";
            if (!runtimeConfig.resolution_scale) runtimeConfig.resolution_scale = "1.0";

            let executables = [];
            try {
                executables = await window.go.rift.App.GetGameExecutables(selectedGameId) || [];
            } catch (err) {
                console.error("Failed to load executables:", err);
            }

            const engine = runtimeConfig.engine || 'd3dmetal';
            const isD3DMetal = (engine === 'd3dmetal' || engine === 'gptk');
            const isDXVK = (engine === 'dxvk');
            const isWine3D = (engine === 'wine3d' || engine === 'wined3d');
            const isManualOverride = runtimeConfig.manual_override || false;

            const gamesOptions = games.map(g =>
                `<option value="${g.id}" ${g.id === selectedGameId ? 'selected' : ''} style="background: #12141f; color: white;">${g.name}</option>`
            ).join('');

            // Executable Selector Card
            let exeSelectorHTML = '';
            if (executables && executables.length > 0) {
                exeSelectorHTML = `
                    <div style="${cardStyle}">
                        <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px;">
                            <div style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif;">Primary Executable (.exe)</div>
                            <div style="color: #38bdf8; font-size: 0.72rem; font-family: 'JetBrains Mono', monospace; background: rgba(56, 189, 248, 0.1); border: 1px solid rgba(56, 189, 248, 0.25); padding: 2px 8px; border-radius: 4px; font-weight: 600;">${executables.length} BINARIES DETECTED</div>
                        </div>
                        <div style="color: rgba(255,255,255,0.45); font-size: 0.82rem; line-height: 1.4; margin-bottom: 16px;">
                            Select which binary RIFT executes. If a launcher was auto-selected, switch to the primary shipping game executable.
                        </div>

                        <div style="display: flex; flex-direction: column; gap: 8px;">
                            ${executables.map(exe => `
                                <label class="exe-option-label" style="display: flex; align-items: center; justify-content: space-between; background: ${exe.isSelected ? 'rgba(56, 189, 248, 0.08)' : 'rgba(0,0,0,0.2)'}; border: 1px solid ${exe.isSelected ? 'rgba(56, 189, 248, 0.35)' : 'rgba(255,255,255,0.05)'}; padding: 12px 16px; border-radius: 8px; cursor: pointer; transition: all 0.2s ease;">
                                    <div style="display: flex; align-items: center; gap: 12px; min-width: 0;">
                                        <input type="radio" name="game-exe-choice" value="${exe.relativePath}" ${exe.isSelected ? 'checked' : ''} style="accent-color: #38bdf8; width: 16px; height: 16px; cursor: pointer;">
                                        <div style="min-width: 0;">
                                            <div style="color: #fff; font-weight: 600; font-size: 0.88rem; font-family: 'JetBrains Mono', monospace; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">${exe.relativePath}</div>
                                            <div style="display: flex; gap: 8px; margin-top: 3px; font-size: 0.72rem; color: rgba(255,255,255,0.4); font-family: 'Inter', sans-serif;">
                                                <span>${(exe.sizeBytes / (1024*1024)).toFixed(1)} MB</span>
                                                <span>•</span>
                                                <span style="color: #38bdf8; font-weight: 600;">${exe.directXAPI}</span>
                                                <span>•</span>
                                                <span>${exe.is32Bit ? '32-bit' : '64-bit'}</span>
                                            </div>
                                        </div>
                                    </div>
                                    ${exe.isSelected ? '<span style="background: rgba(56, 189, 248, 0.15); color: #38bdf8; padding: 3px 8px; border-radius: 4px; font-size: 0.68rem; font-weight: 700; font-family: \'Inter\', sans-serif; letter-spacing: 0.5px;">ACTIVE</span>' : ''}
                                </label>
                            `).join('')}
                        </div>
                    </div>
                `;
            }

            // Build Dynamic Engine-Specific Settings HTML
            let engineSpecificHTML = '';

            if (isD3DMetal) {
                engineSpecificHTML = `
                    <div style="${cardStyle}">
                        <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px;">
                            <div style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif;">Apple D3DMetal (GPTK) Parameters</div>
                            <div style="background: rgba(56, 189, 248, 0.1); border: 1px solid rgba(56, 189, 248, 0.25); color: #38bdf8; padding: 2px 8px; border-radius: 4px; font-size: 0.68rem; font-family: 'Inter', sans-serif; font-weight: 600;">DIRECTX 11 & 12 → METAL</div>
                        </div>
                        <div style="color: rgba(255,255,255,0.45); font-size: 0.82rem; line-height: 1.4; margin-bottom: 18px;">
                            Apple's DirectX translation layer optimized for Apple Silicon hardware graphics pipelines.
                        </div>

                        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-bottom: 14px;">
                            <div style="background: rgba(0,0,0,0.25); border: 1px solid rgba(255,255,255,0.05); padding: 12px 16px; border-radius: 8px;">
                                <label style="display: block; font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.85rem; color: white; margin-bottom: 6px;">Multi-Thread Synchronization</label>
                                <select id="sync-select" style="width: 100%; background: transparent; color: #38bdf8; border: none; font-size: 0.85rem; font-family: 'Inter', sans-serif; outline: none; cursor: pointer; font-weight: 500;">
                                    <option value="msync" ${runtimeConfig.enhanced_sync === 'msync' ? 'selected' : ''} style="background: #12141f; color: white;">MSync (Apple Silicon Native)</option>
                                    <option value="esync" ${runtimeConfig.enhanced_sync === 'esync' ? 'selected' : ''} style="background: #12141f; color: white;">ESync (Eventfd)</option>
                                    <option value="none" ${runtimeConfig.enhanced_sync === 'none' ? 'selected' : ''} style="background: #12141f; color: white;">Disabled</option>
                                </select>
                            </div>

                            <div style="background: rgba(0,0,0,0.25); border: 1px solid rgba(255,255,255,0.05); padding: 12px 16px; border-radius: 8px;">
                                <label style="display: block; font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.85rem; color: white; margin-bottom: 6px;">Resolution Scaling (FSR)</label>
                                <select id="res-scale-select" style="width: 100%; background: transparent; color: #38bdf8; border: none; font-size: 0.85rem; font-family: 'Inter', sans-serif; outline: none; cursor: pointer; font-weight: 500;">
                                    <option value="1.0" ${runtimeConfig.resolution_scale === '1.0' ? 'selected' : ''} style="background: #12141f; color: white;">100% (Native Resolution)</option>
                                    <option value="0.75" ${runtimeConfig.resolution_scale === '0.75' ? 'selected' : ''} style="background: #12141f; color: white;">75% (FSR Quality Mode)</option>
                                    <option value="0.67" ${runtimeConfig.resolution_scale === '0.67' ? 'selected' : ''} style="background: #12141f; color: white;">67% (FSR Balanced Mode)</option>
                                    <option value="0.5" ${runtimeConfig.resolution_scale === '0.5' ? 'selected' : ''} style="background: #12141f; color: white;">50% (FSR Performance Mode)</option>
                                </select>
                            </div>
                        </div>

                        <div style="display: flex; flex-direction: column; gap: 8px;">
                            <label style="${rowStyle} cursor: pointer;">
                                <div>
                                    <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; color: white; margin-bottom: 2px;">Metal Performance HUD</div>
                                    <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: rgba(255,255,255,0.45);">Display real-time Apple Metal FPS, frametime, and GPU load.</div>
                                </div>
                                <input type="checkbox" id="metal-hud" ${runtimeConfig.metal_hud ? 'checked' : ''} style="width: 16px; height: 16px; accent-color: #38bdf8; cursor: pointer;">
                            </label>

                            <label style="${rowStyle} cursor: pointer;">
                                <div>
                                    <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; color: white; margin-bottom: 2px;">Hardware Raytracing (DXR)</div>
                                    <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: rgba(255,255,255,0.45);">DirectX Raytracing acceleration via Metal RT (M3 / M4 chips).</div>
                                </div>
                                <input type="checkbox" id="dxr-enabled" ${runtimeConfig.dxr_enabled ? 'checked' : ''} style="width: 16px; height: 16px; accent-color: #38bdf8; cursor: pointer;">
                            </label>

                            <label style="${rowStyle} cursor: pointer;">
                                <div>
                                    <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; color: white; margin-bottom: 2px;">High-DPI Retina Mode</div>
                                    <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: #f87171;">Forces native 4K/5K resolution. May impact performance.</div>
                                </div>
                                <input type="checkbox" id="retina-mode" ${runtimeConfig.retina_mode ? 'checked' : ''} style="width: 16px; height: 16px; accent-color: #38bdf8; cursor: pointer;">
                            </label>

                            <label style="${rowStyle} cursor: pointer;">
                                <div>
                                    <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; color: white; margin-bottom: 2px;">Rosetta AVX Extensions</div>
                                    <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: rgba(255,255,255,0.45);">Advertise AVX/AVX2 CPU instructions via Rosetta 2 (macOS 15+).</div>
                                </div>
                                <input type="checkbox" id="avx-enabled" ${runtimeConfig.avx_enabled ? 'checked' : ''} style="width: 16px; height: 16px; accent-color: #38bdf8; cursor: pointer;">
                            </label>
                        </div>
                    </div>
                `;
            } else if (isDXVK) {
                engineSpecificHTML = `
                    <div style="${cardStyle}">
                        <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px;">
                            <div style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif;">DXVK (Vulkan / MoltenVK) Parameters</div>
                            <div style="background: rgba(168, 85, 247, 0.1); border: 1px solid rgba(168, 85, 247, 0.25); color: #c084fc; padding: 2px 8px; border-radius: 4px; font-size: 0.68rem; font-family: 'Inter', sans-serif; font-weight: 600;">DIRECTX 9/10/11 → VULKAN</div>
                        </div>
                        <div style="color: rgba(255,255,255,0.45); font-size: 0.82rem; line-height: 1.4; margin-bottom: 18px;">
                            DirectX 9, 10, and 11 translation via Vulkan and MoltenVK Metal backend.
                        </div>

                        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-bottom: 14px;">
                            <div style="background: rgba(0,0,0,0.25); border: 1px solid rgba(255,255,255,0.05); padding: 12px 16px; border-radius: 8px;">
                                <label style="display: block; font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.85rem; color: white; margin-bottom: 6px;">DXVK HUD Overlay</label>
                                <select id="dxvk-hud-select" style="width: 100%; background: transparent; color: #38bdf8; border: none; font-size: 0.85rem; font-family: 'Inter', sans-serif; outline: none; cursor: pointer; font-weight: 500;">
                                    <option value="off" ${runtimeConfig.dxvk_hud === 'off' ? 'selected' : ''} style="background: #12141f; color: white;">Off</option>
                                    <option value="fps" ${runtimeConfig.dxvk_hud === 'fps' ? 'selected' : ''} style="background: #12141f; color: white;">FPS Only</option>
                                    <option value="frametimes" ${runtimeConfig.dxvk_hud === 'frametimes' ? 'selected' : ''} style="background: #12141f; color: white;">Frametimes + FPS</option>
                                    <option value="full" ${runtimeConfig.dxvk_hud === 'full' || runtimeConfig.dxvk_hud === '1' ? 'selected' : ''} style="background: #12141f; color: white;">Full Diagnostics</option>
                                </select>
                            </div>

                            <div style="background: rgba(0,0,0,0.25); border: 1px solid rgba(255,255,255,0.05); padding: 12px 16px; border-radius: 8px;">
                                <label style="display: block; font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.85rem; color: white; margin-bottom: 6px;">Frame Rate Limiter</label>
                                <input type="number" id="dxvk-fps-cap" value="${runtimeConfig.dxvk_frame_rate || 0}" min="0" max="240" placeholder="0 (Unlimited)" style="width: 100%; background: transparent; color: #38bdf8; border: none; font-size: 0.85rem; font-family: 'Inter', sans-serif; outline: none; font-weight: 500;">
                            </div>
                        </div>

                        <div style="display: flex; flex-direction: column; gap: 8px;">
                            <label style="${rowStyle} cursor: pointer;">
                                <div>
                                    <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; color: white; margin-bottom: 2px;">Async Shader Compilation</div>
                                    <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: rgba(255,255,255,0.45);">Enabled by default. Compiles shaders on background threads to prevent frame freezes. Disable only if you see visual glitches.</div>
                                </div>
                                <input type="checkbox" id="dxvk-async" ${runtimeConfig.dxvk_async !== false ? 'checked' : ''} style="width: 16px; height: 16px; accent-color: #38bdf8; cursor: pointer;">
                            </label>

                            <label style="${rowStyle} cursor: pointer;">
                                <div>
                                    <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; color: white; margin-bottom: 2px;">High-DPI Retina Mode</div>
                                    <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: #f87171;">Forces native 4K/5K resolution. May impact performance.</div>
                                </div>
                                <input type="checkbox" id="retina-mode" ${runtimeConfig.retina_mode ? 'checked' : ''} style="width: 16px; height: 16px; accent-color: #38bdf8; cursor: pointer;">
                            </label>

                            <label style="${rowStyle} cursor: pointer;">
                                <div>
                                    <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; color: white; margin-bottom: 2px;">Rosetta AVX Extensions</div>
                                    <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: rgba(255,255,255,0.45);">Advertise AVX/AVX2 CPU instructions via Rosetta 2 (macOS 15+).</div>
                                </div>
                                <input type="checkbox" id="avx-enabled" ${runtimeConfig.avx_enabled ? 'checked' : ''} style="width: 16px; height: 16px; accent-color: #38bdf8; cursor: pointer;">
                            </label>
                        </div>
                    </div>
                `;
            } else if (isWine3D) {
                engineSpecificHTML = `
                    <div style="${cardStyle}">
                        <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px;">
                            <div style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif;">WineD3D (OpenGL Fallback) Parameters</div>
                            <div style="background: rgba(234, 179, 8, 0.1); border: 1px solid rgba(234, 179, 8, 0.25); color: #facc15; padding: 2px 8px; border-radius: 4px; font-size: 0.68rem; font-family: 'Inter', sans-serif; font-weight: 600;">LEGACY SAFE MODE</div>
                        </div>
                        <div style="color: rgba(255,255,255,0.45); font-size: 0.82rem; line-height: 1.4; margin-bottom: 18px;">
                            Direct3D to OpenGL translation for 32-bit legacy titles and simple 2D/indie games.
                        </div>

                        <div style="display: flex; flex-direction: column; gap: 8px;">
                            <label style="${rowStyle} cursor: pointer;">
                                <div>
                                    <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; color: white; margin-bottom: 2px;">High-DPI Retina Mode</div>
                                    <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: rgba(255,255,255,0.45);">Allow high-resolution rendering.</div>
                                </div>
                                <input type="checkbox" id="retina-mode" ${runtimeConfig.retina_mode ? 'checked' : ''} style="width: 16px; height: 16px; accent-color: #38bdf8; cursor: pointer;">
                            </label>

                            <label style="${rowStyle} cursor: pointer;">
                                <div>
                                    <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; color: white; margin-bottom: 2px;">Rosetta AVX Extensions</div>
                                    <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: rgba(255,255,255,0.45);">Advertise AVX/AVX2 CPU instructions via Rosetta 2 (macOS 15+).</div>
                                </div>
                                <input type="checkbox" id="avx-enabled" ${runtimeConfig.avx_enabled ? 'checked' : ''} style="width: 16px; height: 16px; accent-color: #38bdf8; cursor: pointer;">
                            </label>
                        </div>
                    </div>
                `;
            }

            container.innerHTML = `
                <div class="settings-content-wrapper" style="max-width: 880px; margin: 0 auto; padding: 36px 32px 80px;">
                    <!-- Header -->
                    <div style="border-bottom: 1px solid rgba(255, 255, 255, 0.06); padding-bottom: 24px; margin-bottom: 32px;">
                        <h1 style="font-family: 'Plus Jakarta Sans', sans-serif; font-weight: 800; font-size: 2rem; letter-spacing: -0.5px; color: white; margin: 0 0 6px 0;">Configuration</h1>
                        <p style="font-family: 'Inter', sans-serif; font-size: 0.88rem; color: rgba(255, 255, 255, 0.45); margin: 0;">Per-game graphics engine, executable selection, and translation parameters.</p>
                    </div>

                    <!-- Target Game Selector -->
                    <div style="${cardStyle}">
                        <div style="margin-bottom: 16px;">
                            <div style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif; margin-bottom: 4px;">Target Game Capsule</div>
                            <div style="color: rgba(255, 255, 255, 0.45); font-size: 0.82rem; line-height: 1.4;">Select which installed game's translation capsule to configure.</div>
                        </div>

                        <div style="${rowStyle}">
                            <select id="config-game-select" style="flex: 1; background: transparent; color: #38bdf8; border: none; font-size: 0.92rem; font-family: 'Inter', sans-serif; outline: none; cursor: pointer; font-weight: 600;">
                                ${gamesOptions}
                            </select>
                        </div>
                    </div>

                    <!-- Executable Selection -->
                    ${exeSelectorHTML}

                    <!-- Graphics Engine & AI Override Controls -->
                    <div style="${cardStyle}">
                        <div style="margin-bottom: 16px;">
                            <div style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif; margin-bottom: 4px;">Translation Engine</div>
                            <div style="color: rgba(255, 255, 255, 0.45); font-size: 0.82rem; line-height: 1.4;">
                                Choose the translation engine. Enable Manual Override to lock your custom choices against automatic cloud profile syncing.
                            </div>
                        </div>

                        <div style="display: flex; flex-direction: column; gap: 10px;">
                            <label style="${rowStyle} cursor: pointer; background: ${isManualOverride ? 'rgba(56, 189, 248, 0.06)' : 'rgba(0,0,0,0.25)'}; border-color: ${isManualOverride ? 'rgba(56, 189, 248, 0.25)' : 'rgba(255,255,255,0.05)'};">
                                <div>
                                    <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; color: white; margin-bottom: 2px;">Manual Override Lock</div>
                                    <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: ${isManualOverride ? '#38bdf8' : 'rgba(255,255,255,0.45)'};">
                                        ${isManualOverride ? '🔒 Custom settings are locked. RIFT Cloud will not override your choices.' : 'Enable to manually customize and lock engine settings below.'}
                                    </div>
                                </div>
                                <input type="checkbox" id="manual-override" ${isManualOverride ? 'checked' : ''} style="width: 16px; height: 16px; accent-color: #38bdf8; cursor: pointer;">
                            </label>

                            <div style="${rowStyle} ${!isManualOverride ? 'opacity: 0.45; pointer-events: none;' : ''}">
                                <div>
                                    <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; color: white; margin-bottom: 2px;">Graphics Translation Layer</div>
                                    <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: rgba(255,255,255,0.45);">Target API translation for DirectX draw calls.</div>
                                </div>
                                <select id="engine-select" style="background: rgba(255,255,255,0.05); color: #38bdf8; border: 1px solid rgba(255,255,255,0.1); padding: 8px 14px; border-radius: 6px; font-size: 0.82rem; font-family: 'Inter', sans-serif; outline: none; cursor: pointer; font-weight: 600;">
                                    <option value="d3dmetal" ${engine === 'd3dmetal' || engine === 'gptk' ? 'selected' : ''} style="background: #12141f; color: white;">Apple D3DMetal (DX11/12 → Metal)</option>
                                    <option value="dxvk" ${engine === 'dxvk' ? 'selected' : ''} style="background: #12141f; color: white;">DXVK (DX9/10/11 → Vulkan)</option>
                                    <option value="wine3d" ${isWine3D ? 'selected' : ''} style="background: #12141f; color: white;">WineD3D (Legacy OpenGL Safe Mode)</option>
                                </select>
                            </div>
                        </div>
                    </div>

                    <!-- Dynamic Engine Settings Zone -->
                    <div id="engine-settings-zone" style="${!isManualOverride ? 'opacity: 0.45; pointer-events: none;' : ''}">
                        ${engineSpecificHTML}
                    </div>

                    <!-- Launch Arguments -->
                    <div style="${cardStyle} ${!isManualOverride ? 'opacity: 0.45; pointer-events: none;' : ''}">
                        <div style="margin-bottom: 16px;">
                            <div style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif; margin-bottom: 4px;">Launch Arguments</div>
                            <div style="color: rgba(255, 255, 255, 0.45); font-size: 0.82rem; line-height: 1.4;">
                                Pass custom CLI flags directly to the Windows binary at startup.
                            </div>
                        </div>

                        <div style="${rowStyle}">
                            <input type="text" id="launch-args" value="${(runtimeConfig.launch_args || []).join(' ')}" placeholder="-force-d3d12 -windowed -nolog" style="flex: 1; background: transparent; color: #38bdf8; border: none; font-size: 0.85rem; font-family: 'JetBrains Mono', monospace; outline: none;">
                        </div>
                    </div>

                    <!-- Save Action -->
                    <div style="display: flex; justify-content: flex-end; gap: 12px; align-items: center; margin-top: 8px;">
                        <button id="save-config-btn" style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.88rem; padding: 12px 28px; border-radius: 8px; background: #38bdf8; color: #000; border: none; cursor: pointer; transition: all 0.2s ease; box-shadow: 0 0 20px rgba(56, 189, 248, 0.25);">
                            Commit Configuration
                        </button>
                    </div>
                </div>
            `;

            // --- Event Listeners ---

            // Game selector → reload config
            container.querySelector('#config-game-select')?.addEventListener('change', async (e) => {
                selectedGameId = e.target.value;
                runtimeConfig = await window.go.rift.App.GetGameRuntimeConfig(selectedGameId);
                await renderForm();
            });

            // Executable radio selector → switch active binary
            container.querySelectorAll('input[name="game-exe-choice"]').forEach(radio => {
                radio.addEventListener('change', async (e) => {
                    const newExe = e.target.value;
                    try {
                        await window.go.rift.App.SetGameExecutable(selectedGameId, newExe);
                        runtimeConfig = await window.go.rift.App.GetGameRuntimeConfig(selectedGameId);
                        await renderForm();
                    } catch (err) {
                        console.error("Failed to set executable:", err);
                    }
                });
            });

            // Manual override toggle → re-render to update interactivity
            container.querySelector('#manual-override')?.addEventListener('change', async () => {
                runtimeConfig.manual_override = container.querySelector('#manual-override').checked;
                await renderForm();
            });

            // Engine selector → re-render to dynamically swap engine-specific cards
            const engineSelect = container.querySelector('#engine-select');
            if (engineSelect) {
                engineSelect.addEventListener('change', async () => {
                    runtimeConfig.engine = engineSelect.value;
                    await renderForm();
                });
            }

            // Save button
            container.querySelector('#save-config-btn')?.addEventListener('click', async () => {
                const btn = container.querySelector('#save-config-btn');
                btn.textContent = "Saving Configuration...";
                btn.disabled = true;

                const argsString = container.querySelector('#launch-args')?.value.trim() || '';
                const argsArray = argsString ? argsString.split(/\s+/) : [];

                runtimeConfig.manual_override = container.querySelector('#manual-override')?.checked || false;
                runtimeConfig.engine = container.querySelector('#engine-select')?.value || runtimeConfig.engine;
                runtimeConfig.launch_args = argsArray;

                // D3DMetal specific
                if (isD3DMetal) {
                    runtimeConfig.enhanced_sync = container.querySelector('#sync-select')?.value || 'msync';
                    runtimeConfig.resolution_scale = container.querySelector('#res-scale-select')?.value || '1.0';
                    runtimeConfig.metal_hud = container.querySelector('#metal-hud')?.checked || false;
                    runtimeConfig.dxr_enabled = container.querySelector('#dxr-enabled')?.checked || false;
                    runtimeConfig.avx_enabled = container.querySelector('#avx-enabled')?.checked || false;
                    runtimeConfig.retina_mode = container.querySelector('#retina-mode')?.checked || false;
                    // Clear DXVK-only fields
                    runtimeConfig.dxvk_hud = 'off';
                    runtimeConfig.dxvk_async = false;
                    runtimeConfig.dxvk_frame_rate = 0;
                }

                // DXVK specific
                if (isDXVK) {
                    runtimeConfig.dxvk_hud = container.querySelector('#dxvk-hud-select')?.value || 'off';
                    runtimeConfig.dxvk_async = container.querySelector('#dxvk-async')?.checked || false;
                    runtimeConfig.dxvk_frame_rate = parseInt(container.querySelector('#dxvk-fps-cap')?.value || '0', 10);
                    runtimeConfig.avx_enabled = container.querySelector('#avx-enabled')?.checked || false;
                    runtimeConfig.retina_mode = container.querySelector('#retina-mode')?.checked || false;
                    // Clear D3DMetal-only fields
                    runtimeConfig.metal_hud = false;
                    runtimeConfig.dxr_enabled = false;
                    runtimeConfig.resolution_scale = '1.0';
                }

                // Wine3D specific
                if (isWine3D) {
                    runtimeConfig.retina_mode = container.querySelector('#retina-mode')?.checked || false;
                    runtimeConfig.avx_enabled = container.querySelector('#avx-enabled')?.checked || false;
                    runtimeConfig.metal_hud = false;
                    runtimeConfig.dxr_enabled = false;
                    runtimeConfig.dxvk_hud = 'off';
                    runtimeConfig.dxvk_async = false;
                }

                try {
                    await window.go.rift.App.SaveGameRuntimeConfig(selectedGameId, runtimeConfig);
                    btn.textContent = "✓ Configuration Committed";
                    btn.style.background = "#22c55e";
                    setTimeout(() => {
                        btn.textContent = "Commit Configuration";
                        btn.style.background = "#38bdf8";
                        btn.disabled = false;
                    }, 2500);
                } catch (e) {
                    console.error("Save error:", e);
                    btn.textContent = "Error Saving";
                    btn.style.background = "#ef4444";
                    btn.disabled = false;
                }
            });
        };

        renderForm();
    }

    init();
    return container;
}
