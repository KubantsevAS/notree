import type { IException as IMutationException } from './exception.interface';

export interface IResult<
  TData = unknown,
  TVariables = void,
  TError = IMutationException,
> {
  data: TData | undefined;
  isError: boolean;
  error: TError | null;

  mutate: (variables: TVariables) => void;
  mutateAsync: (variables: TVariables) => Promise<TData>;
  isPending: boolean;
}
