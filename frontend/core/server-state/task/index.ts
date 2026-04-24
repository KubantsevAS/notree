export type {
  IClientActions as ITaskClientActions,
  IClientConfig as ITaskClientConfig,
  IMutationParams as ITaskMutationParams,
  IMutationResult as ITaskMutationResult,
  IQueryParams as ITaskQueryParams,
  IQueryResult as ITaskQueryResult,
} from './contract';
export { ClientDefaultConfig as TaskClientDefaultConfig } from './implementation';
export type { IClientProviderProps as ITaskClientProviderProps } from './integration';
export {
  ClientProvider as TaskClientProvider,
  Devtools as TaskDevtools,
  useClientAdapter as useTaskClientAdapter,
  useMutationAdapter as useTaskMutationAdapter,
  useQueryAdapter as useTaskQueryAdapter,
} from './integration';
