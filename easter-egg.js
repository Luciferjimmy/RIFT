/* ==========================================================================
   RIFT Easter Egg — Cinematic Reveal JS
   Lenis, GSAP ScrollTrigger, True Ferris Wheel, Scroll-Driven Glow
   ========================================================================== */

document.addEventListener('DOMContentLoaded', () => {

  gsap.registerPlugin(ScrollTrigger);

  /* ------------------------------------------------------------------------
     1. Lenis Smooth Scroll
     ------------------------------------------------------------------------ */
  const lenis = new Lenis({
    duration: 1.4,
    easing: (t) => Math.min(1, 1.001 - Math.pow(2, -10 * t)),
    smoothWheel: true,
    touchMultiplier: 1.5,
    infinite: false,
  });

  function raf(time) {
    lenis.raf(time);
    requestAnimationFrame(raf);
  }
  requestAnimationFrame(raf);

  lenis.on('scroll', ScrollTrigger.update);
  gsap.ticker.add((time) => lenis.raf(time * 1000));
  gsap.ticker.lagSmoothing(0);

  window.addEventListener('load', () => ScrollTrigger.refresh());

  /* ------------------------------------------------------------------------
     2. Interactive Hover Easter Egg & Mouse Move Parallax
     ------------------------------------------------------------------------ */
  const heroSection = document.getElementById('ee-hero');
  const singleImg = document.querySelector('.ee-hero-single-img');
  const bubble = document.getElementById('ee-speech-bubble');
  const speechText = document.getElementById('ee-speech-text');

  const warningMessages = [
    "DO NOT DISTURB!",
    "I said scroll down, not hover me!",
    "Who invited you here anyway?!",
    "Are you trying to crash my Wine sandbox?",
    "Step away from the character!",
    "Launch the game already!",
    "You're testing my patience...",
    "Back off!",
    "One more hover and we translate you to Metal!"
  ];

  // Synthesize low-frequency rumble / growl using Web Audio API
  let audioCtx = null;
  function playGrowlRumble() {
    try {
      if (!audioCtx) {
        audioCtx = new (window.AudioContext || window.webkitAudioContext)();
      }
      if (audioCtx.state === 'suspended') {
        audioCtx.resume();
      }
      
      const osc = audioCtx.createOscillator();
      const gainNode = audioCtx.createGain();
      const filter = audioCtx.createBiquadFilter();

      osc.type = 'sawtooth';
      osc.frequency.setValueAtTime(80, audioCtx.currentTime);
      osc.frequency.exponentialRampToValueAtTime(30, audioCtx.currentTime + 1.2);

      filter.type = 'lowpass';
      filter.frequency.setValueAtTime(140, audioCtx.currentTime);

      gainNode.gain.setValueAtTime(0.35, audioCtx.currentTime);
      gainNode.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + 1.2);

      osc.connect(filter);
      filter.connect(gainNode);
      gainNode.connect(audioCtx.destination);

      osc.start();
      osc.stop(audioCtx.currentTime + 1.2);
    } catch (e) {
      console.warn("Audio Context blocked by browser autoplay policy:", e);
    }
  }

  // Speak message using Web Speech API (low pitch, robotic warning)
  function speakWarning(text) {
    if ('speechSynthesis' in window) {
      window.speechSynthesis.cancel();
      const utterance = new SpeechSynthesisUtterance(text);
      utterance.rate = 1.05;
      utterance.pitch = 0.55; // Low-pitched warning voice
      
      const voices = window.speechSynthesis.getVoices();
      const maleVoice = voices.find(v => v.name.includes('Male') || v.name.includes('David') || v.name.includes('Google US English'));
      if (maleVoice) {
        utterance.voice = maleVoice;
      }
      window.speechSynthesis.speak(utterance);
    }
  }

  // Manga frustration symbol particle generation
  let angerInterval = null;
  function startAngerSpawning() {
    stopAngerSpawning();
    spawnAngerSymbol();
    angerInterval = setInterval(spawnAngerSymbol, 220);
  }

  function stopAngerSpawning() {
    if (angerInterval) {
      clearInterval(angerInterval);
      angerInterval = null;
    }
  }

  function spawnAngerSymbol() {
    const wrap = document.querySelector('.ee-hero-image-wrap');
    if (!wrap) return;
    
    const symbol = document.createElement('div');
    symbol.className = 'ee-anger-symbol';
    symbol.innerHTML = '💢';
    
    // Spawn near the head area (top-middle of wrapper box)
    const xPos = 25 + Math.random() * 50; 
    const yPos = 5 + Math.random() * 30; 
    
    symbol.style.left = `${xPos}%`;
    symbol.style.top = `${yPos}%`;
    
    wrap.appendChild(symbol);
    
    // Upward float drift dynamics
    const driftX = (Math.random() - 0.5) * 40; 
    const driftY = -50 - Math.random() * 40; 
    const rot = (Math.random() - 0.5) * 60; 
    
    gsap.fromTo(symbol, {
      scale: 0.2,
      opacity: 0,
      rotation: 0
    }, {
      scale: 1.1 + Math.random() * 0.4,
      opacity: 1,
      rotation: rot,
      x: driftX,
      y: driftY,
      duration: 0.9,
      ease: 'power1.out',
      onComplete: () => {
        gsap.to(symbol, {
          opacity: 0,
          scale: 0.6,
          duration: 0.2,
          onComplete: () => symbol.remove()
        });
      }
    });
  }

  // Parallax + Interaction bindings
  if (heroSection && singleImg) {
    // Parallax
    heroSection.addEventListener('mousemove', (e) => {
      const rect = heroSection.getBoundingClientRect();
      const x = (e.clientX - rect.left) / rect.width - 0.5;
      const y = (e.clientY - rect.top) / rect.height - 0.5;
      
      gsap.to(singleImg, {
        x: x * -15,
        y: y * -15,
        duration: 0.6,
        ease: 'power2.out',
        overwrite: 'auto'
      });
    });

    heroSection.addEventListener('mouseleave', () => {
      gsap.to(singleImg, {
        x: 0,
        y: 0,
        duration: 0.8,
        ease: 'power2.out',
        overwrite: 'auto'
      });
    });

    // Hover triggers
    singleImg.addEventListener('mouseenter', () => {
      singleImg.src = 'Players/man_gun.png';
      singleImg.classList.add('shaking');
      
      const randomMsg = warningMessages[Math.floor(Math.random() * warningMessages.length)];
      if (speechText && bubble) {
        speechText.textContent = randomMsg;
        bubble.classList.add('active');
      }
      
      playGrowlRumble();
      speakWarning(randomMsg);
      startAngerSpawning();
    });

    singleImg.addEventListener('mouseleave', () => {
      singleImg.src = 'Players/man_bg.png';
      singleImg.classList.remove('shaking');
      
      if (bubble) {
        bubble.classList.remove('active');
      }
      
      if ('speechSynthesis' in window) {
        window.speechSynthesis.cancel();
      }
      
      stopAngerSpawning();
    });
  }

  /* ------------------------------------------------------------------------
     3. Hero Entrance Animations
     ------------------------------------------------------------------------ */
  const tlHero = gsap.timeline({ delay: 0.3 });

  tlHero.to('.ee-hero-headline', {
    y: 0,
    opacity: 1,
    duration: 1.2,
    ease: 'power4.out'
  })
  .fromTo('.ee-hero-image-wrap', {
    y: 30,
    opacity: 0
  }, {
    y: 0,
    opacity: 1,
    duration: 1.0,
    ease: 'power3.out'
  }, '-=0.8')
  .to('.ee-hero-sub', {
    y: 0,
    opacity: 1,
    duration: 0.8,
    ease: 'power3.out'
  }, '-=0.6');

  /* ------------------------------------------------------------------------
     4. Problem Section — Line-by-line Reveal
     ------------------------------------------------------------------------ */
  function createLineReveal(sectionSelector) {
    const lines = document.querySelectorAll(`${sectionSelector} .ee-reveal-inner`);
    
    gsap.to(lines, {
      scrollTrigger: {
        trigger: sectionSelector,
        start: 'top center+=100',
        toggleActions: 'play none none none'
      },
      y: 0,
      opacity: 1,
      duration: 0.8,
      stagger: 0.15,
      ease: 'power3.out'
    });
  }

  createLineReveal('#ee-problem');
  createLineReveal('#ee-problem-2');

  /* ------------------------------------------------------------------------
     5. Vision Blocks — Staggered Scroll Reveal
     ------------------------------------------------------------------------ */
  gsap.to('.ee-vision-block', {
    scrollTrigger: {
      trigger: '#ee-vision',
      start: 'top center+=50',
      toggleActions: 'play none none none'
    },
    y: 0,
    opacity: 1,
    duration: 0.9,
    stagger: 0.2,
    ease: 'power3.out'
  });

  /* ------------------------------------------------------------------------
     6. Catalog — Scroll-Driven Glow on Game Names
     
     As the user scrolls through the catalog section, a white glow
     spotlight moves from the top game name to the bottom, tracking
     scroll position. Scroll up = glow moves up.
     ------------------------------------------------------------------------ */
  const gameNames = document.querySelectorAll('.ee-game-name');

  // Reveal game names on scroll
  gameNames.forEach((name, index) => {
    gsap.to(name, {
      scrollTrigger: {
        trigger: name,
        start: 'top bottom-=100',
        toggleActions: 'play none none none'
      },
      y: 0,
      opacity: 1,
      duration: 0.7,
      delay: index * 0.08,
      ease: 'power3.out',
    });
  });

  // Optimized Scroll-driven glow: trigger class toggle on center intersection
  gameNames.forEach((name) => {
    ScrollTrigger.create({
      trigger: name,
      start: 'top center+=100',
      end: 'bottom center-=100',
      toggleClass: 'glow-active',
    });
  });

  // Counter reveal
  gsap.to('.ee-catalog-counter', {
    scrollTrigger: {
      trigger: '.ee-catalog-counter',
      start: 'top bottom-=50',
      toggleActions: 'play none none none'
    },
    y: 0,
    opacity: 1,
    duration: 0.8,
    ease: 'power3.out'
  });

  /* ------------------------------------------------------------------------
     7. Close Section — Final Reveal
     ------------------------------------------------------------------------ */
  const tlClose = gsap.timeline({
    scrollTrigger: {
      trigger: '#ee-close',
      start: 'top center+=100',
      toggleActions: 'play none none none'
    }
  });

  tlClose.to('.ee-close-text', {
    y: 0,
    opacity: 1,
    duration: 1,
    ease: 'power4.out'
  })
  .to('.ee-close-sub', {
    y: 0,
    opacity: 1,
    duration: 0.7,
    ease: 'power3.out'
  }, '-=0.5')
  .to('.ee-back-link', {
    y: 0,
    opacity: 1,
    duration: 0.6,
    ease: 'power3.out'
  }, '-=0.3');

});
