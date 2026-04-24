import type { TMethod } from './type';

export interface IRequest<TData = unknown> {
  url: string;
  method: TMethod;
  data?: TData;
  headers?: Record<string, string | number | boolean>;
  params?: Record<string, string | number | boolean>;
  rawConfig?: unknown;
}
