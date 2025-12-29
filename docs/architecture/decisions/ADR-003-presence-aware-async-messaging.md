# ADR-003: Presence-Aware Async Messaging

**Date:** 2025-12-29
**Status:** Accepted
**Deciders:** Product/Engineering Team

## Context

Digital nomads are scattered across time zones. The product must:

1. **Not create reply pressure** - Messages shouldn't demand immediate response
2. **Work async** - Most conversations are 90% async, 10% real-time bursts
3. **Adapt to availability** - Real-time when both online, async otherwise

Traditional chat apps (Slack, Discord) assume "always on" which leads to burnout.

## Decision

Implement presence-aware smart defaults:

### Presence Detection

```
User connects → WebSocket → Redis presence store
                         → PostgreSQL presence table (recovery)
```

- WebSocket connection = online
- Disconnect after 30s without heartbeat = offline
- Presence stored in Redis (fast), backed to PostgreSQL (recovery)

### Message Delivery Modes

| Sender | Recipient | Mode | Notification |
|--------|-----------|------|--------------|
| Online | Online | Real-time | Instant push |
| Online | Offline | Async | Batched digest |
| Any | Any | Thread marked "live" | All get push |

### UI Indicators

When composing a message:
- Recipient online: "Sending..." → "Delivered"
- Recipient offline: "They'll be notified" indicator
- No typing indicator if recipient is offline

### Notification Strategy

| Scenario | Notification |
|----------|--------------|
| Recipient online | WebSocket push (instant) |
| Recipient offline < 1hr | Push notification (delayed 5min) |
| Recipient offline > 1hr | Email digest (hourly/daily based on setting) |

### Summarization Trigger

- AI summarization triggers MORE aggressively on async threads
- Logic: If thread has no online participants for 15min AND 10+ messages → summarize
- This helps people catch up when they come online

## Consequences

### Positive
- No guilt for not responding immediately
- Reduces notification fatigue
- Adapts to natural conversation rhythms
- AI summaries help async catch-up

### Negative
- More complex than simple real-time
- Presence detection has edge cases (network issues)
- Need to tune notification delays

### Neutral
- WebSocket infrastructure required regardless
- Redis adds operational overhead but also used for other caching
- Users can override defaults per-community

## Alternatives Considered

### Alternative 1: Pure Real-Time (Slack Model)
- **Pros**: Simple, immediate delivery
- **Cons**: Creates always-on pressure, exhausting
- **Rejected**: Goes against core product thesis

### Alternative 2: Pure Async (Forum Model)
- **Pros**: No pressure, clear expectations
- **Cons**: Loses chat immediacy, feels slow
- **Rejected**: Want best of both worlds

### Alternative 3: User-Selected Mode (Toggle)
- **Pros**: Maximum control
- **Cons**: Cognitive load, inconsistent experience
- **Rejected**: Smart defaults should just work

### Alternative 4: Schedule-Based (Office Hours)
- **Pros**: Clear boundaries
- **Cons**: Too rigid for global nomads
- **Rejected**: Presence detection is more flexible

## Related Decisions
- ADR-001: Hybrid Database Architecture (Redis for presence)
- ADR-002: AI-Native Knowledge System (summarization on async threads)
