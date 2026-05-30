import React, { useEffect, useCallback, useMemo } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useI18n } from '../i18n';
import { useAuth } from './AuthContext';
import { useIsMobile } from '../lib/useSwipe';
import {
  LayoutDashboard,
  Settings,
  Users,
  Bell,
  FileText,
  Bot,
  Wrench,
  Activity,
  Network,
  Box,
  BarChart3,
  Database,
  Globe,
  LogOut,
  X,
  ChevronRight,
  ChevronLeft,
  Gauge,
  AlertTriangle
} from 'lucide-react';

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
  collapsed?: boolean;
  onToggleCollapse?: () => void;
}

export default function Sidebar({ isOpen, onClose, collapsed = false, onToggleCollapse }: SidebarProps) {
  const { t, language, setLanguage } = useI18n();
  const { user, logout, isAdmin } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const isMobile = useIsMobile();

  // Close sidebar on escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [isOpen, onClose]);

  // Lock body scroll when sidebar is open on mobile
  useEffect(() => {
    if (isMobile && isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    
    return () => {
      document.body.style.overflow = '';
    };
  }, [isMobile, isOpen]);

  const handleLogout = useCallback(() => {
    logout();
    onClose();
    navigate('/login');
  }, [logout, onClose, navigate]);

  const handleNavClick = useCallback(() => {
    if (isMobile) {
      onClose();
    }
  }, [isMobile, onClose]);

  // FE-P1-03: 使用 useMemo 防止每次渲染重建菜单数组
  const menuItems = useMemo(() => [
    { path: '/dashboard', icon: LayoutDashboard, label: t('nav.dashboard') },
    { path: '/digital-twin', icon: Activity, label: t('nav.digitalTwin') },
    { path: '/devices', icon: Settings, label: t('nav.devices') },
    { path: '/knowledge-graph', icon: Network, label: t('nav.knowledgeGraph') },
    { path: '/rules', icon: Bell, label: t('nav.rules') },
    { path: '/work-orders', icon: Wrench, label: t('nav.workOrders') },
    { path: '/notifications', icon: Bell, label: t('nav.notifications') },
    { path: '/ai-agent', icon: Bot, label: t('nav.aiAgent') },
    { path: '/reports', icon: FileText, label: t('nav.reports') },
    { path: '/telemetry', icon: Gauge, label: t('nav.telemetry') },
    { path: '/alerts', icon: AlertTriangle, label: t('nav.alerts') },
    { path: '/blackbox', icon: Box, label: t('nav.blackbox') },
    { path: '/roi', icon: BarChart3, label: t('nav.roi') },
  ], [t]);

  const adminItems = useMemo(() => [
    { path: '/users', icon: Users, label: t('nav.users') },
    { path: '/system', icon: Database, label: t('nav.system') },
  ], [t]);

  const isActive = (path: string) => location.pathname === path;

  return (
    <>
      {/* Desktop sidebar */}
      <aside className={`hidden lg:flex fixed inset-y-0 left-0 z-50 bg-slate-800 border-r border-slate-700 flex-col transition-all duration-300 ${collapsed ? 'w-16' : 'w-64'}`}>
        <SidebarContent
          menuItems={menuItems}
          adminItems={adminItems}
          isAdmin={isAdmin}
          user={user}
          language={language}
          setLanguage={setLanguage}
          t={t}
          onLogout={handleLogout}
          isActive={isActive}
          onNavClick={handleNavClick}
          showCloseButton={false}
          collapsed={collapsed}
          onToggleCollapse={onToggleCollapse}
        />
      </aside>

      {/* Mobile drawer */}
      <div
        className={`
          lg:hidden fixed inset-y-0 left-0 z-50 w-72 max-w-[85vw] bg-slate-800 border-r border-slate-700
          transform transition-transform duration-300 ease-in-out
          ${isOpen ? 'translate-x-0' : '-translate-x-full'}
        `}
      >
        <SidebarContent 
          menuItems={menuItems}
          adminItems={adminItems}
          isAdmin={isAdmin}
          user={user}
          language={language}
          setLanguage={setLanguage}
          t={t}
          onLogout={handleLogout}
          isActive={isActive}
          onNavClick={handleNavClick}
          showCloseButton={true}
          onClose={onClose}
        />
      </div>
    </>
  );
}

interface SidebarContentProps {
  menuItems: Array<{ path: string; icon: React.ComponentType<{ className?: string }>; label: string }>;
  adminItems: Array<{ path: string; icon: React.ComponentType<{ className?: string }>; label: string }>;
  isAdmin: boolean;
  user: { username: string; role: string } | null;
  language: string;
  setLanguage: (lang: 'zh' | 'en') => void;
  t: (key: string) => string;
  onLogout: () => void;
  isActive: (path: string) => boolean;
  onNavClick: () => void;
  showCloseButton: boolean;
  onClose?: () => void;
  collapsed?: boolean;
  onToggleCollapse?: () => void;
}

function SidebarContent({
  menuItems,
  adminItems,
  isAdmin,
  user,
  language,
  setLanguage,
  t,
  onLogout,
  isActive,
  onNavClick,
  showCloseButton,
  onClose,
  collapsed = false,
  onToggleCollapse,
}: SidebarContentProps) {
  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between h-14 lg:h-16 px-4 border-b border-slate-700 shrink-0">
        <div className="flex items-center gap-2 min-w-0">
          <div className="w-8 h-8 rounded-lg bg-primary-600 flex items-center justify-center shrink-0">
            <Activity className="w-5 h-5 text-white" />
          </div>
          {!collapsed && (
            <span className="text-base lg:text-lg font-bold text-slate-100 truncate">Industrial AI</span>
          )}
        </div>
        {!collapsed && onToggleCollapse && (
          <button
            onClick={onToggleCollapse}
            className="p-1.5 text-slate-400 hover:text-slate-200 hover:bg-slate-700 rounded-lg transition-colors"
            aria-label={t('common.close')}
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
        )}
        {collapsed && onToggleCollapse && (
          <button
            onClick={onToggleCollapse}
            className="absolute -right-3 top-4 w-6 h-6 bg-slate-700 border border-slate-600 rounded-full flex items-center justify-center text-slate-400 hover:text-slate-200 hover:bg-slate-600 transition-colors z-10"
            aria-label={t('common.openMenu')}
          >
            <ChevronRight className="w-3 h-3" />
          </button>
        )}
        {showCloseButton && onClose && (
          <button 
            onClick={onClose} 
            className="p-2 text-slate-400 hover:text-slate-200 active:text-slate-100 active:bg-slate-700 rounded-lg touch-manipulation"
            aria-label={t('common.close')}
          >
            <X className="w-5 h-5" />
          </button>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto py-4 scrollbar-thin overscroll-contain">
        <ul className="space-y-0.5 px-2">
          {menuItems.map((item) => (
            <li key={item.path}>
              <Link
                to={item.path}
                title={collapsed ? item.label : undefined}
                className={`
                  flex items-center gap-3 px-3 py-2.5 lg:py-2 rounded-lg text-slate-300
                  transition-colors touch-manipulation
                  ${collapsed ? 'justify-center' : ''}
                  ${isActive(item.path)
                    ? 'bg-primary-600/20 text-primary-400'
                    : 'hover:bg-slate-700/50 hover:text-slate-100 active:bg-slate-700'
                  }
                `}
                onClick={onNavClick}
              >
                <item.icon className="w-5 h-5 shrink-0" />
                {!collapsed && (
                  <span className="flex-1 truncate">{item.label}</span>
                )}
                {!collapsed && isActive(item.path) && (
                  <ChevronRight className="w-4 h-4 text-primary-400" />
                )}
              </Link>
            </li>
          ))}
          
          {isAdmin && (
            <>
              <li className="pt-4">
                {!collapsed && (
                  <span className="px-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                    {t('nav.admin')}
                  </span>
                )}
                {collapsed && (
                  <div className="border-t border-slate-700 mx-2" />
                )}
              </li>
              {adminItems.map((item) => (
                <li key={item.path}>
                  <Link
                    to={item.path}
                    title={collapsed ? item.label : undefined}
                    className={`
                      flex items-center gap-3 px-3 py-2.5 lg:py-2 rounded-lg text-slate-300
                      transition-colors touch-manipulation
                      ${collapsed ? 'justify-center' : ''}
                      ${isActive(item.path)
                        ? 'bg-primary-600/20 text-primary-400'
                        : 'hover:bg-slate-700/50 hover:text-slate-100 active:bg-slate-700'
                      }
                    `}
                    onClick={onNavClick}
                  >
                    <item.icon className="w-5 h-5 shrink-0" />
                    {!collapsed && (
                      <span className="flex-1 truncate">{item.label}</span>
                    )}
                    {!collapsed && isActive(item.path) && (
                      <ChevronRight className="w-4 h-4 text-primary-400" />
                    )}
                  </Link>
                </li>
              ))}
            </>
          )}
        </ul>
      </nav>

      {/* Footer */}
      <div className="border-t border-slate-700 p-3 lg:p-4 shrink-0">
        {/* Language switcher - 隐藏折叠状态 */}
        {!collapsed && (
          <div className="flex items-center gap-2 mb-3">
            <Globe className="w-4 h-4 text-slate-400" />
            <button
              onClick={() => setLanguage(language === 'zh' ? 'en' : 'zh')}
              className="text-sm text-slate-400 hover:text-slate-200 active:text-slate-100 touch-manipulation py-1 px-2 rounded"
            >
              {language === 'zh' ? '中文' : 'English'}
            </button>
          </div>
        )}

        {/* User info */}
        {user && (
          <div data-testid="user-menu" className={`flex items-center justify-between py-2 px-2 bg-slate-700/50 rounded-lg ${collapsed ? 'justify-center' : ''}`}>
            <div className={`flex items-center gap-2 min-w-0 ${collapsed ? 'justify-center' : ''}`}>
              <div className="w-8 h-8 rounded-full bg-slate-600 flex items-center justify-center shrink-0" title={collapsed ? user.username : undefined}>
                <Users className="w-4 h-4 text-slate-300" />
              </div>
              {!collapsed && (
                <div className="min-w-0">
                  <div className="text-sm font-medium text-slate-200 truncate">{user.username}</div>
                  <div className="text-xs text-slate-400 truncate">{user.role}</div>
                </div>
              )}
            </div>
            {!collapsed && (
              <button
                data-testid="logout-btn"
                onClick={onLogout}
                className="p-2 text-slate-400 hover:text-red-400 active:text-red-300 active:bg-slate-600 rounded-lg transition-colors touch-manipulation shrink-0"
                aria-label={t('auth.logout')}
              >
                <LogOut className="w-5 h-5" />
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}