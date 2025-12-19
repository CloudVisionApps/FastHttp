import React, { useEffect, useState } from 'react'
import { configAPI } from '../api/config'

function Config() {
  const [config, setConfig] = useState(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    loadConfig()
  }, [])

  const loadConfig = async () => {
    try {
      const res = await configAPI.get()
      setConfig(res.data)
    } catch (error) {
      console.error('Failed to load config:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await configAPI.update(config)
      alert('Configuration saved successfully')
    } catch (error) {
      alert('Failed to save config: ' + (error.response?.data?.error || error.message))
    } finally {
      setSaving(false)
    }
  }

  const handleReload = async () => {
    try {
      await configAPI.reload()
      loadConfig()
      alert('Configuration reloaded from file')
    } catch (error) {
      alert('Failed to reload config: ' + error.message)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-zinc-500 text-lg">Loading...</div>
      </div>
    )
  }

  if (!config) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-2xl p-4 text-red-700">
        Failed to load configuration
      </div>
    )
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold text-zinc-900">Configuration</h1>
        <div className="flex gap-3">
          <button
            className="px-4 py-2 bg-zinc-200 text-zinc-700 rounded-xl hover:bg-zinc-300 transition-colors font-medium"
            onClick={handleReload}
          >
            Reload from File
          </button>
          <button
            className="px-4 py-2 bg-zinc-700 text-white rounded-xl hover:bg-zinc-800 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? 'Saving...' : 'Save Configuration'}
          </button>
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-sm p-6 border border-zinc-200 mb-6">
        <h3 className="text-lg font-semibold text-zinc-900 mb-6">Global Settings</h3>
        <div className="space-y-5">
          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">User</label>
            <input
              type="text"
              value={config.user || ''}
              onChange={(e) => setConfig({ ...config, user: e.target.value })}
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">Group</label>
            <input
              type="text"
              value={config.group || ''}
              onChange={(e) => setConfig({ ...config, group: e.target.value })}
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">
              Listen Ports (comma-separated)
            </label>
            <input
              type="text"
              value={config.listen?.join(', ') || ''}
              onChange={(e) =>
                setConfig({
                  ...config,
                  listen: e.target.value ? e.target.value.split(',').map((p) => p.trim()) : [],
                })
              }
              placeholder="80, 443"
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">Directory Index</label>
            <input
              type="text"
              value={config.directoryIndex || ''}
              onChange={(e) => setConfig({ ...config, directoryIndex: e.target.value })}
              placeholder="index.php index.html"
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">Admin Port</label>
            <input
              type="text"
              value={config.adminPort || ''}
              onChange={(e) => setConfig({ ...config, adminPort: e.target.value })}
              placeholder="8080"
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>
          <div>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={config.adminEnabled || false}
                onChange={(e) => setConfig({ ...config, adminEnabled: e.target.checked })}
                className="w-4 h-4 text-zinc-700 border-zinc-300 rounded-xl focus:ring-zinc-500"
              />
              <span className="text-sm font-medium text-zinc-700">Enable Admin API</span>
            </label>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-sm p-6 border border-zinc-200">
        <h3 className="text-lg font-semibold text-zinc-900 mb-4">Raw JSON Configuration</h3>
        <textarea
          className="w-full px-4 py-3 border border-zinc-300 rounded-xl font-mono text-sm focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none resize-y"
          value={JSON.stringify(config, null, 2)}
          onChange={(e) => {
            try {
              setConfig(JSON.parse(e.target.value))
            } catch (err) {
              // Invalid JSON, ignore
            }
          }}
          rows={20}
        />
      </div>
    </div>
  )
}

export default Config
