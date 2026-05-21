import React, { useState, useEffect, createContext, useContext, ReactNode } from 'react';
import api from '../lib/api';
import { User } from '../types/api';

// FIX-007: 类型守卫函数，验证 User 必要字段（后端只返回 id, username, role）
function isPartialUser(obj: unknown): obj is { id: number; username: string; role: string } {
  if (!obj || typeof obj !== 'object') return false;
  const user = obj as Record<string, unknown>;
  return (
    typeof user.id === 'number' &&
    typeof user.username === 'string' &&
    typeof user.role === 'string'
  );
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
  isAdmin: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(() => {
    // 从 token 解析 user 信息（页面刷新时恢复）
    const token = api.getToken();
    if (token) {
      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        return {
          id: payload.user_id || 0,
          username: payload.username || 'unknown',
          role: payload.role || 'operator',
          email: `${payload.username || 'unknown'}@example.com`,
          created_at: new Date().toISOString(),
        };
      } catch {
        return null;
      }
    }
    return null;
  });
  const [token, setToken] = useState<string | null>(api.getToken());

  useEffect(() => {
    const handleLogout = () => {
      setUser(null);
      setToken(null);
    };
    window.addEventListener('auth:logout', handleLogout);
    return () => window.removeEventListener('auth:logout', handleLogout);
  }, []);

  const login = async (username: string, password: string) => {
    const response = await api.login(username, password);
    setToken(response.token);
    // FIX-007: 使用类型守卫，补充后端缺失的 email/created_at 字段
    if (isPartialUser(response.user)) {
      setUser({
        id: response.user.id,
        username: response.user.username,
        role: response.user.role,
        email: `${response.user.username}@example.com`, // 默认邮箱
        created_at: new Date().toISOString(), // 默认创建时间
      });
    } else {
      console.error('Invalid user response:', response.user);
      setUser(null);
    }
  };

  const logout = () => {
    api.setToken(null);
    setUser(null);
    setToken(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        login,
        logout,
        isAuthenticated: !!token,
        isAdmin: user?.role === 'admin',
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export default AuthProvider;