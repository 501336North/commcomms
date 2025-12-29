// Package acceptance contains acceptance tests for the CommComms Chat API.
// These tests operate at the HTTP/WebSocket boundary (black-box testing).
package acceptance

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// US-CHAT-001: Send Message
// ============================================

// TestSendMessage_Acceptance tests message sending functionality.
//
// User Story: As a member, I want to send a message to a thread
// so that I can communicate with my community.
func TestSendMessage_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("AC-CHAT-001.1: should send message and receive confirmation", func(t *testing.T) {
		// GIVEN - I am authenticated and in a thread
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		thread := createTestThread(t, token)

		// WHEN - I send a message
		reqBody := map[string]string{
			"content": "Hello, world!",
		}
		resp := postJSONAuth(t, "/api/v1/threads/"+thread.ID+"/messages", reqBody, token)

		// THEN - I should receive confirmation with message details
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.NotEmpty(t, body["id"])
		assert.Equal(t, "Hello, world!", body["content"])
		assert.Equal(t, false, body["isEcho"])
	})

	t.Run("AC-CHAT-001.2: should reject empty message", func(t *testing.T) {
		// GIVEN - I try to send an empty message
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		thread := createTestThread(t, token)

		// WHEN - I send empty content
		reqBody := map[string]string{
			"content": "",
		}
		resp := postJSONAuth(t, "/api/v1/threads/"+thread.ID+"/messages", reqBody, token)

		// THEN - I should see an error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "cannot be empty")
	})

	t.Run("AC-CHAT-001.3: should reject message over 10000 characters", func(t *testing.T) {
		// GIVEN - A message that's too long
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		thread := createTestThread(t, token)

		longContent := make([]byte, 10001)
		for i := range longContent {
			longContent[i] = 'a'
		}

		// WHEN - I try to send it
		reqBody := map[string]string{
			"content": string(longContent),
		}
		resp := postJSONAuth(t, "/api/v1/threads/"+thread.ID+"/messages", reqBody, token)

		// THEN - I should see a length error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "10,000 characters")
	})

	t.Run("AC-CHAT-001.4: should rate limit at 30 messages per minute", func(t *testing.T) {
		// GIVEN - I've sent 30 messages in the last minute
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		thread := createTestThread(t, token)

		// Send 30 messages
		for i := 0; i < 30; i++ {
			reqBody := map[string]string{
				"content": "Message " + string(rune('0'+i)),
			}
			resp := postJSONAuth(t, "/api/v1/threads/"+thread.ID+"/messages", reqBody, token)
			require.Equal(t, http.StatusCreated, resp.StatusCode)
		}

		// WHEN - I try to send another
		reqBody := map[string]string{
			"content": "Message 31",
		}
		resp := postJSONAuth(t, "/api/v1/threads/"+thread.ID+"/messages", reqBody, token)

		// THEN - I should be rate limited
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "Slow down")
	})
}

// ============================================
// US-CHAT-002: Create Thread
// ============================================

// TestCreateThread_Acceptance tests thread creation.
//
// User Story: As a member, I want to start a new thread
// so that I can initiate a focused discussion.
func TestCreateThread_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("AC-CHAT-002.1: should create thread with title", func(t *testing.T) {
		// GIVEN - I am authenticated and in a channel
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		channel := createTestChannel(t, token)

		// WHEN - I create a thread
		reqBody := map[string]string{
			"title":          "Best coworking in Lisbon?",
			"initialMessage": "Looking for recommendations!",
		}
		resp := postJSONAuth(t, "/api/v1/channels/"+channel.ID+"/threads", reqBody, token)

		// THEN - Thread should be created
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.NotEmpty(t, body["id"])
		assert.Equal(t, "Best coworking in Lisbon?", body["title"])
		assert.Equal(t, float64(1), body["messageCount"]) // Initial message
	})

	t.Run("AC-CHAT-002.2: should reject thread without title", func(t *testing.T) {
		// GIVEN - I try to create a thread without title
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		channel := createTestChannel(t, token)

		// WHEN - I submit without title
		reqBody := map[string]string{
			"title": "",
		}
		resp := postJSONAuth(t, "/api/v1/channels/"+channel.ID+"/threads", reqBody, token)

		// THEN - I should see an error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "title required")
	})
}

