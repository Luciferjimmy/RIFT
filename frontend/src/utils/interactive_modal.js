import { appendLog } from '../main.js';

let modalContainer = null;

function ensureModalContainer() {
    if (!modalContainer) {
        modalContainer = document.createElement('div');
        modalContainer.id = 'interactive-modal-container';
        document.body.appendChild(modalContainer);

        // Inject modal styles
        const styleId = 'interactive-modal-css';
        if (!document.getElementById(styleId)) {
            const style = document.createElement('style');
            style.id = styleId;
            style.textContent = `
                .rift-modal-overlay {
                    position: fixed;
                    inset: 0;
                    background: rgba(0, 0, 0, 0.75);
                    backdrop-filter: blur(16px);
                    -webkit-backdrop-filter: blur(16px);
                    z-index: 9999;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    padding: 24px;
                    animation: riftFadeIn 0.2s ease-out;
                }
                @keyframes riftFadeIn {
                    from { opacity: 0; }
                    to { opacity: 1; }
                }
                .rift-modal-card {
                    background: #0d0e14;
                    border: 1px solid rgba(255, 255, 255, 0.1);
                    border-radius: 16px;
                    width: 100%;
                    max-width: 580px;
                    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.7), 0 0 0 1px rgba(255, 255, 255, 0.05);
                    overflow: hidden;
                    display: flex;
                    flex-direction: column;
                    transform: translateY(0);
                    animation: riftSlideUp 0.25s cubic-bezier(0.16, 1, 0.3, 1);
                }
                @keyframes riftSlideUp {
                    from { transform: translateY(16px) scale(0.98); }
                    to { transform: translateY(0) scale(1); }
                }
                .rift-modal-header {
                    padding: 24px 28px 20px;
                    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
                    display: flex;
                    align-items: flex-start;
                    justify-content: space-between;
                }
                .rift-modal-badge {
                    display: inline-flex;
                    align-items: center;
                    gap: 6px;
                    background: rgba(56, 189, 248, 0.12);
                    border: 1px solid rgba(56, 189, 248, 0.3);
                    color: #38bdf8;
                    padding: 4px 10px;
                    border-radius: 6px;
                    font-size: 0.72rem;
                    font-weight: 700;
                    letter-spacing: 0.5px;
                    font-family: 'JetBrains Mono', monospace;
                    text-transform: uppercase;
                    margin-bottom: 8px;
                }
                .rift-modal-title {
                    font-family: 'Plus Jakarta Sans', sans-serif;
                    font-weight: 700;
                    font-size: 1.4rem;
                    color: #fff;
                    margin: 0;
                    letter-spacing: -0.3px;
                }
                .rift-modal-close {
                    background: transparent;
                    border: none;
                    color: rgba(255, 255, 255, 0.4);
                    font-size: 1.25rem;
                    cursor: pointer;
                    padding: 4px 8px;
                    border-radius: 6px;
                    transition: all 0.15s;
                }
                .rift-modal-close:hover {
                    color: #fff;
                    background: rgba(255, 255, 255, 0.08);
                }
                .rift-modal-body {
                    padding: 24px 28px;
                    max-height: 420px;
                    overflow-y: auto;
                    display: flex;
                    flex-direction: column;
                    gap: 16px;
                }
                .rift-pack-card {
                    background: rgba(255, 255, 255, 0.02);
                    border: 1px solid rgba(255, 255, 255, 0.06);
                    border-radius: 10px;
                    padding: 14px 16px;
                    display: flex;
                    align-items: flex-start;
                    gap: 14px;
                    cursor: pointer;
                    transition: all 0.2s ease;
                }
                .rift-pack-card:hover {
                    background: rgba(255, 255, 255, 0.04);
                    border-color: rgba(255, 255, 255, 0.12);
                }
                .rift-pack-card.selected {
                    background: rgba(56, 189, 248, 0.06);
                    border-color: rgba(56, 189, 248, 0.35);
                }
                .rift-pack-card input[type="checkbox"] {
                    margin-top: 3px;
                    width: 18px;
                    height: 18px;
                    accent-color: #38bdf8;
                    cursor: pointer;
                }
                .rift-modal-footer {
                    padding: 20px 28px 24px;
                    border-top: 1px solid rgba(255, 255, 255, 0.06);
                    display: flex;
                    align-items: center;
                    justify-content: flex-end;
                    gap: 12px;
                    background: rgba(0, 0, 0, 0.2);
                }
                .rift-btn-secondary {
                    background: rgba(255, 255, 255, 0.05);
                    border: 1px solid rgba(255, 255, 255, 0.1);
                    color: rgba(255, 255, 255, 0.8);
                    padding: 10px 18px;
                    border-radius: 8px;
                    font-family: 'Inter', sans-serif;
                    font-size: 0.86rem;
                    font-weight: 600;
                    cursor: pointer;
                    transition: all 0.2s;
                }
                .rift-btn-secondary:hover {
                    background: rgba(255, 255, 255, 0.1);
                    color: #fff;
                }
                .rift-btn-primary {
                    background: linear-gradient(135deg, #38bdf8 0%, #0284c7 100%);
                    border: 1px solid rgba(56, 189, 248, 0.4);
                    color: #ffffff;
                    padding: 10px 22px;
                    border-radius: 8px;
                    font-family: 'Inter', sans-serif;
                    font-size: 0.86rem;
                    font-weight: 600;
                    cursor: pointer;
                    box-shadow: 0 4px 16px rgba(2, 132, 199, 0.35);
                    transition: all 0.2s;
                }
                .rift-btn-primary:hover {
                    background: linear-gradient(135deg, #60a5fa 0%, #0369a1 100%);
                    box-shadow: 0 6px 20px rgba(56, 189, 248, 0.45);
                    transform: translateY(-1px);
                }
            `;
            document.head.appendChild(style);
        }
    }
}

