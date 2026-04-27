import { useQuery } from '@tanstack/react-query';

import type {
  IQueryException,
  IQueryParams,
  IQueryResult,
} from '../../contract';
import {
  normalizeGeneralException,
  QueryException,
} from '../../implementation';

export function useQueryAdapter<TData>(
  params: IQueryParams<TData>,
): IQueryResult<TData, IQueryException> {
  const result = useQuery<TData, IQueryException>({
    queryKey: params.queryKey,
    queryFn: async () => {
      try {
        return await params.executor();
      } catch (unknownRawError) {
        const { message, code } =
          normalizeGeneralException<'QUERY_EXECUTOR_ERROR'>(
            unknownRawError,
            'Unknown query error',
            'QUERY_EXECUTOR_ERROR',
          );

        throw new QueryException(message, code, params.queryKey);
      }
    },
    enabled: params.enabled ?? true,
    retry: params.retry,
    staleTime: params.staleTime,
  });

  return {
    data: result.data,
    isPending: result.isPending && result.isFetching,
    isRefetching: !result.isPending && result.isFetching,
    isError: result.isError,
    error: result.error,
    refetch: result.refetch,
  };
}
