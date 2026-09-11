import './style/index.css';
import { initRouter, invalidateViewCache } from './router.js';
import { store } from './state.js';
import { showEpicSelectiveDownloadModal } from './utils/interactive_modal.js';
import { showStartupSplash } from './utils/splash.js';

// ============================================
// QUICK LAUNCH SIDEBAR
// ============================================
function initQuickLaunch() {
    const listContainer = document.getElementById('quick-launch-list');
    if (!listContainer) return;

    store.subscribe('games', (games) => {
        if (!games) return;
        
        // Filter: only show downloaded / ready / running games
        const installedGames = games.filter(g => g.isInstalled || g.status === 'ready' || g.status === 'running');
        
        // Sort: running game always goes to the top, then sort by lastPlayedTs
        const sortedGames = [...installedGames].sort((a, b) => {
            if (a.status === 'running' && b.status !== 'running') return -1;
            if (b.status === 'running' && a.status !== 'running') return 1;
            
            // Sort by most recently played (highest timestamp first)
            const aTs = a.lastPlayedTs || 0;
            const bTs = b.lastPlayedTs || 0;
            return bTs - aTs;
        });

        const recentGames = sortedGames.slice(0, 2);
        
        if (recentGames.length === 0) {
            listContainer.innerHTML = `
                <div style="font-family: 'Inter', sans-serif; font-size: 0.72rem; color: rgba(255,255,255,0.25); padding: 12px 14px; font-style: italic; border: 1px dashed rgba(255,255,255,0.05); border-radius: 6px; text-align: center; letter-spacing: 0.5px;">
                    No Runtimes Prepared
                </div>
            `;
            return;
        }

        const fallbackCover = "data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='120' height='180' viewBox='0 0 120 180'><rect width='100%' height='100%' fill='%231b2838'/><text x='50%' y='50%' dominant-baseline='middle' text-anchor='middle' fill='%232a475e' font-family='sans-serif' font-size='12'>RIFT</text></svg>";
        listContainer.innerHTML = recentGames.map(game => `
            <div class="quick-launch-item" data-id="${game.id}">
                <div class="quick-launch-left">
                    <img class="quick-launch-icon" src="${game.cover || fallbackCover}" alt="${game.name}" onerror="this.src='${fallbackCover}'; this.onerror=null;">
                    <span class="quick-launch-name">${game.name}</span>
                </div>
                <span class="quick-launch-play" title="Launch Game">${game.status === 'running' ? '■' : '▶'}</span>
            </div>
        `).join('');

        // Wire up clicks
        listContainer.querySelectorAll('.quick-launch-item').forEach(el => {
            const gameId = el.dataset.id;
            const game = games.find(g => g.id === gameId);

            el.addEventListener('click', async (e) => {
                if (e.target.classList.contains('quick-launch-play')) {
                    e.stopPropagation();
                    if (!game.isInstalled && game.status !== 'ready' && game.status !== 'running') {
                        appendLog(`[RIFT Engine] '${game.name}' is not fully installed on disk. Redirecting to Library to resume download...`, 'log-warn');
                        window.location.hash = '/library';
                        return;
                    }
                    appendLog(`[RIFT Engine] Launching '${game.name}' via Quick Launch...`, 'log-system');
                    
                    // Update state to running
                    const updatedGames = games.map(g => {
                        if (g.id === gameId) {
                            return { ...g, status: 'running', lastPlayed: 'Playing now' };
                        } else if (g.status === 'running') {
                            return { ...g, status: 'ready', lastPlayed: 'Just now' };
                        }
                        return g;
                    });
                    store.update('games', updatedGames);

                    try {
                        await window.go.rift.App.ExecuteTranslation(game.id, game.platform, game.appID);
                    } catch (err) {
                        appendLog(`[RIFT Engine] Launch failed: ${err}`, 'log-error');
                        const resetGames = games.map(g => {
                            if (g.id === gameId) return { ...g, status: 'ready' };
                            return g;
                        });
                        store.update('games', resetGames);
                    }
                } else {
                    window.location.hash = '/game/' + gameId;
                }
            });
        });
    });
}

// ============================================
// NETWORK MONITORING
// ============================================
function initNetworkMonitor() {
    const updateNetworkStatus = (isOnline) => {
        store.update('offlineMode', !isOnline);
        const iconContainer = document.getElementById('sys-network-status');
        if (!iconContainer) return;
        
        const path = iconContainer.querySelector('path');
        if (isOnline) {
            iconContainer.style.color = 'var(--text-primary)';
            if (path) path.setAttribute('d', 'M5 12.55a11 11 0 0 1 14.08 0M1.42 9a16 16 0 0 1 21.16 0M8.53 16.11a6 6 0 0 1 6.95 0M12 20h.01');
            appendLog('[System] Network connected. Online mode active.', 'log-success');
        } else {
            iconContainer.style.color = 'var(--text-tertiary)';
            // Slashed wifi icon
            if (path) path.setAttribute('d', 'M1 1l22 22M16.72 11.06A10.94 10.94 0 0 1 19 12.55M5 12.55a10.94 10.94 0 0 1 5.17-2.39M10.71 5.05A16 16 0 0 1 22.58 9M1.42 9a15.91 15.91 0 0 1 4.7-2.88M8.53 16.11a6 6 0 0 1 6.95 0M12 20h.01');
            appendLog('[System] Network disconnected.', 'log-warn');
            
            // Show the popup the user asked for
            if (window.showRiftNotification) {
                window.showRiftNotification(
                    'Network Disconnected',
                    'RIFT has lost connection to the internet. Booting into Offline Mode. You can continue to launch already downloaded games.',
                    'warn',
                    () => { appendLog('[System] User confirmed Offline Mode.', 'log-system'); },
                    'ENTER OFFLINE MODE'
                );
            }
        }
    };

    window.addEventListener('online', () => updateNetworkStatus(true));
    window.addEventListener('offline', () => updateNetworkStatus(false));
    
    // Initial check
    setTimeout(() => updateNetworkStatus(navigator.onLine), 500);
}

// ============================================
// GLOBAL SEARCH BAR
// ============================================
function initCommandBar() {
    const globalInput = document.getElementById('global-search-input');
    if (!globalInput) return;

    document.addEventListener('keydown', (e) => {
        if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
            e.preventDefault();
            globalInput.focus();
        }
        if (e.key === 'Escape' && document.activeElement === globalInput) {
            globalInput.blur();
        }
    });

    globalInput.addEventListener('input', (e) => {
        const query = e.target.value;
        window.dispatchEvent(new CustomEvent('global-search', { detail: query }));
    });
}

// ============================================
// CONSOLE & LOGGING
// ============================================
function initConsole() {
    const panel = document.getElementById('console-panel');
    const closeBtn = document.getElementById('console-close');
    const clearBtn = document.getElementById('console-clear');
    const sidebarTerminalBtn = document.getElementById('sidebar-terminal-btn');

    const toggleConsole = () => {
        if (panel) {
            panel.classList.toggle('open');
        }
    };

    if (sidebarTerminalBtn) {
        sidebarTerminalBtn.addEventListener('click', toggleConsole);
    }

    if (closeBtn) {
        closeBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            if (panel) panel.classList.remove('open');
        });
    }

    if (clearBtn) {
        clearBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            const body = document.getElementById('console-body');
            if (body) body.innerHTML = '';
        });
    }

    // Keyboard shortcut to toggle terminal (Ctrl + ` or Cmd + `)
    window.addEventListener('keydown', (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === '`') {
            e.preventDefault();
            toggleConsole();
        }
    });
}

