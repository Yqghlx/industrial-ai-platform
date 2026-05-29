import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';

// 创建一个 mock I18nContext
const MockI18nContext = React.createContext(undefined);

vi.mock('../i18n', () => {
  const context = React.createContext(undefined);
  return {
    useI18n: () => ({ t: (k: string) => k }),
    I18nContext: context,
  };
});

import ErrorBoundary from './ErrorBoundary';

// 制造错误的子组件
const ThrowError = () => {
  throw new Error('Test error message');
};

describe('ErrorBoundary', () => {
  // 抑制 console.error 以保持测试输出干净
  const originalConsoleError = console.error;
  beforeEach(() => {
    console.error = vi.fn();
  });
  afterEach(() => {
    console.error = originalConsoleError;
  });

  it('renders children when no error', () => {
    render(
      <ErrorBoundary>
        <div data-testid="child">正常内容</div>
      </ErrorBoundary>
    );
    expect(screen.getByTestId('child')).toBeInTheDocument();
  });

  it('renders error UI when child throws', async () => {
    render(
      <ErrorBoundary>
        <ThrowError />
      </ErrorBoundary>
    );
    await waitFor(() => {
      // ErrorBoundary 显示 "出错了" 和 "刷新页面"（使用 defaultMessages）
      expect(screen.getByText('出错了')).toBeInTheDocument();
      expect(screen.getByText('刷新页面')).toBeInTheDocument();
    });
  });
});