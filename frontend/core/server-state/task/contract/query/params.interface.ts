// eslint-disable-next-line @typescript-eslint/no-unused-vars
export interface IParams<TData = unknown, TError = Error> {
  queryKey: unknown[];
  executor: () => Promise<TData>;

  enabled?: boolean;
  retry?: number;
  staleTime?: number;
}
