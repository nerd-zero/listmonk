package httpapi

import (
	"listnun/internal/db"
	"listnun/internal/operatorclient"
	"listnun/internal/provisioning"
)

// Response envelope types for swaggo -- writeJSON always wraps the actual
// payload as {"data": ...}, but swaggo needs a concrete named type per
// endpoint (not the `any` writeJSON takes at the call site) to generate an
// accurate schema. These exist purely for documentation generation
// (swag init) and the resulting orval-generated frontend types; nothing
// at runtime constructs or uses these structs directly.

type errorResponse struct {
	Error string `json:"error"`
} // @name ErrorResponse

type healthResponse struct {
	Data struct {
		Status string `json:"status"`
	} `json:"data"`
} // @name HealthResponse

type orgListResponse struct {
	Data []db.ListOrgsByUserRow `json:"data"`
} // @name OrgListResponse

type orgResponse struct {
	Data db.Org `json:"data"`
} // @name OrgResponse

type instanceListResponse struct {
	Data []db.Instance `json:"data"`
} // @name InstanceListResponse

type instanceResponse struct {
	Data db.Instance `json:"data"`
} // @name InstanceResponse

type provisioningJobListResponse struct {
	Data []db.ProvisioningJob `json:"data"`
} // @name ProvisioningJobListResponse

type setupLinkResponse struct {
	Data struct {
		SetupURL string `json:"setup_url"`
	} `json:"data"`
} // @name SetupLinkResponse

type userResponse struct {
	Data db.User `json:"data"`
} // @name UserResponse

type memberListResponse struct {
	Data []db.ListOrgMembersWithUserRow `json:"data"`
} // @name MemberListResponse

type adminOrgListResponse struct {
	Data []db.ListAllOrgsWithInstanceCountRow `json:"data"`
} // @name AdminOrgListResponse

type adminInstanceListResponse struct {
	Data []db.ListAllInstancesWithOrgNameRow `json:"data"`
} // @name AdminInstanceListResponse

type adminInstanceDetailResponse struct {
	Data adminInstanceDetail `json:"data"`
} // @name AdminInstanceDetailResponse

type tenantResponse struct {
	Data operatorclient.Tenant `json:"data"`
} // @name TenantResponse

// senderIdentityDetail bundles a sender identity with the DNS records to
// publish for it (empty for a sender_signature -- only a domain has any).
type senderIdentityDetail struct {
	Identity   db.SenderIdentity `json:"identity"`
	DNSRecords []db.DnsRecord    `json:"dns_records"`
} // @name SenderIdentityDetail

type senderIdentityResponse struct {
	Data senderIdentityDetail `json:"data"`
} // @name SenderIdentityResponse

type postmarkServerResponse struct {
	Data provisioning.PostmarkServerDetail `json:"data"`
} // @name PostmarkServerResponse
