import type { AxiosInstance } from 'axios';
import { type ReactNode, useMemo } from 'react';

import { AxiosContext } from './axios.client.context';
import { AxiosClientRegistry } from './axios.client.registry';
import type { AxiosClientsRegistry } from './axios.client.types';

type ClientProviderProps =
  | {
      client: AxiosInstance;
      clients?: never;
      registry?: never;
      children: ReactNode;
    }
  | {
      client?: never;
      clients: AxiosClientsRegistry;
      registry?: never;
      children: ReactNode;
    }
  | {
      client?: never;
      clients?: never;
      registry: AxiosClientRegistry;
      children: ReactNode;
    };

export function AxiosClientProvider({
  client,
  clients,
  registry: userRegistry,
  children,
}: ClientProviderProps) {
  const clientRegistry = useMemo(() => {
    if (userRegistry) {
      return userRegistry;
    }

    const registry = new AxiosClientRegistry();

    if (clients) {
      registry.setStorage(clients);

      return registry;
    }

    registry.registerClient(client, registry.getDefaultClientKey());

    return registry;
  }, [client, clients, userRegistry]);

  return (
    // eslint-disable-next-line react-x/no-context-provider
    <AxiosContext.Provider value={clientRegistry}>
      {children}
    </AxiosContext.Provider>
  );
}
