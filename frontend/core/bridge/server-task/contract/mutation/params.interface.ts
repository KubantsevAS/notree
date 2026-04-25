export interface IParams<TData = unknown, TVariables = void, TError = Error> {
  mutationKey?: unknown[];
  executor: (variables: TVariables) => Promise<TData>;

  onSuccess?: (data: TData, variables: TVariables) => void;
  onError?: (error: TError, variables: TVariables) => void;
}
