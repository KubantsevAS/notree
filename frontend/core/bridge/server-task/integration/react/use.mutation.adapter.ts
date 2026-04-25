import { useMutation } from '@tanstack/react-query';

import type { IMutationParams, IMutationResult } from '../../contract';
import { MutationException } from '../../implementation';

export function useMutationAdapter<TData, TVariables, TError>(
  params: IMutationParams<TData, TVariables, TError>,
): IMutationResult<TData, TVariables, TError> {
  const result = useMutation<TData, TError, TVariables>({
    mutationKey: params.mutationKey,
    mutationFn: async (variables: TVariables) => {
      try {
        return await params.executor(variables);
      } catch (unknowRawError) {
        const rawError = unknowRawError as Error & { code: string };

        throw new MutationException(
          rawError.message ?? 'Unknown mutation error',
          rawError.code ?? 'MUTATION_EXECUTOR_ERROR',
          params.mutationKey,
          variables,
        );
      }
    },
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
