import { lazy, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import I18nProvider from './i18n';
import { ToastProvider } from './components/Toast';
import { MobileProvider } from './components/MobileProvider';
import AuthProvider from './components/AuthContext';
import { performanceMonitor, usePerformance } from './lib/performance';
import './index.css';

// Lazy load components - 懒加载所有路由组件
const LoginPage = lazy(() => import('./components/LoginPage'));
const App = lazy(() => import('./components/App'));
const FleetDashboard = lazy(() => import('./components/FleetDashboard'));
const DeviceDetail = lazy(() => import('./components/DeviceDetail'));
const DeviceManager = lazy(() => import('./components/DeviceManager'));
const RuleManager = lazy(() => import('./components/RuleManager'));
const UserManager = lazy(() => import('./components/UserManager'));
const SystemStatus = lazy(() => import('./components/SystemStatus'));
const DigitalTwinPanel = lazy(() => import('./components/DigitalTwinPanel'));
const KnowledgeGraph = lazy(() => import('./components/KnowledgeGraph'));
const AITeamDashboard = lazy(() => import('./components/AITeamDashboard'));
const WorkOrderBoard = lazy(() => import('./components/WorkOrderBoard'));
const NotificationCenter = lazy(() => import('./components/NotificationCenter'));
const ROIDashboard = lazy(() => import('./components/ROIDashboard'));
const BlackBoxCenter = lazy(() => import('./components/BlackBoxCenter'));
const ReportCenter = lazy(() => import('./components/ReportCenter'));
const TelemetryPage = lazy(() => import('./components/TelemetryPage'));
const AlertsPage = lazy(() => import('./components/AlertsPage'));
const AlertReportPage = lazy(() => import('./components/AlertReportPage'));

/**
 * 性能监控包装组件
 */
function PerformanceWrapper({ children }: { children: React.ReactNode }) {
  usePerformance('Main');
  
  useEffect(() => {
    // 收集性能指标
    const unsubscribe = performanceMonitor.onMetrics((metrics) => {
      // 仅在开发环境输出
      if (import.meta.env.DEV) {
        // eslint-disable-next-line no-console
        console.log('[Performance Metrics]', {
          FCP: metrics.fcp ? `${metrics.fcp.toFixed(0)}ms` : 'N/A',
          LCP: metrics.lcp ? `${metrics.lcp.toFixed(0)}ms` : 'N/A',
          FID: metrics.fid ? `${metrics.fid.toFixed(0)}ms` : 'N/A',
          CLS: metrics.cls ? metrics.cls.toFixed(4) : 'N/A',
          TTFB: metrics.ttfb ? `${metrics.ttfb.toFixed(0)}ms` : 'N/A',
          Load: metrics.loadComplete ? `${metrics.loadComplete.toFixed(0)}ms` : 'N/A',
        });
      }
      
      // 在生产环境可以上报到监控服务
      if (import.meta.env.PROD && metrics.lcp && metrics.lcp > 2500) {
        console.warn('[Performance] LCP exceeds recommended threshold (2.5s)');
      }
    });

    // 预加载关键路由
    const idleCallback = requestIdleCallback(() => {
      import('./components/FleetDashboard');
      import('./components/DeviceManager');
    });
    
    return () => {
      unsubscribe();
      cancelIdleCallback(idleCallback);
    };
  }, []);

  return <>{children}</>;
}

/**
 * 路由配置
 */
function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<App />}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<FleetDashboard />} />
        <Route path="devices" element={<DeviceManager />} />
        <Route path="devices/:id" element={<DeviceDetail />} />
        <Route path="rules" element={<RuleManager />} />
        <Route path="users" element={<UserManager />} />
        <Route path="system" element={<SystemStatus />} />
        <Route path="digital-twin" element={<DigitalTwinPanel />} />
        <Route path="knowledge-graph" element={<KnowledgeGraph />} />
        <Route path="ai-agent" element={<AITeamDashboard />} />
        <Route path="work-orders" element={<WorkOrderBoard />} />
        <Route path="notifications" element={<NotificationCenter />} />
        <Route path="roi" element={<ROIDashboard />} />
        <Route path="blackbox" element={<BlackBoxCenter />} />
        <Route path="reports" element={<ReportCenter />} />
        <Route path="telemetry" element={<TelemetryPage />} />
        <Route path="alerts" element={<AlertsPage />} />
        <Route path="alerts/report" element={<AlertReportPage />} />
      </Route>
    </Routes>
  );
}

/**
 * 主应用组件
 */
function Main() {
  return (
    <PerformanceWrapper>
      <MobileProvider>
        <I18nProvider defaultLanguage="zh">
          <BrowserRouter>
            <AuthProvider>
              <ToastProvider>
                <AppRoutes />
              </ToastProvider>
            </AuthProvider>
          </BrowserRouter>
        </I18nProvider>
      </MobileProvider>
    </PerformanceWrapper>
  );
}

export default Main;