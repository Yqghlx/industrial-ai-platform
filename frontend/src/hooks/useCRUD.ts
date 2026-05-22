// 通用CRUD Hook
// FIX-038: CRUD逻辑提取
// FE-P2-09: 支持多种API返回类型（直接返回和包装返回）

import { useState, useCallback, useEffect } from 'react';
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

// FE-P2-09: 支持多种API返回类型
interface CRUDConfig<T> {
  // 列表API - 返回分页响应（包含 data 和 total）
  apiGetAll: (page: number, pageSize: number) => Promise<{ data: T[]; total?: number }>;
  // 单项API - 直接返回对象
  apiGetOne: (id: string) => Promise<T>;
  // 创建API - 直接返回创建的对象
  apiCreate: (data: Partial<T>) => Promise<T>;
  // 更新API - 直接返回更新后的对象
  apiUpdate: (id: string, data: Partial<T>) => Promise<T>;
  // 删除API - 返回消息响应
  apiDelete: (id: string) => Promise<{ message?: string }>;
  entityName?: string;
  initialPageSize?: number;
  onError?: (error: string) => void;
  onSuccess?: (action: string) => void;
}

export function useCRUD<T extends { id?: string | number }>(
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
      const result = await config.apiGetOne(id);
      setState((prev) => ({ ...prev, loading: false }));
      handleSuccess('获取详情');
      return result;
    } catch (error) {
      handleError(error, '获取详情');
      return null;
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [config.apiGetOne, handleSuccess, handleError]);

  const create = useCallback(async (data: Partial<T>): Promise<T | null> => {
    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      const result = await config.apiCreate(data);
      if (result) {
        setState((prev) => ({
          ...prev,
          items: [...prev.items, result],
          total: prev.total + 1,
          loading: false,
        }));
        handleSuccess('创建');
        return result;
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
      const result = await config.apiUpdate(id, data);
      if (result) {
        setState((prev) => ({
          ...prev,
          items: prev.items.map((item) =>
            String(item.id) === String(id) ? result : item
          ),
          loading: false,
        }));
        handleSuccess('更新');
        return result;
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
        items: prev.items.filter((item) => String(item.id) !== String(id)),
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