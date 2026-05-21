import '@testing-library/jest-dom';
import { afterEach } from 'vitest';

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

// Mock window.location
// Note: window.location cannot be deleted in jsdom, so we use Object.defineProperty
Object.defineProperty(window, 'location', {
  value: {
    ...window.location,
    origin: 'http://localhost:3000',
  },
  writable: true,
  configurable: true,
});

// Clear all mocks after each test
afterEach(() => {
  localStorage.clear();
});