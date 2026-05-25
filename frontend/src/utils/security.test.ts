// security.ts 测试
// FIX-041-047: 前端安全工具测试

import { describe, it, expect } from 'vitest';
import {
  sanitizeHTML,
  isValidInput,
  safeSubstring,
  isValidURL,
  safeJSONParse,
  sanitizeErrorMessage,
  safeGet,
  hasRequiredProps,
} from '../utils/security';

describe('sanitizeHTML', () => {
  it('should escape HTML entities', () => {
    expect(sanitizeHTML('<script>alert("xss")</script>')).toBe(
      '&lt;script&gt;alert(&quot;xss&quot;)&lt;&#x2F;script&gt;'
    );
  });

  it('should escape ampersand', () => {
    expect(sanitizeHTML('test & value')).toBe('test &amp; value');
  });

  it('should handle empty string', () => {
    expect(sanitizeHTML('')).toBe('');
  });

  it('should handle null/undefined', () => {
    expect(sanitizeHTML(null as unknown as string)).toBe('');
    expect(sanitizeHTML(undefined as unknown as string)).toBe('');
  });

  it('should preserve safe text', () => {
    expect(sanitizeHTML('Hello World')).toBe('Hello World');
  });
});

describe('isValidInput', () => {
  it('should reject script tags', () => {
    expect(isValidInput('<script>alert(1)</script>')).toBe(false);
  });

  it('should reject javascript: URLs', () => {
    expect(isValidInput('javascript:alert(1)')).toBe(false);
  });

  it('should reject event handlers', () => {
    expect(isValidInput('onclick=alert(1)')).toBe(false);
  });

  it('should accept safe input', () => {
    expect(isValidInput('Hello World')).toBe(true);
  });

  it('should reject empty input', () => {
    expect(isValidInput('')).toBe(false);
  });

  it('should reject input exceeding maxLength', () => {
    const longInput = 'a'.repeat(1001);
    expect(isValidInput(longInput, 1000)).toBe(false);
  });

  it('should accept input within maxLength', () => {
    const validInput = 'a'.repeat(100);
    expect(isValidInput(validInput, 1000)).toBe(true);
  });
});

describe('safeSubstring', () => {
  it('should truncate long strings', () => {
    const result = safeSubstring('This is a long string', 10);
    expect(result).toBe('This is a ...');
    expect(result.length).toBe(13); // 10 chars + '...'
  });

  it('should preserve short strings', () => {
    expect(safeSubstring('Short', 10)).toBe('Short');
  });

  it('should handle empty string', () => {
    expect(safeSubstring('', 10)).toBe('');
  });

  it('should escape dangerous content', () => {
    const result = safeSubstring('<script>alert(1)</script>LongText', 8);
    expect(result).toContain('&lt;');
    expect(result).not.toContain('<script>');
  });
});

describe('isValidURL', () => {
  it('should accept http URLs', () => {
    expect(isValidURL('http://example.com')).toBe(true);
  });

  it('should accept https URLs', () => {
    expect(isValidURL('https://example.com/path')).toBe(true);
  });

  it('should reject javascript: URLs', () => {
    expect(isValidURL('javascript:alert(1)')).toBe(false);
  });

  it('should reject data: URLs', () => {
    expect(isValidURL('data:text/html,<script>')).toBe(false);
  });

  it('should reject invalid URLs', () => {
    expect(isValidURL('not-a-url')).toBe(false);
    expect(isValidURL('')).toBe(false);
  });

  it('should reject file: URLs', () => {
    expect(isValidURL('file:///etc/passwd')).toBe(false);
  });
});

describe('safeJSONParse', () => {
  it('should parse valid JSON', () => {
    const result = safeJSONParse<{ name: string }>('{"name":"test"}', { name: 'default' });
    expect(result.name).toBe('test');
  });

  it('should return fallback for invalid JSON', () => {
    const fallback = { name: 'default' };
    const result = safeJSONParse('invalid json', fallback);
    expect(result).toEqual(fallback);
  });

  it('should return fallback for empty string', () => {
    const fallback = { value: 0 };
    const result = safeJSONParse('', fallback);
    expect(result).toEqual(fallback);
  });

  it('should handle arrays', () => {
    const result = safeJSONParse<number[]>('[1,2,3]', []);
    expect(result).toEqual([1, 2, 3]);
  });
});

describe('sanitizeErrorMessage', () => {
  it('should handle Error objects', () => {
    const error = new Error('Test error message');
    const result = sanitizeErrorMessage(error);
    expect(result).toContain('Test error message');
  });

  it('should handle string errors', () => {
    const result = sanitizeErrorMessage('Simple error');
    expect(result).toBe('Simple error');
  });

  it('should truncate long messages', () => {
    const longError = 'a'.repeat(300);
    const result = sanitizeErrorMessage(longError);
    expect(result.length).toBeLessThanOrEqual(203); // 200 + '...'
  });

  it('should sanitize dangerous characters', () => {
    const result = sanitizeErrorMessage('Error <script>alert(1)</script>');
    expect(result).not.toContain('<script>');
  });

  it('should handle null/undefined', () => {
    expect(sanitizeErrorMessage(null)).toBe('Unknown error');
    expect(sanitizeErrorMessage(undefined)).toBe('Unknown error');
  });
});

describe('safeGet', () => {
  it('should access nested properties', () => {
    const obj = { user: { name: 'John', age: 30 } };
    expect(safeGet(obj, 'user.name', 'default')).toBe('John');
    expect(safeGet(obj, 'user.age', 0)).toBe(30);
  });

  it('should return fallback for missing properties', () => {
    const obj = { name: 'John' };
    expect(safeGet(obj, 'email', 'none')).toBe('none');
    expect(safeGet(obj, 'user.name', 'default')).toBe('default');
  });

  it('should handle null/undefined objects', () => {
    expect(safeGet(null, 'name', 'default')).toBe('default');
    expect(safeGet(undefined, 'name', 'default')).toBe('default');
  });

  it('should handle deep paths', () => {
    const obj = { a: { b: { c: { d: 'value' } } } };
    expect(safeGet(obj, 'a.b.c.d', 'fallback')).toBe('value');
    expect(safeGet(obj, 'a.b.c.e', 'fallback')).toBe('fallback');
  });
});

describe('hasRequiredProps', () => {
  it('should return true when all props exist', () => {
    expect(hasRequiredProps({ a: 1, b: 2, c: 3 }, ['a', 'b'])).toBe(true);
  });

  it('should return false when props missing', () => {
    expect(hasRequiredProps({ a: 1 }, ['a', 'b'])).toBe(false);
  });

  it('should handle empty object', () => {
    expect(hasRequiredProps({}, ['a'])).toBe(false);
  });

  it('should return true for empty props array', () => {
    expect(hasRequiredProps({}, [])).toBe(true);
  });

  it('should handle null/undefined objects', () => {
    expect(hasRequiredProps(null, ['a'])).toBe(false);
    expect(hasRequiredProps(undefined, ['a'])).toBe(false);
  });
});