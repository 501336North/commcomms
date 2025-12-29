# CommComms v1 - Requirements Specification

**Feature:** AI-Native Community Platform for Digital Nomads
**Version:** 1.0
**Date:** 2025-12-29

---

## Overview

CommComms is an AI-native community platform where conversations degrade into durable knowledge. Target wedge: digital nomads seeking a persistent "home base" that works across time zones.

---

## Actors

| Actor | Description |
|-------|-------------|
| Guest | Unauthenticated visitor with invite link |
| Member | Authenticated community participant |
| Admin | Community administrator with elevated privileges |
| Owner | Community creator with full control |
| Echo (Bot) | AI assistant that surfaces past knowledge |
| System | Background processes (summarization, cleanup) |

---

## 1. Chat Service Requirements

### US-CHAT-001: Send Message
**As a** member
**I want to** send a message to a thread
**So that** I can communicate with my community

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-CHAT-001.1: Successful Message Send**
- Given I am an authenticated member
- When I submit a message with content
- Then the message is persisted
- And all online thread participants see it within 200ms
- And I receive confirmation of delivery

**AC-CHAT-001.2: Empty Message Prevention**
- Given I try to send an empty message
- When I submit
- Then I see error "Message cannot be empty"
- And no message is created

**AC-CHAT-001.3: Message Size Limit**
- Given I submit a message > 10,000 characters
- When I submit
- Then I see error "Message too long (max 10,000 characters)"

**AC-CHAT-001.4: Rate Limiting**
- Given I send > 30 messages per minute
- When I try to send another
- Then I see error "Slow down! Try again in X seconds"

---

### US-CHAT-002: Create Thread
**As a** member
**I want to** start a new thread
**So that** I can initiate a focused discussion

**Priority:** Must
**Estimate:** Small

#### Acceptance Criteria

**AC-CHAT-002.1: Thread Creation**
- Given I am in a channel
- When I create a thread with a title
- Then the thread is created
- And I am the first participant
- And other members can see and join

**AC-CHAT-002.2: Thread Title Required**
- Given I try to create a thread without a title
- When I submit
- Then I see error "Thread title required"

---

### US-CHAT-003: Real-Time Presence
**As a** member
**I want to** see who is currently online
**So that** I know if I'll get an immediate response

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-CHAT-003.1: Online Status Display**
- Given another member is connected via WebSocket
- When I view the member list
- Then I see them marked as "online"
- And I see their last activity time

**AC-CHAT-003.2: Offline Transition**
- Given a member disconnects
- When 30 seconds pass without reconnection
- Then they are marked as "offline"
- And their presence indicator updates for all viewers

**AC-CHAT-003.3: Typing Indicator**
- Given a member is typing in my thread
- When they are actively typing
- Then I see "[Name] is typing..." indicator
- And it disappears after 3 seconds of inactivity

---

### US-CHAT-004: Async Mode Detection
**As a** member
**I want to** know if my message will be seen immediately or later
**So that** I set correct expectations

**Priority:** Should
**Estimate:** Small

#### Acceptance Criteria

**AC-CHAT-004.1: Async Indicator**
- Given the recipient is offline
- When I send a message
- Then I see "They'll be notified" indicator
- And the message is queued for digest notification

**AC-CHAT-004.2: Real-Time Indicator**
- Given the recipient is online
- When I send a message
- Then I see "Delivered" indicator within 200ms

---

## 2. Knowledge Engine Requirements

### US-KNOW-001: Entity Extraction
**As a** system
**I want to** extract entities from messages
**So that** knowledge is automatically categorized

**Priority:** Must
**Estimate:** Large

#### Acceptance Criteria

**AC-KNOW-001.1: Location Extraction**
- Given a message contains "Best coworking in Lisbon"
- When the message is processed
- Then "Lisbon" is extracted as a Location entity
- And linked in the knowledge graph

**AC-KNOW-001.2: Topic Extraction**
- Given a message discusses "visa requirements"
- When processed
- Then "Visa" is extracted as a Topic entity

