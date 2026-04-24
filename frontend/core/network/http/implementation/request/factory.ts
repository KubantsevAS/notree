import type { IRequest } from '../../contract';
import { Request } from './request';

export class Factory {
  static create<TData = unknown>(options: IRequest): Request<TData> {
    return new Request<TData>(
      options.url,
      options.method,
      options.data as TData,
      options.headers,
      options.params,
      options.rawConfig,
    );
  }
}
