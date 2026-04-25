import type { IQueryException } from '../../contract';
import { Exception as ClientException } from '../client';

export class Exception extends ClientException implements IQueryException {
  public readonly queryKey: unknown[];

  constructor(message: string, code: string, queryKey: unknown[]) {
    super(message, code);
    this.name = 'BridgeServerTaskQueryException';
    this.queryKey = queryKey;

    Object.setPrototypeOf(this, Exception.prototype);
  }
}
