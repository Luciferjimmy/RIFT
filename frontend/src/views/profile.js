import { store } from '../state.js';
import { appendLog, escapeHTML } from '../main.js';

// Setup smart image handlers to detect landscape banners and prevent awkward cropping
window.handleCoverLoad = function(img) {
    if (!img) return;
    const frame = img.closest('.exhibit-frame');
    if (!frame) return;
    if (img.naturalWidth > img.naturalHeight * 1.15) {
        frame.classList.add('has-horizontal-banner');
        const backdrop = frame.querySelector('.cover-art-backdrop');
        if (backdrop) {
            backdrop.style.backgroundImage = `url('${img.src}')`;
            backdrop.style.display = 'block';
        }
    } else {
        frame.classList.remove('has-horizontal-banner');
        const backdrop = frame.querySelector('.cover-art-backdrop');
        if (backdrop) backdrop.style.display = 'none';
    }
};

window.handleCoverErrorCustom = function(img) {
    if (!img) return;
    const fallback = img.dataset.fallback;
    if (fallback && !img.dataset.triedFallback) {
        img.dataset.triedFallback = '1';
        img.src = fallback;
        return;
    }
    if (window.handleCoverError) {
        window.handleCoverError(img, '', img.dataset.gamename || '');
    }
};

export async function renderProfile() {
    const container = document.createElement('div');
    container.className = 'profile-view fade-in';
    container.style.cssText = 'padding: 36px 44px; color: #f5f5f5; height: 100%; overflow-y: auto; font-family: var(--font-sans, "Plus Jakarta Sans", sans-serif); background: #080808;';

    let realStats = {
        alias: "Abhinaw",
        clearance: "Founder / Tier 1",
        rift_id: "RIFT-F-000000",
        chip: "Apple M3",
        os_target: "macOS 26.6.2",
        playtime_minutes: 852,
        games_owned: 0,
        games_played: 15,
        current_streak: 1,
        join_date: "2026-08-19",
        transmissions: []
    };

    let playtimeTotalHours = "0.0";
    let titlesPlayedCount = 0;
    let playtimeMap = {};

    // 1. Fetch live telemetry playtime stats from Go engine (reads ~/.rift/data/playtime.json)
    try {
        if (window.go && window.go.rift && window.go.rift.App && window.go.rift.App.GetPlaytimeStats) {
            const ptStr = await window.go.rift.App.GetPlaytimeStats();
            const pt = JSON.parse(ptStr);
            if (pt && typeof pt.total === 'number' && pt.total > 0) {
                playtimeTotalHours = (pt.total / 3600).toFixed(1);
            }
            if (pt && pt.per_game) {
                playtimeMap = pt.per_game;
                const playedIds = Object.keys(pt.per_game).filter(id => pt.per_game[id] > 0);
                if (playedIds.length > 0) {
                    titlesPlayedCount = playedIds.length;
                }
            }
        }
    } catch (err) {
        console.warn("[Profile] Playtime stats fallback:", err);
    }

    // 2. Fetch User Profile, Chip, and Transmissions from Go Backend
    try {
        if (window.go && window.go.rift && window.go.rift.App && window.go.rift.App.GetUserProfileData) {
            const statsJson = await window.go.rift.App.GetUserProfileData();
            const parsedStats = JSON.parse(statsJson);
            if (!parsedStats.error) {
                realStats = { ...realStats, ...parsedStats };
                if (realStats.alias) realStats.alias = realStats.alias.replace(/_m4$/i, '');
            }
        }
    } catch (err) {
        console.warn("[Profile] Using local profile data:", err);
    }

    // 3. Fetch Last Played Session Translation Core Info (Dynamic from latest session / runtime config)
    let lastCoreData = {
        game_id: "",
        game_name: "No Games Executed Yet",
        engine: "standby",
        core_title: "Translation Engine Standby",
        core_badge: "READY",
        subtitle: "Launch your first title",
        layer_details: "Apple Silicon Translation Pipeline Ready",
        platform: "rift",
        is_mac_native: false
    };
    try {
        if (window.go && window.go.rift && window.go.rift.App && window.go.rift.App.GetLastSessionCoreInfo) {
            const coreJson = await window.go.rift.App.GetLastSessionCoreInfo();
            const parsedCore = JSON.parse(coreJson);
            if (parsedCore && parsedCore.core_title) {
                lastCoreData = parsedCore;
            }
        }
    } catch (err) {
        console.warn("[Profile] Using fallback last core data:", err);
    }

    // Clear notification dot when viewing profile
    const notifyDot = document.getElementById('profile-notify-dot');
    if (notifyDot) notifyDot.style.display = 'none';
    if (realStats.transmissions && realStats.transmissions.length > 0) {
        const latest = realStats.transmissions[0];
        localStorage.setItem('rift_seen_transmission', String(latest.id || latest.date || latest.version_tag));
    }

    const games = store.state.games || [];
    const gamesOwnedCount = games.length > 0 ? games.length : (realStats.games_owned || 0);

    // Avatar Colorways (Choose Color Feature)
    const colorways = [
        { name: "Electric Cyan", bg: "#38bdf8", border: "#0284c7", textColor: "#080808" },
        { name: "Acid Lime", bg: "#00e676", border: "#059669", textColor: "#080808" },
        { name: "Warm Amber", bg: "#fbbf24", border: "#d97706", textColor: "#080808" },
        { name: "Coral Rose", bg: "#fb7185", border: "#e11d48", textColor: "#080808" },
        { name: "Mint Sage", bg: "#2dd4bf", border: "#0d9488", textColor: "#080808" },
        { name: "Neon Violet", bg: "#a855f7", border: "#7c3aed", textColor: "#080808" }
    ];

    let currentAvatarColor = colorways[0];
    const savedAvatarColor = localStorage.getItem('rift_avatar_color');
    if (savedAvatarColor) {
        try { currentAvatarColor = JSON.parse(savedAvatarColor); } catch (e) {}
    } else if (store.state.profile && store.state.profile.avatarColor) {
        currentAvatarColor = store.state.profile.avatarColor;
    }

    // Sort games to get the Top 4 Most Played / Recent Titles
    const sortedGames = [...games].sort((a, b) => {
        const aSecs = playtimeMap[a.id] || playtimeMap[a.appID] || 0;
        const bSecs = playtimeMap[b.id] || playtimeMap[b.appID] || 0;
        if (bSecs !== aSecs) return bSecs - aSecs;
        return (b.lastPlayedTs || 0) - (a.lastPlayedTs || 0);
    });

    const topPlayedGames = sortedGames.slice(0, 4);

    const snakeHighScore = parseInt(localStorage.getItem('rift_snake_highscore') || '0', 10);

    const parseMarkdown = (md) => {
        if (!md) return '';
        let parsed = md.replace(/\\n/g, '\n').replace(/\\r/g, '');
        return parsed
            .replace(/^[ \t]*### (.*$)/gim, '<h3 style="color:#f5f5f5; margin: 12px 0 6px 0; font-size: 0.95rem; font-weight: 700;">$1</h3>')
            .replace(/^[ \t]*## (.*$)/gim, '<h2 style="color:#f5f5f5; margin: 16px 0 8px 0; font-size: 1.1rem; font-weight: 700;">$1</h2>')
            .replace(/^[ \t]*# (.*$)/gim, '<h1 style="color:#fff; margin: 20px 0 10px 0; font-size: 1.3rem; font-weight: 800;">$1</h1>')
            .replace(/^[ \t]*\* (.*$)/gim, '<li style="margin-left: 20px; color: rgba(255,255,255,0.75); line-height: 1.6;">$1</li>')
            .replace(/\n/gim, '<br>');
    };

    container.innerHTML = `
        <style>
            .editorial-profile-container {
                max-width: 1140px;
                margin: 0 auto;
                padding-bottom: 80px;
            }

            /* Editorial Identity Frame with Corner Crop Marks */
            .id-framed-card {
                position: relative;
                background: #0d0f17;
                border: 1px solid rgba(255, 255, 255, 0.08);
                border-radius: 6px;
                padding: 32px 36px;
                margin-bottom: 32px;
                display: flex;
                justify-content: space-between;
                align-items: center;
                gap: 28px;
                flex-wrap: wrap;
                box-shadow: 0 8px 30px rgba(0,0,0,0.5);
            }
            .crop-corner {
                position: absolute;
                width: 10px;
                height: 10px;
                pointer-events: none;
            }
            .crop-tl { top: 6px; left: 6px; border-top: 1.5px solid rgba(255,255,255,0.25); border-left: 1.5px solid rgba(255,255,255,0.25); }
            .crop-tr { top: 6px; right: 6px; border-top: 1.5px solid rgba(255,255,255,0.25); border-right: 1.5px solid rgba(255,255,255,0.25); }
            .crop-bl { bottom: 6px; left: 6px; border-bottom: 1.5px solid rgba(255,255,255,0.25); border-left: 1.5px solid rgba(255,255,255,0.25); }
            .crop-br { bottom: 6px; right: 6px; border-bottom: 1.5px solid rgba(255,255,255,0.25); border-right: 1.5px solid rgba(255,255,255,0.25); }

            /* Refined Editorial Chips (Natural & Professional) */
            .color-block-chip {
                display: inline-flex;
                align-items: center;
                gap: 6px;
                padding: 4px 10px;
                font-family: var(--font-mono, "JetBrains Mono", monospace);
                font-size: 0.72rem;
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.04em;
                border-radius: 4px;
                background: rgba(255, 255, 255, 0.035);
                border: 1px solid rgba(255, 255, 255, 0.08);
                color: rgba(255, 255, 255, 0.75);
            }
            .chip-amber { background: rgba(245, 158, 11, 0.08); color: #fbbf24; border: 1px solid rgba(245, 158, 11, 0.25); }
            .chip-cyan { background: rgba(56, 189, 248, 0.08); color: #38bdf8; border: 1px solid rgba(56, 189, 248, 0.22); }
            .chip-muted { background: rgba(255, 255, 255, 0.035); color: rgba(255, 255, 255, 0.75); border: 1px solid rgba(255, 255, 255, 0.08); }

            /* Section Headers */
            .editorial-section-title {
                margin: 36px 0 16px 0;
                display: flex;
                align-items: baseline;
                justify-content: space-between;
            }
            .title-main {
                font-size: 1.25rem;
                font-weight: 800;
                color: #fff;
                letter-spacing: -0.02em;
                text-transform: uppercase;
                display: flex;
                align-items: baseline;
                gap: 8px;
            }
            .title-main em {
                font-family: var(--font-serif, "Instrument Serif", Georgia, serif);
                font-style: italic;
                font-size: 1.55rem;
                text-transform: none;
                color: #fff;
                font-weight: 400;
            }

            /* Stat Cards */
            .retrospective-stat-card {
                background: #0d0f17;
                border: 1px solid rgba(255, 255, 255, 0.08);
                border-radius: 6px;
                padding: 22px 24px;
                display: flex;
                flex-direction: column;
                justify-content: space-between;
                min-height: 128px;
                position: relative;
                transition: border-color 0.2s ease, transform 0.2s ease;
            }
            .retrospective-stat-card:hover {
                border-color: rgba(255, 255, 255, 0.18);
                transform: translateY(-2px);
            }
            .stat-metric-val {
                font-size: 2.6rem;
                font-weight: 800;
                letter-spacing: -0.03em;
                line-height: 1;
                margin-top: 10px;
                color: #fff;
            }

            /* Epic & Custom Artifacts Style Grid (4 in a Row) */
            .profile-exhibits-grid {
                display: grid;
                grid-template-columns: repeat(4, 1fr);
                gap: 22px;
                margin-bottom: 44px;
            }
            @media (max-width: 1040px) {
                .profile-exhibits-grid {
                    grid-template-columns: repeat(2, 1fr);
                }
            }
            @media (max-width: 580px) {
                .profile-exhibits-grid {
                    grid-template-columns: 1fr;
                }
            }

            .exhibit-card {
                background: transparent;
                border-radius: 8px;
                overflow: visible;
                display: flex;
                flex-direction: column;
                position: relative;
                cursor: pointer;
            }

            .exhibit-frame {
                position: relative;
                aspect-ratio: 10 / 14;
                border-radius: 8px;
                overflow: hidden;
                border: 1px solid rgba(255, 255, 255, 0.08);
                background: #0d0d0d;
                padding: 6px;
                box-shadow: 0 10px 30px rgba(0, 0, 0, 0.6);
                transition: border-color 0.3s ease, box-shadow 0.3s ease, transform 0.25s ease;
            }
            .exhibit-card:hover .exhibit-frame {
                border-color: rgba(255, 255, 255, 0.25);
                transform: translateY(-4px);
                box-shadow: 0 16px 40px rgba(0, 0, 0, 0.8);
            }
            .exhibit-frame img.cover-art-img {
                display: block;
                width: 100%;
                height: 100%;
                object-fit: cover;
                object-position: center;
                border-radius: 4px;
                position: relative;
                z-index: 2;
                transition: transform 0.4s ease, filter 0.4s ease;
                filter: saturate(0.9) brightness(0.95);
            }
            .exhibit-card:hover .exhibit-frame img.cover-art-img {
                transform: scale(1.03);
                filter: saturate(1) brightness(1.05);
            }

            /* Smart Horizontal Banner Support (Prevents Awkward Cropping for Steam Games) */
            .cover-art-backdrop {
                display: none;
                position: absolute;
                inset: -20px;
                background-size: cover;
                background-position: center;
                filter: blur(22px) saturate(1.4) brightness(0.35);
                z-index: 1;
                pointer-events: none;
            }
            .exhibit-frame.has-horizontal-banner {
                display: flex;
                align-items: center;
                justify-content: center;
                background: radial-gradient(circle at center, #181b2a 0%, #090a10 100%);
            }
            .exhibit-frame.has-horizontal-banner .cover-art-backdrop {
                display: block;
            }
            .exhibit-frame.has-horizontal-banner img.cover-art-img {
                width: 100%;
                height: auto;
                max-height: 80%;
                object-fit: contain;
                border-radius: 4px;
                box-shadow: 0 10px 24px rgba(0,0,0,0.8);
            }

            .exhibit-gradient {
                position: absolute;
                inset: 0;
                border-radius: 7px;
                background: linear-gradient(to top, rgba(13, 13, 13, 0.95) 0%, rgba(13, 13, 13, 0.4) 25%, transparent 55%);
                pointer-events: none;
                z-index: 3;
            }

            .exhibit-info {
                padding-top: 10px;
            }
            .exhibit-card-title {
                font-size: 1.2rem;
                margin: 0 0 3px 0;
                color: #fff;
                font-weight: 400;
                letter-spacing: -0.01em;
                line-height: 1.2;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            /* Arcade Cards (Tactile, Clean, Professional) */
            .arcade-curation-tile {
                background: #0d0f17;
                border: 1px solid rgba(255, 255, 255, 0.08);
                border-radius: 8px;
                padding: 26px 28px;
                cursor: pointer;
                transition: border-color 0.25s ease, transform 0.25s ease, box-shadow 0.25s ease;
                display: flex;
                flex-direction: column;
                justify-content: space-between;
                min-height: 190px;
                box-shadow: 0 4px 16px rgba(0,0,0,0.35);
            }
            .arcade-curation-tile:hover {
                border-color: rgba(255, 255, 255, 0.22);
                transform: translateY(-3px);
                box-shadow: 0 12px 32px rgba(0,0,0,0.6);
            }
            .arcade-curation-tile:hover .arcade-launch-link {
                color: #38bdf8;
            }
        </style>

        <div class="editorial-profile-container">

            <!-- SECTION 1: IDENTITY CARD (All Real Telemetry & Dynamic Hardware) -->
            <div class="id-framed-card">
                <div class="crop-corner crop-tl"></div>
                <div class="crop-corner crop-tr"></div>
                <div class="crop-corner crop-bl"></div>
                <div class="crop-corner crop-br"></div>

                <div style="display: flex; align-items: center; gap: 26px; flex-wrap: wrap;">
                    <!-- Avatar Portrait with Choose Colorway Feature -->
                    <div id="btn-open-colorways" style="position: relative; width: 84px; height: 84px; border-radius: 50%; padding: 2px; border: 1.5px solid rgba(255,255,255,0.25); flex-shrink: 0; cursor: pointer;" title="Click to Change Avatar Color">
                        <div id="profile-page-avatar" style="width: 100%; height: 100%; border-radius: 50%; background: ${currentAvatarColor.bg}; display: flex; align-items: center; justify-content: center; font-size: 2rem; font-weight: 800; color: ${currentAvatarColor.textColor}; transition: background 0.2s;">
                            ${escapeHTML(realStats.alias.charAt(0).toUpperCase())}
                        </div>
                        <div style="position: absolute; bottom: 0; right: 0; background: #0d0f17; border: 1px solid rgba(255,255,255,0.25); border-radius: 12px; padding: 2px 6px; font-size: 0.6rem; font-weight: 700; color: #fff; font-family: var(--font-mono);">
                            COLOR
                        </div>
                    </div>

                    <!-- Operator Typography -->
                    <div>
                        <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 4px; flex-wrap: wrap;">
                            <span style="font-size: 1.8rem; font-weight: 800; color: #fff; letter-spacing: -0.02em;">
                                ${escapeHTML(realStats.alias)}
                            </span>
                            <span class="color-block-chip chip-amber">
                                <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon></svg>
                                ${escapeHTML(realStats.clearance)}
                            </span>
                        </div>

                        <div style="font-family: var(--font-mono); font-size: 0.8rem; color: rgba(255,255,255,0.45); margin-bottom: 14px;">
                            @${escapeHTML(realStats.alias.toLowerCase().replace(/[^a-z0-9_]/g, ''))} · ID: ${escapeHTML(realStats.rift_id || "RIFT-F-000000")}
                        </div>

                        <!-- Real Hardware Specs (Real Chip from sysctl, Real OS from sw_vers) -->
                        <div style="display: flex; gap: 8px; flex-wrap: wrap;">
                            <span class="color-block-chip chip-muted" title="Hardware CPU">
                                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
                                ${escapeHTML(realStats.chip || 'Apple M3')}
                            </span>
                            <span class="color-block-chip chip-muted" title="Operating System">
                                <svg width="12" height="12" viewBox="0 0 170 170" fill="currentColor"><path d="M150.37 130.25c-2.45 5.66-5.35 10.87-8.71 15.66-4.58 6.53-8.33 11.05-11.22 13.56-4.48 4.12-9.28 6.23-14.42 6.35-3.69 0-8.14-1.05-13.32-3.18-5.19-2.12-9.97-3.17-14.34-3.17-4.58 0-9.49 1.05-14.75 3.17-5.26 2.13-9.5 3.24-12.74 3.35-4.35.13-9.16-1.9-14.42-6.08-3.7-3.04-7.7-7.85-12.01-14.42-5.74-8.8-10.23-18.78-13.48-29.93-3.26-11.16-4.89-21.84-4.89-32.04 0-14.35 3.7-26.31 11.1-35.91 7.41-9.59 16.75-14.44 28.04-14.54 4.57 0 9.79 1.14 15.65 3.42 5.87 2.28 9.79 3.47 11.75 3.58 1.53 0 5.68-1.25 12.47-3.75 6.78-2.5 12.44-3.64 16.98-3.42 12.63.65 22.84 5.38 30.63 14.19-11.08 6.74-16.52 16.03-16.3 27.87.22 9.57 3.97 17.61 11.27 24.12 7.3 6.52 16.05 10.27 26.24 11.25-2.18 6.74-4.78 13.37-7.83 19.89zM119.22 31.84c0-7.39 2.66-14.24 7.99-20.55 5.33-6.3 11.85-10.33 19.57-12.09.65 1.52.98 3.15.98 4.89 0 7.39-2.77 14.3-8.32 20.71-5.54 6.41-12.18 10.37-19.92 11.87-.22-1.63-.3-3.24-.3-4.83z"/></svg>
                                ${escapeHTML(realStats.os_target || 'macOS 26.6.2')}
                            </span>
                            <span class="color-block-chip chip-muted" title="Member Since">
                                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                                Registered ${escapeHTML(realStats.join_date)}
                            </span>
                        </div>
                    </div>
                </div>

                <!-- Right Brand Callout (v1.0) -->
                <div style="text-align: right; border-left: 1px solid rgba(255,255,255,0.08); padding-left: 24px;">
                    <div style="font-family: var(--font-serif); font-style: italic; font-size: 1.25rem; color: rgba(255,255,255,0.9); line-height: 1.3;">
                        "Every frame,<br><em>translated.</em>"
                    </div>
                    <div style="font-family: var(--font-mono); font-size: 0.68rem; color: rgba(255,255,255,0.35); text-transform: uppercase; margin-top: 6px; letter-spacing: 0.08em;">
                        RIFT Engine v1.0
                    </div>
                </div>
            </div>

            <!-- SECTION 2: THE PLAYTIME RETROSPECTIVE (Real Telemetry from ~/.rift/data/playtime.json) -->
            <div class="editorial-section-title">
                <div class="title-main">
                    <em>The</em> Playtime Retrospective
                </div>
                <div style="font-family: var(--font-mono); font-size: 0.72rem; color: rgba(255,255,255,0.4); text-transform: uppercase;">Real-Time Telemetry</div>
            </div>

            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; margin-bottom: 36px;">
                
                <!-- Stat 1: Global Total Hours -->
                <div class="retrospective-stat-card">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <span style="font-family: var(--font-mono); font-size: 0.72rem; color: rgba(255,255,255,0.5); text-transform: uppercase; letter-spacing: 0.05em;">Global Total</span>
                        <span class="color-block-chip chip-cyan" style="padding: 2px 6px; font-size: 0.65rem;">ACTIVE</span>
                    </div>
                    <div>
                        <div class="stat-metric-val">${playtimeTotalHours}</div>
                        <div style="font-size: 0.75rem; color: rgba(255,255,255,0.45); margin-top: 4px; font-weight: 500;">Hours on Engine</div>
                    </div>
                </div>

                <!-- Stat 2: Titles Played -->
                <div class="retrospective-stat-card">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <span style="font-family: var(--font-mono); font-size: 0.72rem; color: rgba(255,255,255,0.5); text-transform: uppercase; letter-spacing: 0.05em;">Titles Played</span>
                        <span class="color-block-chip chip-muted" style="padding: 2px 6px; font-size: 0.65rem;">LOGGED</span>
                    </div>
                    <div>
                        <div class="stat-metric-val">${titlesPlayedCount}</div>
                        <div style="font-size: 0.75rem; color: rgba(255,255,255,0.45); margin-top: 4px; font-weight: 500;">Executed Titles</div>
                    </div>
                </div>

                <!-- Stat 3: Archive Vault -->
                <div class="retrospective-stat-card">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <span style="font-family: var(--font-mono); font-size: 0.72rem; color: rgba(255,255,255,0.5); text-transform: uppercase; letter-spacing: 0.05em;">Vault Archive</span>
                        <span class="color-block-chip chip-muted" style="padding: 2px 6px; font-size: 0.65rem;">SYNCED</span>
                    </div>
                    <div>
                        <div class="stat-metric-val">${gamesOwnedCount}</div>
                        <div style="font-size: 0.75rem; color: rgba(255,255,255,0.45); margin-top: 4px; font-weight: 500;">Owned Licenses</div>
                    </div>
                </div>

                <!-- Stat 4: Translation Core (Dynamic Telemetry: Last Played Game) -->
                <div class="retrospective-stat-card">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <span style="font-family: var(--font-mono); font-size: 0.72rem; color: rgba(255,255,255,0.5); text-transform: uppercase; letter-spacing: 0.05em;">Translation Core</span>
                        <span class="color-block-chip chip-cyan" style="padding: 2px 6px; font-size: 0.65rem;">${escapeHTML(lastCoreData.core_badge || 'D3DMETAL')}</span>
                    </div>
                    <div>
                        <div style="font-size: 1.45rem; font-weight: 800; color: #fff; margin-top: 14px; line-height: 1.1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;" title="${escapeHTML(lastCoreData.core_title)}">
                            ${escapeHTML(lastCoreData.core_title || 'D3DMetal (GPTK)')}
                        </div>
                        <div style="font-size: 0.75rem; color: #38bdf8; margin-top: 5px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;" title="Used on ${escapeHTML(lastCoreData.game_name || lastCoreData.subtitle)}">
                            for ${escapeHTML(lastCoreData.game_name || lastCoreData.subtitle || 'TrackMania Nations Forever')}
                        </div>
                        <div style="font-size: 0.72rem; color: rgba(255,255,255,0.45); margin-top: 3px; font-weight: 500;">
                            ${escapeHTML(lastCoreData.layer_details || 'DirectX 9/11/12 → Metal 4')}
                        </div>
                    </div>
                </div>

            </div>

            <!-- SECTION 3: TOP EXECUTED ARTIFACTS (Exact Epic & Custom Artifacts Style with Full Vertical Covers) -->
            <div class="editorial-section-title">
                <div class="title-main">
                    <em>Top</em> Executed Artifacts
                </div>
                <a href="#/library" style="font-family: var(--font-mono); font-size: 0.75rem; color: #38bdf8; text-decoration: none; letter-spacing: 0.05em; text-transform: uppercase; display: flex; align-items: center; gap: 4px;">
                    Full Archive
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
                </a>
            </div>

            <div class="profile-exhibits-grid">
                ${topPlayedGames.map((game, idx) => {
                    const isEpic = game.platform && game.platform.toLowerCase() === 'epic';
                    const isSteam = (game.platform && game.platform.toLowerCase() === 'steam') || (game.id && game.id.startsWith('steam-'));
                    const appId = game.appID || (game.id && game.id.startsWith('steam-') ? game.id.replace('steam-', '') : '');
                    
                    // High-res uncropped vertical box art (Steam library_600x900 CDN for Steam games, cover for Epic)
                    const steamVerticalUrl = (isSteam && appId) ? `https://steamcdn-a.akamaihd.net/steam/apps/${appId}/library_600x900.jpg` : '';
                    const steamLandscapeUrl = (isSteam && appId) ? `https://steamcdn-a.akamaihd.net/steam/apps/${appId}/header.jpg` : '';
                    const initialCover = steamVerticalUrl || game.cover || game.heroCover || '';
                    const fallbackCover = steamLandscapeUrl || game.cover || '';

                    const seconds = playtimeMap[game.id] || playtimeMap[game.appID] || 0;
                    const hoursText = seconds > 0 ? (seconds / 3600).toFixed(1) + ' hrs' : (game.playtime || 'Ready');
                    const exhibitNum = String(idx + 1).padStart(2, '0');
                    const safeName = escapeHTML(game.name || 'Unknown Title');
                    const safeBackend = escapeHTML(game.backend || (game.isMacNative ? 'macOS Native' : 'D3DMetal + WoW64'));

                    // Clean SVG Platform Icons (zero emojis)
                    const platformIcon = isEpic ? `
                        <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor" style="flex-shrink: 0;"><path d="M3.5 0h17l-3.5 19-5 5-5-5L3.5 0zm8.5 3.8v16.1l3.3-3.3 2.3-12.8H12z"/></svg>
                    ` : `
                        <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor" style="flex-shrink: 0;"><path d="M12 2a10 10 0 0 0-10 9.77c0 2.2.73 4.23 1.96 5.86l4.24-1.74a3.5 3.5 0 0 1 2.3-3.89V12a3.5 3.5 0 0 1 3.5-3.5 3.5 3.5 0 0 1 3.5 3.5 3.5 3.5 0 0 1-3.5 3.5h-.05l-2.6 3.65c-.27.05-.56.08-.85.08a4.5 4.5 0 0 1-4.46-3.89l-4.13 1.7A10 10 0 1 0 12 2z"/></svg>
                    `;

                    const platformChip = isEpic ? 'chip-amber' : 'chip-cyan';

                    return `
                        <div class="exhibit-card" data-id="${game.id}" data-platform="${game.platform || 'Steam'}" data-appid="${appId}">
                            <!-- Exact Double-Framed Poster matching Epic & Custom Artifacts -->
                            <div class="exhibit-frame">
                                <div class="cover-art-backdrop" style="background-image: url('${initialCover}');"></div>
                                <img 
                                    src="${initialCover}" 
                                    alt="${safeName}" 
                                    class="cover-art-img" 
                                    loading="lazy" 
                                    data-fallback="${fallbackCover}"
                                    data-gamename="${safeName}"
                                    onload="window.handleCoverLoad && window.handleCoverLoad(this)"
                                    onerror="window.handleCoverErrorCustom && window.handleCoverErrorCustom(this)"
                                />
                                <div class="exhibit-gradient"></div>
                                
                                <div style="position: absolute; top: 10px; left: 10px; z-index: 4;">
                                    <span class="color-block-chip ${platformChip}" style="font-size: 0.65rem; padding: 2px 7px; backdrop-filter: blur(8px);">
                                        ${platformIcon} ${escapeHTML(game.platform || 'Steam')}
                                    </span>
                                </div>

                            </div>

                            <!-- Simple, Clean Info Below Poster: Just Name & Playtime -->
                            <div class="exhibit-info">
                                <h3 class="exhibit-card-title text-serif-italic" title="${safeName}">
                                    ${safeName}
                                </h3>
                                <div style="font-family: var(--font-mono); font-size: 0.72rem; color: #38bdf8; font-weight: 600; letter-spacing: 0.04em;">
                                    ${escapeHTML(hoursText)}
                                </div>
                            </div>
                        </div>
                    `;
                }).join('')}
            </div>

            <!-- SECTION 4: SYSTEM TRANSMISSIONS -->
            <div class="editorial-section-title">
                <div class="title-main">
                    <em>System</em> Bulletins & Dispatch
                </div>
                <div style="display: flex; align-items: center; gap: 8px;">
                    <button id="btn-refresh-transmissions" style="background: transparent; border: 1px solid rgba(255,255,255,0.15); color: #38bdf8; padding: 4px 12px; border-radius: 4px; font-family: var(--font-mono); font-size: 0.7rem; text-transform: uppercase; cursor: pointer; display: flex; align-items: center; gap: 4px;">
                        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/></svg>
                        Refresh
                    </button>
                    <button id="btn-view-all-transmissions" style="background: transparent; border: 1px solid rgba(255,255,255,0.15); color: #f5f5f5; padding: 4px 12px; border-radius: 4px; font-family: var(--font-mono); font-size: 0.7rem; text-transform: uppercase; cursor: pointer;">
                        View Complete Dispatch
                    </button>
                </div>
            </div>

            <div style="background: #0d0f17; border: 1px solid rgba(255, 255, 255, 0.08); border-left: 3.5px solid #38bdf8; border-radius: 6px; padding: 20px 24px; margin-bottom: 44px;">
                <div id="transmissions-feed" style="display: flex; flex-direction: column; gap: 12px;">
                    <!-- Injected via JS -->
                </div>
            </div>

            <!-- SECTION 5: RIFT ARCADE -->
            <div class="editorial-section-title">
                <div class="title-main">
                    <em>RIFT</em> Interactive Arcade
                </div>
                <div style="font-family: var(--font-mono); font-size: 0.72rem; color: rgba(255,255,255,0.4); text-transform: uppercase;">Native Canvas Simulations</div>
            </div>

            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 24px;">
                
                <!-- Cyber Snake -->
                <div class="arcade-curation-tile" onclick="window.location.hash = '/arcade/snake'">
                    <div>
                        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px;">
                            <span class="color-block-chip chip-muted">SIMULATION · 01</span>
                            <span style="font-family: var(--font-mono); font-size: 0.74rem; color: #38bdf8; font-weight: 600;">BEST: ${snakeHighScore.toLocaleString()} PTS</span>
                        </div>
                        <h2 style="font-family: var(--font-serif); font-style: italic; font-size: 1.65rem; font-weight: 400; color: #fff; margin: 0 0 8px 0;">Cyber Snake</h2>
                        <p style="margin: 0; font-size: 0.86rem; color: rgba(255,255,255,0.6); line-height: 1.55;">
                            High-precision vector arcade simulation. Features dynamic velocity scaling, combo multipliers, and quantum vector steering.
                        </p>
                    </div>

                    <div style="margin-top: 24px; display: flex; justify-content: space-between; align-items: center; border-top: 1px solid rgba(255,255,255,0.06); padding-top: 14px;">
                        <span style="font-family: var(--font-mono); font-size: 0.72rem; color: rgba(255,255,255,0.4);">WASD / ARROW KEYS</span>
                        <div class="arcade-launch-link" style="font-family: var(--font-mono); font-size: 0.78rem; font-weight: 700; color: #fff; display: flex; align-items: center; gap: 6px; transition: color 0.2s;">
                            LAUNCH SIMULATION
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
                        </div>
                    </div>
                </div>

                <!-- Terminal Golf -->
                <div class="arcade-curation-tile" onclick="window.location.hash = '/arcade/golf'">
                    <div>
                        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px;">
                            <span class="color-block-chip chip-muted">SIMULATION · 02</span>
                            <span style="font-family: var(--font-mono); font-size: 0.74rem; color: rgba(255,255,255,0.5); font-weight: 600;">ISOMETRIC 3D</span>
                        </div>
                        <h2 style="font-family: var(--font-serif); font-style: italic; font-size: 1.65rem; font-weight: 400; color: #fff; margin: 0 0 8px 0;">Terminal Golf</h2>
                        <p style="margin: 0; font-size: 0.86rem; color: rgba(255,255,255,0.6); line-height: 1.55;">
                            Isometric putting physics simulation with procedural green friction, topographic elevation curves, and true parabolic trajectories.
                        </p>
                    </div>

                    <div style="margin-top: 24px; display: flex; justify-content: space-between; align-items: center; border-top: 1px solid rgba(255,255,255,0.06); padding-top: 14px;">
                        <span style="font-family: var(--font-mono); font-size: 0.72rem; color: rgba(255,255,255,0.4);">MOUSE DRAG CONTROLS</span>
                        <div class="arcade-launch-link" style="font-family: var(--font-mono); font-size: 0.78rem; font-weight: 700; color: #fff; display: flex; align-items: center; gap: 6px; transition: color 0.2s;">
                            ENTER COURSE
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
                        </div>
                    </div>
                </div>

            </div>

        </div>

        <!-- Colorway Selector Modal (Choose Color Feature) -->
        <div id="colorways-modal" style="display: none; position: fixed; inset: 0; background: rgba(8,8,8,0.88); backdrop-filter: blur(16px); z-index: 1000; align-items: center; justify-content: center;">
            <div style="background: #0d0f17; border: 1px solid rgba(255,255,255,0.12); border-radius: 8px; padding: 32px; width: 480px; max-width: 90vw; box-shadow: 0 24px 64px rgba(0,0,0,0.9);">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; padding-bottom: 14px; border-bottom: 1px solid rgba(255,255,255,0.08);">
                    <div style="font-size: 1.15rem; font-weight: 800; color: #fff;">Choose Profile Colorway</div>
                    <span class="color-block-chip chip-cyan">PALETTE</span>
                </div>

                <div style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 14px;">
                    ${colorways.map(cw => `
                        <div class="colorway-option" data-bg="${cw.bg}" data-border="${cw.border}" data-text="${cw.textColor}" style="display: flex; flex-direction: column; align-items: center; gap: 10px; padding: 14px; border-radius: 6px; background: rgba(255,255,255,0.03); border: 1.5px solid rgba(255,255,255,0.08); cursor: pointer; transition: transform 0.15s, border-color 0.15s;">
                            <div style="width: 44px; height: 44px; border-radius: 50%; background: ${cw.bg}; border: 1.5px solid ${cw.border}; display: flex; align-items: center; justify-content: center; font-weight: 800; font-size: 1.15rem; color: ${cw.textColor};">
                                ${escapeHTML(realStats.alias.charAt(0).toUpperCase())}
                            </div>
                            <span style="font-size: 0.75rem; font-weight: 600; color: #fff;">${cw.name}</span>
                        </div>
                    `).join('')}
                </div>

                <div style="text-align: right; margin-top: 24px; padding-top: 14px; border-top: 1px solid rgba(255,255,255,0.08);">
                    <button id="close-colorways-modal" style="background: transparent; border: 1px solid rgba(255,255,255,0.15); color: #fff; padding: 6px 18px; border-radius: 4px; font-weight: 600; font-size: 0.78rem; cursor: pointer;">
                        Cancel
                    </button>
                </div>
            </div>
        </div>

        <!-- Transmissions Modal -->
        <div id="all-transmissions-modal" style="display: none; position: fixed; inset: 0; background: rgba(8,8,8,0.92); backdrop-filter: blur(16px); z-index: 1000; align-items: center; justify-content: center;">
            <div style="background: #0d0f17; border: 1px solid rgba(255,255,255,0.12); border-radius: 6px; padding: 36px; width: 780px; max-width: 92vw; max-height: 85vh; display: flex; flex-direction: column; box-shadow: 0 24px 64px rgba(0,0,0,0.9);">
                <div style="display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 20px; padding-bottom: 16px; border-bottom: 1px solid rgba(255,255,255,0.08);">
                    <div>
                        <div style="font-family: var(--font-serif); font-style: italic; font-size: 1.8rem; color: #fff;">System Bulletins & Dispatch</div>
                        <div style="font-family: var(--font-mono); font-size: 0.75rem; color: rgba(255,255,255,0.45); margin-top: 4px;">Engine Engineering Logs · RIFT Cloud</div>
                    </div>
                </div>
                
                <div id="modal-transmissions-list" style="flex-grow: 1; overflow-y: auto; display: flex; flex-direction: column; gap: 16px; padding-right: 12px;">
                    <!-- Populated by JS -->
                </div>
                
                <div style="text-align: right; margin-top: 20px; padding-top: 16px; border-top: 1px solid rgba(255,255,255,0.08);">
                    <button id="close-transmissions-modal" style="background: #38bdf8; color: #080808; border: none; padding: 9px 22px; border-radius: 4px; font-weight: 700; font-size: 0.8rem; cursor: pointer;">
                        Close Dispatch
                    </button>
                </div>
            </div>
        </div>
    `;

    // Setup interactive handlers
    setTimeout(() => {
        // Colorway Modal
        const btnOpenColorways = container.querySelector('#btn-open-colorways');
        const colorwaysModal = container.querySelector('#colorways-modal');
        const closeColorwaysModal = container.querySelector('#close-colorways-modal');
        const profileAvatar = container.querySelector('#profile-page-avatar');
        const topBarAvatar = document.getElementById('top-bar-avatar');

        if (btnOpenColorways && colorwaysModal) {
            btnOpenColorways.addEventListener('click', () => colorwaysModal.style.display = 'flex');
        }
        if (closeColorwaysModal && colorwaysModal) {
            closeColorwaysModal.addEventListener('click', () => colorwaysModal.style.display = 'none');
        }

        const options = container.querySelectorAll('.colorway-option');
        options.forEach(opt => {
            opt.addEventListener('click', () => {
                const bg = opt.dataset.bg;
                const border = opt.dataset.border;
                const textColor = opt.dataset.text;
                const selected = { bg, border, textColor };

                if (profileAvatar) {
                    profileAvatar.style.background = bg;
                    profileAvatar.style.color = textColor;
                }
                if (topBarAvatar) {
                    topBarAvatar.style.background = bg;
                    topBarAvatar.style.color = textColor;
                }

                localStorage.setItem('rift_avatar_color', JSON.stringify(selected));
                store.update('profile', { ...store.state.profile, avatarColor: selected });
                colorwaysModal.style.display = 'none';
            });
        });

        // Card Click: Safely navigate to Game Detail page (Never start downloads on click)
        const cards = container.querySelectorAll('.exhibit-card');
        cards.forEach(card => {
            card.addEventListener('click', () => {
                const gameId = card.dataset.id;
                if (gameId) {
                    window.location.hash = `/game?id=${encodeURIComponent(gameId)}`;
                }
            });
        });

        // Ensure global profile state is synced with real alias
        store.update('profile', {
            ...store.state.profile,
            name: realStats.alias,
            alias: realStats.alias,
            username: realStats.alias.toLowerCase(),
            avatarColor: currentAvatarColor
        });

        // Check already-loaded / cached covers for horizontal aspect ratio
        container.querySelectorAll('.cover-art-img').forEach(img => {
            if (img.complete && img.naturalWidth > 0) {
                window.handleCoverLoad(img);
            }
        });

        // Transmissions Management
        const feedContainer = container.querySelector('#transmissions-feed');
        const viewAllBtn = container.querySelector('#btn-view-all-transmissions');
        const refreshBtn = container.querySelector('#btn-refresh-transmissions');
        const transModal = container.querySelector('#all-transmissions-modal');
        const closeTransModal = container.querySelector('#close-transmissions-modal');
        const modalList = container.querySelector('#modal-transmissions-list');

        function renderTransmissionsFeed() {
            const transmissions = realStats.transmissions || [];
            if (!feedContainer) return;

            if (transmissions.length > 0) {
                feedContainer.innerHTML = transmissions.slice(0, 3).map((t) => {
                    const cleanSnippet = t.content.replace(/\\n/g, ' ').replace(/[*#]/g, '').slice(0, 130);
                    return `
                        <div style="padding: 12px 0; border-bottom: 1px solid rgba(255,255,255,0.05); cursor: pointer;" class="transmission-row">
                            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px;">
                                <span class="color-block-chip chip-cyan" style="padding: 2px 6px; font-size: 0.65rem;">
                                    ${escapeHTML(t.version_tag || 'DISPATCH')}
                                </span>
                                <span style="font-family: var(--font-mono); font-size: 0.72rem; color: rgba(255,255,255,0.4);">${escapeHTML(t.date)}</span>
                            </div>
                            <div style="font-weight: 700; font-size: 0.95rem; color: #fff; margin-bottom: 2px;">${escapeHTML(t.title)}</div>
                            <div style="font-size: 0.8rem; color: rgba(255,255,255,0.6); line-height: 1.4;">${escapeHTML(cleanSnippet)}...</div>
                        </div>
                    `;
                }).join('');

                container.querySelectorAll('.transmission-row').forEach(row => {
                    row.addEventListener('click', () => {
                        if (transModal) transModal.style.display = 'flex';
                        renderAllTransmissions();
                    });
                });
            } else {
                feedContainer.innerHTML = `
                    <div style="padding: 12px 0; color: rgba(255,255,255,0.4); font-size: 0.82rem; font-style: italic;">
                        No system transmissions recorded.
                    </div>
                `;
            }
        }

        function renderAllTransmissions() {
            if (!modalList) return;
            const transmissions = realStats.transmissions || [];
            modalList.innerHTML = transmissions.map(t => `
                <div style="padding: 16px 0; border-bottom: 1px solid rgba(255,255,255,0.06);">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                        <span class="color-block-chip chip-cyan">
                            ${escapeHTML(t.version_tag || 'DISPATCH')}
                        </span>
                        <span style="font-family: var(--font-mono); font-size: 0.75rem; color: rgba(255,255,255,0.4);">${escapeHTML(t.date)}</span>
                    </div>
                    <div style="font-family: var(--font-serif); font-size: 1.25rem; color: #fff; margin-bottom: 6px;">${escapeHTML(t.title)}</div>
                    <div style="font-size: 0.85rem; color: rgba(255,255,255,0.75); line-height: 1.5;">
                        ${parseMarkdown(t.content)}
                    </div>
                </div>
            `).join('');
        }

        renderTransmissionsFeed();

        if (refreshBtn) {
            refreshBtn.addEventListener('click', async () => {
                refreshBtn.disabled = true;
                refreshBtn.innerHTML = `Syncing...`;
                try {
                    if (window.go && window.go.rift && window.go.rift.App && window.go.rift.App.GetUserProfileData) {
                        const statsJson = await window.go.rift.App.GetUserProfileData();
                        const parsed = JSON.parse(statsJson);
                        if (parsed && parsed.transmissions) {
                            realStats.transmissions = parsed.transmissions;
                            renderTransmissionsFeed();
                        }
                    }
                    refreshBtn.innerHTML = `✓ Synced`;
                    setTimeout(() => {
                        refreshBtn.innerHTML = `<svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/></svg> Refresh`;
                        refreshBtn.disabled = false;
                    }, 1500);
                } catch (err) {
                    refreshBtn.innerHTML = `Error`;
                    setTimeout(() => {
                        refreshBtn.innerHTML = `<svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/></svg> Refresh`;
                        refreshBtn.disabled = false;
                    }, 2000);
                }
            });
        }

        if (viewAllBtn && transModal) {
            viewAllBtn.addEventListener('click', () => {
                transModal.style.display = 'flex';
                renderAllTransmissions();
            });
        }

        if (closeTransModal && transModal) {
            closeTransModal.addEventListener('click', () => transModal.style.display = 'none');
        }
    }, 0);

    return container;
}