// ============================================
// US-CHAT-003: Real-Time Presence
// ============================================

// TestPresence_Acceptance tests presence tracking via WebSocket.
//
// User Story: As a member, I want to see who is currently online
// so that I know if I'll get an immediate response.
func TestPresence_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("AC-CHAT-003.1: should show user as online when connected", func(t *testing.T) {
		// GIVEN - A user connects via WebSocket
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		community := createTestCommunity(t, token)

		// Connect WebSocket
		ws := connectWebSocket(t, token, community.ID)
		defer ws.Close()

		// WHEN - Another user checks presence
		otherUser := createTestUser(t)
		otherToken := loginUser(t, otherUser.Email, "TestPass123!").AccessToken
		resp := getJSON(t, "/api/v1/communities/"+community.ID+"/presence", otherToken)

		// THEN - First user should appear online
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		online := body["online"].([]interface{})
		assert.GreaterOrEqual(t, len(online), 1)

		// Find our user
		var found bool
		for _, u := range online {
			userMap := u.(map[string]interface{})
			userInfo := userMap["user"].(map[string]interface{})
			if userInfo["handle"] == user.Handle {
				found = true
				break
			}
		}
		assert.True(t, found, "User should be in online list")
	})

	t.Run("AC-CHAT-003.2: should mark user offline after disconnect", func(t *testing.T) {
		// GIVEN - A connected user
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		community := createTestCommunity(t, token)

		ws := connectWebSocket(t, token, community.ID)

		// WHEN - User disconnects
		ws.Close()

		// Wait for offline detection (30s in production, shorter in test)
		time.Sleep(35 * time.Second)

		// THEN - User should be offline
		otherUser := createTestUser(t)
		otherToken := loginUser(t, otherUser.Email, "TestPass123!").AccessToken
		resp := getJSON(t, "/api/v1/communities/"+community.ID+"/presence", otherToken)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		online := body["online"].([]interface{})

		// User should NOT be in online list
		var found bool
		for _, u := range online {
			userMap := u.(map[string]interface{})
			userInfo := userMap["user"].(map[string]interface{})
			if userInfo["handle"] == user.Handle {
				found = true
				break
			}
		}
		assert.False(t, found, "User should NOT be in online list after disconnect")
	})

	t.Run("AC-CHAT-003.3: should show typing indicator", func(t *testing.T) {
		// GIVEN - Two users in same thread
		user1 := createTestUser(t)
		token1 := loginUser(t, user1.Email, "TestPass123!").AccessToken
		thread := createTestThread(t, token1)

		user2 := createTestUser(t)
		token2 := loginUser(t, user2.Email, "TestPass123!").AccessToken

		// Connect both to WebSocket
		ws1 := connectWebSocket(t, token1, thread.CommunityID)
		ws2 := connectWebSocket(t, token2, thread.CommunityID)
		defer ws1.Close()
		defer ws2.Close()

		// Subscribe to thread
		subscribeToThread(t, ws1, thread.ID)
		subscribeToThread(t, ws2, thread.ID)

		// WHEN - User 1 starts typing
		sendTypingEvent(t, ws1, thread.ID)

		// THEN - User 2 should see typing indicator
		msg := readWebSocketMessage(t, ws2, 5*time.Second)
		assert.Equal(t, "presence:typing", msg["type"])
		payload := msg["payload"].(map[string]interface{})
		assert.Equal(t, thread.ID, payload["threadId"])
		assert.Equal(t, user1.Handle, payload["handle"])
	})
}

// ============================================
// US-CHAT-004: Async Mode Detection
// ============================================

