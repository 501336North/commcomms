# Implementation Notes: CommComms v1

**Feature:** AI-Native Community Platform for Digital Nomads

---

## Quick Reference

### Project Structure

```
commcomms/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go           # Entry point
│   ├── internal/
│   │   ├── ai/                   # AI client abstraction
│   │   ├── auth/                 # JWT, middleware
│   │   ├── chat/                 # Messages, threads, websocket
│   │   ├── community/            # Community management
│   │   ├── config/               # Configuration loading
│   │   ├── db/                   # Database connections, migrations
│   │   ├── echo/                 # Echo bot logic
│   │   ├── governance/           # Proposals, voting
│   │   ├── identity/             # Users, registration, reputation
│   │   ├── jobs/                 # Background job handlers
│   │   ├── knowledge/            # Entity extraction, summarization
│   │   ├── location/             # Check-ins, nearby queries
│   │   ├── presence/             # Online/offline tracking
│   │   ├── search/               # Keyword, semantic, ranking
│   │   └── token/                # Wallet, token deployment
│   ├── pkg/                      # Shared utilities
│   └── go.mod
├── contracts/                    # Solidity smart contracts
│   ├── src/
│   ├── test/
│   └── foundry.toml
├── docs/
│   ├── api/                      # OpenAPI spec, types
│   ├── architecture/decisions/   # ADRs
│   ├── data-model/               # Schema docs
│   └── requirements/             # User stories
└── .oss/dev/active/commcomms/   # Dev docs
```

---

## Key Commands

### Development

```bash
# Start database containers
docker-compose up -d postgres redis neo4j

# Run migrations
go run cmd/migrate/main.go

# Start server
go run cmd/server/main.go

# Run tests
go test -short ./...

# Run with race detector
go test -race ./...

# Generate mocks
go generate ./...
```

### Database

```bash
# Connect to PostgreSQL
psql postgresql://localhost:5432/commcomms

# Connect to Neo4j
cypher-shell -u neo4j -p password

# Connect to Redis
redis-cli
```

### Blockchain (Foundry)

```bash
cd contracts

# Build contracts
forge build

# Run tests
forge test

# Deploy to Base testnet
forge script script/Deploy.s.sol --rpc-url $BASE_SEPOLIA_RPC --broadcast
```

---

## Environment Variables

```bash
# Database
DATABASE_URL=postgresql://localhost:5432/commcomms
REDIS_URL=redis://localhost:6379
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password

# AI
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...

# Auth
JWT_SECRET=your-256-bit-secret

# Blockchain
BASE_RPC_URL=https://sepolia.base.org
PRIVATE_KEY=0x...

# Server
PORT=8080
ENV=development
```

---

## API Patterns

### Response Format

```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  }
}
```

### Error Format

```json
{
  "error": "Human readable message",
  "code": "ERROR_CODE",
  "details": { ... }
}
```

### WebSocket Events

```json
// Outgoing (to client)
{
  "type": "message:new",
  "payload": { "threadId": "...", "message": { ... } },
  "timestamp": "2025-01-01T00:00:00Z"
}

// Incoming (from client)
{
  "action": "subscribe",
  "threadId": "..."
}
```

---

## Testing Patterns

### Unit Test Template

```go
func TestServiceName_MethodName_Scenario(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewRepository(t)
    mockRepo.On("Method", mock.Anything).Return(expected, nil)
    svc := NewService(mockRepo)

    // Act
    result, err := svc.Method(input)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    mockRepo.AssertExpectations(t)
}
```

### Integration Test Template

```go
func TestRepository_Method_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    pool := testutil.SetupPostgres(t)
    repo := NewRepository(pool)

    // Test with real database
    result, err := repo.Method(input)

    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

---

## Performance Considerations

### Message Delivery Path

```
Client WebSocket → Hub.Broadcast → Thread Subscribers → WebSocket.Write
                                                              ↓
                                      Target: < 200ms total latency
```

**Optimization Points:**
- Use goroutine per client for writes
- Batch multiple messages per write (if within 10ms)
- Pre-allocate message buffers

### Echo Response Path

```
Question Detected → Knowledge Retrieval → AI Synthesis → Response
      (~500ms)           (~500ms)          (~1000ms)      (total < 2s)
```

**Optimization Points:**
- Parallel knowledge graph + vector search
- Cache frequent questions
- Use faster model for simple questions

---

## Common Gotchas

### PostgreSQL

1. **UUID generation:** Use `uuid_generate_v4()`, not Go's UUID
2. **Timestamps:** Always use `TIMESTAMPTZ`, store in UTC
3. **Soft deletes:** Check `deleted_at IS NULL` in all queries

### Neo4j

1. **Node IDs:** Neo4j IDs are internal; use our UUIDs
2. **Transactions:** Wrap multi-statement operations in transactions
3. **Indexes:** Create indexes before bulk imports

### WebSocket

1. **Ping/Pong:** Send pings every 30s to detect dead connections
2. **Message size:** Limit to 64KB to prevent memory issues
3. **Reconnection:** Client should auto-reconnect with exponential backoff

### AI APIs

1. **Rate limits:** OpenAI has per-minute token limits; implement queuing
2. **Timeouts:** Set 30s timeout for AI calls
3. **Fallback:** If OpenAI fails, try Anthropic

### Blockchain

1. **Gas estimation:** Add 20% buffer to estimated gas
2. **Nonce management:** Use a nonce manager for concurrent transactions
3. **Confirmation:** Wait for 1 block confirmation before considering done

---

## Debugging

### Common Issues

**"Connection refused" to database:**
- Check if Docker containers are running: `docker ps`
- Check if ports are correct in `.env`

**"Invalid token" errors:**
- Check JWT_SECRET is the same across restarts
- Check token expiry

**"Neo4j connection failed":**
- Neo4j takes longer to start; add retry logic
- Check authentication credentials

**"Rate limited by OpenAI":**
- Check TPM (tokens per minute) usage
- Implement request queuing

---

## Last Updated: 2025-12-29 by /oss:plan
