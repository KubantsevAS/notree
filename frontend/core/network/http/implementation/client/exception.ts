import { Exception as GeneralException } from '../general';

export class Exception extends GeneralException {
  constructor(message: string, statusCode: number = 0, details?: unknown) {
    super('NetworkHttpClientException', message, statusCode, details);

    Object.setPrototypeOf(this, Exception.prototype);
  }
}
