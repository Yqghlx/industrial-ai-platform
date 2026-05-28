import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import Skeleton, { SkeletonGrid, SkeletonStats, SkeletonTable } from './Skeleton';

describe('Skeleton', () => {
  it('renders text variant', () => {
    render(<Skeleton variant="text" />);
    const skeleton = document.querySelector('.animate-pulse');
    expect(skeleton).toBeTruthy();
  });

  it('renders card variant', () => {
    render(<Skeleton variant="card" />);
    const skeleton = document.querySelector('.animate-pulse');
    expect(skeleton).toBeTruthy();
  });

  it('renders circle variant', () => {
    render(<Skeleton variant="circle" />);
    const skeleton = document.querySelector('.rounded-full');
    expect(skeleton).toBeTruthy();
  });

  it('renders avatar variant', () => {
    render(<Skeleton variant="avatar" />);
    const skeleton = document.querySelector('.rounded-full');
    expect(skeleton).toBeTruthy();
  });

  it('renders chart variant', () => {
    render(<Skeleton variant="chart" />);
    const skeleton = document.querySelector('.animate-pulse');
    expect(skeleton).toBeTruthy();
  });

  it('renders button variant', () => {
    render(<Skeleton variant="button" />);
    const skeleton = document.querySelector('.rounded-md');
    expect(skeleton).toBeTruthy();
  });

  it('accepts width and height props', () => {
    render(<Skeleton width={200} height={100} />);
    const skeleton = document.querySelector('.animate-pulse');
    expect(skeleton).toBeTruthy();
  });

  it('renders multiple lines for text variant', () => {
    render(<Skeleton variant="text" lines={3} />);
    const skeletons = document.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBe(3);
  });

  it('applies custom className', () => {
    render(<Skeleton className="custom-skeleton" />);
    const skeleton = document.querySelector('.custom-skeleton');
    expect(skeleton).toBeTruthy();
  });
});

describe('SkeletonGrid', () => {
  it('renders grid with default count', () => {
    render(<SkeletonGrid />);
    const skeletons = document.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBe(6);
  });

  it('renders grid with custom count', () => {
    render(<SkeletonGrid count={4} />);
    const skeletons = document.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBe(4);
  });
});

describe('SkeletonStats', () => {
  it('renders stats with default count', () => {
    render(<SkeletonStats />);
    const metricCards = document.querySelectorAll('.metric-card');
    expect(metricCards.length).toBe(4);
  });

  it('renders stats with custom count', () => {
    render(<SkeletonStats count={8} />);
    const metricCards = document.querySelectorAll('.metric-card');
    expect(metricCards.length).toBe(8);
  });
});

describe('SkeletonTable', () => {
  it('renders table with default rows', () => {
    render(<SkeletonTable />);
    const tableContainer = document.querySelector('.table-container');
    expect(tableContainer).toBeTruthy();
  });

  it('renders table with custom rows', () => {
    render(<SkeletonTable rows={5} />);
    const rows = document.querySelectorAll('.animate-pulse');
    expect(rows.length).toBeGreaterThan(0);
  });
});