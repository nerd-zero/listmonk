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
	"log"
	"regexp"
	"strconv"
	"strings"

	"listnun/internal/cryptoutil"
	"listnun/internal/db"
	"listnun/internal/operatorclient"
	"listnun/internal/postmarkclient"
	"listnun/internal/zitadelmgmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	pm   *PostmarkConfig     // nil unless a Postmark account token is configured
}

// PostmarkConfig bundles everything CreateInstance/AddSenderDomain/
// AddSenderSignature need to talk to Postmark and protect its credentials
// at rest. Passed as a single optional pointer (nil disables all of it)
// rather than separate New params, since both only ever matter together.
type PostmarkConfig struct {
	Client *postmarkclient.Client
	// EncryptionKey encrypts the Postmark server token before it's stored
	// in postmark_servers.api_token_encrypted -- see internal/cryptoutil.
	EncryptionKey [32]byte
	// SharedDomainRoot is the parent domain AddPlatformDomain hands out to
	// an org with no domain of its own -- an instance with slug "acme"
	// gets acme.<this>.
	SharedDomainRoot string
}

func New(pool *pgxpool.Pool, op *operatorclient.Client, zm *zitadelmgmt.Client, pm *PostmarkConfig) *Service {
	return &Service{pool: pool, q: db.New(pool), op: op, zm: zm, pm: pm}
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
	s.createDefaultInstance(ctx, orgID, org.Name, email)
	return user, nil
}

// personalOrgName names a new user's personal org after them directly --
// their first name, or the local part of their email if there's no
// display name -- rather than "<name>'s org": that suffix reads fine on
// its own but slugifies into an awkward "-s-org" (see slugify), doubly so
// once a multi-word display name is in the mix ("Alex Alexson" -> org
// "Alex Alexson's org" -> slug "alex-alexson-s-org").
func personalOrgName(displayName, email string) string {
	if displayName != "" {
		first, _, _ := strings.Cut(displayName, " ")
		return first
	}
	if email != "" {
		local, _, _ := strings.Cut(email, "@")
		return local
	}
	return "My org"
}

var reSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

// slugify mirrors web/src/lib/slug.ts's slugifyPart -- lowercase,
// non-alphanumeric runs collapsed to a single hyphen, no leading/trailing
// hyphen. Used to derive the org's default instance slug server-side,
// where there's no form input to slugify as the user types.
func slugify(s string) string {
	return strings.Trim(reSlugChars.ReplaceAllString(strings.ToLower(s), "-"), "-")
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
	if owner, err := s.q.GetUserByID(ctx, pgUUID(ownerUserID)); err == nil {
		s.createDefaultInstance(ctx, orgID, org.Name, owner.Email)
	}
	return org, nil
}

