import type { AxiosInstance } from 'axios';

export type AxiosClientKey = string & {};

export type AxiosClientsRegistry = Readonly<
  Partial<Record<AxiosClientKey, AxiosInstance>>
>;
export type KnownAxiosClientsRegistry<
  TClientKey extends string = AxiosClientKey,
> = Readonly<Partial<Record<TClientKey, AxiosInstance>>>;
