# ADR-005: Echo Bot with Ephemeral Messages

**Date:** 2025-12-29
**Status:** Accepted
**Deciders:** Product/Engineering Team

## Context

Repeat questions are common in communities:
- "Best coworking in Lisbon?" gets asked every month
- Valuable answers scroll away in chat history
- New members don't know to search before asking

We need a way to surface past knowledge automatically without cluttering the conversation permanently.

## Decision

Implement Echo Bot with ephemeral responses:

### Echo Bot Behavior

```
Message detected → Is it a question? (AI classifier)
                           ↓ Yes
                   Search knowledge base (pgvector + Neo4j)
                           ↓ Confidence >= threshold
                   Post ephemeral response
```

### Question Detection

- Use AI classifier to detect question intent
- Confidence threshold: 0.8 (high precision, avoid false positives)
- Ignore messages in "live" marked threads

### Knowledge Retrieval

1. **Semantic search** (pgvector) - Find similar past messages
2. **Graph query** (Neo4j) - Find related entities and summaries
3. **Combine and rank** - Blend results by recency and relevance
4. **Synthesize answer** - LLM generates response from sources

### Ephemeral Messages

| Property | Value |
|----------|-------|
| `is_echo` | true |
| `expires_at` | NOW() + community.echo_ttl_hours |
| Visual style | Distinct color, "From community memory" label |
| Cleanup | Background job deletes expired messages |

### Configuration (Per-Community)

```yaml
echo_enabled: true              # Toggle on/off
echo_ttl_hours: 24              # Default: 24h
echo_confidence_threshold: 0.7  # Minimum confidence to respond
```

### Why Ephemeral?

- **Human answers take precedence** - Echo bridges the gap, then fades
- **No clutter** - Old Echo responses don't accumulate
- **Encourages human engagement** - People can still answer, their answers persist
- **Humble AI** - Echo acknowledges it may be wrong by being temporary

## Consequences

### Positive
- Repeat questions get instant answers
- Knowledge base proves its value immediately
- Reduces burden on senior community members
- Creates positive feedback loop (more contributions → better Echo)

### Negative
- AI costs for every question (~$0.01-0.05)
- Risk of wrong/outdated answers (mitigated by ephemerality)
- May discourage some human interaction
- Cleanup job adds operational complexity

### Neutral
- Admins can disable Echo per-community
- TTL is tunable based on community preference
- Echo responses link to source threads (attribution)

## Alternatives Considered

### Alternative 1: Permanent Bot Responses (like ChatGPT)
- **Pros**: Answers persist, less AI calls
- **Cons**: Clutters conversation, outdated answers linger
- **Rejected**: Want human answers to dominate

### Alternative 2: Search Prompt (Not Auto-Response)
- **Pros**: User-initiated, no surprise responses
- **Cons**: Users don't search, repeat questions continue
- **Rejected**: Need to surface knowledge proactively

### Alternative 3: Pinned FAQ (Static)
- **Pros**: Curated, high quality
- **Cons**: Manual effort, doesn't scale, misses nuance
- **Rejected**: Want dynamic, AI-powered responses

### Alternative 4: Private Response (Only Asker Sees)
- **Pros**: Non-intrusive
- **Cons**: Others don't learn, no social proof
- **Rejected**: Visibility encourages knowledge contribution

## Implementation Details

### Echo Response Format

```markdown
**Echo** (from community memory)

Based on past discussions:
- [Summary of relevant knowledge]

Sources:
- [Link to thread 1]
- [Link to thread 2]

_This message will disappear in 24h. Human answers persist._
```

### Cleanup Job

```sql
-- Run every hour
DELETE FROM messages
WHERE is_echo = true
  AND expires_at < NOW();
```

### Rate Limiting

- Max 1 Echo response per thread per hour
- Max 10 Echo responses per community per hour
- Prevents Echo from dominating conversations

## Related Decisions
- ADR-001: Hybrid Database Architecture (pgvector, Neo4j)
- ADR-002: AI-Native Knowledge System (retrieval pipeline)
