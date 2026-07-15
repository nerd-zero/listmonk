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

// getMe godoc
//
//	@Summary	The caller's own user record
//	@Description	Includes is_super_admin -- the only client-side way to know whether to show admin-only UI (there's no separate permissions endpoint; this is the single source of truth the frontend's permissions.tsx reads from).
//	@Tags		users
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	userResponse
//	@Failure	401	{object}	errorResponse
//	@Router		/v1/me [get]
func (a *API) getMe(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	writeJSON(w, http.StatusOK, user)
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

// getSenderIdentity godoc
//
//	@Summary		Get an instance's sender identity
//	@Description	Returns the domain or sender signature the org added, plus any DNS records to publish for it (empty for a sender signature). 404 if none added yet.
//	@Tags			instances
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orgID		path		string	true	"Org ID"
//	@Param			instanceID	path		string	true	"Instance ID"
//	@Success		200			{object}	senderIdentityResponse
//	@Failure		400			{object}	errorResponse
//	@Failure		401			{object}	errorResponse
//	@Failure		404			{object}	errorResponse	"No sender identity yet"
//	@Router			/v1/orgs/{orgID}/instances/{instanceID}/sender-identity [get]
func (a *API) getSenderIdentity(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	identity, records, err := a.svc.GetSenderIdentity(r.Context(), orgIDFromRequest(r), instanceID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, senderIdentityDetail{Identity: identity, DNSRecords: records})
}

type addSenderIdentityRequest struct {
	// Kind selects which Postmark identity to add: "domain" (an org's own
	// sending domain, needs DKIM published), "sender_signature" (single
	// address, confirmed by clicking Postmark's own email -- no DNS), or
	// "platform_domain" (a subdomain of ours, for an org with no domain of
	// their own -- no value needed, it's derived from the instance's slug).
	Kind string `json:"kind"`
	// Value is the domain name for kind "domain", or the From email
	// address for kind "sender_signature". Unused for "platform_domain".
	Value string `json:"value"`
	// Name is only used for kind "sender_signature" -- the display name
	// Postmark shows alongside that address.
	Name string `json:"name"`
}

// addSenderIdentity godoc
//
//	@Summary		Add an instance's sender identity
//	@Description	Adds exactly one sender identity per instance (an org's own domain, a sender signature, or an opt-in subdomain of ours), and pushes the resulting SMTP credentials into the listmonk tenant. 409 if the instance already has one, or if the domain/email is already claimed by another workspace.
//	@Tags			instances
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orgID		path		string					true	"Org ID"
//	@Param			instanceID	path		string					true	"Instance ID"
//	@Param			identity	body		addSenderIdentityRequest	true	"Sender identity to add"
//	@Success		200			{object}	senderIdentityResponse
//	@Failure		400			{object}	errorResponse
//	@Failure		401			{object}	errorResponse
//	@Failure		409			{object}	errorResponse	"Already added, or already claimed by another workspace"
//	@Failure		501			{object}	errorResponse	"Postmark not configured"
//	@Router			/v1/orgs/{orgID}/instances/{instanceID}/sender-identity [post]
func (a *API) addSenderIdentity(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	var req addSenderIdentityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	orgID := orgIDFromRequest(r)
	switch req.Kind {
	case "domain":
		if req.Value == "" {
			writeError(w, http.StatusBadRequest, "value is required")
			return
		}
		identity, records, err := a.svc.AddSenderDomain(r.Context(), orgID, instanceID, req.Value)
		if err != nil {
			mapServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, senderIdentityDetail{Identity: identity, DNSRecords: records})
	case "sender_signature":
		if req.Value == "" {
			writeError(w, http.StatusBadRequest, "value is required")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		identity, err := a.svc.AddSenderSignature(r.Context(), orgID, instanceID, req.Value, req.Name)
		if err != nil {
			mapServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, senderIdentityDetail{Identity: identity})
	case "platform_domain":
		identity, records, err := a.svc.AddPlatformDomain(r.Context(), orgID, instanceID)
		if err != nil {
			mapServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, senderIdentityDetail{Identity: identity, DNSRecords: records})
	default:
		writeError(w, http.StatusBadRequest, `kind must be "domain", "sender_signature", or "platform_domain"`)
	}
}

// deleteSenderIdentity godoc
//
//	@Summary		Delete an instance's sender identity
//	@Description	Removes the domain or sender signature from Postmark and locally, along with any DNS records published for it. Irreversible. The instance is left without a confirmed "from" address until a new identity is added.
//	@Tags			instances
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orgID		path	string	true	"Org ID"
//	@Param			instanceID	path	string	true	"Instance ID"
//	@Success		200
//	@Failure		400	{object}	errorResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		404	{object}	errorResponse	"No sender identity yet"
//	@Failure		501	{object}	errorResponse	"Postmark not configured"
//	@Router			/v1/orgs/{orgID}/instances/{instanceID}/sender-identity [delete]
func (a *API) deleteSenderIdentity(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	if err := a.svc.DeleteSenderIdentity(r.Context(), orgIDFromRequest(r), instanceID); err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// deletePostmarkServer godoc
//
//	@Summary		Delete an instance's Postmark server
//	@Description	Removes the Postmark server without touching the instance/tenant itself -- the instance is left without email sending until re-provisioned. Irreversible.
//	@Tags			instances
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orgID		path	string	true	"Org ID"
//	@Param			instanceID	path	string	true	"Instance ID"
//	@Success		200
//	@Failure		400	{object}	errorResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		404	{object}	errorResponse	"Instance has no Postmark server"
//	@Failure		501	{object}	errorResponse	"Postmark not configured"
//	@Router			/v1/orgs/{orgID}/instances/{instanceID}/postmark-server [delete]
func (a *API) deletePostmarkServer(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	if err := a.svc.DeletePostmarkServer(r.Context(), orgIDFromRequest(r), instanceID); err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// getPostmarkServer godoc
//
//	@Summary		Get an instance's Postmark server
//	@Description	Locally-stored info plus live state from Postmark (name, whether SMTP sending is activated). Never includes the API token.
//	@Tags			instances
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orgID		path		string	true	"Org ID"
//	@Param			instanceID	path		string	true	"Instance ID"
//	@Success		200			{object}	postmarkServerResponse
//	@Failure		400			{object}	errorResponse
//	@Failure		401			{object}	errorResponse
//	@Failure		404			{object}	errorResponse	"Instance has no Postmark server"
//	@Failure		501			{object}	errorResponse	"Postmark not configured"
//	@Router			/v1/orgs/{orgID}/instances/{instanceID}/postmark-server [get]
func (a *API) getPostmarkServer(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	server, err := a.svc.GetPostmarkServer(r.Context(), orgIDFromRequest(r), instanceID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, server)
}

// resyncPostmarkServer godoc
//
//	@Summary		Re-push an instance's Postmark SMTP credentials into listmonk
//	@Description	Fixes drift if the tenant's SMTP config was reset or changed by hand. Requires both a Postmark server and a confirmed sender identity to already exist.
//	@Tags			instances
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orgID		path	string	true	"Org ID"
//	@Param			instanceID	path	string	true	"Instance ID"
//	@Success		200
//	@Failure		400	{object}	errorResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		404	{object}	errorResponse	"No sender identity yet"
//	@Failure		501	{object}	errorResponse	"Postmark not configured"
//	@Router			/v1/orgs/{orgID}/instances/{instanceID}/postmark-server/resync [post]
func (a *API) resyncPostmarkServer(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	if err := a.svc.ResyncPostmarkServer(r.Context(), orgIDFromRequest(r), instanceID); err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"resynced": true})
}

// --- members --------------------------------------------------------------

// listMembers godoc
//
//	@Summary	List an org's members
//	@Tags		members
//	@Produce	json
//	@Security	BearerAuth
//	@Param		orgID	path		string	true	"Org ID"
//	@Success	200		{object}	memberListResponse
//	@Failure	401		{object}	errorResponse
//	@Failure	403		{object}	errorResponse
//	@Failure	500		{object}	errorResponse
//	@Router		/v1/orgs/{orgID}/members [get]
func (a *API) listMembers(w http.ResponseWriter, r *http.Request) {
	members, err := a.svc.ListMembers(r.Context(), orgIDFromRequest(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing members")
		return
	}
	writeJSON(w, http.StatusOK, members)
}

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
