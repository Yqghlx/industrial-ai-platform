
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useCRUD } from './useCRUD';

// Mock API
vi.mock('../lib/api', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

describe('useCRUD', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with empty state', () => {
    const { result } = renderHook(() => useCRUD('/api/test'));
    expect(result.current.data).toEqual([]);
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('should fetch data successfully', async () => {
    const mockData = [{ id: 1, name: 'test' }];
    const { api } = await import('./api');
    vi.mocked(api.get).mockResolvedValueOnce({ data: mockData });

    const { result } = renderHook(() => useCRUD('/api/test'));
    
    await waitFor(() => {
      expect(result.current.data).toEqual(mockData);
    });
  });

  it('should handle fetch error', async () => {
    const mockError = new Error('Network error');
    const { api } = await import('./api');
    vi.mocked(api.get).mockRejectedValueOnce(mockError);

    const { result } = renderHook(() => useCRUD('/api/test'));
    
    await waitFor(() => {
      expect(result.current.error).toBe(mockError);
    });
  });

  it('should create new item', async () => {
    const newItem = { name: 'new' };
    const createdItem = { id: 2, name: 'new' };
    const { api } = await import('./api');
    vi.mocked(api.post).mockResolvedValueOnce({ data: createdItem });

    const { result } = renderHook(() => useCRUD('/api/test'));
    
    await result.current.create(newItem);
    
    await waitFor(() => {
      expect(result.current.data).toContainEqual(createdItem);
    });
  });
});
