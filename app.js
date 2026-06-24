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
