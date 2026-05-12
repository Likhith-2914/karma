import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../api';
import { Spinner } from './Loader';

const Auth = ({ setAuth }) => {
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      if (isLogin) {
        // Log in
        const res = await api.post('/login', { email, password });
        localStorage.setItem('token', res.data.token);
        const userTheme = res.data.user.theme || 'dark';
        localStorage.setItem('theme', userTheme);
        document.documentElement.setAttribute('data-theme', userTheme);
        setAuth(true);
        navigate('/dashboard');
      } else {
        // Sign up
        await api.post('/signup', { email, password, name });
        // After signup, switch to login
        setIsLogin(true);
        setError('Signup successful! Please log in.');
      }
    } catch (err) {
      setError(err.response?.data || 'An error occurred');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <img src="/logo.png" alt="Karma Logo" className="app-logo-large" style={{ width: '64px', height: '64px', objectFit: 'contain', marginBottom: '1rem' }} />
        <h1 style={{ marginTop: 0 }}>Karma</h1>
        <p className="subtitle">{isLogin ? 'Welcome back to your planner' : 'Start organizing your day'}</p>
        
        {error && <div className="error-message">{error}</div>}
        
        <form onSubmit={handleSubmit} className="auth-form">
          {!isLogin && (
            <div className="form-group">
              <label>Name</label>
              <input type="text" value={name} onChange={(e) => setName(e.target.value)} required />
            </div>
          )}
          <div className="form-group">
            <label>Email</label>
            <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
          </div>
          <div className="form-group">
            <label>Password</label>
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
          </div>
          <button type="submit" className="primary-btn flex-center" style={{ justifyContent: 'center' }} disabled={isLoading}>
            {isLoading ? <Spinner size={18} color="white" /> : (isLogin ? 'Log In' : 'Sign Up')}
          </button>
        </form>
        
        <p className="toggle-auth">
          {isLogin ? "Don't have an account?" : "Already have an account?"}
          <button onClick={() => setIsLogin(!isLogin)} className="text-btn">
            {isLogin ? 'Sign up' : 'Log in'}
          </button>
        </p>
      </div>
    </div>
  );
};

export default Auth;
