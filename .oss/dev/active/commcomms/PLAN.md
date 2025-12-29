# CommComms v1 - TDD Implementation Plan

**Feature:** AI-Native Community Platform for Digital Nomads
**Version:** 1.0
**Created:** 2025-12-29
**Design Reference:** `.oss/dev/active/commcomms/DESIGN.md`

---

## Summary

Build an AI-native community platform where conversations degrade into durable knowledge. This plan covers 5 implementation phases with strict TDD methodology, delivering 8 services: Identity, Chat, Location, Knowledge, Search, Echo, Governance, and Token.

---

## Service Dependency Graph

```
                           +---------------------+
                           |  Token Service      |
                           |  (blockchain layer) |
                           +----------+----------+
                                      | requires wallet from
                                      v
+-------------+    +-------------+    +-------------+
|  Governance |--->|  Identity   |<---|   Location  |
|   Service   |    |   Service   |    |   Service   |
+------+------+    +------+------+    +-------------+
       |                  |
       | reputation       | auth for all
       v                  v
+-----------------------------------------------------+
|                    Chat Service                      |
|         (messages, threads, presence)               |
+-------------------------+---------------------------+
                          | message data
        +-----------------+-----------------+
        v                 v                 v
+-------------+    +-------------+    +-------------+
|  Knowledge  |    |   Search    |    |  Echo Bot   |
|   Engine    |    |   Service   |    |             |
+-------------+    +-------------+    +-------------+
```

---

## Phase 1: Foundation - Identity Service

**Objective:** Establish core infrastructure and Identity service with full auth flow.

**Estimated Tasks:** 12

### Task 1.1: Project Scaffolding

**Objective:** Set up Go project structure with testing framework

**Tests to Write (RED step):**
- [ ] Test: `TestMainServerStarts` - server binary compiles and starts
  - File: `cmd/server/main_test.go`
  - Assertion: Server starts without panic on valid config

**Implementation (GREEN step):**
- File: `cmd/server/main.go`
- File: `go.mod`, `go.sum`
- Functions to create:
  - `main()` - Entry point with graceful shutdown

**Refactor (REFACTOR step):**
- [ ] Extract config loading to `internal/config/`
- [ ] Add structured logging

**Acceptance Criteria:**
- [ ] `go build ./cmd/server` succeeds
- [ ] `go test ./...` runs (even if no tests yet)

---

### Task 1.2: Database Connection Pool

**Objective:** Create PostgreSQL connection pool with health checks

**Tests to Write (RED step):**
- [ ] Test: `should connect to postgres with valid config`
  - File: `internal/db/postgres_test.go`
  - Assertion: `db.Ping()` returns nil
- [ ] Test: `should return error with invalid config`
  - File: `internal/db/postgres_test.go`
  - Assertion: Error contains connection failure message

**Implementation (GREEN step):**
- File: `internal/db/postgres.go`
- Functions to create:
  - `NewPostgresPool(cfg Config) (*pgxpool.Pool, error)`
  - `Close()` - Graceful connection cleanup

**Refactor (REFACTOR step):**
- [ ] Add connection retry with backoff
- [ ] Add metrics for connection pool stats

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Integration test connects to real PostgreSQL

---

### Task 1.3: Schema Migrations

**Objective:** Run database migrations on startup

**Tests to Write (RED step):**
- [ ] Test: `should apply all migrations in order`
  - File: `internal/db/migrate_test.go`
  - Assertion: Migration version matches expected
- [ ] Test: `should be idempotent on re-run`
  - File: `internal/db/migrate_test.go`
  - Assertion: Second migration run succeeds without changes

**Implementation (GREEN step):**
- File: `internal/db/migrate.go`
- File: `internal/db/migrations/001_initial.sql`
- Functions to create:
  - `RunMigrations(pool *pgxpool.Pool) error`

**Acceptance Criteria:**
- [ ] All tables from schema.sql created
- [ ] Migration is idempotent

---

### Task 1.4: User Registration

**Objective:** Allow users to register with email and invite code

**Tests to Write (RED step):**
- [ ] Test: `should register user with valid email and invite`
  - File: `internal/identity/service_test.go`
  - Assertion: User created with hashed password, reputation=0
- [ ] Test: `should reject registration without valid invite`
  - File: `internal/identity/service_test.go`
  - Assertion: Error "Invalid invite code"
- [ ] Test: `should reject duplicate email`
  - File: `internal/identity/service_test.go`
  - Assertion: Error "Email already registered"
- [ ] Test: `should require password >= 8 characters`
  - File: `internal/identity/service_test.go`
  - Assertion: Error "Password must be at least 8 characters"

**Implementation (GREEN step):**
- File: `internal/identity/service.go`
- Functions to create:
  - `Register(email, password, handle, inviteCode string) (*User, error)`
  - `validatePassword(password string) error`
  - `hashPassword(password string) (string, error)`

**Refactor (REFACTOR step):**
- [ ] Extract password validation to separate package
- [ ] Add input sanitization

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Password stored as bcrypt hash (12 rounds)
- [ ] User added to inviting community

---

### Task 1.5: Handle Validation

**Objective:** Ensure handles are unique and properly formatted

**Tests to Write (RED step):**
- [ ] Test: `should accept valid handle with letters and underscores`
  - File: `internal/identity/service_test.go`
  - Assertion: No error returned
- [ ] Test: `should reject handle with spaces`
  - File: `internal/identity/service_test.go`
  - Assertion: Error "Handle can only contain letters, numbers, underscores"
- [ ] Test: `should reject duplicate handle`
  - File: `internal/identity/service_test.go`
  - Assertion: Error "Handle already taken"
- [ ] Test: `should reject handle > 20 characters`
  - File: `internal/identity/service_test.go`
  - Assertion: Error "Handle must be 20 characters or less"