/**
 * Shows the Epic Selective Download (SDL) Package Customizer Modal.
 * Allows the user to choose optional DLC / HD Texture packs.
 */
export function showEpicSelectiveDownloadModal(data) {
    ensureModalContainer();

    const { gameId, appID, gameName, packs } = data;
    const packList = packs || [];

    const overlay = document.createElement('div');
    overlay.className = 'rift-modal-overlay';

    let packsHTML = `
        <div class="rift-pack-card selected" style="cursor: default; opacity: 0.9;">
            <input type="checkbox" checked disabled>
            <div>
                <div style="color: #fff; font-weight: 600; font-size: 0.95rem;">Base Game Files</div>
                <div style="color: rgba(255,255,255,0.5); font-size: 0.8rem; margin-top: 2px;">Core game binary and mandatory assets (Required)</div>
            </div>
        </div>
    `;

    packList.forEach((pack, idx) => {
        packsHTML += `
            <label class="rift-pack-card" id="pack-card-${idx}">
                <input type="checkbox" class="rift-sdl-pack-cb" data-tag="${pack.tag}">
                <div>
                    <div style="display: flex; align-items: center; gap: 8px;">
                        <span style="color: #fff; font-weight: 600; font-size: 0.95rem;">${pack.name || pack.tag}</span>
                        <span style="background: rgba(255,255,255,0.08); color: rgba(255,255,255,0.6); padding: 1px 6px; border-radius: 4px; font-size: 0.7rem; font-family: 'JetBrains Mono', monospace;">OPTIONAL</span>
                    </div>
                    <div style="color: rgba(255,255,255,0.55); font-size: 0.82rem; margin-top: 3px; line-height: 1.4;">${pack.description || 'Additional package data'}</div>
                </div>
            </label>
        `;
    });

    overlay.innerHTML = `
        <div class="rift-modal-card">
            <div class="rift-modal-header">
                <div>
                    <div class="rift-modal-badge">
                        <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
                        Epic Games • Selective Download
                    </div>
                    <h2 class="rift-modal-title">${gameName || 'Configure Download'}</h2>
                </div>
                <button class="rift-modal-close" id="rift-modal-btn-close">&times;</button>
            </div>

            <div class="rift-modal-body">
                <div style="color: rgba(255,255,255,0.65); font-size: 0.85rem; line-height: 1.45;">
                    This game includes optional component packages. Select which additional packages you want to install:
                </div>
                ${packsHTML}
            </div>

            <div class="rift-modal-footer">
                <button class="rift-btn-secondary" id="rift-modal-btn-base-only">Download Base Only</button>
                <button class="rift-btn-primary" id="rift-modal-btn-download-selected">Download Selected</button>
            </div>
        </div>
    `;

    modalContainer.appendChild(overlay);

    // Toggle card highlighted style on checkbox change
    overlay.querySelectorAll('.rift-sdl-pack-cb').forEach(cb => {
        cb.addEventListener('change', (e) => {
            const card = e.target.closest('.rift-pack-card');
            if (card) {
                if (e.target.checked) card.classList.add('selected');
                else card.classList.remove('selected');
            }
        });
    });

    const close = () => {
        overlay.remove();
    };

    overlay.querySelector('#rift-modal-btn-close')?.addEventListener('click', close);

    // Download Base Only (empty tags)
    overlay.querySelector('#rift-modal-btn-base-only')?.addEventListener('click', async () => {
        close();
        appendLog(`[Download] Initiating Base Game download for ${gameName}...`, 'log-info');
        try {
            await window.go.rift.App.DownloadEpicWithTags(gameId, appID, []);
        } catch (err) {
            appendLog(`[Download] Failed to start download: ${err}`, 'log-error');
        }
    });

    // Download with selected tags
    overlay.querySelector('#rift-modal-btn-download-selected')?.addEventListener('click', async () => {
        const selectedTags = [];
        overlay.querySelectorAll('.rift-sdl-pack-cb:checked').forEach(cb => {
            const tag = cb.dataset.tag;
            if (tag) selectedTags.push(tag);
        });

        close();
        appendLog(`[Download] Initiating download for ${gameName} with tags: ${selectedTags.join(', ') || 'Base Only'}...`, 'log-info');
        try {
            await window.go.rift.App.DownloadEpicWithTags(gameId, appID, selectedTags);
        } catch (err) {
            appendLog(`[Download] Failed to start download: ${err}`, 'log-error');
        }
    });
}

