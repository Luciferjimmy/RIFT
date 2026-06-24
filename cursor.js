/* ==========================================================================
   RIFT Waitlist Custom Cursor Trail (Spring-Damper Inertia)
   ========================================================================== */

document.addEventListener('DOMContentLoaded', () => {
  const dot = document.getElementById('cursor-follower');
  const glow = document.getElementById('cursor-follower-glow');

  if (!dot || !glow) return;

  // Track target mouse position
  let mouseX = window.innerWidth / 2;
  let mouseY = window.innerHeight / 2;

  // Current position of follower dot (fast speed)
  let dotX = mouseX;
  let dotY = mouseY;

  // Current position of background glow (slower speed / higher inertia)
  let glowX = mouseX;
  let glowY = mouseY;

  // Check if touch device - if so, hide custom cursor followers
  const isTouchDevice = window.matchMedia('(pointer: coarse)').matches;
  if (isTouchDevice) {
    dot.style.display = 'none';
    glow.style.display = 'none';
    return;
  }

  // Enable custom cursor styles by adding active class to document root
  document.documentElement.classList.add('custom-cursor-active');

  // Mouse move listener
  window.addEventListener('mousemove', (e) => {
    mouseX = e.clientX;
    mouseY = e.clientY;
  });

  // Inertia Animation loop (lerp: Linear Interpolation)
  function animateCursor() {
    // Lerp formulas: current = current + (target - current) * speed
    // Dot interpolation (fast lag)
    dotX += (mouseX - dotX) * 0.25;
    dotY += (mouseY - dotY) * 0.25;

    // Glow interpolation (slow lag for a fluid drift feel)
    glowX += (mouseX - glowX) * 0.08;
    glowY += (mouseY - glowY) * 0.08;

    dot.style.left = `${dotX}px`;
    dot.style.top = `${dotY}px`;

    glow.style.left = `${glowX}px`;
    glow.style.top = `${glowY}px`;

    requestAnimationFrame(animateCursor);
  }
  
  // Start animation loop
  requestAnimationFrame(animateCursor);

  // Hover states handler
  const hoverSelectors = 'a, button, input, textarea, .game-card, .bento-card, .mockup-btn, .social-icon';
  
  function addHoverEffect() {
    document.body.classList.add('cursor-hovering');
  }

  function removeHoverEffect() {
    document.body.classList.remove('cursor-hovering');
  }

  // Bind initial hoverable elements
  function bindHovers() {
    const hoverables = document.querySelectorAll(hoverSelectors);
    hoverables.forEach(el => {
      el.removeEventListener('mouseenter', addHoverEffect);
      el.removeEventListener('mouseleave', removeHoverEffect);
      el.addEventListener('mouseenter', addHoverEffect);
      el.addEventListener('mouseleave', removeHoverEffect);
    });
  }

  bindHovers();

  // Watch for dynamic DOM changes to bind hovers to new elements
  const observer = new MutationObserver(bindHovers);
  observer.observe(document.body, { childList: true, subtree: true });

  // Hide cursor when leaving window
  document.addEventListener('mouseleave', () => {
    dot.style.opacity = '0';
    glow.style.opacity = '0';
  });

  document.addEventListener('mouseenter', () => {
    dot.style.opacity = '1';
    glow.style.opacity = '1';
  });
});
