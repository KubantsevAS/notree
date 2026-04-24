import React from 'react';

import { RequesterContext } from './requester.context';
import type { IRequesterProviderProps } from './requesterprovider.interface';

export const RequesterProvider: React.FC<IRequesterProviderProps> = ({
  children,
  requesterInstance,
}) => {
  return (
    <RequesterContext value={requesterInstance}>{children}</RequesterContext>
  );
};