func (s *Service) ListOrgsForUser(ctx context.Context, userID uuid.UUID) ([]db.ListOrgsByUserRow, error) {
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

// ListMembers returns everyone in an org, joined with their user profile.
func (s *Service) ListMembers(ctx context.Context, orgID uuid.UUID) ([]db.ListOrgMembersWithUserRow, error) {
	return s.q.ListOrgMembersWithUser(ctx, pgUUID(orgID))
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
	active, err = s.q.UpdateInstanceStatus(ctx, db.UpdateInstanceStatusParams{ID: active.ID, Status: "active"})
	if err != nil {
		return db.Instance{}, fmt.Errorf("update instance status: %w", err)
	}

	// Postmark is a secondary provisioning step, tracked by its own
	// provisioning_jobs row rather than the instance's own status: the
	// tenant is already usable (with listmonk's placeholder SMTP examples)
	// even if this fails, so a Postmark error doesn't flip the instance
	// back to "failed" -- it's surfaced in the timeline instead, and safe
	// to retry by re-running CreateInstance's Postmark step once that's
	// wired to River (see docs/plan.md).
	//
	// This only creates the server itself -- no sending domain or sender
	// signature yet, and so no SMTP push into listmonk yet either. Those
	// need a real "from" identity, which the org supplies themselves via
	// AddSenderDomain/AddSenderSignature below.
	if s.pm != nil {
		if err := s.provisionPostmarkServer(ctx, active); err != nil {
			return active, fmt.Errorf("provision postmark server: %w", err)
		}
	}
	return active, nil
}

// provisionPostmarkServer creates a dedicated Postmark server for inst and
// stores its encrypted token -- see docs/plan.md's create_postmark_server
// step and internal/postmarkclient's doc comment.
func (s *Service) provisionPostmarkServer(ctx context.Context, inst db.Instance) error {
	job, err := s.q.CreateProvisioningJob(ctx, db.CreateProvisioningJobParams{
		ID:         pgUUID(uuid.New()),
		InstanceID: inst.ID,
		JobType:    "provision_postmark_server",
	})
	if err != nil {
		return fmt.Errorf("create provisioning job row: %w", err)
	}

	if jobErr := s.doProvisionPostmarkServer(ctx, inst); jobErr != nil {
		if _, err := s.q.UpdateProvisioningJobStatus(ctx, db.UpdateProvisioningJobStatusParams{
			ID: job.ID, Status: "failed", LastError: nullableString(jobErr.Error()),
		}); err != nil {
			return fmt.Errorf("record job failure: %w", err)
		}
		return jobErr
	}

	if _, err := s.q.UpdateProvisioningJobStatus(ctx, db.UpdateProvisioningJobStatusParams{
		ID: job.ID, Status: "succeeded",
	}); err != nil {
		return fmt.Errorf("record job success: %w", err)
	}
	return nil
}

func (s *Service) doProvisionPostmarkServer(ctx context.Context, inst db.Instance) error {
	server, err := s.pm.Client.CreateServer(ctx, "listnun-"+inst.Slug)
	if err != nil {
		return fmt.Errorf("create postmark server: %w", err)
	}
	if len(server.ApiTokens) == 0 {
		return errors.New("postmark server has no API tokens")
	}

	encryptedToken, err := cryptoutil.Encrypt(s.pm.EncryptionKey, server.ApiTokens[0])
	if err != nil {
		return fmt.Errorf("encrypt postmark token: %w", err)
	}

	if _, err := s.q.CreatePostmarkServer(ctx, db.CreatePostmarkServerParams{
		ID:                pgUUID(uuid.New()),
		InstanceID:        inst.ID,
		PostmarkServerID:  strconv.Itoa(server.ID),
		ApiTokenEncrypted: encryptedToken,
	}); err != nil {
		return fmt.Errorf("store postmark server: %w", err)
	}
	return nil
}

// ErrInstanceHasNoPostmarkServer is returned by DeletePostmarkServer when
// the instance has none to delete -- e.g. Postmark wasn't configured when
// it was created, or one was already removed.
var ErrInstanceHasNoPostmarkServer = errors.New("this instance has no postmark server")

// DeletePostmarkServer removes instance's Postmark server on its own,
// without touching the instance/tenant itself -- e.g. to force
// re-provisioning, or to tear down sending for an instance moving its
// domain elsewhere. The instance is left without email sending until
// CreateInstance's Postmark step (or a future re-provision action) runs
// again.
func (s *Service) DeletePostmarkServer(ctx context.Context, orgID, instanceID uuid.UUID) error {
	inst, err := s.GetInstance(ctx, orgID, instanceID)
	if err != nil {
		return fmt.Errorf("get instance: %w", err)
	}
	return s.deletePostmarkServerFor(ctx, inst)
}

// AdminDeletePostmarkServer is DeletePostmarkServer without the
// org-membership scope -- a super admin can remove any instance's server.
func (s *Service) AdminDeletePostmarkServer(ctx context.Context, instanceID uuid.UUID) error {
	inst, err := s.q.GetInstanceByID(ctx, pgUUID(instanceID))
	if err != nil {
		return fmt.Errorf("get instance: %w", err)
	}
	return s.deletePostmarkServerFor(ctx, inst)
}

func (s *Service) deletePostmarkServerFor(ctx context.Context, inst db.Instance) error {
	if s.pm == nil {
		return ErrPostmarkNotConfigured
	}
	server, err := s.q.GetPostmarkServerByInstanceID(ctx, inst.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInstanceHasNoPostmarkServer
		}
		return fmt.Errorf("get postmark server: %w", err)
	}

	pmID, err := strconv.Atoi(server.PostmarkServerID)
	if err != nil {
		return fmt.Errorf("parse postmark server id: %w", err)
	}
	if err := s.pm.Client.DeleteServer(ctx, pmID); err != nil && !postmarkclient.IsNotFound(err) {
		return fmt.Errorf("delete postmark server: %w", err)
	}

	if err := s.q.DeletePostmarkServerByInstanceID(ctx, inst.ID); err != nil {
		return fmt.Errorf("delete postmark server row: %w", err)
	}
	return nil
}

