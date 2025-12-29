# CommComms Data Model

## Overview

CommComms uses a hybrid data architecture:

| Store | Technology | Purpose |
|-------|------------|---------|
| **Relational** | PostgreSQL 15+ | Core entities, transactions, auth |
| **Knowledge Graph** | Neo4j 5.x | Entity relationships, semantic queries |
| **Vector Store** | pgvector | Semantic search embeddings |
| **Cache** | Redis | Sessions, presence, rate limiting |

## PostgreSQL Schema

Located at: `backend/internal/db/schema.sql`

### Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                   USERS                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌───────────────┐         ┌─────────────────┐         ┌─────────────────┐ │
│   │    users      │────────▶│ refresh_tokens  │         │    wallets      │ │
│   ├───────────────┤   1:N   ├─────────────────┤         ├─────────────────┤ │
│   │ id (PK)       │         │ id (PK)         │         │ id (PK)         │ │
│   │ email         │         │ user_id (FK)    │         │ user_id (FK)    │ │
│   │ password_hash │         │ token_hash      │         │ address         │ │
│   │ handle        │         │ expires_at      │         │ chain_id        │ │
│   │ display_name  │◀────────│ created_at      │         │ connected_at    │ │
│   │ reputation    │    1:1  └─────────────────┘         └─────────────────┘ │
│   └───────────────┘                                            ▲             │
│          │                                                     │ 1:1         │
│          │                                                     │             │
└──────────┼─────────────────────────────────────────────────────┼─────────────┘
           │                                                     │
           │ N:M                                                 │
           ▼                                                     │
┌─────────────────────────────────────────────────────────────────────────────┐
│                              COMMUNITIES                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌───────────────┐         ┌───────────────────┐       ┌─────────────────┐ │
│   │  communities  │◀───────▶│ community_members │       │     invites     │ │
│   ├───────────────┤   N:M   ├───────────────────┤       ├─────────────────┤ │
│   │ id (PK)       │         │ community_id (FK) │       │ id (PK)         │ │
│   │ name          │         │ user_id (FK)      │       │ community_id    │ │
│   │ description   │         │ role              │       │ code            │ │
│   │ is_private    │         │ reputation_in_... │       │ max_uses        │ │
│   │ echo_enabled  │         │ joined_at         │       │ expires_at      │ │
│   │ echo_ttl_hrs  │         └───────────────────┘       └─────────────────┘ │
│   └───────────────┘                                                          │
│          │                                                                   │
│          │ 1:N                                                               │
│          ▼                                                                   │
│   ┌───────────────┐         ┌───────────────────┐                           │
│   │   channels    │────────▶│ community_tokens  │                           │
│   ├───────────────┤   1:1   ├───────────────────┤                           │
│   │ id (PK)       │         │ id (PK)           │                           │
│   │ community_id  │         │ community_id (FK) │                           │
│   │ name          │         │ contract_address  │                           │
│   │ description   │         │ name, symbol      │                           │
│   └───────────────┘         │ total_supply      │                           │
│          │                  └───────────────────┘                           │
│          │ 1:N                                                               │
└──────────┼───────────────────────────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                               THREADS                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌───────────────┐         ┌─────────────────────┐                         │
│   │    threads    │◀───────▶│ thread_participants │                         │
│   ├───────────────┤   N:M   ├─────────────────────┤                         │
│   │ id (PK)       │         │ thread_id (FK)      │                         │
│   │ channel_id    │         │ user_id (FK)        │                         │
│   │ author_id     │         │ last_read_at        │                         │
│   │ title         │         │ joined_at           │                         │
│   │ message_count │         └─────────────────────┘                         │
│   │ last_msg_at   │                                                          │
│   └───────────────┘                                                          │
│          │                  ┌─────────────────────┐                         │
│          │ 1:N             │  thread_summaries   │                         │
│          │                  ├─────────────────────┤                         │
│          ├─────────────────▶│ id (PK)             │                         │
│          │            1:1   │ thread_id (FK)      │                         │
│          │                  │ summary             │                         │
│          │                  │ key_points (JSONB)  │                         │
│          ▼                  │ entity_refs (JSONB) │                         │
│   ┌───────────────┐         └─────────────────────┘                         │
│   │   messages    │                   │                                      │
│   ├───────────────┤                   │ 1:N                                  │
│   │ id (PK)       │                   ▼                                      │
│   │ thread_id     │         ┌─────────────────────┐                         │
│   │ author_id     │         │ summary_entity_links│ ──────▶ Neo4j           │
│   │ content       │         ├─────────────────────┤                         │
│   │ is_echo       │         │ summary_id (FK)     │                         │
│   │ expires_at    │         │ entity_id           │                         │
│   │ content_tsv   │         │ entity_type         │                         │
│   └───────────────┘         │ entity_name         │                         │
│          │                  └─────────────────────┘                         │
│          │ 1:N                                                               │
│          ▼                                                                   │
│   ┌───────────────┐         ┌─────────────────────┐                         │
│   │   reactions   │         │ message_embeddings  │ ──────▶ pgvector        │
│   ├───────────────┤         ├─────────────────────┤                         │
│   │ id (PK)       │         │ message_id (FK)     │                         │
│   │ message_id    │         │ embedding (1536)    │                         │
│   │ user_id       │         └─────────────────────┘                         │
│   │ emoji         │                                                          │
│   └───────────────┘                                                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                              GOVERNANCE                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌───────────────┐         ┌─────────────────────┐                         │
│   │   proposals   │────────▶│  proposal_options   │                         │
│   ├───────────────┤   1:N   ├─────────────────────┤                         │
│   │ id (PK)       │         │ id (PK)             │                         │
│   │ community_id  │         │ proposal_id (FK)    │◀────┐                   │
│   │ author_id     │         │ label               │     │                   │
│   │ title         │         │ vote_weight         │     │ N:1               │
│   │ description   │         │ vote_count          │     │                   │
│   │ status        │         └─────────────────────┘     │                   │
│   │ voting_ends   │                                      │                   │
│   │ on_chain_tx   │         ┌─────────────────────┐     │                   │
│   └───────────────┘         │       votes         │─────┘                   │
│                             ├─────────────────────┤                         │
│                             │ id (PK)             │                         │
│                             │ proposal_id (FK)    │                         │
│                             │ option_id (FK)      │                         │
│                             │ user_id (FK)        │                         │
│                             │ weight              │                         │
│                             │ tx_hash             │                         │
│                             └─────────────────────┘                         │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                              LOCATION                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌───────────────┐                                                          │
│   │   check_ins   │                                                          │
│   ├───────────────┤                                                          │
│   │ id (PK)       │                                                          │
│   │ user_id (FK)  │                                                          │
│   │ city          │                                                          │
│   │ country       │                                                          │
│   │ latitude      │                                                          │
│   │ longitude     │                                                          │
│   │ visibility    │                                                          │
│   │ is_current    │                                                          │
│   └───────────────┘                                                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                              SYSTEM                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌───────────────┐         ┌─────────────────────┐       ┌───────────────┐ │
│   │  async_jobs   │         │  reputation_events  │       │   presence    │ │
│   ├───────────────┤         ├─────────────────────┤       ├───────────────┤ │
│   │ id (PK)       │         │ id (PK)             │       │ user_id (PK)  │ │
│   │ type          │         │ user_id (FK)        │       │ community_id  │ │
│   │ status        │         │ community_id (FK)   │       │ last_seen_at  │ │
│   │ payload       │         │ event_type          │       │ socket_id     │ │
│   │ result        │         │ points              │       └───────────────┘ │
│   │ error         │         │ reference_id        │                         │
│   └───────────────┘         └─────────────────────┘                         │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Neo4j Knowledge Graph

