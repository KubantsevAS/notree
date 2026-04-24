import axios, { type AxiosInstance } from 'axios';

import {
  type IClientException,
  type IClientResponse,
  type IRequest,
  type IRequester,
  type IRequesterInterceptor,
  type IRequesterMiddleware,
  type TRequesterDispatch,
  type TRequestMethod,
} from '../../contract';
import {
  ClientException,
  RequesterException,
  RequestFactory,
} from '../../implementation';
import { Client } from './client';

export class Requester implements IRequester {
  private readonly client: Client;
  private requestInterceptors: Map<string, IRequesterInterceptor['onRequest']> =
    new Map();
  private responseInterceptors: Map<
    string,
    IRequesterInterceptor['onResponse']
  > = new Map();
  private errorInterceptors: Map<
    string,
    NonNullable<IRequesterInterceptor['onError']>
  > = new Map();
  private middlewares: Set<IRequesterMiddleware> = new Set();
  private interceptorCounter = 0;

  constructor(private readonly instance: AxiosInstance) {
    this.client = new Client(instance);
  }

  private async executeWithPipeline<T>(
    request: IRequest,
  ): Promise<IClientResponse<T>> {
    let finalRequest: IRequest;

    try {
      finalRequest = await this.processRequestInterceptions(request);
    } catch (error) {
      return this.handlePipelineError(error);
    }

    try {
      const response = await this.findInitialDispatch<T>()(finalRequest);

      return this.processResponseInterceptions(response);
    } catch (error) {
      return this.handlePipelineError(error);
    }
  }

  private async processRequestInterceptions(req: IRequest) {
    let finalRequest = req;

    for (const interceptor of this.requestInterceptors.values()) {
      const prevResult = finalRequest;

      finalRequest = await interceptor(prevResult);
    }

    return finalRequest;
  }

  private findInitialDispatch<T>() {
    let initialDispatch: TRequesterDispatch = this.buildInitialRequest<T>();

    const middlewareArray = Array.from(this.middlewares);

    for (let i = middlewareArray.length - 1; i >= 0; i--) {
      const prevDispatch = initialDispatch;

      initialDispatch = (req) => middlewareArray[i](req, prevDispatch);
    }

    return initialDispatch;
  }

  private async processResponseInterceptions<T>(res: IClientResponse) {
    let finalResponse = res;

    for (const interceptor of this.responseInterceptors.values()) {
      const prevValue = finalResponse as IClientResponse<T>;

      finalResponse = await interceptor(prevValue);
    }

    return finalResponse as IClientResponse<T>;
  }

  private async handlePipelineError(error: unknown): Promise<never> {
    let clientError: IClientException;

    if (error instanceof ClientException) {
      clientError = error;
    } else {
      clientError = new ClientException(
        (error as Error).message || 'Unknown pipeline error',
        0,
        error,
      );
    }

    let processedError = clientError;
    for (const interceptor of this.errorInterceptors.values()) {
      processedError = await interceptor(processedError);
    }

    if (
      processedError === clientError ||
      processedError instanceof RequesterException
    ) {
      throw processedError;
    }

    throw new RequesterException(
      processedError.message,
      processedError.statusCode,
      processedError,
    );
  }

  private buildRequest(
    method: TRequestMethod,
    url: string,
    data?: unknown,
    config?: unknown,
  ): IRequest {
    const rawConfig = (config || {}) as Record<string, unknown>;

    return RequestFactory.create({
      url,
      method,
      data,
      headers: rawConfig.headers as IRequest['headers'],
      params: rawConfig.params as IRequest['params'],
      rawConfig,
    });
  }

  private buildInitialRequest<T>() {
    return async (req: IRequest) => {
      const axiosConfig = {
        url: req.url,
        method: req.method,
        data: req.data,
        headers: req.headers,
        params: req.params,
        ...((req.rawConfig as object) || {}),
      };

      return this.client.request<T>(req.method, req.url, axiosConfig);
    };
  }

  async get<T = unknown>(
    url: string,
    config?: unknown,
  ): Promise<IClientResponse<T>> {
    return this.executeWithPipeline<T>(
      this.buildRequest('get', url, undefined, config),
    );
  }

  async post<T = unknown>(
    url: string,
    data?: unknown,
    config?: unknown,
  ): Promise<IClientResponse<T>> {
    return this.executeWithPipeline<T>(
      this.buildRequest('post', url, data, config),
    );
  }

  async patch<T = unknown>(
    url: string,
    data?: unknown,
    config?: unknown,
  ): Promise<IClientResponse<T>> {
    return this.executeWithPipeline<T>(
      this.buildRequest('patch', url, data, config),
    );
  }

  async put<T = unknown>(
    url: string,
    data?: unknown,
    config?: unknown,
  ): Promise<IClientResponse<T>> {
    return this.executeWithPipeline<T>(
      this.buildRequest('put', url, data, config),
    );
  }

  async delete<T = unknown>(
    url: string,
    config?: unknown,
  ): Promise<IClientResponse<T>> {
    return this.executeWithPipeline<T>(
      this.buildRequest('delete', url, undefined, config),
    );
  }

  create<C extends Record<string, unknown>, I extends IRequester>(
    config: C,
  ): I {
    const currentDefaults = this.instance.defaults;
    const mergedConfig = {
      ...currentDefaults,
      ...config,
      headers: { ...currentDefaults.headers, ...(config.headers || {}) },
    };
    const newInstance = axios.create(mergedConfig);

    return new Requester(newInstance) as unknown as I;
  }

  addRequestInterceptor(
    interceptor: IRequesterInterceptor['onRequest'],
  ): string {
    const id = `req_${++this.interceptorCounter}`;
    this.requestInterceptors.set(id, interceptor);

    return id;
  }

  addResponseInterceptor(
    interceptor: IRequesterInterceptor['onResponse'],
  ): string {
    const id = `res_${++this.interceptorCounter}`;
    this.responseInterceptors.set(id, interceptor);

    return id;
  }

  addErrorInterceptor(
    interceptor: NonNullable<IRequesterInterceptor['onError']>,
  ): string {
    const id = `err_${++this.interceptorCounter}`;
    this.errorInterceptors.set(id, interceptor);

    return id;
  }

  removeRequestInterceptor(id: string): void {
    this.requestInterceptors.delete(id);
  }

  removeResponseInterceptor(id: string): void {
    this.responseInterceptors.delete(id);
  }

  removeErrorInterceptor(id: string): void {
    this.errorInterceptors.delete(id);
  }

  // eslint-disable-next-line react-x/no-unnecessary-use-prefix
  use(middleware: IRequesterMiddleware): void {
    this.middlewares.add(middleware);
  }

  eject(middleware: IRequesterMiddleware): void {
    this.middlewares.delete(middleware);
  }
}
