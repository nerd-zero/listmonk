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

func (a *API) adminListOrgs(w http.ResponseWriter, r *http.Request) {
	orgs, err := a.svc.AdminListOrgs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing orgs")
		return
	}
	writeJSON(w, http.StatusOK, orgs)
}

func (a *API) adminListInstances(w http.ResponseWriter, r *http.Request) {
	instances, err := a.svc.AdminListInstances(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing instances")
		return
	}
	writeJSON(w, http.StatusOK, instances)
}

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
