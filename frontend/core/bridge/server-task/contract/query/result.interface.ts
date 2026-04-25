import type { IException as IQueryException } from './exception.interface';

export interface IResult<TData = unknown, TError = IQueryException> {
  data: TData | undefined;
  isLoading: boolean;
  isRefetching: boolean;

  isError: boolean;
  error: TError | null;

  refetch: () => void;
}
