// 通用CRUD Hook
// FIX-038: CRUD逻辑提取

import { useState, useCallback, useEffect } from 'react';
import { ApiResponse } from '../lib/typeGuards';
import { sanitizeErrorMessage } from '../utils/security';

// Simple toast fallback
const showToast = (_options: { type: string; message: string }) => {
  // Toast placeholder - replace with actual toast implementation
};

interface CRUDState<T> {
  items: T[];
  loading: boolean;
  error: string | null;
  total: number;
  page: number;
  pageSize: number;
}

interface CRUDActions<T> {
  fetchAll: (page?: number, pageSize?: number) => Promise<void>;
  fetchOne: (id: string) => Promise<T | null>;
  create: (data: Partial<T>) => Promise<T | null>;
  update: (id: string, data: Partial<T>) => Promise<T | null>;
  delete: (id: string) => Promise<boolean>;
  refresh: () => Promise<void>;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  resetError: () => void;
}

interface CRUDConfig<T> {
  apiGetAll: (page: number, pageSize: number) => Promise<ApiResponse<T[]>>;
  apiGetOne: (id: string) => Promise<ApiResponse<T>>;
  apiCreate: (data: Partial<T>) => Promise<ApiResponse<T>>;
  apiUpdate: (id: string, data: Partial<T>) => Promise<ApiResponse<T>>;
  apiDelete: (id: string) => Promise<ApiResponse<void>>;
  entityName?: string;
  initialPageSize?: number;
  onError?: (error: string) => void;
  onSuccess?: (action: string) => void;
}

export function useCRUD<T extends { id?: string }>(
  config: CRUDConfig<T>
): [CRUDState<T>, CRUDActions<T>] {
  const [state, setState] = useState<CRUDState<T>>({
    items: [],
    loading: false,
    error: null,
    total: 0,
    page: 1,
    pageSize: config.initialPageSize || 20,
  });

  const handleError = useCallback((error: unknown, action: string) => {
    const message = sanitizeErrorMessage(error);
    setState((prev) => ({ ...prev, error: message, loading: false }));
    if (config.onError) {
      config.onError(message);
    } else {
      showToast({ type: 'error', message: `${action}失败: ${message}` });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [config.onError]);

  const handleSuccess = useCallback((action: string) => {
    setState((prev) => ({ ...prev, error: null }));
    if (config.onSuccess) {
      config.onSuccess(action);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [config.onSuccess]);

  const fetchAll = useCallback(async (page?: number, pageSize?: number) => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    const p = page ?? state.page;
    const ps = pageSize ?? state.pageSize;

    try {
      const response = await config.apiGetAll(p, ps);
      setState((prev) => ({
        ...prev,
        items: response.data || [],
        total: response.total || (response.data?.length || 0),
        page: p,
        pageSize: ps,
        loading: false,
      }));
      handleSuccess('获取列表');
    } catch (error) {
      handleError(error, '获取列表');
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [state.page, state.pageSize, config.apiGetAll, handleSuccess, handleError]);

  const fetchOne = useCallback(async (id: string): Promise<T | null> => {
    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      const response = await config.apiGetOne(id);
      setState((prev) => ({ ...prev, loading: false }));
      handleSuccess('获取详情');
      return response.data || null;
    } catch (error) {
      handleError(error, '获取详情');
      return null;
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [config.apiGetOne, handleSuccess, handleError]);

  const create = useCallback(async (data: Partial<T>): Promise<T | null> => {
    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      const response = await config.apiCreate(data);
      if (response.data) {
        setState((prev) => ({
          ...prev,
          items: [...prev.items, response.data!],
          total: prev.total + 1,
          loading: false,
        }));
        handleSuccess('创建');
        return response.data;
      }
      setState((prev) => ({ ...prev, loading: false }));
      return null;
    } catch (error) {
      handleError(error, '创建');
      return null;
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [config.apiCreate, handleSuccess, handleError]);

  const update = useCallback(async (id: string, data: Partial<T>): Promise<T | null> => {
    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      const response = await config.apiUpdate(id, data);
      if (response.data) {
        setState((prev) => ({
          ...prev,
          items: prev.items.map((item) =>
            item.id === id ? response.data! : item
          ),
          loading: false,
        }));
        handleSuccess('更新');
        return response.data;
      }
      setState((prev) => ({ ...prev, loading: false }));
      return null;
    } catch (error) {
      handleError(error, '更新');
      return null;
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [config.apiUpdate, handleSuccess, handleError]);

  const deleteItem = useCallback(async (id: string): Promise<boolean> => {
    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      await config.apiDelete(id);
      setState((prev) => ({
        ...prev,
        items: prev.items.filter((item) => item.id !== id),
        total: prev.total - 1,
        loading: false,
      }));
      handleSuccess('删除');
      return true;
    } catch (error) {
      handleError(error, '删除');
      return false;
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [config.apiDelete, handleSuccess, handleError]);

  const refresh = useCallback(async () => {
    await fetchAll(state.page, state.pageSize);
  }, [fetchAll, state.page, state.pageSize]);

  const setPage = useCallback((page: number) => {
    setState((prev) => ({ ...prev, page }));
  }, []);

  const setPageSize = useCallback((size: number) => {
    setState((prev) => ({ ...prev, pageSize: size, page: 1 }));
  }, []);

  const resetError = useCallback(() => {
    setState((prev) => ({ ...prev, error: null }));
  }, []);

  // 初始加载
  useEffect(() => {
    fetchAll();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // 仅初始加载一次

  const actions: CRUDActions<T> = {
    fetchAll,
    fetchOne,
    create,
    update,
    delete: deleteItem,
    refresh,
    setPage,
    setPageSize,
    resetError,
  };

  return [state, actions];
}

export default useCRUD;