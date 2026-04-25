import { Exception as RequestException } from '../request';

export class Exception extends RequestException {
  public readonly name: string;

  constructor(message: string, statusCode: number = 0, details?: unknown) {
    super(message, statusCode, details);

    this.name = 'NetworkHttpRequesterException';

    Object.setPrototypeOf(this, Exception.prototype);
  }
}
