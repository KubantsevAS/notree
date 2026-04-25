import { useQuery } from '@tanstack/react-query';

import type { IQueryParams, IQueryResult } from '../../contract';
import { QueryException } from '../../implementation';

export function useQueryAdapter<TData, TError>(
  params: IQueryParams<TData, TError>,
): IQueryResult<TData, TError> {
  const result = useQuery<TData, TError>({
    queryKey: params.queryKey,
    queryFn: async () => {
      try {
        return await params.executor();
      } catch (unknowRawError) {
        const rawError = unknowRawError as Error & { code: string };

        throw new QueryException(
          rawError.message ?? 'Unknown query error',
          rawError.code ?? 'QUERY_EXECUTOR_ERROR',
          params.queryKey,
        );
      }
    },
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
