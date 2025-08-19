import React, { useState } from 'react';
import { useAuthStore } from '../store/authStore';
import styles from './UserInfo.module.css';

const UserInfo: React.FC = () => {
  const { user, logout, register } = useAuthStore();
  const [showUpgrade, setShowUpgrade] = useState(false);
  const [upgradeForm, setUpgradeForm] = useState({
    email: '',
    password: '',
    confirmPassword: '',
  });

  if (!user) return null;

  const handleLogout = async () => {
    if (window.confirm('Are you sure you want to logout?')) {
      await logout();
    }
  };

  const handleUpgradeAccount = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (upgradeForm.password !== upgradeForm.confirmPassword) {
      alert('Passwords do not match');
      return;
    }

    try {
      await register(user.name, upgradeForm.email, upgradeForm.password);
      setShowUpgrade(false);
      setUpgradeForm({ email: '', password: '', confirmPassword: '' });
    } catch (error) {
      console.error('Failed to upgrade account:', error);
    }
  };

  const handleUpgradeInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUpgradeForm({
      ...upgradeForm,
      [e.target.name]: e.target.value,
    });
  };

  const getAuthTypeLabel = (authType: string) => {
    switch (authType) {
      case 'google':
        return '🚀 Google';
      case 'password':
        return '🔐 Account';
      case 'guest':
        return '👤 Guest';
      default:
        return '👤 Player';
    }
  };

  const getAuthTypeClass = (authType: string) => {
    switch (authType) {
      case 'google':
        return styles.google;
      case 'password':
        return styles.account;
      case 'guest':
        return styles.guest;
      default:
        return styles.default;
    }
  };

  if (showUpgrade && user.auth_type === 'guest') {
    return (
      <div className={styles.userInfo}>
        <form onSubmit={handleUpgradeAccount} className={styles.upgradeForm}>
          <div className={styles.upgradeHeader}>
            <h4>Upgrade to Full Account</h4>
            <p>Keep your progress and name: <strong>{user.name}</strong></p>
          </div>
          
          <div className={styles.formGroup}>
            <input
              name="email"
              type="email"
              value={upgradeForm.email}
              onChange={handleUpgradeInputChange}
              placeholder="Your email"
              required
            />
          </div>
          
          <div className={styles.formGroup}>
            <input
              name="password"
              type="password"
              value={upgradeForm.password}
              onChange={handleUpgradeInputChange}
              placeholder="Choose password"
              minLength={6}
              required
            />
          </div>
          
          <div className={styles.formGroup}>
            <input
              name="confirmPassword"
              type="password"
              value={upgradeForm.confirmPassword}
              onChange={handleUpgradeInputChange}
              placeholder="Confirm password"
              minLength={6}
              required
            />
          </div>
          
          <div className={styles.upgradeActions}>
            <button type="submit" className={styles.upgradeBtn}>
              Create Account
            </button>
            <button
              type="button"
              onClick={() => setShowUpgrade(false)}
              className={styles.cancelBtn}
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    );
  }

  return (
    <div className={styles.userInfo}>
      <div className={styles.userDetails}>
        <div className={styles.userName}>{user.name}</div>
        <div className={`${styles.authType} ${getAuthTypeClass(user.auth_type)}`}>
          {getAuthTypeLabel(user.auth_type)}
        </div>
        {user.email && (
          <div className={styles.userEmail}>{user.email}</div>
        )}
      </div>
      
      <div className={styles.userActions}>
        {user.auth_type === 'guest' && (
          <button
            onClick={() => setShowUpgrade(true)}
            className={styles.upgradeBtn}
            title="Create a permanent account"
          >
            ⬆️ Sign Up
          </button>
        )}
        <button
          onClick={handleLogout}
          className={styles.logoutBtn}
          title="Logout"
        >
          ↗️ Logout
        </button>
      </div>
    </div>
  );
};

export default UserInfo;
