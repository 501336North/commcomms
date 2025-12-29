-- CommComms Database Schema
-- PostgreSQL 15+
--
-- This schema is for relational data. Knowledge graph lives in Neo4j.
-- Vector embeddings stored in pgvector extension.

-- ============================================
-- EXTENSIONS
-- ============================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";  -- For semantic search embeddings

-- ============================================
-- ENUMS
-- ============================================

CREATE TYPE member_role AS ENUM ('owner', 'admin', 'moderator', 'member');
CREATE TYPE proposal_status AS ENUM ('active', 'passed', 'rejected', 'expired');
CREATE TYPE location_visibility AS ENUM ('precise', 'city', 'country');
CREATE TYPE async_job_status AS ENUM ('pending', 'processing', 'completed', 'failed');

-- ============================================
-- USERS & AUTH
-- ============================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    handle VARCHAR(20) NOT NULL UNIQUE,
    display_name VARCHAR(50),
    bio VARCHAR(500),
    avatar_url VARCHAR(500),
    reputation INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  -- Soft delete
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_handle ON users(handle) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_reputation ON users(reputation DESC);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at);

-- ============================================
-- WALLETS (for token layer)
-- ============================================

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    address VARCHAR(42) NOT NULL UNIQUE,  -- Ethereum address
    chain_id INTEGER NOT NULL DEFAULT 8453,  -- Base L2
    connected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallets_address ON wallets(address);

-- ============================================
-- COMMUNITIES
-- ============================================

CREATE TABLE communities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL,
    description VARCHAR(500),
    is_private BOOLEAN NOT NULL DEFAULT FALSE,

    -- Echo bot settings
    echo_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    echo_ttl_hours INTEGER NOT NULL DEFAULT 24,
    echo_confidence_threshold NUMERIC(3,2) NOT NULL DEFAULT 0.70,

    -- Governance settings
    min_reputation_to_propose INTEGER NOT NULL DEFAULT 100,
    voting_quorum_percent INTEGER NOT NULL DEFAULT 20,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_communities_name ON communities(name);

CREATE TABLE community_members (
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role member_role NOT NULL DEFAULT 'member',
    reputation_in_community INTEGER NOT NULL DEFAULT 0,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (community_id, user_id)
);

CREATE INDEX idx_community_members_user ON community_members(user_id);
CREATE INDEX idx_community_members_reputation ON community_members(community_id, reputation_in_community DESC);

CREATE TABLE invites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    code VARCHAR(32) NOT NULL UNIQUE,
    created_by UUID NOT NULL REFERENCES users(id),
    max_uses INTEGER,  -- NULL = unlimited
    uses INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invites_code ON invites(code);
CREATE INDEX idx_invites_community ON invites(community_id);

-- ============================================
-- CHANNELS & THREADS
-- ============================================

CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channels_community ON channels(community_id);
CREATE UNIQUE INDEX idx_channels_community_name ON channels(community_id, name);

CREATE TABLE threads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(200) NOT NULL,
    message_count INTEGER NOT NULL DEFAULT 0,
    participant_count INTEGER NOT NULL DEFAULT 0,
    last_message_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_threads_channel ON threads(channel_id);
CREATE INDEX idx_threads_channel_activity ON threads(channel_id, last_message_at DESC);
CREATE INDEX idx_threads_author ON threads(author_id);

CREATE TABLE thread_participants (
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (thread_id, user_id)
);

CREATE INDEX idx_thread_participants_user ON thread_participants(user_id);

-- ============================================
-- MESSAGES
-- ============================================

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    is_echo BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ,  -- For ephemeral Echo messages

    -- For full-text search
    content_tsv TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', content)) STORED,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_messages_thread ON messages(thread_id, created_at);
CREATE INDEX idx_messages_author ON messages(author_id);
CREATE INDEX idx_messages_expires ON messages(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_messages_fulltext ON messages USING GIN(content_tsv);

-- Vector embedding for semantic search
CREATE TABLE message_embeddings (
    message_id UUID PRIMARY KEY REFERENCES messages(id) ON DELETE CASCADE,
    embedding vector(1536) NOT NULL,  -- OpenAI ada-002 dimensions
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_message_embeddings_vector ON message_embeddings USING ivfflat (embedding vector_cosine_ops);

CREATE TABLE reactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (message_id, user_id, emoji)
);

CREATE INDEX idx_reactions_message ON reactions(message_id);

-- ============================================
-- KNOWLEDGE (summaries, linked to Neo4j entities)
-- ============================================

CREATE TABLE thread_summaries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    thread_id UUID NOT NULL UNIQUE REFERENCES threads(id) ON DELETE CASCADE,
    summary TEXT NOT NULL,
    key_points JSONB NOT NULL DEFAULT '[]',  -- Array of strings
    entity_refs JSONB NOT NULL DEFAULT '[]',  -- Array of {id, type, name}
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    message_count_at_generation INTEGER NOT NULL
);

CREATE INDEX idx_thread_summaries_thread ON thread_summaries(thread_id);

-- Link between summaries and Neo4j entity IDs
CREATE TABLE summary_entity_links (
    summary_id UUID NOT NULL REFERENCES thread_summaries(id) ON DELETE CASCADE,
    entity_id VARCHAR(100) NOT NULL,  -- Neo4j node ID
    entity_type VARCHAR(20) NOT NULL,  -- 'location', 'topic', 'person'
    entity_name VARCHAR(100) NOT NULL,

    PRIMARY KEY (summary_id, entity_id)
);

