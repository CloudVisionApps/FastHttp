import React from 'react'
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import VirtualHosts from './pages/VirtualHosts'
import Config from './pages/Config'
import Locations from './pages/Locations'
import './App.css'

function App() {
  return (
    <Router>
      <div className="app">
        <nav className="sidebar">
          <div className="logo">
            <h2>FastHTTP</h2>
            <p>Admin Panel</p>
          </div>
          <ul className="nav-menu">
            <li>
              <Link to="/">Dashboard</Link>
            </li>
            <li>
              <Link to="/virtualhosts">Virtual Hosts</Link>
            </li>
            <li>
              <Link to="/config">Configuration</Link>
            </li>
            <li>
              <Link to="/locations">Locations</Link>
            </li>
          </ul>
        </nav>
        <main className="content">
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
