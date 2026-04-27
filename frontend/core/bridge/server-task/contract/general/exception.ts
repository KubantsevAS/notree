export interface IException {
  message: string;
  code: string;
  timestamp: Date;
  details?: unknown;
}
