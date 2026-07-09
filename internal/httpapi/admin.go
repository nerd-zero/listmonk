package httpapi

import (
	"encoding/json"
	"net/http"
)

// adminInstanceDetail merges listnun's own instance row with the tenant's
// live status/counts from listmonk -- two different dimensions (see
// provisioning.AdminGetTenantLiveStatus's doc comment), both useful on one
// screen.
type adminInstanceDetail struct {
	Instance any `json:"instance"`
	Tenant   any `json:"listmonk_tenant,omitempty"`
}

// adminListOrgs godoc
//
//	@Summary		List every org on the platform
//	@Description	Super-admin only. Bypasses org membership.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	adminOrgListResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		403	{object}	errorResponse	"Not a super admin"
//	@Failure		500	{object}	errorResponse
//	@Router			/v1/admin/orgs [get]
func (a *API) adminListOrgs(w http.ResponseWriter, r *http.Request) {
	orgs, err := a.svc.AdminListOrgs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing orgs")
		return
	}
	writeJSON(w, http.StatusOK, orgs)
}

// adminListInstances godoc
//
//	@Summary		List every instance on the platform
//	@Description	Super-admin only. Bypasses org membership; includes each instance's org name.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	adminInstanceListResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		403	{object}	errorResponse	"Not a super admin"
//	@Failure		500	{object}	errorResponse
//	@Router			/v1/admin/instances [get]
func (a *API) adminListInstances(w http.ResponseWriter, r *http.Request) {
	instances, err := a.svc.AdminListInstances(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing instances")
		return
	}
	writeJSON(w, http.StatusOK, instances)
}

// adminGetInstance godoc
//
//	@Summary		Get instance detail (super admin)
//	@Description	Merges listnun's own provisioning status with the tenant's live status/counts straight from listmonk -- two different dimensions, see provisioning.AdminGetTenantLiveStatus's doc comment.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			instanceID	path		string	true	"Instance ID"
//	@Success		200			{object}	adminInstanceDetailResponse
//	@Failure		400			{object}	errorResponse
//	@Failure		401			{object}	errorResponse
//	@Failure		403			{object}	errorResponse	"Not a super admin"
//	@Failure		404			{object}	errorResponse
//	@Router			/v1/admin/instances/{instanceID} [get]
func (a *API) adminGetInstance(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	inst, err := a.svc.AdminGetInstance(r.Context(), instanceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "instance not found")
		return
	}

	detail := adminInstanceDetail{Instance: inst}
	if tenant, err := a.svc.AdminGetTenantLiveStatus(r.Context(), instanceID); err == nil {
		detail.Tenant = tenant
	}
	writeJSON(w, http.StatusOK, detail)
}

type adminSetStatusRequest struct {
	Status string `json:"status"`
}

// adminSetTenantStatus godoc
//
//	@Summary		Suspend, reactivate, or disable a tenant
//	@Description	Super-admin only. Writes directly to listmonk; deliberately not mirrored into instances.status (a different dimension -- provisioning state vs. account lifecycle).
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			instanceID	path		string					true	"Instance ID"
//	@Param			status		body		adminSetStatusRequest	true	"New status"
//	@Success		200			{object}	tenantResponse
//	@Failure		400			{object}	errorResponse
//	@Failure		401			{object}	errorResponse
//	@Failure		403			{object}	errorResponse	"Not a super admin"
//	@Failure		500			{object}	errorResponse
//	@Router			/v1/admin/instances/{instanceID}/status [put]
func (a *API) adminSetTenantStatus(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	var req adminSetStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Status != "active" && req.Status != "suspended" && req.Status != "disabled" {
		writeError(w, http.StatusBadRequest, "status must be one of active, suspended, disabled")
		return
	}

	tenant, err := a.svc.AdminSetTenantStatus(r.Context(), instanceID, req.Status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "updating tenant status")
		return
	}
	writeJSON(w, http.StatusOK, tenant)
}

// adminResendSetupLink godoc
//
//	@Summary		Reissue any instance's setup link (super admin)
//	@Description	Same as the org-scoped resend-setup-link endpoint, without the org-membership requirement.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			instanceID	path		string	true	"Instance ID"
//	@Success		200			{object}	setupLinkResponse
//	@Failure		400			{object}	errorResponse
//	@Failure		401			{object}	errorResponse
//	@Failure		403			{object}	errorResponse	"Not a super admin"
//	@Failure		500			{object}	errorResponse
//	@Router			/v1/admin/instances/{instanceID}/setup-link [post]
func (a *API) adminResendSetupLink(w http.ResponseWriter, r *http.Request) {
	instanceID, err := instanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid instance id")
		return
	}

	url, err := a.svc.AdminResendSetupLink(r.Context(), instanceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reissuing setup link")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"setup_url": url})
}
