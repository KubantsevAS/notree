export const normalizeException = <TCode extends string>(
  rawError: unknown,
  messageFallback: string,
  codeFallback: TCode,
): {
  message: string;
  code: TCode;
} => {
  const isRawErrorInstanceofError = rawError instanceof Error;
  const isRawErrorHasCode =
    isRawErrorInstanceofError &&
    'code' in rawError &&
    typeof rawError.code === 'string';

  const normalizedMessage = isRawErrorInstanceofError
    ? (rawError.message ?? messageFallback)
    : String(rawError);
  const normalizedCode = isRawErrorHasCode
    ? (rawError.code as TCode)
    : codeFallback;

  return {
    message: normalizedMessage,
    code: normalizedCode,
  };
};
