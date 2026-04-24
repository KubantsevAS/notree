import type { IClient } from '../client';
import type { IInterceptor } from './interceptor.interface';
import type { IMiddleware } from './middleware.interface';

export interface IRequester extends IClient {
  create<C extends Record<string, unknown>, I extends IRequester>(config: C): I;

  addRequestInterceptor(interceptor: IInterceptor['onRequest']): string;

  addResponseInterceptor(interceptor: IInterceptor['onResponse']): string;

  addErrorInterceptor(
    interceptor: NonNullable<IInterceptor['onError']>,
  ): string;

  removeRequestInterceptor(id: string): void;
  removeResponseInterceptor(id: string): void;
  removeErrorInterceptor(id: string): void;

  use(middleware: IMiddleware): void;
  eject(middleware: IMiddleware): void;
}
