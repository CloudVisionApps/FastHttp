import axios from 'axios'

// Auto-detect API URL based on current hostname
const getApiUrl = () => {
  // If VITE_API_URL is set, use it
  if (import.meta.env.VITE_API_URL) {
    return import.meta.env.VITE_API_URL
  }
  
  // If served from the same origin (admin panel built and served by the API server),
  // always use relative URL to avoid CORS issues
  // This works when the React app is served from the same server on port 8080
  return '/api/v1'
}

const API_BASE_URL = getApiUrl()

// Log for debugging
console.log('API Base URL:', API_BASE_URL)

const client = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor for auth (if needed)
client.interceptors.request.use(
  (config) => {
    // Add auth token if available
    const token = localStorage.getItem('authToken')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for error handling
client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized
      localStorage.removeItem('authToken')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default client
