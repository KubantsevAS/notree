export interface IActions {
  invalidateQueries: (key: unknown[]) => Promise<void> | void;
  removeQueries: (key?: unknown[]) => Promise<void> | void;
  resetQueries: (key?: unknown[]) => Promise<void> | void;
}
