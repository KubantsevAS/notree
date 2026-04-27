import { Exception as GeneralException } from '../general';

export class Exception extends GeneralException {
  public readonly queryKey: unknown[];

  constructor(
    message: string,
    code: string,
    queryKey: unknown[],
    details?: unknown,
  ) {
    super('BridgeServerTaskQueryException', message, code, details);

    this.queryKey = queryKey;

    Object.setPrototypeOf(this, Exception.prototype);
  }
}
