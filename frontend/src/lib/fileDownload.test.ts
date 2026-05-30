import { describe, it, expect, vi, beforeEach } from 'vitest';
import { downloadBlob, downloadCSV, generateCSV } from './fileDownload';

// Mock DOM APIs
const mockCreateObjectURL = vi.fn(() => 'blob:mock-url');
const mockRevokeObjectURL = vi.fn();
const mockClick = vi.fn();
const mockAppendChild = vi.fn();
const mockRemoveChild = vi.fn();

beforeEach(() => {
  vi.stubGlobal('URL', {
    createObjectURL: mockCreateObjectURL,
    revokeObjectURL: mockRevokeObjectURL,
  });

  // Mock document.createElement to return a minimal anchor element
  vi.spyOn(document, 'createElement').mockReturnValue({
    href: '',
    download: '',
    click: mockClick,
  } as unknown as HTMLAnchorElement);

  vi.spyOn(document.body, 'appendChild').mockImplementation(mockAppendChild);
  vi.spyOn(document.body, 'removeChild').mockImplementation(mockRemoveChild);
});

describe('generateCSV', () => {
  it('生成标准 CSV 字符串', () => {
    const headers = ['姓名', '年龄', '城市'];
    const rows = [['张三', 25, '北京'], ['李四', 30, '上海']];
    const result = generateCSV(headers, rows);
    expect(result).toBe('姓名,年龄,城市\n张三,25,北京\n李四,30,上海');
  });

  it('空行数据时只返回表头', () => {
    const result = generateCSV(['A', 'B'], []);
    expect(result).toBe('A,B');
  });

  it('处理空表头和空行', () => {
    const result = generateCSV([], []);
    expect(result).toBe('');
  });

  it('数值类型正确转换为字符串', () => {
    const result = generateCSV(['值'], [[3.14], [0], [-1]]);
    expect(result).toBe('值\n3.14\n0\n-1');
  });
});

describe('downloadBlob', () => {
  it('创建 Blob 并触发下载', () => {
    const data = new Uint8Array([1, 2, 3]);
    downloadBlob(data, 'test.pdf', 'application/pdf');

    expect(mockCreateObjectURL).toHaveBeenCalled();
    expect(mockClick).toHaveBeenCalled();
    expect(mockRevokeObjectURL).toHaveBeenCalledWith('blob:mock-url');
    expect(mockAppendChild).toHaveBeenCalled();
    expect(mockRemoveChild).toHaveBeenCalled();
  });

  it('字符串数据也能正确下载', () => {
    downloadBlob('hello world', 'test.txt', 'text/plain');
    expect(mockClick).toHaveBeenCalled();
  });
});

describe('downloadCSV', () => {
  it('下载 CSV 内容并使用正确的 MIME 类型', () => {
    downloadCSV('a,b\n1,2', 'data.csv');
    expect(mockCreateObjectURL).toHaveBeenCalled();
    expect(mockClick).toHaveBeenCalled();
  });
});
