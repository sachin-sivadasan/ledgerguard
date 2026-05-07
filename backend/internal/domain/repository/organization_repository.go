package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// OrganizationRepository defines the interface for organization persistence.
//
// Error handling contract:
//   - Single-item Find methods return a sentinel error (ErrOrgNotFound)
//     when no matching record exists.
//   - List methods return an empty slice when no records match.
//   - All methods may return other errors for database/network failures.
type OrganizationRepository interface {
	// Create persists a new organization.
	Create(ctx context.Context, org *entity.Organization) error

	// FindByID returns the organization with the given ID.
	// Returns ErrOrgNotFound if no org exists with that ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error)

	// FindBySlug returns the organization with the given slug.
	// Returns ErrOrgNotFound if no org exists with that slug.
	FindBySlug(ctx context.Context, slug string) (*entity.Organization, error)

	// Update updates an existing organization. Returns error if not found.
	Update(ctx context.Context, org *entity.Organization) error

	// Delete removes an organization by ID. Returns error if not found.
	Delete(ctx context.Context, id uuid.UUID) error
}

// MemberRepository defines the interface for org member persistence.
//
// Error handling contract:
//   - Single-item Find methods return ErrMemberNotFound.
//   - List methods return an empty slice when no records match.
type MemberRepository interface {
	// Create persists a new org member.
	Create(ctx context.Context, member *entity.OrgMember) error

	// FindByID returns the member with the given ID.
	// Returns ErrMemberNotFound if not found.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.OrgMember, error)

	// FindByOrgAndUser returns the member for a specific org+user pair.
	// Returns ErrMemberNotFound if the user is not a member of the org.
	FindByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*entity.OrgMember, error)

	// FindByOrgID returns all members of an organization.
	// Returns empty slice if org has no members.
	FindByOrgID(ctx context.Context, orgID uuid.UUID) ([]*entity.OrgMember, error)

	// FindByUserID returns all org memberships for a user.
	// Returns empty slice if user belongs to no orgs.
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.OrgMember, error)

	// CountByOrgID returns the count of active members in an org.
	CountByOrgID(ctx context.Context, orgID uuid.UUID) (int, error)

	// Update updates an existing member (role, status, notification prefs, etc.).
	Update(ctx context.Context, member *entity.OrgMember) error

	// Delete removes a member from an org.
	Delete(ctx context.Context, id uuid.UUID) error
}

// InvitationRepository defines the interface for org invitation persistence.
//
// Error handling contract:
//   - Single-item Find methods return ErrInvitationNotFound.
//   - List methods return an empty slice when no records match.
type InvitationRepository interface {
	// Create persists a new invitation.
	Create(ctx context.Context, invitation *entity.OrgInvitation) error

	// FindByToken returns the invitation with the given token.
	// Returns ErrInvitationNotFound if not found.
	FindByToken(ctx context.Context, token string) (*entity.OrgInvitation, error)

	// FindByID returns the invitation with the given ID.
	// Returns ErrInvitationNotFound if not found.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.OrgInvitation, error)

	// FindPendingByOrgID returns all pending invitations for an org.
	// Returns empty slice if none exist.
	FindPendingByOrgID(ctx context.Context, orgID uuid.UUID) ([]*entity.OrgInvitation, error)

	// FindPendingByEmail returns all pending invitations for an email address.
	// Returns empty slice if none exist.
	FindPendingByEmail(ctx context.Context, email string) ([]*entity.OrgInvitation, error)

	// Update updates an existing invitation (status changes).
	Update(ctx context.Context, invitation *entity.OrgInvitation) error
}

// OrgAuditRepository defines the interface for org audit log persistence.
//
// Error handling contract:
//   - Append is write-only; never returns "not found".
//   - FindByOrgID returns an empty slice when no entries match.
type OrgAuditRepository interface {
	// Append creates a new audit log entry.
	Append(ctx context.Context, entry *entity.OrgAuditEntry) error

	// FindByOrgID returns audit entries for an org, ordered by created_at DESC.
	// Supports pagination via limit and offset.
	FindByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*entity.OrgAuditEntry, error)

	// CountByOrgID returns the total number of audit entries for an org.
	CountByOrgID(ctx context.Context, orgID uuid.UUID) (int, error)
}
