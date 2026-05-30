import React, { createContext, useContext, useState, ReactNode, useCallback } from 'react';
import { useI18n } from '../i18n';

// Close icon component
const CloseIcon = () => (
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    width="14" 
    height="14" 
    viewBox="0 0 24 24" 
    fill="none" 
    stroke="currentColor" 
    strokeWidth="2" 
    strokeLinecap="round" 
    strokeLinejoin="round"
  >
    <line x1="18" y1="6" x2="6" y2="18"></line>
    <line x1="6" y1="6" x2="18" y2="18"></line>
  </svg>
);

interface ToastMessage {
  id: number;
  type: 'success' | 'error' | 'warning' | 'info';
  message: string;
}

interface ToastContextType {
  showToast: (toast: Omit<ToastMessage, 'id'>) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within ToastProvider');
  }
  return context;
}

interface ToastProviderProps {
  children: ReactNode;
}

export function ToastProvider({ children }: ToastProviderProps) {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  const showToast = useCallback((toast: Omit<ToastMessage, 'id'>) => {
    const id = Date.now();
    setToasts((prev) => [...prev, { ...toast, id }]);
    
    // Auto remove after 5 seconds
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 5000);
  }, []);

  const removeToast = (id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      <Toast toasts={toasts} removeToast={removeToast} />
    </ToastContext.Provider>
  );
}

interface ToastProps {
  toasts: ToastMessage[];
  removeToast: (id: number) => void;
}

export default function Toast({ toasts, removeToast }: ToastProps) {
  const { t } = useI18n();
  if (toasts.length === 0) return null;

  return (
    <div className="fixed top-4 left-4 right-4 lg:left-auto lg:right-4 z-50 space-y-2 flex flex-col items-center lg:items-end" role="status" aria-live="polite">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={`
            w-full max-w-sm lg:w-80
            bg-slate-800 border rounded-lg shadow-lg px-4 py-3
            animate-slide-down touch-manipulation relative
            ${toast.type === 'success' ? 'border-green-500' :
              toast.type === 'error' ? 'border-red-500' :
              toast.type === 'warning' ? 'border-yellow-500' :
              'border-slate-500'}
          `}
        >
          <div className="flex items-center gap-3">
            {toast.type === 'success' && (
              <div className="w-5 h-5 rounded-full bg-green-500 flex items-center justify-center shrink-0">
                <span className="text-white text-xs">✓</span>
              </div>
            )}
            {toast.type === 'error' && (
              <div className="w-5 h-5 rounded-full bg-red-500 flex items-center justify-center shrink-0">
                <span className="text-white text-xs">✕</span>
              </div>
            )}
            {toast.type === 'warning' && (
              <div className="w-5 h-5 rounded-full bg-yellow-500 flex items-center justify-center shrink-0">
                <span className="text-white text-xs">!</span>
              </div>
            )}
            {toast.type === 'info' && (
              <div className="w-5 h-5 rounded-full bg-blue-500 flex items-center justify-center shrink-0">
                <span className="text-white text-xs">i</span>
              </div>
            )}
            <span className="text-slate-100 text-sm flex-1">{toast.message}</span>
          </div>
          <button
            onClick={() => removeToast(toast.id)}
            className="absolute top-2 right-2 p-1 text-slate-400 hover:text-slate-200 active:text-slate-100 touch-manipulation min-w-[44px] min-h-[44px] flex items-center justify-center"
            aria-label={t('common.close')}
          >
            <CloseIcon />
          </button>
        </div>
      ))}
    </div>
  );
}