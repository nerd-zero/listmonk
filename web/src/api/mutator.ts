import { userManager } from "@/auth/user-manager";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

// Custom fetch instance for the orval-generated client (see
// ../../orval.config.ts's override.mutator) -- attaches the current
// Zitadel access token as a Bearer header, matching internal/authn's
// verification on the Go side. No signup/login endpoints of our own to
// call here, per docs/plan.md's Auth section.
//
// Must resolve to { data, status, headers } (an axios-like envelope), not
// the parsed body directly -- orval's `httpClient: "fetch"` generator
// types every endpoint's return as exactly that shape (see any generated
// endpoints/*.ts's `...ResponseSuccess` type) and its own code accesses
// `.data`/`.status` uniformly. Our backend's own {"data": ...} envelope
// (see internal/httpapi's writeJSON) then lives one level down, at
// `.data.data` -- e.g. a hook's `query.data?.data.data` to reach the
// actual array/object.
export const customFetch = async <T,>(
  url: string,
  options: RequestInit,
): Promise<T> => {
  const user = await userManager.getUser();

  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");
  if (user?.access_token) {
    headers.set("Authorization", `Bearer ${user.access_token}`);
  }

  const response = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
  });

  const body = response.status === 204 ? undefined : await response.json();

  if (!response.ok) {
    throw new Error(body?.error ?? `Request failed: ${response.status}`);
  }

  return {
    data: body,
    status: response.status,
    headers: response.headers,
  } as T;
};

export default customFetch;
