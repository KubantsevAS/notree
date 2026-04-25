export interface IException extends Error {
  code: string;
  timestamp: Date;
}
