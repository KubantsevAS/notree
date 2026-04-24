import type { IException as IClientException, TResponse } from '../client';
import type { IRequest } from '../request';

export interface IInterceptor {
  onRequest(request: IRequest): IRequest | Promise<IRequest>;

  onResponse<T = unknown>(
    response: TResponse<T>,
  ): TResponse<T> | Promise<TResponse<T>>;

  onError?(
    error: IClientException,
  ): IClientException | Promise<IClientException>;
}
