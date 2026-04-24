export interface IResult<TData = unknown, TError = Error, TVariables = void> {
  data: TData | undefined;
  isError: boolean;
  error: TError | null;

  mutate: (variables: TVariables) => void;
  mutateAsync: (variables: TVariables) => Promise<TData>;
  isPending: boolean;
}
