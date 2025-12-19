import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { locationsAPI, virtualHostsAPI } from '../api/config'
import './Locations.css'

function Locations() {
  const { serverName } = useParams()
  const navigate = useNavigate()
  const [locations, setLocations] = useState([])
  const [vhost, setVhost] = useState(null)
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState(null)

  useEffect(() => {
    loadData()
  }, [serverName])

  const loadData = async () => {
    try {
      const [vhostRes, locationsRes] = await Promise.all([
        virtualHostsAPI.get(serverName),
        locationsAPI.getAll(serverName),
      ])
      setVhost(vhostRes.data)
      setLocations(locationsRes.data)
    } catch (error) {
      console.error('Failed to load locations:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (index) => {
    if (!confirm('Are you sure you want to delete this location?')) {
      return
    }

    try {
      await locationsAPI.delete(serverName, index)
      loadData()
    } catch (error) {
      alert('Failed to delete location: ' + error.message)
    }
  }

  if (loading) {
    return <div className="loading">Loading...</div>
  }

  return (
    <div className="locations">
      <div className="page-header">
        <div>
          <button className="btn" onClick={() => navigate('/virtualhosts')}>
            ← Back
          </button>
          <h1>Locations for {serverName}</h1>
        </div>
        <button className="btn btn-primary" onClick={() => setShowForm(true)}>
          Add Location
        </button>
      </div>

      {showForm && (
        <LocationForm
          onClose={() => {
            setShowForm(false)
            setEditing(null)
          }}
          onSave={() => {
            loadData()
            setShowForm(false)
            setEditing(null)
          }}
          editing={editing}
          serverName={serverName}
        />
      )}

      <div className="card">
        <table className="table">
          <thead>
            <tr>
              <th>Path</th>
              <th>Match Type</th>
              <th>Handler</th>
              <th>Proxy Type</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {locations.length === 0 ? (
              <tr>
                <td colSpan="5" style={{ textAlign: 'center', padding: '20px' }}>
                  No locations configured
                </td>
              </tr>
            ) : (
              locations.map((location, index) => (
                <tr key={index}>
                  <td>{location.path}</td>
                  <td>{location.matchType || 'prefix'}</td>
                  <td>{location.handler}</td>
                  <td>{location.proxyType || '-'}</td>
                  <td>
                    <button
                      className="btn btn-sm"
                      onClick={() => {
                        setEditing({ ...location, index })
                        setShowForm(true)
                      }}
                    >
                      Edit
                    </button>
                    <button
                      className="btn btn-sm btn-danger"
                      onClick={() => handleDelete(index)}
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

function LocationForm({ onClose, onSave, editing, serverName }) {
  const [formData, setFormData] = useState(
    editing
      ? { ...editing }
      : {
          path: '',
          matchType: 'prefix',
          handler: 'static',
          proxyUnixSocket: '',
          proxyType: 'http',
          cgiPath: '',
          phpProxyFCGI: '',
          directoryIndex: '',
        }
  )

  const handleSubmit = async (e) => {
    e.preventDefault()
    try {
      const locationData = { ...formData }
      delete locationData.index // Remove index if present

      if (editing && editing.index !== undefined) {
        await locationsAPI.update(serverName, editing.index, locationData)
      } else {
        await locationsAPI.create(serverName, locationData)
      }
      onSave()
    } catch (error) {
      alert('Failed to save location: ' + (error.response?.data?.error || error.message))
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>{editing ? 'Edit' : 'Add'} Location</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Path *</label>
            <input
              type="text"
              value={formData.path}
              onChange={(e) => setFormData({ ...formData, path: e.target.value })}
              required
              placeholder="/api or \\.php$ for regex"
            />
          </div>

          <div className="form-group">
            <label>Match Type</label>
            <select
              value={formData.matchType || 'prefix'}
              onChange={(e) => setFormData({ ...formData, matchType: e.target.value })}
            >
              <option value="prefix">Prefix</option>
              <option value="regex">Regex</option>
              <option value="regexCaseInsensitive">Regex (Case Insensitive)</option>
            </select>
          </div>

          <div className="form-group">
            <label>Handler *</label>
            <select
              value={formData.handler}
              onChange={(e) => setFormData({ ...formData, handler: e.target.value })}
              required
            >
              <option value="static">Static</option>
              <option value="proxy">Proxy</option>
              <option value="cgi">CGI</option>
              <option value="php">PHP</option>
            </select>
          </div>

          {formData.handler === 'proxy' && (
            <>
              <div className="form-group">
                <label>Proxy Unix Socket</label>
                <input
                  type="text"
                  value={formData.proxyUnixSocket || ''}
                  onChange={(e) => setFormData({ ...formData, proxyUnixSocket: e.target.value })}
                  placeholder="/var/run/app.sock"
                />
              </div>
              <div className="form-group">
                <label>Proxy Type</label>
                <select
                  value={formData.proxyType || 'http'}
                  onChange={(e) => setFormData({ ...formData, proxyType: e.target.value })}
                >
                  <option value="http">HTTP</option>
                  <option value="fcgi">FastCGI</option>
                </select>
              </div>
            </>
          )}

          {formData.handler === 'cgi' && (
            <div className="form-group">
              <label>CGI Path</label>
              <input
                type="text"
                value={formData.cgiPath || ''}
                onChange={(e) => setFormData({ ...formData, cgiPath: e.target.value })}
                placeholder="/cgi-bin"
              />
            </div>
          )}

          {formData.handler === 'php' && (
            <div className="form-group">
              <label>PHP FastCGI Address (TCP)</label>
              <input
                type="text"
                value={formData.phpProxyFCGI || ''}
                onChange={(e) => setFormData({ ...formData, phpProxyFCGI: e.target.value })}
                placeholder="127.0.0.1:9000"
              />
            </div>
          )}

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

export default Locations
