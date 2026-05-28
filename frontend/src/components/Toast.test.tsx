import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

// Mock i18n
vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}));

// Import after mocks
import { ToastProvider, useToast, Toast } from './Toast';

// Test component to trigger toast
const TestComponent = () => {
  const { showToast } = useToast();
  return (
    <div>
      <button 
        onClick={() => showToast({ type: 'success', message: 'Success message' })}
        data-testid="success-btn"
      >
        Success
      </button>
      <button 
        onClick={() => showToast({ type: 'error', message: 'Error message' })}
        data-testid="error-btn"
      >
        Error
      </button>
    </div>
  );
};

describe('Toast', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('provides useToast hook within provider', () => {
    render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    );

    expect(screen.getByTestId('success-btn')).toBeInTheDocument();
    expect(screen.getByTestId('error-btn')).toBeInTheDocument();
  });

  it('shows success toast on button click', async () => {
    render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    );

    const successBtn = screen.getByTestId('success-btn');
    fireEvent.click(successBtn);

    await waitFor(() => {
      expect(screen.getByText('Success message')).toBeInTheDocument();
    });
  });

  it('shows error toast on button click', async () => {
    render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    );

    const errorBtn = screen.getByTestId('error-btn');
    fireEvent.click(errorBtn);

    await waitFor(() => {
      expect(screen.getByText('Error message')).toBeInTheDocument();
    });
  });

  it('throws error when useToast is used outside provider', () => {
    // Suppress error boundary
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    expect(() => {
      render(<TestComponent />);
    }).toThrow('useToast must be used within ToastProvider');

    consoleSpy.mockRestore();
  });
});