// ============================================
// SECURITY & ESCAPING UTILITIES
// ============================================

export function escapeHTML(str) {
    if (!str) return '';
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

// Global cover error handler for offline / missing game covers
window.handleCoverError = function(img, fallbackUrl, gameName) {
    if (fallbackUrl && !img.dataset.triedFallback) {
        img.dataset.triedFallback = '1';
        img.src = fallbackUrl;
        return;
    }
    
    // Fallback failed or offline: render premium RIFT classified dossier cover
    const frame = img.closest('.exhibit-frame');
    if (frame) {
        img.style.display = 'none';
        if (!frame.querySelector('.rift-offline-placeholder')) {
            const ph = document.createElement('div');
            ph.className = 'rift-offline-placeholder';
            ph.innerHTML = `
                <div class="rift-offline-watermark">RI<span>f</span>T</div>
                <div class="rift-offline-badge">OFFLINE DOSSIER</div>
                <div class="rift-offline-name">${escapeHTML(gameName || 'Unknown Title')}</div>
            `;
            frame.insertBefore(ph, frame.firstChild);
        }
    }
};

export function appendLog(content, className = 'log-info') {
    const body = document.getElementById('console-body');
    if (!body) return;
    const line = document.createElement('div');
    line.className = `log-line ${className}`;
    line.style.display = 'flex';
    line.style.alignItems = 'center';
    line.style.gap = '6px';
    
    const timeStr = new Date().toLocaleTimeString();
    const timeSpan = document.createElement('span');
    timeSpan.style.opacity = '0.5';
    timeSpan.style.fontSize = '0.9em';
    timeSpan.textContent = `[${timeStr}]`;

    const textSpan = document.createElement('span');
    textSpan.textContent = content;
    
    line.appendChild(timeSpan);
    line.appendChild(textSpan);
    
    body.appendChild(line);
    body.scrollTop = body.scrollHeight;
}

// ============================================
// BOOT INITIALIZATION
// ============================================



function initAuthSidebar() {
    const epicBtn = document.getElementById('epic-connect-btn');
    const steamBtn = document.getElementById('steam-connect-btn');

    if (epicBtn) {
        epicBtn.addEventListener('click', async () => {
            if (epicBtn.dataset.loading === 'true') return;
            epicBtn.dataset.loading = 'true';
            
            appendLog('[Auth] Initiating secure Epic Games login...', 'log-system');
            const originalStatus = document.getElementById('epic-status').textContent;
            document.getElementById('epic-status').textContent = "Authenticating in popup...";
            epicBtn.style.opacity = "0.7";

            try {
                const username = await window.go.rift.App.StartEpicAuth();
                if (username === "Failed") throw new Error("Auth failed or cancelled");
                
                document.getElementById('epic-username').textContent = username;
                document.getElementById('epic-status').textContent = "Status: Connected";
                epicBtn.style.opacity = "1";
                epicBtn.style.background = "rgba(46, 204, 113, 0.1)";
                epicBtn.style.borderLeft = "3px solid #2ecc71";
                appendLog(`[Auth] Epic connected as ${username}`, 'log-success');
                
                // Auto-scan local apps directory for preinstalled Epic matches
                appendLog('[Sync] Scanning Applications folder for preinstalled Epic titles...', 'log-system');
                await window.go.rift.App.ScanMacForInstalledGames();

                syncLibrary();
            } catch (e) {
                appendLog(`[Auth] Epic login failed: ${e}`, 'log-error');
                document.getElementById('epic-status').textContent = originalStatus;
                epicBtn.style.opacity = "1";
            } finally {
                epicBtn.dataset.loading = 'false';
            }
        });
    }

    if (steamBtn) {
        steamBtn.addEventListener('click', async () => {
            if (steamBtn.dataset.loading === 'true') return;
            
            // If already connected, show status
            if (store.state.authStatus?.steamConnected) {
                appendLog(`[Auth] Steam already connected as ${store.state.authStatus.steamUsername}`, 'log-info');
                return;
            }
            
            // Show the Steam login modal
            const overlay = document.getElementById('steam-login-overlay');
            if (overlay) {
                const guardGroup = document.getElementById('steam-guard-group');
                const errEl = document.getElementById('steam-error-msg');
                if (guardGroup) guardGroup.style.display = 'none';
                if (errEl) {
                    errEl.innerHTML = '';
                    errEl.style.display = 'none';
                }
                const guardInput = document.getElementById('steam-input-guard');
                if (guardInput) guardInput.value = '';
                overlay.style.display = 'flex';
                setTimeout(() => {
                    const uInput = document.getElementById('steam-input-username');
                    if (uInput) uInput.focus();
                }, 50);
            }
        });
    }

    // Steam Login Modal Logic
    const steamModalOverlay = document.getElementById('steam-login-overlay');
    const steamModalClose = document.getElementById('steam-modal-close');
    const steamSubmitBtn = document.getElementById('steam-submit-btn');

    const renderSteamError = (msg) => {
        const errEl = document.getElementById('steam-error-msg');
        if (!errEl) return;
        errEl.innerHTML = `
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#ef4444" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0;margin-top:2px;">
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="12" y1="8" x2="12" y2="12"></line>
                <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
            <span style="flex:1;line-height:1.4;">${msg}</span>
        `;
        errEl.style.display = 'flex';
    };

    const clearSteamError = () => {
        const errEl = document.getElementById('steam-error-msg');
        if (errEl) {
            errEl.innerHTML = '';
            errEl.style.display = 'none';
        }
    };

    ['steam-input-username', 'steam-input-password', 'steam-input-guard'].forEach(id => {
        const el = document.getElementById(id);
        if (el) {
            el.addEventListener('input', clearSteamError);
            el.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' && steamSubmitBtn && !steamSubmitBtn.disabled) {
                    steamSubmitBtn.click();
                }
            });
        }
    });

    if (steamModalClose) {
        steamModalClose.addEventListener('click', () => {
            if (steamModalOverlay) steamModalOverlay.style.display = 'none';
            clearSteamError();
        });
    }

    if (steamModalOverlay) {
        steamModalOverlay.addEventListener('click', (e) => {
            if (e.target === steamModalOverlay) {
                steamModalOverlay.style.display = 'none';
                clearSteamError();
            }
        });
    }

    if (steamSubmitBtn) {
        steamSubmitBtn.addEventListener('click', async () => {
            const username = (document.getElementById('steam-input-username')?.value || '').trim();
            const password = document.getElementById('steam-input-password')?.value || '';
            const authCode = (document.getElementById('steam-input-guard')?.value || '').trim();
            
            if (!username || !password) {
                renderSteamError('Please enter your Steam account name and password.');
                return;
            }

            steamSubmitBtn.disabled = true;
            steamSubmitBtn.textContent = 'Signing in...';
            clearSteamError();

            appendLog(`[Auth] Connecting Steam as ${username}...`, 'log-system');

            try {
                const result = await window.go.rift.App.StartSteamAuth(username, password, authCode);

                if (result && result.needsGuard) {
                    // Show the Steam Guard field for a code request
                    const guardGroup = document.getElementById('steam-guard-group');
                    if (guardGroup) guardGroup.style.display = 'block';
                    steamSubmitBtn.disabled = false;
                    steamSubmitBtn.textContent = 'Submit Code';
                    
                    const guardInput = document.getElementById('steam-input-guard');
                    if (authCode) {
                        // User already submitted a code, but it was rejected or expired
                        if (guardInput) {
                            guardInput.value = '';
                            guardInput.focus();
                        }
                        const msg = result.error || 'The Steam Guard code was wrong or expired. Please enter a fresh code.';
                        renderSteamError(msg);
                        appendLog('[Auth] Steam Guard code was wrong or expired. Try again.', 'log-error');
                    } else {
                        // First time code prompt
                        if (guardInput) guardInput.focus();
                        appendLog('[Auth] Steam Guard / 2FA code required. Check your email or authenticator app.', 'log-warn');
                    }
                    return;
                }

                if (!result || result.success === false) {
                    const msg = result?.error || 'Steam login failed. Check your credentials and try again.';
                    renderSteamError(msg);
                    steamSubmitBtn.disabled = false;
                    steamSubmitBtn.textContent = 'Sign in';
                    appendLog(`[Auth] Steam login failed: ${msg}`, 'log-error');
                    return;
                }

                // Success
                appendLog(`[Auth] Steam connected as ${result.username || username}`, 'log-success');
                if (steamModalOverlay) steamModalOverlay.style.display = 'none';
                checkAuthStatus();
                syncLibrary();
            } catch (err) {
                renderSteamError(err.message || 'An unexpected error occurred during sign in.');
                appendLog(`[Auth Error] Steam login failed: ${err}`, 'log-error');
            } finally {
                steamSubmitBtn.disabled = false;
                const guardGroup = document.getElementById('steam-guard-group');
                if (guardGroup && guardGroup.style.display === 'block') {
                    steamSubmitBtn.textContent = 'Submit Code';
                } else {
                    steamSubmitBtn.textContent = 'Sign in';
                }
            }
        });
    }


    // Reactive subscriptions to update sidebar UI elements
    store.subscribe('steamLoading', (loading) => {
        if (!steamBtn) return;
        if (loading) {
            steamBtn.style.opacity = "0.7";
            const isInstalled = store.state.authStatus?.steamInstalled;
            document.getElementById('steam-status').textContent = isInstalled ? "Launching Steam (Please wait)..." : "Downloading Engine & Steam...";
        } else {
            steamBtn.style.opacity = "1";
        }
    });

    store.subscribe('authStatus', (status) => {
        if (!status) return;
        
        if (steamBtn) {
            if (status.steamConnected) {
                document.getElementById('steam-username').textContent = status.steamUsername;
                document.getElementById('steam-status').textContent = "Status: Connected";
                steamBtn.style.background = "rgba(46, 204, 113, 0.1)";
                steamBtn.style.borderLeft = "3px solid #2ecc71";
            } else {
                document.getElementById('steam-username').textContent = "Connect Steam";
                document.getElementById('steam-status').textContent = "Status: Disconnected";
                steamBtn.style.background = "rgba(255, 255, 255, 0.05)";
                steamBtn.style.borderLeft = "none";
            }
        }
        
        if (epicBtn) {
            if (status.epicConnected) {
                document.getElementById('epic-username').textContent = status.epicUsername;
                document.getElementById('epic-status').textContent = "Status: Connected";
                epicBtn.style.background = "rgba(46, 204, 113, 0.1)";
                epicBtn.style.borderLeft = "3px solid #2ecc71";
            } else {
                document.getElementById('epic-username').textContent = "Connect Epic";
                document.getElementById('epic-status').textContent = "Status: Disconnected";
                epicBtn.style.background = "rgba(255, 255, 255, 0.05)";
                epicBtn.style.borderLeft = "none";
            }
        }
    });

    const topBarProfile = document.getElementById('top-bar-profile');
    if (topBarProfile) {
        topBarProfile.addEventListener('click', () => {
            window.location.hash = '/profile';
        });
    }

    const updateTopBarAvatar = () => {
        const topBarAvatar = document.getElementById('top-bar-avatar');
        if (!topBarAvatar) return;
        const profile = store.state.profile || {};
        let avatarColor = profile.avatarColor;
        if (!avatarColor) {
            try {
                const saved = localStorage.getItem('rift_avatar_color');
                if (saved) avatarColor = JSON.parse(saved);
            } catch (e) {}
        }
        if (avatarColor && avatarColor.bg) {
            topBarAvatar.style.background = avatarColor.bg;
            topBarAvatar.style.color = avatarColor.textColor || '#080808';
        } else {
            topBarAvatar.style.background = '#38bdf8';
            topBarAvatar.style.color = '#080808';
        }
        const getCleanAlias = () => {
            if (profile.name && profile.name !== "Unknown") return profile.name;
            if (profile.alias && profile.alias !== "Unknown") return profile.alias;
            if (profile.username && profile.username !== "Unknown") return profile.username;
            return "Abhinaw";
        };
        const alias = getCleanAlias();
        topBarAvatar.textContent = alias.charAt(0).toUpperCase();
    };

    updateTopBarAvatar();
    store.subscribe('profile', () => updateTopBarAvatar());

    // Transmission Notification Check
    const checkTransmissions = async () => {
        try {
            if (window.go && window.go.rift && window.go.rift.App && window.go.rift.App.GetUserProfileData) {
                const statsJson = await window.go.rift.App.GetUserProfileData();
                const parsed = JSON.parse(statsJson);
                if (parsed && parsed.alias && parsed.alias !== "Unknown") {
                    const cleanAlias = parsed.alias.replace(/_m4$/i, '');
                    if (store.state.profile.name !== cleanAlias) {
                        store.update('profile', {
                            ...store.state.profile,
                            name: cleanAlias,
                            alias: cleanAlias,
                            username: cleanAlias
                        });
                    }
                }
                if (parsed && parsed.transmissions && parsed.transmissions.length > 0) {
                    const latest = parsed.transmissions[0];
                    const latestId = String(latest.id || latest.date || latest.version_tag);
                    const seenKey = localStorage.getItem('rift_seen_transmission');
                    const notifyDot = document.getElementById('profile-notify-dot');
                    if (notifyDot) {
                        if (!seenKey || seenKey !== latestId) {
                            notifyDot.style.display = 'block';
                        } else {
                            notifyDot.style.display = 'none';
                        }
                    }
                }
            }
        } catch (err) {
            console.warn("[Main] Transmissions check failed:", err);
        }
    };
    checkTransmissions();
    setInterval(checkTransmissions, 60000); // Check every minute
}

