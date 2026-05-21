import React, { createContext, useContext, useState, ReactNode, useCallback } from 'react';
import zhTranslations from './zh';
import enTranslations from './en';

type Language = 'zh' | 'en';
type Translations = typeof zhTranslations;

// 插值参数类型
type InterpolationParams = Record<string, string | number>;

interface I18nContextType {
  language: Language;
  setLanguage: (lang: Language) => void;
  t: (key: string, params?: InterpolationParams) => string;
  // 切换语言
  toggleLanguage: () => void;
  // 获取当前语言的显示名称
  languageName: string;
}

const translations: Record<Language, Translations> = {
  zh: zhTranslations,
  en: enTranslations,
};

const languageNames: Record<Language, string> = {
  zh: '中文',
  en: 'English',
};

const I18nContext = createContext<I18nContextType | undefined>(undefined);

export { I18nContext };

interface I18nProviderProps {
  children: ReactNode;
  defaultLanguage?: Language;
}

/**
 * 插值处理函数
 * 支持变量插值，格式为 {variable}
 * 例如: "已选择 {count} 条规则" => "已选择 5 条规则"
 */
function interpolate(template: string, params?: InterpolationParams): string {
  if (!params) {
    return template;
  }

  return template.replace(/\{(\w+)\}/g, (match, key) => {
    if (Object.prototype.hasOwnProperty.call(params, key)) {
      return String(params[key]);
    }
    return match; // 如果找不到参数，保留原始占位符
  });
}

export function I18nProvider({ children, defaultLanguage = 'zh' }: I18nProviderProps) {
  const [language, setLanguage] = useState<Language>(
    (localStorage.getItem('language') as Language) || defaultLanguage
  );

  const handleSetLanguage = useCallback((lang: Language) => {
    setLanguage(lang);
    localStorage.setItem('language', lang);
  }, []);

  const toggleLanguage = useCallback(() => {
    const newLang = language === 'zh' ? 'en' : 'zh';
    handleSetLanguage(newLang);
  }, [language, handleSetLanguage]);

  const t = useCallback((key: string, params?: InterpolationParams): string => {
    const keys = key.split('.');
    let value: unknown = translations[language];
    
    for (const k of keys) {
      if (typeof value === 'object' && value !== null) {
        value = (value as Record<string, unknown>)[k];
      } else {
        return key; // Return key if not found
      }
    }
    
    const result = typeof value === 'string' ? value : key;
    
    // 应用插值
    return interpolate(result, params);
  }, [language]);

  return (
    <I18nContext.Provider value={{ 
      language, 
      setLanguage: handleSetLanguage, 
      t,
      toggleLanguage,
      languageName: languageNames[language],
    }}>
      {children}
    </I18nContext.Provider>
  );
}

export function useI18n() {
  const context = useContext(I18nContext);
  if (!context) {
    throw new Error('useI18n must be used within an I18nProvider');
  }
  return context;
}

export default I18nProvider;