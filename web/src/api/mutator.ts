import { userManager } from "@/auth/user-manager";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

// Custom fetch instance for the orval-generated client (see
// ../../orval.config.ts's override.mutator) -- attaches the current
// Zitadel access token as a Bearer header, matching internal/authn's
// verification on the Go side. No signup/login endpoints of our own to
// call here, per docs/plan.md's Auth section.
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

  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new Error(body?.error ?? `Request failed: ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }
  return response.json() as Promise<T>;
};

export default customFetch;
