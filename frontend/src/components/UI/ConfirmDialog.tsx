// 确认对话框组件
// FIX-040: 自定义确认框替换原生confirm

import React, { useState, useCallback, createContext, useContext } from 'react';

// Simple translation fallback if i18next not available
const t = (key: string): string => {
  const translations: Record<string, string> = {
    'common.confirm': '确认',
    'common.cancel': '取消',
  };
  return translations[key] || key;
};

interface ConfirmDialogState {
  isOpen: boolean;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'danger' | 'warning' | 'info';
  onConfirm: () => void;
  onCancel: () => void;
}

interface ConfirmDialogProps {
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'danger' | 'warning' | 'info';
  onConfirm: () => void;
  onCancel: () => void;
  isOpen: boolean;
}

// Context for global confirm dialog
const ConfirmDialogContext = createContext<{
  showConfirm: (options: Omit<ConfirmDialogState, 'isOpen' | 'onConfirm' | 'onCancel'>) => Promise<boolean>;
}>({
  showConfirm: () => Promise.resolve(false),
});

export function useConfirmDialog() {
  return useContext(ConfirmDialogContext);
}

// Confirm Dialog Component
export function ConfirmDialog({
  title,
  message,
  confirmText,
  cancelText,
  variant = 'warning',
  onConfirm,
  onCancel,
  isOpen,
}: ConfirmDialogProps) {
  if (!isOpen) return null;

  const variantColors = {
    danger: 'bg-red-600 hover:bg-red-700',
    warning: 'bg-yellow-600 hover:bg-yellow-700',
    info: 'bg-blue-600 hover:bg-blue-700',
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      role="dialog"
      aria-modal="true"
      aria-labelledby="confirm-dialog-title"
      aria-describedby="confirm-dialog-message"
    >
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onCancel}
        aria-hidden="true"
      />

      {/* Dialog */}
      <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full mx-4 p-6">
        {/* Title */}
        <h2
          id="confirm-dialog-title"
          className="text-lg font-semibold text-gray-900 dark:text-white mb-2"
        >
          {title || t('common.confirm')}
        </h2>

        {/* Message */}
        <p
          id="confirm-dialog-message"
          className="text-sm text-gray-600 dark:text-gray-300 mb-6"
        >
          {message}
        </p>

        {/* Actions */}
        <div className="flex gap-3 justify-end">
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 
                     bg-gray-100 dark:bg-gray-700 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600
                     focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500"
            aria-label={cancelText || t('common.cancel')}
          >
            {cancelText || t('common.cancel')}
          </button>

          <button
            type="button"
            onClick={onConfirm}
            className={`px-4 py-2 text-sm font-medium text-white rounded-md
                     focus:outline-none focus:ring-2 focus:ring-offset-2
                     ${variantColors[variant]}`}
            aria-label={confirmText || t('common.confirm')}
          >
            {confirmText || t('common.confirm')}
          </button>
        </div>
      </div>
    </div>
  );
}

// Provider for global confirm dialog usage
export function ConfirmDialogProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<ConfirmDialogState>({
    isOpen: false,
    title: '',
    message: '',
    onConfirm: () => {},
    onCancel: () => {},
  });

  // resolveRef kept for potential future use (confirm dialog promise resolution)
  const [, setResolveRef] = useState<((value: boolean) => void) | null>(null);

  const showConfirm = useCallback(
    (options: Omit<ConfirmDialogState, 'isOpen' | 'onConfirm' | 'onCancel'>): Promise<boolean> => {
      return new Promise<boolean>((resolve) => {
        setResolveRef(() => resolve);
        setState({
          ...options,
          isOpen: true,
          onConfirm: () => {
            setState((prev) => ({ ...prev, isOpen: false }));
            resolve(true);
          },
          onCancel: () => {
            setState((prev) => ({ ...prev, isOpen: false }));
            resolve(false);
          },
        });
      });
    },
    []
  );

  return (
    <ConfirmDialogContext.Provider value={{ showConfirm }}>
      {children}
      <ConfirmDialog
        title={state.title}
        message={state.message}
        confirmText={state.confirmText}
        cancelText={state.cancelText}
        variant={state.variant}
        onConfirm={state.onConfirm}
        onCancel={state.onCancel}
        isOpen={state.isOpen}
      />
    </ConfirmDialogContext.Provider>
  );
}

// 使用示例:
// function MyComponent() {
//   const { showConfirm } = useConfirmDialog();
//   
//   const handleDelete = async () => {
//     const confirmed = await showConfirm({
//       title: '删除确认',
//       message: '确定要删除此项目吗？此操作无法撤销。',
//       variant: 'danger',
//       confirmText: '删除',
//       cancelText: '取消',
//     });
//     if (confirmed) {
//       // 执行删除操作
//     }
//   };
//   
//   return <button onClick={handleDelete}>删除</button>;
// }

export default ConfirmDialog;