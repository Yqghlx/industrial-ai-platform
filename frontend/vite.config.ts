import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { visualizer } from 'rollup-plugin-visualizer';

export default defineConfig(({ mode }) => ({
  plugins: [
    react(),
    // 打包分析插件
    mode === 'analyze' && visualizer({
      open: true,
      filename: 'stats.html',
      gzipSize: true,
      brotliSize: true,
    }),
  ].filter(Boolean),
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/docs': {
        target: 'http://localhost',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false, // 禁用 sourcemap 以减小体积
    minify: 'esbuild', // 使用 esbuild 压缩
    target: 'es2020', // 现代浏览器
    cssMinify: true, // CSS 压缩
    chunkSizeWarningLimit: 1000, // 警告阈值
    rollupOptions: {
      output: {
        // 更细粒度的代码分割
        manualChunks: (id) => {
          // React 核心
          if (id.includes('node_modules/react/') || 
              id.includes('node_modules/react-dom/') ||
              id.includes('node_modules/scheduler/')) {
            return 'react-core';
          }
          // React Router
          if (id.includes('node_modules/react-router-dom/') ||
              id.includes('node_modules/@remix-run/')) {
            return 'react-router';
          }
          // 图表库
          if (id.includes('node_modules/recharts/')) {
            return 'charts';
          }
          // 图标库
          if (id.includes('node_modules/lucide-react/')) {
            return 'icons';
          }
          // 知识图谱
          if (id.includes('node_modules/react-force-graph-2d/') ||
              id.includes('node_modules/d3-')) {
            return 'graph';
          }
          // 其他第三方库
          if (id.includes('node_modules/')) {
            return 'vendor';
          }
        },
        // 优化文件命名
        chunkFileNames: (chunkInfo) => {
          const facadeModuleId = chunkInfo.facadeModuleId 
            ? chunkInfo.facadeModuleId.replace(/\\/g, '/') 
            : '';
          
          // 按功能分组
          if (facadeModuleId.includes('/components/')) {
            const componentName = facadeModuleId.split('/components/')[1]?.split('/')[0];
            return `assets/components/${componentName || 'other'}-[hash].js`;
          }
          return 'assets/[name]-[hash].js';
        },
        assetFileNames: (assetInfo) => {
          // 按类型分组资源
          const info = assetInfo.name || '';
          if (/\.(png|jpe?g|gif|svg|webp|ico)$/i.test(info)) {
            return 'assets/images/[name]-[hash][extname]';
          }
          if (/\.(woff2?|eot|ttf|otf)$/i.test(info)) {
            return 'assets/fonts/[name]-[hash][extname]';
          }
          if (/\.css$/i.test(info)) {
            return 'assets/css/[name]-[hash][extname]';
          }
          return 'assets/[name]-[hash][extname]';
        },
      },
    },
    // 压缩选项
    esbuild: {
      drop: ['console', 'debugger'], // 生产环境移除 console 和 debugger
      legalComments: 'none', // 移除注释
    },
    // 启用 CSS 代码分割
    cssCodeSplit: true,
  },
  // 优化依赖预构建
  optimizeDeps: {
    include: ['react', 'react-dom', 'react-router-dom', 'lucide-react'],
    exclude: ['react-force-graph-2d'], // 大型库延迟加载
    // 强制预构建，避免动态发现
    force: false,
  },
  // 启用 esbuild 优化
  esbuild: {
    logOverride: { 'this-is-undefined-in-esm': 'silent' },
  },
}));