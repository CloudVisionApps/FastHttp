import React, { useEffect, useState } from 'react'
import { configAPI } from '../api/config'
import './Config.css'

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
    return <div className="loading">Loading...</div>
  }

  if (!config) {
    return <div>Failed to load configuration</div>
  }

  return (
    <div className="config">
      <div className="page-header">
        <h1>Configuration</h1>
        <div>
          <button className="btn" onClick={handleReload}>
            Reload from File
          </button>
          <button className="btn btn-primary" onClick={handleSave} disabled={saving}>
            {saving ? 'Saving...' : 'Save Configuration'}
          </button>
        </div>
      </div>

      <div className="card">
        <h3>Global Settings</h3>
        <div className="form-group">
          <label>User</label>
          <input
            type="text"
            value={config.user || ''}
            onChange={(e) => setConfig({ ...config, user: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label>Group</label>
          <input
            type="text"
            value={config.group || ''}
            onChange={(e) => setConfig({ ...config, group: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label>Listen Ports (comma-separated)</label>
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
          />
        </div>
        <div className="form-group">
          <label>Directory Index</label>
          <input
            type="text"
            value={config.directoryIndex || ''}
            onChange={(e) => setConfig({ ...config, directoryIndex: e.target.value })}
            placeholder="index.php index.html"
          />
        </div>
        <div className="form-group">
          <label>Admin Port</label>
          <input
            type="text"
            value={config.adminPort || ''}
            onChange={(e) => setConfig({ ...config, adminPort: e.target.value })}
            placeholder="8080"
          />
        </div>
        <div className="form-group">
          <label>
            <input
              type="checkbox"
              checked={config.adminEnabled || false}
              onChange={(e) => setConfig({ ...config, adminEnabled: e.target.checked })}
            />
            Enable Admin API
          </label>
        </div>
      </div>

      <div className="card">
        <h3>Raw JSON Configuration</h3>
        <textarea
          className="json-editor"
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
