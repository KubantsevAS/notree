import { useQueryClient } from '@tanstack/react-query';

import type { IClientActions } from '../../contract';
import { ClientException } from '../../implementation';

export function useClientAdapter(): IClientActions {
  const queryClient = useQueryClient();

  const handleAction = async (
    action: () => Promise<void> | void,
    code: string,
  ) => {
    try {
      return await action();
    } catch (unknowRawError) {
      const rawError = unknowRawError as Error & { code: string };

      throw new ClientException(
        rawError.message ?? 'Client action failed',
        code,
      );
    }
  };

  return {
    invalidateQueries: (key) => {
      return handleAction(
        () => queryClient.invalidateQueries({ queryKey: key }),
        'CLIENT_INVALIDATE_ERROR',
      );
    },
    removeQueries: (key) => {
      return handleAction(
        () => queryClient.removeQueries({ queryKey: key }),
        'CLIENT_REMOVE_ERROR',
      );
    },
    resetQueries: (key) => {
      return handleAction(
        () => queryClient.resetQueries({ queryKey: key }),
        'CLIENT_RESET_ERROR',
      );
    },
  };
}
