import { Exception as GeneralException } from '../general';

export class Exception extends GeneralException {
  constructor(message: string, statusCode: number = 0, details?: unknown) {
    super('NetworkHttpRequestException', message, statusCode, details);

    Object.setPrototypeOf(this, Exception.prototype);
  }
}
