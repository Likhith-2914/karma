import axios from 'axios';

// Use VITE_API_URL if it exists (for Vercel), otherwise fallback to localhost (for local dev)
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

// Create an Axios instance with a 15-second timeout
const api = axios.create({
  baseURL: API_URL,
  timeout: 15000,
});

// Add a request interceptor
// This function runs before every request we make to the Go backend.
// It checks if we have a token in localStorage and attaches it to the Authorization header.
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
}, (error) => {
  return Promise.reject(error);
});

export default api;
