import { Engine } from './Engine.js';

export function renderIntro() {
    const container = document.createElement('div');
    container.style.width = '100vw';
    container.style.height = '100vh';
    container.style.background = '#000000'; // Pure black background
    container.style.position = 'relative';
    container.style.overflow = 'hidden';
    container.style.setProperty('--wails-draggable', 'drag');
    
    // UI Overlay Container
    const overlay = document.createElement('div');
    overlay.style.position = 'absolute';
    overlay.style.top = '0';
    overlay.style.left = '0';
    overlay.style.width = '100%';
    overlay.style.height = '100%';
    overlay.style.pointerEvents = 'none'; 
    
    const glassStyle = `
        position: absolute;
        top: 50%;
        transform: translateY(-50%);
        width: 450px;
        max-width: calc(100vw - 80px);
        padding: 34px 38px;
        background: rgba(10, 10, 14, 0.78);
        backdrop-filter: blur(24px);
        -webkit-backdrop-filter: blur(24px);
        border: 1px solid rgba(255, 255, 255, 0.12);
        border-radius: 22px;
        color: white;
        font-family: var(--font-sans, 'Plus Jakarta Sans', sans-serif);
        opacity: 0;
        box-shadow: 0 24px 60px rgba(0, 0, 0, 0.75), inset 0 1px 0 rgba(255, 255, 255, 0.12);
        transition: opacity 0.4s cubic-bezier(0.16, 1, 0.3, 1), transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
    `;

    overlay.innerHTML = `
        <!-- Skip Button -->
        <button id="intro-skip-btn" style="position: absolute; top: 28px; right: 32px; z-index: 100; pointer-events: auto; background: rgba(255, 255, 255, 0.06); border: 1px solid rgba(255, 255, 255, 0.12); color: rgba(255, 255, 255, 0.65); padding: 8px 18px; border-radius: 8px; font-family: var(--font-mono, monospace); font-size: 0.72rem; letter-spacing: 1.5px; text-transform: uppercase; cursor: pointer; backdrop-filter: blur(12px); transition: all 0.2s;">
            SKIP INTRO ✕
        </button>

        <!-- Title & Subtitle Screen -->
        <h1 id="intro-title" style="color: white; font-family: var(--font-display, 'Smooch Sans'), 'Smooch Sans', sans-serif; font-weight: 800; font-size: clamp(6.5rem, 16vw, 10.5rem); letter-spacing: 0.10em; text-transform: uppercase; margin: 0; position: absolute; top: 34%; left: 50%; transform: translate(-50%, -50%); transition: opacity 0.3s; text-shadow: 0 10px 40px rgba(0,0,0,0.9), 0 0 50px rgba(255,255,255,0.25); text-align: center; line-height: 1;">
            RI<span style="font-family: var(--font-serif, 'Instrument Serif'), 'Instrument Serif', Georgia, serif; font-style: italic; font-weight: 400; text-transform: lowercase; margin-left: -0.06em; margin-right: 0.02em; color: #ffffff;">f</span>T
        </h1>
        <p id="intro-subtitle" style="color: rgba(255,255,255,0.55); font-family: var(--font-mono, 'JetBrains Mono'), monospace; font-size: 0.85rem; letter-spacing: 3px; text-transform: uppercase; position: absolute; top: 62%; left: 50%; transform: translate(-50%, -50%); transition: opacity 0.3s; text-align: center; white-space: nowrap;">
            Scroll down · Let’s talk about that $2,500 laptop
        </p>
        
        <!-- Panel 1 (Left): The Expensive Lie -->
        <div id="panel-1" style="${glassStyle} left: 8%;">
            <span style="font-family: var(--font-mono, monospace); font-size: 0.68rem; letter-spacing: 2.5px; text-transform: uppercase; color: rgba(255, 255, 255, 0.4); display: block; margin-bottom: 12px; font-weight: 600;">01 // THE EXPENSIVE LIE</span>
            <h3 style="font-size: 1.65rem; margin: 0 0 14px 0; font-weight: 700; color: #ffffff; letter-spacing: -0.01em; line-height: 1.25;">Nice fancy typewriter you got there.</h3>
            <p style="font-size: 0.95rem; line-height: 1.65; margin: 0; color: rgba(255,255,255,0.78);">You told your parents or tax accountant you bought Apple Silicon for "video editing" and "pro productivity." Don't lie to us. You bought 18 GPU cores secretly hoping you could finally game without your lap catching fire. Then you opened Mac Steam and found three compatible indie titles from 2012. Welcome to group therapy.</p>
        </div>

        <!-- Panel 2 (Right): The Reddit Ritual -->
        <div id="panel-2" style="${glassStyle} right: 8%;">
            <span style="font-family: var(--font-mono, monospace); font-size: 0.68rem; letter-spacing: 2.5px; text-transform: uppercase; color: rgba(255, 255, 255, 0.4); display: block; margin-bottom: 12px; font-weight: 600;">02 // THE REDDIT RITUAL</span>
            <h3 style="font-size: 1.65rem; margin: 0 0 14px 0; font-weight: 700; color: #ffffff; letter-spacing: -0.01em; line-height: 1.25;">The 2 AM Terminal breakdown.</h3>
            <p style="font-size: 0.95rem; line-height: 1.65; margin: 0; color: rgba(255,255,255,0.78);">Usually, running a Windows game on Mac means 14 open Reddit tabs, copying random "sudo chmod" commands from a post written in 2021, and praying Steve Jobs blesses your Wine prefix. If you enjoy debugging DLLs on a Friday night, respect. If you don't, that's literally why we built this.</p>
        </div>

        <!-- Panel 3 (Left): The Real Translation -->
        <div id="panel-3" style="${glassStyle} left: 8%;">
            <span style="font-family: var(--font-mono, monospace); font-size: 0.68rem; letter-spacing: 2.5px; text-transform: uppercase; color: rgba(255, 255, 255, 0.4); display: block; margin-bottom: 12px; font-weight: 600;">03 // THE TRANSLATION</span>
            <h3 style="font-size: 1.65rem; margin: 0 0 14px 0; font-weight: 700; color: #ffffff; letter-spacing: -0.01em; line-height: 1.25;">DirectX meets Tim Cook's ego.</h3>
            <p style="font-size: 0.95rem; line-height: 1.65; margin: 0; color: rgba(255,255,255,0.78);">Windows games speak DirectX. Your Mac exclusively speaks Apple Metal. RIFT sits in the middle like an exhausted marriage counselor, intercepting every graphics draw call in real time. Is it dark magic? Pretty much. Will anti-cheat still complain? Occasionally. But it works.</p>
        </div>

        <!-- Panel 4 (Right): Zero Plumbing -->
        <div id="panel-4" style="${glassStyle} right: 8%;">
            <span style="font-family: var(--font-mono, monospace); font-size: 0.68rem; letter-spacing: 2.5px; text-transform: uppercase; color: rgba(255, 255, 255, 0.4); display: block; margin-bottom: 12px; font-weight: 600;">04 // ZERO PLUMBING</span>
            <h3 style="font-size: 1.65rem; margin: 0 0 14px 0; font-weight: 700; color: #ffffff; letter-spacing: -0.01em; line-height: 1.25;">We already did the crying for you.</h3>
            <p style="font-size: 0.95rem; line-height: 1.65; margin: 0; color: rgba(255,255,255,0.78);">Does game X need DXVK? Does game Y need D3DMetal? Does game Z explode unless you inject three random C++ runtime DLLs? Yes, yes, and yes. You shouldn't have to troubleshoot like an IT intern. RIFT tunes every knob under the hood before you even hit install.</p>
        </div>

        <!-- Panel 5 (Left): The Brutal Truth -->
        <div id="panel-5" style="${glassStyle} left: 8%;">
            <span style="font-family: var(--font-mono, monospace); font-size: 0.68rem; letter-spacing: 2.5px; text-transform: uppercase; color: rgba(255, 255, 255, 0.4); display: block; margin-bottom: 12px; font-weight: 600;">05 // THE HONEST TRUTH</span>
            <h3 style="font-size: 1.65rem; margin: 0 0 14px 0; font-weight: 700; color: #ffffff; letter-spacing: -0.01em; line-height: 1.25;">No corporate hype. Just honesty.</h3>
            <p style="font-size: 0.95rem; line-height: 1.65; margin: 0; color: rgba(255,255,255,0.78);">Look, kernel-level rootkits like Valorant won't work — Vanguard treats Wine like a crime scene. But for thousands of real games across Steam and Epic? You click play. It runs. It's the closest your Mac has ever been to an actual gaming PC without voiding your AppleCare.</p>
        </div>

        <!-- Final Button -->
        <div id="intro-start-wrap" style="position: absolute; bottom: 12%; left: 50%; transform: translateX(-50%); display: flex; flex-direction: column; align-items: center; gap: 10px; pointer-events: none;">
            <button id="intro-start-btn" style="opacity: 0; pointer-events: none; background: #ffffff; color: #000000; border: none; padding: 18px 46px; border-radius: 9999px; font-family: var(--font-sans, sans-serif); font-weight: 700; font-size: 0.95rem; cursor: pointer; text-transform: uppercase; letter-spacing: 2px; box-shadow: 0 10px 35px rgba(255,255,255,0.25); transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);">
                LET'S PROVE TIM COOK WRONG →
            </button>
            <span id="intro-start-sub" style="opacity: 0; font-family: var(--font-mono, monospace); font-size: 0.72rem; letter-spacing: 1.5px; color: rgba(255, 255, 255, 0.45); text-transform: uppercase; transition: opacity 0.3s;">
                Zero terminal commands required
            </span>
        </div>
    `;
    
    container.appendChild(overlay);

    const onComplete = () => {
        window.localStorage.setItem('hasSeenIntro', 'true');
        if (engine) engine.dispose();
        const hasLoggedIn = window.localStorage.getItem('hasLoggedIn');
        window.location.hash = hasLoggedIn ? '/library' : '/login';
    };

    let engine;
    
    setTimeout(() => {
        engine = new Engine(container);
        
        const btn = container.querySelector('#intro-start-btn');
        if (btn) {
            btn.addEventListener('mouseenter', () => { btn.style.transform = 'scale(1.05)'; });
            btn.addEventListener('mouseleave', () => { btn.style.transform = 'scale(1)'; });
            btn.addEventListener('click', onComplete);
        }

        const skipBtn = container.querySelector('#intro-skip-btn');
        if (skipBtn) {
            skipBtn.addEventListener('mouseenter', () => {
                skipBtn.style.background = 'rgba(255, 255, 255, 0.14)';
                skipBtn.style.color = '#ffffff';
            });
            skipBtn.addEventListener('mouseleave', () => {
                skipBtn.style.background = 'rgba(255, 255, 255, 0.06)';
                skipBtn.style.color = 'rgba(255, 255, 255, 0.65)';
            });
            skipBtn.addEventListener('click', onComplete);
        }
    }, 0);

    container._cleanup = () => {
        if (engine) engine.dispose();
    };

    return container;
}
