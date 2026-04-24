import type { IRequest, TRequestMethod } from '../../contract';

export class Request<TData = unknown> implements IRequest<TData> {
  public readonly url: string;
  public readonly method: TRequestMethod;
  public readonly data?: TData;
  public readonly headers?: Record<string, string | number | boolean>;
  public readonly params?: Record<string, string | number | boolean>;
  public readonly rawConfig?: unknown;

  constructor(
    url: string,
    method: TRequestMethod,
    data?: TData,
    headers?: Record<string, string | number | boolean>,
    params?: Record<string, string | number | boolean>,
    rawConfig?: unknown,
  ) {
    this.url = url;
    this.method = method;
    this.data = data;
    this.headers = headers;
    this.params = params;
    this.rawConfig = rawConfig;
  }

  public clone(partial: Partial<IRequest<TData>>): Request<TData> {
    return new Request<TData>(
      partial.url ?? this.url,
      partial.method ?? this.method,
      partial.data !== undefined ? (partial.data as TData) : this.data,
      partial.headers ? { ...this.headers, ...partial.headers } : this.headers,
      partial.params ? { ...this.params, ...partial.params } : this.params,
      partial.rawConfig
        ? { ...(this.rawConfig as object), ...(partial.rawConfig as object) }
        : this.rawConfig,
    );
  }
}
