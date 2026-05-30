import React, { useState, useEffect, useCallback, Suspense } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from './AuthContext';
import Sidebar from './Sidebar';
import MobileNavBar from './MobileNavBar';
import { Menu, Bell, Activity } from 'lucide-react';
import { useToast } from './Toast';
import { PerformanceButton } from './PerformancePanel';
import { usePerformance } from '../lib/performance';
import { useSwipe, useIsMobile, useViewportHeight } from '../lib/useSwipe';
import { useWebSocket } from '../hooks/useWebSocket';
import { useI18n } from '../i18n';
import { ConfirmDialogProvider } from './UI/ConfirmDialog';
import { RouteLoader } from './LoadingSpinner';

export default function App() {
  usePerformance('App');
  useViewportHeight(); // Handle iOS Safari viewport height
  
  const { t } = useI18n();
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const { showToast } = useToast();
  const isMobile = useIsMobile();

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, navigate]);

  // Close sidebar on route change (mobile)
  useEffect(() => {
    if (isMobile) {
      setSidebarOpen(false);
    }
  }, [location.pathname, isMobile]);

  // WebSocket connection with compression support
  const { isConnected: wsConnected } = useWebSocket({
    onMessage: (message) => {
      if (message.type === 'alert') {
        showToast({
          type: 'warning',
          message: (message.payload as { message: string }).message,
        });
      } else if (message.type === 'telemetry') {
        // Update in real-time
      } else if (message.type === 'ping') {
        // Handle ping
      } else if (message.type === 'connected') {
        // Connection established
      }
    },
    onError: (error) => {
      console.error('WebSocket error:', error);
    },
  });

  // Swipe handlers for sidebar
  const swipeHandlers = useSwipe({
    onSwipeRight: () => {
      if (isMobile && !sidebarOpen) {
        setSidebarOpen(true);
      }
    },
    onSwipeLeft: () => {
      if (isMobile && sidebarOpen) {
        setSidebarOpen(false);
      }
    },
    threshold: 50,
  });

  const closeSidebar = useCallback(() => {
    setSidebarOpen(false);
  }, []);

  const toggleSidebarCollapsed = useCallback(() => {
    setSidebarCollapsed(prev => !prev);
  }, []);

  if (!isAuthenticated) {
    return null;
  }

  return (
    <ConfirmDialogProvider>
      <div 
        className="flex h-screen bg-slate-900 overflow-hidden"
        {...swipeHandlers}
      >
      {/* Sidebar */}
      <Sidebar isOpen={sidebarOpen} onClose={closeSidebar} collapsed={sidebarCollapsed} onToggleCollapse={toggleSidebarCollapsed} />

      {/* Main content */}
      <div className={`flex-1 flex flex-col overflow-hidden ${sidebarCollapsed ? 'lg:ml-16' : 'lg:ml-64'} transition-all duration-300`}>
        {/* Top navbar */}
        <header className="h-14 lg:h-16 bg-slate-800 border-b border-slate-700 flex items-center justify-between px-3 lg:px-4 safe-area-top">
          <div className="flex items-center gap-2">
            <button
              onClick={() => isMobile ? setSidebarOpen(true) : toggleSidebarCollapsed()}
              className="p-2 text-slate-400 hover:text-slate-200 active:text-slate-100 active:bg-slate-700 rounded-lg touch-manipulation"
              aria-label={sidebarCollapsed ? t('common.openMenu') : t('common.close')}
            >
              <Menu className="w-5 h-5 lg:w-6 lg:h-6" />
            </button>
            
            {/* Mobile: Show page title */}
            <h1 className="text-base font-medium text-slate-100 lg:hidden truncate">
              Industrial AI
            </h1>
          </div>
          
          <div className="flex items-center gap-2 lg:gap-4">
            <div className="hidden sm:flex items-center gap-2">
              <Activity className={`w-5 h-5 ${wsConnected ? 'text-green-500 animate-pulse' : 'text-red-500'}`} />
              <span className="text-sm text-slate-400">{wsConnected ? t('common.connected') : t('common.disconnected')}</span>
            </div>
            
            <button 
              className="relative p-2 text-slate-400 hover:text-slate-200 active:text-slate-100 active:bg-slate-700 rounded-lg touch-manipulation"
              aria-label={t('nav.notifications')}
            >
              <Bell className="w-5 h-5" />
              <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full" aria-label={t('notification.unread')} />
            </button>
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 overflow-y-auto overflow-x-hidden p-3 lg:p-4 bg-slate-900 pb-20 lg:pb-4">
          <div className="max-w-7xl mx-auto">
            <Suspense fallback={<RouteLoader />}>
              <Outlet />
            </Suspense>
          </div>
        </main>
      </div>

      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 lg:hidden backdrop-blur-sm"
          onClick={closeSidebar}
          aria-hidden="true"
        />
      )}
      
      {/* Mobile bottom navigation */}
      <MobileNavBar />
      
      {/* Performance Monitor Button (Dev only) */}
      <PerformanceButton />
    </div>
    </ConfirmDialogProvider>
  );
}