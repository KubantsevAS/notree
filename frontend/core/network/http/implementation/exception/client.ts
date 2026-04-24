export class ClientException extends Error {
  public readonly name: string;
  public readonly statusCode: number;
  public readonly details?: unknown;

  constructor(message: string, statusCode: number = 0, details?: unknown) {
    super(message);

    this.name = 'HttpClientException';

    this.statusCode = statusCode;
    this.details = details;

    Object.setPrototypeOf(this, ClientException.prototype);
  }
}
