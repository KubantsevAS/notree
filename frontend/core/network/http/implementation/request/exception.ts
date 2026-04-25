import { Exception as ClientException } from '../client';

export class Exception extends ClientException {
  public readonly name: string;

  constructor(message: string, statusCode: number = 400, details?: unknown) {
    super(message, statusCode, details);

    this.name = 'NetworkHttpRequestException';

    Object.setPrototypeOf(this, Exception.prototype);
  }
}
