import React, { useState } from 'react';

function FindingsTable({ findings = [] }) {
  const [filter, setFilter] = useState('');
  const [severityFilter, setSeverityFilter] = useState('all');

  const filtered = findings.filter((f) => {
    if (severityFilter !== 'all' && f.severity !== severityFilter) return false;
    if (filter && !f.title.toLowerCase().includes(filter.toLowerCase()) &&
        !f.target.toLowerCase().includes(filter.toLowerCase())) return false;
    return true;
  });

  const severityColor = (s) => {
    switch (s) {
      case 'critical': return '#ff1744';
      case 'high':     return '#ff9100';
      case 'medium':   return '#ffea00';
      case 'low':      return '#64dd17';
      case 'info':     return '#448aff';
      default:         return '#888';
    }
  };

  return (
    <div>
      <h2 style={{ color: '#ff5252' }}>Findings</h2>
      <div style={{ marginBottom: '15px', display: 'flex', gap: '10px' }}>
        <input
          type="text"
          placeholder="Filter findings..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          style={{
            padding: '8px 12px',
            background: '#1a1a1a',
            color: '#e0e0e0',
            border: '1px solid #333',
            borderRadius: '4px',
            flex: 1,
          }}
        />
        <select
          value={severityFilter}
          onChange={(e) => setSeverityFilter(e.target.value)}
          style={{
            padding: '8px 12px',
            background: '#1a1a1a',
            color: '#e0e0e0',
            border: '1px solid #333',
            borderRadius: '4px',
          }}
        >
          <option value="all">All severities</option>
          <option value="critical">Critical</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
          <option value="info">Info</option>
        </select>
      </div>
      {filtered.length === 0 ? (
        <p style={{ color: '#666' }}>No findings to display.</p>
      ) : (
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '14px' }}>
          <thead>
            <tr style={{ background: '#1a1a1a' }}>
              <th style={{ padding: '10px', textAlign: 'left', borderBottom: '2px solid #333' }}>Target</th>
              <th style={{ padding: '10px', textAlign: 'left', borderBottom: '2px solid #333' }}>Title</th>
              <th style={{ padding: '10px', textAlign: 'center', borderBottom: '2px solid #333' }}>Severity</th>
              <th style={{ padding: '10px', textAlign: 'left', borderBottom: '2px solid #333' }}>CVE</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((f, i) => (
              <tr key={f.id || i} style={{ background: i % 2 === 0 ? '#0d0d0d' : '#111' }}>
                <td style={{ padding: '8px 10px', borderBottom: '1px solid #222' }}>{f.target}</td>
                <td style={{ padding: '8px 10px', borderBottom: '1px solid #222' }}>{f.title}</td>
                <td style={{ padding: '8px 10px', borderBottom: '1px solid #222', textAlign: 'center' }}>
                  <span style={{
                    background: severityColor(f.severity),
                    color: '#000',
                    padding: '2px 8px',
                    borderRadius: '10px',
                    fontSize: '12px',
                    fontWeight: 'bold',
                  }}>
                    {f.severity}
                  </span>
                </td>
                <td style={{ padding: '8px 10px', borderBottom: '1px solid #222', color: '#448aff' }}>
                  {f.cve || '-'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

export default FindingsTable;
