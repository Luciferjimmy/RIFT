/* ==========================================================================
   RIFT Waitlist Main JS: Lenis Scroll, GSAP Parallax, and Scrollytelling
   ========================================================================== */

document.addEventListener('DOMContentLoaded', () => {
  
  // Register GSAP ScrollTrigger
  gsap.registerPlugin(ScrollTrigger);

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
    const nameValue = footerNameInput.value.trim();
    const emailValue = footerEmailInput.value.trim();

    if (!emailValue) {
      showFeedback(footerFeedback, 'Email address is required.', 'error');
      return;
    }
    if (!validateEmail(emailValue)) {
      showFeedback(footerFeedback, 'Please enter a valid email address.', 'error');
      return;
    }

    // Start simulation
    footerSubmitBtn.disabled = true;
    footerSubmitBtn.querySelector('span').innerText = 'Adding user...';
    showFeedback(footerFeedback, '', '');

    setTimeout(() => {
      // Simulate successful registration
      gsap.to(footerForm, {
        opacity: 0,
        y: -15,
        duration: 0.3,
        onComplete: () => {
          footerForm.style.display = 'none';
          successBox.style.display = 'flex';
          successEmailDisplay.innerText = emailValue;
          
          // Animate success ring rotates/scale
          gsap.from('.success-ring', { rotate: -180, scale: 0, duration: 0.8, ease: 'back.out(1.5)' });
          gsap.from('.success-checkmark', { scale: 0, delay: 0.3, duration: 0.4, ease: 'back.out(1.8)' });
        }
      });
    }, 1200);
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
