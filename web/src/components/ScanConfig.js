import React, { useState } from 'react';

function ScanConfig({ targets = [], onAddTarget, onRemoveTarget, onStartScan, loading }) {
  const [newTarget, setNewTarget] = useState('');
  const [ports, setPorts] = useState('22,80,443,8080,8443');
  const [modules, setModules] = useState(['tcp_syn', 'service_detect', 'dns_recon', 'cve_match']);

  const handleAddTarget = () => {
    if (newTarget.trim()) {
      onAddTarget(newTarget.trim());
      setNewTarget('');
    }
  };

  const handleStartScan = () => {
    if (targets.length === 0) return;
    onStartScan({
      targets: targets,
      ports: ports.split(',').map(p => parseInt(p.trim())).filter(Boolean),
      modules: modules,
    });
  };

  const toggleModule = (mod) => {
    setModules(prev =>
      prev.includes(mod) ? prev.filter(m => m !== mod) : [...prev, mod]
    );
  };

  const availableModules = [
    { id: 'tcp_syn', label: 'TCP SYN' },
    { id: 'udp', label: 'UDP' },
    { id: 'service_detect', label: 'Service Detect' },
    { id: 'dns_recon', label: 'DNS Recon' },
    { id: 'subdomain', label: 'Subdomain' },
    { id: 'dirbuster', label: 'Dir Buster' },
    { id: 'fuzzer', label: 'Fuzzer' },
    { id: 'cve_match', label: 'CVE Match' },
    { id: 'nuclei', label: 'Nuclei' },
    { id: 'github', label: 'GitHub' },
    { id: 'shodan', label: 'Shodan' },
    { id: 'cert', label: 'Cert Transparency' },
  ];

  return (
    <div>
      <h2 style={{ color: '#ff5252' }}>Scan Configuration</h2>

      <div style={{ marginBottom: '15px' }}>
        <h4 style={{ color: '#aaa', margin: '10px 0 5px' }}>Targets</h4>
        <div style={{ display: 'flex', gap: '10px', marginBottom: '10px' }}>
          <input
            type="text"
            placeholder="example.com or 10.0.0.1"
            value={newTarget}
            onChange={(e) => setNewTarget(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAddTarget()}
            style={{
              padding: '8px 12px',
              background: '#1a1a1a',
              color: '#e0e0e0',
              border: '1px solid #333',
              borderRadius: '4px',
              flex: 1,
            }}
          />
          <button
            onClick={handleAddTarget}
            style={{
              padding: '8px 16px',
              background: '#ff5252',
              color: '#fff',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontWeight: 'bold',
            }}
          >
            Add
          </button>
        </div>
        {targets.length > 0 && (
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '5px' }}>
            {targets.map((t, i) => (
              <span
                key={i}
                style={{
                  background: '#1a1a1a',
                  color: '#448aff',
                  padding: '4px 10px',
                  borderRadius: '12px',
                  fontSize: '13px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '5px',
                }}
              >
                {t}
                <button
                  onClick={() => onRemoveTarget(t)}
                  style={{
                    background: 'transparent',
                    color: '#ff5252',
                    border: 'none',
                    cursor: 'pointer',
                    fontWeight: 'bold',
                    fontSize: '14px',
                  }}
                >
                  ×
                </button>
              </span>
            ))}
          </div>
        )}
      </div>

      <div style={{ marginBottom: '15px' }}>
        <h4 style={{ color: '#aaa', margin: '10px 0 5px' }}>Ports</h4>
        <input
          type="text"
          value={ports}
          onChange={(e) => setPorts(e.target.value)}
          style={{
            padding: '8px 12px',
            background: '#1a1a1a',
            color: '#e0e0e0',
            border: '1px solid #333',
            borderRadius: '4px',
            width: '100%',
          }}
        />
      </div>

      <div style={{ marginBottom: '15px' }}>
        <h4 style={{ color: '#aaa', margin: '10px 0 5px' }}>Modules</h4>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
          {availableModules.map((mod) => (
            <button
              key={mod.id}
              onClick={() => toggleModule(mod.id)}
              style={{
                padding: '6px 14px',
                background: modules.includes(mod.id) ? '#ff5252' : '#1a1a1a',
                color: modules.includes(mod.id) ? '#fff' : '#888',
                border: `1px solid ${modules.includes(mod.id) ? '#ff5252' : '#333'}`,
                borderRadius: '16px',
                cursor: 'pointer',
                fontSize: '13px',
              }}
            >
              {mod.label}
            </button>
          ))}
        </div>
      </div>

      <button
        onClick={handleStartScan}
        disabled={loading || targets.length === 0}
        style={{
          padding: '12px 24px',
          background: loading ? '#555' : '#448aff',
          color: '#fff',
          border: 'none',
          borderRadius: '4px',
          cursor: loading || targets.length === 0 ? 'not-allowed' : 'pointer',
          fontWeight: 'bold',
          fontSize: '16px',
          width: '100%',
        }}
      >
        {loading ? 'Scanning...' : 'Start Scan'}
      </button>
    </div>
  );
}

export default ScanConfig;