Located at: `backend/internal/db/neo4j_schema.cypher`

### Graph Structure

```
                              ┌─────────────────┐
                              │  CommunityRef   │
                              │  (pgId, name)   │
                              └────────▲────────┘
                                       │
                                 BELONGS_TO
                                       │
┌─────────────┐   RELATED_TO   ┌──────┴───────┐   RELATED_TO   ┌─────────────┐
│   Topic     │◀──────────────▶│   Location   │◀──────────────▶│   Person    │
│             │                │              │                │             │
│ Coworking   │                │   Lisbon     │                │  John Doe   │
│ Visa        │                │   Portugal   │                │             │
│ Remote Work │                │   Chiang Mai │                │             │
└──────┬──────┘                └──────┬───────┘                └──────┬──────┘
       │                              │                               │
       │      ┌───────────────────────┼───────────────────────┐       │
       │      │                       │                       │       │
       │      │                 SUMMARIZES                    │       │
       │      │                       │                       │       │
       │      │                ┌──────▼───────┐               │       │
       │      │                │   Summary    │               │       │
       │      │                │              │               │       │
       │      │                │ "Best cowork │               │       │
       │      │                │  in Lisbon"  │               │       │
       │      │                └──────┬───────┘               │       │
       │      │                       │                       │       │
       │      │                  FROM_THREAD                  │       │
       │      │                       │                       │       │
       │      │                ┌──────▼───────┐               │       │
       │      │                │  ThreadRef   │               │       │
       │      │                │   (pgId)     │               │       │
       │      │                └──────┬───────┘               │       │
       │      │                       │                       │       │
       │      │                   CONTAINS                    │       │
       │      │                       │                       │       │
       │      │    ┌──────────────────┼──────────────────┐    │       │
       │      │    │                  │                  │    │       │
       │      │    ▼                  ▼                  ▼    │       │
       │   ┌──┴────────┐      ┌───────────────┐   ┌──────────┴──┐    │
       │   │ MessageRef│      │  MessageRef   │   │  MessageRef │    │
       │   │  (pgId)   │      │    (pgId)     │   │   (pgId)    │    │
       │   └──────┬────┘      └───────┬───────┘   └──────┬──────┘    │
       │          │                   │                  │           │
       └───────── │ ──── MENTIONS ────│────── MENTIONS ──│───────────┘
                  │                   │                  │
                  └───── AUTHORED_BY ─┴─── AUTHORED_BY ──┘
                                      │
                                ┌─────▼─────┐
                                │  UserRef  │
                                │  (pgId)   │
                                └───────────┘
```

