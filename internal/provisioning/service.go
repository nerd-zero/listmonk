// Package provisioning holds the framework-agnostic business logic for
// onboarding tenants: creating orgs, JIT-provisioning users into their own
// personal org on first login, and turning "create an instance" into a real
// tenant in the listmonk fork via internal/operatorclient. No HTTP here by
// design, so it's directly exercisable from a test/smoke harness without a
// router or auth middleware in the way.
package provisioning

import (
	"context"
	"errors"
	"fmt"

	"listnun/internal/db"
	"listnun/internal/operatorclient"
	"listnun/internal/zitadelmgmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrSlugTaken is returned by CreateInstance when the listmonk fork already
// has a tenant with the requested slug (its own 409 response).
var ErrSlugTaken = errors.New("a workspace with this slug already exists")

// ErrNotMember is returned by RequireMembership when the user isn't a
// member of the org -- the boundary that keeps one org's tenants from
// being reachable by another org's users.
var ErrNotMember = errors.New("not a member of this org")

// ErrNotOwner is returned by RequireOrgOwner -- inviting new members is an
// owner-only action, unlike most other org-scoped reads/writes which any
// member can do.
var ErrNotOwner = errors.New("only an org owner can do this")

// ErrInvitesNotConfigured is returned by InviteMember when no Zitadel
// service account is configured (see internal/zitadelmgmt). Inviting users
// is optional scaffolding, unlike the listmonk operator client the API
// can't run without.
var ErrInvitesNotConfigured = errors.New("invites are not configured")

type Service struct {
	pool *pgxpool.Pool
	q    *db.Queries
	op   *operatorclient.Client
	zm   *zitadelmgmt.Client // nil unless a Zitadel service account is configured
}

func New(pool *pgxpool.Pool, op *operatorclient.Client, zm *zitadelmgmt.Client) *Service {
	return &Service{pool: pool, q: db.New(pool), op: op, zm: zm}
}

// JITProvisionUser looks up a user by their verified Zitadel subject,
// creating the user plus a personal org (and owner membership) on first
// sight. Returns the user and whether this was a first-time provision.
func (s *Service) JITProvisionUser(ctx context.Context, zitadelSubject, email, displayName string) (db.User, error) {
	existing, err := s.q.GetUserByZitadelSubject(ctx, zitadelSubject)
	if err == nil {
		return existing, nil
	}

	orgID := uuid.New()
	lmOrgID, err := s.createListmonkOrganization(ctx, orgID, personalOrgName(displayName, email))
	if err != nil {
		return db.User{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)

	user, err := q.CreateUser(ctx, db.CreateUserParams{
		ID:             pgUUID(uuid.New()),
		ZitadelSubject: zitadelSubject,
		Email:          email,
		DisplayName:    nullableString(displayName),
	})
	if err != nil {
		return db.User{}, fmt.Errorf("create user: %w", err)
	}

	org, err := q.CreateOrg(ctx, db.CreateOrgParams{
		ID:                     pgUUID(orgID),
		Name:                   personalOrgName(displayName, email),
		ListmonkOrganizationID: &lmOrgID,
	})
	if err != nil {
		return db.User{}, fmt.Errorf("create personal org: %w", err)
	}

	if _, err := q.AddOrgMember(ctx, db.AddOrgMemberParams{
		OrgID:  org.ID,
		UserID: user.ID,
		Role:   "owner",
	}); err != nil {
		return db.User{}, fmt.Errorf("add owner membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.User{}, fmt.Errorf("commit tx: %w", err)
	}
	return user, nil
}

func personalOrgName(displayName, email string) string {
	if displayName != "" {
		return displayName + "'s org"
	}
	return email + "'s org"
}

// createListmonkOrganization creates the org's twin in the listmonk fork
// first, so a local org row is never written without one -- nothing to
// reconcile if this succeeds but the local insert that follows fails,
// since orgs.listmonk_organization_id is nullable and the caller simply
// doesn't get a local row at all in that case.
//
// The listmonk-side name is disambiguated with the local org's own id
// because personal-org names ("Alex's org") collide often across
// different users, and the fork enforces organizations_name_key globally.
// This name is never user-facing -- only listnun's own orgs.name is ever
// displayed, and Organization itself is Operator-API-only by design.
func (s *Service) createListmonkOrganization(ctx context.Context, localOrgID uuid.UUID, name string) (int32, error) {
	lmName := name + " (" + localOrgID.String()[:8] + ")"
	lmOrg, err := s.op.CreateOrganization(ctx, lmName)
	if err != nil {
		return 0, fmt.Errorf("create listmonk organization: %w", err)
	}
	return int32(lmOrg.ID), nil
}

// CreateOrg creates an additional org for a user who already has one,
// e.g. to separate a second company/brand's tenants from the first.
func (s *Service) CreateOrg(ctx context.Context, ownerUserID uuid.UUID, name string) (db.Org, error) {
	orgID := uuid.New()
	lmOrgID, err := s.createListmonkOrganization(ctx, orgID, name)
	if err != nil {
		return db.Org{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Org{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)

	org, err := q.CreateOrg(ctx, db.CreateOrgParams{ID: pgUUID(orgID), Name: name, ListmonkOrganizationID: &lmOrgID})
	if err != nil {
		return db.Org{}, fmt.Errorf("create org: %w", err)
	}
	if _, err := q.AddOrgMember(ctx, db.AddOrgMemberParams{
		OrgID:  org.ID,
		UserID: pgUUID(ownerUserID),
		Role:   "owner",
	}); err != nil {
		return db.Org{}, fmt.Errorf("add owner membership: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return db.Org{}, fmt.Errorf("commit tx: %w", err)
	}
	return org, nil
}

func (s *Service) ListOrgsForUser(ctx context.Context, userID uuid.UUID) ([]db.Org, error) {
	return s.q.ListOrgsByUser(ctx, pgUUID(userID))
}

// RequireMembership fails closed: any error from the membership lookup
// (no such row, or a real DB problem) is reported the same way, since both
// mean "can't confirm this user belongs to this org."
func (s *Service) RequireMembership(ctx context.Context, orgID, userID uuid.UUID) error {
	if _, err := s.q.GetOrgMember(ctx, db.GetOrgMemberParams{OrgID: pgUUID(orgID), UserID: pgUUID(userID)}); err != nil {
		return ErrNotMember
	}
	return nil
}

// RequireOrgOwner is RequireMembership plus a role check -- same
// fail-closed treatment of lookup errors.
func (s *Service) RequireOrgOwner(ctx context.Context, orgID, userID uuid.UUID) error {
	member, err := s.q.GetOrgMember(ctx, db.GetOrgMemberParams{OrgID: pgUUID(orgID), UserID: pgUUID(userID)})
	if err != nil {
		return ErrNotMember
	}
	if member.Role != "owner" {
		return ErrNotOwner
	}
	return nil
}

// InviteMember adds a new person to an org. If a Zitadel service account
// is configured, this creates their Zitadel identity outright (see
// internal/zitadelmgmt) and links it into org_members in one step -- no
// separate "pending invite" state to track, since Zitadel hands back the
// new user's ID synchronously and JITProvisionUser already treats "row
// exists" as a no-op on whatever login eventually completes their
// Zitadel-side signup.
func (s *Service) InviteMember(ctx context.Context, orgID uuid.UUID, email, displayName, role string) (db.User, error) {
	if s.zm == nil {
		return db.User{}, ErrInvitesNotConfigured
	}

	zitadelSubject, err := s.zm.InviteHuman(ctx, email, displayName)
	if err != nil {
		return db.User{}, fmt.Errorf("invite in zitadel: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)

	user, err := q.CreateUser(ctx, db.CreateUserParams{
		ID:             pgUUID(uuid.New()),
		ZitadelSubject: zitadelSubject,
		Email:          email,
		DisplayName:    nullableString(displayName),
	})
	if err != nil {
		return db.User{}, fmt.Errorf("create user: %w", err)
	}

	if _, err := q.AddOrgMember(ctx, db.AddOrgMemberParams{OrgID: pgUUID(orgID), UserID: user.ID, Role: role}); err != nil {
		return db.User{}, fmt.Errorf("add org membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.User{}, fmt.Errorf("commit tx: %w", err)
	}
	return user, nil
}

// CreateInstanceParams is what the dashboard's "New instance" form collects.
type CreateInstanceParams struct {
	Slug          string
	Name          string
	AdminUsername string
	AdminEmail    string
}

// CreateInstance provisions a tenant in the listmonk fork for this org: an
// instances row is created up front (status "created"), a provisioning_jobs
// row tracks the single "provision_listmonk_tenant" step for the UI's
// timeline, then the Operator API is called. Success records the fork's
// tenant id and one-time setup link and flips the instance to "active";
// failure (most commonly a taken slug) flips it to "failed" with the
// error recorded on the job.
//
// This step runs synchronously rather than through a job queue for now --
// docs/plan.md's River-backed job chain (Postmark, DNS) is still ahead of
// this in the pipeline and isn't wired up yet, so there's nothing to retry
// asynchronously for *this* step yet. Revisit once those steps land.
func (s *Service) CreateInstance(ctx context.Context, orgID uuid.UUID, p CreateInstanceParams) (db.Instance, error) {
	org, err := s.q.GetOrgByID(ctx, pgUUID(orgID))
	if err != nil {
		return db.Instance{}, fmt.Errorf("get org: %w", err)
	}

	inst, err := s.q.CreateInstance(ctx, db.CreateInstanceParams{
		ID:            pgUUID(uuid.New()),
		OrgID:         pgUUID(orgID),
		Slug:          p.Slug,
		Name:          p.Name,
		AdminUsername: p.AdminUsername,
		AdminEmail:    p.AdminEmail,
		Status:        "created",
	})
	if err != nil {
		// listmonk tenant slugs are a single flat namespace shared by every
		// org on this platform (see internal/operatorclient's reTenantSlug
		// note), so instances.slug is globally unique here too, not just
		// per-org -- this local constraint can catch a collision before a
		// round trip to the Operator API does.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "instances_slug_key" {
			return db.Instance{}, ErrSlugTaken
		}
		return db.Instance{}, fmt.Errorf("create instance row: %w", err)
	}

	job, err := s.q.CreateProvisioningJob(ctx, db.CreateProvisioningJobParams{
		ID:         pgUUID(uuid.New()),
		InstanceID: inst.ID,
		JobType:    "provision_listmonk_tenant",
	})
	if err != nil {
		return db.Instance{}, fmt.Errorf("create provisioning job row: %w", err)
	}

	if _, err := s.q.UpdateInstanceStatus(ctx, db.UpdateInstanceStatusParams{
		ID: inst.ID, Status: "listmonk_tenant_provisioning",
	}); err != nil {
		return db.Instance{}, fmt.Errorf("update instance status: %w", err)
	}

	tenantParams := operatorclient.CreateTenantParams{
		Slug:          p.Slug,
		Name:          p.Name,
		AdminUsername: p.AdminUsername,
		AdminEmail:    p.AdminEmail,
	}
	if org.ListmonkOrganizationID != nil {
		tenantParams.OrganizationID = int(*org.ListmonkOrganizationID)
	}
	result, opErr := s.op.CreateTenant(ctx, tenantParams)
	if opErr != nil {
		failErr := ErrSlugTaken
		if !operatorclient.IsConflict(opErr) {
			failErr = opErr
		}
		if _, err := s.q.UpdateProvisioningJobStatus(ctx, db.UpdateProvisioningJobStatusParams{
			ID: job.ID, Status: "failed", LastError: nullableString(opErr.Error()),
		}); err != nil {
			return db.Instance{}, fmt.Errorf("record job failure: %w", err)
		}
		failed, err := s.q.UpdateInstanceStatus(ctx, db.UpdateInstanceStatusParams{ID: inst.ID, Status: "failed"})
		if err != nil {
			return db.Instance{}, fmt.Errorf("update instance status: %w", err)
		}
		return failed, failErr
	}

	if _, err := s.q.UpdateProvisioningJobStatus(ctx, db.UpdateProvisioningJobStatusParams{
		ID: job.ID, Status: "succeeded",
	}); err != nil {
		return db.Instance{}, fmt.Errorf("record job success: %w", err)
	}

	tenantID := int32(result.Tenant.ID)
	active, err := s.q.SetInstanceListmonkTenant(ctx, db.SetInstanceListmonkTenantParams{
		ID: inst.ID, ListmonkTenantID: &tenantID, AdminSetupUrl: nullableString(result.SetupURL),
	})
	if err != nil {
		return db.Instance{}, fmt.Errorf("record listmonk tenant: %w", err)
	}
	return s.q.UpdateInstanceStatus(ctx, db.UpdateInstanceStatusParams{ID: active.ID, Status: "active"})
}

func (s *Service) ListInstances(ctx context.Context, orgID uuid.UUID) ([]db.Instance, error) {
	return s.q.ListInstancesByOrg(ctx, pgUUID(orgID))
}

func (s *Service) GetInstance(ctx context.Context, orgID, instanceID uuid.UUID) (db.Instance, error) {
	return s.q.GetInstanceForOrg(ctx, db.GetInstanceForOrgParams{ID: pgUUID(instanceID), OrgID: pgUUID(orgID)})
}

func (s *Service) ListProvisioningEvents(ctx context.Context, instanceID uuid.UUID) ([]db.ProvisioningJob, error) {
	return s.q.ListProvisioningJobsByInstance(ctx, pgUUID(instanceID))
}

// ResendSetupLink reissues the new admin's one-time setup link -- needed
// because the token behind it lives only in the listmonk fork's memory and
// is lost on its restart (see internal/operatorclient's CreateTenantResult
// doc comment).
func (s *Service) ResendSetupLink(ctx context.Context, orgID, instanceID uuid.UUID) (string, error) {
	inst, err := s.GetInstance(ctx, orgID, instanceID)
	if err != nil {
		return "", fmt.Errorf("get instance: %w", err)
	}
	return s.resendSetupLinkFor(ctx, inst)
}

func (s *Service) resendSetupLinkFor(ctx context.Context, inst db.Instance) (string, error) {
	if inst.ListmonkTenantID == nil {
		return "", errors.New("instance has no listmonk tenant yet")
	}

	result, err := s.op.CreateSetupLink(ctx, int(*inst.ListmonkTenantID), inst.AdminEmail)
	if err != nil {
		return "", fmt.Errorf("reissue setup link: %w", err)
	}

	if _, err := s.q.UpdateInstanceSetupURL(ctx, db.UpdateInstanceSetupURLParams{
		ID: inst.ID, AdminSetupUrl: nullableString(result.SetupURL),
	}); err != nil {
		return "", fmt.Errorf("store setup link: %w", err)
	}
	return result.SetupURL, nil
}

// --- super admin: platform-wide, bypasses org membership ----------------
//
// A super admin sees every org's instances, not just their own -- distinct
// from an org_members "owner" (scoped to one org) and from listmonk's own
// per-tenant "Super Admin" role (scoped to one tenant). There's no
// self-service grant; see users.is_super_admin's doc comment in the
// migration.

// ErrNotSuperAdmin is returned by RequireSuperAdmin.
var ErrNotSuperAdmin = errors.New("not a super admin")

// RequireSuperAdmin fails closed, same as RequireMembership: a lookup
// error and "found but not a super admin" are indistinguishable to the
// caller.
func (s *Service) RequireSuperAdmin(ctx context.Context, userID uuid.UUID) error {
	user, err := s.q.GetUserByID(ctx, pgUUID(userID))
	if err != nil || !user.IsSuperAdmin {
		return ErrNotSuperAdmin
	}
	return nil
}

func (s *Service) AdminListOrgs(ctx context.Context) ([]db.ListAllOrgsWithInstanceCountRow, error) {
	return s.q.ListAllOrgsWithInstanceCount(ctx)
}

func (s *Service) AdminListInstances(ctx context.Context) ([]db.ListAllInstancesWithOrgNameRow, error) {
	return s.q.ListAllInstancesWithOrgName(ctx)
}

func (s *Service) AdminGetInstance(ctx context.Context, instanceID uuid.UUID) (db.Instance, error) {
	return s.q.GetInstanceByID(ctx, pgUUID(instanceID))
}

// AdminGetTenantLiveStatus fetches the tenant's live status + cross-tenant
// counts directly from listmonk. Deliberately not read from
// instances.status: that column tracks provisioning state
// (created/.../active/failed), a different dimension from the tenant's
// actual active/suspended/disabled lifecycle in the fork.
func (s *Service) AdminGetTenantLiveStatus(ctx context.Context, instanceID uuid.UUID) (operatorclient.TenantWithCounts, error) {
	inst, err := s.q.GetInstanceByID(ctx, pgUUID(instanceID))
	if err != nil {
		return operatorclient.TenantWithCounts{}, fmt.Errorf("get instance: %w", err)
	}
	if inst.ListmonkTenantID == nil {
		return operatorclient.TenantWithCounts{}, errors.New("instance has no listmonk tenant yet")
	}
	return s.op.GetTenant(ctx, int(*inst.ListmonkTenantID))
}

// AdminSetTenantStatus suspends/reactivates/disables a tenant directly in
// listmonk. Deliberately not mirrored into instances.status -- see
// AdminGetTenantLiveStatus's doc comment.
func (s *Service) AdminSetTenantStatus(ctx context.Context, instanceID uuid.UUID, status string) (operatorclient.Tenant, error) {
	inst, err := s.q.GetInstanceByID(ctx, pgUUID(instanceID))
	if err != nil {
		return operatorclient.Tenant{}, fmt.Errorf("get instance: %w", err)
	}
	if inst.ListmonkTenantID == nil {
		return operatorclient.Tenant{}, errors.New("instance has no listmonk tenant yet")
	}
	return s.op.UpdateTenantStatus(ctx, int(*inst.ListmonkTenantID), status)
}

// AdminResendSetupLink is ResendSetupLink without the org-membership scope
// -- a super admin can reissue any tenant's setup link.
func (s *Service) AdminResendSetupLink(ctx context.Context, instanceID uuid.UUID) (string, error) {
	inst, err := s.q.GetInstanceByID(ctx, pgUUID(instanceID))
	if err != nil {
		return "", fmt.Errorf("get instance: %w", err)
	}
	return s.resendSetupLinkFor(ctx, inst)
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
