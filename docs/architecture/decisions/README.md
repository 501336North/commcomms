# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for CommComms.

## Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-001](ADR-001-hybrid-database-architecture.md) | Hybrid Database Architecture | Accepted | 2025-12-29 |
| [ADR-002](ADR-002-ai-native-knowledge-system.md) | AI-Native Knowledge System | Accepted | 2025-12-29 |
| [ADR-003](ADR-003-presence-aware-async-messaging.md) | Presence-Aware Async Messaging | Accepted | 2025-12-29 |
| [ADR-004](ADR-004-hybrid-dao-governance.md) | Hybrid DAO Governance | Accepted | 2025-12-29 |
| [ADR-005](ADR-005-echo-bot-ephemeral-messages.md) | Echo Bot Ephemeral Messages | Accepted | 2025-12-29 |

## Summary

### ADR-001: Hybrid Database Architecture
**Decision:** Use PostgreSQL (relational), Neo4j (graph), pgvector (embeddings), Redis (cache).
**Why:** Each technology excels at its specific use case. No single DB handles all requirements.

### ADR-002: AI-Native Knowledge System
**Decision:** Automatic entity extraction, thread summarization, and semantic search.
**Why:** Core thesis is "conversations → knowledge" without manual curation.

### ADR-003: Presence-Aware Async Messaging
**Decision:** Real-time when both online, async otherwise. Smart notification batching.
**Why:** Avoid always-on pressure. Digital nomads span many time zones.

### ADR-004: Hybrid DAO Governance
**Decision:** Reputation + optional token weighted voting. Results recorded on Base L2.
**Why:** Give members real ownership without pure plutocracy.

### ADR-005: Echo Bot Ephemeral Messages
**Decision:** AI bot answers questions from knowledge base. Responses expire after 24h.
**Why:** Surface past knowledge automatically. Human answers take precedence.

## How to Add an ADR

1. Create new file: `ADR-XXX-kebab-case-title.md`
2. Use the template in any existing ADR
3. Set status to "Proposed"
4. Add to this README index
5. After review, update status to "Accepted"

## ADR Lifecycle

```
Proposed → Accepted → Superseded/Deprecated
```

- **Proposed**: Under discussion
- **Accepted**: Active decision
- **Deprecated**: No longer relevant
- **Superseded by ADR-XXX**: Replaced by newer decision

---

*Last updated: 2025-12-29*
