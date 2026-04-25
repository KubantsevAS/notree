import type { IClientConfig } from '../../contract';

export const DefaultConfig: Required<IClientConfig> = {
  queries: {
    staleTime: 1000 * 60 * 5,
    retry: 1,
    refetchOnWindowFocus: false,
  },
  mutations: {
    retry: false,
  },
};
