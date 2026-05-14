import React, { useRef, useMemo } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { OrbitControls, Sphere, Text } from '@react-three/drei';

// SwarmNode renders a single peer as a sphere with label.
function SwarmNode({ position, color, name, size = 0.3 }) {
  const meshRef = useRef();
  
  useFrame((state, delta) => {
    if (meshRef.current) {
      meshRef.current.rotation.y += delta * 0.5;
    }
  });

  return (
    <group position={position}>
      <Sphere ref={meshRef} args={[size, 16, 16]}>
        <meshStandardMaterial color={color} roughness={0.3} metalness={0.7} />
      </Sphere>
      <Text
        position={[0, size + 0.3, 0]}
        fontSize={0.2}
        color="white"
        anchorX="center"
        anchorY="middle"
        outlineWidth={0.02}
        outlineColor="black"
      >
        {name}
      </Text>
    </group>
  );
}

// Generate positions on a sphere surface for given number of nodes.
function generatePositions(count, radius = 3) {
  const positions = [];
  const phi = Math.PI * (3 - Math.sqrt(5)); // golden angle

  for (let i = 0; i < count; i++) {
    const y = 1 - (i / (count - 1 || 1)) * 2;
    const radiusAtY = Math.sqrt(1 - y * y);
    const theta = phi * i;
    
    const x = Math.cos(theta) * radiusAtY;
    const z = Math.sin(theta) * radiusAtY;
    
    positions.push([x * radius, y * radius, z * radius]);
  }
  return positions;
}

function SwarmMap({ peers = [] }) {
  const positions = useMemo(() => generatePositions(peers.length || 1, 4), [peers.length]);

  const getColor = (mode) => {
    switch (mode) {
      case 'strategic': return '#ff5252';
      case 'scanner': return '#448aff';
      case 'hybrid': return '#69f0ae';
      default: return '#888888';
    }
  };

  return (
    <div style={{ width: '100%', height: '400px', background: '#111', borderRadius: '8px', overflow: 'hidden' }}>
      <Canvas camera={{ position: [6, 3, 8], fov: 50 }}>
        <ambientLight intensity={0.4} />
        <pointLight position={[10, 10, 10]} intensity={0.8} />
        <OrbitControls enableZoom={true} enablePan={true} />
        
        {peers.length === 0 && (
          <Text position={[0, 0, 0]} fontSize={0.5} color="gray">
            No peers connected
          </Text>
        )}
        
        {peers.map((peer, i) => (
          <SwarmNode
            key={peer.id || i}
            position={positions[i] || [0, 0, 0]}
            color={getColor(peer.mode)}
            name={peer.id ? peer.id.substring(0, 8) : '?'}
            size={0.3 + (peer.load ? peer.load * 0.2 : 0)}
          />
        ))}
      </Canvas>
    </div>
  );
}

export default SwarmMap;
