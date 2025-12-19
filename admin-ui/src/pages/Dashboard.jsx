import React, { useEffect, useState } from 'react'
import { statsAPI, serverAPI } from '../api/config'

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
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-zinc-500 text-lg">Loading...</div>
      </div>
    )
  }

  const statusColor = serverStatus?.status === 'running' ? 'text-green-600' : 'text-red-600'

  return (
    <div>
      <h1 className="text-3xl font-bold text-zinc-900 mb-8">Dashboard</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-white rounded-2xl shadow-sm p-6 border border-zinc-200">
          <h3 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-2">
            Server Status
          </h3>
          <p className={`text-3xl font-semibold ${statusColor}`}>
            {serverStatus?.status || 'Unknown'}
          </p>
        </div>

        <div className="bg-white rounded-2xl shadow-sm p-6 border border-zinc-200">
          <h3 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-2">
            Virtual Hosts
          </h3>
          <p className="text-3xl font-semibold text-zinc-700">{stats?.virtualHosts || 0}</p>
        </div>

        <div className="bg-white rounded-2xl shadow-sm p-6 border border-zinc-200">
          <h3 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-2">
            Active Ports
          </h3>
          <p className="text-3xl font-semibold text-zinc-700">{stats?.ports?.length || 0}</p>
          <p className="text-xs text-zinc-500 mt-1">{stats?.ports?.join(', ') || 'None'}</p>
        </div>

        <div className="bg-white rounded-2xl shadow-sm p-6 border border-zinc-200">
          <h3 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-2">
            MIME Types
          </h3>
          <p className="text-3xl font-semibold text-zinc-700">{stats?.mimeTypes || 0}</p>
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-sm p-6 border border-zinc-200">
        <h3 className="text-lg font-semibold text-zinc-900 mb-4">Server Control</h3>
        <div className="flex gap-3">
          <button
            className="px-4 py-2 bg-zinc-700 text-white rounded-xl hover:bg-zinc-800 transition-colors font-medium"
            onClick={handleReload}
          >
            Reload Configuration
          </button>
          <button
            className="px-4 py-2 bg-red-600 text-white rounded-xl hover:bg-red-700 transition-colors font-medium"
            onClick={() => alert('Restart not implemented yet')}
          >
            Restart Server
          </button>
        </div>
      </div>
    </div>
  )
}

export default Dashboard
