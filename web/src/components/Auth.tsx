import React, { useState, useEffect } from 'react';
import { useAuthStore } from '../store/authStore';
import styles from './Auth.module.css';

const Auth: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'login' | 'register' | 'guest'>('guest');
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    password: '',
    confirmPassword: '',
  });

  const {
    isLoading,
    error,
    ssoEnabled,
    authMethods,
    login,
    register,
    loginAsGuest,
    loginWithGoogle,
    checkAuthStatus,
    clearError,
  } = useAuthStore();

  useEffect(() => {
    checkAuthStatus();
    
    // Try to restore guest name from localStorage if available
    const savedGuestName = localStorage.getItem('dixitme-guest-name');
    if (savedGuestName && !formData.name) {
      setFormData(prev => ({ ...prev, name: savedGuestName }));
    }
  }, [checkAuthStatus]);

  useEffect(() => {
    if (error) {
      const timer = setTimeout(() => clearError(), 5000);
      return () => clearTimeout(timer);
    }
  }, [error, clearError]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.email.trim() || !formData.password.trim()) {
      return;
    }
    try {
      await login(formData.email.trim(), formData.password.trim());
    } catch (error) {
      // Error is handled by store
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim() || !formData.email.trim() || !formData.password.trim()) {
      return;
    }
    if (formData.password !== formData.confirmPassword) {
      useAuthStore.getState().setError('Passwords do not match');
      return;
    }
    try {
      await register(formData.name.trim(), formData.email.trim(), formData.password.trim());
    } catch (error) {
      // Error is handled by store
    }
  };

  const handleGuestLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim()) {
      return;
    }
    try {
      // Save guest name to localStorage for future visits
      localStorage.setItem('dixitme-guest-name', formData.name.trim());
      await loginAsGuest(formData.name.trim());
    } catch (error) {
      // Error is handled by store
    }
  };

  const handleGoogleLogin = async () => {
    if (!ssoEnabled) {
      useAuthStore.getState().setError('Google Sign-In is temporarily disabled');
      return;
    }
    
    // For demo purposes, we'll show a placeholder
    // In production, you'd integrate with Google OAuth
    useAuthStore.getState().setError('Google Sign-In integration coming soon!');
  };

  const resetForm = () => {
    setFormData({
      name: '',
      email: '',
      password: '',
      confirmPassword: '',
    });
  };

  const handleTabChange = (tab: 'login' | 'register' | 'guest') => {
    setActiveTab(tab);
    resetForm();
    clearError();
  };

  return (
    <div className={styles.authContainer}>
      <div className={styles.authCard}>
        <div className={styles.authHeader}>
          <h1>DixitMe</h1>
          <p>Online Dixit Card Game</p>
        </div>

        <div className={styles.tabs}>
          <button
            className={`${styles.tab} ${activeTab === 'guest' ? styles.active : ''}`}
            onClick={() => handleTabChange('guest')}
          >
            Quick Play
          </button>
          {authMethods.password && (
            <button
              className={`${styles.tab} ${activeTab === 'login' ? styles.active : ''}`}
              onClick={() => handleTabChange('login')}
            >
              Login
            </button>
          )}
          {authMethods.password && (
            <button
              className={`${styles.tab} ${activeTab === 'register' ? styles.active : ''}`}
              onClick={() => handleTabChange('register')}
            >
              Register
            </button>
          )}
        </div>

        {activeTab === 'guest' && (
          <form onSubmit={handleGuestLogin} className={styles.authForm}>
            <div className={styles.formGroup}>
              <label htmlFor="guest-name">Your Name</label>
              <input
                id="guest-name"
                name="name"
                type="text"
                value={formData.name}
                onChange={handleInputChange}
                placeholder="Enter your name"
                maxLength={20}
                required
              />
            </div>
            <button
              type="submit"
              className={styles.submitBtn}
              disabled={isLoading || !formData.name.trim()}
            >
              {isLoading ? 'Joining...' : 'Play as Guest'}
            </button>
          </form>
        )}

        {activeTab === 'login' && (
          <form onSubmit={handleLogin} className={styles.authForm}>
            <div className={styles.formGroup}>
              <label htmlFor="login-email">Email</label>
              <input
                id="login-email"
                name="email"
                type="email"
                value={formData.email}
                onChange={handleInputChange}
                placeholder="Enter your email"
                required
              />
            </div>
            <div className={styles.formGroup}>
              <label htmlFor="login-password">Password</label>
              <input
                id="login-password"
                name="password"
                type="password"
                value={formData.password}
                onChange={handleInputChange}
                placeholder="Enter your password"
                required
              />
            </div>
            <button
              type="submit"
              className={styles.submitBtn}
              disabled={isLoading || !formData.email.trim() || !formData.password.trim()}
            >
              {isLoading ? 'Logging in...' : 'Login'}
            </button>
          </form>
        )}

        {activeTab === 'register' && (
          <form onSubmit={handleRegister} className={styles.authForm}>
            <div className={styles.formGroup}>
              <label htmlFor="register-name">Name</label>
              <input
                id="register-name"
                name="name"
                type="text"
                value={formData.name}
                onChange={handleInputChange}
                placeholder="Enter your name"
                maxLength={50}
                required
              />
            </div>
            <div className={styles.formGroup}>
              <label htmlFor="register-email">Email</label>
              <input
                id="register-email"
                name="email"
                type="email"
                value={formData.email}
                onChange={handleInputChange}
                placeholder="Enter your email"
                required
              />
            </div>
            <div className={styles.formGroup}>
              <label htmlFor="register-password">Password</label>
              <input
                id="register-password"
                name="password"
                type="password"
                value={formData.password}
                onChange={handleInputChange}
                placeholder="Enter your password"
                minLength={6}
                required
              />
            </div>
            <div className={styles.formGroup}>
              <label htmlFor="register-confirm-password">Confirm Password</label>
              <input
                id="register-confirm-password"
                name="confirmPassword"
                type="password"
                value={formData.confirmPassword}
                onChange={handleInputChange}
                placeholder="Confirm your password"
                minLength={6}
                required
              />
            </div>
            <button
              type="submit"
              className={styles.submitBtn}
              disabled={
                isLoading ||
                !formData.name.trim() ||
                !formData.email.trim() ||
                !formData.password.trim() ||
                !formData.confirmPassword.trim()
              }
            >
              {isLoading ? 'Creating account...' : 'Create Account'}
            </button>
          </form>
        )}

        {authMethods.google && (
          <div className={styles.socialLogin}>
            <div className={styles.divider}>
              <span>or</span>
            </div>
            <button
              onClick={handleGoogleLogin}
              className={`${styles.googleBtn} ${!ssoEnabled ? styles.disabled : ''}`}
              disabled={!ssoEnabled || isLoading}
            >
              <span className={styles.googleIcon}>🚀</span>
              {ssoEnabled ? 'Continue with Google' : 'Google Sign-In Disabled'}
            </button>
          </div>
        )}

        {error && (
          <div className={styles.errorMessage}>
            {error}
          </div>
        )}

        <div className={styles.authFooter}>
          <p>
            {activeTab === 'guest' && 'Guest sessions remember your name but progress isn\'t saved permanently. You can upgrade to a full account later!'}
            {activeTab === 'login' && 'Don\'t have an account? Click Register above.'}
            {activeTab === 'register' && 'Already have an account? Click Login above.'}
          </p>
        </div>
      </div>
    </div>
  );
};

export default Auth;
