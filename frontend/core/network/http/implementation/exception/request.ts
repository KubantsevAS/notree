import { ClientException } from './client';

export class RequestException extends ClientException {
  public readonly name: string;

  constructor(message: string, statusCode: number = 400, details?: unknown) {
    super(message, statusCode, details);

    this.name = 'RequestException';

    Object.setPrototypeOf(this, RequestException.prototype);
  }
}
