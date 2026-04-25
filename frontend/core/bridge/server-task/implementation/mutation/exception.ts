import type { IMutationException } from '../../contract';
import { Exception as ClientException } from '../client';

export class Exception extends ClientException implements IMutationException {
  public readonly mutationKey?: unknown[];
  public readonly variables?: unknown;

  constructor(
    message: string,
    code: string,
    mutationKey?: unknown[],
    variables?: unknown,
  ) {
    super(message, code);
    this.name = 'BridgeServerTaskMutationException';
    this.mutationKey = mutationKey;
    this.variables = variables;

    Object.setPrototypeOf(this, Exception.prototype);
  }
}
