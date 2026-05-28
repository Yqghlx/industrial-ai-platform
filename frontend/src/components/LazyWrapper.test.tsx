import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

import { LazyWrapper } from './LazyWrapper';

describe('LazyWrapper', () => {
  it('renders children', () => {
    render(
      <LazyWrapper>
        <div data-testid="child">Content</div>
      </LazyWrapper>
    );
    expect(screen.getByTestId('child')).toBeTruthy();
  });

  it('renders with spinner variant', () => {
    render(
      <LazyWrapper variant="spinner">
        <div>Content</div>
      </LazyWrapper>
    );
    expect(screen.getByText('Content')).toBeTruthy();
  });

  it('renders with minimal variant', () => {
    render(
      <LazyWrapper variant="minimal">
        <div>Content</div>
      </LazyWrapper>
    );
    expect(screen.getByText('Content')).toBeTruthy();
  });

  it('accepts delay prop', () => {
    render(
      <LazyWrapper delay={300}>
        <div>Content</div>
      </LazyWrapper>
    );
    expect(screen.getByText('Content')).toBeTruthy();
  });
});