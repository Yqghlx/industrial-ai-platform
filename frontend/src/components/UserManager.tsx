import React, { useState, useEffect } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { useAuth } from './AuthContext';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { Plus, Trash2, Shield, User as UserIcon } from 'lucide-react';
import { User, UserRole } from '../types/api';
import { asUserArraySafe } from '../types/typeGuards';

export default function UserManager() {
  const { t } = useI18n();
  const { isAdmin } = useAuth();
  const { showToast } = useToast();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    if (!isAdmin) return;
    loadUsers();
  }, [isAdmin]);

  const loadUsers = async () => {
    setLoading(true);
    try {
      const res = await api.getUsers(1, 50);
      // FE-P1-01: 使用类型守卫安全转换数组
      setUsers(asUserArraySafe(res.data));
    } catch (error) {
      console.error('Failed to load users:', error);
      showToast({ type: 'error', message: t('errors.loadFailedUsers') });
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm(t('user.confirmDelete'))) return;
    try {
      await api.deleteUser(id);
      showToast({ type: 'success', message: t('user.deleteSuccess') });
      loadUsers();
    } catch (error) {
      showToast({ type: 'error', message: t('errors.unknown') });
    }
  };

  if (!isAdmin) {
    return (
      <div className="text-center py-12">
        <Shield className="w-12 h-12 text-red-400 mx-auto mb-4" />
        <h2 className="text-xl font-bold text-slate-100">{t('user.adminRequired')}</h2>
        <p className="text-slate-400">{t('user.noPermission')}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.users')}</h1>
          <p className="text-slate-400">{t('user.title')}</p>
        </div>
        <button 
          onClick={() => setShowCreateModal(true)}
          className="btn btn-primary flex items-center gap-2"
          aria-label={t('user.createUser')}
        >
          <Plus className="w-5 h-5" />
          <span>{t('user.createUser')}</span>
        </button>
      </div>

      {/* Users table */}
      <div className="card">
        <div className="card-body">
          {loading ? (
            <div className="space-y-4">
              {[1, 2, 3, 4].map(i => <Skeleton key={i} variant="card" />)}
            </div>
          ) : (
            <div className="table-container">
              <table className="table">
                <thead>
                  <tr>
                    <th>{t('auth.username')}</th>
                    <th>{t('auth.email')}</th>
                    <th>{t('user.role')}</th>
                    <th>{t('workOrder.createdAt')}</th>
                    <th>{t('common.delete')}</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((user) => (
                    <tr key={user.id}>
                      <td>
                        <div className="flex items-center gap-2">
                          <div className="w-8 h-8 rounded-full bg-slate-600 flex items-center justify-center">
                            <UserIcon className="w-4 h-4 text-slate-300" />
                          </div>
                          <span className="font-medium">{user.username}</span>
                        </div>
                      </td>
                      <td>{user.email}</td>
                      <td>
                        <span className={`status-badge ${
                          user.role === 'admin' ? 'bg-primary-500/20 text-primary-400' :
                          'bg-slate-500/20 text-slate-300'
                        }`}>
                          {user.role === 'admin' ? t('user.admin') : t('user.user')}
                        </span>
                      </td>
                      <td>{new Date(user.created_at).toLocaleDateString()}</td>
                      <td>
                        <button
                          onClick={() => handleDelete(user.id)}
                          className="p-1 text-slate-400 hover:text-red-400"
                          aria-label={t('common.delete')}
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center">
          <div className="card max-w-md">
            <div className="card-header">
              <h2 className="text-lg font-semibold">{t('user.createUser')}</h2>
            </div>
            <div className="card-body">
              <form onSubmit={async (e) => {
                e.preventDefault();
                const formData = new FormData(e.target as HTMLFormElement);
                const data = {
                  username: formData.get('username') as string,
                  password: formData.get('password') as string,
                  email: formData.get('email') as string,
                  role: formData.get('role') as UserRole,
                };

                try {
                  await api.createUser(data);
                  showToast({ type: 'success', message: '用户已创建' });
                  setShowCreateModal(false);
                  loadUsers();
                } catch (error) {
                  showToast({ type: 'error', message: '创建失败' });
                }
              }}>
                <div className="space-y-4">
                  <div>
                    <label className="label">{t('auth.username')}</label>
                    <input name="username" className="input" required minLength={3} />
                  </div>
                  <div>
                    <label className="label">{t('auth.password')}</label>
                    <input name="password" type="password" className="input" required minLength={6} />
                  </div>
                  <div>
                    <label className="label">{t('auth.email')}</label>
                    <input name="email" type="email" className="input" required />
                  </div>
                  <div>
                    <label className="label">{t('user.role')}</label>
                    <select name="role" className="input">
                      <option value="user">{t('user.user')}</option>
                      <option value="admin">{t('user.admin')}</option>
                    </select>
                  </div>
                  <div className="flex gap-2">
                    <button type="submit" className="btn btn-primary flex-1">
                      {t('common.create')}
                    </button>
                    <button 
                      type="button"
                      onClick={() => setShowCreateModal(false)}
                      className="btn btn-secondary flex-1"
                      aria-label={t('common.cancel')}
                    >
                      {t('common.cancel')}
                    </button>
                  </div>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
      
    </div>
  );
}