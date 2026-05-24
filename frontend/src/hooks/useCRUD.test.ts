import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useCRUD } from './useCRUD';

interface TestItem {
  id?: number;
  name?: string;
}

describe('useCRUD', () => {
  const mockConfig = {
    apiGetAll: vi.fn().mockResolvedValue({ data: [], total: 0 }),
    apiGetOne: vi.fn(),
    apiCreate: vi.fn(),
    apiUpdate: vi.fn(),
    apiDelete: vi.fn(),
    entityName: 'TestItem',
    initialPageSize: 10,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockConfig.apiGetAll.mockResolvedValue({ data: [], total: 0 });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize with correct config', async () => {
    const { result } = renderHook(() => useCRUD<TestItem>(mockConfig));

    // Wait for initial useEffect fetch to complete
    await waitFor(() => {
      expect(result.current[0].loading).toBe(false);
    });

    const [state, actions] = result.current;
    expect(state.items).toEqual([]);
    expect(state.total).toBe(0);
    expect(state.pageSize).toBe(10);
    expect(typeof actions.fetchAll).toBe('function');
    expect(typeof actions.create).toBe('function');
  });

  it('should fetch data successfully', async () => {
    const mockData = [{ id: 1, name: 'test' }];
    
    // Override mock for this test - needs to return for initial AND manual fetch
    mockConfig.apiGetAll.mockResolvedValueOnce({ data: mockData, total: 1 });

    const { result } = renderHook(() => useCRUD<TestItem>(mockConfig));

    // Wait for initial load to complete
    await waitFor(() => {
      expect(result.current[0].loading).toBe(false);
    });

    // Now fetch manually with different data
    mockConfig.apiGetAll.mockResolvedValueOnce({ data: mockData, total: 1 });

    await act(async () => {
      await result.current[1].fetchAll();
    });

    await waitFor(() => {
      expect(result.current[0].items).toEqual(mockData);
      expect(result.current[0].total).toBe(1);
    });
  });

  it('should handle fetch error', async () => {
    // Override mock to throw error
    mockConfig.apiGetAll.mockRejectedValueOnce(new Error('Network error'));

    const { result } = renderHook(() => useCRUD<TestItem>(mockConfig));

    // Wait for initial load error
    await waitFor(() => {
      expect(result.current[0].loading).toBe(false);
    });

    // Error should be set
    await waitFor(() => {
      expect(result.current[0].error).toBeTruthy();
    });
  });

  it('should create new item', async () => {
    const newItem = { name: 'new' };
    const createdItem = { id: 2, name: 'new' };
    mockConfig.apiCreate.mockResolvedValueOnce(createdItem);

    const { result } = renderHook(() => useCRUD<TestItem>(mockConfig));

    // Wait for initial load
    await waitFor(() => {
      expect(result.current[0].loading).toBe(false);
    });

    await act(async () => {
      const res = await result.current[1].create(newItem);
      expect(res).toEqual(createdItem);
    });

    await waitFor(() => {
      expect(result.current[0].items).toContainEqual(createdItem);
    });
  });
});