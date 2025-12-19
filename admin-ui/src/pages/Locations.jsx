import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { locationsAPI, virtualHostsAPI } from '../api/config'

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
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-zinc-500 text-lg">Loading...</div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-8">
        <div className="flex items-center gap-4">
          <button
            className="px-4 py-2 bg-zinc-200 text-zinc-700 rounded-xl hover:bg-zinc-300 transition-colors font-medium"
            onClick={() => navigate('/virtualhosts')}
          >
            ← Back
          </button>
          <h1 className="text-3xl font-bold text-zinc-900">Locations for {serverName}</h1>
        </div>
        <button
          className="px-4 py-2 bg-zinc-700 text-white rounded-xl hover:bg-zinc-800 transition-colors font-medium"
          onClick={() => setShowForm(true)}
        >
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

      <div className="bg-white rounded-2xl shadow-sm border border-zinc-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-zinc-200">
            <thead className="bg-zinc-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Path
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Match Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Handler
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Proxy Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-zinc-200">
              {locations.length === 0 ? (
                <tr>
                  <td colSpan="5" className="px-6 py-12 text-center text-zinc-500">
                    No locations configured
                  </td>
                </tr>
              ) : (
                locations.map((location, index) => (
                  <tr key={index} className="hover:bg-zinc-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-zinc-900">
                      {location.path}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-zinc-500">
                      {location.matchType || 'prefix'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-zinc-500">
                      {location.handler}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-zinc-500">
                      {location.proxyType || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                      <button
                        className="text-zinc-700 hover:text-zinc-900"
                        onClick={() => {
                          setEditing({ ...location, index })
                          setShowForm(true)
                        }}
                      >
                        Edit
                      </button>
                      <button
                        className="text-red-600 hover:text-red-900"
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
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <div
        className="bg-white rounded-2xl shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6 border-b border-zinc-200">
          <h2 className="text-2xl font-bold text-zinc-900">
            {editing ? 'Edit' : 'Add'} Location
          </h2>
        </div>
        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">Path *</label>
            <input
              type="text"
              value={formData.path}
              onChange={(e) => setFormData({ ...formData, path: e.target.value })}
              required
              placeholder="/api or \\.php$ for regex"
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">Match Type</label>
            <select
              value={formData.matchType || 'prefix'}
              onChange={(e) => setFormData({ ...formData, matchType: e.target.value })}
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            >
              <option value="prefix">Prefix</option>
              <option value="regex">Regex</option>
              <option value="regexCaseInsensitive">Regex (Case Insensitive)</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">Handler *</label>
            <select
              value={formData.handler}
              onChange={(e) => setFormData({ ...formData, handler: e.target.value })}
              required
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            >
              <option value="static">Static</option>
              <option value="proxy">Proxy</option>
              <option value="cgi">CGI</option>
              <option value="php">PHP</option>
            </select>
          </div>

          {formData.handler === 'proxy' && (
            <>
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-2">
                  Proxy Unix Socket
                </label>
                <input
                  type="text"
                  value={formData.proxyUnixSocket || ''}
                  onChange={(e) => setFormData({ ...formData, proxyUnixSocket: e.target.value })}
                  placeholder="/var/run/app.sock"
                  className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-2">Proxy Type</label>
                <select
                  value={formData.proxyType || 'http'}
                  onChange={(e) => setFormData({ ...formData, proxyType: e.target.value })}
                  className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
                >
                  <option value="http">HTTP</option>
                  <option value="fcgi">FastCGI</option>
                </select>
              </div>
            </>
          )}

          {formData.handler === 'cgi' && (
            <div>
              <label className="block text-sm font-medium text-zinc-700 mb-2">CGI Path</label>
              <input
                type="text"
                value={formData.cgiPath || ''}
                onChange={(e) => setFormData({ ...formData, cgiPath: e.target.value })}
                placeholder="/cgi-bin"
                className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
              />
            </div>
          )}

          {formData.handler === 'php' && (
            <div>
              <label className="block text-sm font-medium text-zinc-700 mb-2">
                PHP FastCGI Address (TCP)
              </label>
              <input
                type="text"
                value={formData.phpProxyFCGI || ''}
                onChange={(e) => setFormData({ ...formData, phpProxyFCGI: e.target.value })}
                placeholder="127.0.0.1:9000"
                className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
              />
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">Directory Index</label>
            <input
              type="text"
              value={formData.directoryIndex || ''}
              onChange={(e) => setFormData({ ...formData, directoryIndex: e.target.value })}
              placeholder="index.php index.html"
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>

          <div className="flex gap-3 pt-4 border-t border-zinc-200">
            <button
              type="submit"
              className="px-4 py-2 bg-zinc-700 text-white rounded-xl hover:bg-zinc-800 transition-colors font-medium"
            >
              {editing ? 'Update' : 'Create'}
            </button>
            <button
              type="button"
              className="px-4 py-2 bg-zinc-200 text-zinc-700 rounded-xl hover:bg-zinc-300 transition-colors font-medium"
              onClick={onClose}
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default Locations
