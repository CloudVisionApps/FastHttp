import React from 'react'
import { BrowserRouter as Router, Routes, Route, NavLink } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import VirtualHosts from './pages/VirtualHosts'
import Config from './pages/Config'
import Locations from './pages/Locations'

function App() {
  return (
    <Router>
      <div className="flex min-h-screen bg-zinc-50">
        <nav className="w-64 bg-gradient-to-b from-zinc-800 to-zinc-900 text-white shadow-lg">
          <div className="p-6 border-b border-white/10">
            <h2 className="text-2xl font-bold mb-1">FastHTTP</h2>
            <p className="text-sm text-white/80">Admin Panel</p>
          </div>
          <ul className="p-4 space-y-2">
            <li>
              <NavLink
                to="/"
                className={({ isActive }) =>
                  `block px-4 py-3 rounded-xl transition-colors ${
                    isActive
                      ? 'bg-white/20 text-white font-medium'
                      : 'text-white/90 hover:bg-white/10'
                  }`
                }
              >
                Dashboard
              </NavLink>
            </li>
            <li>
              <NavLink
                to="/virtualhosts"
                className={({ isActive }) =>
                  `block px-4 py-3 rounded-xl transition-colors ${
                    isActive
                      ? 'bg-white/20 text-white font-medium'
                      : 'text-white/90 hover:bg-white/10'
                  }`
                }
              >
                Virtual Hosts
              </NavLink>
            </li>
            <li>
              <NavLink
                to="/config"
                className={({ isActive }) =>
                  `block px-4 py-3 rounded-xl transition-colors ${
                    isActive
                      ? 'bg-white/20 text-white font-medium'
                      : 'text-white/90 hover:bg-white/10'
                  }`
                }
              >
                Configuration
              </NavLink>
            </li>
          </ul>
        </nav>
        <main className="flex-1 p-8 overflow-y-auto">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/virtualhosts" element={<VirtualHosts />} />
            <Route path="/config" element={<Config />} />
            <Route path="/locations/:serverName" element={<Locations />} />
          </Routes>
        </main>
      </div>
    </Router>
  )
}

export default App
