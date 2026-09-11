// Cyber Snake — Modern Arcade Game for RIFT
export function renderSnake() {
    const container = document.createElement('div');
    container.className = 'snake-view fade-in';
    container.style.cssText = 'width: 100%; height: 100%; display: flex; flex-direction: column; background: #07090e; position: relative; overflow: hidden; font-family: "Plus Jakarta Sans", -apple-system, sans-serif;';

    container.innerHTML = `
        <!-- Top Navigation & Score Bar -->
        <div style="position: absolute; top: 0; left: 0; right: 0; padding: 18px 28px; z-index: 20; display: flex; justify-content: space-between; align-items: center; background: linear-gradient(180deg, rgba(7,9,14,0.95), transparent); backdrop-filter: blur(8px);">
            <div style="display: flex; align-items: center; gap: 10px;">
                <button id="btn-home-snake" style="background: rgba(255,255,255,0.06); color: #fff; border: 1px solid rgba(255,255,255,0.12); padding: 8px 14px; border-radius: 8px; cursor: pointer; font-size: 0.8rem; font-weight: 600; transition: all 0.2s; display: flex; align-items: center; gap: 6px;" title="Go to Library">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>
                    Home
                </button>
                <button id="btn-exit-snake" style="background: rgba(255,255,255,0.06); color: #fff; border: 1px solid rgba(255,255,255,0.12); padding: 8px 14px; border-radius: 8px; cursor: pointer; font-size: 0.8rem; font-weight: 600; transition: all 0.2s; display: flex; align-items: center; gap: 6px;" title="Back to Profile">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
                    Profile
                </button>
                <div style="font-size: 1.25rem; font-weight: 800; color: #fff; letter-spacing: -0.02em; margin-left: 8px;">
                    CYBER <span style="color: #00ffcc;">SNAKE</span>
                </div>
            </div>

            <!-- Active Power-Up Notification Indicator -->
            <div id="active-power-hud" style="display: none; padding: 6px 16px; border-radius: 20px; font-size: 0.78rem; font-weight: 700; letter-spacing: 0.5px; animation: pulse 1s infinite alternate;">
                <!-- Filled dynamically -->
            </div>

            <!-- Score Counters -->
            <div style="display: flex; align-items: center; gap: 28px;">
                <div style="display: flex; flex-direction: column; align-items: flex-end;">
                    <span style="font-size: 0.7rem; color: rgba(255,255,255,0.45); text-transform: uppercase; font-weight: 600; letter-spacing: 0.05em;">Score</span>
                    <span id="snake-score" style="font-size: 1.4rem; font-weight: 800; color: #fff; line-height: 1;">0</span>
                </div>
                <div style="display: flex; flex-direction: column; align-items: flex-end;">
                    <span style="font-size: 0.7rem; color: rgba(255,255,255,0.45); text-transform: uppercase; font-weight: 600; letter-spacing: 0.05em;">Multiplier</span>
                    <span id="snake-combo" style="font-size: 1.4rem; font-weight: 800; color: #38bdf8; line-height: 1;">1x</span>
                </div>
                <div style="display: flex; flex-direction: column; align-items: flex-end;">
                    <span style="font-size: 0.7rem; color: rgba(255,255,255,0.45); text-transform: uppercase; font-weight: 600; letter-spacing: 0.05em;">Best</span>
                    <span id="snake-high" style="font-size: 1.4rem; font-weight: 800; color: #34d399; line-height: 1;">0</span>
                </div>
                <button id="btn-sound-toggle" style="background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.12); color: #fff; width: 36px; height: 36px; border-radius: 8px; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s;" title="Toggle Sound">
                    <svg id="sound-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"></polygon><path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"></path></svg>
                </button>
            </div>
        </div>

        <canvas id="snakeCanvas" style="flex-grow: 1; width: 100%; height: 100%; display: block;"></canvas>

        <!-- In-Game Bottom Legend (Shows what each orb does clearly) -->
        <div style="position: absolute; bottom: 16px; left: 50%; transform: translateX(-50%); z-index: 15; display: flex; align-items: center; gap: 20px; background: rgba(15,18,26,0.85); backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.08); padding: 8px 20px; border-radius: 30px; font-size: 0.75rem; color: rgba(255,255,255,0.7);">
            <div style="display: flex; align-items: center; gap: 6px;">
                <span style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: #00ffcc; box-shadow: 0 0 8px #00ffcc;"></span>
                <span><strong>Food</strong> (+50)</span>
            </div>
            <span style="color: rgba(255,255,255,0.15);">|</span>
            <div style="display: flex; align-items: center; gap: 6px;">
                <span style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: #ffd700; box-shadow: 0 0 8px #ffd700;"></span>
                <span style="display: flex; align-items: center; gap: 4px;"><svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg><strong>Overdrive</strong> (2x Pts)</span>
            </div>
            <span style="color: rgba(255,255,255,0.15);">|</span>
            <div style="display: flex; align-items: center; gap: 6px;">
                <span style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: #38bdf8; box-shadow: 0 0 8px #38bdf8;"></span>
                <span style="display: flex; align-items: center; gap: 4px;"><svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg><strong>Slow-Mo</strong> (Dilation)</span>
            </div>
            <span style="color: rgba(255,255,255,0.15);">|</span>
            <div style="display: flex; align-items: center; gap: 6px;">
                <span style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: #f43f5e; box-shadow: 0 0 8px #f43f5e;"></span>
                <span style="display: flex; align-items: center; gap: 4px;"><svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor"><polygon points="6 2 18 2 22 8 12 22 2 8 6 2"></polygon></svg><strong>Gem</strong> (+500 Bonus)</span>
            </div>
        </div>

        <!-- Clean, Non-AI-Slop Overlay (Start, Pause, Game Over) -->
        <div id="snake-overlay" style="position: absolute; inset: 0; background: rgba(7,9,14,0.85); backdrop-filter: blur(14px); display: flex; flex-direction: column; align-items: center; justify-content: center; z-index: 30;">
            <div style="text-align: center; max-width: 480px; width: 90%; padding: 36px; border-radius: 16px; background: #0f121a; border: 1px solid rgba(255,255,255,0.1); box-shadow: 0 20px 50px rgba(0,0,0,0.7);">
                <h1 id="overlay-title" style="margin: 0 0 8px 0; font-size: 2rem; font-weight: 800; color: #fff; letter-spacing: -0.02em;">CYBER SNAKE</h1>
                <p id="overlay-desc" style="margin: 0 0 24px 0; font-size: 0.9rem; color: rgba(255,255,255,0.6); line-height: 1.5;">
                    Eat nodes, chain combos, and pick up power orbs to reach the high score.
                </p>

                <!-- Clear Orb Legend inside Start Screen -->
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-bottom: 24px; text-align: left;">
                    <div style="background: rgba(255,255,255,0.03); border: 1px solid rgba(255,255,255,0.06); padding: 10px 14px; border-radius: 10px; display: flex; align-items: center; gap: 10px;">
                        <span style="display: inline-block; width: 12px; height: 12px; border-radius: 50%; background: #00ffcc; box-shadow: 0 0 8px #00ffcc; flex-shrink: 0;"></span>
                        <div>
                            <div style="font-size: 0.8rem; font-weight: 700; color: #fff;">Green Core</div>
                            <div style="font-size: 0.72rem; color: rgba(255,255,255,0.5);">+50 Pts · Grows Snake</div>
                        </div>
                    </div>
                    <div style="background: rgba(255,255,255,0.03); border: 1px solid rgba(255,255,255,0.06); padding: 10px 14px; border-radius: 10px; display: flex; align-items: center; gap: 10px;">
                        <span style="display: inline-block; width: 12px; height: 12px; border-radius: 50%; background: #ffd700; box-shadow: 0 0 8px #ffd700; flex-shrink: 0;"></span>
                        <div>
                            <div style="font-size: 0.8rem; font-weight: 700; color: #ffd700; display: flex; align-items: center; gap: 4px;">
                                <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
                                Overdrive
                            </div>
                            <div style="font-size: 0.72rem; color: rgba(255,255,255,0.5);">Speed Boost & 2x Pts (5s)</div>
                        </div>
                    </div>
                    <div style="background: rgba(255,255,255,0.03); border: 1px solid rgba(255,255,255,0.06); padding: 10px 14px; border-radius: 10px; display: flex; align-items: center; gap: 10px;">
                        <span style="display: inline-block; width: 12px; height: 12px; border-radius: 50%; background: #38bdf8; box-shadow: 0 0 8px #38bdf8; flex-shrink: 0;"></span>
                        <div>
                            <div style="font-size: 0.8rem; font-weight: 700; color: #38bdf8; display: flex; align-items: center; gap: 4px;">
                                <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                                Slow-Mo
                            </div>
                            <div style="font-size: 0.72rem; color: rgba(255,255,255,0.5);">Bullet-Time Turns (5s)</div>
                        </div>
                    </div>
                    <div style="background: rgba(255,255,255,0.03); border: 1px solid rgba(255,255,255,0.06); padding: 10px 14px; border-radius: 10px; display: flex; align-items: center; gap: 10px;">
                        <span style="display: inline-block; width: 12px; height: 12px; border-radius: 50%; background: #f43f5e; box-shadow: 0 0 8px #f43f5e; flex-shrink: 0;"></span>
                        <div>
                            <div style="font-size: 0.8rem; font-weight: 700; color: #f43f5e; display: flex; align-items: center; gap: 4px;">
                                <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor"><polygon points="6 2 18 2 22 8 12 22 2 8 6 2"></polygon></svg>
                                Bonus Gem
                            </div>
                            <div style="font-size: 0.72rem; color: rgba(255,255,255,0.5);">+500 Instant Points</div>
                        </div>
                    </div>
                </div>

                <div style="display: flex; gap: 10px; justify-content: center; margin-bottom: 24px; font-size: 0.75rem; color: rgba(255,255,255,0.5);">
                    <span>Controls: <strong>WASD / Arrow Keys</strong> to steer · <strong>SPACE</strong> to pause · <strong>ESC</strong> to exit</span>
                </div>

                <div style="display: flex; gap: 10px; justify-content: center; align-items: center; flex-wrap: wrap;">
                    <button id="btn-overlay-home" style="background: rgba(255,255,255,0.06); color: #fff; border: 1px solid rgba(255,255,255,0.15); padding: 12px 18px; border-radius: 8px; font-size: 0.85rem; font-weight: 600; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: all 0.2s;">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>
                        Home
                    </button>
                    <button id="btn-overlay-back" style="background: rgba(255,255,255,0.06); color: #fff; border: 1px solid rgba(255,255,255,0.15); padding: 12px 18px; border-radius: 8px; font-size: 0.85rem; font-weight: 600; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: all 0.2s;">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
                        Profile
                    </button>
                    <button id="btn-start-snake" style="background: #38bdf8; color: #080808; border: none; padding: 12px 30px; border-radius: 8px; font-size: 0.95rem; font-weight: 700; letter-spacing: 0.02em; cursor: pointer; transition: transform 0.15s, background 0.15s;">
                        START GAME
                    </button>
                </div>
            </div>
        </div>
    `;

    setTimeout(() => {
        const btnExit = container.querySelector('#btn-exit-snake');
        const btnHome = container.querySelector('#btn-home-snake');
        const btnOverlayBack = container.querySelector('#btn-overlay-back');
        const btnOverlayHome = container.querySelector('#btn-overlay-home');

        const exitToProfile = () => {
            if (window.snakeCleanup) window.snakeCleanup();
            window.location.hash = '/profile';
        };

        const exitToHome = () => {
            if (window.snakeCleanup) window.snakeCleanup();
            window.location.hash = '/library';
        };

        if (btnExit) btnExit.addEventListener('click', exitToProfile);
        if (btnOverlayBack) btnOverlayBack.addEventListener('click', exitToProfile);
        if (btnHome) btnHome.addEventListener('click', exitToHome);
        if (btnOverlayHome) btnOverlayHome.addEventListener('click', exitToHome);

        initCyberSnake(container);
    }, 0);

    return container;
}

