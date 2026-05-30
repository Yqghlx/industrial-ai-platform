import React, { useState, useEffect, useRef } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { useToast } from './Toast';
import { Bot, Send, User, Loader, Sparkles, Trash2 } from 'lucide-react';
import { AgentResponse } from '../types/api';
import { getAgentColor } from '../lib/colorUtils';
import ReactMarkdown from 'react-markdown';

const MAX_RESPONSES = 50;
const STORAGE_KEY = 'ai-chat-history';

interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  agent?: string;
  timestamp: number;
}

export default function AITeamDashboard() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [query, setQuery] = useState('');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const chatEndRef = useRef<HTMLDivElement>(null);

  // 加载历史会话
  useEffect(() => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY);
      if (saved) {
        const parsed = JSON.parse(saved) as ChatMessage[];
        if (Array.isArray(parsed) && parsed.length > 0) {
          setMessages(parsed);
        }
      }
    } catch { /* 忽略解析错误 */ }
  }, []);

  // 保存会话到 localStorage
  useEffect(() => {
    if (messages.length > 0) {
      const toSave = messages.slice(-MAX_RESPONSES * 2);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(toSave));
    }
  }, [messages]);

  // 自动滚动到底部
  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, loading]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = query.trim();
    if (!trimmed || loading) return;

    // 立即添加用户消息
    const userMsg: ChatMessage = {
      id: `user-${Date.now()}`,
      role: 'user',
      content: trimmed,
      timestamp: Date.now(),
    };
    setMessages(prev => [...prev, userMsg]);
    setQuery('');
    setLoading(true);

    try {
      const res = await api.agentQuery(trimmed) as AgentResponse;
      const aiMsg: ChatMessage = {
        id: `ai-${Date.now()}`,
        role: 'assistant',
        content: res.response || '',
        agent: res.agent,
        timestamp: Date.now(),
      };
      setMessages(prev => [...prev, aiMsg]);
    } catch (error) {
      const errorMsg: ChatMessage = {
        id: `error-${Date.now()}`,
        role: 'assistant',
        content: t('ai.queryFailed'),
        agent: 'system',
        timestamp: Date.now(),
      };
      setMessages(prev => [...prev, errorMsg]);
      showToast({ type: 'error', message: t('ai.queryFailed') });
    } finally {
      setLoading(false);
    }
  };

  const clearHistory = () => {
    setMessages([]);
    localStorage.removeItem(STORAGE_KEY);
  };

  return (
    <div className="space-y-6 h-[calc(100vh-120px)] flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.aiAgent')}</h1>
          <p className="text-slate-400">{t('ai.title')}</p>
        </div>
        <div className="flex items-center gap-3">
          {messages.length > 0 && (
            <button
              onClick={clearHistory}
              className="btn btn-secondary flex items-center gap-2 text-sm"
              aria-label="清除历史"
            >
              <Trash2 className="w-4 h-4" />
              <span>清除历史</span>
            </button>
          )}
          <div className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-primary-500" />
            <span className="text-sm text-slate-400">GLM-4-flash</span>
          </div>
        </div>
      </div>

      {/* Chat container */}
      <div className="card flex-1 flex flex-col min-h-0">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <Bot className="w-5 h-5 text-primary-500" />
            <span className="font-medium text-slate-100">{t('ai.chatTitle')}</span>
          </div>
        </div>
        <div className="card-body flex-1 overflow-y-auto scrollbar-thin space-y-4">
          {/* Welcome message */}
          {messages.length === 0 && !loading && (
            <div className="text-center py-8">
              <Bot className="w-12 h-12 text-primary-500 mx-auto mb-4" />
              <p className="text-slate-300 mb-2">{t('ai.welcome')}</p>
              <p className="text-sm text-slate-400">
                {t('ai.welcomeDescription')}
              </p>
            </div>
          )}

          {/* Messages */}
          {messages.map((msg) => (
            <div key={msg.id} className={`flex items-start gap-3 ${msg.role === 'user' ? 'justify-end' : ''}`}>
              {msg.role === 'assistant' ? (
                <>
                  <div className="w-8 h-8 rounded-full bg-primary-600 flex items-center justify-center flex-shrink-0">
                    <Bot className="w-4 h-4 text-white" />
                  </div>
                  <div className="bg-slate-800 border border-slate-700 rounded-lg px-4 py-3 max-w-[70%]">
                    {msg.agent && (
                      <div className={`text-sm mb-2 font-medium ${getAgentColor(msg.agent)}`}>
                        {msg.agent}
                      </div>
                    )}
                    <div className="text-slate-200 prose prose-invert prose-sm max-w-none
                      prose-headings:text-slate-100 prose-p:text-slate-200 prose-li:text-slate-200
                      prose-code:text-primary-300 prose-code:bg-slate-700 prose-code:px-1 prose-code:rounded
                      prose-pre:bg-slate-900 prose-pre:border prose-pre:border-slate-700">
                      <ReactMarkdown>{msg.content}</ReactMarkdown>
                    </div>
                  </div>
                </>
              ) : (
                <>
                  <div className="bg-primary-600 rounded-lg px-4 py-2 max-w-[70%]">
                    <p className="text-white">{msg.content}</p>
                  </div>
                  <div className="w-8 h-8 rounded-full bg-slate-600 flex items-center justify-center flex-shrink-0">
                    <User className="w-4 h-4 text-slate-300" />
                  </div>
                </>
              )}
            </div>
          ))}

          {/* Loading indicator */}
          {loading && (
            <div className="flex items-start gap-3">
              <div className="w-8 h-8 rounded-full bg-primary-600 flex items-center justify-center flex-shrink-0">
                <Bot className="w-4 h-4 text-white" />
              </div>
              <div className="bg-slate-800 border border-slate-700 rounded-lg px-4 py-3">
                <Loader className="w-5 h-5 animate-spin text-primary-400" />
              </div>
            </div>
          )}

          <div ref={chatEndRef} />
        </div>

        {/* Quick prompts + Input */}
        <div className="p-4 border-t border-slate-700 space-y-3">
          {/* 快捷建议按钮放在输入框紧邻上方，确保用户能看到填入效果 */}
          <div className="flex flex-wrap gap-2">
            {[
              t('ai.prompt1'),
              t('ai.prompt2'),
              t('ai.prompt3'),
              t('ai.prompt4'),
              t('ai.prompt5'),
            ].map((prompt) => (
              <button
                key={prompt}
                onClick={() => setQuery(prompt)}
                className="px-3 py-1 bg-slate-700 rounded-lg text-sm text-slate-300 hover:bg-slate-600 transition-colors"
                aria-label={prompt}
              >
                {prompt}
              </button>
            ))}
          </div>
          <form onSubmit={handleSubmit} className="flex gap-2">
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="input flex-1"
              placeholder={t('ai.placeholder')}
              disabled={loading}
            />
            <button
              type="submit"
              disabled={loading || !query.trim()}
              className="btn btn-primary flex items-center gap-2"
              aria-label={t('ai.askQuestion')}
            >
              {loading ? (
                <Loader className="w-5 h-5 animate-spin" />
              ) : (
                <Send className="w-5 h-5" />
              )}
              <span>{t('ai.askQuestion')}</span>
            </button>
          </form>
        </div>
      </div>

    </div>
  );
}