**AC-KNOW-001.3: Extraction Latency**
- Given a message is sent
- When entity extraction runs
- Then extraction completes within 5 seconds of message send

**AC-KNOW-001.4: Confidence Threshold**
- Given an entity is extracted with < 0.6 confidence
- When processing completes
- Then the entity is NOT added to the graph
- And logged for review

---

### US-KNOW-002: Thread Summarization
**As a** member
**I want to** see a summary of long threads
**So that** I can catch up quickly

**Priority:** Must
**Estimate:** Large

#### Acceptance Criteria

**AC-KNOW-002.1: Auto-Summarization Trigger**
- Given a thread has been inactive for 15 minutes
- And the thread has > 10 messages
- When the summarization job runs
- Then a summary is generated
- And displayed as a collapsible header

**AC-KNOW-002.2: Summary Quality**
- Given a thread about "Lisbon coworking recommendations"
- When summarized
- Then the summary includes key recommendations
- And mentions who recommended what
- And is < 300 words

**AC-KNOW-002.3: Summary Update**
- Given a summarized thread receives new messages
- When the thread goes quiet again
- Then the summary updates to include new information

**AC-KNOW-002.4: Summary Latency**
- Given a thread triggers summarization
- When processed
- Then the summary appears within 30 seconds

---

### US-KNOW-003: Knowledge Card Creation
**As a** system
**I want to** create knowledge cards from discussions
**So that** valuable information is preserved

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-KNOW-003.1: Card Attachment**
- Given a thread summary about "Lisbon"
- When the summary is created
- Then it is attached to the "Lisbon" entity card
- And visible when browsing Lisbon-related content

**AC-KNOW-003.2: Card Aggregation**
- Given multiple threads discuss "Lisbon coworking"
- When viewing the Lisbon card
- Then I see aggregated insights from all threads
- And links to source threads

---

## 3. Echo Bot Requirements

### US-ECHO-001: Question Detection
**As** Echo
**I want to** detect when someone asks a question
**So that** I can offer relevant knowledge

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-ECHO-001.1: Question Recognition**
- Given a message "What's the best coworking in Lisbon?"
- When processed
- Then it is classified as a question with > 0.8 confidence

**AC-ECHO-001.2: Non-Question Filtering**
- Given a message "The coworking in Lisbon is great"
- When processed
- Then it is NOT classified as a question
- And Echo does not respond

**AC-ECHO-001.3: Detection Latency**
- Given a question is asked
- When detected
- Then classification completes within 500ms

---

### US-ECHO-002: Knowledge Retrieval
**As** Echo
**I want to** find relevant past discussions
**So that** I can answer questions from community memory

**Priority:** Must
**Estimate:** Large

#### Acceptance Criteria

**AC-ECHO-002.1: Semantic Search**
- Given the question "Best place to work in Lisboa?"
- When searching knowledge base
- Then results include threads about "Lisbon coworking"
- And semantic similarity overcomes spelling variations

**AC-ECHO-002.2: Source Attribution**
- Given Echo finds relevant content
- When responding
- Then the response includes links to source threads
- And mentions who originally contributed

**AC-ECHO-002.3: Confidence Threshold**
- Given no relevant content exists (confidence < 0.7)
- When Echo evaluates
- Then Echo does NOT respond
- And logs "No confident answer found"

**AC-ECHO-002.4: Retrieval Latency**
- Given a question is detected
- When knowledge retrieval runs
- Then Echo responds within 2 seconds

---

### US-ECHO-003: Ephemeral Messages
**As a** member
**I want to** see Echo's answers without permanent clutter
**So that** human answers take precedence

**Priority:** Must
**Estimate:** Small

#### Acceptance Criteria

**AC-ECHO-003.1: TTL Application**
- Given Echo responds to a question
- When the message is created
- Then it has expires_at = now + community.echo_ttl_hours

**AC-ECHO-003.2: Expiration Cleanup**
- Given an Echo message has expired
- When the cleanup job runs
- Then the message is deleted
- And no longer visible to anyone

