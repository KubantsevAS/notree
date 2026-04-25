import type { IException as IMutationException } from './exception.interface';

export interface IParams<
  TData = unknown,
  TVariables = void,
  TError = IMutationException,
> {
  mutationKey?: unknown[];
  executor: (variables: TVariables) => Promise<TData>;

  onSuccess?: (data: TData, variables: TVariables) => void;
  onError?: (error: TError, variables: TVariables) => void;
}
