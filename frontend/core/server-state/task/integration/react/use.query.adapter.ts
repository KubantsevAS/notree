import { useQuery } from '@tanstack/react-query';

import type { IQueryParams, IQueryResult } from '../../contract';

export function useQueryAdapter<TData, TError = Error>(
  params: IQueryParams<TData, TError>,
): IQueryResult<TData, TError> {
  const result = useQuery<TData, TError>({
    queryKey: params.queryKey,
    queryFn: params.executor,
    enabled: params.enabled ?? true,
    retry: params.retry,
    staleTime: params.staleTime,
  });

  return {
    data: result.data,
    isLoading: result.isPending && result.isFetching,
    isRefetching: !result.isPending && result.isFetching,
    isError: result.isError,
    error: result.error,
    refetch: result.refetch,
  };
}
