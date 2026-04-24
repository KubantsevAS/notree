import { QueryClientProvider } from '@tanstack/react-query';
import React, { useMemo } from 'react';

import { ClientFactory } from './client.factory';
import type { IClientProviderProps } from './client.provider.interface';

export const ClientProvider: React.FC<IClientProviderProps> = ({
  children,
  clientConfig,
}) => {
  const queryClient = useMemo(() => {
    return ClientFactory.create(clientConfig);
  }, [clientConfig]);

  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};