// ============================================
// STEAM CHAT (opens locked-down WebView)
// ============================================
function initSteamChat() {
    const chatBtn = document.getElementById('steam-chat-btn');
    if (!chatBtn) return;

    chatBtn.addEventListener('click', async () => {
        appendLog('[Steam Chat] Opening Steam Chat browser...', 'log-system');
        try {
            await window.go.rift.App.OpenSteamChat();
            appendLog('[Steam Chat] Steam Chat window opened.', 'log-success');
        } catch (err) {
            appendLog(`[Steam Chat] Failed: ${err}`, 'log-error');
        }
    });
}

async function checkAuthStatus() {
    try {
        const status = await window.go.rift.App.GetAuthStatus();
        store.update('authStatus', status);
        
        try {
            const supUserStr = await window.go.rift.App.SupabaseGetUser();
            const supUser = JSON.parse(supUserStr);
            if (supUser.email) {
                const uname = (supUser.user_metadata && supUser.user_metadata.name) || supUser.email.split('@')[0];
                const cleanUname = uname.replace(/_m4$/i, '');
                store.update('profile', {
                    ...store.state.profile,
                    name: cleanUname,
                    alias: cleanUname,
                    username: cleanUname
                });
            }
        } catch (e) {
            appendLog(`[Auth] Failed to load Supabase profile: ${e}`, 'log-error');
        }

        // Always sync library on boot (local capsules + store games)
        syncLibrary();
    } catch (e) {
        appendLog(`[Auth] Failed to retrieve login status: ${e}`, 'log-error');
        syncLibrary();
    }
}

