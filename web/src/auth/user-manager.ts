import { UserManager } from "oidc-client-ts";
import { userManagerSettings } from "./oidc-config";

// A standalone UserManager reading the same persisted session
// react-oidc-context's AuthProvider writes to localStorage (see
// oidc-config.ts's userStore) -- lets the API layer (src/api/mutator.ts)
// attach the current access token outside the React tree, where
// react-query's queryFn/mutationFn run.
export const userManager = new UserManager(userManagerSettings);
