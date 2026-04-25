import { useMutation } from '@tanstack/react-query';

import type { IMutationParams, IMutationResult } from '../../contract';

export function useMutationAdapter<TData, TVariables, TError = Error>(
  params: IMutationParams<TData, TVariables, TError>,
): IMutationResult<TData, TError, TVariables> {
  const result = useMutation<TData, TError, TVariables>({
    mutationKey: params.mutationKey,
    mutationFn: params.executor,
    onSuccess: params.onSuccess,
    onError: params.onError,
  });

  return {
    data: result.data,
    isError: result.isError,
    error: result.error,
    mutate: result.mutate,
    mutateAsync: result.mutateAsync,
    isPending: result.isPending,
  };
}
