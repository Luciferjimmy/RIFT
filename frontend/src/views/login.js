export function renderLogin() {
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

    const inputStyle = `
        width: 100%;
        background: rgba(255, 255, 255, 0.05);
        border: 1px solid rgba(255, 255, 255, 0.1);
        padding: 12px 16px;
        color: white;
        font-family: var(--font-sans);
        font-size: 1rem;
        border-radius: 8px;
        margin-bottom: 16px;
        outline: none;
        box-sizing: border-box;
        transition: border-color 0.2s, background 0.2s;
    `;

    container.innerHTML = `
        <div style="position: absolute; top: -50%; left: -50%; width: 200%; height: 200%; background: radial-gradient(circle at 50% 50%, rgba(255,255,255,0.02) 0%, transparent 50%); pointer-events: none;"></div>
        
        <div style="${glassStyle} width: 420px; padding: 40px; text-align: center; position: relative; z-index: 10; pointer-events: auto;">
            <div style="font-family: var(--font-display); font-weight: 800; font-size: 3rem; letter-spacing: 10px; text-transform: uppercase; margin: 0 0 10px 0; color: white;">
                RI<span style="font-family: var(--font-serif); font-style: italic; font-weight: 400; text-transform: lowercase; margin-left: -5px; margin-right: 5px;">f</span>T
            </div>
            <p id="form-subtitle" style="color: rgba(255,255,255,0.5); font-family: var(--font-sans); margin: 0 0 32px 0; font-size: 0.92rem; letter-spacing: 0.5px;">Sign in to access your archive</p>
            
            <div id="error-banner" style="display: none; background: rgba(248, 113, 113, 0.1); border: 1px solid rgba(248, 113, 113, 0.2); color: #f87171; padding: 12px; border-radius: 8px; margin-bottom: 20px; font-family: var(--font-sans); font-size: 0.85rem; text-align: left; line-height: 1.4;"></div>
            <div id="success-banner" style="display: none; background: rgba(34, 197, 94, 0.1); border: 1px solid rgba(34, 197, 94, 0.25); color: #4ade80; padding: 12px; border-radius: 8px; margin-bottom: 20px; font-family: var(--font-sans); font-size: 0.85rem; text-align: left; line-height: 1.4;"></div>
            
            <div style="text-align: left;">
                <!-- Signup Name Field -->
                <div id="field-name-container" style="display: none;">
                    <label style="display: block; color: rgba(255,255,255,0.7); font-family: var(--font-sans); font-size: 0.78rem; letter-spacing: 1px; text-transform: uppercase; margin-bottom: 6px;">Display Name</label>
                    <input type="text" id="login-name" style="${inputStyle}" placeholder="Enter your name..." autocomplete="off" />
                </div>

                <!-- Email Field (Login, Signup, Forgot Step 1) -->
                <div id="field-email-container">
                    <label style="display: block; color: rgba(255,255,255,0.7); font-family: var(--font-sans); font-size: 0.78rem; letter-spacing: 1px; text-transform: uppercase; margin-bottom: 6px;">Email Address</label>
                    <input type="text" id="login-email" style="${inputStyle}" placeholder="Enter your email..." autocomplete="off" />
                </div>
                
                <!-- Password Field (Login, Signup) -->
                <div id="field-pass-container">
                    <label id="pass-label" style="display: block; color: rgba(255,255,255,0.7); font-family: var(--font-sans); font-size: 0.78rem; letter-spacing: 1px; text-transform: uppercase; margin-bottom: 6px;">Password</label>
                    <input type="password" id="login-pass" style="${inputStyle}" placeholder="Enter your password..." />
                </div>

                <!-- 6-digit Code Field (Forgot Step 2) -->
                <div id="field-code-container" style="display: none;">
                    <label style="display: block; color: #38bdf8; font-family: var(--font-sans); font-size: 0.78rem; letter-spacing: 1px; text-transform: uppercase; margin-bottom: 6px;">6-Digit Reset Code</label>
                    <input type="text" id="reset-code" maxlength="6" style="${inputStyle} border-color: rgba(56, 189, 248, 0.4); text-align: center; letter-spacing: 4px; font-weight: 700; font-family: 'JetBrains Mono', monospace;" placeholder="123456" autocomplete="off" />
                    <div style="font-size: 0.75rem; color: rgba(255,255,255,0.45); margin-top: -10px; margin-bottom: 16px;">Check your email inbox for the 6-digit code.</div>
                </div>

                <!-- New Password Field (Forgot Step 2) -->
                <div id="field-newpass-container" style="display: none;">
                    <label style="display: block; color: rgba(255,255,255,0.7); font-family: var(--font-sans); font-size: 0.78rem; letter-spacing: 1px; text-transform: uppercase; margin-bottom: 6px;">New Password</label>
                    <input type="password" id="new-pass" style="${inputStyle}" placeholder="Enter new password (min 6 chars)..." />
                </div>
            </div>

            <button id="main-btn" style="width: 100%; background: white; color: black; border: none; padding: 14px; border-radius: 8px; font-family: var(--font-sans); font-weight: 700; font-size: 1rem; cursor: pointer; transition: transform 0.1s, opacity 0.2s; margin-top: 6px;">
                Authenticate
            </button>
            
            <div style="margin-top: 20px; display: flex; flex-direction: column; gap: 8px; color: rgba(255,255,255,0.4); font-family: var(--font-sans); font-size: 0.82rem;">
                <span class="hover-white" id="toggle-mode-btn" style="cursor: pointer; transition: color 0.2s;">Create an account</span>
                <span class="hover-white" id="forgot-pass-btn" style="cursor: pointer; transition: color 0.2s; font-size: 0.78rem; color: rgba(255,255,255,0.35);">Forgot your password?</span>
            </div>
        </div>
    `;

    setTimeout(() => {
        // Modes: 'login', 'signup', 'forgot_email', 'forgot_code'
        let mode = 'login';
        let resetEmailTarget = '';
        
        const btn = container.querySelector('#main-btn');
        const toggleBtn = container.querySelector('#toggle-mode-btn');
        const forgotBtn = container.querySelector('#forgot-pass-btn');
        
        const fieldName = container.querySelector('#field-name-container');
        const fieldEmail = container.querySelector('#field-email-container');
        const fieldPass = container.querySelector('#field-pass-container');
        const fieldCode = container.querySelector('#field-code-container');
        const fieldNewPass = container.querySelector('#field-newpass-container');
        
        const subtitle = container.querySelector('#form-subtitle');
        const errorBanner = container.querySelector('#error-banner');
        const successBanner = container.querySelector('#success-banner');
        const inputs = container.querySelectorAll('input');
        
        function showError(msg) {
            errorBanner.innerHTML = msg;
            errorBanner.style.display = 'block';
            successBanner.style.display = 'none';
        }
        
        function showSuccess(msg) {
            successBanner.innerHTML = msg;
            successBanner.style.display = 'block';
            errorBanner.style.display = 'none';
        }
        
        function clearBanners() {
            errorBanner.style.display = 'none';
            successBanner.style.display = 'none';
        }

        function setMode(newMode) {
            mode = newMode;
            clearBanners();

            if (mode === 'login') {
                subtitle.textContent = "Sign in to access your archive";
                fieldName.style.display = 'none';
                fieldEmail.style.display = 'block';
                fieldPass.style.display = 'block';
                fieldCode.style.display = 'none';
                fieldNewPass.style.display = 'none';
                btn.textContent = "Authenticate";
                toggleBtn.textContent = "Create an account";
                toggleBtn.style.display = 'inline';
                forgotBtn.textContent = "Forgot your password?";
                forgotBtn.style.display = 'inline';
            } else if (mode === 'signup') {
                subtitle.textContent = "Register a new archive account";
                fieldName.style.display = 'block';
                fieldEmail.style.display = 'block';
                fieldPass.style.display = 'block';
                fieldCode.style.display = 'none';
                fieldNewPass.style.display = 'none';
                btn.textContent = "Register";
                toggleBtn.textContent = "Already have an account? Sign in";
                toggleBtn.style.display = 'inline';
                forgotBtn.style.display = 'none';
            } else if (mode === 'forgot_email') {
                subtitle.textContent = "Reset your account password";
                fieldName.style.display = 'none';
                fieldEmail.style.display = 'block';
                fieldPass.style.display = 'none';
                fieldCode.style.display = 'none';
                fieldNewPass.style.display = 'none';
                btn.textContent = "Send Reset Code";
                toggleBtn.textContent = "Back to Login";
                toggleBtn.style.display = 'inline';
                forgotBtn.style.display = 'none';
            } else if (mode === 'forgot_code') {
                subtitle.textContent = `Enter the 6-digit code sent to ${resetEmailTarget}`;
                fieldName.style.display = 'none';
                fieldEmail.style.display = 'none';
                fieldPass.style.display = 'none';
                fieldCode.style.display = 'block';
                fieldNewPass.style.display = 'block';
                btn.textContent = "Update Password & Sign In";
                toggleBtn.textContent = "Back to Login";
                toggleBtn.style.display = 'inline';
                forgotBtn.textContent = "Resend Code";
                forgotBtn.style.display = 'inline';
            }
        }
        
        // Mode toggle logic
        toggleBtn.addEventListener('click', () => {
            if (mode === 'login') setMode('signup');
            else setMode('login');
        });

        forgotBtn.addEventListener('click', () => {
            if (mode === 'forgot_code') {
                setMode('forgot_email');
            } else {
                setMode('forgot_email');
            }
        });

        inputs.forEach(input => {
            input.addEventListener('focus', () => {
                input.style.borderColor = 'rgba(255,255,255,0.5)';
                input.style.background = 'rgba(255,255,255,0.1)';
            });
            input.addEventListener('blur', () => {
                input.style.borderColor = 'rgba(255,255,255,0.1)';
                input.style.background = 'rgba(255,255,255,0.05)';
            });
        });

        const hoverSpans = container.querySelectorAll('.hover-white');
        hoverSpans.forEach(span => {
            span.addEventListener('mouseover', () => span.style.color = 'white');
            span.addEventListener('mouseout', () => span.style.color = 'rgba(255,255,255,0.4)');
        });

        btn.addEventListener('mouseover', () => btn.style.opacity = '0.8');
        btn.addEventListener('mouseout', () => btn.style.opacity = '1');
        btn.addEventListener('mousedown', () => btn.style.transform = 'scale(0.98)');
        btn.addEventListener('mouseup', () => btn.style.transform = 'scale(1)');

        btn.addEventListener('click', async () => {
            clearBanners();

            // Guard against running in browser without Wails Go backend
            if (!window.go || !window.go.rift || !window.go.rift.App) {
                showError("Wails Go desktop runtime not detected. Please run RIFT inside the desktop app window, not a standalone web browser tab.");
                return;
            }

            const email = document.getElementById('login-email').value.trim();
            const pass = document.getElementById('login-pass').value;
            const name = document.getElementById('login-name').value.trim();
            const code = document.getElementById('reset-code').value.trim();
            const newPass = document.getElementById('new-pass').value;
            
            // Validation
            if (mode === 'login') {
                if (!email || !pass) {
                    showError("Please enter your email and password.");
                    return;
                }
            } else if (mode === 'signup') {
                if (!name || !email || !pass) {
                    showError("Please fill in your name, email, and password.");
                    return;
                }
                if (pass.length < 6) {
                    showError("Password must be at least 6 characters.");
                    return;
                }
            } else if (mode === 'forgot_email') {
                if (!email) {
                    showError("Please enter your account email address.");
                    return;
                }
            } else if (mode === 'forgot_code') {
                if (!code || !newPass) {
                    showError("Please enter the 6-digit code and your new password.");
                    return;
                }
                if (newPass.length < 6) {
                    showError("New password must be at least 6 characters.");
                    return;
                }
            }

            btn.textContent = 'Processing...';
            btn.style.opacity = '0.5';
            btn.style.pointerEvents = 'none';
            
            try {
                if (mode === 'signup') {
                    const resStr = await window.go.rift.App.SupabaseSignUp(email, name, pass);
                    const res = JSON.parse(resStr);
                    if (res.error) throw new Error(res.error);
                    
                    setMode('login');
                    document.getElementById('login-email').value = email;
                    document.getElementById('login-pass').value = '';
                    btn.textContent = 'Authenticate';
                    btn.style.opacity = '1';
                    btn.style.pointerEvents = 'auto';
                    showSuccess("Confirm your account: check your email, and sign in here.");
                    return;
                } else if (mode === 'login') {
                    const resStr = await window.go.rift.App.SupabaseSignIn(email, pass);
                    const res = JSON.parse(resStr);
                    if (res.error) throw new Error(res.error);
                    
                    window.localStorage.setItem('hasLoggedIn', 'true');
                    const hasPassedSetup = window.localStorage.getItem('hasPassedSetup') === 'true';
                    if (hasPassedSetup) {
                        window.hasPassedSetup = true;
                        window.location.hash = '/library';
                    } else {
                        window.location.hash = '/setup';
                    }
                } else if (mode === 'forgot_email') {
                    resetEmailTarget = email;
                    const resStr = await window.go.rift.App.SupabaseResetPasswordForEmail(email);
                    const res = JSON.parse(resStr);
                    if (res.error) throw new Error(res.error);
                    
                    setMode('forgot_code');
                    showSuccess(`A 6-digit verification code was sent to ${email}. Check your inbox!`);
                } else if (mode === 'forgot_code') {
                    const resStr = await window.go.rift.App.SupabaseVerifyResetCodeAndSetPassword(resetEmailTarget, code, newPass);
                    const res = JSON.parse(resStr);
                    if (res.error) throw new Error(res.error);
                    
                    window.localStorage.setItem('hasLoggedIn', 'true');
                    showSuccess("Password reset successful! Logging into your archive...");
                    setTimeout(() => {
                        const hasPassedSetup = window.localStorage.getItem('hasPassedSetup') === 'true';
                        if (hasPassedSetup) {
                            window.hasPassedSetup = true;
                            window.location.hash = '/library';
                        } else {
                            window.location.hash = '/setup';
                        }
                    }, 1000);
                }
            } catch (err) {
                btn.textContent = mode === 'signup' ? 'Register' : (mode === 'login' ? 'Authenticate' : (mode === 'forgot_email' ? 'Send Reset Code' : 'Update Password'));
                btn.style.opacity = '1';
                btn.style.pointerEvents = 'auto';
                
                const errStr = err.message.toLowerCase();
                if (errStr.includes('invalid login credentials')) {
                    showError("Incorrect email or password.");
                } else if (errStr.includes('email not confirmed')) {
                    showError("✉️ <strong>Account not confirmed yet!</strong> Please check your inbox and click the confirmation link before signing in.");
                } else if (errStr.includes('already registered')) {
                    showError("An account with this email already exists.");
                } else if (errStr.includes('rate limit')) {
                    showError("Too many attempts. Please wait a minute and try again.");
                } else if (errStr.includes('invalid verification code') || errStr.includes('token has expired') || errStr.includes('otp')) {
                    showError("Invalid or expired 6-digit code. Please check your email or click Resend Code.");
                } else {
                    showError(err.message || "An authentication error occurred.");
                }
            }
        });
    }, 0);

    return container;
}
