import type { TResponse } from '../client';
import type { IRequest } from '../request';

export type TDispatch = (request: IRequest) => Promise<TResponse>;
