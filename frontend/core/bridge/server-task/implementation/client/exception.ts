import type { IClientException } from '../../contract';

export class Exception extends Error implements IClientException {
  public readonly code: string;
  public readonly timestamp: Date;

  constructor(message: string, code: string) {
    super(message);

    this.name = 'BridgeServerTaskClientException';

    this.code = code;
    this.timestamp = new Date();

    Object.setPrototypeOf(this, Exception.prototype);
  }
}
