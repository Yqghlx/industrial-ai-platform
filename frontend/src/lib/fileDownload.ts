/**
 * 文件下载工具函数
 * 将 DOM 操作隔离，便于测试和复用
 */

/** 下载二进制文件（Blob） */
export function downloadBlob(data: ArrayBuffer | BlobPart, filename: string, mimeType: string): void {
  const blob = new Blob([data], { type: mimeType });
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  window.URL.revokeObjectURL(url);
  document.body.removeChild(link);
}

/** 下载 CSV 文本（自动加 BOM 头保证中文兼容） */
export function downloadCSV(content: string, filename: string): void {
  downloadBlob('﻿' + content, filename, 'text/csv;charset=utf-8;');
}

/** 生成 CSV 字符串 */
export function generateCSV(headers: string[], rows: (string | number)[][]): string {
  return [
    headers.join(','),
    ...rows.map(r => r.join(',')),
  ].join('\n');
}