**AC-ECHO-003.3: Visual Distinction**
- Given Echo responds
- When viewing the thread
- Then Echo's message is visually distinct (different style/color)
- And marked as "From community memory"

---

## 4. Search Service Requirements

### US-SEARCH-001: Keyword Search
**As a** member
**I want to** search by keywords
**So that** I can find specific content

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-SEARCH-001.1: Full-Text Search**
- Given I search for "coworking Lisbon"
- When results return
- Then I see messages containing those words
- And results are ranked by relevance

**AC-SEARCH-001.2: Search Scope**
- Given I search within a community
- When results return
- Then only content from that community appears

---

### US-SEARCH-002: Semantic Search
**As a** member
**I want to** search by meaning, not just keywords
**So that** I find relevant content even with different wording

**Priority:** Must
**Estimate:** Large

#### Acceptance Criteria

**AC-SEARCH-002.1: Semantic Matching**
- Given I search "remote work spaces in Portugal"
- When results return
- Then I see content about "coworking in Lisbon"
- And other semantically related content

**AC-SEARCH-002.2: Combined Ranking**
- Given search executes
- When results return
- Then keyword matches and semantic matches are blended
- And the most relevant results appear first

---

### US-SEARCH-003: Entity Filtering
**As a** member
**I want to** filter search by entity
**So that** I can narrow results to specific topics/locations

**Priority:** Should
**Estimate:** Small

#### Acceptance Criteria

**AC-SEARCH-003.1: Location Filter**
- Given I filter by entity "Lisbon"
- When results return
- Then only content linked to Lisbon entity appears

---

## 5. Identity Service Requirements

### US-ID-001: User Registration
**As a** guest
**I want to** register with my email
**So that** I can join communities

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-ID-001.1: Email Registration**
- Given I have a valid invite link
- When I submit email and password
- Then my account is created
- And I am added to the inviting community
- And I receive a welcome email

**AC-ID-001.2: No Phone Required**
- Given the registration form
- When I view it
- Then there is NO phone number field
- And email is the only identifier required

**AC-ID-001.3: Password Requirements**
- Given I submit a password < 8 characters
- When I submit
- Then I see error "Password must be at least 8 characters"

**AC-ID-001.4: Email Uniqueness**
- Given an account exists with this email
- When I try to register
- Then I see error "Email already registered"

---

### US-ID-002: Pseudonymous Profile
**As a** new member
**I want to** create a pseudonymous handle
**So that** I control my identity

**Priority:** Must
**Estimate:** Small

#### Acceptance Criteria

**AC-ID-002.1: Handle Creation**
- Given I complete registration
- When I set up my profile
- Then I choose a unique handle
- And this is my public identifier

**AC-ID-002.2: Handle Uniqueness**
- Given the handle "nomad_sarah" exists
- When I try to use it
- Then I see error "Handle already taken"

**AC-ID-002.3: Handle Format**
- Given I enter a handle with spaces
- When I submit
- Then I see error "Handle can only contain letters, numbers, underscores"

---

### US-ID-003: Reputation Tracking
**As a** member
**I want to** build reputation over time
**So that** my contributions are recognized

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-ID-003.1: Contribution Points**
- Given I send a message that gets marked "helpful"
- When the action is recorded
- Then my reputation score increases

**AC-ID-003.2: Reputation Display**
- Given I view a member's profile
- When I see their reputation
- Then I see their total score
- And their rank within the community

**AC-ID-003.3: Reputation Decay Prevention**
- Given I was active 6 months ago but not recently
- When I view my reputation
- Then my score has NOT decayed
- And historical contributions are preserved

---

### US-ID-004: Invite-Only Access
**As an** admin
**I want to** control who joins my community
**So that** I maintain quality

**Priority:** Must
**Estimate:** Small

#### Acceptance Criteria

**AC-ID-004.1: Invite Link Generation**
- Given I am an admin
- When I generate an invite link
- Then I get a unique URL
- And I can set expiration (default: 7 days)
- And I can set max uses (default: unlimited)

