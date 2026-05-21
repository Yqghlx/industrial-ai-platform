declare module 'pako' {
  export function inflate(data: Uint8Array, options?: InflateOptions): Uint8Array;
  export function deflate(data: Uint8Array, options?: DeflateOptions): Uint8Array;
  
  export interface InflateOptions {
    level?: number;
    to?: string;
  }
  
  export interface DeflateOptions {
    level?: number;
  }
}