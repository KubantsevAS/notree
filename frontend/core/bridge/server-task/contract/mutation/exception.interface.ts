import type { IException as IGeneralException } from '../general/exception';

export interface IException extends IGeneralException {
  mutationKey?: unknown[];
  variables?: unknown;
}
