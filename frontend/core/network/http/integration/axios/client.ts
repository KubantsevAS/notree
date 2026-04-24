import axios, { type AxiosError, type AxiosInstance } from 'axios';

import {
  type IClient,
  type IClientResponse,
  type TRequestMethod,
} from '../../contract';
import { ClientException } from '../../implementation';

export class Client implements IClient {
  constructor(private readonly instance: AxiosInstance) {
    this.instance = instance;
  }

  async request<T>(
    method: TRequestMethod,
    url: string,
    config?: unknown,
  ): Promise<IClientResponse<T>> {
    try {
      const axiosConfig = config as Record<string, unknown>;

      const response = await this.instance.request<T>({
        url,
        method,
        ...axiosConfig,
      });

      return {
        data: response.data,
        status: response.status,
      };
    } catch (error) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError;

        throw new ClientException(
          axiosError.message,
          axiosError.response?.status || 0,
          axiosError.response?.data || axiosError.config,
        );
      }

      throw new ClientException(
        (error as Error).message || 'Unknown HTTP error',
        0,
        error,
      );
    }
  }

  async get<T = unknown>(
    url: string,
    config?: unknown,
  ): Promise<IClientResponse<T>> {
    return this.request<T>('get', url, config);
  }

  async post<T = unknown>(
    url: string,
    data?: unknown,
    config?: unknown,
  ): Promise<IClientResponse<T>> {
    return this.request<T>('post', url, { ...(config as object), data });
  }

  async patch<T = unknown>(
    url: string,
    data?: unknown,
    config?: unknown,
  ): Promise<IClientResponse<T>> {
    return this.request<T>('patch', url, { ...(config as object), data });
  }

  async put<T = unknown>(
    url: string,
    data?: unknown,
    config?: unknown,
  ): Promise<IClientResponse<T>> {
    return this.request<T>('put', url, { ...(config as object), data });
  }

  async delete<T = unknown>(
    url: string,
    config?: unknown,
  ): Promise<IClientResponse<T>> {
    return this.request<T>('delete', url, config);
  }
}
