import { useContext } from 'react';

import { AxiosContext } from './axios.client.context';
import { type AxiosClientKey } from './axios.client.types';

export function useAxiosClient(clientKey?: AxiosClientKey) {
  // eslint-disable-next-line react-x/no-use-context
  const clientRegistry = useContext(AxiosContext);

  if (!clientRegistry) {
    throw new Error('useAxiosClient must be used within a ClientProvider');
  }

  const resolvedClientKey = clientKey ?? clientRegistry.getDefaultClientKey();
  const client = clientRegistry.getClient(resolvedClientKey);

  if (!client) {
    const availableClients = clientRegistry.getRegisteredKeys();

    throw new Error(
      `Axios client "${resolvedClientKey}" is not registered. Available clients: ${
        availableClients.length > 0 ? availableClients.join(', ') : 'none'
      }`,
    );
  }

  return client;
}
