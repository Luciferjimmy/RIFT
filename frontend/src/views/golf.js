export function renderGolf() {
    const container = document.createElement('div');
    container.className = 'golf-view';
    container.style.cssText = 'width: 100%; height: 100%; display: flex; flex-direction: column; background: #000; position: relative;';
    
    container.innerHTML = `
        <div style="position: absolute; top: 20px; left: 20px; z-index: 10; display: flex; gap: 20px; align-items: center;">
            <button id="btn-exit-golf" style="background: transparent; color: #fff; border: 1px solid rgba(255,255,255,0.3); padding: 8px 16px; border-radius: 4px; cursor: pointer; font-family: inherit; text-transform: uppercase; font-size: 0.8rem; letter-spacing: 1px; transition: all 0.2s;" onmouseover="this.style.background='rgba(255,255,255,0.1)'" onmouseout="this.style.background='transparent'">Exit Arcade</button>
            <div style="color: #fff; font-size: 1.2rem; font-weight: 300; letter-spacing: 2px;">TERMINAL <span style="font-weight: 800; color: #01ffff;">GOLF</span></div>
            <div style="color: rgba(255,255,255,0.5); font-size: 0.9rem;">LEVEL <span id="golf-level" style="color: #fff; font-weight: bold;">1</span></div>
        </div>
        <canvas id="golfCanvas" style="flex-grow: 1; width: 100%; height: 100%; cursor: crosshair;"></canvas>
    `;

    setTimeout(() => {
        const btnExit = container.querySelector('#btn-exit-golf');
        if (btnExit) btnExit.addEventListener('click', () => window.location.hash = '/profile');
        
        initGolf(container);
    }, 0);

    return container;
}

function initGolf(container) {
    const canvas = container.querySelector('#golfCanvas');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const levelDisplay = container.querySelector('#golf-level');

    function resize() {
        canvas.width = canvas.clientWidth;
        canvas.height = canvas.clientHeight;
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
        
        const safeTop = 80; 
        const safeH = canvas.height - safeTop - 40; 
        const safeW = canvas.width - 40;
        
        scale = Math.min(1, safeW / gridW, safeH / gridH);
        
        ox = canvas.width / 2;
        oy = safeTop + safeH / 2 - ((size * TILE_SIZE) * Math.sin(Math.PI / 6) * scale);
    }

    function generateLevel() {
        strokes = 0;
        const size = Math.min(6 + level * 2, 16);
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
        ctx.strokeStyle = 'rgba(255,255,255,0.2)'; ctx.stroke();
        ctx.fillStyle = type === 2 ? '#222' : '#333';
        ctx.beginPath(); ctx.moveTo(p3.x,p3.y); ctx.lineTo(p2.x,p2.y); ctx.lineTo(b2.x,b2.y); ctx.lineTo(b3.x,b3.y); ctx.fill();
        ctx.strokeStyle = 'rgba(255,255,255,0.2)'; ctx.stroke();
        ctx.fillStyle = type === 2 ? '#333' : type === 3 ? '#111' : '#111';
        ctx.beginPath(); ctx.moveTo(p1.x,p1.y); ctx.lineTo(p2.x,p2.y); ctx.lineTo(p3.x,p3.y); ctx.lineTo(p4.x,p4.y); ctx.fill();
        ctx.strokeStyle = 'rgba(255,255,255,0.3)'; ctx.stroke();
        if (type === 3) {
            const c = project(x+TILE_SIZE/2, y+TILE_SIZE/2, z+h);
            ctx.beginPath(); ctx.ellipse(c.x, c.y, TILE_SIZE*0.3*scale, TILE_SIZE*0.15*scale, 0, 0, Math.PI*2);
            ctx.fillStyle = '#000'; ctx.fill();
            ctx.strokeStyle = 'rgba(255,255,255,0.6)'; ctx.lineWidth = 1.5; ctx.stroke();
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
                ctx.fillStyle = '#fff'; ctx.shadowColor = '#fff'; ctx.shadowBlur = 10; ctx.fill(); ctx.shadowBlur = 0;
            }
        }

        if (isDragging) {
            const pb = project(ball.x, ball.y, ball.z);
            ctx.beginPath(); ctx.moveTo(pb.x, pb.y);
            const ws = unproject(dragStart.x, dragStart.y), wc = unproject(dragCurrent.x, dragCurrent.y);
            let dx = ws.x - wc.x, dy = ws.y - wc.y, pvx = dx*0.05, pvy = dy*0.05, pvz = Math.min(Math.sqrt(pvx*pvx+pvy*pvy)*0.3, 10);
            let sx = ball.x, sy = ball.y, sz = ball.z;
            for (let s = 0; s < 20; s++) { pvz += GRAVITY; sx += pvx; sy += pvy; sz += pvz; if (sz < 0) break; const pp = project(sx, sy, sz); ctx.lineTo(pp.x, pp.y); }
            ctx.strokeStyle = 'rgba(255,255,255,0.4)'; ctx.lineWidth = 1; ctx.setLineDash([4,4]); ctx.stroke(); ctx.setLineDash([]);
        }
    }

    let animationId;
    function loop() { update(); draw(); animationId = requestAnimationFrame(loop); }

    generateLevel(); resize(); loop();

    canvas.addEventListener('mousedown', (e) => {
        if (ball.state !== 'stationary' && ball.z <= ball.radius + 1) return;
        const r = canvas.getBoundingClientRect();
        dragStart.x = e.clientX - r.left; dragStart.y = e.clientY - r.top;
        isDragging = true; dragCurrent = { ...dragStart };
    });
    window.addEventListener('mousemove', (e) => {
        if (!isDragging) return;
        const r = canvas.getBoundingClientRect();
        dragCurrent.x = e.clientX - r.left; dragCurrent.y = e.clientY - r.top;
    });
    window.addEventListener('mouseup', () => {
        if (!isDragging) return;
        isDragging = false;
        const ws = unproject(dragStart.x, dragStart.y), wc = unproject(dragCurrent.x, dragCurrent.y);
        let dx = ws.x - wc.x, dy = ws.y - wc.y;
        ball.vx = dx * 0.05; ball.vy = dy * 0.05; ball.vz = Math.min(Math.sqrt(dx*dx + dy*dy)*0.015, 10);
        ball.state = 'air'; strokes++;
    });
    
    // Cleanup function exposed on the container so router can stop the animation frame
    container._cleanup = () => {
        if (animationId) cancelAnimationFrame(animationId);
        window.removeEventListener('resize', resize);
    };
}