// TestAsyncMode_Acceptance tests async/real-time detection.
//
// User Story: As a member, I want to know if my message will be seen
// immediately or later so that I set correct expectations.
func TestAsyncMode_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("AC-CHAT-004.1: should indicate async when recipient offline", func(t *testing.T) {
		// GIVEN - Recipient is offline
		sender := createTestUser(t)
		senderToken := loginUser(t, sender.Email, "TestPass123!").AccessToken
		thread := createTestThread(t, senderToken)

		// Recipient exists but is not connected
		createTestUserInThread(t, thread.ID)

		// WHEN - I send a message
		reqBody := map[string]string{
			"content": "Hello, are you there?",
		}
		resp := postJSONAuth(t, "/api/v1/threads/"+thread.ID+"/messages", reqBody, senderToken)

		// THEN - Response should indicate async delivery
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		// deliveryMode should indicate this will be delivered later
		assert.Contains(t, body, "deliveryMode")
		assert.Equal(t, "async", body["deliveryMode"])
	})

	t.Run("AC-CHAT-004.2: should indicate real-time when recipient online", func(t *testing.T) {
		// GIVEN - Recipient is online
		sender := createTestUser(t)
		senderToken := loginUser(t, sender.Email, "TestPass123!").AccessToken
		thread := createTestThread(t, senderToken)

		recipient := createTestUserInThread(t, thread.ID)
		recipientToken := loginUser(t, recipient.Email, "TestPass123!").AccessToken

		// Connect recipient
		ws := connectWebSocket(t, recipientToken, thread.CommunityID)
		defer ws.Close()
		subscribeToThread(t, ws, thread.ID)

		// WHEN - I send a message
		reqBody := map[string]string{
			"content": "Hello, are you there?",
		}
		resp := postJSONAuth(t, "/api/v1/threads/"+thread.ID+"/messages", reqBody, senderToken)

		// THEN - Response should indicate real-time delivery
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Equal(t, "realtime", body["deliveryMode"])
	})
}

// ============================================
// Real-Time Message Delivery
// ============================================

// TestRealTimeDelivery_Acceptance tests WebSocket message delivery.
func TestRealTimeDelivery_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("should deliver message to all online participants within 200ms", func(t *testing.T) {
		// GIVEN - Multiple users connected to same thread
		user1 := createTestUser(t)
		token1 := loginUser(t, user1.Email, "TestPass123!").AccessToken
		thread := createTestThread(t, token1)

		user2 := createTestUser(t)
		token2 := loginUser(t, user2.Email, "TestPass123!").AccessToken

		// Connect both via WebSocket
		ws1 := connectWebSocket(t, token1, thread.CommunityID)
		ws2 := connectWebSocket(t, token2, thread.CommunityID)
		defer ws1.Close()
		defer ws2.Close()

		subscribeToThread(t, ws1, thread.ID)
		subscribeToThread(t, ws2, thread.ID)

		// WHEN - User 1 sends a message
		start := time.Now()
		reqBody := map[string]string{
			"content": "Real-time test message",
		}
		resp := postJSONAuth(t, "/api/v1/threads/"+thread.ID+"/messages", reqBody, token1)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// THEN - User 2 should receive it within 200ms
		msg := readWebSocketMessage(t, ws2, 200*time.Millisecond)
		elapsed := time.Since(start)

		assert.Equal(t, "message:new", msg["type"])
		assert.Less(t, elapsed, 200*time.Millisecond, "Message delivery should be < 200ms")

		payload := msg["payload"].(map[string]interface{})
		message := payload["message"].(map[string]interface{})
		assert.Equal(t, "Real-time test message", message["content"])
	})
}

// ============================================
// Message Editing and Deletion
// ============================================

