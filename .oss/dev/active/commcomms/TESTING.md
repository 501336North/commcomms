# Testing Strategy: CommComms v1

**Feature:** AI-Native Community Platform for Digital Nomads

---

## TDD Methodology: London TDD

CommComms uses **London TDD** (Mockist TDD):
- Mock collaborator dependencies
- Test behavior, not implementation
- Focus on message passing between objects

---

## Test Categories

### Unit Tests (Mocked Dependencies)

**Purpose:** Test service layer logic in isolation

**Characteristics:**
- Fast execution (< 1ms per test)
- No external dependencies
- Mocked repositories and clients
- 90% coverage target

**File Pattern:** `*_test.go`

**Example:**
```go
func TestRegister_ValidInput_CreatesUser(t *testing.T) {
    mockRepo := mocks.NewUserRepository(t)
    mockRepo.On("GetByEmail", "test@example.com").Return(nil, nil)
    mockRepo.On("Create", mock.Anything).Return(nil)

    svc := identity.NewService(mockRepo)
    user, err := svc.Register("test@example.com", "password123", "testuser", "INVITE123")

    assert.NoError(t, err)
    assert.Equal(t, "testuser", user.Handle)
}
```

---

### Integration Tests (Real Databases)

**Purpose:** Verify database interactions work correctly

**Characteristics:**
- Uses test containers (Docker)
- Real PostgreSQL, Neo4j, Redis
- Slower execution (100ms - 1s per test)
- Tests repository implementations

**File Pattern:** `*_integration_test.go`

**Example:**
```go
func TestUserRepository_Create_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    pool := testutil.SetupPostgres(t)
    repo := db.NewUserRepository(pool)

    user, err := repo.Create(&User{Email: "test@example.com", Handle: "testuser"})

    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

---

### E2E Tests (Full Stack)

**Purpose:** Verify complete user flows

**Characteristics:**
- Uses real external services (AI APIs, blockchain testnet)
- Slowest execution (2-30s per test)
- Limited test count (expensive)
- Run in CI, not locally

**File Pattern:** `*_e2e_test.go`

**Example:**
```go
func TestUserRegistrationFlow_E2E(t *testing.T) {
    if os.Getenv("RUN_E2E") != "true" {
        t.Skip("E2E tests disabled")
    }

    client := testutil.NewAPIClient(t)

    // Create invite
    invite := client.CreateInvite(t, communityID)

    // Register user
    user := client.Register(t, RegisterRequest{
        Email: "new@example.com",
        Password: "password123",
        Handle: "newuser",
        InviteCode: invite.Code,
    })

    assert.NotEmpty(t, user.ID)
    assert.Equal(t, "newuser", user.Handle)
}
```

---

## Performance Tests

**Purpose:** Verify latency requirements are met

**Targets:**
| Metric | Target | Test File |
|--------|--------|-----------|
| Message delivery | < 200ms P95 | `chat/delivery_perf_test.go` |
| Echo response | < 2s P95 | `echo/response_perf_test.go` |
| Search results | < 500ms P95 | `search/search_perf_test.go` |
| Summarization | < 30s | `knowledge/summarize_perf_test.go` |

**Example:**
```go
func BenchmarkMessageDelivery(b *testing.B) {
    hub := websocket.NewHub()
    go hub.Run()

    // Setup clients
    clients := make([]*websocket.Client, 100)
    for i := range clients {
        clients[i] = testutil.NewTestClient(hub)
    }

    msg := &Message{Content: "test message"}

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        hub.Broadcast("thread-1", msg)
    }

    // Assert P95 < 200ms
    assert.Less(b, b.Elapsed().Milliseconds()/int64(b.N), int64(200))
}
```

---

## Test Data Management

### Fixtures

**Location:** `testdata/fixtures/`

**Contents:**
- `users.json` - Sample user data
- `communities.json` - Sample community data
- `messages.json` - Sample message data
- `embeddings.json` - Pre-computed embeddings for semantic search tests

### Test Containers

**Setup:** `testutil/containers.go`

```go
func SetupPostgres(t *testing.T) *pgxpool.Pool {
    ctx := context.Background()
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:15",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_PASSWORD": "test",
                "POSTGRES_DB":       "commcomms_test",
            },
            WaitingFor: wait.ForListeningPort("5432/tcp"),
        },
        Started: true,
    })
    require.NoError(t, err)
    t.Cleanup(func() { container.Terminate(ctx) })

    // Return pool connection
}
```

---

## Mock Generation

**Tool:** [mockery](https://github.com/vektra/mockery)

**Configuration:** `.mockery.yaml`

```yaml
all: true
output: mocks
dir: internal
packages:
  github.com/yourorg/commcomms/internal/identity:
    interfaces:
      UserRepository:
      InviteRepository:
  github.com/yourorg/commcomms/internal/chat:
    interfaces:
      MessageRepository:
      ThreadRepository:
  github.com/yourorg/commcomms/internal/ai:
    interfaces:
      Client:
