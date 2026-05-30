import React, { useState, useCallback } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { useEscapeKey } from '../lib/hooks';
import { useAuth } from './AuthContext';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { Plus, Trash2, Shield, User as UserIcon } from 'lucide-react';
import { User, UserRole, UserCreateInput } from '../types/api';
import { useConfirmDialog } from './UI/ConfirmDialog';
import { useCRUD } from '../hooks/useCRUD';

// FE-P2: React.memo 优化用户行渲染
interface UserRowProps {
  user: User;
  t: (key: string) => string;
  onDelete: (id: number) => void;
}

const UserRow = React.memo(function UserRow({ user, t, onDelete }: UserRowProps) {
  return (
    <tr>
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
          onClick={() => onDelete(user.id)}
          className="p-1 text-slate-400 hover:text-red-400 min-w-[44px] min-h-[44px] flex items-center justify-center"
          aria-label={t('common.delete')}
        >
          <Trash2 className="w-4 h-4" />
        </button>
      </td>
    </tr>
  );
});

export default function UserManager() {
  const { t } = useI18n();
  const { isAdmin } = useAuth();
  const { showToast } = useToast();
  const { showConfirm } = useConfirmDialog();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [creating, setCreating] = useState(false);

  // C1: Escape 键关闭模态框
  const closeCreateModal = useCallback(() => {
    setShowCreateModal(false);
  }, []);

  useEscapeKey(closeCreateModal, showCreateModal);

  // FE-P2-09: 使用通用 useCRUD hook 替代重复的 CRUD 逻辑
  const [state, actions] = useCRUD<User>({
    apiGetAll: (page, pageSize) => api.getUsers(page, pageSize),
    apiGetOne: async (id) => {
      // 注: User API 没有 getOne，这里通过 getAll 获取并查找
      const res = await api.getUsers(1, 100);
      return res.data?.find(u => String(u.id) === String(id)) || null as unknown as User;
    },
    apiCreate: (data) => api.createUser(data as UserCreateInput),
    apiUpdate: (_id, _data) => Promise.resolve(null as unknown as User), // User API 没有 update
    apiDelete: (id) => api.deleteUser(Number(id)),
    entityName: 'User',
    initialPageSize: 50,
    onError: (error) => showToast({ type: 'error', message: error }),
    onSuccess: (_action) => {},
  });

  const { items: users, loading } = state;
  const { create, delete: deleteItem } = actions;

  const handleDelete = useCallback(async (id: number) => {
    // FE-P2-11: 使用自定义确认框替代原生 confirm()
    const confirmed = await showConfirm({
      title: t('user.confirmDeleteTitle'),
      message: t('user.confirmDelete'),
      variant: 'danger',
      confirmText: t('common.delete'),
      cancelText: t('common.cancel'),
    });
    if (!confirmed) return;
    const success = await deleteItem(String(id));
    if (success) {
      showToast({ type: 'success', message: t('user.deleteSuccess') });
    } else {
      showToast({ type: 'error', message: t('errors.unknown') });
    }
  }, [showConfirm, deleteItem, showToast, t]);

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
                    <UserRow
                      key={user.id}
                      user={user}
                      t={t}
                      onDelete={handleDelete}
                    />
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" role="dialog" aria-modal="true" onClick={(e) => { if (e.target === e.currentTarget) closeCreateModal(); }}>
          <div className="card max-w-md">
            <div className="card-header">
              <h2 className="text-lg font-semibold">{t('user.createUser')}</h2>
            </div>
            <div className="card-body">
              <form onSubmit={async (e) => {
                e.preventDefault();
                setCreating(true);
                const formData = new FormData(e.target as HTMLFormElement);
                const data = {
                  username: formData.get('username') as string,
                  password: formData.get('password') as string,
                  email: formData.get('email') as string,
                  role: formData.get('role') as UserRole,
                };

                // FE-P2-09: 使用 useCRUD hook 的 create 方法
                const success = await create(data) !== null;
                if (success) {
                  showToast({ type: 'success', message: t('user.createSuccess') });
                } else {
                  showToast({ type: 'error', message: t('user.createFailed') });
                }
                setCreating(false);
                setShowCreateModal(false);
              }}>
                <div className="space-y-4">
                  <div>
                    <label className="label" htmlFor="user-create-username">{t('auth.username')}</label>
                    <input id="user-create-username" name="username" className="input" required minLength={3} />
                  </div>
                  <div>
                    <label className="label" htmlFor="user-create-password">{t('auth.password')}</label>
                    <input id="user-create-password" name="password" type="password" className="input" required minLength={6} />
                  </div>
                  <div>
                    <label className="label" htmlFor="user-create-email">{t('auth.email')}</label>
                    <input id="user-create-email" name="email" type="email" className="input" required />
                  </div>
                  <div>
                    <label className="label" htmlFor="user-create-role">{t('user.role')}</label>
                    <select id="user-create-role" name="role" className="input">
                      <option value="user">{t('user.user')}</option>
                      <option value="admin">{t('user.admin')}</option>
                    </select>
                  </div>
                  <div className="flex gap-2">
                    <button type="submit" className="btn btn-primary flex-1" disabled={creating}>
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