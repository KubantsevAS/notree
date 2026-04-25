import type { IException as IQueryException } from './exception.interface';

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export interface IParams<TData = unknown, TError = IQueryException> {
  queryKey: unknown[];
  executor: () => Promise<TData>;

  enabled?: boolean;
  retry?: number;
  staleTime?: number;
}
