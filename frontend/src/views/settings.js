import { store } from '../state.js';
import { appendLog } from '../main.js';

export function renderSettings() {
    const container = document.createElement('div');
    container.className = 'view-premium';
    container.style.width = '100%';
    container.style.height = '100%';
    container.style.overflowY = 'auto';

    async function init() {
        const cfg = await window.go.rift.App.GetFullConfig();
        const authStatus = await window.go.rift.App.GetAuthStatus();
        const steamUser = authStatus.steamUsername;
        const epicUser = authStatus.epicUsername;

        let userEmail = 'Connected';
        try {
            const userStr = await window.go.rift.App.SupabaseGetUser();
            const userObj = JSON.parse(userStr || '{}');
            if (userObj.email) userEmail = userObj.email;
        } catch (_) {}

        let buildInfo = { version: '1.0.0-beta.2', stage: 'Beta 2', platform: 'darwin', architecture: 'arm64' };
        try {
            if (window.go?.rift?.App?.GetAppBuildInfo) {
                buildInfo = await window.go.rift.App.GetAppBuildInfo();
            }
        } catch (_) {}

        const cardStyle = `
            background: rgba(255, 255, 255, 0.02);
            border: 1px solid rgba(255, 255, 255, 0.06);
            border-radius: 12px;
            padding: 24px 28px;
            margin-bottom: 24px;
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

        container.innerHTML = `
            <div class="settings-content-wrapper" style="max-width: 880px; margin: 0 auto; padding: 36px 32px 80px;">
                <!-- Header -->
                <div style="border-bottom: 1px solid rgba(255, 255, 255, 0.06); padding-bottom: 24px; margin-bottom: 32px;">
                    <h1 style="font-family: 'Plus Jakarta Sans', sans-serif; font-weight: 800; font-size: 2rem; letter-spacing: -0.5px; color: white; margin: 0 0 6px 0;">Global Settings</h1>
                    <p style="font-family: 'Inter', sans-serif; font-size: 0.88rem; color: rgba(255, 255, 255, 0.45); margin: 0;">Manage application preferences, storage directories, and connected accounts.</p>
                </div>

                <!-- Section 0: Client Version & Updates -->
                <div style="${cardStyle}">
                    <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 18px;">
                        <div>
                            <div style="display: flex; align-items: center; gap: 10px; margin-bottom: 4px;">
                                <span style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif;">RIFT Client</span>
                                <span style="background: rgba(56, 189, 248, 0.15); color: #38bdf8; border: 1px solid rgba(56, 189, 248, 0.3); font-size: 0.72rem; font-family: 'JetBrains Mono', monospace; font-weight: 700; padding: 2px 8px; border-radius: 999px;">
                                    ${buildInfo.stage || 'Beta 2'}
                                </span>
                            </div>
                            <div style="color: rgba(255, 255, 255, 0.45); font-size: 0.82rem; line-height: 1.4;">Native Windows gaming translation runtime for Apple Silicon.</div>
                        </div>
                        <button id="settings-check-update-btn" style="font-family: 'Inter', sans-serif; font-size: 0.82rem; font-weight: 600; padding: 8px 18px; border-radius: 6px; background: rgba(56, 189, 248, 0.15); color: #38bdf8; border: 1px solid rgba(56, 189, 248, 0.3); cursor: pointer; transition: all 0.2s; display: flex; align-items: center; gap: 6px;">
                            <span id="update-btn-label">Check for Updates</span>
                        </button>
                    </div>

                    <div style="${rowStyle}">
                        <div>
                            <div style="color: white; font-weight: 600; font-size: 0.9rem; font-family: 'Inter', sans-serif;">Installed Version</div>
                            <div style="color: rgba(255, 255, 255, 0.5); font-size: 0.8rem; font-family: 'JetBrains Mono', monospace; margin-top: 2px;">
                                v${buildInfo.version || '1.0.0-beta.2'} (${buildInfo.platform || 'darwin'}-${buildInfo.architecture || 'arm64'})
                            </div>
                        </div>
                        <div id="settings-update-status" style="font-family: 'Inter', sans-serif; font-size: 0.82rem; color: rgba(255, 255, 255, 0.5);">
                            Up to date
                        </div>
                    </div>
                </div>

                <!-- Section 1: Account & Security -->
                <div style="${cardStyle}">
                    <div style="margin-bottom: 18px;">
                        <div style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif; margin-bottom: 4px;">RIFT Cloud Account</div>
                        <div style="color: rgba(255, 255, 255, 0.45); font-size: 0.82rem; line-height: 1.4;">Manage your archive profile and master authentication credentials.</div>
                    </div>

                    <!-- Cloud Session Row -->
                    <div style="${rowStyle} margin-bottom: 12px;">
                        <div>
                            <div style="color: white; font-weight: 600; font-size: 0.9rem; font-family: 'Inter', sans-serif;">Active Session</div>
                            <div style="color: rgba(255, 255, 255, 0.5); font-size: 0.8rem; font-family: 'Inter', sans-serif; margin-top: 2px;">Logged in as <span style="color: #38bdf8; font-weight: 600;">${userEmail}</span></div>
                        </div>
                        <button id="settings-logout-btn" style="font-family: 'Inter', sans-serif; font-size: 0.8rem; font-weight: 600; padding: 8px 18px; border-radius: 6px; background: rgba(239, 68, 68, 0.12); color: #f87171; border: 1px solid rgba(239, 68, 68, 0.25); cursor: pointer; transition: all 0.2s;">
                            Sign Out
                        </button>
                    </div>

                    <!-- Inline Password Update -->
                    <div style="background: rgba(0, 0, 0, 0.2); border: 1px solid rgba(255, 255, 255, 0.05); border-radius: 8px; padding: 16px 20px;">
                        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;">
                            <div style="color: rgba(255, 255, 255, 0.8); font-weight: 600; font-size: 0.85rem; font-family: 'Inter', sans-serif;">Update Master Password</div>
                            <div id="settings-pass-status" style="font-size: 0.8rem; font-family: 'Inter', sans-serif;"></div>
                        </div>
                        <div style="display: flex; gap: 10px; align-items: center;">
                            <input type="password" id="settings-new-pass" placeholder="New Password (min 6 chars)..." style="flex: 1; background: rgba(255, 255, 255, 0.04); border: 1px solid rgba(255, 255, 255, 0.1); padding: 10px 14px; border-radius: 6px; color: #fff; font-family: 'Inter', sans-serif; font-size: 0.85rem; outline: none; box-sizing: border-box; transition: border-color 0.2s;">
                            <input type="password" id="settings-confirm-pass" placeholder="Confirm Password..." style="flex: 1; background: rgba(255, 255, 255, 0.04); border: 1px solid rgba(255, 255, 255, 0.1); padding: 10px 14px; border-radius: 6px; color: #fff; font-family: 'Inter', sans-serif; font-size: 0.85rem; outline: none; box-sizing: border-box; transition: border-color 0.2s;">
                            <button id="settings-save-pass-btn" style="font-family: 'Inter', sans-serif; font-size: 0.82rem; font-weight: 600; padding: 10px 20px; border-radius: 6px; white-space: nowrap; background: #38bdf8; color: #000; border: none; cursor: pointer; transition: all 0.2s;">
                                Save Password
                            </button>
                        </div>
                    </div>
                </div>

                <!-- Section 2: Connected Storefronts -->
                <div style="${cardStyle}">
                    <div style="margin-bottom: 18px;">
                        <div style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif; margin-bottom: 4px;">Connected Storefronts</div>
                        <div style="color: rgba(255, 255, 255, 0.45); font-size: 0.82rem; line-height: 1.4;">Manage authentication credentials for Steam and Epic Games libraries.</div>
                    </div>

                    <!-- Steam -->
                    <div style="${rowStyle} margin-bottom: 10px;">
                        <div>
                            <div style="color: white; font-weight: 600; font-size: 0.9rem; font-family: 'Inter', sans-serif;">Steam</div>
                            <div style="color: rgba(255, 255, 255, 0.5); font-size: 0.8rem; font-family: 'Inter', sans-serif; margin-top: 2px;">
                                ${steamUser ? `Connected as <span style="color: #38bdf8; font-weight: 600;">${steamUser}</span>` : 'Not Connected'}
                            </div>
                        </div>
                        ${steamUser ? '<button id="settings-logout-steam-btn" style="font-family: Inter, sans-serif; font-size: 0.8rem; font-weight: 500; padding: 7px 16px; border-radius: 6px; background: rgba(255,255,255,0.06); color: rgba(255,255,255,0.8); border: 1px solid rgba(255,255,255,0.1); cursor: pointer; transition: all 0.2s;">Disconnect</button>' : ''}
                    </div>

                    <!-- Epic Games -->
                    <div style="${rowStyle}">
                        <div>
                            <div style="color: white; font-weight: 600; font-size: 0.9rem; font-family: 'Inter', sans-serif;">Epic Games</div>
                            <div style="color: rgba(255, 255, 255, 0.5); font-size: 0.8rem; font-family: 'Inter', sans-serif; margin-top: 2px;">
                                ${epicUser ? `Connected as <span style="color: #38bdf8; font-weight: 600;">${epicUser}</span>` : 'Not Connected'}
                            </div>
                        </div>
                        ${epicUser ? '<button id="settings-logout-epic-btn" style="font-family: Inter, sans-serif; font-size: 0.8rem; font-weight: 500; padding: 7px 16px; border-radius: 6px; background: rgba(255,255,255,0.06); color: rgba(255,255,255,0.8); border: 1px solid rgba(255,255,255,0.1); cursor: pointer; transition: all 0.2s;">Disconnect</button>' : ''}
                    </div>
                </div>

                <!-- Section 3: Storage & Library Locations -->
                <div style="${cardStyle}">
                    <div style="margin-bottom: 18px;">
                        <div style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif; margin-bottom: 4px;">Storage Locations</div>
                        <div style="color: rgba(255, 255, 255, 0.45); font-size: 0.82rem; line-height: 1.4;">Install paths where RIFT provisions Wine translation capsules and downloads games.</div>
                    </div>

                    <div style="${rowStyle}">
                        <div style="font-family: 'JetBrains Mono', monospace; font-size: 0.8rem; color: #38bdf8; word-break: break-all; flex: 1;" id="settings-dir-path">
                            ${cfg.library_folders ? cfg.library_folders.map((folder, index) => 
                                `<div style="margin-bottom: ${index === cfg.library_folders.length - 1 ? '0' : '6px'}; display: flex; align-items: center; gap: 8px;">
                                    <div style="background: ${index === cfg.default_library ? 'rgba(56, 189, 248, 0.15)' : 'rgba(255,255,255,0.05)'}; 
                                                color: ${index === cfg.default_library ? '#38bdf8' : 'rgba(255,255,255,0.5)'}; 
                                                padding: 2px 6px; border-radius: 4px; font-size: 0.65rem; font-family: 'Inter', sans-serif; font-weight: 600;">
                                        ${index === cfg.default_library ? 'DEFAULT' : 'LIBRARY'}
                                    </div>
                                    <span style="color: rgba(255,255,255,0.85);">${folder}</span>
                                </div>`
                            ).join('') : 'Default Directory (~/.rift/games)'}
                        </div>
                        <button id="settings-change-dir-btn" style="font-family: 'Inter', sans-serif; font-size: 0.8rem; font-weight: 600; padding: 8px 16px; border-radius: 6px; background: rgba(255,255,255,0.08); color: white; border: 1px solid rgba(255,255,255,0.12); white-space: nowrap; cursor: pointer; transition: all 0.2s;">
                            Add Location
                        </button>
                    </div>
                </div>


                <!-- Section 4: Telemetry & Privacy -->
                <div style="${cardStyle}">
                    <div style="margin-bottom: 18px;">
                        <div style="color: white; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif; margin-bottom: 4px;">Telemetry & Privacy</div>
                        <div style="color: rgba(255, 255, 255, 0.45); font-size: 0.82rem; line-height: 1.4;">Configure anonymous hardware statistics and diagnostic crash reports.</div>
                    </div>

                    <label style="${rowStyle} cursor: pointer;">
                        <div>
                            <div style="font-family: 'Inter', sans-serif; font-weight: 600; font-size: 0.9rem; color: white; margin-bottom: 2px;">Anonymous Performance Telemetry</div>
                            <div style="font-family: 'Inter', sans-serif; font-size: 0.78rem; color: rgba(255,255,255,0.45);">Help optimize automated community game profiles by sharing non-identifiable FPS and hardware stats.</div>
                        </div>
                        <input type="checkbox" id="toggle-telemetry" ${!cfg.disable_telemetry ? 'checked' : ''} style="width: 17px; height: 17px; accent-color: #38bdf8; cursor: pointer;">
                    </label>
                </div>

                <!-- Section 5: Maintenance & Recovery -->
                <div style="background: rgba(239, 68, 68, 0.03); border: 1px solid rgba(239, 68, 68, 0.15); border-radius: 12px; padding: 24px 28px; margin-bottom: 24px;">
                    <div style="margin-bottom: 16px;">
                        <div style="color: #f87171; font-weight: 700; font-size: 1.05rem; font-family: 'Plus Jakarta Sans', sans-serif; margin-bottom: 4px;">System Recovery</div>
                        <div style="color: rgba(255, 255, 255, 0.45); font-size: 0.82rem; line-height: 1.4;">Forcefully terminate all active background Wine servers, translations, and orphaned subprocesses.</div>
                    </div>

                    <div style="display: flex; justify-content: space-between; align-items: center; background: rgba(0, 0, 0, 0.3); border: 1px solid rgba(239, 68, 68, 0.12); padding: 14px 20px; border-radius: 8px;">
                        <div>
                            <div style="color: white; font-weight: 600; font-size: 0.9rem; font-family: 'Inter', sans-serif;">Terminate All Game Environments</div>
                            <div style="color: rgba(255, 255, 255, 0.45); font-size: 0.78rem; font-family: 'Inter', sans-serif; margin-top: 2px;">Use if a game freezes or fails to release system GPU/RAM resources.</div>
                        </div>
                        <button id="settings-force-kill-all-btn" style="font-family: 'Inter', sans-serif; font-size: 0.8rem; font-weight: 600; padding: 8px 18px; border-radius: 6px; white-space: nowrap; background: #ef4444; color: white; border: none; cursor: pointer; transition: all 0.2s;">
                            FORCE TERMINATE
                        </button>
                    </div>
                </div>
            </div>
        `;

        // === Event Handlers ===

        const checkUpdateBtn = container.querySelector('#settings-check-update-btn');
        const updateStatus = container.querySelector('#settings-update-status');
        const updateBtnLabel = container.querySelector('#update-btn-label');

        if (checkUpdateBtn) {
            checkUpdateBtn.addEventListener('click', async () => {
                checkUpdateBtn.disabled = true;
                checkUpdateBtn.style.opacity = '0.6';
                if (updateBtnLabel) updateBtnLabel.textContent = 'Checking...';
                if (updateStatus) {
                    updateStatus.style.color = '#38bdf8';
                    updateStatus.textContent = 'Checking GitHub releases...';
                }
                try {
                    let update = null;
                    if (window.checkForAppUpdates) {
                        update = await window.checkForAppUpdates(true);
                    } else if (window.go?.rift?.App?.CheckForUpdates) {
                        update = await window.go.rift.App.CheckForUpdates();
                        if (update && update.updateAvailable && window.showUpdateModal) {
                            window.showUpdateModal(update);
                        }
                    }
                    if (update && update.updateAvailable) {
                        if (updateStatus) {
                            updateStatus.style.color = '#f59e0b';
                            updateStatus.textContent = `Update available: v${update.latestVersion}`;
                        }
                    } else {
                        if (updateStatus) {
                            updateStatus.style.color = '#4ade80';
                            updateStatus.textContent = `✓ Up to date (v${update?.currentVersion || buildInfo.version || '1.0.0-beta.2'})`;
                        }
                    }
                } catch (err) {
                    if (updateStatus) {
                        updateStatus.style.color = '#f87171';
                        updateStatus.textContent = 'Check failed (offline or rate-limited)';
                    }
                } finally {
                    checkUpdateBtn.disabled = false;
                    checkUpdateBtn.style.opacity = '1';
                    if (updateBtnLabel) updateBtnLabel.textContent = 'Check for Updates';
                }
            });
        }

        container.querySelector('#settings-change-dir-btn')?.addEventListener('click', async () => {
            appendLog('[Config] Selecting RIFT Games Directory...', 'log-system');
            try {
                const newPath = await window.go.rift.App.SelectLibraryFolder();
                if (newPath) {
                    appendLog(`[Config] RIFT Library Directory added: ${newPath}`, 'log-success');
                    init();
                }
            } catch (err) {
                appendLog(`[Config] Directory selection cancelled or failed: ${err}`, 'log-error');
            }
        });

        const newPassInput = container.querySelector('#settings-new-pass');
        const confirmPassInput = container.querySelector('#settings-confirm-pass');
        const savePassBtn = container.querySelector('#settings-save-pass-btn');
        const statusMsg = container.querySelector('#settings-pass-status');

        if (savePassBtn) {
            savePassBtn.addEventListener('click', async () => {
                const pass1 = newPassInput.value;
                const pass2 = confirmPassInput.value;

                if (!pass1 || pass1.length < 6) {
                    statusMsg.style.color = '#f87171';
                    statusMsg.textContent = 'Password must be at least 6 characters.';
                    return;
                }
                if (pass1 !== pass2) {
                    statusMsg.style.color = '#f87171';
                    statusMsg.textContent = 'Passwords do not match.';
                    return;
                }

                savePassBtn.textContent = 'Saving...';
                savePassBtn.disabled = true;

                try {
                    const resStr = await window.go.rift.App.SupabaseUpdatePassword(pass1);
                    const res = JSON.parse(resStr);
                    if (res.error) throw new Error(res.error);

                    statusMsg.style.color = '#4ade80';
                    statusMsg.textContent = '✓ Password successfully updated!';
                    newPassInput.value = '';
                    confirmPassInput.value = '';
                    savePassBtn.textContent = 'Saved';
                    setTimeout(() => {
                        savePassBtn.textContent = 'Save Password';
                        savePassBtn.disabled = false;
                        statusMsg.textContent = '';
                    }, 3500);
                } catch (err) {
                    statusMsg.style.color = '#f87171';
                    statusMsg.textContent = 'Update failed: ' + err.message;
                    savePassBtn.textContent = 'Save Password';
                    savePassBtn.disabled = false;
                }
            });
        }

        const logoutBtn = container.querySelector('#settings-logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', async () => {
                await window.go.rift.App.SupabaseSignOut();
                window.localStorage.removeItem('hasLoggedIn');
                window.location.hash = '/login';
            });
        }
        
        const logoutSteamBtn = container.querySelector('#settings-logout-steam-btn');
        if (logoutSteamBtn) {
            logoutSteamBtn.addEventListener('click', async () => {
                appendLog('[Auth] Purging Steam session data...', 'log-system');
                await window.go.rift.App.LogoutSteam();
                await window.go.rift.App.SyncLibrary();
                init();
            });
        }
        
        const logoutEpicBtn = container.querySelector('#settings-logout-epic-btn');
        if (logoutEpicBtn) {
            logoutEpicBtn.addEventListener('click', async () => {
                appendLog('[Auth] Purging Epic session data...', 'log-system');
                await window.go.rift.App.LogoutEpic();
                await window.go.rift.App.SyncLibrary();
                init();
            });
        }

        const forceKillBtn = container.querySelector('#settings-force-kill-all-btn');
        if (forceKillBtn) {
            forceKillBtn.addEventListener('click', async () => {
                appendLog('[System Recovery] Sending global termination signal...', 'log-warn');
                forceKillBtn.innerText = 'TERMINATING...';
                try {
                    const res = await window.go.rift.App.ForceQuitAll();
                    appendLog(`[System Recovery] ${res}`, 'log-success');
                    
                    const updatedGames = await window.go.rift.App.SyncLibrary();
                    store.update('games', updatedGames);
                } catch(err) {
                    appendLog(`[System Recovery] Force quit failed: ${err}`, 'log-error');
                }
                setTimeout(() => { forceKillBtn.innerText = 'FORCE TERMINATE'; }, 2000);
            });
        }

        container.querySelector('#toggle-telemetry')?.addEventListener('change', async (e) => {
            cfg.disable_telemetry = !e.target.checked;
            await window.go.rift.App.SaveFullConfig(cfg);
            appendLog(`[Config] Telemetry preference updated: ${!e.target.checked}`, 'log-info');
        });
    }

    init();
    return container;
}
