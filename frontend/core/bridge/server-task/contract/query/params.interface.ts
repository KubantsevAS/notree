export interface IParams<TData = unknown> {
  queryKey: unknown[];
  executor: () => Promise<TData>;

  enabled?: boolean;
  retry?: number;
  staleTime?: number;
}
