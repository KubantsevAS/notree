export interface IException extends Error {
  message: string;
  statusCode: number;
  details?: unknown;
}
