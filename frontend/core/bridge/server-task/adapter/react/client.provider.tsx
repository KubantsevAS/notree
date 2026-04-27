import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React, { useRef } from 'react';

import { ClientFactory } from './client.factory';
import type { IClientProviderProps } from './client.provider.interface';

export const ClientProvider: React.FC<IClientProviderProps> = ({
  children,
  clientConfig,
}) => {
  const client = useRef<QueryClient>(ClientFactory.create(clientConfig));

  return (
    <QueryClientProvider client={client.current}>
      {children}
    </QueryClientProvider>
  );
};
