import type { IException as IClientException } from '../client/exception.interface';

export interface IException extends IClientException {
  mutationKey?: unknown[];
  variables?: unknown;
}
