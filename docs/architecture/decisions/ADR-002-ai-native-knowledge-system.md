# ADR-002: AI-Native Knowledge System

**Date:** 2025-12-29
**Status:** Accepted
**Deciders:** Product/Engineering Team

## Context

The core thesis of CommComms is "conversations degrade into knowledge." This requires:

1. **Entity extraction** - Identify locations, topics, people from messages
2. **Thread summarization** - Generate concise summaries of discussions
3. **Knowledge cards** - Aggregate insights across threads by entity
4. **Semantic search** - Find content by meaning, not just keywords

This is inherently AI-powered and must happen without manual curation.

## Decision

Implement an AI-native knowledge system with:

### 1. Entity Extraction Pipeline

```
Message → AI (LLM) → Entities → Neo4j Knowledge Graph
                            ↓
                  PostgreSQL summary_entity_links
```

- Extract entities on every message (async job)
- Use structured output (JSON mode) for reliability
- Store entities in Neo4j with relationships
- Link back to PostgreSQL via `summary_entity_links`

### 2. Summarization Pipeline

```
Thread quiescent (15min) → AI Summarization → thread_summaries
                                           ↓
                                    Update Neo4j entities
```

- Trigger summarization after thread inactivity
- Update summary when thread receives new messages
- Summaries attach to entities in Neo4j

### 3. Dual Search System

```
Query → [Keyword Search (PostgreSQL)] → Combined
      → [Semantic Search (pgvector)]  → Ranking
```

- Keyword: PostgreSQL full-text search (tsvector)
- Semantic: pgvector similarity on embeddings
- Hybrid ranking blends both result sets

### AI Provider Strategy

- Primary: OpenAI API (GPT-4o for extraction, ada-002 for embeddings)
- Fallback: Anthropic Claude (for complex summarization)
- Cost control: Cache embeddings, batch entity extraction

## Consequences

### Positive
- Knowledge accumulates automatically (no manual wiki curation)
- Semantic search finds content by meaning
- Entity relationships enable "what do people say about Lisbon?"
- Summaries reduce time to catch up on long threads

### Negative
- AI costs scale with message volume (~$0.01-0.03 per message for full pipeline)
- Extraction quality varies (need confidence thresholds)
- Summarization latency (15min delay + 5-30s generation)
- AI hallucination risk in summaries

### Neutral
- Need monitoring for AI service availability
- Will cache embeddings (1536 dimensions per message)
- Entity extraction is async (not blocking message send)

## Alternatives Considered

### Alternative 1: Manual Curation (Wiki Model)
- **Pros**: Higher quality, no AI costs, no hallucination
- **Cons**: Requires user effort, won't scale, defeats the purpose
- **Rejected**: Core thesis is conversations → knowledge without effort

### Alternative 2: Keyword-Only Search
- **Pros**: Simple, fast, no AI
- **Cons**: Misses semantic similarity, can't answer "best coworking near beach"
- **Rejected**: Semantic understanding is key differentiator

### Alternative 3: Client-Side AI (Local LLMs)
- **Pros**: No API costs, privacy
- **Cons**: Quality varies, device constraints, inconsistent
- **Rejected**: Need consistent quality across all users

### Alternative 4: RAG Only (No Knowledge Graph)
- **Pros**: Simpler, just embeddings + retrieval
- **Cons**: Lose relationship queries, can't traverse "Lisbon → Portugal → Europe"
- **Rejected**: Graph relationships enable powerful queries

## Cost Analysis

Estimated costs per 1000 messages:
- Entity extraction (GPT-4o): ~$5-10
- Summarization (per 100 threads): ~$2-5
- Embeddings (ada-002): ~$0.10
- Neo4j queries: ~$0.01

Monthly estimate for 100K messages: ~$500-1000

## Related Decisions
- ADR-001: Hybrid Database Architecture
- ADR-005: Echo Bot Ephemeral Messages
