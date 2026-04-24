import type { TResponse } from '../client';
import type { IRequest } from '../request';
import type { TDispatch } from './middleware.type';

export interface IMiddleware {
  (request: IRequest, next: TDispatch): Promise<TResponse>;
}
