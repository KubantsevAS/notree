export abstract class Exception extends Error {
  public readonly name: string;
  public readonly statusCode: number;
  public readonly details?: unknown;

  constructor(
    name: string,
    message: string,
    statusCode: number,
    details?: unknown,
  ) {
    super(message);

    this.name = name;
    this.statusCode = statusCode;
    this.details = details;

    Object.setPrototypeOf(this, new.target.prototype);
  }
}
