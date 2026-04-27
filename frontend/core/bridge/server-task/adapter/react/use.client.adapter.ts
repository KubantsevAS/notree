import { useQueryClient } from '@tanstack/react-query';

import type { IClientActions } from '../../contract';
import {
  ClientException,
  normalizeGeneralException,
} from '../../implementation';

export function useClientAdapter(): IClientActions {
  const queryClient = useQueryClient();

  const handleAction = async <TCode extends string>(
    action: () => Promise<void> | void,
    code: TCode,
  ) => {
    try {
      return await action();
    } catch (unknownRawError) {
      const { message, code: normalizedCode } = normalizeGeneralException<
        typeof code
      >(unknownRawError, 'Unknown client error', code);

      throw new ClientException(message, normalizedCode);
    }
  };

  return {
    invalidateQueries: (key) => {
      return handleAction<'CLIENT_INVALIDATE_ERROR'>(
        () => queryClient.invalidateQueries({ queryKey: key }),
        'CLIENT_INVALIDATE_ERROR',
      );
    },
    removeQueries: (key) => {
      return handleAction<'CLIENT_REMOVE_ERROR'>(
        () => queryClient.removeQueries(key ? { queryKey: key } : undefined),
        'CLIENT_REMOVE_ERROR',
      );
    },
    resetQueries: (key) => {
      return handleAction<'CLIENT_RESET_ERROR'>(
        () => queryClient.resetQueries(key ? { queryKey: key } : undefined),
        'CLIENT_RESET_ERROR',
      );
    },
  };
}
