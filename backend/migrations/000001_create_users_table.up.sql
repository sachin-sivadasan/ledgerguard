-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid VARCHAR(128) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('OWNER', 'ADMIN')),
    plan_tier VARCHAR(20) DEFAULT 'FREE' CHECK (plan_tier IN ('FREE', 'PRO')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for Firebase UID lookups
CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);