/**
 * Shows the In-App Security OTP Password Change Modal.
 * Dispatches a 6-digit OTP to userEmail and prompts for code + new password.
 */
export async function showChangePasswordModal(userEmail) {
    ensureModalContainer();

    const overlay = document.createElement('div');
    overlay.className = 'rift-modal-overlay';
    overlay.style.background = 'rgba(4, 6, 12, 0.85)';

    overlay.innerHTML = `
        <div class="rift-modal-card" style="max-width: 460px; background: linear-gradient(180deg, #111420 0%, #0a0c14 100%); border: 1px solid rgba(255, 255, 255, 0.12); box-shadow: 0 32px 80px rgba(0, 0, 0, 0.85), 0 0 0 1px rgba(56, 189, 248, 0.15); border-radius: 20px;">
            <div class="rift-modal-header" style="padding: 24px 28px 18px; border-bottom: 1px solid rgba(255, 255, 255, 0.07); display: flex; align-items: flex-start; justify-content: space-between;">
                <div>
                    <div style="display: inline-flex; align-items: center; gap: 6px; padding: 3px 10px; border-radius: 20px; background: rgba(56, 189, 248, 0.1); border: 1px solid rgba(56, 189, 248, 0.25); color: #38bdf8; font-size: 0.7rem; font-weight: 700; letter-spacing: 1px; text-transform: uppercase; margin-bottom: 8px;">
                        <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
                        Archive Security
                    </div>
                    <h2 style="font-family: 'Plus Jakarta Sans', sans-serif; font-weight: 800; font-size: 1.35rem; color: #fff; margin: 0; letter-spacing: -0.5px;">Update Master Password</h2>
                </div>
                <button class="rift-modal-close" id="pw-modal-close" style="background: rgba(255,255,255,0.06); border: none; color: rgba(255,255,255,0.6); width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; cursor: pointer; font-size: 18px; transition: all 0.2s;">&times;</button>
            </div>

            <div class="rift-modal-body" style="padding: 24px 28px; display: flex; flex-direction: column; gap: 20px;">
                <div id="pw-modal-banner" style="display: none; padding: 12px 16px; border-radius: 10px; font-size: 0.83rem; line-height: 1.45;"></div>
                
                <div style="color: rgba(255,255,255,0.6); font-size: 0.86rem; line-height: 1.5;">
                    Enter the 6-digit security code dispatched to <span style="color: #fff; font-weight: 600;">${userEmail}</span>:
                </div>

                <!-- 6 Individual OTP Digit Inputs -->
                <div>
                    <label style="display: block; font-size: 0.72rem; text-transform: uppercase; letter-spacing: 1px; color: #38bdf8; font-weight: 700; margin-bottom: 10px;">Security Verification Code</label>
                    <div style="display: flex; gap: 8px; justify-content: space-between;" id="otp-digit-container">
                        <input type="text" maxlength="1" class="otp-box" data-index="0" style="width: 50px; height: 56px; background: rgba(0,0,0,0.4); border: 1px solid rgba(255,255,255,0.12); border-radius: 12px; color: #38bdf8; font-size: 1.5rem; font-weight: 800; text-align: center; font-family: 'JetBrains Mono', monospace; outline: none; transition: all 0.2s;" autocomplete="off" />
                        <input type="text" maxlength="1" class="otp-box" data-index="1" style="width: 50px; height: 56px; background: rgba(0,0,0,0.4); border: 1px solid rgba(255,255,255,0.12); border-radius: 12px; color: #38bdf8; font-size: 1.5rem; font-weight: 800; text-align: center; font-family: 'JetBrains Mono', monospace; outline: none; transition: all 0.2s;" autocomplete="off" />
                        <input type="text" maxlength="1" class="otp-box" data-index="2" style="width: 50px; height: 56px; background: rgba(0,0,0,0.4); border: 1px solid rgba(255,255,255,0.12); border-radius: 12px; color: #38bdf8; font-size: 1.5rem; font-weight: 800; text-align: center; font-family: 'JetBrains Mono', monospace; outline: none; transition: all 0.2s;" autocomplete="off" />
                        <input type="text" maxlength="1" class="otp-box" data-index="3" style="width: 50px; height: 56px; background: rgba(0,0,0,0.4); border: 1px solid rgba(255,255,255,0.12); border-radius: 12px; color: #38bdf8; font-size: 1.5rem; font-weight: 800; text-align: center; font-family: 'JetBrains Mono', monospace; outline: none; transition: all 0.2s;" autocomplete="off" />
                        <input type="text" maxlength="1" class="otp-box" data-index="4" style="width: 50px; height: 56px; background: rgba(0,0,0,0.4); border: 1px solid rgba(255,255,255,0.12); border-radius: 12px; color: #38bdf8; font-size: 1.5rem; font-weight: 800; text-align: center; font-family: 'JetBrains Mono', monospace; outline: none; transition: all 0.2s;" autocomplete="off" />
                        <input type="text" maxlength="1" class="otp-box" data-index="5" style="width: 50px; height: 56px; background: rgba(0,0,0,0.4); border: 1px solid rgba(255,255,255,0.12); border-radius: 12px; color: #38bdf8; font-size: 1.5rem; font-weight: 800; text-align: center; font-family: 'JetBrains Mono', monospace; outline: none; transition: all 0.2s;" autocomplete="off" />
                    </div>
                </div>

                <!-- New Password Field -->
                <div>
                    <label style="display: block; font-size: 0.72rem; text-transform: uppercase; letter-spacing: 1px; color: rgba(255,255,255,0.6); margin-bottom: 6px;">New Password</label>
                    <input type="password" id="pw-new-pass" placeholder="Minimum 6 characters..." style="width: 100%; background: rgba(0,0,0,0.3); border: 1px solid rgba(255,255,255,0.1); padding: 12px 16px; border-radius: 10px; color: #fff; font-size: 0.92rem; outline: none; box-sizing: border-box; transition: all 0.2s;">
                </div>

                <!-- Confirm Password Field -->
                <div>
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
                        <label style="font-size: 0.72rem; text-transform: uppercase; letter-spacing: 1px; color: rgba(255,255,255,0.6);">Confirm Password</label>
                        <span id="pw-match-badge" style="font-size: 0.72rem; color: rgba(255,255,255,0.3);"></span>
                    </div>
                    <input type="password" id="pw-confirm-pass" placeholder="Re-enter new password..." style="width: 100%; background: rgba(0,0,0,0.3); border: 1px solid rgba(255,255,255,0.1); padding: 12px 16px; border-radius: 10px; color: #fff; font-size: 0.92rem; outline: none; box-sizing: border-box; transition: all 0.2s;">
                </div>
            </div>

            <div class="rift-modal-footer" style="padding: 16px 28px 24px; border-top: 1px solid rgba(255, 255, 255, 0.07); display: flex; justify-content: space-between; align-items: center;">
                <button class="rift-btn-secondary" id="pw-resend-btn" style="background: rgba(255,255,255,0.05); color: rgba(255,255,255,0.7); border: 1px solid rgba(255,255,255,0.1); padding: 10px 18px; border-radius: 8px; font-size: 0.82rem; font-weight: 600; cursor: pointer; transition: all 0.2s;">Resend Code</button>
                <button class="rift-btn-primary" id="pw-submit-btn" style="background: #38bdf8; color: #000; border: none; padding: 11px 24px; border-radius: 8px; font-size: 0.88rem; font-weight: 700; cursor: pointer; transition: all 0.2s; box-shadow: 0 0 20px rgba(56, 189, 248, 0.25);">Update & Protect</button>
            </div>
        </div>
    `;

    modalContainer.appendChild(overlay);

    const banner = overlay.querySelector('#pw-modal-banner');
    const showBanner = (msg, isSuccess = false) => {
        banner.style.display = 'block';
        banner.style.background = isSuccess ? 'rgba(34, 197, 94, 0.12)' : 'rgba(239, 68, 68, 0.12)';
        banner.style.border = isSuccess ? '1px solid rgba(34, 197, 94, 0.3)' : '1px solid rgba(239, 68, 68, 0.3)';
        banner.style.color = isSuccess ? '#4ade80' : '#f87171';
        banner.textContent = msg;
    };

    // Auto-dispatch OTP code on modal open
    try {
        await window.go.rift.App.SupabaseResetPasswordForEmail(userEmail);
        showBanner(`A 6-digit security code was dispatched to ${userEmail}.`, true);
    } catch (_) {
        showBanner(`Check your email for the security code, or click Resend.`);
    }

    // OTP Box Navigation & Paste Handling
    const otpBoxes = overlay.querySelectorAll('.otp-box');
    otpBoxes.forEach((box, i) => {
        box.addEventListener('focus', () => {
            box.style.borderColor = '#38bdf8';
            box.style.boxShadow = '0 0 12px rgba(56, 189, 248, 0.3)';
            box.select();
        });
        box.addEventListener('blur', () => {
            box.style.borderColor = 'rgba(255,255,255,0.12)';
            box.style.boxShadow = 'none';
        });
        box.addEventListener('input', (e) => {
            if (e.target.value.length === 1 && i < 5) {
                otpBoxes[i + 1].focus();
            }
        });
        box.addEventListener('keydown', (e) => {
            if (e.key === 'Backspace' && !e.target.value && i > 0) {
                otpBoxes[i - 1].focus();
            }
        });
        box.addEventListener('paste', (e) => {
            e.preventDefault();
            const pasteData = (e.clipboardData || window.clipboardData).getData('text').trim();
            if (/^\d{6}$/.test(pasteData)) {
                pasteData.split('').forEach((digit, idx) => {
                    if (otpBoxes[idx]) otpBoxes[idx].value = digit;
                });
                otpBoxes[5].focus();
            }
        });
    });

    // Auto focus first box
    setTimeout(() => otpBoxes[0]?.focus(), 100);

    // Password Match Checking
    const p1 = overlay.querySelector('#pw-new-pass');
    const p2 = overlay.querySelector('#pw-confirm-pass');
    const matchBadge = overlay.querySelector('#pw-match-badge');

    const checkMatch = () => {
        if (!p1.value || !p2.value) {
            matchBadge.textContent = '';
            return;
        }
        if (p1.value === p2.value) {
            matchBadge.textContent = '✓ Passwords Match';
            matchBadge.style.color = '#4ade80';
        } else {
            matchBadge.textContent = 'Passwords do not match';
            matchBadge.style.color = '#f87171';
        }
    };
    p1.addEventListener('input', checkMatch);
    p2.addEventListener('input', checkMatch);

    const close = () => overlay.remove();
    overlay.querySelector('#pw-modal-close')?.addEventListener('click', close);

    // Resend handler
    overlay.querySelector('#pw-resend-btn')?.addEventListener('click', async () => {
        const btn = overlay.querySelector('#pw-resend-btn');
        btn.textContent = 'Sending...';
        btn.disabled = true;
        try {
            await window.go.rift.App.SupabaseResetPasswordForEmail(userEmail);
            showBanner(`A fresh 6-digit code has been dispatched to ${userEmail}!`, true);
        } catch (err) {
            showBanner(`Failed to resend code: ${err.message || err}`);
        }
        setTimeout(() => {
            btn.textContent = 'Resend Code';
            btn.disabled = false;
        }, 5000);
    });

    // Submit handler
    overlay.querySelector('#pw-submit-btn')?.addEventListener('click', async () => {
        let code = '';
        otpBoxes.forEach(b => code += b.value.trim());

        if (code.length !== 6) {
            showBanner('Please fill in all 6 digits of the verification code.');
            return;
        }
        if (!p1.value || p1.value.length < 6) {
            showBanner('New password must be at least 6 characters long.');
            return;
        }
        if (p1.value !== p2.value) {
            showBanner('Passwords do not match. Please verify.');
            return;
        }

        const submitBtn = overlay.querySelector('#pw-submit-btn');
        submitBtn.textContent = 'Verifying...';
        submitBtn.disabled = true;

        try {
            const resStr = await window.go.rift.App.SupabaseVerifyResetCodeAndSetPassword(userEmail, code, p1.value);
            const res = JSON.parse(resStr);
            if (res.error) throw new Error(res.error);

            showBanner('✓ Master password successfully updated! Closing...', true);
            setTimeout(() => {
                close();
            }, 1200);
        } catch (err) {
            showBanner('Verification failed: ' + (err.message || 'Invalid or expired code.'));
            submitBtn.textContent = 'Update & Protect';
            submitBtn.disabled = false;
        }
    });
}
