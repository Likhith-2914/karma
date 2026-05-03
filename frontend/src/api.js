import axios from 'axios';

// Automatically route to '/api' in production (Nginx proxy), but use localhost in local dev
const API_URL = import.meta.env.MODE === 'production' ? '/api' : 'http://localhost:8080/api';

// Create an Axios instance
const api = axios.create({
  baseURL: API_URL,
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
