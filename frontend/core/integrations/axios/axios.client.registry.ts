import type { AxiosInstance } from 'axios';

import type {
  AxiosClientKey,
  KnownAxiosClientsRegistry,
} from './axios.client.types';

export class AxiosClientRegistry {
  private defaultClientKey: AxiosClientKey;
  private storage: KnownAxiosClientsRegistry;

  constructor(defaultClientKey: AxiosClientKey = 'default') {
    this.defaultClientKey = defaultClientKey;
    this.storage = {};
  }

  setDefaultClientKey(key: AxiosClientKey) {
    this.defaultClientKey = key;
  }

  getDefaultClientKey() {
    return this.defaultClientKey;
  }

  createStorage() {
    this.storage = {};

    return this.storage;
  }

  setStorage<TRegistry extends KnownAxiosClientsRegistry>(registry: TRegistry) {
    this.storage = { ...registry };

    return this.storage as Readonly<TRegistry>;
  }

  getStorage() {
    return this.storage;
  }

  registerClient<TKey extends string>(
    client: AxiosInstance,
    key?: TKey,
  ): Readonly<
    Omit<KnownAxiosClientsRegistry, TKey> & Record<TKey, AxiosInstance>
  > {
    const nextStorage = {
      ...this.storage,
      [key ?? this.defaultClientKey]: client,
    } as Readonly<
      Omit<KnownAxiosClientsRegistry, TKey> & Record<TKey, AxiosInstance>
    >;
    this.storage = nextStorage;
    return nextStorage;
  }

  registerClients<TRegistry extends KnownAxiosClientsRegistry>(
    clients: TRegistry,
  ) {
    const nextStorage = {
      ...this.storage,
      ...clients,
    } as Readonly<KnownAxiosClientsRegistry & TRegistry>;
    this.storage = nextStorage;
    return nextStorage;
  }

  getClient<TKey extends string>(key: TKey): AxiosInstance | undefined {
    return this.storage[key];
  }

  hasClient<TKey extends string>(key: TKey) {
    return Boolean(this.storage[key]);
  }

  getRegisteredKeys<TKey extends string = AxiosClientKey>() {
    return Object.keys(this.storage) as TKey[];
  }
}
