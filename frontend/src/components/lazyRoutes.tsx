import { lazy, Suspense, ComponentType } from 'react';
import { RouteLoader } from './LoadingSpinner';

// 路由组件懒加载配置
// 使用命名导出以便预加载

// 认证相关
export const LoginPage = lazy(() => import('./LoginPage'));
export const App = lazy(() => import('./App'));

// 仪表板
export const FleetDashboard = lazy(() => import('./FleetDashboard'));
export const ROIDashboard = lazy(() => import('./ROIDashboard'));
export const SystemStatus = lazy(() => import('./SystemStatus'));

// 设备管理
export const DeviceManager = lazy(() => import('./DeviceManager'));
export const DeviceDetail = lazy(() => import('./DeviceDetail'));
export const DigitalTwinPanel = lazy(() => import('./DigitalTwinPanel'));

// AI 功能
export const AITeamDashboard = lazy(() => import('./AITeamDashboard'));
export const KnowledgeGraph = lazy(() => import('./KnowledgeGraph'));

// 运维管理
export const WorkOrderBoard = lazy(() => import('./WorkOrderBoard'));
export const RuleManager = lazy(() => import('./RuleManager'));
export const NotificationCenter = lazy(() => import('./NotificationCenter'));
export const BlackBoxCenter = lazy(() => import('./BlackBoxCenter'));

// 报表
export const ReportCenter = lazy(() => import('./ReportCenter'));

// 用户管理
export const UserManager = lazy(() => import('./UserManager'));

/**
 * 创建带加载状态的懒加载组件
 * FE-P3-11: 移除泛型约束，使用简单类型定义，移除 any
 */
export function createLazyPage(
  importFn: () => Promise<{ default: ComponentType }>
) {
  const LazyComponent = lazy(importFn);
  
  return (
    <Suspense fallback={<RouteLoader />}>
      <LazyComponent />
    </Suspense>
  );
}

/**
 * 预加载关键路由
 * 在应用空闲时预加载可能访问的页面
 */
export function preloadCriticalRoutes() {
  // 首页仪表板通常是第一个访问的页面
  import('./FleetDashboard');
  
  // 设备管理是常用功能
  import('./DeviceManager');
}

/**
 * 预加载所有路由
 * 用于快速导航
 */
export function preloadAllRoutes() {
  Promise.all([
    import('./LoginPage'),
    import('./App'),
    import('./FleetDashboard'),
    import('./DeviceManager'),
    import('./DeviceDetail'),
    import('./DigitalTwinPanel'),
    import('./AITeamDashboard'),
    import('./KnowledgeGraph'),
    import('./WorkOrderBoard'),
    import('./RuleManager'),
    import('./NotificationCenter'),
    import('./BlackBoxCenter'),
    import('./ReportCenter'),
    import('./ROIDashboard'),
    import('./SystemStatus'),
    import('./UserManager'),
  ]).catch(console.warn);
}