**AC-ID-004.2: Invite Validation**
- Given an expired invite link
- When someone tries to register
- Then they see error "Invite link has expired"

---

## 6. Governance Service Requirements

### US-GOV-001: Create Proposal
**As a** member with sufficient reputation
**I want to** create a proposal
**So that** the community can vote on decisions

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-GOV-001.1: Proposal Creation**
- Given I have reputation >= threshold
- When I submit a proposal
- Then it is created with status "active"
- And all members can view and vote

**AC-GOV-001.2: Reputation Gate**
- Given I have reputation < threshold
- When I try to create a proposal
- Then I see error "Need more reputation to create proposals"

**AC-GOV-001.3: Proposal Content**
- Given I create a proposal
- When I fill the form
- Then I provide: title, description, options, voting period

---

### US-GOV-002: Cast Vote
**As a** member
**I want to** vote on proposals
**So that** I participate in governance

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-GOV-002.1: Reputation-Weighted Voting**
- Given my reputation is 100
- And another member's reputation is 50
- When we both vote
- Then my vote counts 2x theirs

**AC-GOV-002.2: Vote Recording**
- Given I cast a vote
- When the vote is processed
- Then it is recorded on-chain (Base L2)
- And I receive transaction confirmation

**AC-GOV-002.3: Vote Changing**
- Given I already voted
- When I try to change my vote
- Then my previous vote is replaced
- And only my final vote counts

---

### US-GOV-003: Proposal Resolution
**As** the system
**I want to** resolve proposals when complete
**So that** decisions are finalized

**Priority:** Must
**Estimate:** Small

#### Acceptance Criteria

**AC-GOV-003.1: Quorum Check**
- Given voting period ends
- When calculating results
- Then proposal passes only if quorum reached (default: 20% participation)

**AC-GOV-003.2: Result Recording**
- Given proposal resolves
- When finalized
- Then result is recorded on-chain
- And all members notified

---

## 7. Token Service Requirements

### US-TOKEN-001: Wallet Connection
**As a** member
**I want to** connect my crypto wallet
**So that** I can participate in token governance

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-TOKEN-001.1: WalletConnect Support**
- Given I click "Connect Wallet"
- When I select WalletConnect
- Then I can connect any compatible wallet
- And my address is linked to my profile

**AC-TOKEN-001.2: Coinbase Wallet Support**
- Given I click "Connect Wallet"
- When I select Coinbase Wallet
- Then I can connect via Coinbase Wallet
- And my address is linked

**AC-TOKEN-001.3: Address Display**
- Given I connected a wallet
- When I view my profile
- Then I see my truncated address (0x1234...5678)

---

### US-TOKEN-002: Community Token
**As an** owner
**I want to** deploy a community token
**So that** we can have token-gated access and governance

**Priority:** Must
**Estimate:** Large

#### Acceptance Criteria

**AC-TOKEN-002.1: Token Deployment**
- Given I am a community owner
- When I deploy a token
- Then an ERC-20 token is deployed on Base L2
- And I set: name, symbol, initial supply

**AC-TOKEN-002.2: Token Distribution**
- Given the token is deployed
- When I distribute tokens
- Then members receive tokens to their connected wallets

---

### US-TOKEN-003: Token-Weighted Voting
**As a** token holder
**I want to** vote with my tokens
**So that** stake-holders have proportional say

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-TOKEN-003.1: Balance Reflection**
- Given I hold 1000 tokens
- When I vote
- Then my voting power is 1000

**AC-TOKEN-003.2: Combined Weight**
- Given community uses hybrid voting
- When I vote
- Then my weight = reputation_weight + token_weight

---

## 8. Location Service Requirements

### US-LOC-001: Check-In
**As a** member
**I want to** check in to my current location
**So that** others know where I am

**Priority:** Must
**Estimate:** Small

#### Acceptance Criteria

**AC-LOC-001.1: Location Recording**
- Given I grant location permission
- When I check in
- Then my current city/country is recorded
- And visible on my profile

