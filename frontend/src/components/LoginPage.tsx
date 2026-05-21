import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useI18n } from '../i18n';
import { useAuth } from './AuthContext';
import api from '../lib/api';
import { useToast } from './Toast';
import { Activity, Lock, User, Mail, AlertCircle } from 'lucide-react';
import { parseApiError, ErrorType, getErrorType } from '../lib/errorHelper';

export default function LoginPage() {
  const { t } = useI18n();
  const { login } = useAuth();
  const navigate = useNavigate();
  const { showToast } = useToast();
  
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      if (mode === 'login') {
        await login(username, password);
        showToast({ type: 'success', message: t('auth.loginSuccess') });
        navigate('/');
      } else {
        await api.register({ username, password, email });
        await login(username, password);
        showToast({ type: 'success', message: t('auth.registerSuccess') });
        navigate('/');
      }
    } catch (err) {
      // 获取错误类型以提供更具体的提示
      const errorType = getErrorType(err);
      let errorMessage: string;
      
      switch (errorType) {
        case ErrorType.NETWORK:
          errorMessage = t('errors.networkError');
          break;
        case ErrorType.TIMEOUT:
          errorMessage = t('errors.timeout');
          break;
        case ErrorType.RATE_LIMIT:
          errorMessage = t('errors.rateLimit');
          showToast({ type: 'error', message: t('errors.rateLimit') });
          break;
        case ErrorType.UNAUTHORIZED:
          errorMessage = mode === 'login' ? t('auth.invalidCredentials') : t('errors.unauthorized');
          break;
        case ErrorType.VALIDATION:
          errorMessage = err instanceof Error ? err.message : t('errors.validation');
          break;
        default:
          // 使用智能错误解析
          errorMessage = parseApiError(err, t);
      }
      
      setError(errorMessage);
      
      // 只有非速率限制错误才显示 toast（速率限制已单独显示）
      if (errorType !== ErrorType.RATE_LIMIT) {
        showToast({ 
          type: 'error', 
          message: mode === 'login' ? t('auth.loginFailed') : t('auth.registerFailed') 
        });
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-xl bg-primary-600 mb-4">
            <Activity className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-2xl font-bold text-slate-100">Industrial AI Platform</h1>
          <p className="text-slate-400 mt-2">工业AI代理平台</p>
        </div>

        {/* Form */}
        <div className="card">
          <div className="card-body">
            {/* Tabs */}
            <div className="flex gap-2 mb-6">
              <button
                onClick={() => setMode('login')}
                className={`flex-1 py-2 rounded-lg text-center transition-colors ${
                  mode === 'login'
                    ? 'bg-primary-600 text-white'
                    : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                }`}
              >
                {t('auth.login')}
              </button>
              <button
                onClick={() => setMode('register')}
                className={`flex-1 py-2 rounded-lg text-center transition-colors ${
                  mode === 'register'
                    ? 'bg-primary-600 text-white'
                    : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                }`}
              >
                {t('auth.register')}
              </button>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              {/* Username */}
              <div>
                <label className="label">{t('auth.username')}</label>
                <div className="relative">
                  <User className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                  <input
                    type="text"
                    name="username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    className="input pl-10"
                    placeholder={t('auth.username')}
                    required
                  />
                </div>
              </div>

              {/* Password */}
              <div>
                <label className="label">{t('auth.password')}</label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                  <input
                    type="password"
                    name="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="input pl-10"
                    placeholder={t('auth.password')}
                    required
                  />
                </div>
              </div>

              {/* Email (register only) */}
              {mode === 'register' && (
                <div>
                  <label className="label">{t('auth.email')}</label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                    <input
                      type="email"
                      name="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      className="input pl-10"
                      placeholder={t('auth.email')}
                      required
                    />
                  </div>
                </div>
              )}

              {/* Error */}
              {error && (
                <div className="flex items-center gap-2 p-3 bg-red-500/20 rounded-lg text-red-400">
                  <AlertCircle className="w-5 h-5" />
                  <span className="text-sm">{error}</span>
                </div>
              )}

              {/* Submit */}
              <button
                type="submit"
                disabled={loading}
                className="btn btn-primary w-full flex items-center justify-center gap-2"
              >
                {loading ? (
                  <>
                    <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    <span>{t('common.loading')}</span>
                  </>
                ) : (
                  <span>{mode === 'login' ? t('auth.login') : t('auth.register')}</span>
                )}
              </button>
            </form>

            {/* Demo hint - only show in development mode */}
            {import.meta.env.DEV && (
              <div className="mt-6 pt-4 border-t border-slate-700 text-center text-sm text-slate-400">
                <p>演示账户: admin / Admin@123456</p>
              </div>
            )}
          </div>
        </div>
      </div>
      
    </div>
  );
}