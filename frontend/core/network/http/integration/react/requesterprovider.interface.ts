import { type ReactNode } from 'react';

import type { IRequester } from '../../contract';

export interface IRequesterProviderProps {
  children: ReactNode;
  requesterInstance: IRequester;
}
