import type { IException as IQueryException } from './exception.interface';

export interface IResult<TData = unknown, TError = IQueryException> {
  data: TData | undefined;
  isPending: boolean;
  isRefetching: boolean;

  isError: boolean;
  error: TError | null;

  refetch: () => Promise<TData | unknown>;
}
