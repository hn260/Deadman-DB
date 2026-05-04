import React, { useState, useEffect } from 'react';
import { ShieldCheck, Database, Clock, DownloadCloud, Activity, RefreshCw } from 'lucide-react';

function App() {
  const [snapshots, setSnapshots] = useState([]);
  const [loading, setLoading] = useState(true);

  // Use environment variable for API URL in production, default to empty string (which uses the Vite proxy locally)
  const apiUrl = import.meta.env.VITE_API_URL || '';

  // In a real app, this would fetch from /api/v1/snapshots
  // Since we might not have the Go server running, we use mock data for the UI demonstration
  useEffect(() => {
    // Example fetch call for when you connect the real backend:
    // fetch(`${apiUrl}/api/v1/snapshots`).then(res => res.json()).then(...)
    setTimeout(() => {
      setSnapshots([
        { id: 'snap_1715421295', db_name: 'production_pg', timestamp: 1715421295, size: 154859012, status: 'success' },
        { id: 'snap_1715334895', db_name: 'production_pg', timestamp: 1715334895, size: 154792100, status: 'success' },
        { id: 'snap_1715248495', db_name: 'production_pg', timestamp: 1715248495, size: 154512330, status: 'success' },
        { id: 'snap_1715162095', db_name: 'legacy_mongo', timestamp: 1715162095, size: 8432100, status: 'failed' },
        { id: 'snap_1715075695', db_name: 'production_pg', timestamp: 1715075695, size: 154100200, status: 'success' },
      ]);
      setLoading(false);
    }, 1500);
  }, []);

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (timestamp) => {
    return new Date(timestamp * 1000).toLocaleString();
  };

  const totalSize = snapshots.reduce((acc, curr) => acc + (curr.status === 'success' ? curr.size : 0), 0);

  return (
    <div className="app-container">
      <header>
        <div className="logo-container">
          <div className="logo-icon">
            <ShieldCheck color="#000" size={24} />
          </div>
          <h1 className="logo-text">Deadman DB</h1>
        </div>
        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--accent-green)', fontSize: '0.875rem', fontWeight: 500 }}>
            <Activity size={16} /> Daemon Active
          </div>
          <button className="btn btn-primary">
            <Database size={16} /> Trigger Backup
          </button>
        </div>
      </header>

      <div className="dashboard-grid">
        <div className="glass-panel metric-card">
          <span className="metric-title">Total Protected Databases</span>
          <span className="metric-value">2</span>
        </div>
        <div className="glass-panel metric-card">
          <span className="metric-title">Total Storage Used</span>
          <span className="metric-value">{loading ? '...' : formatBytes(totalSize)}</span>
        </div>
        <div className="glass-panel metric-card">
          <span className="metric-title">Last Backup</span>
          <span className="metric-value" style={{ fontSize: '1.5rem', marginTop: '0.5rem' }}>
            {loading ? '...' : 'Today at 2:00 AM'}
          </span>
        </div>

        <div className="glass-panel snapshots-section">
          <div className="section-header">
            <h2>Recent Snapshots</h2>
            <button className="btn">
              <RefreshCw size={16} /> Refresh
            </button>
          </div>

          {loading ? (
            <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
              Loading snapshot history...
            </div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Snapshot ID</th>
                  <th>Database</th>
                  <th>Timestamp</th>
                  <th>Size</th>
                  <th>Status</th>
                  <th style={{ textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {snapshots.map((snap) => (
                  <tr key={snap.id}>
                    <td style={{ fontFamily: 'monospace', color: 'var(--text-secondary)' }}>{snap.id}</td>
                    <td style={{ fontWeight: 500 }}>{snap.db_name}</td>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
                        <Clock size={14} /> {formatDate(snap.timestamp)}
                      </div>
                    </td>
                    <td>{formatBytes(snap.size)}</td>
                    <td>
                      <span className={`status-badge ${snap.status === 'success' ? 'status-success' : 'status-failed'}`}>
                        {snap.status}
                      </span>
                    </td>
                    <td style={{ textAlign: 'right' }}>
                      <button 
                        className="btn btn-danger" 
                        style={{ padding: '0.35rem 0.75rem', fontSize: '0.875rem', display: 'inline-flex' }}
                        disabled={snap.status !== 'success'}
                      >
                        <DownloadCloud size={14} /> Restore
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;
