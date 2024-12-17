package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"testing"
	"time"
)

type mockWriter struct {
	sync.Mutex
	written []byte
}

type mockServer struct{}

func TestNewServer(t *testing.T) {
	tests := []struct {
		name     string
		options  []ServerOption
		expected ServerCapabilities
	}{
		{
			name:     "empty server",
			options:  nil,
			expected: ServerCapabilities{},
		},
		{
			name: "with prompt server",
			options: []ServerOption{
				WithPromptServer(&mockPromptServer{}),
			},
			expected: ServerCapabilities{
				Prompts: &PromptsCapability{},
			},
		},
		{
			name: "with prompt server and watcher",
			options: []ServerOption{
				WithPromptServer(&mockPromptServer{}),
				WithPromptListUpdater(&mockPromptListWatcher{}),
			},
			expected: ServerCapabilities{
				Prompts: &PromptsCapability{
					ListChanged: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &mockServer{}
			s := newServer(srv, tt.options...)
			if !reflect.DeepEqual(tt.expected, s.capabilities) {
				t.Errorf("got capabilities %+v, want %+v", s.capabilities, tt.expected)
			}
		})
	}
}

func TestServerStart(t *testing.T) {
	srv := &mockServer{}
	s := newServer(srv)

	s.start()

	if s.sessionStopChan == nil {
		t.Error("sessionStopChan is nil")
	}
	if s.closeChan == nil {
		t.Error("closeChan is nil")
	}

	// Clean up
	s.stop()
}

func TestServerHandleMsg(t *testing.T) {
	writer := &mockWriter{}
	sess := &serverSession{
		id:           "test-session",
		ctx:          context.Background(),
		writter:      writer,
		writeTimeout: time.Second,
	}

	srv := &mockServer{}
	s := newServer(srv, WithWriteTimeout(time.Second))

	s.sessions.Store(sess.id, sess)

	// Test ping message
	pingMsg := `{"jsonrpc": "2.0", "method": "ping", "id": "1"}`
	err := s.handleMsg(bytes.NewReader([]byte(pingMsg)), sess.id)
	if err != nil {
		t.Fatalf("handleMsg failed: %v", err)
	}

	var response jsonRPCMessage
	err = json.NewDecoder(bytes.NewReader(writer.getWritten())).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.ID != MustString("1") {
		t.Errorf("got ID %v, want 1", response.ID)
	}
}

func TestServerStartSession(t *testing.T) {
	srv := &mockServer{}
	s := newServer(srv, WithWriteTimeout(time.Second))
	s.start()

	writer := &mockWriter{}
	ctx := context.Background()

	s.startSession(ctx, writer)

	var sessionCount int
	s.sessions.Range(func(_, _ interface{}) bool {
		sessionCount++
		return true
	})

	if sessionCount != 1 {
		t.Errorf("got %d sessions, want 1", sessionCount)
	}

	// Clean up
	s.stop()
}

func TestServerRootsList(t *testing.T) {
	writer := &mockWriter{}
	srv := &mockServer{}
	s := newServer(srv, WithWriteTimeout(time.Second), WithReadTimeout(time.Second))

	ctx := context.Background()
	sessID := s.startSession(ctx, writer)

	// Start goroutine to handle mock response
	go func() {
		// Loop until we get a request
		var msg jsonRPCMessage
		for {
			wbs := writer.getWritten()
			if len(wbs) == 0 {
				continue
			}
			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
			if err != nil {
				t.Errorf("failed to decode request: %v", err)
				return
			}
			break
		}

		// Verify request
		if msg.Method != methodRootsList {
			t.Errorf("expected method %s, got %s", methodRootsList, msg.Method)
		}

		// Send mock response
		mockResponse := RootList{
			Roots: []Root{
				{URI: "test://root1", Name: "Root 1"},
				{URI: "test://root2", Name: "Root 2"},
			},
		}

		responseMsg := jsonRPCMessage{
			JSONRPC: jsonRPCVersion,
			ID:      msg.ID,
			Result:  json.RawMessage(mustMarshal(mockResponse)),
		}

		responseBs, _ := json.Marshal(responseMsg)
		err := s.handleMsg(bytes.NewReader(responseBs), sessID)
		if err != nil {
			t.Errorf("handleMsg failed: %v", err)
		}
	}()

	// Call rootsList
	ctx = ctxWithSessionID(ctx, sessID)
	result, err := s.rootsList(ctx)
	if err != nil {
		t.Fatalf("rootsList failed: %v", err)
	}

	// Verify result
	if len(result.Roots) != 2 {
		t.Errorf("expected 2 roots, got %d", len(result.Roots))
	}
	if result.Roots[0].URI != "test://root1" || result.Roots[0].Name != "Root 1" {
		t.Errorf("unexpected root[0]: %+v", result.Roots[0])
	}
	if result.Roots[1].URI != "test://root2" || result.Roots[1].Name != "Root 2" {
		t.Errorf("unexpected root[1]: %+v", result.Roots[1])
	}
}

func TestServerCreateSampleMessage(t *testing.T) {
	writer := &mockWriter{}
	srv := &mockServer{}
	s := newServer(srv, WithWriteTimeout(time.Second), WithReadTimeout(time.Second))

	ctx := context.Background()
	sessID := s.startSession(ctx, writer)

	// Set up mock response
	ss, _ := s.sessions.Load(sessID)
	sess, _ := ss.(*serverSession)

	// Start goroutine to handle mock response
	go func() {
		// Loop until we get a request
		var msg jsonRPCMessage
		for {
			wbs := writer.getWritten()
			if len(wbs) == 0 {
				continue
			}
			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
			if err != nil {
				t.Errorf("failed to decode request: %v", err)
				return
			}
			break
		}

		// Verify request
		if msg.Method != methodSamplingCreateMessage {
			t.Errorf("expected method %s, got %s", methodSamplingCreateMessage, msg.Method)
		}

		// Verify params
		var receivedParams SamplingParams
		if err := json.Unmarshal(msg.Params, &receivedParams); err != nil {
			t.Errorf("failed to decode params: %v", err)
			return
		}

		// Send mock response
		mockResponse := SamplingResult{
			Role: PromptRoleAssistant,
			Content: SamplingContent{
				Type: "text",
				Text: "Hello! How can I help you?",
			},
			Model:      "test-model",
			StopReason: "completed",
		}

		responseMsg := jsonRPCMessage{
			JSONRPC: jsonRPCVersion,
			ID:      msg.ID,
			Result:  json.RawMessage(mustMarshal(mockResponse)),
		}

		rc, _ := sess.serverRequests.Load(string(msg.ID))
		resChan, _ := rc.(chan jsonRPCMessage)
		resChan <- responseMsg
	}()

	// Call createSampleMessage
	ctx = ctxWithSessionID(ctx, sessID)
	// Test params
	result, err := s.createSampleMessage(ctx, SamplingParams{
		Messages: []SamplingMessage{
			{
				Role: PromptRoleUser,
				Content: SamplingContent{
					Type: "text",
					Text: "Hello",
				},
			},
		},
		ModelPreferences: SamplingModelPreferences{
			CostPriority:         1,
			SpeedPriority:        2,
			IntelligencePriority: 3,
		},
		SystemPrompts: "Be helpful",
		MaxTokens:     100,
	})
	if err != nil {
		t.Fatalf("createSampleMessage failed: %v", err)
	}

	// Verify result
	if result.Role != PromptRoleAssistant {
		t.Errorf("expected role %s, got %s", PromptRoleAssistant, result.Role)
	}
	if result.Content.Text != "Hello! How can I help you?" {
		t.Errorf("unexpected content text: %s", result.Content.Text)
	}
	if result.Model != "test-model" {
		t.Errorf("expected model test-model, got %s", result.Model)
	}
	if result.StopReason != "completed" {
		t.Errorf("expected stop reason completed, got %s", result.StopReason)
	}
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func (w *mockWriter) Write(p []byte) (int, error) {
	w.Lock()
	defer w.Unlock()
	w.written = append(w.written, p...)
	return len(p), nil
}

func (w *mockWriter) getWritten() []byte {
	w.Lock()
	defer w.Unlock()
	return w.written
}

func (m *mockServer) Info() Info {
	return Info{Name: "test-server", Version: "1.0"}
}

func (m *mockServer) RequiredClientCapabilities() ClientCapabilities {
	return ClientCapabilities{}
}
