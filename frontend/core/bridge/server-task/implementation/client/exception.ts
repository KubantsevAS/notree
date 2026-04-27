import { Exception as GeneralException } from '../general';

export class Exception extends GeneralException {
  constructor(message: string, code: string, details?: unknown) {
    super('BridgeServerTaskClientException', message, code, details);

    Object.setPrototypeOf(this, Exception.prototype);
  }
}
