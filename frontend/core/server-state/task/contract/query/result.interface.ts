export interface IResult<TData = unknown, TError = Error> {
  data: TData | undefined;
  isLoading: boolean;
  isRefetching: boolean;

  isError: boolean;
  error: TError | null;

  refetch: () => void;
}
