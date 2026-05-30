import React, { useState } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { useToast } from './Toast';
import { Bot, Send, User, Loader, Sparkles } from 'lucide-react';
import { AgentResponse } from '../types/api';
import { getAgentColor } from '../lib/colorUtils';

// Maximum number of responses to keep in memory
const MAX_RESPONSES = 50;

export default function AITeamDashboard() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [query, setQuery] = useState('');
  const [responses, setResponses] = useState<AgentResponse[]>([]);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    setLoading(true);
    try {
      const res = await api.agentQuery(query);
      setResponses(prev => {
        // Limit responses to prevent memory overflow
        const newResponses = [...prev, res as AgentResponse];
        if (newResponses.length > MAX_RESPONSES) {
          // Keep only the most recent responses
          return newResponses.slice(-MAX_RESPONSES);
        }
        return newResponses;
      });
      setQuery('');
    } catch (error) {
      showToast({ type: 'error', message: t('ai.queryFailed') });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6 h-[calc(100vh-120px)] flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.aiAgent')}</h1>
          <p className="text-slate-400">{t('ai.title')}</p>
        </div>
        <div className="flex items-center gap-2">
          <Sparkles className="w-5 h-5 text-primary-500" />
          <span className="text-sm text-slate-400">GLM-4-flash</span>
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
          {responses.length === 0 && (
            <div className="text-center py-8">
              <Bot className="w-12 h-12 text-primary-500 mx-auto mb-4" />
              <p className="text-slate-300 mb-2">{t('ai.welcome')}</p>
              <p className="text-sm text-slate-400">
                {t('ai.welcomeDescription')}
              </p>
            </div>
          )}

          {/* Messages */}
          {responses.map((r, index) => (
            // FE-P1-04: 使用 session_id 作为稳定 key，避免使用 slice 内容
            <div key={r.session_id || `response-${index}`} className="space-y-3">
              {/* User query */}
              <div className="flex items-start gap-3 justify-end">
                <div className="bg-primary-600 rounded-lg px-4 py-2 max-w-[70%]">
                  <p className="text-white">{r.session_id ? t('ai.queryLabel') : query}</p>
                </div>
                <div className="w-8 h-8 rounded-full bg-slate-600 flex items-center justify-center">
                  <User className="w-4 h-4 text-slate-300" />
                </div>
              </div>

              {/* AI response */}
              <div className="flex items-start gap-3">
                <div className="w-8 h-8 rounded-full bg-primary-600 flex items-center justify-center">
                  <Bot className="w-4 h-4 text-white" />
                </div>
                <div className="bg-slate-800 border border-slate-700 rounded-lg px-4 py-3 max-w-[70%]">
                  <div className={`text-sm mb-2 ${getAgentColor(r.agent)}`}>
                    {r.agent}
                  </div>
                  <div className="text-slate-200 whitespace-pre-wrap">
                    {r.response}
                  </div>
                </div>
              </div>
            </div>
          ))}
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