### Node Types

| Node | Labels | Key Properties |
|------|--------|----------------|
| Entity | `:Entity`, `:Location`/`:Topic`/`:Person` | id, name, type, mentionCount |
| Location | `:Entity`, `:Location` | city, country, countryCode, lat, lng |
| Topic | `:Entity`, `:Topic` | category |
| Person | `:Entity`, `:Person` | description |
| Summary | `:Summary` | id, threadId, content, keyPoints |
| ThreadRef | `:ThreadRef` | pgId, title |
| MessageRef | `:MessageRef` | pgId |
| UserRef | `:UserRef` | pgId, handle |
| CommunityRef | `:CommunityRef` | pgId, name |

### Relationship Types

| Relationship | From | To | Properties |
|--------------|------|-----|------------|
| MENTIONS | MessageRef | Entity | confidence, timestamp |
| SUMMARIZES | Summary | Entity | relevance |
| FROM_THREAD | Summary | ThreadRef | - |
| CONTAINS | ThreadRef | MessageRef | - |
| AUTHORED_BY | MessageRef/ThreadRef | UserRef | - |
| BELONGS_TO | Entity | CommunityRef | - |
| RELATED_TO | Entity | Entity | strength, type |
| IN_LOCATION | Location | Location | - |
| ABOUT | Topic | Topic | - |

## Data Flow

```
User sends message
        │
        ▼
┌───────────────────┐
│   PostgreSQL      │
│   messages table  │
└────────┬──────────┘
         │
         │  Async job triggered
         ▼
┌───────────────────┐     ┌───────────────────┐
│  Entity Extract   │────▶│      Neo4j        │
│  (AI service)     │     │  Knowledge Graph  │
└────────┬──────────┘     └───────────────────┘
         │
         │  Generate embedding
         ▼
┌───────────────────┐
│    pgvector       │
│   embeddings      │
└───────────────────┘
         │
         │  After thread quiescence
         ▼
┌───────────────────┐     ┌───────────────────┐
│   Summarization   │────▶│ thread_summaries  │
│   (AI service)    │     │ summary_entity_   │
└───────────────────┘     │ links → Neo4j     │
                          └───────────────────┘
```

## Table Summary

| Table | Purpose | Row Estimate |
|-------|---------|--------------|
| users | User accounts | 10K-100K |
| refresh_tokens | JWT refresh tokens | 50K-500K |
| wallets | Connected crypto wallets | 5K-50K |
| communities | Community definitions | 100-1K |
| community_members | User-community membership | 100K-1M |
| invites | Invite links | 10K-100K |
| channels | Community channels | 1K-10K |
| threads | Discussion threads | 100K-1M |
| thread_participants | Thread membership | 500K-5M |
| messages | Chat messages | 1M-100M |
| message_embeddings | Vector embeddings | 1M-100M |
| reactions | Message reactions | 500K-10M |
| thread_summaries | AI-generated summaries | 100K-1M |
| summary_entity_links | Summary-to-Neo4j links | 200K-2M |
| proposals | DAO proposals | 1K-10K |
| proposal_options | Voting options | 5K-50K |
| votes | Cast votes | 50K-500K |
| community_tokens | Deployed tokens | 100-1K |
| token_balance_cache | Cached balances | 100K-1M |
| check_ins | Location check-ins | 100K-1M |
| async_jobs | Background job queue | 10K-100K |
| presence | Online status | 1K-10K |
| reputation_events | Reputation audit log | 500K-5M |

## Indexes

### Performance-Critical Indexes

| Table | Index | Purpose |
|-------|-------|---------|
| messages | (thread_id, created_at) | Thread message pagination |
| messages | GIN(content_tsv) | Full-text search |
| message_embeddings | IVFFlat(embedding) | Semantic search |
| threads | (channel_id, last_message_at DESC) | Channel thread list |
| check_ins | (country, city) WHERE is_current | Nearby search |

### Unique Constraints

| Table | Columns | Purpose |
|-------|---------|---------|
| users | email | Prevent duplicate accounts |
| users | handle | Unique handles |
| wallets | address | One wallet per address |
| invites | code | Unique invite codes |
| votes | (proposal_id, user_id) | One vote per user |

## Migrations

Migrations are managed via `golang-migrate` or similar:

```bash
# Create new migration
migrate create -ext sql -dir migrations add_feature_x

# Run migrations
migrate -path migrations -database $DATABASE_URL up

# Rollback
migrate -path migrations -database $DATABASE_URL down 1
```

---

*Schema designed: 2025-12-29*
*Powered by OSS Dev Workflow*