function initCyberSnake(container) {
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

    // Audio Engine
    let audioCtx = null;
    let soundEnabled = true;

    function initAudio() {
        if (!audioCtx) {
            try {
                audioCtx = new (window.AudioContext || window.webkitAudioContext)();
            } catch (e) {}
        }
    }

    function playTone(freq, type, duration, gainVal = 0.08) {
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

    btnSound.addEventListener('click', () => {
        soundEnabled = !soundEnabled;
        btnSound.style.opacity = soundEnabled ? '1' : '0.4';
    });

    let highScore = parseInt(localStorage.getItem('rift_snake_highscore') || '0', 10);
    highEl.textContent = highScore.toLocaleString();

    let width = 0, height = 0;
    const GRID_SIZE = 24;
    let cols = 0, rows = 0;

    function resize() {
        const dpr = window.devicePixelRatio || 1;
        width = canvas.clientWidth;
        height = canvas.clientHeight;
        canvas.width = width * dpr;
        canvas.height = height * dpr;
        ctx.scale(dpr, dpr);
        cols = Math.floor(width / GRID_SIZE);
        rows = Math.floor(height / GRID_SIZE);
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
    let powerEffect = null; // { type: 'overdrive'|'chrono', duration: frames }

    let particles = [];
    let floatingTexts = [];

    function spawnParticles(x, y, color, count = 16) {
        for (let i = 0; i < count; i++) {
            const angle = Math.random() * Math.PI * 2;
            const speed = Math.random() * 3.5 + 1.2;
            particles.push({
                x: x * GRID_SIZE + GRID_SIZE / 2,
                y: y * GRID_SIZE + GRID_SIZE / 2,
                vx: Math.cos(angle) * speed,
                vy: Math.sin(angle) * speed,
                color: color,
                size: Math.random() * 2.5 + 1.5,
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
        scoreEl.textContent = '0';
        comboEl.textContent = '1x';
        if (powerHud) powerHud.style.display = 'none';
        placeFood();
        isRunning = true;
        isPaused = false;
        overlay.style.display = 'none';
        playTone(520, 'sine', 0.1, 0.12);
    }

    function gameOver() {
        isRunning = false;
        playTone(140, 'sawtooth', 0.35, 0.2);
        if (score > highScore) {
            highScore = score;
            localStorage.setItem('rift_snake_highscore', highScore.toString());
            highEl.textContent = highScore.toLocaleString();
        }

        if (powerHud) powerHud.style.display = 'none';
        overlayTitle.textContent = 'GAME OVER';
        overlayDesc.innerHTML = `Score: <strong style="color: #fff; font-size: 1.3rem;">${score.toLocaleString()}</strong> · Highest Multiplier: <strong style="color: #38bdf8;">${combo}x</strong>`;
        btnStart.textContent = 'PLAY AGAIN';
        overlay.style.display = 'flex';
    }

    function handleKeyDown(e) {
        if (e.key === 'Escape') {
            if (window.snakeCleanup) window.snakeCleanup();
            window.location.hash = '/profile';
            return;
        }

        if (!isRunning && (e.key === ' ' || e.key === 'Enter')) {
            resetGame();
            return;
        }

        if (e.key === ' ' || e.key.toLowerCase() === 'p') {
            isPaused = !isPaused;
            if (isPaused) {
                overlayTitle.textContent = 'PAUSED';
                overlayDesc.textContent = 'Press SPACE to continue.';
                btnStart.textContent = 'RESUME';
                overlay.style.display = 'flex';
            } else {
                overlay.style.display = 'none';
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

    btnStart.addEventListener('click', () => {
        if (isPaused) {
            isPaused = false;
            overlay.style.display = 'none';
        } else {
            resetGame();
        }
    });

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
                    comboEl.textContent = '1x';
                    comboEl.style.color = '#38bdf8';
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
                        powerHud.textContent = `⚡ OVERDRIVE (2X SPEED) · ${secondsLeft}s`;
                    } else if (powerEffect.type === 'chrono') {
                        powerHud.style.background = 'rgba(56, 189, 248, 0.15)';
                        powerHud.style.border = '1px solid rgba(56, 189, 248, 0.4)';
                        powerHud.style.color = '#38bdf8';
                        powerHud.textContent = `⏳ SLOW MOTION · ${secondsLeft}s`;
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
                    comboEl.textContent = `${combo}x`;
                    comboEl.style.color = combo >= 3 ? '#34d399' : '#38bdf8';

                    let points = 50 * combo;
                    let particleColor = '#00ffcc';

                    if (food.type === 'overdrive') {
                        powerEffect = { type: 'overdrive', duration: 300 };
                        points = 150 * combo;
                        particleColor = '#ffd700';
                        addFloatingText(food.x, food.y, `⚡ SPEED BOOST! +${points}`, '#ffd700');
                        playTone(660, 'triangle', 0.15, 0.12);
                    } else if (food.type === 'chrono') {
                        powerEffect = { type: 'chrono', duration: 300 };
                        points = 100 * combo;
                        particleColor = '#38bdf8';
                        addFloatingText(food.x, food.y, `⏳ SLOW-MO! +${points}`, '#38bdf8');
                        playTone(440, 'sine', 0.25, 0.12);
                    } else if (food.type === 'quantum') {
                        points = 500 * combo;
                        particleColor = '#f43f5e';
                        addFloatingText(food.x, food.y, `💎 +${points} BONUS!`, '#f43f5e');
                        playTone(880, 'sine', 0.2, 0.15);
                    } else {
                        addFloatingText(food.x, food.y, `+${points}`, '#00ffcc');
                        playTone(480 + combo * 50, 'sine', 0.08, 0.1);
                    }

                    score += points;
                    scoreEl.textContent = score.toLocaleString();
                    spawnParticles(food.x, food.y, particleColor, 18);
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

        // Subtle Grid
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.025)';
        ctx.lineWidth = 1;
        for (let x = 0; x < width; x += GRID_SIZE) {
            ctx.beginPath();
            ctx.moveTo(x, 0);
            ctx.lineTo(x, height);
            ctx.stroke();
        }
        for (let y = 0; y < height; y += GRID_SIZE) {
            ctx.beginPath();
            ctx.moveTo(0, y);
            ctx.lineTo(width, y);
            ctx.stroke();
        }

        // Render Food with Explicit Distinct Icons
        const fx = food.x * GRID_SIZE + GRID_SIZE / 2;
        const fy = food.y * GRID_SIZE + GRID_SIZE / 2;
        const pulse = Math.sin(Date.now() * 0.008) * 1.5;

        ctx.save();
        if (food.type === 'overdrive') {
            // Yellow Overdrive with ⚡
            ctx.shadowColor = '#ffd700';
            ctx.shadowBlur = 14;
            ctx.fillStyle = '#ffd700';
            ctx.beginPath();
            ctx.arc(fx, fy, (GRID_SIZE / 2 - 2) + pulse, 0, Math.PI * 2);
            ctx.fill();

            ctx.fillStyle = '#07090e';
            ctx.font = 'bold 11px sans-serif';
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText('⚡', fx, fy);
        } else if (food.type === 'chrono') {
            // Cyan Chrono with ⏳
            ctx.shadowColor = '#38bdf8';
            ctx.shadowBlur = 14;
            ctx.fillStyle = '#38bdf8';
            ctx.beginPath();
            ctx.arc(fx, fy, (GRID_SIZE / 2 - 2) + pulse, 0, Math.PI * 2);
            ctx.fill();

            ctx.fillStyle = '#07090e';
            ctx.font = 'bold 10px sans-serif';
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText('⏳', fx, fy);
        } else if (food.type === 'quantum') {
            // Pink Gem with 💎
            ctx.shadowColor = '#f43f5e';
            ctx.shadowBlur = 18;
            ctx.fillStyle = '#f43f5e';
            ctx.beginPath();
            ctx.arc(fx, fy, (GRID_SIZE / 2 - 1) + pulse, 0, Math.PI * 2);
            ctx.fill();

            ctx.fillStyle = '#ffffff';
            ctx.font = 'bold 10px sans-serif';
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText('💎', fx, fy);
        } else {
            // Standard Green Core with pulsing target
            ctx.shadowColor = '#00ffcc';
            ctx.shadowBlur = 12;
            ctx.fillStyle = '#00ffcc';
            ctx.beginPath();
            ctx.arc(fx, fy, (GRID_SIZE / 2 - 3) + pulse, 0, Math.PI * 2);
            ctx.fill();

            ctx.fillStyle = '#07090e';
            ctx.beginPath();
            ctx.arc(fx, fy, 2, 0, Math.PI * 2);
            ctx.fill();
        }
        ctx.restore();

        // Render Snake Body
        for (let i = snake.length - 1; i >= 0; i--) {
            const seg = snake[i];
            const px = seg.x * GRID_SIZE;
            const py = seg.y * GRID_SIZE;

            ctx.save();
            if (i === 0) {
                ctx.shadowColor = powerEffect ? (powerEffect.type === 'overdrive' ? '#ffd700' : '#38bdf8') : '#00ffcc';
                ctx.shadowBlur = 14;
                ctx.fillStyle = powerEffect ? (powerEffect.type === 'overdrive' ? '#ffd700' : '#38bdf8') : '#00ffcc';
                ctx.beginPath();
                ctx.roundRect(px + 1, py + 1, GRID_SIZE - 2, GRID_SIZE - 2, 5);
                ctx.fill();

                // Eyes
                ctx.fillStyle = '#07090e';
                const eyeOffset = 5;
                if (dir.x === 1) {
                    ctx.fillRect(px + GRID_SIZE - 6, py + eyeOffset, 3, 3);
                    ctx.fillRect(px + GRID_SIZE - 6, py + GRID_SIZE - eyeOffset - 3, 3, 3);
                } else if (dir.x === -1) {
                    ctx.fillRect(px + 3, py + eyeOffset, 3, 3);
                    ctx.fillRect(px + 3, py + GRID_SIZE - eyeOffset - 3, 3, 3);
                } else if (dir.y === 1) {
                    ctx.fillRect(px + eyeOffset, py + GRID_SIZE - 6, 3, 3);
                    ctx.fillRect(px + GRID_SIZE - eyeOffset - 3, py + GRID_SIZE - 6, 3, 3);
                } else {
                    ctx.fillRect(px + eyeOffset, py + 3, 3, 3);
                    ctx.fillRect(px + GRID_SIZE - eyeOffset - 3, py + 3, 3, 3);
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
                ctx.beginPath();
                ctx.roundRect(px + 2, py + 2, GRID_SIZE - 4, GRID_SIZE - 4, 3);
                ctx.fill();
            }
            ctx.restore();
        }

        // Render Particles
        for (let i = particles.length - 1; i >= 0; i--) {
            const p = particles[i];
            p.x += p.vx;
            p.y += p.vy;
            p.alpha -= p.decay;
            if (p.alpha <= 0) {
                particles.splice(i, 1);
                continue;
            }
            ctx.save();
            ctx.globalAlpha = p.alpha;
            ctx.fillStyle = p.color;
            ctx.beginPath();
            ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2);
            ctx.fill();
            ctx.restore();
        }

        // Render Floating Score Texts
        for (let i = floatingTexts.length - 1; i >= 0; i--) {
            const ft = floatingTexts[i];
            ft.y += ft.vy;
            ft.alpha -= 0.025;
            if (ft.alpha <= 0) {
                floatingTexts.splice(i, 1);
                continue;
            }
            ctx.save();
            ctx.globalAlpha = ft.alpha;
            ctx.fillStyle = ft.color;
            ctx.font = 'bold 12px "Plus Jakarta Sans", sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText(ft.text, ft.x, ft.y);
            ctx.restore();
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
