import type { AuthProviderProps } from "react-oidc-context";
import { WebStorageStateStore } from "oidc-client-ts";

// Zitadel via Authorization Code + PKCE -- the frontend talks to Zitadel
// directly (no auth endpoints of our own on the Go backend, see
// docs/plan.md's Auth section). The access token this produces is sent as
// a Bearer header on every API call; internal/authn verifies it against
// the same issuer's JWKS.
export const oidcConfig: AuthProviderProps = {
  authority: import.meta.env.VITE_OIDC_AUTHORITY,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  redirect_uri: import.meta.env.VITE_OIDC_REDIRECT_URI,
  post_logout_redirect_uri: import.meta.env.VITE_OIDC_POST_LOGOUT_REDIRECT_URI,
  scope: "openid profile email",
  userStore: new WebStorageStateStore({ store: window.localStorage }),
  onSigninCallback: () => {
    // Strip the ?code=&state= params Zitadel appended so a refresh doesn't
    // try to re-process a spent authorization code.
    window.history.replaceState({}, document.title, window.location.pathname);
  },
};
