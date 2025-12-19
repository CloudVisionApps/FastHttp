import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { virtualHostsAPI } from '../api/config'
import './VirtualHosts.css'

function VirtualHosts() {
  const [virtualHosts, setVirtualHosts] = useState([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState(null)
  const navigate = useNavigate()

  useEffect(() => {
    loadVirtualHosts()
  }, [])

  const loadVirtualHosts = async () => {
    try {
      const res = await virtualHostsAPI.getAll()
      setVirtualHosts(res.data)
    } catch (error) {
      console.error('Failed to load virtual hosts:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (serverName) => {
    if (!confirm(`Are you sure you want to delete ${serverName}?`)) {
      return
    }

    try {
      await virtualHostsAPI.delete(serverName)
      loadVirtualHosts()
    } catch (error) {
      alert('Failed to delete virtual host: ' + error.message)
    }
  }

  if (loading) {
    return <div className="loading">Loading...</div>
  }

  return (
    <div className="virtualhosts">
      <div className="page-header">
        <h1>Virtual Hosts</h1>
        <button className="btn btn-primary" onClick={() => setShowForm(true)}>
          Add Virtual Host
        </button>
      </div>

      {showForm && (
        <VirtualHostForm
          onClose={() => {
            setShowForm(false)
            setEditing(null)
          }}
          onSave={() => {
            loadVirtualHosts()
            setShowForm(false)
            setEditing(null)
          }}
          editing={editing}
        />
      )}

      <div className="card">
        <table className="table">
          <thead>
            <tr>
              <th>Server Name</th>
              <th>Document Root</th>
              <th>Listen Ports</th>
              <th>Locations</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {virtualHosts.length === 0 ? (
              <tr>
                <td colSpan="5" style={{ textAlign: 'center', padding: '20px' }}>
                  No virtual hosts configured
                </td>
              </tr>
            ) : (
              virtualHosts.map((vhost) => (
                <tr key={vhost.serverName}>
                  <td>{vhost.serverName}</td>
                  <td>{vhost.documentRoot}</td>
                  <td>{vhost.listen?.length > 0 ? vhost.listen.join(', ') : 'All ports'}</td>
                  <td>{vhost.locations?.length || 0}</td>
                  <td>
                    <button
                      className="btn btn-sm"
                      onClick={() => navigate(`/locations/${vhost.serverName}`)}
                    >
                      Locations
                    </button>
                    <button
                      className="btn btn-sm"
                      onClick={() => {
                        setEditing(vhost)
                        setShowForm(true)
                      }}
                    >
                      Edit
                    </button>
                    <button
                      className="btn btn-sm btn-danger"
                      onClick={() => handleDelete(vhost.serverName)}
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function VirtualHostForm({ onClose, onSave, editing }) {
  const [formData, setFormData] = useState(
    editing || {
      serverName: '',
      documentRoot: '',
      listen: [],
      user: '',
      group: '',
      directoryIndex: '',
      locations: [],
    }
  )

  const handleSubmit = async (e) => {
    e.preventDefault()
    try {
      if (editing) {
        await virtualHostsAPI.update(editing.serverName, formData)
      } else {
        await virtualHostsAPI.create(formData)
      }
      onSave()
    } catch (error) {
      alert('Failed to save virtual host: ' + (error.response?.data?.error || error.message))
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>{editing ? 'Edit' : 'Add'} Virtual Host</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Server Name *</label>
            <input
              type="text"
              value={formData.serverName}
              onChange={(e) => setFormData({ ...formData, serverName: e.target.value })}
              required
            />
          </div>

          <div className="form-group">
            <label>Document Root *</label>
            <input
              type="text"
              value={formData.documentRoot}
              onChange={(e) => setFormData({ ...formData, documentRoot: e.target.value })}
              required
            />
          </div>

          <div className="form-group">
            <label>Listen Ports (comma-separated, empty for all)</label>
            <input
              type="text"
              value={formData.listen?.join(', ') || ''}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  listen: e.target.value ? e.target.value.split(',').map((p) => p.trim()) : [],
                })
              }
              placeholder="80, 8080"
            />
          </div>

          <div className="form-group">
            <label>User</label>
            <input
              type="text"
              value={formData.user || ''}
              onChange={(e) => setFormData({ ...formData, user: e.target.value })}
            />
          </div>

          <div className="form-group">
            <label>Group</label>
            <input
              type="text"
              value={formData.group || ''}
              onChange={(e) => setFormData({ ...formData, group: e.target.value })}
            />
          </div>

          <div className="form-group">
            <label>Directory Index</label>
            <input
              type="text"
              value={formData.directoryIndex || ''}
              onChange={(e) => setFormData({ ...formData, directoryIndex: e.target.value })}
              placeholder="index.php index.html"
            />
          </div>

          <div className="form-actions">
            <button type="submit" className="btn btn-primary">
              {editing ? 'Update' : 'Create'}
            </button>
            <button type="button" className="btn" onClick={onClose}>
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default VirtualHosts