// ErrPostmarkNotConfigured is returned by AddSenderDomain/AddSenderSignature
// when no Postmark account token is configured.
var ErrPostmarkNotConfigured = errors.New("postmark is not configured")

// ErrSenderIdentityExists is returned when the instance already has a
// sender identity -- exactly one (domain or sender signature) per instance
// for now, so a second attempt is a no-op rather than a silent overwrite.
var ErrSenderIdentityExists = errors.New("this instance already has a sender identity")

// ErrSenderIdentityTaken is returned when the domain or sender email is
// already claimed by another instance -- sender_identities_value_key is
// global, so two different orgs can never share one.
var ErrSenderIdentityTaken = errors.New("this domain or sender email is already in use by another workspace")

// ErrSenderIdentityNotFound is returned by GetSenderIdentity when the
// instance hasn't added one yet -- an expected, common state (not yet
// configured), distinct from a real lookup failure.
var ErrSenderIdentityNotFound = errors.New("this instance has no sender identity yet")

// GetSenderIdentity returns instance's sender identity, if any, plus the
// DNS records to publish for it (empty for a sender signature -- those
// don't need DNS, just clicking Postmark's confirmation email).
func (s *Service) GetSenderIdentity(ctx context.Context, orgID, instanceID uuid.UUID) (db.SenderIdentity, []db.DnsRecord, error) {
	inst, err := s.GetInstance(ctx, orgID, instanceID)
	if err != nil {
		return db.SenderIdentity{}, nil, fmt.Errorf("get instance: %w", err)
	}
	identity, err := s.q.GetSenderIdentityByInstanceID(ctx, inst.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.SenderIdentity{}, nil, ErrSenderIdentityNotFound
		}
		return db.SenderIdentity{}, nil, fmt.Errorf("get sender identity: %w", err)
	}

	// Postmark doesn't push a webhook when DKIM verifies or a sender
	// signature is confirmed -- it's only ever discoverable by asking. So
	// every fetch of a still-pending identity checks Postmark live and
	// persists the flip to "confirmed" if it's happened. A failed check
	// here (Postmark down, network hiccup) isn't fatal to the read --
	// the org just sees "pending" a little longer, not an error.
	if identity.Status == "pending" && s.pm != nil {
		if refreshed, ok := s.refreshSenderIdentityStatus(ctx, identity); ok {
			identity = refreshed
		}
	}

	records, err := s.q.ListDNSRecordsByInstance(ctx, inst.ID)
	if err != nil {
		return db.SenderIdentity{}, nil, fmt.Errorf("list dns records: %w", err)
	}
	return identity, records, nil
}

// refreshSenderIdentityStatus asks Postmark whether identity has been
// verified/confirmed yet and, if so, persists that. ok is false whenever
// nothing changed -- either it's still pending, or the check itself
// failed -- so the caller can just keep using the identity it already had.
func (s *Service) refreshSenderIdentityStatus(ctx context.Context, identity db.SenderIdentity) (db.SenderIdentity, bool) {
	postmarkID, err := strconv.Atoi(identity.PostmarkID)
	if err != nil {
		return identity, false
	}

	var verified bool
	switch identity.Kind {
	case "domain", "platform_domain":
		domain, err := s.pm.Client.VerifyDKIM(ctx, postmarkID)
		if err != nil {
			return identity, false
		}
		verified = domain.DKIMVerified
	case "sender_signature":
		sig, err := s.pm.Client.GetSenderSignature(ctx, postmarkID)
		if err != nil {
			return identity, false
		}
		verified = sig.Confirmed
	default:
		return identity, false
	}

	if !verified {
		return identity, false
	}
	confirmed, err := s.q.MarkSenderIdentityConfirmed(ctx, identity.ID)
	if err != nil {
		return identity, false
	}
	return confirmed, true
}

// AddSenderDomain gives instance a full sending domain of the org's own:
// Postmark's Domain API returns a DKIM record, which is stored so the
// dashboard can show the org what to publish themselves. SMTP credentials
// are pushed into the listmonk tenant using this domain's "noreply@"
// address as the From address -- deliverability isn't gated on DKIM
// actually verifying first, matching Postmark's own model (a domain can
// send, just with weaker trust, before DKIM is confirmed).
func (s *Service) AddSenderDomain(ctx context.Context, orgID, instanceID uuid.UUID, domain string) (db.SenderIdentity, []db.DnsRecord, error) {
	if s.pm == nil {
		return db.SenderIdentity{}, nil, ErrPostmarkNotConfigured
	}
	inst, err := s.GetInstance(ctx, orgID, instanceID)
	if err != nil {
		return db.SenderIdentity{}, nil, fmt.Errorf("get instance: %w", err)
	}
	return s.addDomainIdentity(ctx, inst, "domain", domain)
}

