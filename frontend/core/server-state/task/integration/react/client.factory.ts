import { QueryClient } from '@tanstack/react-query';

import type { IClientConfig } from '../../contract';
import { ClientDefaultConfig } from '../../implementation';

export class ClientFactory {
  static create(config?: IClientConfig): QueryClient {
    const queryDefaults = config?.queries ?? {};
    const mutationDefaults = config?.mutations ?? {};

    return new QueryClient({
      defaultOptions: {
        queries: {
          staleTime:
            queryDefaults.staleTime ?? ClientDefaultConfig.queries.staleTime,
          retry: queryDefaults.retry ?? ClientDefaultConfig.queries.retry,
          refetchOnWindowFocus:
            queryDefaults.refetchOnWindowFocus ??
            ClientDefaultConfig.queries.refetchOnWindowFocus,
        },
        mutations: {
          retry: mutationDefaults.retry ?? ClientDefaultConfig.mutations.retry,
        },
      },
    });
  }
}
