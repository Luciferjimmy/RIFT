import * as THREE from 'three';

export function createDiorama() {
    const diorama = new THREE.Group();
    
    // Aesthetic Settings matching the Golf game exactly (Black & White wireframe)
    const groundMat = new THREE.MeshBasicMaterial({
        color: 0x111111
    });
    
    const wallMat = new THREE.MeshBasicMaterial({
        color: 0x222222
    });

    const edgeMat = new THREE.LineBasicMaterial({
        color: 0x999999, // Brighter white/grey wireframe lines
        linewidth: 1
    });

    const ballMaterial = new THREE.MeshBasicMaterial({
        color: 0xFFFFFF // Pure white glowing ball
    });

    // Spaced out path to ensure platforms never overlap visually (island size is 40x40)
    const platformsData = [
        { x: 0,   y: 0,    z: 0 },       // Start
        { x: 60,  y: -20,  z: -60 },     // Right, Down, Away
        { x: -30, y: -40,  z: -120 },    // Left, Down, Away
        { x: 50,  y: -20,  z: -180 },    // Right, Up, Away
        { x: -40, y: 10,   z: -240 },    // Left, Up, Away
        { x: 40,  y: -30,  z: -300 },    // Right, Down, Away
        { x: 0,   y: -50,  z: -380 }     // Center, Down, Away (End)
    ];

    const platforms = [];
    
    const boxGeo = new THREE.BoxGeometry(4, 4, 4);
    const boxEdges = new THREE.EdgesGeometry(boxGeo);
    
    platformsData.forEach((data, index) => {
        const islandGroup = new THREE.Group();
        islandGroup.position.set(data.x, data.y + 400, data.z); 
        
        islandGroup.userData = { 
            targetY: data.y, 
            index: index, 
            hasDropped: false 
        };
        
        // Generate a voxel island (10x10 grid for larger landing zones)
        const size = 10;
        for(let i=0; i<size; i++) {
            for(let j=0; j<size; j++) {
                // Random shape logic
                if (Math.random() > 0.2 || (i===3 && j===3)) {
                    // Ground block
                    const block = new THREE.Mesh(boxGeo, groundMat);
                    block.position.set(i*4 - (size*2), 0, j*4 - (size*2));
                    
                    const edges = new THREE.LineSegments(boxEdges, edgeMat);
                    block.add(edges);
                    
                    islandGroup.add(block);
                    
                    // Random wall block on top
                    if (Math.random() > 0.85 && !(i===3 && j===3) && !(i===4 && j===4)) {
                        const wall = new THREE.Mesh(boxGeo, wallMat);
                        wall.position.set(i*4 - (size*2), 4, j*4 - (size*2));
                        
                        const wallEdges = new THREE.LineSegments(boxEdges, edgeMat);
                        wall.add(wallEdges);
                        
                        islandGroup.add(wall);
                    }
                }
            }
        }
        
        diorama.add(islandGroup);
        platforms.push(islandGroup);
    });

    // 2. The Ball (Protagonist) - smaller
    const ballGeo = new THREE.SphereGeometry(1.5, 32, 32);
    const ball = new THREE.Mesh(ballGeo, ballMaterial);
    
    // Add point light to the ball to give it that "glowing" effect on the dark tiles
    const ballLight = new THREE.PointLight(0xffffff, 1.0, 30);
    ballLight.position.set(0, 0, 0);
    ball.add(ballLight);

    ball.position.set(platformsData[0].x, platformsData[0].y + 8, platformsData[0].z);
    diorama.add(ball);

    return { diorama, platforms, platformsData, ball };
}
