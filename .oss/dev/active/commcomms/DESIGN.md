# CommComms - Design Brief

**Tagline:** "Your conversations become your community's memory."

---

## Problem

Digital nomads lack a persistent "home base" community that:
- Works across time zones without creating reply pressure
- Retains valuable knowledge instead of letting it scroll away
- Provides real identity/reputation without phone numbers
- Gives members true ownership over governance

Current solutions fail:
- **WhatsApp/Telegram**: Chat entropy, no knowledge retention, phone-number identity
- **Discord/Slack**: Always-on pressure, chaotic at scale, weak ownership
- **Forums**: Too slow, not chat-native, poor mobile experience

---

## Solution

An AI-native community platform where **conversations degrade into knowledge**:
- Chat threads auto-summarize and attach to entities (locations, topics)
- Presence-aware messaging adapts to async/real-time based on who's online
- Reputation-weighted governance gives members real ownership
- Optional token layer for paid communities and treasuries

---

## Target Wedge

**Digital nomads** - but architected for broader verticals:
- Founder/operator communities
- Paid membership communities
- Professional guilds
- Learning cohorts

---

## Approach

### Architecture: Event-sourced, AI-augmented, graph-backed

| Layer | Technology | Purpose |
|-------|------------|---------|
| Frontend | Next.js (PWA) | Web-first, responsive, installable |
| Backend | Go | High-performance API, websocket handling |
| Database | PostgreSQL | Core relational data |
| Graph DB | Neo4j | Knowledge graph, semantic relationships |
| Vector Store | pgvector or Pinecone | Semantic search embeddings |
| Blockchain | Base (L2) | Token layer, DAO voting |
| AI | OpenAI/Anthropic API | Summarization, entity extraction |
| Real-time | WebSockets | Presence-aware messaging |
| Cache | Redis | Sessions, presence, rate limiting |

### Key Technical Decisions

1. **Entity extraction on write** - When messages are sent, AI extracts entities (locations, topics, people) and updates the knowledge graph in near-real-time

2. **Summarization on thread quiescence** - Threads auto-summarize after N minutes of inactivity, not on every message (cost control)

3. **Dual search** - Keyword search (Postgres full-text) + semantic search (vector similarity) combined for best results

4. **Hybrid on-chain/off-chain** - Votes recorded on-chain for transparency, chat data off-chain for speed/cost

---

## Echo - Community Knowledge Bot

A named bot persona that surfaces past knowledge when questions are asked.

**Behavior:**
- Triggers on question-like messages (AI detects intent)
- Searches knowledge graph + vector store for relevant past discussions
- Replies immediately with synthesized answer + source links
- Message is ephemeral - disappears after TTL (default: 24h)
- Community members can still answer; human answers persist

**Configuration (per-community):**
```
echo_ttl_hours: 24              # How long Echo messages persist
echo_enabled: true              # Toggle on/off
echo_confidence_threshold: 0.7  # Minimum confidence to reply
```

---

## Components

| Component | Responsibility |
|-----------|---------------|
| **Chat Service** | Real-time messaging, thread management, presence tracking |
| **Knowledge Engine** | Entity extraction, summarization, graph updates |
| **Echo Bot** | Question detection, knowledge retrieval, ephemeral replies |
| **Search Service** | Keyword + semantic search, result ranking |
| **Identity Service** | Auth, pseudonymous profiles, reputation tracking |
| **Governance Service** | DAO proposals, voting, reputation-weighted tallies |
| **Token Service** | Base L2 integration, wallet connection, token operations |
| **Location Service** | Check-ins, geo-tagging, "who's nearby" queries |

---

## Identity & Reputation

- **Pseudonymous but persistent** - No phone numbers, email-only registration
- **Reputation accrues over time** - Based on contributions, verifications, activity
- **Hybrid badges (v2)** - Earned, not claimed (e.g., "Lisbon Local," "Visa Wizard")

---

## Async Model

- **Presence-aware smart defaults** - Real-time when both online, async otherwise
- **AI summarization more aggressive on async threads**
- **Batched notifications** for offline users

---

## Governance

- **Hybrid DAO** - Reputation-based for daily decisions, optional token layer for paid communities
- **No forking** - Knowledge is collective property; you can leave but can't take the wiki
- **Tiered voting** - Higher reputation = more voting weight

---

## Business Model

- **Freemium + Premium communities**
- Free to join public communities
- Pay to create/join private communities
- Optional token layer for treasuries and advanced governance

---

## Success Criteria

### Core Thesis Validation
- [ ] Users search the knowledge base before asking (measurable: search-before-post ratio)
- [ ] Echo answers >30% of repeat questions correctly (user feedback: helpful/not helpful)
- [ ] Thread summaries are used (click-through rate on collapsed summaries)

### Engagement
- [ ] DAU/MAU ratio >40% (healthy async community benchmark)
- [ ] Average knowledge contribution per active user >2/month
- [ ] Communities retain >60% of members at 90 days

### Technical
- [ ] Message delivery <200ms (real-time path)
- [ ] Echo response <2s (knowledge retrieval)
- [ ] Summarization latency <30s after thread quiescence

### Business
- [ ] 10+ active communities within 3 months of launch
- [ ] 20% of communities upgrade to premium within 6 months
- [ ] Token layer adopted by 5+ communities in first year

---

## Out of Scope (v1)

Explicitly NOT building:

- Native mobile apps (PWA only)
- Reputation badges (v2)
- Multi-language support
- Audio/video calls
- File storage beyond images
- Public community discovery (invite-only for v1)
- Cross-community identity federation
- Advanced analytics dashboard

---

## TDD Test Plan

Tests to write FIRST, before any implementation:

### Chat Service
- [ ] User can send message to thread
- [ ] Message appears to all thread participants in real-time
- [ ] Presence status updates when user connects/disconnects
- [ ] Async mode triggers when recipient is offline

### Knowledge Engine
- [ ] Entity extraction identifies locations from message text
- [ ] Thread summarization triggers after 15min of inactivity
- [ ] Summary attaches to extracted entities in graph
- [ ] Summary updates when new messages added to thread

### Echo Bot
- [ ] Question detection identifies question-like messages
- [ ] Knowledge retrieval returns relevant past discussions
- [ ] Echo message has correct TTL and expires_at timestamp
- [ ] Expired Echo messages are cleaned up by background job

### Search Service
- [ ] Keyword search returns matching messages
- [ ] Semantic search returns conceptually similar content
- [ ] Combined ranking blends both result sets
- [ ] Entity filter narrows results (e.g., "Lisbon" only)

### Identity Service
- [ ] User can register with email (no phone)
- [ ] Pseudonymous profile created with unique handle
- [ ] Reputation score updates on contribution
- [ ] Invite link grants community access

### Governance Service
- [ ] Proposal can be created by eligible member
- [ ] Votes are weighted by reputation
- [ ] Proposal resolves when quorum reached
- [ ] Results recorded on-chain (Base L2)

### Token Service
- [ ] Wallet connection via WalletConnect/Coinbase Wallet
- [ ] Community token deployment on Base
- [ ] Token balance reflected in voting weight
- [ ] Transaction history viewable

### Location Service
- [ ] Check-in records current location
- [ ] "Who's nearby" query returns users within radius
- [ ] Location history contributes to badges (v2 prep)

---

*Design validated: 2025-12-29*
*Powered by OSS Dev Workflow*
