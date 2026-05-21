// 安全工具函数
// FIX-041-047: 前端安全增强

/**
 * FIX-041: XSS防护 - HTML转义
 */
export function sanitizeHTML(str: string): string {
  if (!str) return '';
  const htmlEntities: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#x27;',
    '/': '&#x2F;',
  };
  return str.replace(/[&<>"'/]/g, (char) => htmlEntities[char] || char);
}

/**
 * FIX-042: 输入验证 - 防止注入
 */
export function isValidInput(input: string, maxLength: number = 1000): boolean {
  if (!input || input.length > maxLength) return false;
  // 检查危险字符序列
  const dangerousPatterns = [
    /<script/i,
    /javascript:/i,
    /on\w+=/i,
    /data:/i,
    /vbscript:/i,
  ];
  return !dangerousPatterns.some((pattern) => pattern.test(input));
}

/**
 * FIX-043: 安全截取 - 避免截断造成问题
 */
export function safeSubstring(str: string, maxLength: number): string {
  if (!str || str.length <= maxLength) return str || '';
  // 截取时避免截断HTML实体或危险内容
  const safeStr = sanitizeHTML(str.substring(0, maxLength));
  return safeStr + '...';
}

/**
 * FIX-044: URL验证 - 只允许安全URL
 */
export function isValidURL(url: string): boolean {
  if (!url) return false;
  try {
    const parsed = new URL(url);
    // 只允许http和https
    return parsed.protocol === 'http:' || parsed.protocol === 'https:';
  } catch {
    return false;
  }
}

/**
 * FIX-045: 安全JSON解析
 */
export function safeJSONParse<T>(jsonStr: string, fallback: T): T {
  if (!jsonStr) return fallback;
  try {
    const parsed = JSON.parse(jsonStr);
    // 检查解析后的对象是否安全
    if (typeof parsed === 'object' && parsed !== null) {
      return parsed as T;
    }
    return fallback;
  } catch {
    return fallback;
  }
}

/**
 * FIX-046: 安全存储 - 加密敏感数据
 */
export const secureStorage = {
  set(key: string, value: unknown, encrypt: boolean = false): void {
    try {
      const data = JSON.stringify(value);
      const stored = encrypt ? btoa(data) : data;
      localStorage.setItem(key, stored);
    } catch {
      console.warn('Storage set failed');
    }
  },

  get<T>(key: string, fallback: T, encrypted: boolean = false): T {
    try {
      const stored = localStorage.getItem(key);
      if (!stored) return fallback;
      const data = encrypted ? atob(stored) : stored;
      return safeJSONParse(data, fallback);
    } catch {
      return fallback;
    }
  },

  remove(key: string): void {
    localStorage.removeItem(key);
  },

  clear(): void {
    localStorage.clear();
  },
};

/**
 * FIX-047: 错误消息净化
 */
export function sanitizeErrorMessage(error: unknown): string {
  if (!error) return 'Unknown error';
  if (typeof error === 'string') {
    // 移除可能敏感的错误信息
    return safeSubstring(error.replace(/[^\w\s:.-]/g, ''), 200);
  }
  if (error instanceof Error) {
    return safeSubstring(error.message, 200);
  }
  return 'An error occurred';
}

/**
 * 安全的对象属性访问
 */
export function safeGet<T>(obj: unknown, path: string, fallback: T): T {
  if (!obj || typeof obj !== 'object') return fallback;
  const keys = path.split('.');
  let current: unknown = obj;
  for (const key of keys) {
    if (current && typeof current === 'object' && key in current) {
      current = (current as Record<string, unknown>)[key];
    } else {
      return fallback;
    }
  }
  return (current as T) ?? fallback;
}

/**
 * 检查对象是否有必要的属性
 */
export function hasRequiredProps(obj: unknown, props: string[]): boolean {
  if (!obj || typeof obj !== 'object') return false;
  return props.every((prop) => prop in (obj as object));
}