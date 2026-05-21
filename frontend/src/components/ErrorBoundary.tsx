import React, { Component, ErrorInfo, ReactNode } from 'react';
import { I18nContext } from '../i18n';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

// 国际化消息
const getMessages = (t: (key: string, params?: Record<string, string | number>) => string) => ({
  errorTitle: t('errorBoundary.title', { defaultValue: '出错了' }),
  errorMessage: t('errorBoundary.message', { defaultValue: '发生了未知错误' }),
  refreshPage: t('errorBoundary.refresh', { defaultValue: '刷新页面' }),
  goBack: t('errorBoundary.goBack', { defaultValue: '返回上一页' }),
});

// 默认消息（备用）
const defaultMessages = {
  errorTitle: '出错了',
  errorMessage: '发生了未知错误',
  refreshPage: '刷新页面',
  goBack: '返回上一页',
};

export default class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught error:', error, errorInfo);
  }

  private handleGoBack = () => {
    window.history.back();
  };

  public render() {
    if (this.state.hasError) {
      return (
        <I18nContext.Consumer>
          {(context: { t: (key: string, params?: Record<string, string | number>) => string } | undefined) => {
            const messages = context ? getMessages(context.t) : defaultMessages;
            
            return (
              <div className="min-h-screen bg-slate-900 flex items-center justify-center p-4">
                <div className="card max-w-md">
                  <div className="card-body text-center">
                    <div className="text-6xl mb-4">⚠️</div>
                    <h2 className="text-xl font-bold text-slate-100 mb-2">
                      {messages.errorTitle}
                    </h2>
                    <p className="text-slate-400 mb-4">
                      {this.state.error?.message || messages.errorMessage}
                    </p>
                    <div className="flex gap-3 justify-center">
                      <button
                        onClick={() => window.location.reload()}
                        className="btn btn-primary"
                      >
                        {messages.refreshPage}
                      </button>
                      <button
                        onClick={this.handleGoBack}
                        className="btn btn-secondary"
                      >
                        {messages.goBack}
                      </button>
                    </div>
                    {import.meta.env.DEV && this.state.error && (
                      <details className="mt-4 text-left">
                        <summary className="text-slate-500 cursor-pointer text-sm">
                          错误详情 (仅开发环境显示)
                        </summary>
                        <pre className="mt-2 p-3 bg-slate-800 rounded text-xs text-red-400 overflow-auto max-h-40">
                          {this.state.error.stack}
                        </pre>
                      </details>
                    )}
                  </div>
                </div>
              </div>
            );
          }}
        </I18nContext.Consumer>
      );
    }

    return this.props.children;
  }
}