async function syncLibrary() {
    appendLog('[Sync] Syncing games from local capsules and store configs...', 'log-system');
    try {
        const ownedGames = await window.go.rift.App.SyncLibrary();
        store.update('games', ownedGames);
        appendLog(`[Sync] Successfully loaded ${ownedGames ? ownedGames.length : 0} games in RIFT Archive.`, 'log-success');
        if (ownedGames && ownedGames.length > 0) {
            appendLog(`[Sync] Sample game: ${ownedGames[0].name} [${ownedGames[0].platform || 'unknown'}]`, 'log-info');
        }

        // Restore any active or interrupted downloads from disk
        if (window.go?.rift?.App?.GetActiveDownloads) {
            try {
                const activeDls = await window.go.rift.App.GetActiveDownloads();
                if (activeDls && activeDls.length > 0) {
                    const latest = activeDls[0];
                    const g = (ownedGames || []).find(x => x.id === latest.gameId);
                    store.update('activeDownload', {
                        gameId: latest.gameId,
                        gameName: latest.gameName || (g ? g.name : latest.gameId),
                        percent: latest.percent || 0,
                        speed: latest.speed || 0,
                        status: latest.status || 'download_interrupted',
                        stage: latest.status === 'download_interrupted' ? 'Download interrupted. Ready to resume.' : (latest.stage || 'Downloading...')
                    });
                }
            } catch (dlErr) {
                console.warn('[Sync] Failed querying active downloads:', dlErr);
            }
        }
    } catch (e) {
        appendLog(`[Sync ERROR] Library sync failed: ${e}`, 'log-error');
    }
}

export async function checkForAppUpdates(manual = false) {
    if (!window.go?.rift?.App?.CheckForUpdates) return null;
    try {
        const update = await window.go.rift.App.CheckForUpdates();
        if (update && update.updateAvailable) {
            appendLog(`[Updater] New version available: v${update.latestVersion} (Current: v${update.currentVersion})`, 'log-warn');
            showUpdateModal(update);
        } else if (manual) {
            appendLog(`[Updater] RIFT is up to date (v${update?.currentVersion || '1.0.0-beta.2'}).`, 'log-success');
        }
        return update;
    } catch (err) {
        console.warn('[Updater] Update check failed:', err);
        if (manual) appendLog(`[Updater] Check failed: ${err}`, 'log-error');
        throw err;
    }
}

// Expose on window for runtime and views access
window.checkForAppUpdates = checkForAppUpdates;

