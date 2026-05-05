import { createContext } from 'react';

import type { AxiosClientRegistry } from './axios.client.registry';

export const AxiosContext = createContext<AxiosClientRegistry | null>(null);
