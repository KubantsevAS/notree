import { Exception as GeneralException } from '../general';

export class Exception extends GeneralException {
  public readonly mutationKey?: unknown[];
  public readonly variables?: unknown;

  constructor(
    message: string,
    code: string,
    mutationKey?: unknown[],
    variables?: unknown,
    details?: unknown,
  ) {
    super('BridgeServerTaskMutationException', message, code, details);

    this.mutationKey = mutationKey;
    this.variables = variables;

    Object.setPrototypeOf(this, Exception.prototype);
  }
}
