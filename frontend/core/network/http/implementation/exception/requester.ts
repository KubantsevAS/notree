import { RequestException } from './request';

export class RequesterException extends RequestException {
  public readonly name: string;

  constructor(message: string, statusCode: number = 0, details?: unknown) {
    super(message, statusCode, details);

    this.name = 'RequesterException';

    Object.setPrototypeOf(this, RequesterException.prototype);
  }
}
