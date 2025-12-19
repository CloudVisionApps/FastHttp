import React, { useEffect, useState } from 'react'
import { statsAPI, serverAPI } from '../api/config'
import './Dashboard.css'

function Dashboard() {
  const [stats, setStats] = useState(null)
  const [serverStatus, setServerStatus] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      const [statsRes, statusRes] = await Promise.all([
        statsAPI.get(),
        serverAPI.getStatus(),
      ])
      setStats(statsRes.data)
      setServerStatus(statusRes.data)
    } catch (error) {
      console.error('Failed to load dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleReload = async () => {
    try {
      await serverAPI.reload()
      alert('Server reloaded successfully')
      loadData()
    } catch (error) {
      alert('Failed to reload server: ' + error.message)
    }
  }

  if (loading) {
    return <div className="loading">Loading...</div>
  }

  return (
    <div className="dashboard">
      <h1>Dashboard</h1>

      <div className="stats-grid">
        <div className="stat-card">
          <h3>Server Status</h3>
          <p className="stat-value">{serverStatus?.status || 'Unknown'}</p>
        </div>

        <div className="stat-card">
          <h3>Virtual Hosts</h3>
          <p className="stat-value">{stats?.virtualHosts || 0}</p>
        </div>

        <div className="stat-card">
          <h3>Active Ports</h3>
          <p className="stat-value">{stats?.ports?.length || 0}</p>
          <p className="stat-detail">{stats?.ports?.join(', ') || 'None'}</p>
        </div>

        <div className="stat-card">
          <h3>MIME Types</h3>
          <p className="stat-value">{stats?.mimeTypes || 0}</p>
        </div>
      </div>

      <div className="card">
        <h3>Server Control</h3>
        <div className="control-buttons">
          <button className="btn btn-primary" onClick={handleReload}>
            Reload Configuration
          </button>
          <button className="btn btn-danger" onClick={() => alert('Restart not implemented yet')}>
            Restart Server
          </button>
        </div>
      </div>
    </div>
  )
}

export default Dashboard