**AC-LOC-001.2: Privacy Control**
- Given I check in
- When I choose visibility
- Then I can set: precise, city-only, or country-only

---

### US-LOC-002: Nearby Query
**As a** member
**I want to** find who's nearby
**So that** I can meet up in person

**Priority:** Must
**Estimate:** Medium

#### Acceptance Criteria

**AC-LOC-002.1: Radius Search**
- Given I search "Who's in Lisbon?"
- When results return
- Then I see members checked in to Lisbon
- And how long they've been there

**AC-LOC-002.2: Privacy Respect**
- Given a member set visibility to "country-only"
- When I search their city
- Then they do NOT appear in city-level results

---

## Edge Cases

| Scenario | Expected Behavior |
|----------|-------------------|
| User loses connection mid-message | Message queued locally, sent on reconnect |
| AI summarization fails | Thread marked "pending summary", retried |
| Echo confidence exactly 0.7 | Echo responds (threshold is >=) |
| Token deployment fails | Transaction reverted, user notified, can retry |
| Vote cast after deadline | Vote rejected with error |
| Duplicate invite link use (at limit) | Second user sees "Invite link exhausted" |
| Empty search query | Return empty results, no error |
| Very long thread (1000+ messages) | Paginated load, progressive summarization |

---

## Non-Functional Requirements

### Performance
- Message delivery: < 200ms (real-time)
- Echo response: < 2s
- Summarization: < 30s after quiescence
- Search results: < 500ms

### Security
- Passwords: bcrypt (12 rounds)
- Sessions: JWT with 24h expiry, refresh tokens
- Rate limiting: Per endpoint, per user
- Input sanitization: All user input
- On-chain data: Immutable, transparent

### Scalability
- Support 10,000 concurrent WebSocket connections
- Handle 100 messages/second per community
- Knowledge graph: 1M+ nodes

### Availability
- 99.9% uptime for core messaging
- Graceful degradation if AI services fail

---

## Dependencies

| Dependency | Purpose | Fallback |
|------------|---------|----------|
| OpenAI/Anthropic API | Summarization, entity extraction | Queue and retry |
| Neo4j | Knowledge graph | Read-only mode |
| Base L2 | Token operations | Pending queue |
| Redis | Presence, caching | Direct DB queries |

---

## Out of Scope (v1)

- Native mobile apps
- Reputation badges
- Multi-language support
- Audio/video calls
- File storage beyond images
- Public community discovery
- Cross-community federation
- Advanced analytics

---

## Traceability Matrix

| Requirement | User Story | Service | Priority |
|-------------|-----------|---------|----------|
| Real-time messaging | US-CHAT-001 | Chat Service | Must |
| Thread management | US-CHAT-002 | Chat Service | Must |
| Presence tracking | US-CHAT-003 | Chat Service | Must |
| Entity extraction | US-KNOW-001 | Knowledge Engine | Must |
| Thread summarization | US-KNOW-002 | Knowledge Engine | Must |
| Question detection | US-ECHO-001 | Echo Bot | Must |
| Knowledge retrieval | US-ECHO-002 | Echo Bot | Must |
| Ephemeral messages | US-ECHO-003 | Echo Bot | Must |
| Keyword search | US-SEARCH-001 | Search Service | Must |
| Semantic search | US-SEARCH-002 | Search Service | Must |
| User registration | US-ID-001 | Identity Service | Must |
| Pseudonymous profiles | US-ID-002 | Identity Service | Must |
| Reputation tracking | US-ID-003 | Identity Service | Must |
| Proposal creation | US-GOV-001 | Governance Service | Must |
| Voting | US-GOV-002 | Governance Service | Must |
| Wallet connection | US-TOKEN-001 | Token Service | Must |
| Token deployment | US-TOKEN-002 | Token Service | Must |
| Location check-in | US-LOC-001 | Location Service | Must |
| Nearby search | US-LOC-002 | Location Service | Must |

---

*Requirements extracted: 2025-12-29*
*Source: DESIGN.md*
*Powered by OSS Dev Workflow*
