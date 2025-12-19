import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { virtualHostsAPI } from '../api/config'

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
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-zinc-500 text-lg">Loading...</div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold text-zinc-900">Virtual Hosts</h1>
        <button
          className="px-4 py-2 bg-zinc-700 text-white rounded-xl hover:bg-zinc-800 transition-colors font-medium"
          onClick={() => setShowForm(true)}
        >
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

      <div className="bg-white rounded-2xl shadow-sm border border-zinc-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-zinc-200">
            <thead className="bg-zinc-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Server Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Document Root
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Listen Ports
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Locations
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-zinc-200">
              {virtualHosts.length === 0 ? (
                <tr>
                  <td colSpan="5" className="px-6 py-12 text-center text-zinc-500">
                    No virtual hosts configured
                  </td>
                </tr>
              ) : (
                virtualHosts.map((vhost) => (
                  <tr key={vhost.serverName} className="hover:bg-zinc-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-zinc-900">
                      {vhost.serverName}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-zinc-500">
                      {vhost.documentRoot}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-zinc-500">
                      {vhost.listen?.length > 0 ? vhost.listen.join(', ') : 'All ports'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-zinc-500">
                      {vhost.locations?.length || 0}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                      <button
                        className="text-zinc-700 hover:text-zinc-900"
                        onClick={() => navigate(`/locations/${vhost.serverName}`)}
                      >
                        Locations
                      </button>
                      <button
                        className="text-zinc-600 hover:text-zinc-900"
                        onClick={() => {
                          setEditing(vhost)
                          setShowForm(true)
                        }}
                      >
                        Edit
                      </button>
                      <button
                        className="text-red-600 hover:text-red-900"
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
            {editing ? 'Edit' : 'Add'} Virtual Host
          </h2>
        </div>
        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">
              Server Name *
            </label>
            <input
              type="text"
              value={formData.serverName}
              onChange={(e) => setFormData({ ...formData, serverName: e.target.value })}
              required
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">
              Document Root *
            </label>
            <input
              type="text"
              value={formData.documentRoot}
              onChange={(e) => setFormData({ ...formData, documentRoot: e.target.value })}
              required
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">
              Listen Ports (comma-separated, empty for all)
            </label>
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
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">User</label>
            <input
              type="text"
              value={formData.user || ''}
              onChange={(e) => setFormData({ ...formData, user: e.target.value })}
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2">Group</label>
            <input
              type="text"
              value={formData.group || ''}
              onChange={(e) => setFormData({ ...formData, group: e.target.value })}
              className="w-full px-4 py-2 border border-zinc-300 rounded-xl focus:ring-2 focus:ring-zinc-500 focus:border-zinc-500 outline-none"
            />
          </div>

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

export default VirtualHosts
