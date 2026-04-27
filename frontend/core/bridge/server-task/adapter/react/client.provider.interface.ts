import type { ReactNode } from 'react';

import type { IClientConfig } from '../../contract';

export interface IClientProviderProps {
  children: ReactNode;
  clientConfig?: IClientConfig;
}