// AddPlatformDomain is AddSenderDomain's alternative for an org with no
// domain of its own: a subdomain of PostmarkConfig.SharedDomainRoot is
// derived from the instance's slug (already globally unique, so this can
// never collide) rather than typed in. The resulting DKIM record is ours
// to publish, not the org's -- there's nothing further for them to do
// beyond clicking to opt in.
func (s *Service) AddPlatformDomain(ctx context.Context, orgID, instanceID uuid.UUID) (db.SenderIdentity, []db.DnsRecord, error) {
	if s.pm == nil {
		return db.SenderIdentity{}, nil, ErrPostmarkNotConfigured
	}
	inst, err := s.GetInstance(ctx, orgID, instanceID)
	if err != nil {
		return db.SenderIdentity{}, nil, fmt.Errorf("get instance: %w", err)
	}
	domain := inst.Slug + "." + s.pm.SharedDomainRoot
	return s.addDomainIdentity(ctx, inst, "platform_domain", domain)
}

// addDomainIdentity is AddSenderDomain and AddPlatformDomain's shared
// tail: both create a real Postmark domain and store it the same way,
// differing only in whose domain it is (kind) and where the name came
// from (the org's own input, vs derived from the instance's slug).
func (s *Service) addDomainIdentity(ctx context.Context, inst db.Instance, kind, domain string) (db.SenderIdentity, []db.DnsRecord, error) {
	if err := s.requireNoSenderIdentity(ctx, inst.ID); err != nil {
		return db.SenderIdentity{}, nil, err
	}

	pmDomain, err := s.pm.Client.CreateDomain(ctx, domain)
	if err != nil {
		return db.SenderIdentity{}, nil, fmt.Errorf("create postmark domain: %w", err)
	}

	identity, err := s.q.CreateSenderIdentity(ctx, db.CreateSenderIdentityParams{
		ID:         pgUUID(uuid.New()),
		InstanceID: inst.ID,
		Kind:       kind,
		Value:      domain,
		PostmarkID: strconv.Itoa(pmDomain.ID),
	})
	if err != nil {
		return db.SenderIdentity{}, nil, mapSenderIdentityConstraint(err)
	}

	var records []db.DnsRecord
	if pmDomain.DKIMHost != "" {
		rec, err := s.q.CreateDNSRecord(ctx, db.CreateDNSRecordParams{
			ID:         pgUUID(uuid.New()),
			InstanceID: inst.ID,
			RecordType: "dkim",
			Host:       pmDomain.DKIMHost,
			Value:      pmDomain.DKIMTextValue,
		})
		if err != nil {
			return identity, nil, fmt.Errorf("store dkim record: %w", err)
		}
		records = append(records, rec)
	}

	if err := s.pushSMTPCredentials(ctx, inst, "noreply@"+domain); err != nil {
		return identity, records, err
	}
	return identity, records, nil
}

// AddSenderSignature gives instance a single verified sender address
// instead of a full domain -- no DNS involved; Postmark emails a
// confirmation link straight to fromEmail; identity.Status starts and
// stays "pending" until that's clicked (there's no API to poll or push
// that state yet -- see the sender_identities migration's own comment).
func (s *Service) AddSenderSignature(ctx context.Context, orgID, instanceID uuid.UUID, fromEmail, name string) (db.SenderIdentity, error) {
	if s.pm == nil {
		return db.SenderIdentity{}, ErrPostmarkNotConfigured
	}
	inst, err := s.GetInstance(ctx, orgID, instanceID)
	if err != nil {
		return db.SenderIdentity{}, fmt.Errorf("get instance: %w", err)
	}
	if err := s.requireNoSenderIdentity(ctx, inst.ID); err != nil {
		return db.SenderIdentity{}, err
	}

	sig, err := s.pm.Client.CreateSenderSignature(ctx, fromEmail, name)
	if err != nil {
		return db.SenderIdentity{}, fmt.Errorf("create postmark sender signature: %w", err)
	}

	identity, err := s.q.CreateSenderIdentity(ctx, db.CreateSenderIdentityParams{
		ID:         pgUUID(uuid.New()),
		InstanceID: inst.ID,
		Kind:       "sender_signature",
		Value:      fromEmail,
		PostmarkID: strconv.Itoa(sig.ID),
	})
	if err != nil {
		return db.SenderIdentity{}, mapSenderIdentityConstraint(err)
	}

	if err := s.pushSMTPCredentials(ctx, inst, fromEmail); err != nil {
		return identity, err
	}
	return identity, nil
}

