import { useMutation } from '@tanstack/react-query';

import type {
  IMutationException,
  IMutationParams,
  IMutationResult,
} from '../../contract';
import {
  MutationException,
  normalizeGeneralException,
} from '../../implementation';

export function useMutationAdapter<TData, TVariables>(
  params: IMutationParams<TData, TVariables, IMutationException>,
): IMutationResult<TData, TVariables, IMutationException> {
  const result = useMutation<TData, IMutationException, TVariables>({
    mutationKey: params.mutationKey,
    mutationFn: async (variables: TVariables) => {
      try {
        return await params.executor(variables);
      } catch (unknownRawError) {
        const { message, code: normalizedCode } =
          normalizeGeneralException<'MUTATION_EXECUTOR_ERROR'>(
            unknownRawError,
            'Unknown mutation error',
            'MUTATION_EXECUTOR_ERROR',
          );

        throw new MutationException(
          message,
          normalizedCode,
          params.mutationKey,
          variables,
        );
      }
    },
    onSuccess: async (data: TData, variables: TVariables) => {
      return params.onSuccess?.(data, variables);
    },
    onError: async (error: IMutationException, variables: TVariables) => {
      params.onError?.(error, variables);
    },
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