CREATE INDEX idx_summary_entity_links_entity ON summary_entity_links(entity_id);

-- ============================================
-- GOVERNANCE
-- ============================================

CREATE TABLE proposals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    status proposal_status NOT NULL DEFAULT 'active',
    voting_ends_at TIMESTAMPTZ NOT NULL,

    -- On-chain data
    on_chain_tx_hash VARCHAR(66),  -- Transaction hash

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_proposals_community ON proposals(community_id);
CREATE INDEX idx_proposals_status ON proposals(community_id, status);
CREATE INDEX idx_proposals_voting_ends ON proposals(voting_ends_at) WHERE status = 'active';

CREATE TABLE proposal_options (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    proposal_id UUID NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
    label VARCHAR(200) NOT NULL,
    vote_weight NUMERIC(20,6) NOT NULL DEFAULT 0,
    vote_count INTEGER NOT NULL DEFAULT 0,
    position INTEGER NOT NULL
);

CREATE INDEX idx_proposal_options_proposal ON proposal_options(proposal_id);

CREATE TABLE votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    proposal_id UUID NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
    option_id UUID NOT NULL REFERENCES proposal_options(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    weight NUMERIC(20,6) NOT NULL,  -- Reputation + token weight
    tx_hash VARCHAR(66),  -- On-chain tx hash
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (proposal_id, user_id)
);

CREATE INDEX idx_votes_proposal ON votes(proposal_id);
CREATE INDEX idx_votes_user ON votes(user_id);

-- ============================================
-- TOKENS
-- ============================================

CREATE TABLE community_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    community_id UUID NOT NULL UNIQUE REFERENCES communities(id) ON DELETE CASCADE,
    contract_address VARCHAR(42) NOT NULL UNIQUE,
    name VARCHAR(50) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    total_supply VARCHAR(78) NOT NULL,  -- Store as string for big numbers
    decimals INTEGER NOT NULL DEFAULT 18,
    chain_id INTEGER NOT NULL DEFAULT 8453,  -- Base L2
    deploy_tx_hash VARCHAR(66) NOT NULL,
    deployed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_community_tokens_address ON community_tokens(contract_address);

-- Token balances are read from chain, but we cache for performance
CREATE TABLE token_balance_cache (
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance VARCHAR(78) NOT NULL,  -- Store as string for big numbers
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (community_id, user_id)
);

-- ============================================
-- LOCATION
-- ============================================

CREATE TABLE check_ins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    city VARCHAR(100),
    country VARCHAR(100) NOT NULL,
    country_code CHAR(2) NOT NULL,
    latitude NUMERIC(10,7),
    longitude NUMERIC(10,7),
    visibility location_visibility NOT NULL DEFAULT 'city',
    checked_in_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Only one active check-in per user
    is_current BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_check_ins_user ON check_ins(user_id) WHERE is_current = TRUE;
CREATE INDEX idx_check_ins_location ON check_ins(country, city) WHERE is_current = TRUE;
CREATE INDEX idx_check_ins_country ON check_ins(country_code) WHERE is_current = TRUE;

-- ============================================
-- ASYNC JOBS
-- ============================================

CREATE TABLE async_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,  -- 'summarize', 'embed', 'deploy_token', etc.
    status async_job_status NOT NULL DEFAULT 'pending',
    payload JSONB NOT NULL,
    result JSONB,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_async_jobs_status ON async_jobs(status) WHERE status IN ('pending', 'processing');
CREATE INDEX idx_async_jobs_type ON async_jobs(type, status);

-- ============================================
-- PRESENCE (ephemeral, but track in DB for recovery)
-- ============================================

CREATE TABLE presence (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    socket_id VARCHAR(100)
);

CREATE INDEX idx_presence_community ON presence(community_id);
CREATE INDEX idx_presence_stale ON presence(last_seen_at);

-- ============================================
-- REPUTATION EVENTS (for auditing)
-- ============================================

CREATE TABLE reputation_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    community_id UUID REFERENCES communities(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,  -- 'message_helpful', 'thread_created', etc.
    points INTEGER NOT NULL,
    reference_id UUID,  -- ID of related entity (message, thread, etc.)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reputation_events_user ON reputation_events(user_id);
CREATE INDEX idx_reputation_events_community ON reputation_events(community_id, user_id);

-- ============================================
-- FUNCTIONS & TRIGGERS
-- ============================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER communities_updated_at
    BEFORE UPDATE ON communities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Update thread stats on message insert
CREATE OR REPLACE FUNCTION update_thread_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE threads SET
            message_count = message_count + 1,
            last_message_at = NEW.created_at
        WHERE id = NEW.thread_id;
    ELSIF TG_OP = 'DELETE' AND OLD.deleted_at IS NULL THEN
        UPDATE threads SET
            message_count = message_count - 1
        WHERE id = OLD.thread_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER messages_thread_stats
    AFTER INSERT OR DELETE ON messages
    FOR EACH ROW
    EXECUTE FUNCTION update_thread_stats();

-- Cleanup expired Echo messages (run via pg_cron or application)
-- SELECT * FROM messages WHERE expires_at IS NOT NULL AND expires_at < NOW();
