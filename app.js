/* ==========================================================================
   RIFT Waitlist Main JS: Lenis Scroll, GSAP Parallax, and Scrollytelling
   ========================================================================== */

document.addEventListener('DOMContentLoaded', () => {
  
  // Register GSAP ScrollTrigger
  gsap.registerPlugin(ScrollTrigger);

  // Supabase Configuration: Connects waitlist directly to database table
  const SUPABASE_CONFIG = {
    url: "https://kkucgbskyyynyoryymcw.supabase.co",
    anonKey: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImtrdWNnYnNreXl5bnlvcnl5bWN3Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODIxMzcyNDgsImV4cCI6MjA5NzcxMzI0OH0.BkW9-GbDpOi4GhVy74wyu8huWdnuN7WazHMKV60-bzI"
  };

  /* ------------------------------------------------------------------------
     1. Lenis Smooth Scroll Initialization
     ------------------------------------------------------------------------ */
  const lenis = new Lenis({
    duration: 1.1,
    easing: (t) => Math.min(1, 1.001 - Math.pow(2, -10 * t)), // easeOutExpo
    smoothWheel: true,
    touchMultiplier: 1.5,
    infinite: false,
  });

  // RAF loop for Lenis
  function raf(time) {
    lenis.raf(time);
    requestAnimationFrame(raf);
  }
  requestAnimationFrame(raf);

  // Link Lenis to GSAP ScrollTrigger
  lenis.on('scroll', ScrollTrigger.update);
  
  gsap.ticker.add((time) => {
    lenis.raf(time * 1000);
  });
  gsap.ticker.lagSmoothing(0);

  // Recalculate ScrollTrigger offsets once images and fonts are fully loaded
  window.addEventListener('load', () => {
    ScrollTrigger.refresh();
  });

  /* ------------------------------------------------------------------------
     2. GSAP Entrance Animations (Hero Section)
     ------------------------------------------------------------------------ */
  const tlHero = gsap.timeline();

  tlHero.to('.hero-meta-lead', {
    y: 0,
    opacity: 1,
    duration: 0.6,
    ease: 'power3.out'
  })
  .to('.hero-title', {
    y: 0,
    opacity: 1,
    duration: 0.8,
    ease: 'power4.out'
  }, '-=0.4')
  .to('.hero-subtitle', {
    y: 0,
    opacity: 1,
    duration: 0.8,
    ease: 'power3.out'
  }, '-=0.6')
  .to('.hero-cta-wrapper', {
    y: 0,
    opacity: 1,
    duration: 0.6,
    ease: 'power3.out'
  }, '-=0.6')
  .to('.mockup-wrapper', {
    y: 0,
    scale: 1,
    opacity: 1,
    duration: 0.8,
    ease: 'power4.out'
  }, '-=0.4');

  /* ------------------------------------------------------------------------
     3. Parallax Scroll Animations for Floating 3D Spheres
     ------------------------------------------------------------------------ */
  // Check if media query supports desktop view (spheres are visible)
  if (window.innerWidth > 1024) {
    gsap.to('#bg-sphere-left', {
      scrollTrigger: {
        trigger: '#hero',
        start: 'top top',
        end: 'bottom top',
        scrub: 1.2
      },
      y: -100,
      rotate: 20,
      scale: 1.05,
      ease: 'none'
    });

    gsap.to('#bg-sphere-right', {
      scrollTrigger: {
        trigger: '#hero',
        start: 'top top',
        end: 'bottom top',
        scrub: 1.2
      },
      y: -150,
      rotate: -25,
      scale: 0.95,
      ease: 'none'
    });
  }

  /* ------------------------------------------------------------------------
     4. Scrollytelling - Prompt & Response Steps
     ------------------------------------------------------------------------ */
  const descBlocks = document.querySelectorAll('.scrolly-desc-block');
  
  descBlocks.forEach((block, index) => {
    ScrollTrigger.create({
      trigger: block,
      start: "top center+=80",
      end: "bottom center-=80",
      onEnter: () => activateStep(index + 1),
      onEnterBack: () => activateStep(index + 1),
    });
  });

  function activateStep(stepNum) {
    // Fade inactive description blocks
    descBlocks.forEach((block, idx) => {
      if (idx === stepNum - 1) {
        block.classList.add('active-step');
        gsap.to(block, { opacity: 1, duration: 0.3 });
      } else {
        block.classList.remove('active-step');
        gsap.to(block, { opacity: 0.25, duration: 0.3 });
      }
    });

    // Update active terminal prompt cards on the left
    const scrollyCards = document.querySelectorAll('.scrolly-card');
    scrollyCards.forEach((card, idx) => {
      if (idx === stepNum - 1) {
        card.classList.add('active');
        // Simple entry pop
        gsap.fromTo(card, { scale: 0.96, y: 10 }, { scale: 1, y: 0, duration: 0.4, ease: 'back.out(1.2)' });
      } else {
        card.classList.remove('active');
      }
    });

    // Trigger custom animations inside each launcher mockup card
    if (stepNum === 1) {
      // Step 1: Account sync logs stagger fade-in
      const syncLines = document.querySelectorAll('#visual-step-1 .sync-line');
      gsap.killTweensOf(syncLines);
      // Reset
      gsap.set(syncLines, { opacity: 0, y: 5 });
      // Animate
      gsap.to(syncLines, {
        opacity: (index, target) => target.classList.contains('text-dim') ? 0.45 : 1,
        y: 0,
        stagger: 0.25,
        duration: 0.4,
        ease: 'power2.out',
        delay: 0.2
      });
    } else if (stepNum === 2) {
      // Step 2: Store grid items stagger pop-in
      const discoverCards = document.querySelectorAll('#visual-step-2 .discover-card');
      gsap.killTweensOf(discoverCards);
      // Reset
      gsap.set(discoverCards, { opacity: 0, scale: 0.95, y: 8 });
      // Animate
      gsap.to(discoverCards, {
        opacity: 1,
        scale: 1,
        y: 0,
        stagger: 0.08,
        duration: 0.4,
        ease: 'back.out(1.1)',
        delay: 0.15
      });
    } else if (stepNum === 3) {
      // Step 3: Library PLAY button press & launch console typing/fade-in
      const playBtn = document.querySelector('#visual-step-3 .library-play-btn');
      const playLines = document.querySelectorAll('#visual-step-3 .play-line');
      
      gsap.killTweensOf([playBtn, playLines]);
      
      // Reset PLAY button to default white style
      gsap.set(playBtn, { 
        backgroundColor: '#FFFFFF', 
        color: '#000000', 
        borderColor: 'transparent',
        boxShadow: '0 0 8px rgba(255, 255, 255, 0.15)',
        textShadow: 'none',
        textContent: '▶ PLAY'
      });
      playBtn.innerText = '▶ PLAY';
      
      // Reset console lines
      gsap.set(playLines, { opacity: 0, x: -5 });
      
      // Create animation timeline
      const tl = gsap.timeline({ delay: 0.2 });
      
      // 1. Simulate button hover/press animation
      tl.to(playBtn, { scale: 0.94, duration: 0.1, ease: 'sine.inOut' })
        .to(playBtn, { 
          scale: 1, 
          backgroundColor: '#00DF6C', 
          color: '#0A0A0F', 
          boxShadow: '0 0 15px rgba(0, 223, 108, 0.6)', 
          duration: 0.15,
          onComplete: () => {
            playBtn.innerText = '● RUNNING';
          }
        });
      
      // 2. Fade in WINE logs sequentially
      tl.to(playLines, {
        opacity: 1,
        x: 0,
        stagger: 0.3,
        duration: 0.3,
        ease: 'power2.out'
      }, "+=0.15");
    }
  }

  /* ------------------------------------------------------------------------
     5. Scroll-Triggered Grid Fade-In Reveals
     ------------------------------------------------------------------------ */
  // Bento features grid reveal
  gsap.fromTo('.bento-card', 
    { y: 40, opacity: 0 },
    {
      scrollTrigger: {
        trigger: '.bento-grid',
        start: 'top bottom-=100',
        toggleActions: 'play none none none'
      },
      y: 0,
      opacity: 1,
      duration: 0.7,
      stagger: 0.12,
      ease: 'power3.out'
    }
  );

  // Comparison Table reveal
  gsap.fromTo('.comparison-table-wrapper',
    { y: 40, opacity: 0 },
    {
      scrollTrigger: {
        trigger: '.comparison-section',
        start: 'top bottom-=100',
        toggleActions: 'play none none none'
      },
      y: 0,
      opacity: 1,
      duration: 0.8,
      ease: 'power3.out'
    }
  );

  // FAQ Grid items reveal
  gsap.fromTo('.faq-item',
    { y: 30, opacity: 0 },
    {
      scrollTrigger: {
        trigger: '.faq-grid',
        start: 'top bottom-=100',
        toggleActions: 'play none none none'
      },
      y: 0,
      opacity: 1,
      duration: 0.7,
      stagger: 0.1,
      ease: 'power3.out'
    }
  );

  // Waitlist card block reveal
  gsap.fromTo('.waitlist-card',
    { y: 50, opacity: 0, scale: 0.98 },
    {
      scrollTrigger: {
        trigger: '.waitlist-section',
        start: 'top bottom-=150',
        toggleActions: 'play none none none'
      },
      y: 0,
      opacity: 1,
      scale: 1,
      duration: 0.8,
      ease: 'power3.out'
    }
  );

  /* ------------------------------------------------------------------------
     6. Waitlist Form Submission & Portal Animate Sequences
     ------------------------------------------------------------------------ */
  const footerForm = document.getElementById('footer-signup-form');
  const footerNameInput = document.getElementById('waitlist-name');
  const footerEmailInput = document.getElementById('waitlist-email');
  const footerFeedback = document.getElementById('waitlist-form-feedback');
  const footerSubmitBtn = document.getElementById('waitlist-submit-btn');
  const successBox = document.getElementById('waitlist-success-box');
  const successEmailDisplay = document.getElementById('success-email-display');

  // Simple Email Validation pattern
  function validateEmail(email) {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(String(email).toLowerCase());
  }

  // Footer / Main Waitlist Form handler
  footerForm.addEventListener('submit', (e) => {
    e.preventDefault();
    const nameValue = footerNameInput ? footerNameInput.value.trim() : '';
    const emailValue = footerEmailInput.value.trim();

    if (!nameValue) {
      showFeedback(footerFeedback, 'Name is required.', 'error');
      return;
    }
    if (!emailValue) {
      showFeedback(footerFeedback, 'Email address is required.', 'error');
      return;
    }
    if (!validateEmail(emailValue)) {
      showFeedback(footerFeedback, 'Please enter a valid email address.', 'error');
      return;
    }

    // Start submission process
    footerSubmitBtn.disabled = true;
    footerSubmitBtn.querySelector('span').innerText = 'Processing...';
    showFeedback(footerFeedback, '', '');

    function showSuccessView(email) {
      gsap.to(footerForm, {
        opacity: 0,
        y: -15,
        duration: 0.3,
        onComplete: () => {
          footerForm.style.display = 'none';
          successBox.style.display = 'flex';
          successEmailDisplay.innerText = email;
          
          // Animate success ring rotates/scale
          gsap.from('.success-ring', { rotate: -180, scale: 0, duration: 0.8, ease: 'back.out(1.5)' });
          gsap.from('.success-checkmark', { scale: 0, delay: 0.3, duration: 0.4, ease: 'back.out(1.8)' });

          // Auto-redirect to the cinematic easter-egg page after 3 seconds
          setTimeout(() => {
            window.location.href = 'easter-egg.html';
          }, 3000);
        }
      });
    }

    // Fallback/Demo mode if Supabase credentials are not configured yet
    if (!SUPABASE_CONFIG.url || !SUPABASE_CONFIG.anonKey) {
      setTimeout(() => {
        showSuccessView(emailValue);
      }, 1000);
      return;
    }

    // Direct HTTP POST to Supabase secure RPC function
    const rpcUrl = `${SUPABASE_CONFIG.url}/rest/v1/rpc/join_waitlist`;
    fetch(rpcUrl, {
      method: 'POST',
      headers: {
        'apikey': SUPABASE_CONFIG.anonKey,
        'Authorization': `Bearer ${SUPABASE_CONFIG.anonKey}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        user_name: nameValue || null,
        user_email: emailValue
      })
    })
    .then(response => {
      if (!response.ok) {
        throw new Error('Server error');
      }
      return response.json();
    })
    .then(data => {
      if (data && data.success) {
        showSuccessView(emailValue);
      } else {
        // Obscured errors: handle simple status messages safely
        if (data && data.error === 'Already registered') {
          showFeedback(footerFeedback, 'This email is already registered on the waitlist.', 'error');
        } else {
          showFeedback(footerFeedback, 'Failed to join waitlist. Please verify your email.', 'error');
        }
        footerSubmitBtn.disabled = false;
        footerSubmitBtn.querySelector('span').innerText = 'Request Invitation';
      }
    })
    .catch(error => {
      showFeedback(footerFeedback, 'Network error. Please try again.', 'error');
      footerSubmitBtn.disabled = false;
      footerSubmitBtn.querySelector('span').innerText = 'Request Invitation';
    });
  });

  /* ==========================================================================
     18. Physics Playground (Antigravity Mode)
     ========================================================================== */
  const toggleBtn = document.getElementById('physics-toggle');
  let physicsActive = false;
  let engine, runner, mouseConstraint;
  let boundaryBodies = [];
  let elementBodies = [];
  let spawnedSpheres = [];
  let originalScrollY = 0;

  // Targets: headings, bento cards, buttons, mockups, descriptions, stickies
  const targetSelectors = [
    '.hero-intro-text',
    '.hero-cta-wrapper',
    '#hero-mockup',
    '.pipeline-title-wrapper',
    '.scrolly-visual-sticky',
    '.scrolly-desc-block',
    '.games-section .section-container',
    '.bento-card',
    '.comparison-table-wrapper',
    '.faq-card',
    '.waitlist-card'
  ].join(', ');

  if (toggleBtn) {
    toggleBtn.addEventListener('click', () => {
      if (!physicsActive) {
        startPhysicsPlayground();
      } else {
        stopPhysicsPlayground();
      }
    });
  }

  function startPhysicsPlayground() {
    if (typeof Matter === 'undefined') {
      console.error('Matter.js not loaded yet');
      return;
    }
    physicsActive = true;
    toggleBtn.classList.add('active');
    toggleBtn.innerText = '💥 Restore Layout';

    // 1. Save scroll position and scroll to top instantly to align coordinates
    originalScrollY = window.scrollY;
    window.scrollTo({ top: 0, behavior: 'instant' });

    // 2. Lock screen scroll
    document.body.style.overflow = 'hidden';
    if (window.lenis) {
      window.lenis.stop();
    }

    // 3. Select target elements
    const elements = document.querySelectorAll(targetSelectors);
    
    // Save original styles for restoration
    elementBodies = [];
    
    // 4. Initialize Matter.js Engine & World
    const { Engine, World, Bodies, Runner, Mouse, MouseConstraint } = Matter;
    engine = Engine.create();
    
    // Set zero gravity explicitly
    if (engine.gravity) {
      engine.gravity.y = 0;
    } else if (engine.world && engine.world.gravity) {
      engine.world.gravity.y = 0;
    }
    
    const world = engine.world;
    const viewWidth = window.innerWidth;
    const viewHeight = window.innerHeight;

    // 5. Create viewport boundary walls so elements are locked inside the screen and can't go outside
    const wallThickness = 100;
    const leftWall = Bodies.rectangle(-wallThickness / 2, viewHeight / 2, wallThickness, viewHeight, { isStatic: true });
    const rightWall = Bodies.rectangle(viewWidth + wallThickness / 2, viewHeight / 2, wallThickness, viewHeight, { isStatic: true });
    const topWall = Bodies.rectangle(viewWidth / 2, -wallThickness / 2, viewWidth, wallThickness, { isStatic: true });
    const bottomWall = Bodies.rectangle(viewWidth / 2, viewHeight + wallThickness / 2, viewWidth, wallThickness, { isStatic: true });
    
    boundaryBodies = [leftWall, rightWall, topWall, bottomWall];
    World.add(world, boundaryBodies);

    // 6. Convert DOM elements into Matter bodies
    elements.forEach((el, index) => {
      // Get current bounding box relative to viewport
      const rect = el.getBoundingClientRect();
      if (rect.width === 0 || rect.height === 0) return; // skip hidden elements

      // Store initial layout style and DOM parent structure
      const parent = el.parentNode;
      const nextSibling = el.nextSibling;
      const originalStyle = {
        position: el.style.position,
        left: el.style.left,
        top: el.style.top,
        width: el.style.width,
        height: el.style.height,
        margin: el.style.margin,
        transform: el.style.transform,
        zIndex: el.style.zIndex,
        transition: el.style.transition
      };

      // Set fixed viewport bounds
      el.style.width = `${rect.width}px`;
      el.style.height = `${rect.height}px`;
      el.style.left = `${rect.left}px`;
      el.style.top = `${rect.top}px`;
      
      el.classList.add('physics-body-active');
      
      // Portal to body to escape parent transforms and offsets
      document.body.appendChild(el);

      // Create Matter physics body centered at element center
      const cx = rect.left + rect.width / 2;
      const cy = rect.top + rect.height / 2;
      const body = Bodies.rectangle(cx, cy, rect.width, rect.height, {
        restitution: 0.75, // bouncy
        friction: 0.1,
        frictionAir: 0.025
      });

      // Link body to DOM element
      body.domElement = el;
      elementBodies.push({
        el: el,
        body: body,
        parent: parent,
        nextSibling: nextSibling,
        originalStyle: originalStyle
      });

      World.add(world, body);
    });

    // 7. Enable dragging
    const mouse = Mouse.create(document.body);
    mouseConstraint = MouseConstraint.create(engine, {
      mouse: mouse,
      constraint: {
        stiffness: 0.2,
        render: { visible: false }
      }
    });

    World.add(world, mouseConstraint);

    // 8. Tick updates loop
    runner = Runner.create();
    Runner.run(runner, engine);

    Matter.Events.on(engine, 'afterUpdate', () => {
      elementBodies.forEach(item => {
        const { el, body } = item;
        const x = body.position.x - el.offsetWidth / 2;
        const y = body.position.y - el.offsetHeight / 2;
        el.style.transform = `translate3d(${x - parseFloat(el.style.left)}px, ${y - parseFloat(el.style.top)}px, 0) rotate(${body.angle}rad)`;
      });

      // Update custom spawned spheres positions
      spawnedSpheres.forEach(sphereItem => {
        const { el, body } = sphereItem;
        el.style.left = `${body.position.x}px`;
        el.style.top = `${body.position.y}px`;
      });
    });

    // 9. Spawn exactly one big sphere at the start (no click-spawning anymore)
    spawnSphere(viewWidth / 2, viewHeight / 4);
  }

  function spawnSphere(x, y) {
    const { Bodies, World, Body } = Matter;
    const radius = 24; // 48px diameter (larger sphere)

    // Create DOM element
    const sphereEl = document.createElement('div');
    sphereEl.className = 'physics-sphere';
    document.body.appendChild(sphereEl);

    // Create physical body
    const sphereBody = Bodies.circle(x, y, radius, {
      restitution: 0.85,
      friction: 0.05,
      density: 0.005
    });

    // Give it a small starting velocity
    Body.setVelocity(sphereBody, {
      x: (Math.random() - 0.5) * 6,
      y: -5 // throw upwards slightly
    });

    World.add(engine.world, sphereBody);

    spawnedSpheres.push({
      el: sphereEl,
      body: sphereBody
    });
  }

  function stopPhysicsPlayground() {
    physicsActive = false;
    toggleBtn.classList.remove('active');
    toggleBtn.innerText = 'Antigravity Mode';

    // 1. Destroy Matter runner and engine
    if (runner) {
      Matter.Runner.stop(runner);
    }
    if (engine) {
      Matter.World.clear(engine.world, false);
      Matter.Engine.clear(engine);
    }

    // 2. Remove spawned spheres from DOM
    spawnedSpheres.forEach(item => {
      if (item.el && item.el.parentNode) {
        item.el.parentNode.removeChild(item.el);
      }
    });
    spawnedSpheres = [];

    // 3. Restore original styles and layout positions
    elementBodies.forEach(item => {
      const { el, parent, nextSibling, originalStyle } = item;
      el.classList.remove('physics-body-active');
      
      // Restore CSS
      el.style.position = originalStyle.position;
      el.style.left = originalStyle.left;
      el.style.top = originalStyle.top;
      el.style.width = originalStyle.width;
      el.style.height = originalStyle.height;
      el.style.margin = originalStyle.margin;
      el.style.transform = originalStyle.transform;
      el.style.zIndex = originalStyle.zIndex;
      el.style.transition = originalStyle.transition;
      
      // Put back in original DOM place
      if (nextSibling) {
        parent.insertBefore(el, nextSibling);
      } else {
        parent.appendChild(el);
      }
    });

    elementBodies = [];

    // 4. Restore screen scrolling and position
    document.body.style.overflow = 'auto';
    if (window.lenis) {
      window.lenis.start();
    }
    window.scrollTo(0, originalScrollY);
  }

  // Helpers
  function showFeedback(element, message, status) {
    element.innerText = message;
    element.className = 'form-message'; // Reset
    
    if (status === 'error') {
      element.classList.add('error');
    } else if (status === 'success') {
      element.classList.add('success');
    }
  }
});