```

**Generate:** `go generate ./...`

---

## Test Execution

### Local Development

```bash
# Run unit tests only (fast)
go test -short ./...

# Run all tests including integration
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./internal/identity/...

# Run specific test
go test -run TestRegister ./internal/identity/...
```

### CI Pipeline

```yaml
# .github/workflows/ci.yml
test:
  runs-on: ubuntu-latest
  services:
    postgres:
      image: postgres:15
      env:
        POSTGRES_PASSWORD: test
        POSTGRES_DB: commcomms_test
      ports:
        - 5432:5432
    redis:
      image: redis:7
      ports:
        - 6379:6379
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.22'
    - name: Run tests
      run: go test -race -coverprofile=coverage.out ./...
    - name: Upload coverage
      uses: codecov/codecov-action@v4
```

---

## Acceptance Tests (Written)

### test/acceptance/identity_test.go

| Test | User Story | Status |
|------|-----------|--------|
| TestUserRegistration_Acceptance | US-ID-001 | SKIP (awaiting implementation) |
| TestHandleValidation_Acceptance | US-ID-002 | SKIP (awaiting implementation) |
| TestUserLogin_Acceptance | - | SKIP (awaiting implementation) |
| TestTokenRefresh_Acceptance | - | SKIP (awaiting implementation) |
| TestProtectedRoutes_Acceptance | - | SKIP (awaiting implementation) |
| TestInviteManagement_Acceptance | US-ID-004 | SKIP (awaiting implementation) |
| TestReputation_Acceptance | US-ID-003 | SKIP (awaiting implementation) |

### test/acceptance/chat_test.go

| Test | User Story | Status |
|------|-----------|--------|
| TestSendMessage_Acceptance | US-CHAT-001 | SKIP (awaiting implementation) |
| TestCreateThread_Acceptance | US-CHAT-002 | SKIP (awaiting implementation) |
| TestPresence_Acceptance | US-CHAT-003 | SKIP (awaiting implementation) |
| TestAsyncMode_Acceptance | US-CHAT-004 | SKIP (awaiting implementation) |
| TestRealTimeDelivery_Acceptance | - | SKIP (awaiting implementation) |
| TestMessageEditing_Acceptance | - | SKIP (awaiting implementation) |
| TestMessageDeletion_Acceptance | - | SKIP (awaiting implementation) |
| TestCommunityManagement_Acceptance | - | SKIP (awaiting implementation) |

---

## Unit Test Results (Pending)

### Phase 1: Identity Service

| Test Suite | Tests | Passed | Failed | Coverage |
|------------|-------|--------|--------|----------|
| identity/service_test.go | - | - | - | - |
| identity/invite_test.go | - | - | - | - |
| identity/reputation_test.go | - | - | - | - |
| auth/jwt_test.go | - | - | - | - |
| auth/middleware_test.go | - | - | - | - |

*Results will be updated as unit tests are written and executed during `/oss:build`.*

---

## Last Updated: 2025-12-29 by /oss:acceptance
