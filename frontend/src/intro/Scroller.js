import * as THREE from 'three';

// Easing function for the bounce drop
function easeOutBounce(x) {
    const n1 = 7.5625;
    const d1 = 2.75;
    if (x < 1 / d1) {
        return n1 * x * x;
    } else if (x < 2 / d1) {
        return n1 * (x -= 1.5 / d1) * x + 0.75;
    } else if (x < 2.5 / d1) {
        return n1 * (x -= 2.25 / d1) * x + 0.9375;
    } else {
        return n1 * (x -= 2.625 / d1) * x + 0.984375;
    }
}

function lerp(start, end, amt) {
    return (1 - amt) * start + amt * end;
}

export function initScroller(camera, dioramaData) {
    let scrollTarget = 0;
    let scrollCurrent = 0;
    const maxScroll = 12000;
    
    const numJumps = 6; // 0->1, 1->2, 2->3, 3->4, 4->5, 5->6
    
    // UI Elements
    const uiTitle = document.getElementById('intro-title');
    const uiSubtitle = document.getElementById('intro-subtitle');
    const uiStartBtn = document.getElementById('intro-start-btn');
    const panels = [];
    for(let i=1; i<=5; i++) panels.push(document.getElementById(`panel-${i}`));

    function onWheel(e) {
        scrollTarget += e.deltaY * 1.5;
        if (scrollTarget < 0) scrollTarget = 0;
        if (scrollTarget > maxScroll) scrollTarget = maxScroll;
    }
    
    window.addEventListener('wheel', onWheel, { passive: false });
    
    return {
        update: () => {
            scrollCurrent += (scrollTarget - scrollCurrent) * 0.05;
            const p = scrollCurrent / maxScroll; // 0.0 to 1.0
            
            const currentSegment = Math.min(Math.floor(p * numJumps), numJumps - 1);
            let localP = (p * numJumps) - currentSegment; // 0.0 to 1.0 within the jump
            
            // Allow trailing past the last platform for the final button reveal
            if (p >= 1.0) {
                localP = 1.0;
            }

            const pStart = dioramaData.platformsData[currentSegment];
            const pEnd = dioramaData.platformsData[currentSegment + 1];

            // 1. Ball Jump Physics (Parabola)
            const ballX = lerp(pStart.x, pEnd.x, localP);
            const ballZ = lerp(pStart.z, pEnd.z, localP);
            
            // Linear Y interpolation + Sine wave for the arc
            const baseBallY = lerp(pStart.y, pEnd.y, localP) + 12; // 12 units above platform center
            const jumpHeight = Math.sin(localP * Math.PI) * 35; // 35 unit jump height
            
            dioramaData.ball.position.set(ballX, baseBallY + jumpHeight, ballZ);
            
            // Spin the ball as it moves
            dioramaData.ball.rotation.x += 0.05;
            dioramaData.ball.rotation.z -= 0.03;

            // 2. Camera Tracking (Isometric/Trailing perspective)
            // Camera stays slightly above and behind the ball
            const camTargetX = ballX + 40;
            const camTargetY = baseBallY + 50;
            const camTargetZ = ballZ + 80;
            
            camera.position.set(camTargetX, camTargetY, camTargetZ);
            camera.lookAt(ballX, baseBallY + 5, ballZ);

            // 3. Platform Drop Animation
            // Platforms should drop into place *just* before the ball needs to jump to them
            dioramaData.platforms.forEach((mesh, index) => {
                if (index === 0) {
                    mesh.position.y = mesh.userData.targetY; // Start platform always there
                    return;
                }
                
                // Trigger drop when we are halfway through the previous jump
                const dropTriggerP = (index - 1.5) / numJumps;
                if (p > dropTriggerP) {
                    const dropProgress = Math.min((p - dropTriggerP) / (1.0 / numJumps), 1.0);
                    const eased = easeOutBounce(dropProgress);
                    const startY = mesh.userData.targetY + 400; // Drop from 400 units high
                    mesh.position.y = lerp(startY, mesh.userData.targetY, eased);
                } else {
                    mesh.position.y = mesh.userData.targetY + 400;
                }
            });

            // 4. UI Panel Fading
            if (uiTitle && uiSubtitle) {
                if (p < 0.05) {
                    uiTitle.style.opacity = 1 - (p / 0.05);
                    uiSubtitle.style.opacity = 1 - (p / 0.05);
                } else {
                    uiTitle.style.opacity = 0;
                    uiSubtitle.style.opacity = 0;
                }
            }

            // Show panel N when resting on Platform N
            panels.forEach((panel, i) => {
                if (!panel) return;
                const panelIndex = i + 1; // Panel 1 is for Platform 1
                
                // Panel is fully visible when p is exactly at (panelIndex / 6)
                // We give it a small window of visibility (e.g. +/- 0.08)
                const targetP = panelIndex / numJumps;
                const dist = Math.abs(p - targetP);
                
                if (dist < 0.08) {
                    const opacity = 1.0 - (dist / 0.08);
                    panel.style.opacity = opacity;
                    // Slide in effect depending on left/right position
                    const isLeft = (panelIndex % 2 !== 0);
                    const slideOffset = (1.0 - opacity) * (isLeft ? -50 : 50);
                    panel.style.transform = `translateY(-50%) translateX(${slideOffset}px)`;
                } else {
                    panel.style.opacity = 0;
                }
            });

            // Show start button at the very end
            const uiStartWrap = document.getElementById('intro-start-wrap');
            const uiStartSub = document.getElementById('intro-start-sub');
            if (uiStartBtn) {
                if (p > 0.95) {
                    const btnP = (p - 0.95) / 0.05;
                    uiStartBtn.style.opacity = btnP;
                    uiStartBtn.style.pointerEvents = 'auto';
                    uiStartBtn.style.transform = `translateY(${16 - (btnP * 16)}px)`;
                    if (uiStartSub) uiStartSub.style.opacity = btnP;
                    if (uiStartWrap) uiStartWrap.style.pointerEvents = 'auto';
                } else {
                    uiStartBtn.style.opacity = 0;
                    uiStartBtn.style.pointerEvents = 'none';
                    uiStartBtn.style.transform = `translateY(16px)`;
                    if (uiStartSub) uiStartSub.style.opacity = 0;
                    if (uiStartWrap) uiStartWrap.style.pointerEvents = 'none';
                }
            }
        },
        dispose: () => {
            window.removeEventListener('wheel', onWheel);
        }
    };
}
