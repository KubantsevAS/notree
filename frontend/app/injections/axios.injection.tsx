import axios from 'axios';

import {
  AxiosClientProvider,
  AxiosClientRegistry,
} from '../../core/integrations';

const REGISTRY_CLIENT_KEY = {
  DEFAULT: 'default',
} as const;

const defaultHttpClient = axios.create({
  baseURL: import.meta.env.VITE_DEFAULT_API_URL,
});

const axiosClientRegistry = new AxiosClientRegistry();

axiosClientRegistry.setDefaultClientKey(REGISTRY_CLIENT_KEY.DEFAULT);
axiosClientRegistry.createStorage();
axiosClientRegistry.registerClient(defaultHttpClient);

export function AxiosInjection({ children }: { children: React.ReactNode }) {
  return (
    <AxiosClientProvider registry={axiosClientRegistry}>
      {children}
    </AxiosClientProvider>
  );
}
