import React, { useState, useEffect } from 'react';
import SwarmMap from './components/SwarmMap';
import FindingsTable from './components/FindingsTable';
import ScanConfig from './components/ScanConfig';

const API_BASE = process.env.REACT_APP_API_BASE || '';

function App() {
  const [targets, setTargets] = useState([]);
  const [scans, setScans] = useState([]);
  const [findings, setFindings] = useState([]);
  const [peers, setPeers] = useState([]);
  const [loading, setLoading] = useState(false);

  // Fetch targets on mount
  useEffect(() => {
    fetchTargets();
    fetchPeers();
    fetchFindings();
    const interval = setInterval(() => {
      fetchPeers();
      fetchFindings();
    }, 10000);
    return () => clearInterval(interval);
  }, []);

  const fetchTargets = async () => {
    try {
      const res = await fetch(`${API_BASE}/api/targets`);
      const data = await res.json();
      setTargets(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Failed to fetch targets', err);
    }
  };

  const fetchPeers = async () => {
    try {
      const res = await fetch(`${API_BASE}/api/swarm/peers`);
      const data = await res.json();
      setPeers(Array.isArray(data) ? data : []);
    } catch (err) {
      // Swarm endpoint may not be available yet
    }
  };

  const fetchFindings = async () => {
    try {
      const res = await fetch(`${API_BASE}/api/findings?limit=50`);
      const data = await res.json();
      setFindings(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Failed to fetch findings', err);
    }
  };

  const handleAddTarget = async (target) => {
    try {
      await fetch(`${API_BASE}/api/targets`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target }),
      });
      fetchTargets();
    } catch (err) {
      console.error('Failed to add target', err);
    }
  };

  const handleRemoveTarget = async (target) => {
    try {
      await fetch(`${API_BASE}/api/targets`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target }),
      });
      fetchTargets();
    } catch (err) {
      console.error('Failed to remove target', err);
    }
  };

  const handleStartScan = async (config) => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/api/scans`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      });
      const data = await res.json();
      setScans((prev) => [...prev, data]);
    } catch (err) {
      console.error('Failed to start scan', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: '20px', fontFamily: 'monospace', background: '#0a0a0a', color: '#e0e0e0', minHeight: '100vh' }}>
      <h1 style={{ color: '#ff5252', marginBottom: '5px' }}>Siege</h1>
      <p style={{ color: '#888', marginTop: 0 }}>Black Swarm 3.0 — Distributed Reconnaissance</p>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px', marginTop: '20px' }}>
        <div>
          <ScanConfig
            targets={targets}
            onAddTarget={handleAddTarget}
            onRemoveTarget={handleRemoveTarget}
            onStartScan={handleStartScan}
            loading={loading}
          />
        </div>
        <div>
          <SwarmMap peers={peers} />
        </div>
      </div>

      <div style={{ marginTop: '30px' }}>
        <FindingsTable findings={findings} />
      </div>

      <div style={{ marginTop: '40px', textAlign: 'center', color: '#555', fontSize: '12px' }}>
        Siege v3.0 — {peers.length} nodes online — {findings.length} findings
      </div>
    </div>
  );
}

export default App;
