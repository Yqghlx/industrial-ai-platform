import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Mock i18n
vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}));

import MobileNavBar from './MobileNavBar';

const renderWithRouter = (component: React.ReactNode) => {
  return render(<MemoryRouter>{component}</MemoryRouter>);
};

describe('MobileNavBar', () => {
  it('renders navigation bar', () => {
    renderWithRouter(<MobileNavBar />);
    // 检查nav元素存在
    const nav = document.querySelector('nav');
    expect(nav).toBeTruthy();
  });

  it('renders navigation links', () => {
    renderWithRouter(<MobileNavBar />);
    // 检查是否有链接
    const links = document.querySelectorAll('a');
    expect(links.length).toBeGreaterThan(0);
  });

  it('has correct navigation items', () => {
    renderWithRouter(<MobileNavBar />);
    // 检查是否有dashboard路径
    const links = Array.from(document.querySelectorAll('a'));
    const hasDashboard = links.some(link => link.getAttribute('href')?.includes('dashboard'));
    expect(hasDashboard).toBe(true);
  });
});