export function showUpdateModal(update) {
    if (document.getElementById('rift-update-modal')) return;


    const modal = document.createElement('div');
    modal.id = 'rift-update-modal';
    modal.style.cssText = 'position:fixed;inset:0;background:rgba(0,0,0,0.75);backdrop-filter:blur(14px);z-index:99999;display:flex;align-items:center;justify-content:center;animation:fadeIn 0.3s ease;';

    const safeNotes = escapeHTML(update.releaseNotes || 'A new build of RIFT is ready with critical fixes and performance improvements.');
    const safeVersion = escapeHTML(update.latestVersion);
    const currentVer = escapeHTML(update.currentVersion);

    modal.innerHTML = `
        <div style="background:linear-gradient(135deg,#0d0e12,#181a20);border:1px solid rgba(255,255,255,0.15);border-radius:16px;width:90%;max-width:520px;padding:32px;box-shadow:0 24px 64px rgba(0,0,0,0.7);display:flex;flex-direction:column;gap:18px;">
            <div style="display:flex;align-items:center;justify-content:space-between;">
                <div style="display:flex;align-items:center;gap:10px;">
                    <div style="width:36px;height:36px;border-radius:8px;background:linear-gradient(135deg,#38bdf8,#818cf8);display:flex;align-items:center;justify-content:center;font-weight:900;color:#000;font-size:16px;">R</div>
                    <div>
                        <h2 style="font-family:'Plus Jakarta Sans',sans-serif;font-size:1.15rem;font-weight:700;color:#fff;margin:0;">RIFT Update Ready</h2>
                        <span style="font-family:'JetBrains Mono',monospace;font-size:0.75rem;color:rgba(255,255,255,0.5);">v${currentVer} → <strong style="color:#38bdf8;">v${safeVersion}</strong></span>
                    </div>
                </div>
                <button id="update-close-btn" style="background:none;border:none;color:rgba(255,255,255,0.4);font-size:1.2rem;cursor:pointer;">✕</button>
            </div>

            <div style="background:rgba(0,0,0,0.3);border:1px solid rgba(255,255,255,0.06);border-radius:10px;padding:16px;max-height:160px;overflow-y:auto;font-family:'Inter',sans-serif;font-size:0.82rem;line-height:1.5;color:rgba(255,255,255,0.8);white-space:pre-line;">
                ${safeNotes}
            </div>

            <div id="update-dl-progress-box" style="display:none;flex-direction:column;gap:8px;">
                <div style="display:flex;justify-content:space-between;font-family:'JetBrains Mono',monospace;font-size:0.75rem;color:rgba(255,255,255,0.7);">
                    <span id="update-dl-status">Downloading update...</span>
                    <span id="update-dl-percent">0%</span>
                </div>
                <div style="height:6px;background:rgba(255,255,255,0.08);border-radius:4px;overflow:hidden;">
                    <div id="update-dl-fill" style="height:100%;width:0%;background:linear-gradient(90deg,#38bdf8,#818cf8);transition:width 0.2s ease;"></div>
                </div>
            </div>

            <div style="display:flex;gap:12px;margin-top:8px;">
                <button id="update-now-btn" style="flex:1;background:linear-gradient(135deg,#38bdf8,#818cf8);color:#000;border:none;padding:12px 20px;border-radius:8px;font-family:'Plus Jakarta Sans',sans-serif;font-weight:700;font-size:0.9rem;cursor:pointer;transition:opacity 0.2s;">
                    UPDATE NOW
                </button>
                <button id="update-later-btn" style="background:rgba(255,255,255,0.06);color:rgba(255,255,255,0.7);border:1px solid rgba(255,255,255,0.1);padding:12px 20px;border-radius:8px;font-family:'Plus Jakarta Sans',sans-serif;font-weight:600;font-size:0.85rem;cursor:pointer;">
                    LATER
                </button>
            </div>
        </div>
    `;

    document.body.appendChild(modal);

    const closeBtn = modal.querySelector('#update-close-btn');
    const laterBtn = modal.querySelector('#update-later-btn');
    const updateBtn = modal.querySelector('#update-now-btn');
    const progressBox = modal.querySelector('#update-dl-progress-box');
    const fillBar = modal.querySelector('#update-dl-fill');
    const pctText = modal.querySelector('#update-dl-percent');
    const statusText = modal.querySelector('#update-dl-status');

    const closeModal = () => modal.remove();
    closeBtn.addEventListener('click', closeModal);
    laterBtn.addEventListener('click', closeModal);

    let eventsRegistered = false;
    updateBtn.addEventListener('click', async () => {
        updateBtn.disabled = true;
        updateBtn.style.opacity = '0.5';
        updateBtn.innerText = 'INITIALIZING...';
        progressBox.style.display = 'flex';

        if (window.runtime && window.runtime.EventsOn && !eventsRegistered) {
            eventsRegistered = true;
            window.runtime.EventsOn('update_progress', (data) => {
                const pct = data.percent || 0;
                fillBar.style.width = `${pct}%`;
                pctText.innerText = `${pct.toFixed(0)}%`;
                const mbDone = ((data.downloaded || 0) / (1024 * 1024)).toFixed(1);
                const mbTotal = ((data.total || 0) / (1024 * 1024)).toFixed(1);
                statusText.innerText = `Downloading: ${mbDone} MB / ${mbTotal} MB`;
            });

            window.runtime.EventsOn('update_completed', (path) => {
                fillBar.style.width = '100%';
                pctText.innerText = '100%';
                statusText.innerText = 'Download complete! Opening installer...';
                updateBtn.innerText = 'INSTALLING...';
                setTimeout(closeModal, 4000);
            });

            window.runtime.EventsOn('update_failed', (err) => {
                statusText.innerText = `Error: ${err}`;
                updateBtn.disabled = false;
                updateBtn.style.opacity = '1';
                updateBtn.innerText = 'RETRY';
            });
        }

        try {
            if (window.go?.rift?.App?.DownloadAndApplyUpdate) {
                await window.go.rift.App.DownloadAndApplyUpdate(update.downloadUrl);
            } else if (window.runtime?.BrowserOpenURL) {
                window.runtime.BrowserOpenURL(update.downloadUrl);
                closeModal();
            } else {
                window.open(update.downloadUrl, '_blank');
                closeModal();
            }
        } catch (e) {
            statusText.innerText = `Update error: ${e}`;
            updateBtn.disabled = false;
            updateBtn.style.opacity = '1';
            updateBtn.innerText = 'OPEN IN BROWSER';
            updateBtn.onclick = () => {
                if (window.runtime?.BrowserOpenURL) window.runtime.BrowserOpenURL(update.downloadUrl);
                else window.open(update.downloadUrl, '_blank');
            };
        }
    });
}

