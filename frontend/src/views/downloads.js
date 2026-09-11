import { store } from '../state.js';
import { appendLog } from '../main.js';

export function renderDownloads() {
    const container = document.createElement('div');
    container.className = 'view-premium';
    
    // Add custom styling for the horizontal layout
    const styleId = 'downloads-minimal-css';
    if (!document.getElementById(styleId)) {
        const style = document.createElement('style');
        style.id = styleId;
        style.textContent = `
            .dl-minimal-card {
                display: flex;
                width: 100%;
                background: rgba(15, 15, 20, 0.8);
                border: 1px solid rgba(255, 255, 255, 0.08);
                border-radius: 12px;
                overflow: hidden;
                margin-top: 24px;
                box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
                backdrop-filter: blur(12px);
            }
            .dl-minimal-poster {
                width: 220px;
                min-height: 300px;
                background-size: cover;
                background-position: center;
                background-color: #111;
                border-right: 1px solid rgba(255, 255, 255, 0.08);
            }
            .dl-minimal-content {
                flex: 1;
                padding: 36px 40px;
                display: flex;
                flex-direction: column;
                justify-content: center;
                position: relative;
            }
            .dl-action-group {
                position: absolute;
                top: 24px;
                right: 24px;
                display: flex;
                align-items: center;
                gap: 10px;
            }
            .dl-minimal-title {
                font-family: 'Plus Jakarta Sans', sans-serif;
                font-weight: 700;
                font-size: 1.85rem;
                color: #fff;
                margin: 0 0 12px 0;
                letter-spacing: -0.5px;
            }
            .dl-minimal-stats {
                display: flex;
                gap: 36px;
                margin-bottom: 24px;
                color: rgba(255, 255, 255, 0.45);
                font-family: 'JetBrains Mono', monospace;
                font-size: 0.82rem;
            }
            .dl-minimal-stat-val {
                color: #fff;
                margin-top: 6px;
                font-size: 1.05rem;
                font-weight: 500;
            }
            .dl-minimal-progress-bg {
                width: 100%;
                height: 5px;
                background: rgba(255, 255, 255, 0.06);
                border-radius: 3px;
                overflow: hidden;
            }
            .dl-minimal-progress-fill {
                height: 100%;
                background: linear-gradient(90deg, #38bdf8 0%, #818cf8 100%);
                border-radius: 3px;
                transition: width 0.3s ease;
                box-shadow: 0 0 12px rgba(56, 189, 248, 0.5);
            }
            .dl-minimal-cancel {
                display: inline-flex;
                align-items: center;
                gap: 6px;
                background: rgba(255, 255, 255, 0.04);
                border: 1px solid rgba(255, 255, 255, 0.12);
                color: rgba(255, 255, 255, 0.7);
                padding: 7px 14px;
                border-radius: 8px;
                font-family: 'Inter', sans-serif;
                font-size: 0.82rem;
                font-weight: 500;
                cursor: pointer;
                transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
            }
            .dl-minimal-cancel:hover {
                background: rgba(233, 69, 96, 0.15);
                border-color: rgba(233, 69, 96, 0.4);
                color: #ff6b81;
                transform: translateY(-1px);
            }
            .dl-minimal-retry {
                display: inline-flex;
                align-items: center;
                gap: 7px;
                background: linear-gradient(135deg, #38bdf8 0%, #0284c7 100%);
                border: 1px solid rgba(56, 189, 248, 0.4);
                color: #ffffff;
                padding: 8px 18px;
                border-radius: 8px;
                font-family: 'Inter', sans-serif;
                font-size: 0.84rem;
                font-weight: 600;
                letter-spacing: 0.2px;
                cursor: pointer;
                box-shadow: 0 4px 14px rgba(2, 132, 199, 0.35), inset 0 1px 0 rgba(255, 255, 255, 0.25);
                transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
            }
            .dl-minimal-retry:hover {
                background: linear-gradient(135deg, #60a5fa 0%, #0369a1 100%);
                box-shadow: 0 6px 20px rgba(56, 189, 248, 0.5), inset 0 1px 0 rgba(255, 255, 255, 0.35);
                transform: translateY(-1px);
            }
            .dl-minimal-retry:active {
                transform: translateY(0);
                box-shadow: 0 2px 8px rgba(2, 132, 199, 0.3);
            }
            .dl-minimal-resume {
                display: inline-flex;
                align-items: center;
                gap: 7px;
                background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
                border: 1px solid rgba(245, 158, 11, 0.5);
                color: #ffffff;
                padding: 8px 18px;
                border-radius: 8px;
                font-family: 'Inter', sans-serif;
                font-size: 0.84rem;
                font-weight: 600;
                letter-spacing: 0.2px;
                cursor: pointer;
                box-shadow: 0 4px 14px rgba(217, 119, 6, 0.35), inset 0 1px 0 rgba(255, 255, 255, 0.25);
                transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
            }
            .dl-minimal-resume:hover {
                background: linear-gradient(135deg, #fbbf24 0%, #b45309 100%);
                box-shadow: 0 6px 20px rgba(245, 158, 11, 0.5), inset 0 1px 0 rgba(255, 255, 255, 0.35);
                transform: translateY(-1px);
            }
            .dl-minimal-resume:active {
                transform: translateY(0);
                box-shadow: 0 2px 8px rgba(217, 119, 6, 0.3);
            }
        `;
        document.head.appendChild(style);
    }

    function render(activeDownload) {
        container.innerHTML = `
            <div class="downloads-content-wrapper" style="max-width: 960px; margin: 0 auto; padding: 40px 32px 80px;">
                <div class="section-header" style="border-bottom: 1px solid rgba(255,255,255,0.05); padding-bottom: 20px; margin-bottom: 36px;">
                    <h1 style="font-family: 'Plus Jakarta Sans', sans-serif; font-weight: 800; font-size: 2.5rem; letter-spacing: -1px; color: white;">Downloads</h1>
                    <p style="font-family: 'Inter', sans-serif; font-size: 0.875rem; color: rgba(255,255,255,0.5); margin-top: 6px;">Active Translation & Game Acquisition Stream</p>
                </div>

                ${activeDownload ? renderActiveDownload(activeDownload) : renderEmptyState()}
            </div>
        `;

        // Wire cancel / dismiss buttons
        container.querySelectorAll('.dl-minimal-cancel').forEach(btn => {
            btn.addEventListener('click', async () => {
                if (!activeDownload) return;
                appendLog('[Downloads] Cancelling/clearing download...', 'log-warn');
                try {
                    await window.go.rift.App.CancelDownload(activeDownload.gameId);
                    store.update('activeDownload', null);
                } catch (err) {
                    appendLog(`[Downloads] Cancel failed: ${err}`, 'log-error');
                }
            });
        });

        // Wire resume button
        container.querySelector('.dl-minimal-resume')?.addEventListener('click', async () => {
            if (!activeDownload) return;
            appendLog(`[Downloads] Resuming download for ${activeDownload.gameId}...`, 'log-info');
            store.update('activeDownload', {
                ...activeDownload,
                status: 'downloading',
                speed: 0,
                stage: 'Resuming download...'
            });
            try {
                await window.go.rift.App.DownloadGame(activeDownload.gameId);
            } catch (err) {
                appendLog(`[Downloads] Resume failed: ${err}`, 'log-error');
                store.update('activeDownload', {
                    ...activeDownload,
                    status: 'download_interrupted',
                    stage: 'Download interrupted. Ready to resume.'
                });
            }
        });

        // Wire retry button
        container.querySelector('.dl-minimal-retry')?.addEventListener('click', async () => {
            if (!activeDownload) return;
            appendLog('[Downloads] Retrying download...', 'log-info');
            try {
                await window.go.rift.App.DownloadGame(activeDownload.gameId);
            } catch (err) {
                appendLog(`[Downloads] Retry failed: ${err}`, 'log-error');
            }
        });
    }

    function renderActiveDownload(dl) {
        const game = store.state.games.find(g => g.id === dl.gameId);
        let gameName = dl.gameName || (game ? game.name : '');
        if (!gameName && dl.gameId && dl.gameId.startsWith('steam-')) {
            const appId = dl.gameId.replace('steam-', '');
            if (appId === '11020') {
                gameName = 'TrackMania Nations Forever';
            } else {
                gameName = `Steam Game (${appId})`;
            }
        } else if (!gameName) {
            gameName = dl.gameId;
        }

        let cover = game ? (game.cover || game.heroCover) : '';
        if (!cover && dl.gameId && dl.gameId.startsWith('steam-')) {
            const appId = dl.gameId.replace('steam-', '');
            cover = `https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/${appId}/header.jpg`;
        }
        const pct = dl.percent || 0;
        
        let speedStr = (dl.status === 'download_interrupted' || dl.status === 'failed')
            ? '-- MB/s'
            : (typeof dl.speed === 'number' ? `${dl.speed.toFixed(1)} MB/s` : '-- MB/s');
        
        let sizesStr = '-- MB / -- MB';
        if (dl.total > 0) {
            if (dl.total >= 1024) {
                sizesStr = `${(dl.downloaded / 1024).toFixed(2)} GB / ${(dl.total / 1024).toFixed(2)} GB`;
            } else {
                sizesStr = `${dl.downloaded.toFixed(1)} MB / ${dl.total.toFixed(1)} MB`;
            }
        }
        
        const stage = dl.status === 'failed' ? 'Download Interrupted' : (dl.stage || 'Starting...');

        let actionHtml = `
            <div class="dl-action-group">
                <button class="dl-minimal-cancel">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                    Cancel
                </button>
            </div>
        `;
        let message = '';
        if (dl.status === 'download_interrupted') {
            actionHtml = `
                <div class="dl-action-group">
                    <button class="dl-minimal-cancel" title="Cancel download">Cancel</button>
                    <button class="dl-minimal-resume">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>
                        Resume Download
                    </button>
                </div>
            `;
        } else if (dl.status === 'failed') {
            actionHtml = `
                <div class="dl-action-group">
                    <button class="dl-minimal-cancel" title="Dismiss failed download">Dismiss</button>
                    <button class="dl-minimal-retry">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="1 4 1 10 7 10"></polyline><polyline points="23 20 23 14 17 14"></polyline><path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15"></path></svg>
                        Retry Download
                    </button>
                </div>
            `;
            
            const errorMsg = dl.error || (dl.stage ? dl.stage.replace(/^Download failed\.\s*/i, '') : 'Download failed');
            const errorType = dl.errorType || 'generic';
            
            let badgeTitle = 'Download Interrupted';
            let badgeIcon = `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#ff6b81" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>`;
            let badgeDesc = errorMsg;

            if (errorType === 'network' || errorMsg.toLowerCase().includes('connect') || errorMsg.toLowerCase().includes('internet')) {
                badgeTitle = 'Network Connection Lost';
                badgeIcon = `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="2" y1="12" x2="22" y2="12"></line><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path></svg>`;
                badgeDesc = 'Unable to reach game servers. Check your internet connection and click Retry to resume.';
            } else if (errorType === 'auth_expired' || errorMsg.toLowerCase().includes('session expired') || errorMsg.toLowerCase().includes('login failed') || errorMsg.toLowerCase().includes('invalid credentials')) {
                badgeTitle = 'Session Expired';
                badgeIcon = `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 2l-2 2m-1.5 1.5L16 7l-1.5-1.5M13 8.5l-1.5 1.5-3-3L2 13.5V22h8.5l6.5-6.5"></path><circle cx="16.5" cy="7.5" r="2.5"></circle></svg>`;
                badgeDesc = 'Your Epic Games / Steam session has expired. Please re-authenticate in Settings.';
            } else if (errorType === 'disk_full' || errorMsg.toLowerCase().includes('space') || errorMsg.toLowerCase().includes('disk is full')) {
                badgeTitle = 'Disk Space Full';
                badgeIcon = `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#f43f5e" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="12" x2="2" y2="12"></line><path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"></path><line x1="6" y1="16" x2="6.01" y2="16"></line><line x1="10" y1="16" x2="10.01" y2="16"></line></svg>`;
                badgeDesc = 'Not enough free disk space on your drive to complete the download.';
            } else if (badgeDesc.length > 150 || badgeDesc.includes('[cli]') || badgeDesc.includes('[Core]') || badgeDesc.includes('Traceback')) {
                badgeDesc = 'An error occurred during download. Click the Active Runtime Log drawer below for technical details.';
            }

            message = `
                <div style="margin-top: 18px; padding: 14px 18px; background: rgba(233, 69, 96, 0.08); border: 1px solid rgba(233, 69, 96, 0.25); border-radius: 10px; font-size: 13px; color: #fff; display: flex; gap: 14px; align-items: flex-start; backdrop-filter: blur(8px);">
                    <div style="display: flex; align-items: center; justify-content: center; min-width: 24px; padding-top: 2px;">${badgeIcon}</div>
                    <div style="flex: 1;">
                        <div style="font-weight: 700; color: #ff6b81; margin-bottom: 4px; font-family: 'Plus Jakarta Sans', sans-serif;">${badgeTitle}</div>
                        <div style="color: rgba(255,255,255,0.85); line-height: 1.45;">${badgeDesc}</div>
                        <div style="margin-top: 8px; font-size: 11px; color: rgba(255,255,255,0.45);">Previously downloaded files are cached on disk. Tapping Retry resumes without re-downloading finished chunks.</div>
                    </div>
                </div>
            `;
        }

        return `
            <div class="dl-minimal-card">
                <div class="dl-minimal-poster" style="background-image: url('${cover}');"></div>
                <div class="dl-minimal-content">
                    ${actionHtml}
                    <h2 class="dl-minimal-title">${gameName}</h2>
                    
                    <div class="dl-minimal-stats">
                        <div>
                            <div>NETWORK</div>
                            <div class="dl-minimal-stat-val">${speedStr}</div>
                        </div>
                        <div>
                            <div>PROGRESS</div>
                            <div class="dl-minimal-stat-val">${pct.toFixed(1)}%</div>
                        </div>
                        <div>
                            <div>TRANSFERRED</div>
                            <div class="dl-minimal-stat-val">${sizesStr}</div>
                        </div>
                    </div>

                    <div class="dl-minimal-progress-bg">
                        <div class="dl-minimal-progress-fill" style="width: ${pct}%;"></div>
                    </div>

                    <div style="display: flex; justify-content: space-between; font-size: 0.75rem; color: rgba(255,255,255,0.4); margin-top: 10px; font-family: 'JetBrains Mono', monospace;">
                        <span>${stage}</span>
                        <span>Wine 64-bit / Metal Translation</span>
                    </div>

                    ${message}
                </div>
            </div>
        `;
    }

    function renderEmptyState() {
        return `
            <div class="empty-state" style="background: rgba(255, 255, 255, 0.02); border: 1px dashed rgba(255, 255, 255, 0.08); border-radius: 12px; padding: 80px 40px; text-align: center; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px;">
                <div style="width: 48px; height: 48px; border-radius: 50%; background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.08); display: flex; align-items: center; justify-content: center; font-size: 1.2rem; color: rgba(255,255,255,0.4); margin-bottom: 8px;">↓</div>
                <h2 style="font-family: 'Plus Jakarta Sans', sans-serif; font-weight: 700; font-size: 1.4rem; color: white; margin: 0;">Zero Active Downloads</h2>
                <p style="font-family: 'Inter', sans-serif; font-size: 0.9rem; color: rgba(255,255,255,0.6); margin: 0; max-width: 450px; line-height: 1.6;">Your Wi-Fi router is finally getting a break. Pick a game from The Exhibition to start downloading.</p>
            </div>
        `;
    }

    const unsubDownload = store.subscribe('activeDownload', (dl) => {
        render(dl);
    });

    const unsubGames = store.subscribe('games', () => {
        const dl = store.state.activeDownload;
        if (dl) render(dl);
    });

    render(store.state.activeDownload);

    container._cleanup = () => {
        unsubDownload();
        unsubGames();
    };

    return container;
}
