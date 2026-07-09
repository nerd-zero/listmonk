package httpapi

import (
	"encoding/json"
	"net/http"
	"regexp"

	"listnun/internal/provisioning"

	"github.com/google/uuid"
)

// Same rule the listmonk fork's Operator API enforces (cmd/operator.go's
// reTenantSlug) -- checked here too so a bad slug 400s before a round trip
// to that API or a DB constraint violation.
var reSlug = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// --- orgs -------------------------------------------------------------

func (a *API) listOrgs(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	orgs, err := a.svc.ListOrgsForUser(r.Context(), uuid.UUID(user.ID.Bytes))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing orgs")
		return
	}
	writeJSON(w, http.StatusOK, orgs)
}

type createOrgRequest struct {
	Name string `json:"name"`
}

func (a *API) createOrg(w http.ResponseWriter, r *http.Request) {
	var req createOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	user, _ := userFromContext(r.Context())
	org, err := a.svc.CreateOrg(r.Context(), uuid.UUID(user.ID.Bytes), req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "creating org")
		return
	}
	writeJSON(w, http.StatusOK, org)
}

// --- instances ----------------------------------------------------------

func (a *API) listInstances(w http.ResponseWriter, r *http.Request) {
	instances, err := a.svc.ListInstances(r.Context(), orgIDFromRequest(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing instances")
		return
	}
	writeJSON(w, http.StatusOK, instances)
}

type createInstanceRequest struct {
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	AdminUsername string `json:"admin_username"`
	AdminEmail    string `json:"admin_email"`
}

func (a *API) createInstance(w http.ResponseWriter, r *http.Request) {
	var req createInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !reSlug.MatchString(req.Slug) {
		writeError(w, http.StatusBadRequest, "invalid slug: use lowercase letters, numbers, hyphens")
		return
	}
	if req.Name == "" || req.AdminUsername == "" || req.AdminEmail == "" {
		writeError(w, http.StatusBadRequest, "name, admin_username, and admin_email are required")
		return
	}

	inst, err := a.svc.CreateInstance(r.Context(), orgIDFromRequest(r), provisioning.CreateInstanceParams{
		Slug:          req.Slug,
		Name:          req.Name,
		AdminUsername: req.AdminUsername,
		AdminEmail:    req.AdminEmail,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, inst)
}

func (a *API) getInstance(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	inst, err := a.svc.GetInstance(r.Context(), orgIDFromRequest(r), instanceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "instance not found")
		return
	}
	writeJSON(w, http.StatusOK, inst)
}

func (a *API) listEvents(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	// Confirms org ownership before listing this instance's events.
	if _, err := a.svc.GetInstance(r.Context(), orgIDFromRequest(r), instanceID); err != nil {
		writeError(w, http.StatusNotFound, "instance not found")
		return
	}

	events, err := a.svc.ListProvisioningEvents(r.Context(), instanceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing events")
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (a *API) resendSetupLink(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	url, err := a.svc.ResendSetupLink(r.Context(), orgIDFromRequest(r), instanceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reissuing setup link")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"setup_url": url})
}

// --- members --------------------------------------------------------------

type inviteMemberRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func (a *API) inviteMember(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	orgID := orgIDFromRequest(r)

	// Inviting is owner-only -- stricter than the plain-membership check
	// requireOrgMembership already ran for this whole {orgID} subtree.
	if err := a.svc.RequireOrgOwner(r.Context(), orgID, uuid.UUID(user.ID.Bytes)); err != nil {
		writeError(w, http.StatusForbidden, "only an org owner can invite members")
		return
	}

	var req inviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Role != "owner" && req.Role != "member" {
		req.Role = "member"
	}

	invited, err := a.svc.InviteMember(r.Context(), orgID, req.Email, req.DisplayName, req.Role)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, invited)
}
