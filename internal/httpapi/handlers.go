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

// listOrgs godoc
//
//	@Summary	List the caller's orgs
//	@Tags		orgs
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	orgListResponse
//	@Failure	401	{object}	errorResponse
//	@Failure	500	{object}	errorResponse
//	@Router		/v1/orgs [get]
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

// createOrg godoc
//
//	@Summary	Create an additional org
//	@Description	Also creates the org's mirrored listmonk Organization (see internal/provisioning.createListmonkOrganization).
//	@Tags		orgs
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		org	body		createOrgRequest	true	"Org to create"
//	@Success	200	{object}	orgResponse
//	@Failure	400	{object}	errorResponse
//	@Failure	401	{object}	errorResponse
//	@Failure	500	{object}	errorResponse
//	@Router		/v1/orgs [post]
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

// listInstances godoc
//
//	@Summary	List an org's instances
//	@Tags		instances
//	@Produce	json
//	@Security	BearerAuth
//	@Param		orgID	path		string	true	"Org ID"
//	@Success	200		{object}	instanceListResponse
//	@Failure	401		{object}	errorResponse
//	@Failure	403		{object}	errorResponse
//	@Failure	500		{object}	errorResponse
//	@Router		/v1/orgs/{orgID}/instances [get]
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

// createInstance godoc
//
//	@Summary		Create an instance
//	@Description	Provisions a real tenant via the listmonk fork's Operator API, synchronously (see docs/plan.md's Provisioning state machine section).
//	@Tags			instances
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orgID		path		string					true	"Org ID"
//	@Param			instance	body		createInstanceRequest	true	"Instance to create"
//	@Success		200			{object}	instanceResponse
//	@Failure		400			{object}	errorResponse
//	@Failure		401			{object}	errorResponse
//	@Failure		403			{object}	errorResponse
//	@Failure		409			{object}	errorResponse	"Slug already in use"
//	@Failure		500			{object}	errorResponse
//	@Router			/v1/orgs/{orgID}/instances [post]
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

// getInstance godoc
//
//	@Summary	Get instance detail
//	@Tags		instances
//	@Produce	json
//	@Security	BearerAuth
//	@Param		orgID		path		string	true	"Org ID"
//	@Param		instanceID	path		string	true	"Instance ID"
//	@Success	200			{object}	instanceResponse
//	@Failure	400			{object}	errorResponse
//	@Failure	401			{object}	errorResponse
//	@Failure	404			{object}	errorResponse
//	@Router		/v1/orgs/{orgID}/instances/{instanceID} [get]
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

// listEvents godoc
//
//	@Summary		Provisioning timeline for an instance
//	@Description	Maps to provisioning_jobs; backs the dashboard's provisioning-status UI.
//	@Tags			instances
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orgID		path		string	true	"Org ID"
//	@Param			instanceID	path		string	true	"Instance ID"
//	@Success		200			{object}	provisioningJobListResponse
//	@Failure		400			{object}	errorResponse
//	@Failure		401			{object}	errorResponse
//	@Failure		404			{object}	errorResponse
//	@Router			/v1/orgs/{orgID}/instances/{instanceID}/events [get]
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

// resendSetupLink godoc
//
//	@Summary		Reissue an instance admin's one-time setup link
//	@Description	Needed because the original link's token is lost on a listmonk restart.
//	@Tags			instances
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orgID		path		string	true	"Org ID"
//	@Param			instanceID	path		string	true	"Instance ID"
//	@Success		200			{object}	setupLinkResponse
//	@Failure		400			{object}	errorResponse
//	@Failure		401			{object}	errorResponse
//	@Failure		500			{object}	errorResponse
//	@Router			/v1/orgs/{orgID}/instances/{instanceID}/setup-link [post]
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

// inviteMember godoc
//
//	@Summary		Invite a person into an org
//	@Description	Owner-only. Creates the person's Zitadel identity directly if a service account is configured (see internal/zitadelmgmt); otherwise returns 501.
//	@Tags			members
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orgID	path		string				true	"Org ID"
//	@Param			member	body		inviteMemberRequest	true	"Person to invite"
//	@Success		200		{object}	userResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		401		{object}	errorResponse
//	@Failure		403		{object}	errorResponse	"Not an org owner"
//	@Failure		501		{object}	errorResponse	"Invites not configured"
//	@Router			/v1/orgs/{orgID}/members [post]
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
