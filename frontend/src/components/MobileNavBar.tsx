import React from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { 
  LayoutDashboard, 
  Settings, 
  Bell, 
  Bot, 
  Activity 
} from 'lucide-react';
import { useI18n } from '../i18n';

interface NavItem {
  path: string;
  icon: React.ComponentType<{ className?: string }>;
  labelKey: string;
}

export default function MobileNavBar() {
  const { t } = useI18n();
  const location = useLocation();

  const navItems: NavItem[] = [
    { path: '/dashboard', icon: LayoutDashboard, labelKey: 'nav.dashboard' },
    { path: '/devices', icon: Settings, labelKey: 'nav.devices' },
    { path: '/digital-twin', icon: Activity, labelKey: 'nav.digitalTwin' },
    { path: '/notifications', icon: Bell, labelKey: 'nav.notifications' },
    { path: '/ai-agent', icon: Bot, labelKey: 'nav.aiAgent' },
  ];

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 lg:hidden bg-slate-800 border-t border-slate-700 safe-area-bottom">
      <div className="flex items-center justify-around h-16 px-2">
        {navItems.map((item) => {
          const isActive = location.pathname === item.path;
          const Icon = item.icon;
          
          return (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) =>
                `flex flex-col items-center justify-center flex-1 h-full py-1 transition-colors touch-manipulation ${
                  isActive
                    ? 'text-primary-400'
                    : 'text-slate-400 active:text-slate-200'
                }`
              }
            >
              <Icon className={`w-5 h-5 ${isActive ? 'scale-110' : ''} transition-transform`} />
              <span className="text-xs mt-1 truncate max-w-full">
                {t(item.labelKey)}
              </span>
              {isActive && (
                <div className="absolute bottom-0 w-8 h-0.5 bg-primary-500 rounded-full" />
              )}
            </NavLink>
          );
        })}
      </div>
    </nav>
  );
}