/**
 * Unwraps an orval-generated response's `.data` field, asserting the
 * success variant. Safe because mutator.ts's customFetch throws on any
 * non-2xx response -- react-query routes those to `.error` instead, so a
 * resolved `.data` here is always the success shape, never the error one.
 */
export function unwrap<TSuccess>(response: {
  data: TSuccess | { error?: string };
}): TSuccess {
  return response.data as TSuccess;
}
