# Technical Decisions: CommComms v1

**Feature:** AI-Native Community Platform for Digital Nomads

---

## Active Decisions

### DEC-001: TDD Methodology - London TDD

**Date:** 2025-12-29
**Status:** Decided

**Context:**
CommComms requires extensive testing of service interactions with databases, AI APIs, and blockchain. We need a consistent testing approach.

**Decision:**
Use **London TDD** (Mockist TDD):
- Mock collaborator dependencies
- Test behavior, not implementation
- Focus on message passing

**Rationale:**
- Faster test execution (no I/O)
- Better isolation of units
- Clearer API boundaries
- Easier to test error conditions

**Alternatives Considered:**
- **Detroit TDD:** Tests more implementation details, requires real dependencies
- **No specific methodology:** Inconsistent test quality

---

### DEC-002: Phase Order - Identity First

**Date:** 2025-12-29
**Status:** Decided

**Context:**
8 services to build, need to determine optimal order.

**Decision:**
Build in this order:
1. Identity (auth)
2. Chat (core messaging)
3. Knowledge + Location
4. Search + Echo
5. Governance + Token

**Rationale:**
- Identity is a hard dependency for all other services
- Chat is the core value proposition
- Knowledge/Search depend on messages existing
- Governance/Token are optional for MVP

---

### DEC-003: Database Strategy - Hybrid Approach

**Date:** 2025-12-29
**Status:** Decided

**Context:**
Need to store relational data (users, messages), knowledge graph (entities), and vector embeddings.

**Decision:**
- **PostgreSQL:** Core relational data, full-text search
- **Neo4j:** Knowledge graph, entity relationships
- **pgvector:** Vector embeddings (part of PostgreSQL)
- **Redis:** Presence, sessions, caching

**Rationale:**
- PostgreSQL is proven for transactional data
- Neo4j excels at graph traversal (entity relationships)
- pgvector avoids adding another database (Pinecone)
- Redis provides sub-millisecond reads for presence

**Reference:** ADR-001 (docs/architecture/decisions/)

---

### DEC-004: AI Provider Strategy

**Date:** 2025-12-29
**Status:** Decided

**Context:**
Need AI for entity extraction, summarization, question detection, and embeddings.

**Decision:**
- **Primary:** OpenAI (gpt-4o-mini for fast tasks, gpt-4o for complex)
- **Fallback:** Anthropic Claude (for redundancy)
- **Embeddings:** OpenAI text-embedding-3-small (1536 dimensions)

**Rationale:**
- OpenAI has mature API and good latency
- Fallback prevents single point of failure
- Smaller embedding model reduces storage costs

---

### DEC-005: Blockchain Choice - Base L2

**Date:** 2025-12-29
**Status:** Decided

**Context:**
Need blockchain for token deployment and vote recording.

**Decision:**
Use **Base L2** (Coinbase's Ethereum L2)

**Rationale:**
- Low gas fees (< $0.01 per transaction)
- Ethereum compatibility (ERC-20 tokens work natively)
- Growing ecosystem, good tooling
- Backed by Coinbase (reliability)

**Alternatives Considered:**
- **Ethereum mainnet:** Too expensive
- **Polygon:** Good but Base has better Coinbase wallet integration
- **Solana:** Different ecosystem, less EVM compatibility

**Reference:** ADR-004 (docs/architecture/decisions/)

---

### DEC-006: WebSocket Framework

**Date:** 2025-12-29
**Status:** Decided

**Context:**
Need real-time messaging with < 200ms latency.

**Decision:**
Use **gorilla/websocket** with custom hub pattern

**Rationale:**
- Battle-tested in production
- Good performance characteristics
- Full control over connection handling
- No vendor lock-in

**Alternatives Considered:**
- **Socket.IO:** Adds complexity, Go support is unofficial
- **Centrifugo:** Good but adds another service to manage
- **gRPC streaming:** Not ideal for browser clients

---

### DEC-007: Authentication - JWT + Refresh Tokens

**Date:** 2025-12-29
**Status:** Decided

**Context:**
Need stateless auth for API, with session management.

**Decision:**
- **Access Token:** JWT, 24h expiry, HS256 signing
- **Refresh Token:** Opaque token stored in DB, 7 day expiry
- **Storage:** Access token in memory, refresh token in httpOnly cookie

**Rationale:**
- Stateless auth scales well
- Refresh tokens allow revocation
- HttpOnly cookies prevent XSS theft

---

### DEC-008: Rate Limiting Strategy

**Date:** 2025-12-29
**Status:** Decided

**Context:**
Need to prevent abuse (message spam, API abuse).

**Decision:**
- **Message rate:** 30/minute per user
- **API rate:** 100/minute per user (general endpoints)
- **Storage:** Redis with sliding window algorithm

**Rationale:**
- Per-user limits are fair
- Sliding window is more accurate than fixed window
- Redis provides O(1) check time

---

## Pending Decisions

### DEC-009: Notification Strategy

**Status:** Pending

**Question:** How do we deliver notifications for async messages?

**Options:**
1. Email digests (batched)
2. Push notifications (PWA)
3. Third-party service (OneSignal, Firebase)

**Decision Date:** Phase 2

---

### DEC-010: Image Storage

**Status:** Pending

**Question:** Where do we store user-uploaded images?

**Options:**
1. S3 + CloudFront
2. Cloudinary
3. imgix

**Decision Date:** Phase 2

---

## Last Updated: 2025-12-29 by /oss:plan
