export function renderVerify() {
    const registeredEmail = sessionStorage.getItem('just_registered_email');
    const emailDisplay = registeredEmail 
        ? `<strong style="color: #38bdf8;">${registeredEmail}</strong>`
        : 'your email address';

    const container = document.createElement('div');
    container.style.width = '100vw';
    container.style.height = '100vh';
    container.style.background = '#000000';
    container.style.display = 'flex';
    container.style.alignItems = 'center';
    container.style.justifyContent = 'center';
    container.style.position = 'relative';
    container.style.overflow = 'hidden';
    container.style.setProperty('--wails-draggable', 'drag');

    const glassStyle = `
        background: rgba(255, 255, 255, 0.03);
        backdrop-filter: blur(20px);
        -webkit-backdrop-filter: blur(20px);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 16px;
        box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    `;

    container.innerHTML = `
        <div style="position: absolute; top: -50%; left: -50%; width: 200%; height: 200%; background: radial-gradient(circle at 50% 50%, rgba(56, 189, 248, 0.02) 0%, transparent 50%); pointer-events: none;"></div>
        
        <div style="${glassStyle} width: 450px; padding: 50px 40px; text-align: center; position: relative; z-index: 10; pointer-events: auto;">
            
            <div style="margin-bottom: 30px; display: flex; justify-content: center;">
                <div style="width: 80px; height: 80px; border-radius: 50%; background: rgba(56, 189, 248, 0.1); border: 1px solid rgba(56, 189, 248, 0.3); display: flex; align-items: center; justify-content: center;">
                    <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
                        <polyline points="22,6 12,13 2,6"></polyline>
                    </svg>
                </div>
            </div>

            <h1 style="color: white; font-family: var(--font-display); font-size: 1.8rem; margin: 0 0 16px 0; font-weight: 700;">Check Your Inbox</h1>
            
            <p style="color: rgba(255,255,255,0.6); font-family: var(--font-sans); margin: 0 0 40px 0; font-size: 0.95rem; line-height: 1.5;">
                We've sent a verification link to ${emailDisplay}.<br/>
                Please click it to activate your account, then return to sign in.
            </p>
            
            <button id="back-btn" style="width: 100%; background: rgba(255,255,255,0.05); color: white; border: 1px solid rgba(255,255,255,0.1); padding: 14px; border-radius: 8px; font-family: var(--font-sans); font-weight: 600; font-size: 1rem; cursor: pointer; transition: all 0.2s;">
                Return to Login
            </button>
        </div>
    `;

    setTimeout(() => {
        const btn = container.querySelector('#back-btn');
        btn.addEventListener('mouseover', () => {
            btn.style.background = 'rgba(255,255,255,0.1)';
            btn.style.borderColor = 'rgba(255,255,255,0.2)';
        });
        btn.addEventListener('mouseout', () => {
            btn.style.background = 'rgba(255,255,255,0.05)';
            btn.style.borderColor = 'rgba(255,255,255,0.1)';
        });
        btn.addEventListener('click', () => {
            window.location.hash = '/login';
        });
    }, 0);

    return container;
}
