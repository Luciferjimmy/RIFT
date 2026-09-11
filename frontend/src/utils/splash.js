// RIFT Cinematic Opening Splash Animation
// Pure black-and-white monochrome cinematic sequence when the application launches.
// Matches exact sidebar typography (Smooch Sans + Instrument Serif) with slow majestic pacing.

export function showStartupSplash() {
    // Avoid running if already present
    if (document.getElementById('rift-startup-splash')) return;

    const splash = document.createElement('div');
    splash.id = 'rift-startup-splash';
    splash.style.cssText = `
        position: fixed;
        inset: 0;
        z-index: 999999;
        background: #000000;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        overflow: hidden;
        user-select: none;
        cursor: pointer;
        opacity: 1;
        transition: opacity 0.6s cubic-bezier(0.16, 1, 0.3, 1), transform 0.6s cubic-bezier(0.16, 1, 0.3, 1), filter 0.6s ease;
    `;

    splash.innerHTML = `
        <!-- Monochrome Hyperspace Warp Canvas -->
        <canvas id="splash-canvas" style="position: absolute; inset: 0; width: 100%; height: 100%; pointer-events: none;"></canvas>

        <!-- Ambient Optical Halo (Pure White Diffusion) -->
        <div style="position: absolute; width: 700px; height: 350px; background: radial-gradient(ellipse at center, rgba(255, 255, 255, 0.09) 0%, rgba(255, 255, 255, 0.02) 45%, transparent 75%); filter: blur(60px); pointer-events: none; animation: glowPulse 2.8s ease-in-out infinite alternate;"></div>

        <!-- Monolithic Center Wordmark (Exact Sidebar Typography) -->
        <div style="position: relative; z-index: 10; display: flex; flex-direction: column; align-items: center; text-align: center;">
            <div id="splash-logo" style="opacity: 0; filter: blur(20px); transform: scale(0.91); letter-spacing: 0.22em; transition: opacity 1.2s cubic-bezier(0.16, 1, 0.3, 1), filter 1.2s cubic-bezier(0.16, 1, 0.3, 1), transform 1.2s cubic-bezier(0.16, 1, 0.3, 1), letter-spacing 1.2s cubic-bezier(0.16, 1, 0.3, 1); font-family: var(--font-display, 'Smooch Sans'), 'Smooch Sans', sans-serif; font-weight: 800; font-size: clamp(6.5rem, 16vw, 11rem); line-height: 1; text-transform: uppercase; color: #ffffff; text-shadow: 0 0 45px rgba(255, 255, 255, 0.45), 0 0 95px rgba(255, 255, 255, 0.18), 0 10px 40px rgba(0, 0, 0, 0.9);">
                RI<span style="font-family: var(--font-serif, 'Instrument Serif'), 'Instrument Serif', Georgia, serif; font-style: italic; font-weight: 400; text-transform: lowercase; margin-left: -0.06em; margin-right: 0.02em; color: #ffffff; text-shadow: 0 0 35px rgba(255, 255, 255, 0.65);">f</span>T
            </div>
        </div>
    `;

    document.body.appendChild(splash);

    // ============================================
    // 3D MONOCHROME WARP ENGINE (3.0s Calibrated Flow)
    // ============================================
    const canvas = splash.querySelector('#splash-canvas');
    const ctx = canvas.getContext('2d');
    let animId = null;

    function resize() {
        if (!canvas) return;
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;
    }
    resize();
    window.addEventListener('resize', resize);

    const numStars = 85;
    const stars = [];
    for (let i = 0; i < numStars; i++) {
        stars.push({
            x: (Math.random() - 0.5) * window.innerWidth,
            y: (Math.random() - 0.5) * window.innerHeight,
            z: Math.random() * window.innerWidth,
            pz: window.innerWidth,
            size: Math.random() * 1.5 + 0.5
        });
    }

    // Calibrated for 3-second sequence
    let speed = 1.3;

    function renderParticles() {
        if (!ctx) return;
        ctx.fillStyle = 'rgba(0, 0, 0, 0.28)';
        ctx.fillRect(0, 0, canvas.width, canvas.height);

        const cx = canvas.width / 2;
        const cy = canvas.height / 2;

        speed += 0.045; // Smooth acceleration

        for (let s of stars) {
            s.pz = s.z;
            s.z -= speed;

            if (s.z <= 0) {
                s.z = canvas.width;
                s.pz = s.z;
                s.x = (Math.random() - 0.5) * canvas.width;
                s.y = (Math.random() - 0.5) * canvas.height;
            }

            const k = 180 / s.z;
            const px = s.x * k + cx;
            const py = s.y * k + cy;

            const pk = 180 / s.pz;
            const ppx = s.x * pk + cx;
            const ppy = s.y * pk + cy;

            if (px >= 0 && px <= canvas.width && py >= 0 && py <= canvas.height) {
                const alpha = Math.min(1, (1 - s.z / canvas.width) * 1.5);
                ctx.strokeStyle = `rgba(255, 255, 255, ${alpha})`;
                ctx.lineWidth = Math.max(0.6, (1 - s.z / canvas.width) * 2.3);
                ctx.beginPath();
                ctx.moveTo(ppx, ppy);
                ctx.lineTo(px, py);
                ctx.stroke();
            }
        }

        animId = requestAnimationFrame(renderParticles);
    }
    renderParticles();

    // ============================================
    // CINEMATIC REVEAL TIMELINE
    // ============================================
    const logo = splash.querySelector('#splash-logo');

    // T+200ms: Wordmark emerges from optical blur into focus
    setTimeout(() => {
        if (logo) {
            logo.style.opacity = '1';
            logo.style.filter = 'blur(0px)';
            logo.style.transform = 'scale(1)';
            logo.style.letterSpacing = '0.10em';
        }
    }, 200);

    // ============================================
    // DISMISS & CLEANUP (Smooth 0.6s dissolve at 2.4s -> 3.0s total)
    // ============================================
    let dismissed = false;

    function dismissSplash() {
        if (dismissed) return;
        dismissed = true;

        window.removeEventListener('resize', resize);
        document.removeEventListener('keydown', handleKey);
        splash.removeEventListener('click', dismissSplash);

        splash.style.opacity = '0';
        splash.style.transform = 'scale(1.05)';
        splash.style.filter = 'blur(8px)';

        setTimeout(() => {
            if (animId) cancelAnimationFrame(animId);
            splash.remove();
        }, 600);
    }

    function handleKey() {
        dismissSplash();
    }

    splash.addEventListener('click', dismissSplash);
    document.addEventListener('keydown', handleKey);

    // Auto-dismiss starts at 2.4s so the full sequence completes at 3.0s
    setTimeout(dismissSplash, 2400);
}
