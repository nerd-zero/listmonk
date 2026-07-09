import type { AuthProviderProps } from "react-oidc-context";
import { WebStorageStateStore, type UserManagerSettings } from "oidc-client-ts";

// Zitadel via Authorization Code + PKCE -- the frontend talks to Zitadel
// directly (no auth endpoints of our own on the Go backend, see
// docs/plan.md's Auth section). The access token this produces is sent as
// a Bearer header on every API call; internal/authn verifies it against
// the same issuer's JWKS.
//
// Typed as UserManagerSettings (not react-oidc-context's looser
// AuthProviderProps, which allows omitting these since it also accepts a
// pre-built UserManager) so user-manager.ts's standalone UserManager --
// built from this same object, to read the session outside the React
// tree -- gets real type-checking on the required fields.
export const userManagerSettings: UserManagerSettings = {
  authority: import.meta.env.VITE_OIDC_AUTHORITY,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  redirect_uri: import.meta.env.VITE_OIDC_REDIRECT_URI,
  post_logout_redirect_uri: import.meta.env.VITE_OIDC_POST_LOGOUT_REDIRECT_URI,
  scope: "openid profile email",
  userStore: new WebStorageStateStore({ store: window.localStorage }),
};

export const oidcConfig: AuthProviderProps = {
  ...userManagerSettings,
  onSigninCallback: () => {
    // Strip the ?code=&state= params Zitadel appended so a refresh doesn't
    // try to re-process a spent authorization code.
    window.history.replaceState({}, document.title, window.location.pathname);
  },
};