// DeleteSenderIdentity removes instance's sender identity (domain or
// sender signature) from Postmark and locally, along with any DNS records
// published for it -- irreversible. The instance is left without a
// confirmed "from" address until a new identity is added; its existing
// listmonk SMTP config is left as-is (still pointed at the now-orphaned
// address) since there's nothing sensible to fall back to automatically.
func (s *Service) DeleteSenderIdentity(ctx context.Context, orgID, instanceID uuid.UUID) error {
	inst, err := s.GetInstance(ctx, orgID, instanceID)
	if err != nil {
		return fmt.Errorf("get instance: %w", err)
	}
	if s.pm == nil {
		return ErrPostmarkNotConfigured
	}

	identity, err := s.q.GetSenderIdentityByInstanceID(ctx, inst.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSenderIdentityNotFound
		}
		return fmt.Errorf("get sender identity: %w", err)
	}

	postmarkID, err := strconv.Atoi(identity.PostmarkID)
	if err != nil {
		return fmt.Errorf("parse postmark id: %w", err)
	}
	switch identity.Kind {
	case "domain", "platform_domain":
		if err := s.pm.Client.DeleteDomain(ctx, postmarkID); err != nil && !postmarkclient.IsNotFound(err) {
			return fmt.Errorf("delete postmark domain: %w", err)
		}
	case "sender_signature":
		if err := s.pm.Client.DeleteSenderSignature(ctx, postmarkID); err != nil && !postmarkclient.IsNotFound(err) {
			return fmt.Errorf("delete postmark sender signature: %w", err)
		}
	}

	if err := s.q.DeleteDNSRecordsByInstance(ctx, inst.ID); err != nil {
		return fmt.Errorf("delete dns records: %w", err)
	}
	if err := s.q.DeleteSenderIdentityByInstanceID(ctx, inst.ID); err != nil {
		return fmt.Errorf("delete sender identity row: %w", err)
	}
	return nil
}

func (s *Service) requireNoSenderIdentity(ctx context.Context, instanceID pgtype.UUID) error {
	_, err := s.q.GetSenderIdentityByInstanceID(ctx, instanceID)
	switch {
	case err == nil:
		return ErrSenderIdentityExists
	case errors.Is(err, pgx.ErrNoRows):
		return nil
	default:
		return fmt.Errorf("check existing sender identity: %w", err)
	}
}

func mapSenderIdentityConstraint(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.ConstraintName == "sender_identities_value_key" {
		return ErrSenderIdentityTaken
	}
	return fmt.Errorf("store sender identity: %w", err)
}

// pushSMTPCredentials sends the instance's Postmark server credentials
// into its listmonk tenant, replacing the placeholder SMTP examples --
// see internal/operatorclient.SetTenantSMTP's doc comment. fromAddress is
// whatever identity (domain or sender signature) was just confirmed.
func (s *Service) pushSMTPCredentials(ctx context.Context, inst db.Instance, fromAddress string) error {
	pmServer, err := s.q.GetPostmarkServerByInstanceID(ctx, inst.ID)
	if err != nil {
		return fmt.Errorf("get postmark server: %w", err)
	}
	token, err := cryptoutil.Decrypt(s.pm.EncryptionKey, pmServer.ApiTokenEncrypted)
	if err != nil {
		return fmt.Errorf("decrypt postmark token: %w", err)
	}
	if inst.ListmonkTenantID == nil {
		return errors.New("instance has no listmonk tenant yet")
	}

	err = s.op.SetTenantSMTP(ctx, int(*inst.ListmonkTenantID), operatorclient.SMTPEntry{
		Name:          "Postmark",
		Enabled:       true,
		Host:          "smtp.postmarkapp.com",
		Port:          587,
		AuthProtocol:  "plain",
		Username:      token,
		Password:      token,
		EmailHeaders:  []map[string]string{},
		MaxConns:      10,
		MaxMsgRetries: 2,
		MsgRetryDelay: "10ms",
		IdleTimeout:   "15s",
		WaitTimeout:   "5s",
		TLSType:       "STARTTLS",
		FromAddresses: []string{fromAddress},
	})
	if err != nil {
		return fmt.Errorf("push smtp credentials to listmonk: %w", err)
	}
	return nil
}

