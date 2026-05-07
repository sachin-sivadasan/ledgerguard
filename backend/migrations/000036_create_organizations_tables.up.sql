-- Multi-User Team Access Model: Organizations, Members, Invitations, Audit Log
-- See DECISIONS.md ADR-034 and DATABASE_SCHEMA.md for full documentation.

-- 1. Organizations — top-level data-owning entity
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE,
    plan_tier VARCHAR(20) NOT NULL DEFAULT 'FREE'
        CHECK (plan_tier IN ('FREE', 'STARTER', 'PRO')),
    webhook_url TEXT,
    webhook_secret VARCHAR(64),
    sso_provider VARCHAR(50),
    sso_config JSONB,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Org Members — users belonging to an organization with roles
CREATE TABLE org_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'VIEWER'
        CHECK (role IN ('OWNER', 'ADMIN', 'VIEWER')),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'SUSPENDED')),
    notification_prefs JSONB DEFAULT '{"email": true, "push": true}'::jsonb,
    invited_by UUID REFERENCES users(id),
    suspended_by UUID REFERENCES users(id),
    suspended_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, user_id)
);

CREATE INDEX idx_org_members_user ON org_members(user_id);
CREATE INDEX idx_org_members_org ON org_members(org_id);
CREATE INDEX idx_org_members_status ON org_members(org_id, status);

-- 3. Org Invitations — pending invites
CREATE TABLE org_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'VIEWER'
        CHECK (role IN ('ADMIN', 'VIEWER')),
    invited_by UUID NOT NULL REFERENCES users(id),
    token VARCHAR(64) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'ACCEPTED', 'EXPIRED', 'REVOKED')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, email, status)
);

-- 4. Org Audit Log — who did what
CREATE TABLE org_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    target_type VARCHAR(30),
    target_id UUID,
    metadata JSONB,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_org ON org_audit_log(org_id, created_at DESC);
CREATE INDEX idx_audit_log_actor ON org_audit_log(actor_id, created_at DESC);

-- 5. Add org_id to partner_accounts (nullable first, backfilled later)
ALTER TABLE partner_accounts ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- 6. Add org_id to billing_subscriptions (nullable first)
ALTER TABLE billing_subscriptions ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- 7. Add org_id to api_keys (nullable first)
ALTER TABLE api_keys ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
