import client from './client'

export const configAPI = {
  get: () => client.get('/config'),
  update: (config) => client.put('/config', config),
  reload: () => client.post('/config/reload'),
}

export const virtualHostsAPI = {
  getAll: () => client.get('/virtualhosts'),
  get: (serverName) => client.get(`/virtualhosts/${serverName}`),
  create: (vhost) => client.post('/virtualhosts', vhost),
  update: (serverName, vhost) => client.put(`/virtualhosts/${serverName}`, vhost),
  delete: (serverName) => client.delete(`/virtualhosts/${serverName}`),
}

export const locationsAPI = {
  getAll: (serverName) => client.get(`/virtualhosts/${serverName}/locations`),
  create: (serverName, location) => client.post(`/virtualhosts/${serverName}/locations`, location),
  update: (serverName, index, location) => client.put(`/virtualhosts/${serverName}/locations/${index}`, location),
  delete: (serverName, index) => client.delete(`/virtualhosts/${serverName}/locations/${index}`),
}

export const serverAPI = {
  getStatus: () => client.get('/server/status'),
  reload: () => client.post('/server/reload'),
  restart: () => client.post('/server/restart'),
}

export const statsAPI = {
  get: () => client.get('/stats'),
}
