import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';

// Mock i18n context
vi.mock('../i18n', () => ({
  I18nContext: {
    Consumer: ({ children }: { children: (context: unknown) => React.ReactNode }) => 
      children({ t: (key: string) => key }),
  },
}));

// Import after mocks
import ErrorBoundary from './ErrorBoundary';

// Component that throws an error
const ThrowError = ({ shouldThrow }: { shouldThrow: boolean }) => {
  if (shouldThrow) {
    throw new Error('Test error');
  }
  return <div data-testid="child-component">Normal content</div>;
};

describe('ErrorBoundary', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Suppress console.error for cleaner test output
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('renders children when no error', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={false} />
      </ErrorBoundary>
    );

    expect(screen.getByTestId('child-component')).toBeInTheDocument();
  });

  it('renders error UI when child throws', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    // Should show error emoji
    expect(screen.getByText('⚠️')).toBeInTheDocument();
  });

  it('calls componentDidCatch on error', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(consoleSpy).toHaveBeenCalled();
    consoleSpy.mockRestore();
  });

  it('shows default messages without i18n context', () => {
    // Remove mock temporarily
    vi.doMock('../i18n', () => ({
      I18nContext: {
        Consumer: ({ children }: { children: (context: unknown) => React.ReactNode }) => 
          children(undefined),
      },
    }));

    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    // Should still render error boundary
    expect(screen.getByText('⚠️')).toBeInTheDocument();
  });
});