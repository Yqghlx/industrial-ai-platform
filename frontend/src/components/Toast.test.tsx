import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ToastProvider, useToast } from './Toast';
import React from 'react';

vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));

describe('Toast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  it('renders ToastProvider without crashing', () => {
    vi.useRealTimers();
    render(
      <ToastProvider>
        <div>Test</div>
      </ToastProvider>
    );
    expect(screen.getByText('Test')).toBeInTheDocument();
    vi.useFakeTimers();
  });

  it('shows toast when showToast is called', async () => {
    vi.useRealTimers();
    const TestComponent = () => {
      const { showToast } = useToast();
      return (
        <button onClick={() => showToast({ type: 'success', message: 'Test Toast' })}>
          Show Toast
        </button>
      );
    };

    render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    );

    fireEvent.click(screen.getByText('Show Toast'));
    
    await waitFor(() => {
      expect(screen.getByText('Test Toast')).toBeInTheDocument();
    }, { timeout: 1000 });
    
    vi.useFakeTimers();
  });

  it('throws error when useToast is used outside ToastProvider', () => {
    const TestComponent = () => {
      useToast();
      return null;
    };

    expect(() => render(<TestComponent />)).toThrow('useToast must be used within ToastProvider');
  });
});