package httpapi

import (
	"context"
	"errors"
	"net/http"

	"listnun/internal/authn"
	"listnun/internal/db"
	"listnun/internal/operatorclient"
	"listnun/internal/postmarkclient"
	"listnun/internal/provisioning"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ctxKey int

const userCtxKey ctxKey = iota

func withUser(ctx context.Context, u db.User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

func userFromContext(ctx context.Context) (db.User, bool) {
	u, ok := ctx.Value(userCtxKey).(db.User)
	return u, ok
}

// authMiddleware verifies the bearer token against Zitadel and
// JIT-provisions a users row (plus a personal org, on first sight) --
// there's no login/signup handler of our own, per docs/plan.md.
func (a *API) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := authn.BearerToken(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		claims, err := a.verifier.Verify(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid bearer token")
			return
		}

		user, err := a.svc.JITProvisionUser(r.Context(), claims.Subject, claims.Email, claims.DisplayName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "provisioning user")
			return
		}

		next.ServeHTTP(w, r.WithContext(withUser(r.Context(), user)))
	})
}

// requireOrgMembership enforces the "each org manages only its own
// tenants" boundary: the authenticated user must be a member of the
// {orgID} in the URL, or every route nested under it 404s/403s rather than
// trusting the path segment. Runs after authMiddleware.
func (a *API) requireOrgMembership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := userFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthenticated")
			return
		}

		orgID, err := uuid.Parse(chi.URLParam(r, "orgID"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid org id")
			return
		}

		if err := a.svc.RequireMembership(r.Context(), orgID, uuid.UUID(user.ID.Bytes)); err != nil {
			writeError(w, http.StatusForbidden, "not a member of this org")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requireSuperAdmin gates the whole /api/v1/admin group. Runs after
// authMiddleware, same pattern as requireOrgMembership.
func (a *API) requireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := userFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthenticated")
			return
		}

		if err := a.svc.RequireSuperAdmin(r.Context(), uuid.UUID(user.ID.Bytes)); err != nil {
			writeError(w, http.StatusForbidden, "not a super admin")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func orgIDFromRequest(r *http.Request) uuid.UUID {
	id, _ := uuid.Parse(chi.URLParam(r, "orgID"))
	return id
}

func instanceIDFromRequest(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "instanceID"))
}

// mapServiceError translates known provisioning errors to HTTP status
// codes; anything unrecognized is a 500.
func mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case err == provisioning.ErrSlugTaken:
		writeError(w, http.StatusConflict, err.Error())
	case err == provisioning.ErrNotMember:
		writeError(w, http.StatusForbidden, err.Error())
	case err == provisioning.ErrNotOwner:
		writeError(w, http.StatusForbidden, err.Error())
	case err == provisioning.ErrInvitesNotConfigured:
		writeError(w, http.StatusNotImplemented, err.Error())
	case err == provisioning.ErrSenderIdentityNotFound:
		writeError(w, http.StatusNotFound, err.Error())
	case err == provisioning.ErrSenderIdentityExists, err == provisioning.ErrSenderIdentityTaken:
		writeError(w, http.StatusConflict, err.Error())
	case err == provisioning.ErrPostmarkNotConfigured:
		writeError(w, http.StatusNotImplemented, err.Error())
	case err == provisioning.ErrInstanceHasNoPostmarkServer:
		writeError(w, http.StatusNotFound, err.Error())
	default:
		var pmErr *postmarkclient.APIError
		var opErr *operatorclient.APIError
		switch {
		case errors.As(err, &pmErr):
			writeError(w, http.StatusBadGateway, "postmark: "+pmErr.Message)
		case errors.As(err, &opErr):
			writeError(w, http.StatusBadGateway, "listmonk: "+opErr.Message)
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
	}
}
