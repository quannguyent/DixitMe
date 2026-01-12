import React from 'react';
import { useAuthStore } from '../features/auth/stores/authStore';
import styles from './MainLayout.module.css';

interface MainLayoutProps {
  children: React.ReactNode;
  headerActions?: React.ReactNode;
}

export const MainLayout: React.FC<MainLayoutProps> = ({ children, headerActions }) => {
  const { user, logout } = useAuthStore();

  const handleLogout = async () => {
    if (window.confirm('Are you sure you want to logout?')) {
      await logout();
    }
  };

  const getInitials = (name: string) => {
    return name.slice(0, 2).toUpperCase();
  };

  return (
    <div className={styles.container}>
      <div className={styles.starsBackground}>
        <div className={styles.clouds} />
      </div>

      <header className={styles.header}>
        <div className={styles.logo}>
          <span className={styles.logoEmoji}>✨</span>
          DixitMe
        </div>

        <div className={styles.userProfile}>
          {user ? (
            <>
              <div className={styles.userInfo}>
                <span className={styles.userName}>{user.name}</span>
                <span className={styles.userRole}>
                  {user.auth_type === 'guest' ? 'Guest' : 'Member'}
                </span>
              </div>
              <div className={styles.avatar} title={user.name}>
                {getInitials(user.name)}
              </div>
              <button
                onClick={handleLogout}
                className="btn btn-ghost"
                style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem' }}
              >
                Logout
              </button>
            </>
          ) : (
            <div className={styles.authButtons}>
              {headerActions}
            </div>
          )}
        </div>
      </header>

      <main className={styles.main}>
        {children}
      </main>

      <footer className={styles.footer}>
        <p>© 2024 DixitMe • Built with ✨ for the Advanced Coding Agent</p>
      </footer>
    </div>
  );
};
