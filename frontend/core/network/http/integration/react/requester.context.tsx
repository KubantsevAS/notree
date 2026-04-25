import { createContext, useContext } from 'react';

import type { IRequester } from '../../contract';

export const RequesterContext = createContext<IRequester | null>(null);

export const useRequester = (): IRequester => {
  const context = useContext(RequesterContext);

  if (!context) {
    throw new Error(
      'useHttpRequester must be used within a <HttpRequesterProvider />. Make sure to wrap your component tree with the Provider.',
    );
  }

  return context;
};
