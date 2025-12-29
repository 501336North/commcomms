# ADR-001: Hybrid Database Architecture

**Date:** 2025-12-29
**Status:** Accepted
**Deciders:** Product/Engineering Team

## Context

CommComms needs to store:
1. **Relational data** - Users, communities, messages, governance (ACID required)
2. **Knowledge relationships** - Entity connections, semantic queries (graph traversal)
3. **Semantic search** - Find similar content by meaning (vector similarity)
4. **Real-time state** - Presence, sessions, rate limits (ephemeral)

No single database technology excels at all four.

## Decision

Use a hybrid database architecture:

| Store | Technology | Purpose |
|-------|------------|---------|
| Relational | PostgreSQL 15+ | Core entities, transactions, full-text search |
| Graph | Neo4j 5.x | Knowledge graph, entity relationships |
| Vector | pgvector extension | Semantic search embeddings |
| Cache | Redis | Sessions, presence, rate limiting |

### Integration Pattern

- PostgreSQL is the source of truth for all transactional data
- Neo4j mirrors entity/relationship data for graph queries
- pgvector stores embeddings for messages (attached to PostgreSQL)
- Redis stores ephemeral data only

### Consistency Model

- **PostgreSQL → Neo4j**: Eventually consistent via async jobs
- **PostgreSQL → pgvector**: Strongly consistent (same transaction)
- **Redis**: Best effort (recovery from PostgreSQL)

## Consequences

### Positive
- Each technology used for its strength
- PostgreSQL handles 80% of queries (simple, well-understood)
- Neo4j enables queries impossible in SQL ("find all entities mentioned by users who visited Lisbon")
- pgvector integrates cleanly with PostgreSQL (same query interface)
- Redis provides sub-millisecond presence lookups

### Negative
- Operational complexity: 4 databases to manage
- Sync logic between PostgreSQL and Neo4j requires careful handling
- Team needs expertise in multiple technologies
- Higher infrastructure cost

### Neutral
- Will use managed services where possible (Render PostgreSQL, Neo4j AuraDB)
- Need monitoring for sync lag between PostgreSQL and Neo4j
- Backup strategy must cover all four stores

## Alternatives Considered

### Alternative 1: PostgreSQL Only
- **Pros**: Simpler, single store, team familiar
- **Cons**: Graph queries would be inefficient (recursive CTEs), no native vector search
- **Rejected**: Knowledge graph queries are core to the product

### Alternative 2: MongoDB + Atlas Search
- **Pros**: Flexible schema, built-in search, vector support coming
- **Cons**: No ACID for multi-document, weaker relational support
- **Rejected**: Need ACID for governance votes and token operations

### Alternative 3: Neo4j Only
- **Pros**: Native graph, can store properties
- **Cons**: Not designed for high-volume transactional writes, no vector search
- **Rejected**: Messages need high write throughput

### Alternative 4: Supabase (PostgreSQL + pgvector)
- **Pros**: Managed, includes auth, real-time subscriptions
- **Cons**: Still need separate graph DB, vendor lock-in
- **Considered**: Good option if we drop knowledge graph requirement

## Related Decisions
- ADR-002: AI-Native Knowledge System (uses Neo4j heavily)
- ADR-005: Echo Bot (uses pgvector for retrieval)
