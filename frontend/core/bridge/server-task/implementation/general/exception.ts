export class Exception extends Error {
  public readonly code: string;
  public readonly details?: unknown;
  public readonly timestamp: Date;

  constructor(name: string, message: string, code: string, details?: unknown) {
    super(message);

    this.name = name;
    this.code = code;
    this.details = details;
    this.timestamp = new Date();

    Object.setPrototypeOf(this, new.target.prototype);
  }
}