// TestMessageEditing_Acceptance tests message modification.
func TestMessageEditing_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("should edit own message", func(t *testing.T) {
		// GIVEN - A message I sent
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		thread := createTestThread(t, token)

		msgResp := postJSONAuth(t, "/api/v1/threads/"+thread.ID+"/messages",
			map[string]string{"content": "Original message"}, token)
		require.Equal(t, http.StatusCreated, msgResp.StatusCode)

		var msgBody map[string]interface{}
		json.NewDecoder(msgResp.Body).Decode(&msgBody)
		messageID := msgBody["id"].(string)

		// WHEN - I edit it
		editResp := patchJSONAuth(t, "/api/v1/messages/"+messageID,
			map[string]string{"content": "Edited message"}, token)

		// THEN - Edit should succeed
		assert.Equal(t, http.StatusOK, editResp.StatusCode)

		var editBody map[string]interface{}
		json.NewDecoder(editResp.Body).Decode(&editBody)
		assert.Equal(t, "Edited message", editBody["content"])
		assert.NotEmpty(t, editBody["editedAt"])
	})

	t.Run("should not edit others message", func(t *testing.T) {
		// GIVEN - A message from another user
		user1 := createTestUser(t)
		token1 := loginUser(t, user1.Email, "TestPass123!").AccessToken
		thread := createTestThread(t, token1)

		msgResp := postJSONAuth(t, "/api/v1/threads/"+thread.ID+"/messages",
			map[string]string{"content": "Original message"}, token1)
		require.Equal(t, http.StatusCreated, msgResp.StatusCode)

		var msgBody map[string]interface{}
		json.NewDecoder(msgResp.Body).Decode(&msgBody)
		messageID := msgBody["id"].(string)

		// Another user
		user2 := createTestUser(t)
		token2 := loginUser(t, user2.Email, "TestPass123!").AccessToken

		// WHEN - I try to edit it
		editResp := patchJSONAuth(t, "/api/v1/messages/"+messageID,
			map[string]string{"content": "Hacked!"}, token2)

		// THEN - Should be forbidden
		assert.Equal(t, http.StatusForbidden, editResp.StatusCode)
	})
}

// TestMessageDeletion_Acceptance tests message deletion.
func TestMessageDeletion_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("should delete own message", func(t *testing.T) {
		// GIVEN - A message I sent
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		thread := createTestThread(t, token)

		msgResp := postJSONAuth(t, "/api/v1/threads/"+thread.ID+"/messages",
			map[string]string{"content": "Delete me"}, token)
		require.Equal(t, http.StatusCreated, msgResp.StatusCode)

		var msgBody map[string]interface{}
		json.NewDecoder(msgResp.Body).Decode(&msgBody)
		messageID := msgBody["id"].(string)

		// WHEN - I delete it
		deleteResp := deleteJSONAuth(t, "/api/v1/messages/"+messageID, token)

		// THEN - Delete should succeed
		assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

		// Verify it's gone
		getResp := getJSON(t, "/api/v1/threads/"+thread.ID, token)
		var threadBody map[string]interface{}
		json.NewDecoder(getResp.Body).Decode(&threadBody)
		messages := threadBody["messages"].([]interface{})
		assert.Empty(t, messages)
	})
}

// ============================================
// Community & Channel Management
// ============================================

// TestCommunityManagement_Acceptance tests community operations.
func TestCommunityManagement_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("should create community", func(t *testing.T) {
		// GIVEN - An authenticated user
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken

		// WHEN - I create a community
		reqBody := map[string]interface{}{
			"name":        "Digital Nomads",
			"description": "A community for remote workers",
			"isPrivate":   true,
		}
		resp := postJSONAuth(t, "/api/v1/communities", reqBody, token)

		// THEN - Community should be created with defaults
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.NotEmpty(t, body["id"])
		assert.Equal(t, "Digital Nomads", body["name"])

		settings := body["settings"].(map[string]interface{})
		assert.Equal(t, true, settings["echoEnabled"])
		assert.Equal(t, float64(24), settings["echoTtlHours"])
	})

	t.Run("should create channel in community", func(t *testing.T) {
		// GIVEN - A community I own
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		community := createTestCommunity(t, token)

		// WHEN - I create a channel
		reqBody := map[string]string{
			"name":        "general",
			"description": "General discussion",
		}
		resp := postJSONAuth(t, "/api/v1/communities/"+community.ID+"/channels", reqBody, token)

		// THEN - Channel should be created
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.NotEmpty(t, body["id"])
		assert.Equal(t, "general", body["name"])
	})

	t.Run("should reject duplicate channel name", func(t *testing.T) {
		// GIVEN - A channel already exists
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken
		community := createTestCommunity(t, token)

		// Create first channel
		reqBody := map[string]string{"name": "general"}
		resp1 := postJSONAuth(t, "/api/v1/communities/"+community.ID+"/channels", reqBody, token)
		require.Equal(t, http.StatusCreated, resp1.StatusCode)

		// WHEN - I try to create another with same name
		resp2 := postJSONAuth(t, "/api/v1/communities/"+community.ID+"/channels", reqBody, token)

		// THEN - Should be rejected
		assert.Equal(t, http.StatusConflict, resp2.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&body)
		assert.Contains(t, body["error"], "already exists")
	})
}

