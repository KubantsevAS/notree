import { type TResponse } from './type';

export interface IClient {
  get<T = unknown>(url: string, config?: unknown): Promise<TResponse<T>>;

  post<T = unknown>(
    url: string,
    data?: unknown,
    config?: unknown,
  ): Promise<TResponse<T>>;

  patch<T = unknown>(
    url: string,
    data?: unknown,
    config?: unknown,
  ): Promise<TResponse<T>>;

  put<T = unknown>(
    url: string,
    data?: unknown,
    config?: unknown,
  ): Promise<TResponse<T>>;

  delete<T = unknown>(url: string, config?: unknown): Promise<TResponse<T>>;
}