document.addEventListener('DOMContentLoaded', () => {
    showStartupSplash();
    initRouter();
    initCommandBar();
    initNetworkMonitor();

    // Setup Top Navigation Buttons
    const backBtn = document.querySelector('.nav-controls .control-btn[title="Back"]');
    const fwdBtn = document.querySelector('.nav-controls .control-btn[title="Forward"]');
    const topBarProfile = document.getElementById('top-bar-profile');
    
    if (backBtn) backBtn.addEventListener('click', () => window.history.back());
    if (fwdBtn) fwdBtn.addEventListener('click', () => window.history.forward());
    
    if (topBarProfile) {
        topBarProfile.addEventListener('click', () => {
            if (window.location.hash === '#/profile') {
                window.history.back(); // If already on profile, act as a close/back button
            } else {
                window.location.hash = '/profile';
            }
        });
    }

    window.showStartupSplash = showStartupSplash;
    window.addEventListener('keydown', (e) => {
        if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key.toLowerCase() === 'i') {
            e.preventDefault();
            showStartupSplash();
        }
    });

    initConsole();
    initQuickLaunch();
    initAuthSidebar();
    initSteamChat();
    checkAuthStatus();
    checkForAppUpdates();
    appendLog("RIFT v1.0.0-beta.2 initialized successfully.", "log-system");

    // Bind Wails Events
    if (window.runtime && window.runtime.EventsOn) {

        // --- Acquisition pipeline events ---

        // Game setup complete (AI + prefix ready, download about to start)
        window.runtime.EventsOn('game_setup_complete', (gameID) => {
            appendLog(`[RIFT Engine] Game environment configured: ${gameID}`, 'log-success');
            syncLibrary();
        });

        // Scraper background update
        window.runtime.EventsOn('scraper_update', () => {
            syncLibrary();
        });

        // AI config ready
        window.runtime.EventsOn('game_config_ready', (data) => {
            const gid = typeof data === 'string' ? data : data.gameId;
            appendLog(`[RIFT Engine] AI configuration ready for ${gid}`, 'log-success');
        });

        // Download started — set activeDownload state (preserve existing progress)
        window.runtime.EventsOn('download_started', (data) => {
            const gameId = typeof data === 'string' ? data : data.gameId;
            const gameName = data.gameName || store.state.games?.find(g => g.id === gameId)?.name || gameId;
            appendLog(`[RIFT Engine] Legendary download started for ${gameId}`, 'log-info');
            
            // Show notification dot if not already on downloads page
            if (window.location.hash !== '#/downloads') {
                const dlDot = document.getElementById('nav-dl-dot');
                if (dlDot) dlDot.style.display = 'inline-flex';
            }

            const existing = store.state.activeDownload;
            store.update('activeDownload', {
                gameId,
                gameName,
                percent: existing?.percent || 0,
                speed: 0,
                status: 'downloading',
                stage: 'Downloading game files...',
                details: existing?.details || ''
            });
            // Mark game as downloading in library
            const updated = (store.state.games || []).map(g => {
                if (g.id === gameId) return { ...g, status: 'downloading' };
                return g;
            });
            store.update('games', updated);
        });

        // Download progress update from legendary / SteamCMD
        window.runtime.EventsOn('download_progress', (data) => {
            const current = store.state.activeDownload;
            if (current) {
                const speedVal = typeof data.speed === 'number' ? data.speed : (parseFloat(data.speed) || 0);
                const speedStr = typeof speedVal === 'number' ? `${speedVal.toFixed(1)} MB/s` : '-- MB/s';
                const downVal = typeof data.downloaded === 'number' ? data.downloaded : (parseFloat(data.downloaded) || 0);
                const totalVal = typeof data.total === 'number' ? data.total : (parseFloat(data.total) || 0);

                store.update('activeDownload', {
                    ...current,
                    percent: typeof data.percent === 'number' ? data.percent : current.percent,
                    speed: speedVal,
                    downloaded: downVal,
                    total: totalVal,
                    status: 'downloading',
                    stage: `Downloading... ${(data.percent || 0).toFixed(1)}%`,
                    details: speedStr,
                    progressText: totalVal > 0 ? `${downVal.toFixed(1)} MB / ${totalVal.toFixed(1)} MB` : (data.stage || '-- MB / -- MB')
                });
            }
        });

        // Orchestration progress (prefix/AI/engine phases)
        window.runtime.EventsOn('orchestration_progress', (data) => {
            const current = store.state.activeDownload;
            if (current && current.gameId === data.gameId) {
                // User requested orchestration steps not to advance the progress bar, so it stays true to download %
                const newPercent = (data.percent === 100 || data.percent === 0) ? data.percent : current.percent;
                
                store.update('activeDownload', {
                    ...current,
                    percent: newPercent,
                    stage: data.stage || '',
                    details: data.details || ''
                });
            }
            // Log every orchestration step
            appendLog(`[RIFT Engine] [${data.percent || '?'}%] ${data.stage}: ${data.details}`, 'log-info');
        });

        // Download completed successfully
        window.runtime.EventsOn('game_installed', (gameID) => {
            appendLog(`[RIFT Engine] ${gameID} ready to play!`, 'log-success');
            store.update('activeDownload', null);
            syncLibrary();
        });

        // Library changed (re-sync)
        window.runtime.EventsOn('library_changed', () => {
            appendLog('[RIFT Engine] Library updated.', 'log-info');
            invalidateViewCache('/library');
            invalidateViewCache('/profile');
            checkAuthStatus();
            syncLibrary();
        });

        // Auth status changed (user login / logout / account switch)
        window.runtime.EventsOn('auth_status_changed', () => {
            appendLog('[Auth] User authentication state updated.', 'log-info');
            invalidateViewCache(); // Reset all views so new user starts completely fresh
            checkAuthStatus();
        });

        // Playtime updated (game session ended or synced from cloud)
        window.runtime.EventsOn('playtime_updated', (data) => {
            appendLog(`[Telemetry] Playtime updated: ${data && data.totalHours ? data.totalHours : 0} hrs.`, 'log-info');
            invalidateViewCache('/profile');
        });

        // Download failed
        window.runtime.EventsOn('download_failed', (data) => {
            const gameId = data.gameId || data.game_id || data;
            const error = data.error || "Unknown error";
            const errorType = data.errorType || 'generic';
            const requireLogin = data.requireLogin || false;
            
            const icons = {
                warn: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>`,
                error: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>`,
                disk: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"></path></svg>`,
                network: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12.55a11 11 0 0 1 14.08 0"></path><path d="M1.42 9a16 16 0 0 1 21.16 0"></path><path d="M8.53 16.11a6 6 0 0 1 6.95 0"></path><line x1="12" y1="20" x2="12.01" y2="20"></line></svg>`,
                clock: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>`
            };
            
            // Log with appropriate severity
            if (errorType === 'auth_expired') {
                appendLog(`[RIFT Engine] ${icons.warn} Steam session expired! ${error}`, 'log-error');
            } else if (errorType === 'not_owned') {
                appendLog(`[RIFT Engine] ${icons.error} ${error}`, 'log-error');
            } else if (errorType === 'disk_full') {
                appendLog(`[RIFT Engine] ${icons.disk} ${error}`, 'log-error');
            } else if (errorType === 'network') {
                appendLog(`[RIFT Engine] ${icons.network} ${error}`, 'log-error');
            } else if (errorType === 'rate_limit') {
                appendLog(`[RIFT Engine] ${icons.clock} ${error}`, 'log-error');
            } else {
                appendLog(`[RIFT Engine] Download failed: ${error}`, 'log-error');
            }
            
            const current = store.state.activeDownload;
            if (current && current.gameId === gameId) {
                store.update('activeDownload', { ...current, status: 'failed', stage: 'Download Interrupted', error: error, errorType: errorType });
            }
            
            store.update('games', store.state.games.map(g => {
                if (g.id === gameId) return { ...g, status: 'failed' };
                return g;
            }));
            
            // If login is required, refresh auth status so UI reflects logged-out state
            if (requireLogin) {
                checkAuthStatus();
            }
        });

        // Download cancelled by user
        window.runtime.EventsOn('download_cancelled', (gameID) => {
            appendLog(`[RIFT Engine] Download cancelled.`, 'log-warn');
            store.update('activeDownload', null);
            const updated = (store.state.games || []).map(g => {
                if (g.id === gameID) return { ...g, status: 'not_downloaded' };
                return g;
            });
            store.update('games', updated);
        });

        // Download timed out
        window.runtime.EventsOn('download_timedout', (gameID) => {
            appendLog(`[RIFT Engine] Download timed out after 30 minutes.`, 'log-error');
            store.update('activeDownload', null);
            const updated = (store.state.games || []).map(g => {
                if (g.id === gameID) return { ...g, status: 'not_downloaded' };
                return g;
            });
            store.update('games', updated);
        });

        // --- Legacy / other events ---

        // Listen to theme sync
        window.runtime.EventsOn("theme-sync", (themeName) => {
            document.documentElement.setAttribute('data-theme', themeName);
            appendLog(`[HUD] Theme synchronized to engine state: ${themeName}`, 'log-system');
        });

        // Listen for game process exits to revert UI from LIVE to PLAY
        window.runtime.EventsOn('game_crashed', (data) => {
            const gameId = data.gameId || data;
            const error = data.error || "Unknown crash";
            appendLog(`[RIFT Engine] Game crashed: ${error}`, 'log-error');
            const g = store.state.games.find(x => x.id === gameId);
            if (g) {
                store.update('games', store.state.games.map(x => {
                    if (x.id === gameId) return { ...x, status: 'ready' };
                    return x;
                }));
            }
            alert(`RIFT Engine detected a crash: ${error}`);
        });

        window.runtime.EventsOn('game_process_exited', (gameID) => {
            console.log(`Game process exited for ${gameID}, reverting UI state.`);
            const games = store.state.games;
            const updated = games.map(g => {
                if (g.id === gameID && g.status === 'running') {
                    return { ...g, status: 'ready', lastPlayed: 'Just now' };
                }
                return g;
            });
            store.update('games', updated);
        });

        // --- Game Launch Overlay Logic ---
        
        // --- Game Launch Overlay Logic ---
        let launchQuoteInterval = null;
        const launchQuotes = [
            "Bribing DirectX to speak Apple Metal...",
            "Convincing Windows runtimes this is actually a PC...",
            "Injecting C++ DLLs so you never have to open Terminal...",
            "Prefix armed. Launching game binary — don't touch anything..."
        ];
        
        function getOrCreateLaunchOverlay() {
            let overlay = document.getElementById('launch-overlay');
            if (!overlay) {
                overlay = document.createElement('div');
                overlay.id = 'launch-overlay';
                overlay.style = 'position: fixed; inset: 0; background: rgba(0, 0, 0, 0.94); backdrop-filter: blur(24px); -webkit-backdrop-filter: blur(24px); z-index: 10000; display: none; flex-direction: column; align-items: center; justify-content: center; opacity: 0; transition: opacity 0.5s cubic-bezier(0.16, 1, 0.3, 1); cursor: pointer; user-select: none;';
                
                overlay.innerHTML = `
                    <div style="font-family: var(--font-display, 'Smooch Sans'), 'Smooch Sans', sans-serif; font-size: 54px; font-weight: 800; color: #fff; margin-bottom: 20px; letter-spacing: 2px; text-transform: uppercase;">
                        RI<span style="font-family: var(--font-serif, 'Instrument Serif'), 'Instrument Serif', Georgia, serif; font-style: italic; font-weight: 400; text-transform: lowercase; margin-left: -4px;">f</span>T ENGINE
                    </div>
                    
                    <div style="width: 280px; height: 3px; background: rgba(255,255,255,0.08); overflow: hidden; border-radius: 3px; margin-bottom: 24px; position: relative;">
                        <div id="launch-overlay-bar" style="position: absolute; top: 0; left: 0; height: 100%; width: 40%; background: linear-gradient(90deg, #38bdf8, #818cf8); border-radius: 3px; box-shadow: 0 0 12px rgba(56, 189, 248, 0.6); animation: pulseBar 1.4s infinite ease-in-out;"></div>
                    </div>
                    
                    <style>
                        @keyframes pulseBar {
                            0% { left: -40%; width: 40%; }
                            50% { width: 70%; }
                            100% { left: 100%; width: 40%; }
                        }
                        @keyframes pulseDot {
                            0% { opacity: 0.4; transform: scale(0.85); }
                            100% { opacity: 1; transform: scale(1.15); }
                        }
                    </style>
                    
                    <div style="display: flex; flex-direction: column; align-items: center; gap: 10px; max-width: 480px; padding: 0 20px;">
                        <div style="display: inline-flex; align-items: center; gap: 8px; background: rgba(56, 189, 248, 0.08); border: 1px solid rgba(56, 189, 248, 0.25); padding: 6px 18px; border-radius: 20px;">
                            <span id="launch-overlay-dot" style="width: 8px; height: 8px; border-radius: 50%; background: #38bdf8; box-shadow: 0 0 8px #38bdf8; animation: pulseDot 1s infinite alternate;"></span>
                            <span id="launch-overlay-status" style="font-family: 'JetBrains Mono', monospace; font-size: 11.5px; font-weight: 600; color: #fff; letter-spacing: 0.6px; text-transform: uppercase;">
                                Configuring Environment...
                            </span>
                        </div>
                        <div id="launch-overlay-quote" style="font-family: 'Inter', sans-serif; font-size: 11px; color: rgba(255,255,255,0.45); text-align: center; line-height: 1.4; min-height: 16px;">
                            Bribing DirectX to speak Apple Metal...
                        </div>
                    </div>

                    <div style="position: absolute; bottom: 28px; font-family: 'JetBrains Mono', monospace; font-size: 10px; color: rgba(255,255,255,0.3); text-transform: uppercase; letter-spacing: 1.5px;">
                        Click anywhere to dismiss
                    </div>
                `;
                
                overlay.addEventListener('click', () => {
                    overlay.style.opacity = '0';
                    setTimeout(() => { overlay.style.display = 'none'; }, 400);
                    if (launchQuoteInterval) clearInterval(launchQuoteInterval);
                });

                document.body.appendChild(overlay);
            }
            return overlay;
        }

        window.runtime.EventsOn('launch_started', (data) => {
            const overlay = getOrCreateLaunchOverlay();
            const statusEl = document.getElementById('launch-overlay-status');
            const quoteEl = document.getElementById('launch-overlay-quote');
            const dotEl = document.getElementById('launch-overlay-dot');
            
            if (dotEl) {
                dotEl.style.background = '#38bdf8';
                dotEl.style.boxShadow = '0 0 8px #38bdf8';
            }
            if (statusEl) {
                statusEl.textContent = data?.status || 'Configuring Environment...';
                statusEl.style.color = '#fff';
            }
            
            let quoteIdx = 0;
            if (quoteEl) quoteEl.textContent = launchQuotes[0];
            
            if (launchQuoteInterval) clearInterval(launchQuoteInterval);
            launchQuoteInterval = setInterval(() => {
                quoteIdx = (quoteIdx + 1) % launchQuotes.length;
                if (quoteEl && overlay.style.display !== 'none') {
                    quoteEl.style.opacity = '0';
                    setTimeout(() => {
                        quoteEl.textContent = launchQuotes[quoteIdx];
                        quoteEl.style.opacity = '1';
                    }, 200);
                }
            }, 2600);
            
            overlay.style.display = 'flex';
            setTimeout(() => overlay.style.opacity = '1', 10);
            appendLog(`[RIFT Engine] Launch sequence initiated for ${data?.gameId || ''}`, 'log-system');
        });
        
        window.runtime.EventsOn('launch_status', (data) => {
            const overlay = document.getElementById('launch-overlay');
            if (overlay && overlay.style.display !== 'none') {
                const statusEl = document.getElementById('launch-overlay-status');
                const text = typeof data === 'string' ? data : (data?.status || 'Processing...');
                if (statusEl) {
                    statusEl.textContent = text;
                }
            }
        });
        
        window.runtime.EventsOn('launch_complete', (gameID) => {
            if (launchQuoteInterval) clearInterval(launchQuoteInterval);
            const overlay = document.getElementById('launch-overlay');
            if (overlay) {
                const statusEl = document.getElementById('launch-overlay-status');
                const dotEl = document.getElementById('launch-overlay-dot');
                if (statusEl) {
                    statusEl.textContent = 'Game process active! Opening window...';
                    statusEl.style.color = '#34d399';
                }
                if (dotEl) {
                    dotEl.style.background = '#34d399';
                    dotEl.style.boxShadow = '0 0 10px #34d399';
                }
                
                // Keep overlay up briefly to prevent awkward black gap before game window paints
                setTimeout(() => {
                    overlay.style.opacity = '0';
                    setTimeout(() => {
                        overlay.style.display = 'none';
                    }, 500);
                }, 1400);
            }
            appendLog(`[RIFT Engine] Translation pipeline ready. Executing game binary.`, 'log-success');
        });

        // Runtime Cache Downloads
        window.runtime.EventsOn('runtime_download_progress', (data) => {
            if (document.body.classList.contains('setup-active')) return; // Do not interrupt cinematic intros or setup

            let toast = document.getElementById('runtime-toast');
            if (!toast) {
                toast = document.createElement('div');
                toast.id = 'runtime-toast';
                toast.style = 'position: fixed; bottom: 24px; right: 24px; background: rgba(0,0,0,0.8); border: 1px solid rgba(56, 189, 248, 0.3); border-radius: 12px; padding: 16px; z-index: 10000; box-shadow: 0 10px 40px rgba(0,0,0,0.6); backdrop-filter: blur(20px); font-family: "Inter", sans-serif; min-width: 280px; transform: translateY(100px); opacity: 0; transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);';
                document.body.appendChild(toast);
                // Trigger animation
                requestAnimationFrame(() => {
                    toast.style.transform = 'translateY(0)';
                    toast.style.opacity = '1';
                });
            }
            
            toast.innerHTML = `
                <div style="display: flex; align-items: center; margin-bottom: 10px;">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" style="margin-right: 8px; animation: pulse-dot 2s infinite;"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                    <span style="color: white; font-weight: 600; font-size: 0.95rem;">Caching Master Library</span>
                </div>
                <div style="color: rgba(255,255,255,0.7); font-size: 0.75rem; margin-bottom: 12px; font-family: 'JetBrains Mono', monospace; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
                    ${data.filename} (${data.current}/${data.total})
                </div>
                <div style="height: 4px; background: rgba(255,255,255,0.1); border-radius: 4px; overflow: hidden;">
                    <div style="height: 100%; width: ${(data.current / data.total) * 100}%; background: linear-gradient(90deg, #38bdf8, #818cf8); transition: width 0.3s ease; box-shadow: 0 0 10px rgba(56, 189, 248, 0.5);"></div>
                </div>
            `;
        });

        window.runtime.EventsOn('runtime_download_complete', () => {
            if (document.body.classList.contains('setup-active')) return; // Do not interrupt cinematic intros or setup

            const toast = document.getElementById('runtime-toast');
            if (toast) {
                toast.innerHTML = `
                    <div style="display: flex; align-items: center; margin-bottom: 4px;">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#34d399" stroke-width="2" style="margin-right: 8px;"><polyline points="20 6 9 17 4 12"></polyline></svg>
                        <span style="color: #34d399; font-weight: 600; font-size: 0.95rem;">Cache Complete</span>
                    </div>
                    <div style="color: rgba(255,255,255,0.5); font-size: 0.75rem; margin-left: 26px;">Runtimes are ready.</div>
                `;
                setTimeout(() => {
                    toast.style.transform = 'translateY(100px)';
                    toast.style.opacity = '0';
                    setTimeout(() => toast.remove(), 400);
                }, 3500);
            }
        });

        // Real-Time System Telemetry (HUD)
        window.runtime.EventsOn("sys-stats", (stats) => {
            const cpuEl = document.getElementById('sys-cpu');
            const ramEl = document.getElementById('sys-ram');
            const swapEl = document.getElementById('sys-swap');
            const diskEl = document.getElementById('sys-disk');
            const dlEl = document.getElementById('sys-net-dl');
            const ulEl = document.getElementById('sys-net-ul');

            if (cpuEl) cpuEl.innerText = stats.cpu;
            if (ramEl) ramEl.innerText = stats.ram;
            if (swapEl) swapEl.innerText = stats.swap;
            if (diskEl) diskEl.innerText = stats.disk;
            if (dlEl) dlEl.innerText = stats.netDl;
            if (ulEl) ulEl.innerText = stats.netUl;
        });

        window.runtime.EventsOn('steam_logged_in', (username) => {
            appendLog(`[Auth] Steam successfully connected as user: ${username}`, 'log-success');
            checkAuthStatus();
        });

        // Global Notification Helper
        window.showRiftNotification = function(title, msg, type = 'error', onAction = null, actionText = 'DISMISS') {
            const notification = document.createElement('div');
            notification.className = 'rift-global-notification';
            const color = type === 'error' ? '#ff4444' : (type === 'warn' ? '#fbbf24' : '#34d399');
            
            const safeTitle = escapeHTML(title);
            const safeMsg = escapeHTML(msg);

            notification.innerHTML = `
                <div style="position:fixed;top:20px;right:20px;background:rgba(10,10,10,0.95);border:1px solid rgba(255,255,255,0.1);border-left:3px solid ${color};border-radius:4px;padding:20px;z-index:9999;color:#fff;font-family:'Plus Jakarta Sans',sans-serif;font-size:13px;max-width:340px;box-shadow:0 10px 30px rgba(0,0,0,0.8);backdrop-filter:blur(10px);animation:slideIn 0.3s cubic-bezier(0.16,1,0.3,1)">
                    <div style="font-family:'Instrument Serif',serif;font-size:22px;letter-spacing:0.5px;margin-bottom:6px;display:flex;align-items:center;color:#fff;">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right:8px;"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
                        ${safeTitle}
                    </div>
                    <div style="opacity:0.6;line-height:1.5;margin-bottom:16px;">${safeMsg}</div>
                    <div style="display:flex;justify-content:flex-end;gap:8px;">
                        <button class="dismiss-btn" style="background:rgba(255,255,255,0.05);color:#fff;border:1px solid rgba(255,255,255,0.1);padding:6px 14px;border-radius:4px;cursor:pointer;font-family:'JetBrains Mono',monospace;font-size:11px;transition:background 0.2s;">DISMISS</button>
                        ${onAction ? `<button class="action-btn" style="background:${color};color:#fff;border:1px solid ${color};padding:6px 14px;border-radius:4px;cursor:pointer;font-family:'JetBrains Mono',monospace;font-size:11px;font-weight:600;transition:opacity 0.2s;">${actionText}</button>` : ''}
                    </div>
                </div>`;
            
            document.body.appendChild(notification);
            
            notification.querySelector('.dismiss-btn').onclick = () => notification.remove();
            if (onAction) {
                notification.querySelector('.action-btn').onclick = () => {
                    onAction();
                    notification.remove();
                };
            }
            
            setTimeout(() => { if (document.body.contains(notification)) notification.remove(); }, 15000);
        };

        window.runtime.EventsOn('steam_session_expired', (data) => {
            const msg = data && data.message ? data.message : 'Your Steam session has expired. Please log in again.';
            appendLog(`[Auth] [Session Expired] ${msg}`, 'log-error');
            checkAuthStatus();
            window.showRiftNotification('Session Expired', msg, 'error');
        });

        window.runtime.EventsOn('epic_session_expired', (data) => {
            const msg = data && data.message ? data.message : 'Your Epic Games session has expired. Please log in again.';
            appendLog(`[Auth] [Session Expired] ${msg}`, 'log-error');
            checkAuthStatus();
            window.showRiftNotification('Session Expired', msg, 'error');
        });

        window.runtime.EventsOn('epic_sdl_prompt_required', (data) => {
            appendLog(`[Download] Optional packages available for ${data.gameName || data.gameId}. Showing options...`, 'log-info');
            showEpicSelectiveDownloadModal(data);
        });

        window.runtime.EventsOn('steam_logged_out', () => {
            appendLog('[Auth] Steam has been disconnected.', 'log-warn');
            checkAuthStatus();
        });
        window.runtime.EventsOn('steam_download_progress', (data) => {
            const match = data.raw.match(/progress:\s*([\d.]+)/);
            if (match) {
                const percent = parseFloat(match[1]).toFixed(1);
                const updatedGames = store.state.games.map(g => {
                    if (g.appID === data.appId) {
                        return { ...g, status: 'downloading', downloadProgress: percent };
                    }
                    return g;
                });
                store.update('games', updatedGames);
            }
        });
    }
});