// ============================================
// Helper Types & Functions
// ============================================

// TestThread represents a test thread.
type TestThread struct {
	ID          string
	CommunityID string
	ChannelID   string
}

// TestChannel represents a test channel.
type TestChannel struct {
	ID          string
	CommunityID string
}

// TestCommunity represents a test community.
type TestCommunity struct {
	ID string
}

// createTestThread creates a test thread.
func createTestThread(t *testing.T, token string) TestThread {
	t.Helper()
	// TODO: Implement when server exists
	return TestThread{
		ID:          "test-thread-id",
		CommunityID: "test-community-id",
		ChannelID:   "test-channel-id",
	}
}

// createTestChannel creates a test channel.
func createTestChannel(t *testing.T, token string) TestChannel {
	t.Helper()
	// TODO: Implement when server exists
	return TestChannel{
		ID:          "test-channel-id",
		CommunityID: "test-community-id",
	}
}

// createTestCommunity creates a test community.
func createTestCommunity(t *testing.T, token string) TestCommunity {
	t.Helper()
	// TODO: Implement when server exists
	return TestCommunity{
		ID: "test-community-id",
	}
}

// createTestUserInThread adds a user to a thread.
func createTestUserInThread(t *testing.T, threadID string) TestUser {
	t.Helper()
	// TODO: Implement when server exists
	return TestUser{
		ID:     "thread-user-id",
		Email:  "threaduser@example.com",
		Handle: "threaduser",
	}
}

// connectWebSocket connects to the WebSocket endpoint.
func connectWebSocket(t *testing.T, token, communityID string) *websocket.Conn {
	t.Helper()

	url := "ws" + TestServer.URL[4:] + "/api/v1/ws?token=" + token + "&community=" + communityID
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)

	return conn
}

// subscribeToThread subscribes WebSocket to a thread.
func subscribeToThread(t *testing.T, ws *websocket.Conn, threadID string) {
	t.Helper()

	msg := map[string]string{
		"action":   "subscribe",
		"threadId": threadID,
	}
	err := ws.WriteJSON(msg)
	require.NoError(t, err)
}

// sendTypingEvent sends a typing indicator.
func sendTypingEvent(t *testing.T, ws *websocket.Conn, threadID string) {
	t.Helper()

	msg := map[string]string{
		"action":   "typing",
		"threadId": threadID,
	}
	err := ws.WriteJSON(msg)
	require.NoError(t, err)
}

// readWebSocketMessage reads a message with timeout.
func readWebSocketMessage(t *testing.T, ws *websocket.Conn, timeout time.Duration) map[string]interface{} {
	t.Helper()

	ws.SetReadDeadline(time.Now().Add(timeout))

	var msg map[string]interface{}
	err := ws.ReadJSON(&msg)
	require.NoError(t, err)

	return msg
}

// patchJSONAuth sends a PATCH request with auth.
func patchJSONAuth(t *testing.T, path string, body interface{}, token string) *http.Response {
	t.Helper()

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPatch, TestServer.URL+path, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

// deleteJSONAuth sends a DELETE request with auth.
func deleteJSONAuth(t *testing.T, path string, token string) *http.Response {
	t.Helper()

	req, _ := http.NewRequest(http.MethodDelete, TestServer.URL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}
