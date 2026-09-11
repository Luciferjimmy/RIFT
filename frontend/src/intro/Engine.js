import * as THREE from 'three';
import { createDiorama } from './Diorama.js';
import { initScroller } from './Scroller.js';

export class Engine {
    constructor(container) {
        this.container = container;
        
        // Scene setup
        this.scene = new THREE.Scene();
        // Match the HTML background color
        this.scene.background = new THREE.Color('#000000'); 
        
        // Gentle fog to blend platforms out in the distance
        this.scene.fog = new THREE.Fog('#000000', 50, 300);
        
        // Camera setup (Perspective but positioned isometrically by Scroller)
        this.camera = new THREE.PerspectiveCamera(45, window.innerWidth / window.innerHeight, 0.1, 1000);
        
        // Soft Minimalist Lighting
        const ambientLight = new THREE.AmbientLight(0xffffff, 0.6);
        this.scene.add(ambientLight);
        
        const directionalLight = new THREE.DirectionalLight(0xffffff, 0.8);
        directionalLight.position.set(50, 100, 50);
        this.scene.add(directionalLight);
        
        const fillLight = new THREE.DirectionalLight(0xaaccff, 0.3);
        fillLight.position.set(-50, 50, -50);
        this.scene.add(fillLight);

        // Renderer setup
        this.renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
        this.renderer.setSize(window.innerWidth, window.innerHeight);
        this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
        this.container.appendChild(this.renderer.domElement);
        
        // Build the Diorama World
        this.diorama = createDiorama();
        this.scene.add(this.diorama.diorama);

        // Scroll Logic
        this.scroller = initScroller(this.camera, this.diorama);

        // Resize handler
        this.onWindowResize = this.onWindowResize.bind(this);
        window.addEventListener('resize', this.onWindowResize);

        // Animation Loop
        this.animationId = null;
        this.animate = this.animate.bind(this);
        this.animate();
    }

    onWindowResize() {
        this.camera.aspect = window.innerWidth / window.innerHeight;
        this.camera.updateProjectionMatrix();
        this.renderer.setSize(window.innerWidth, window.innerHeight);
    }

    animate() {
        this.animationId = requestAnimationFrame(this.animate);
        
        // Update scroller physics (Ball jump & camera move)
        this.scroller.update();

        this.renderer.render(this.scene, this.camera);
    }

    dispose() {
        cancelAnimationFrame(this.animationId);
        window.removeEventListener('resize', this.onWindowResize);
        this.scroller.dispose();
        
        this.renderer.dispose();
        if(this.renderer.domElement.parentNode) {
            this.renderer.domElement.parentNode.removeChild(this.renderer.domElement);
        }
    }
}