**Implementation (GREEN step):**
- File: `internal/identity/service.go`
- Functions to create:
  - `validateHandle(handle string) error`
  - `isHandleAvailable(handle string) (bool, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Handle uniqueness enforced at service and DB level

---

### Task 1.6: JWT Token Generation

**Objective:** Generate access and refresh tokens on login

**Tests to Write (RED step):**
- [ ] Test: `should generate valid access token with user claims`
  - File: `internal/auth/jwt_test.go`
  - Assertion: Token contains user_id, expires in 24h
- [ ] Test: `should generate refresh token with 7 day expiry`
  - File: `internal/auth/jwt_test.go`
  - Assertion: Token expires in 7 days
- [ ] Test: `should validate token signature`
  - File: `internal/auth/jwt_test.go`
  - Assertion: Invalid signature returns error
- [ ] Test: `should reject expired tokens`
  - File: `internal/auth/jwt_test.go`
  - Assertion: Expired token returns "Token expired" error

**Implementation (GREEN step):**
- File: `internal/auth/jwt.go`
- Functions to create:
  - `GenerateAccessToken(userID string) (string, error)`
  - `GenerateRefreshToken(userID string) (string, error)`
  - `ValidateToken(token string) (*Claims, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Tokens use HS256 signing

---

### Task 1.7: User Login

**Objective:** Authenticate users and return tokens

**Tests to Write (RED step):**
- [ ] Test: `should login with valid credentials`
  - File: `internal/identity/service_test.go`
  - Assertion: Returns access + refresh tokens
- [ ] Test: `should reject invalid password`
  - File: `internal/identity/service_test.go`
  - Assertion: Error "Invalid credentials"
- [ ] Test: `should reject non-existent email`
  - File: `internal/identity/service_test.go`
  - Assertion: Error "Invalid credentials" (same as above for security)

**Implementation (GREEN step):**
- File: `internal/identity/service.go`
- Functions to create:
  - `Login(email, password string) (*AuthResponse, error)`
  - `verifyPassword(hash, password string) bool`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Constant-time password comparison

---

### Task 1.8: Token Refresh

**Objective:** Extend sessions with refresh tokens

**Tests to Write (RED step):**
- [ ] Test: `should issue new tokens with valid refresh token`
  - File: `internal/identity/service_test.go`
  - Assertion: New access + refresh tokens returned
- [ ] Test: `should reject revoked refresh token`
  - File: `internal/identity/service_test.go`
  - Assertion: Error "Token revoked"
- [ ] Test: `should reject expired refresh token`
  - File: `internal/identity/service_test.go`
  - Assertion: Error "Token expired"

**Implementation (GREEN step):**
- File: `internal/identity/service.go`
- Functions to create:
  - `RefreshTokens(refreshToken string) (*AuthResponse, error)`
  - `revokeRefreshToken(tokenHash string) error`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Old refresh token invalidated on refresh

---

### Task 1.9: Auth Middleware

**Objective:** Protect routes with JWT authentication

**Tests to Write (RED step):**
- [ ] Test: `should allow request with valid token`
  - File: `internal/auth/middleware_test.go`
  - Assertion: Request passes through, user ID in context
- [ ] Test: `should reject request without token`
  - File: `internal/auth/middleware_test.go`
  - Assertion: 401 Unauthorized response
- [ ] Test: `should reject request with invalid token`
  - File: `internal/auth/middleware_test.go`
  - Assertion: 401 Unauthorized response

**Implementation (GREEN step):**
- File: `internal/auth/middleware.go`
- Functions to create:
  - `AuthMiddleware(next http.Handler) http.Handler`
  - `GetUserFromContext(ctx context.Context) (*User, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] User ID available in all protected route handlers

---

### Task 1.10: Invite Generation

**Objective:** Allow admins to create invite links

**Tests to Write (RED step):**
- [ ] Test: `should generate unique invite code`
  - File: `internal/identity/invite_test.go`
  - Assertion: Code is 32 characters, alphanumeric
- [ ] Test: `should set default expiry of 7 days`
  - File: `internal/identity/invite_test.go`
  - Assertion: expires_at = now + 7 days
- [ ] Test: `should respect custom max uses`
  - File: `internal/identity/invite_test.go`
  - Assertion: max_uses = specified value

**Implementation (GREEN step):**
- File: `internal/identity/invite.go`
- Functions to create:
  - `CreateInvite(communityID, creatorID string, opts InviteOptions) (*Invite, error)`
  - `generateInviteCode() string`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Invite URL format: `{baseURL}/join/{code}`

---

### Task 1.11: Invite Validation

**Objective:** Validate invite codes on registration

**Tests to Write (RED step):**
- [ ] Test: `should accept valid invite code`
  - File: `internal/identity/invite_test.go`
  - Assertion: Returns community ID
- [ ] Test: `should reject expired invite`
  - File: `internal/identity/invite_test.go`
  - Assertion: Error "Invite link has expired"
- [ ] Test: `should reject exhausted invite`
  - File: `internal/identity/invite_test.go`
  - Assertion: Error "Invite link exhausted"
- [ ] Test: `should increment use count on valid use`
  - File: `internal/identity/invite_test.go`
  - Assertion: uses = uses + 1

**Implementation (GREEN step):**
- File: `internal/identity/invite.go`
- Functions to create:
  - `ValidateInvite(code string) (*Community, error)`
  - `useInvite(code string) error`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Atomic increment prevents race conditions

---

### Task 1.12: Reputation Initialization

**Objective:** Track reputation events for users

**Tests to Write (RED step):**
- [ ] Test: `should initialize reputation to 0`
  - File: `internal/identity/reputation_test.go`
  - Assertion: New user has reputation = 0
- [ ] Test: `should record reputation event`
  - File: `internal/identity/reputation_test.go`
  - Assertion: Event stored with type, points, reference
- [ ] Test: `should not decay reputation over time`
  - File: `internal/identity/reputation_test.go`
  - Assertion: Old reputation unchanged after 6 months

**Implementation (GREEN step):**
- File: `internal/identity/reputation.go`
- Functions to create:
  - `GetReputation(userID string) (int, error)`
  - `RecordReputationEvent(userID, eventType string, points int, refID string) error`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Reputation events stored in reputation_events table

---

## Phase 2: Core Messaging - Chat Service

**Objective:** Real-time chat with presence awareness.

**Estimated Tasks:** 14

### Task 2.1: Community Creation

**Objective:** Allow users to create communities

**Tests to Write (RED step):**
- [ ] Test: `should create community with name and description`
  - File: `internal/community/service_test.go`
  - Assertion: Community created, creator is owner
- [ ] Test: `should set default settings`
  - File: `internal/community/service_test.go`
  - Assertion: echo_enabled=true, voting_quorum=20%
- [ ] Test: `should reject empty name`
  - File: `internal/community/service_test.go`
  - Assertion: Error "Community name required"

**Implementation (GREEN step):**
- File: `internal/community/service.go`
- Functions to create:
  - `Create(name, description string, creatorID string) (*Community, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Creator automatically added as owner

---

### Task 2.2: Channel Management

**Objective:** Create channels within communities

**Tests to Write (RED step):**
- [ ] Test: `should create channel in community`
  - File: `internal/chat/channel_test.go`
  - Assertion: Channel created with unique name in community
- [ ] Test: `should reject duplicate channel name in same community`
  - File: `internal/chat/channel_test.go`
  - Assertion: Error "Channel name already exists"
- [ ] Test: `should list channels for community`
  - File: `internal/chat/channel_test.go`
  - Assertion: Returns all channels in community

**Implementation (GREEN step):**
- File: `internal/chat/channel.go`
- Functions to create:
  - `CreateChannel(communityID, name, description string) (*Channel, error)`
  - `ListChannels(communityID string) ([]Channel, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Channel names unique per community

---

### Task 2.3: Thread Creation

**Objective:** Start new discussion threads

**Tests to Write (RED step):**
- [ ] Test: `should create thread with title`
  - File: `internal/chat/thread_test.go`
  - Assertion: Thread created, author is participant
- [ ] Test: `should reject empty title`
  - File: `internal/chat/thread_test.go`
  - Assertion: Error "Thread title required"
- [ ] Test: `should add initial message if provided`
  - File: `internal/chat/thread_test.go`
  - Assertion: First message created with thread

**Implementation (GREEN step):**
- File: `internal/chat/thread.go`
- Functions to create:
  - `CreateThread(channelID, title, initialMessage, authorID string) (*Thread, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Author automatically a participant

---

### Task 2.4: Message Sending

**Objective:** Send messages to threads

**Tests to Write (RED step):**
- [ ] Test: `should send message to thread`
  - File: `internal/chat/message_test.go`
  - Assertion: Message persisted, thread stats updated
- [ ] Test: `should reject empty message`
  - File: `internal/chat/message_test.go`
  - Assertion: Error "Message cannot be empty"
- [ ] Test: `should reject message > 10000 characters`
  - File: `internal/chat/message_test.go`
  - Assertion: Error "Message too long (max 10,000 characters)"
- [ ] Test: `should add sender as participant if not already`
  - File: `internal/chat/message_test.go`
  - Assertion: Sender in thread_participants table

**Implementation (GREEN step):**
- File: `internal/chat/message.go`
- Functions to create:
  - `SendMessage(threadID, content, authorID string) (*Message, error)`
  - `validateMessageContent(content string) error`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] message_count and last_message_at updated

---

### Task 2.5: Rate Limiting

**Objective:** Prevent message spam

**Tests to Write (RED step):**
- [ ] Test: `should allow 30 messages per minute`
  - File: `internal/chat/ratelimit_test.go`
  - Assertion: First 30 messages succeed
- [ ] Test: `should reject message 31 within a minute`
  - File: `internal/chat/ratelimit_test.go`
  - Assertion: Error "Slow down! Try again in X seconds"
- [ ] Test: `should reset after minute passes`
  - File: `internal/chat/ratelimit_test.go`
  - Assertion: Message succeeds after 60s

**Implementation (GREEN step):**
- File: `internal/chat/ratelimit.go`
- Functions to create:
  - `CheckRateLimit(userID string) error`
  - `RecordMessage(userID string) error`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Rate limiting per user, not global

---

### Task 2.6: WebSocket Hub

**Objective:** Manage WebSocket connections for real-time messaging

**Tests to Write (RED step):**
- [ ] Test: `should register client on connect`
  - File: `internal/chat/websocket/hub_test.go`
  - Assertion: Client in hub's client map
- [ ] Test: `should unregister client on disconnect`
  - File: `internal/chat/websocket/hub_test.go`
  - Assertion: Client removed from map
- [ ] Test: `should broadcast to all clients in thread`
  - File: `internal/chat/websocket/hub_test.go`
  - Assertion: All thread participants receive message

**Implementation (GREEN step):**
- File: `internal/chat/websocket/hub.go`
- Functions to create:
  - `NewHub() *Hub`
  - `Run()` - Main event loop
  - `Register(client *Client)`
  - `Unregister(client *Client)`
  - `Broadcast(threadID string, message []byte)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Thread-scoped broadcast

---

### Task 2.7: WebSocket Client Handler

**Objective:** Handle individual WebSocket connections

**Tests to Write (RED step):**
- [ ] Test: `should authenticate WebSocket connection`
  - File: `internal/chat/websocket/client_test.go`
  - Assertion: Valid JWT required for connection
- [ ] Test: `should subscribe to threads`
  - File: `internal/chat/websocket/client_test.go`
  - Assertion: Client added to thread's subscriber list
- [ ] Test: `should send typing indicator`
  - File: `internal/chat/websocket/client_test.go`
  - Assertion: Other participants see typing event

**Implementation (GREEN step):**
- File: `internal/chat/websocket/client.go`
- Functions to create:
  - `NewClient(conn *websocket.Conn, hub *Hub, userID string) *Client`
  - `ReadPump()` - Read messages from WebSocket
  - `WritePump()` - Write messages to WebSocket

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Ping/pong for connection health

---

### Task 2.8: Real-Time Message Delivery

**Objective:** Deliver messages to all thread participants within 200ms

**Tests to Write (RED step):**
- [ ] Test: `should deliver message to all online participants`
  - File: `internal/chat/delivery_test.go`
  - Assertion: All connected participants receive message event
- [ ] Test: `should deliver within 200ms`
  - File: `internal/chat/delivery_test.go`
  - Assertion: P95 latency < 200ms
- [ ] Test: `should include message details in event`
  - File: `internal/chat/delivery_test.go`
  - Assertion: Event contains message ID, content, author

**Implementation (GREEN step):**
- File: `internal/chat/delivery.go`
- Functions to create:
  - `DeliverMessage(threadID string, message *Message) error`
  - `formatMessageEvent(message *Message) []byte`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Performance test confirms < 200ms P95

---

### Task 2.9: Presence Tracking (Online)

**Objective:** Track when users are online

**Tests to Write (RED step):**
- [ ] Test: `should mark user online on WebSocket connect`
  - File: `internal/presence/service_test.go`
  - Assertion: User in presence table, last_seen_at = now
- [ ] Test: `should update last_seen_at on activity`
  - File: `internal/presence/service_test.go`
  - Assertion: last_seen_at updated on message send
- [ ] Test: `should list online users in community`
  - File: `internal/presence/service_test.go`
  - Assertion: Returns all users with recent last_seen_at

**Implementation (GREEN step):**
- File: `internal/presence/service.go`
- Functions to create:
  - `SetOnline(userID, communityID, socketID string) error`
  - `UpdateActivity(userID string) error`
  - `ListOnline(communityID string) ([]OnlineUser, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Presence stored in Redis (fallback: Postgres)

---

### Task 2.10: Presence Tracking (Offline)

**Objective:** Detect when users go offline

**Tests to Write (RED step):**
- [ ] Test: `should mark user offline on WebSocket disconnect`
  - File: `internal/presence/service_test.go`
  - Assertion: User removed from presence table
- [ ] Test: `should mark offline after 30s without heartbeat`
  - File: `internal/presence/service_test.go`
  - Assertion: Stale connections cleaned up
- [ ] Test: `should broadcast offline event to community`
  - File: `internal/presence/service_test.go`
  - Assertion: All community members see offline event

**Implementation (GREEN step):**
- File: `internal/presence/service.go`
- Functions to create:
  - `SetOffline(userID string) error`
  - `CleanupStale() error` - Background job

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Offline detection within 30 seconds

---

### Task 2.11: Typing Indicators

**Objective:** Show when users are typing

**Tests to Write (RED step):**
- [ ] Test: `should broadcast typing event to thread participants`
  - File: `internal/presence/typing_test.go`
  - Assertion: Other participants see typing indicator
- [ ] Test: `should auto-stop after 3 seconds of inactivity`
  - File: `internal/presence/typing_test.go`
  - Assertion: Typing indicator disappears after 3s
- [ ] Test: `should not show typing to self`
  - File: `internal/presence/typing_test.go`
  - Assertion: Sender doesn't receive own typing event

**Implementation (GREEN step):**
- File: `internal/presence/typing.go`
- Functions to create:
  - `StartTyping(userID, threadID string) error`
  - `StopTyping(userID, threadID string) error`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Typing indicators are ephemeral (not persisted)

---

### Task 2.12: Async Mode Detection

**Objective:** Indicate when messages will be delivered asynchronously

**Tests to Write (RED step):**
- [ ] Test: `should return async indicator when recipient offline`
  - File: `internal/chat/delivery_test.go`
  - Assertion: Response includes `async: true`
- [ ] Test: `should return real-time indicator when recipient online`
  - File: `internal/chat/delivery_test.go`
  - Assertion: Response includes `async: false`, `delivered_at`

**Implementation (GREEN step):**
- File: `internal/chat/delivery.go`
- Functions to create:
  - `GetDeliveryMode(threadID string) DeliveryMode`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Accurate presence check for delivery mode

---

### Task 2.13: Message Editing

**Objective:** Allow users to edit their messages

**Tests to Write (RED step):**
- [ ] Test: `should edit message content`
  - File: `internal/chat/message_test.go`
  - Assertion: Content updated, edited_at set
- [ ] Test: `should reject edit from non-author`
  - File: `internal/chat/message_test.go`
  - Assertion: Error "Only author can edit message"
- [ ] Test: `should broadcast edit event`
  - File: `internal/chat/message_test.go`
  - Assertion: Participants receive edit event

**Implementation (GREEN step):**
- File: `internal/chat/message.go`
- Functions to create:
  - `EditMessage(messageID, content, editorID string) (*Message, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Edit history not preserved (v1)

---

### Task 2.14: Message Deletion

**Objective:** Allow users to delete their messages

**Tests to Write (RED step):**
- [ ] Test: `should soft delete message`
  - File: `internal/chat/message_test.go`
  - Assertion: deleted_at set, content hidden
- [ ] Test: `should reject delete from non-author (unless admin)`
  - File: `internal/chat/message_test.go`
  - Assertion: Error or success based on role
- [ ] Test: `should broadcast delete event`
  - File: `internal/chat/message_test.go`
  - Assertion: Participants receive delete event

**Implementation (GREEN step):**
- File: `internal/chat/message.go`
- Functions to create:
  - `DeleteMessage(messageID, deleterID string) error`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Soft delete preserves audit trail

---

## Phase 3: Knowledge Layer

**Objective:** AI-powered entity extraction, summarization, and semantic search.

**Estimated Tasks:** 12

### Task 3.1: AI Client Abstraction

**Objective:** Create abstraction for AI API calls

**Tests to Write (RED step):**
- [ ] Test: `should call OpenAI completion API`
  - File: `internal/ai/client_test.go`
  - Assertion: Returns completion text
- [ ] Test: `should handle API errors gracefully`
  - File: `internal/ai/client_test.go`
  - Assertion: Retryable errors trigger retry
- [ ] Test: `should respect rate limits`
  - File: `internal/ai/client_test.go`
  - Assertion: 429 triggers backoff

**Implementation (GREEN step):**
- File: `internal/ai/client.go`
- Functions to create:
  - `NewAIClient(apiKey string) *AIClient`
  - `Complete(prompt string, opts Options) (string, error)`
  - `Embed(text string) ([]float64, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Supports both OpenAI and Anthropic

---

### Task 3.2: Entity Extraction - Locations

**Objective:** Extract location entities from messages

**Tests to Write (RED step):**
- [ ] Test: `should extract "Lisbon" from "Best coworking in Lisbon"`
  - File: `internal/knowledge/extractor_test.go`
  - Assertion: Returns Entity{type: "location", name: "Lisbon"}
- [ ] Test: `should extract multiple locations`
  - File: `internal/knowledge/extractor_test.go`
  - Assertion: "Paris and Berlin" returns 2 entities
- [ ] Test: `should reject low confidence extractions`
  - File: `internal/knowledge/extractor_test.go`
  - Assertion: Confidence < 0.6 not returned

**Implementation (GREEN step):**
- File: `internal/knowledge/extractor.go`
- Functions to create:
  - `ExtractEntities(message string) ([]Entity, error)`
  - `extractLocations(message string) ([]Entity, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Extraction completes within 5 seconds

---

### Task 3.3: Entity Extraction - Topics

**Objective:** Extract topic entities from messages

**Tests to Write (RED step):**
- [ ] Test: `should extract "Visa" from "visa requirements"`
  - File: `internal/knowledge/extractor_test.go`
  - Assertion: Returns Entity{type: "topic", name: "Visa"}
- [ ] Test: `should normalize topic names`
  - File: `internal/knowledge/extractor_test.go`
  - Assertion: "visas", "VISA" both normalize to "Visa"

**Implementation (GREEN step):**
- File: `internal/knowledge/extractor.go`
- Functions to create:
  - `extractTopics(message string) ([]Entity, error)`
  - `normalizeEntity(name string) string`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Topics are deduplicated

---

### Task 3.4: Neo4j Graph Repository

**Objective:** Store entities in knowledge graph

**Tests to Write (RED step):**
- [ ] Test: `should create location node`
  - File: `internal/knowledge/neo4j_repository_test.go`
  - Assertion: Node exists with correct properties
- [ ] Test: `should create relationship between thread and entity`
  - File: `internal/knowledge/neo4j_repository_test.go`
  - Assertion: MENTIONS relationship exists
- [ ] Test: `should increment mention count`
  - File: `internal/knowledge/neo4j_repository_test.go`
  - Assertion: mentionCount increases on duplicate

**Implementation (GREEN step):**
- File: `internal/knowledge/neo4j_repository.go`
- Functions to create:
  - `CreateOrUpdateEntity(entity Entity) error`
  - `LinkThreadToEntity(threadID, entityID string) error`
  - `GetEntityByName(name, entityType string) (*Entity, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Sync to Neo4j within 60s of extraction

---

### Task 3.5: Thread Summarization

**Objective:** Auto-summarize threads after quiescence

**Tests to Write (RED step):**
- [ ] Test: `should summarize thread with > 10 messages`
  - File: `internal/knowledge/summarizer_test.go`
  - Assertion: Summary generated, < 300 words
- [ ] Test: `should not summarize thread with < 10 messages`
  - File: `internal/knowledge/summarizer_test.go`
  - Assertion: No summary created
- [ ] Test: `should extract key points`
  - File: `internal/knowledge/summarizer_test.go`
  - Assertion: key_points array populated
- [ ] Test: `should complete within 30 seconds`
  - File: `internal/knowledge/summarizer_test.go`
  - Assertion: Summarization latency < 30s

**Implementation (GREEN step):**
- File: `internal/knowledge/summarizer.go`
- Functions to create:
  - `SummarizeThread(threadID string) (*ThreadSummary, error)`
  - `shouldSummarize(thread *Thread) bool`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Summary attached to entities in graph

---

### Task 3.6: Quiescence Detection

**Objective:** Trigger summarization after thread goes quiet

**Tests to Write (RED step):**
- [ ] Test: `should trigger after 15 minutes of inactivity`
  - File: `internal/knowledge/quiescence_test.go`
  - Assertion: Summarization job queued
- [ ] Test: `should reset timer on new message`
  - File: `internal/knowledge/quiescence_test.go`
  - Assertion: Timer resets, no premature trigger
- [ ] Test: `should not re-trigger if already summarized at same message count`
  - File: `internal/knowledge/quiescence_test.go`
  - Assertion: No duplicate summary

**Implementation (GREEN step):**
- File: `internal/knowledge/quiescence.go`
- Functions to create:
  - `StartQuiescenceTimer(threadID string)`
  - `CancelQuiescenceTimer(threadID string)`
  - `CheckQuiescence() []string` - Returns threads to summarize

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Timer persisted across server restarts

---

### Task 3.7: Message Embedding

**Objective:** Generate vector embeddings for messages

**Tests to Write (RED step):**
- [ ] Test: `should generate 1536-dimension embedding`
  - File: `internal/knowledge/embedding_test.go`
  - Assertion: Embedding vector has 1536 dimensions
- [ ] Test: `should store embedding in pgvector`
  - File: `internal/knowledge/embedding_test.go`
  - Assertion: message_embeddings row created
- [ ] Test: `should batch multiple messages`
  - File: `internal/knowledge/embedding_test.go`
  - Assertion: Batch of 10 messages embedded in single call

**Implementation (GREEN step):**
- File: `internal/knowledge/embedding.go`
- Functions to create:
  - `EmbedMessage(messageID, content string) error`
  - `BatchEmbed(messages []Message) error`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Embeddings created asynchronously

---

### Task 3.8: Keyword Search

**Objective:** Full-text search across messages

**Tests to Write (RED step):**
- [ ] Test: `should find messages containing "coworking"`
  - File: `internal/search/keyword_test.go`
  - Assertion: Results include matching messages
- [ ] Test: `should rank by relevance`
  - File: `internal/search/keyword_test.go`
  - Assertion: Best matches first
- [ ] Test: `should scope to community`
  - File: `internal/search/keyword_test.go`
  - Assertion: Only messages from specified community

**Implementation (GREEN step):**
- File: `internal/search/keyword.go`
- Functions to create:
  - `KeywordSearch(communityID, query string, opts Options) ([]SearchResult, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Uses PostgreSQL full-text search (ts_vector)

---

### Task 3.9: Semantic Search

**Objective:** Find conceptually similar content

**Tests to Write (RED step):**
- [ ] Test: `should find "Lisbon coworking" when searching "remote work Lisboa"`
  - File: `internal/search/semantic_test.go`
  - Assertion: Semantically similar results returned
- [ ] Test: `should return similarity score`
  - File: `internal/search/semantic_test.go`
  - Assertion: Each result has score 0-1
- [ ] Test: `should complete within 500ms`
  - File: `internal/search/semantic_test.go`
  - Assertion: Search latency < 500ms

**Implementation (GREEN step):**
- File: `internal/search/semantic.go`
- Functions to create:
  - `SemanticSearch(communityID, query string, opts Options) ([]SearchResult, error)`
  - `cosineSimilarity(a, b []float64) float64`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Uses pgvector for nearest neighbor search

---

### Task 3.10: Combined Search Ranking

**Objective:** Blend keyword and semantic results

**Tests to Write (RED step):**
- [ ] Test: `should combine keyword and semantic scores`
  - File: `internal/search/ranker_test.go`
  - Assertion: Final score = weighted blend
- [ ] Test: `should prefer exact keyword matches`
  - File: `internal/search/ranker_test.go`
  - Assertion: Exact match ranks higher
- [ ] Test: `should deduplicate results`
  - File: `internal/search/ranker_test.go`
  - Assertion: Same message not returned twice

**Implementation (GREEN step):**
- File: `internal/search/ranker.go`
- Functions to create:
  - `CombinedSearch(communityID, query string, opts Options) ([]SearchResult, error)`
  - `blendScores(keyword, semantic float64) float64`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Configurable blend weights

---

### Task 3.11: Entity Filtering

**Objective:** Filter search results by entity

**Tests to Write (RED step):**
- [ ] Test: `should filter by location entity`
  - File: `internal/search/filter_test.go`
  - Assertion: Only Lisbon-related results returned
- [ ] Test: `should filter by topic entity`
  - File: `internal/search/filter_test.go`
  - Assertion: Only Visa-related results returned

**Implementation (GREEN step):**
- File: `internal/search/filter.go`
- Functions to create:
  - `FilterByEntity(results []SearchResult, entityID string) []SearchResult`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Efficient filtering using Neo4j relationships

---

### Task 3.12: Location Check-in

**Objective:** Allow users to check in to locations

**Tests to Write (RED step):**
- [ ] Test: `should record check-in with coordinates`
  - File: `internal/location/service_test.go`
  - Assertion: Check-in created with city/country
- [ ] Test: `should respect visibility setting`
  - File: `internal/location/service_test.go`
  - Assertion: Only requested precision stored
- [ ] Test: `should only allow one active check-in`
  - File: `internal/location/service_test.go`
  - Assertion: Previous check-in marked inactive

**Implementation (GREEN step):**
- File: `internal/location/service.go`
- Functions to create:
  - `CheckIn(userID string, lat, lng float64, visibility string) (*CheckIn, error)`
  - `GetCurrentCheckIn(userID string) (*CheckIn, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Reverse geocoding for city/country

---

## Phase 4: Echo Bot

**Objective:** AI assistant that surfaces past knowledge.

**Estimated Tasks:** 6

### Task 4.1: Question Detection

**Objective:** Identify when messages are questions

**Tests to Write (RED step):**
- [ ] Test: `should detect "What's the best coworking?" as question`
  - File: `internal/echo/question_detector_test.go`
  - Assertion: is_question=true, confidence > 0.8
- [ ] Test: `should NOT detect "The coworking is great" as question`
  - File: `internal/echo/question_detector_test.go`
  - Assertion: is_question=false
- [ ] Test: `should detect implicit questions`
  - File: `internal/echo/question_detector_test.go`
  - Assertion: "Looking for visa advice" detected as question
- [ ] Test: `should complete within 500ms`
  - File: `internal/echo/question_detector_test.go`
  - Assertion: Detection latency < 500ms

**Implementation (GREEN step):**
- File: `internal/echo/question_detector.go`
- Functions to create:
  - `IsQuestion(message string) (bool, float64, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Fast classification using lightweight model

---

### Task 4.2: Knowledge Retrieval

**Objective:** Find relevant past discussions for questions

**Tests to Write (RED step):**
- [ ] Test: `should retrieve relevant threads for question`
  - File: `internal/echo/retriever_test.go`
  - Assertion: Returns threads with related content
- [ ] Test: `should include source attribution`
  - File: `internal/echo/retriever_test.go`
  - Assertion: Each result has thread ID and contributors
- [ ] Test: `should respect confidence threshold`
  - File: `internal/echo/retriever_test.go`
  - Assertion: Low confidence results filtered out

**Implementation (GREEN step):**
- File: `internal/echo/retriever.go`
- Functions to create:
  - `RetrieveKnowledge(question, communityID string) ([]KnowledgeResult, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Combines knowledge graph and semantic search

---

### Task 4.3: Answer Synthesis

**Objective:** Generate helpful answers from retrieved knowledge

**Tests to Write (RED step):**
- [ ] Test: `should synthesize answer from multiple sources`
  - File: `internal/echo/synthesizer_test.go`
  - Assertion: Answer includes information from all sources
- [ ] Test: `should include source links`
  - File: `internal/echo/synthesizer_test.go`
  - Assertion: Answer references source threads
- [ ] Test: `should mention original contributors`
  - File: `internal/echo/synthesizer_test.go`
  - Assertion: "According to @user..."

**Implementation (GREEN step):**
- File: `internal/echo/synthesizer.go`
- Functions to create:
  - `SynthesizeAnswer(question string, sources []KnowledgeResult) (string, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Answer is helpful and concise

---

### Task 4.4: Echo Response Delivery

**Objective:** Deliver Echo responses as ephemeral messages

**Tests to Write (RED step):**
- [ ] Test: `should create message with is_echo=true`
  - File: `internal/echo/service_test.go`
  - Assertion: Message marked as Echo
- [ ] Test: `should set expires_at based on community TTL`
  - File: `internal/echo/service_test.go`
  - Assertion: expires_at = now + echo_ttl_hours
- [ ] Test: `should respond within 2 seconds`
  - File: `internal/echo/service_test.go`
  - Assertion: Total latency < 2s

**Implementation (GREEN step):**
- File: `internal/echo/service.go`
- Functions to create:
  - `RespondToQuestion(messageID, threadID, communityID string) error`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] End-to-end response < 2s

---

### Task 4.5: Echo Configuration

**Objective:** Respect per-community Echo settings

**Tests to Write (RED step):**
- [ ] Test: `should not respond when echo_enabled=false`
  - File: `internal/echo/service_test.go`
  - Assertion: No Echo message created
- [ ] Test: `should respect confidence threshold`
  - File: `internal/echo/service_test.go`
  - Assertion: Low confidence = no response
- [ ] Test: `should use community-specific TTL`
  - File: `internal/echo/service_test.go`
  - Assertion: Different communities, different TTLs

**Implementation (GREEN step):**
- File: `internal/echo/config.go`
- Functions to create:
  - `GetEchoConfig(communityID string) (*EchoConfig, error)`
  - `ShouldRespond(confidence float64, config *EchoConfig) bool`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Per-community configuration respected

---

### Task 4.6: Echo Cleanup Job

**Objective:** Delete expired Echo messages

**Tests to Write (RED step):**
- [ ] Test: `should delete messages past expires_at`
  - File: `internal/jobs/echo_cleanup_test.go`
  - Assertion: Expired messages removed
- [ ] Test: `should run on schedule`
  - File: `internal/jobs/echo_cleanup_test.go`
  - Assertion: Job runs every minute
- [ ] Test: `should not delete non-Echo messages`
  - File: `internal/jobs/echo_cleanup_test.go`
  - Assertion: Regular messages preserved

**Implementation (GREEN step):**
- File: `internal/jobs/echo_cleanup.go`
- Functions to create:
  - `CleanupExpiredEchoMessages() (int, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Cleanup runs as background job

---

## Phase 5: Governance & Token

**Objective:** Community ownership with optional blockchain layer.

**Estimated Tasks:** 10

### Task 5.1: Proposal Creation

**Objective:** Allow qualified members to create proposals

**Tests to Write (RED step):**
- [ ] Test: `should create proposal with sufficient reputation`
  - File: `internal/governance/service_test.go`
  - Assertion: Proposal created with status "active"
- [ ] Test: `should reject proposal with insufficient reputation`
  - File: `internal/governance/service_test.go`
  - Assertion: Error "Need more reputation to create proposals"
- [ ] Test: `should require title, description, options`
  - File: `internal/governance/service_test.go`
  - Assertion: Validation errors for missing fields

**Implementation (GREEN step):**
- File: `internal/governance/service.go`
- Functions to create:
  - `CreateProposal(communityID, authorID string, req CreateProposalRequest) (*Proposal, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Reputation threshold configurable per community

---

### Task 5.2: Vote Casting

**Objective:** Allow members to vote on proposals

**Tests to Write (RED step):**
- [ ] Test: `should cast vote on active proposal`
  - File: `internal/governance/service_test.go`
  - Assertion: Vote recorded with weight
- [ ] Test: `should weight vote by reputation`
  - File: `internal/governance/service_test.go`
  - Assertion: Higher reputation = higher weight
- [ ] Test: `should reject vote on expired proposal`
  - File: `internal/governance/service_test.go`
  - Assertion: Error "Voting period has ended"

**Implementation (GREEN step):**
- File: `internal/governance/service.go`
- Functions to create:
  - `CastVote(proposalID, optionID, userID string) (*VoteReceipt, error)`
  - `calculateVoteWeight(userID, communityID string) (float64, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Vote weight = reputation + token balance

---

### Task 5.3: Vote Changing

**Objective:** Allow members to change their vote

**Tests to Write (RED step):**
- [ ] Test: `should replace previous vote`
  - File: `internal/governance/service_test.go`
  - Assertion: Only one vote per user per proposal
- [ ] Test: `should update option vote counts`
  - File: `internal/governance/service_test.go`
  - Assertion: Old option decremented, new option incremented

**Implementation (GREEN step):**
- File: `internal/governance/service.go`
- Functions to create:
  - `ChangeVote(proposalID, newOptionID, userID string) (*VoteReceipt, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Atomic vote change

---

### Task 5.4: Proposal Resolution

**Objective:** Resolve proposals when voting period ends

**Tests to Write (RED step):**
- [ ] Test: `should resolve after voting period`
  - File: `internal/governance/resolver_test.go`
  - Assertion: Status changes from "active" to "passed" or "rejected"
- [ ] Test: `should check quorum`
  - File: `internal/governance/resolver_test.go`
  - Assertion: Quorum not reached = "rejected"
- [ ] Test: `should determine winner by vote weight`
  - File: `internal/governance/resolver_test.go`
  - Assertion: Option with most weight wins

**Implementation (GREEN step):**
- File: `internal/governance/resolver.go`
- Functions to create:
  - `ResolveProposal(proposalID string) (*ProposalResult, error)`
  - `CheckQuorum(proposalID string) (bool, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Resolution runs as background job

---

### Task 5.5: Wallet Connection

**Objective:** Allow users to connect crypto wallets

**Tests to Write (RED step):**
- [ ] Test: `should verify wallet signature`
  - File: `internal/token/wallet_test.go`
  - Assertion: Address linked to user on valid signature
- [ ] Test: `should reject invalid signature`
  - File: `internal/token/wallet_test.go`
  - Assertion: Error "Invalid wallet signature"
- [ ] Test: `should store wallet address`
  - File: `internal/token/wallet_test.go`
  - Assertion: Wallet row created in database

**Implementation (GREEN step):**
- File: `internal/token/wallet.go`
- Functions to create:
  - `ConnectWallet(userID, address, signature, message string) (*Wallet, error)`
  - `verifySignature(address, signature, message string) bool`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Supports EIP-191 personal_sign

---

### Task 5.6: Token Deployment

**Objective:** Deploy community ERC-20 tokens

**Tests to Write (RED step):**
- [ ] Test: `should deploy token on Base L2`
  - File: `internal/token/service_test.go`
  - Assertion: Contract address returned
- [ ] Test: `should set token name and symbol`
  - File: `internal/token/service_test.go`
  - Assertion: Token metadata correct
- [ ] Test: `should only allow owner to deploy`
  - File: `internal/token/service_test.go`
  - Assertion: Non-owner gets error

**Implementation (GREEN step):**
- File: `internal/token/service.go`
- Functions to create:
  - `DeployToken(communityID, ownerID string, req DeployTokenRequest) (*TokenDeployment, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Uses Base L2 testnet initially

---

### Task 5.7: Token Distribution

**Objective:** Distribute tokens to community members

**Tests to Write (RED step):**
- [ ] Test: `should transfer tokens to member wallets`
  - File: `internal/token/service_test.go`
  - Assertion: Token balances updated
- [ ] Test: `should batch multiple distributions`
  - File: `internal/token/service_test.go`
  - Assertion: Single transaction for multiple recipients
- [ ] Test: `should require owner authorization`
  - File: `internal/token/service_test.go`
  - Assertion: Non-owner gets error

**Implementation (GREEN step):**
- File: `internal/token/service.go`
- Functions to create:
  - `DistributeTokens(communityID string, distributions []Distribution) error`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Gas-efficient batch transfers

---

### Task 5.8: Token Balance Cache

**Objective:** Cache token balances for fast lookup

**Tests to Write (RED step):**
- [ ] Test: `should cache balance after query`
  - File: `internal/token/cache_test.go`
  - Assertion: Balance stored in token_balance_cache
- [ ] Test: `should return cached balance`
  - File: `internal/token/cache_test.go`
  - Assertion: No blockchain call on cache hit
- [ ] Test: `should refresh cache periodically`
  - File: `internal/token/cache_test.go`
  - Assertion: Stale cache refreshed

**Implementation (GREEN step):**
- File: `internal/token/cache.go`
- Functions to create:
  - `GetBalance(communityID, userID string) (string, error)`
  - `RefreshBalances(communityID string) error`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Cache TTL: 5 minutes

---

### Task 5.9: On-Chain Vote Recording

**Objective:** Record vote results on blockchain

**Tests to Write (RED step):**
- [ ] Test: `should submit proposal result to chain`
  - File: `internal/governance/onchain_test.go`
  - Assertion: Transaction hash returned
- [ ] Test: `should store tx hash in proposal`
  - File: `internal/governance/onchain_test.go`
  - Assertion: on_chain_tx_hash populated
- [ ] Test: `should handle chain errors gracefully`
  - File: `internal/governance/onchain_test.go`
  - Assertion: Retry on failure, notify admin

**Implementation (GREEN step):**
- File: `internal/governance/onchain.go`
- Functions to create:
  - `RecordResultOnChain(proposal *Proposal, result *ProposalResult) (string, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Results are immutable once recorded

---

### Task 5.10: Token-Weighted Voting

**Objective:** Include token balance in vote weight

**Tests to Write (RED step):**
- [ ] Test: `should add token balance to vote weight`
  - File: `internal/governance/vote_calculator_test.go`
  - Assertion: Weight = reputation + token_balance
- [ ] Test: `should work when community has no token`
  - File: `internal/governance/vote_calculator_test.go`
  - Assertion: Weight = reputation only
- [ ] Test: `should work when user has no wallet`
  - File: `internal/governance/vote_calculator_test.go`
  - Assertion: Weight = reputation only

**Implementation (GREEN step):**
- File: `internal/governance/vote_calculator.go`
- Functions to create:
  - `CalculateCombinedWeight(userID, communityID string) (float64, error)`

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Configurable weight ratio (reputation vs token)

---

## Testing Strategy

### Unit Tests
- All service layer logic with mocked repositories
- Pure functions for business rules
- Target: 90% coverage

### Integration Tests
- PostgreSQL queries with test database
- Neo4j operations with test database
- Redis operations with test instance
- WebSocket connections with test client

### E2E Tests (with real external services)
- AI API calls (rate limited, cached)
- Blockchain operations (testnet only)
- Full user flows

### Performance Tests
- Message delivery < 200ms P95
- Echo response < 2s P95
- Search results < 500ms P95
- Summarization < 30s

---

## Security Checklist

- [ ] Passwords: bcrypt (12 rounds)
- [ ] Sessions: JWT with 24h expiry
- [ ] Rate limiting: Per endpoint, per user
- [ ] Input sanitization: All user input
- [ ] SQL injection: Parameterized queries
- [ ] XSS: Content sanitization
- [ ] CORS: Strict origin policy
- [ ] Secrets: Environment variables only
- [ ] Blockchain: Testnet for development

---

## Rollout Strategy

### Phase 1-2 (Weeks 1-6)
- Deploy Identity + Chat services
- Internal testing only
- No external users

### Phase 3-4 (Weeks 7-12)
- Add Knowledge + Echo
- Limited beta with 50 users
- Monitor AI costs and latency

### Phase 5 (Weeks 13-16)
- Add Governance + Token
- Expand beta to 500 users
- Testnet blockchain only

### Production (Week 17+)
- Full launch
- Mainnet token layer (optional per community)
- 99.9% uptime target

---

## Estimated Totals

| Phase | Tasks | Test Cases |
|-------|-------|------------|
| Phase 1: Identity | 12 | 36 |
| Phase 2: Chat | 14 | 42 |
| Phase 3: Knowledge | 12 | 36 |
| Phase 4: Echo | 6 | 18 |
| Phase 5: Governance/Token | 10 | 30 |
| **Total** | **54** | **162** |

---

*Plan created: 2025-12-29*
*Methodology: London TDD (mock collaborators, test behaviors)*
*Powered by OSS Dev Workflow*