// createDefaultInstance provisions a new org's root workspace, named after
// the org itself (e.g. org "Acme" gets instance slug "acme") so there's
// always something to land on right after signup instead of an empty
// dashboard. Best-effort: called after the org's own transaction has
// already committed, so a failure here (operator API down, no ownerEmail
// to register the tenant admin with, ...) is logged and swallowed rather
// than rolling back or failing the signup/org-creation that triggered it.
func (s *Service) createDefaultInstance(ctx context.Context, orgID uuid.UUID, orgName, ownerEmail string) {
	if ownerEmail == "" {
		log.Printf("provisioning: skipping default instance for org %s: no owner email to register the tenant admin with", orgID)
		return
	}

	base := slugify(orgName)
	if base == "" {
		base = "org"
	}

	params := CreateInstanceParams{
		Slug: base, Name: orgName, AdminUsername: "admin", AdminEmail: ownerEmail,
	}
	if _, err := s.CreateInstance(ctx, orgID, params); err != nil {
		if !errors.Is(err, ErrSlugTaken) {
			log.Printf("provisioning: default instance for org %s: %v", orgID, err)
			return
		}
		// The flat slug namespace (see CreateInstance's doc comment) means
		// another org already has this name -- disambiguate with a chunk
		// of this org's own id, the same way createListmonkOrganization
		// disambiguates the listmonk-side org name.
		params.Slug = base + "-" + orgID.String()[:8]
		if _, err := s.CreateInstance(ctx, orgID, params); err != nil {
			log.Printf("provisioning: default instance for org %s: %v", orgID, err)
		}
	}
}

func (s *Service) ListInstances(ctx context.Context, orgID uuid.UUID) ([]db.Instance, error) {
	return s.q.ListInstancesByOrg(ctx, pgUUID(orgID))
}

func (s *Service) GetInstance(ctx context.Context, orgID, instanceID uuid.UUID) (db.Instance, error) {
	return s.q.GetInstanceForOrg(ctx, db.GetInstanceForOrgParams{ID: pgUUID(instanceID), OrgID: pgUUID(orgID)})
}

// DeleteInstance permanently deletes instance: its Postmark server (if
// any), its listmonk tenant (which cascades into all of the tenant's own
// data -- subscribers, campaigns, etc., see operatorclient.DeleteTenant's
// doc comment), and finally the local row (which cascades into
// sender_identities/dns_records/provisioning_jobs). Irreversible.
//
// External resources are deleted before the local row, in that order, so a
// failure partway through leaves the local row in place to retry against
// rather than orphaning a Postmark server or listmonk tenant whose id we've
// just lost.
func (s *Service) DeleteInstance(ctx context.Context, orgID, instanceID uuid.UUID) error {
	inst, err := s.GetInstance(ctx, orgID, instanceID)
	if err != nil {
		return fmt.Errorf("get instance: %w", err)
	}
	return s.deleteInstanceFor(ctx, inst)
}

// AdminDeleteInstance is DeleteInstance without the org-membership scope --
// a super admin can delete any instance.
func (s *Service) AdminDeleteInstance(ctx context.Context, instanceID uuid.UUID) error {
	inst, err := s.q.GetInstanceByID(ctx, pgUUID(instanceID))
	if err != nil {
		return fmt.Errorf("get instance: %w", err)
	}
	return s.deleteInstanceFor(ctx, inst)
}

func (s *Service) deleteInstanceFor(ctx context.Context, inst db.Instance) error {
	if s.pm != nil {
		if err := s.deletePostmarkServerFor(ctx, inst); err != nil && !errors.Is(err, ErrInstanceHasNoPostmarkServer) {
			return fmt.Errorf("delete postmark server: %w", err)
		}
	}

	if inst.ListmonkTenantID != nil {
		if err := s.op.DeleteTenant(ctx, int(*inst.ListmonkTenantID)); err != nil && !operatorclient.IsNotFound(err) {
			return fmt.Errorf("delete listmonk tenant: %w", err)
		}
	}

	if err := s.q.DeleteInstance(ctx, inst.ID); err != nil {
		return fmt.Errorf("delete instance row: %w", err)
	}
	return nil
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
