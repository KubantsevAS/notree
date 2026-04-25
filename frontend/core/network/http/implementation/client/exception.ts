export class Exception extends Error {
  public readonly name: string;
  public readonly statusCode: number;
  public readonly details?: unknown;

  constructor(message: string, statusCode: number = 0, details?: unknown) {
    super(message);

    this.name = 'NetworkHttpClientException';

    this.statusCode = statusCode;
    this.details = details;

    Object.setPrototypeOf(this, Exception.prototype);
  }
}
