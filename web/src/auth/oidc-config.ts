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
  // oidc-client-ts defaults this to false -- without it, user.profile only
  // has whatever's directly embedded in the ID token, and Zitadel's ID
  // token is minimal (idTokenUserinfoAssertion is off), so email/name
  // never showed up. This makes it fetch /oidc/v1/userinfo after login.
  loadUserInfo: true,
  userStore: new WebStorageStateStore({ store: window.localStorage }),
};

export const oidcConfig: AuthProviderProps = {
  ...userManagerSettings,
  onSigninCallback: () => {
    // redirect_uri points at the dedicated /auth/callback path, which has
    // no matching <Route> in App.tsx -- land back on "/" instead of just
    // stripping ?code=&state= and leaving the browser stuck on a route
    // that renders nothing.
    window.history.replaceState({}, document.title, "/");
  },
};
