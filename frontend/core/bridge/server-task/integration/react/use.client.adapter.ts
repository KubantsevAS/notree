import { useQueryClient } from '@tanstack/react-query';

import type { IClientActions } from '../../contract';

export function useClientAdapter(): IClientActions {
  const queryClient = useQueryClient();

  return {
    invalidateQueries: async (key) => {
      await queryClient.invalidateQueries({ queryKey: key });
    },
    removeQueries: (key) => {
      queryClient.removeQueries({ queryKey: key });
    },
    resetQueries: async (key) => {
      await queryClient.resetQueries({ queryKey: key });
    },
  };
}
