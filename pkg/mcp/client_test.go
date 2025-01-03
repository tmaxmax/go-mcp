package mcp_test

import (
	"context"

	"github.com/MegaGrindStone/go-mcp/pkg/mcp"
)

type mockPromptListWatcher struct{}

type mockResourceListWatcher struct{}

type mockResourceSubscribedWatcher struct{}

type mockToolListWatcher struct{}

type mockRootsListHandler struct{}

type mockRootsListUpdater struct {
	ch chan struct{}
}

type mockSamplingHandler struct{}

type mockLogReceiver struct{}

func (m mockPromptListWatcher) OnPromptListChanged() {
}

func (m mockResourceListWatcher) OnResourceListChanged() {
}

func (m mockResourceSubscribedWatcher) OnResourceSubscribedChanged(string) {
}

func (m mockToolListWatcher) OnToolListChanged() {
}

func (m mockRootsListHandler) RootsList(context.Context) (mcp.RootList, error) {
	return mcp.RootList{
		Roots: []mcp.Root{
			{URI: "test://root", Name: "Test Root"},
		},
	}, nil
}

func (m mockRootsListUpdater) RootsListUpdates() <-chan struct{} {
	if m.ch == nil {
		m.ch = make(chan struct{})
	}
	return m.ch
}

func (m mockSamplingHandler) CreateSampleMessage(context.Context, mcp.SamplingParams) (mcp.SamplingResult, error) {
	return mcp.SamplingResult{
		Role: mcp.PromptRoleAssistant,
		Content: mcp.SamplingContent{
			Type: "text",
			Text: "Test response",
		},
		Model:      "test-model",
		StopReason: "completed",
	}, nil
}

func (m mockLogReceiver) OnLog(_ mcp.LogParams) {
}

// func TestNewClient(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		options  []ClientOption
// 		expected ClientCapabilities
// 	}{
// 		{
// 			name:     "empty client",
// 			options:  nil,
// 			expected: ClientCapabilities{},
// 		},
// 		{
// 			name: "with roots handler",
// 			options: []ClientOption{
// 				WithRootsListHandler(&mockRootsListHandler{}),
// 			},
// 			expected: ClientCapabilities{
// 				Roots: &RootsCapability{},
// 			},
// 		},
// 		{
// 			name: "with roots handler and updater",
// 			options: []ClientOption{
// 				WithRootsListHandler(&mockRootsListHandler{}),
// 				WithRootsListUpdater(&mockRootsListUpdater{}),
// 			},
// 			expected: ClientCapabilities{
// 				Roots: &RootsCapability{
// 					ListChanged: true,
// 				},
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cli := &mockClient{}
// 			c := newClient(cli, tt.options...)
// 			if !reflect.DeepEqual(tt.expected, c.capabilities) {
// 				t.Errorf("got capabilities %+v, want %+v", c.capabilities, tt.expected)
// 			}
// 		})
// 	}
// }
//
// func TestClientStart(t *testing.T) {
// 	cli := &mockClient{}
// 	c := newClient(cli)
//
// 	c.start()
//
// 	if c.sessionStopChan == nil {
// 		t.Error("sessionStopChan is nil")
// 	}
// 	if c.closeChan == nil {
// 		t.Error("closeChan is nil")
// 	}
//
// 	// Clean up
// 	c.stop()
// }
//
// func TestClientHandleMsg(t *testing.T) {
// 	writer := &mockWriter{}
// 	sess := &clientSession{
// 		id:           "test-session",
// 		ctx:          context.Background(),
// 		writter:      writer,
// 		writeTimeout: time.Second,
// 	}
//
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second))
//
// 	c.sessions.Store(sess.id, sess)
//
// 	// Test ping message
// 	pingMsg := `{"jsonrpc": "2.0", "method": "ping", "id": "1"}`
// 	err := c.handleMsg(bytes.NewReader([]byte(pingMsg)), sess.id)
// 	if err != nil {
// 		t.Fatalf("handleMsg failed: %v", err)
// 	}
//
// 	var response JSONRPCMessage
// 	err = json.NewDecoder(bytes.NewReader(writer.getWritten())).Decode(&response)
// 	if err != nil {
// 		t.Fatalf("failed to decode response: %v", err)
// 	}
// 	if response.ID != MustString("1") {
// 		t.Errorf("got ID %v, want 1", response.ID)
// 	}
// }
//
// func TestClientStartSession(t *testing.T) {
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second))
// 	c.start()
//
// 	writer := &mockWriter{}
// 	ctx := context.Background()
//
// 	c.startSession(ctx, writer, "test-session")
//
// 	var sessionCount int
// 	c.sessions.Range(func(_, _ interface{}) bool {
// 		sessionCount++
// 		return true
// 	})
//
// 	if sessionCount != 1 {
// 		t.Errorf("got %d sessions, want 1", sessionCount)
// 	}
//
// 	// Clean up
// 	c.stop()
// }
//
// func TestClientGetPrompt(t *testing.T) {
// 	writer := &mockWriter{}
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second), WithClientReadTimeout(time.Second))
//
// 	ctx := context.Background()
// 	sessID := "test-session"
// 	c.startSession(ctx, writer, sessID)
//
// 	// Start goroutine to handle mock response
// 	go func() {
// 		// Loop until we get a request
// 		var msg JSONRPCMessage
// 		for {
// 			wbs := writer.getWritten()
// 			if len(wbs) == 0 {
// 				continue
// 			}
// 			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
// 			if err != nil {
// 				t.Errorf("failed to decode request: %v", err)
// 				return
// 			}
// 			break
// 		}
//
// 		// Verify request
// 		if msg.Method != MethodPromptsGet {
// 			t.Errorf("expected method %s, got %s", MethodPromptsGet, msg.Method)
// 		}
//
// 		// Verify params
// 		var params PromptsGetParams
// 		if err := json.Unmarshal(msg.Params, &params); err != nil {
// 			t.Errorf("failed to decode params: %v", err)
// 			return
// 		}
// 		if params.Name != "test-prompt" {
// 			t.Errorf("expected prompt name test-prompt, got %s", params.Name)
// 		}
// 		if params.Meta.ProgressToken != "123" {
// 			t.Errorf("expected progress token 123, got %s", params.Meta.ProgressToken)
// 		}
// 		testArg, ok := params.Arguments["test-arg"]
// 		if !ok {
// 			t.Error("test-arg not found in arguments")
// 			return
// 		}
// 		if testArg != "test-value" {
// 			t.Errorf("expected argument value test-value, got %v", testArg)
// 		}
//
// 		// Send mock response
// 		mockResponse := PromptResult{
// 			Description: "Test Prompt",
// 			Messages: []PromptMessage{
// 				{
// 					Role: PromptRoleAssistant,
// 					Content: Content{
// 						Type: ContentTypeText,
// 						Text: "Test response message",
// 					},
// 				},
// 			},
// 		}
// 		mockBs, err := json.Marshal(mockResponse)
// 		if err != nil {
// 			t.Errorf("failed to marshal mock response: %v", err)
// 		}
//
// 		responseMsg := JSONRPCMessage{
// 			JSONRPC: JSONRPCVersion,
// 			ID:      msg.ID,
// 			Result:  mockBs,
// 		}
//
// 		responseBs, _ := json.Marshal(responseMsg)
// 		err = c.handleMsg(bytes.NewReader(responseBs), sessID)
// 		if err != nil {
// 			t.Errorf("handleMsg failed: %v", err)
// 		}
// 	}()
//
// 	// Call getPrompt
// 	ctx = ctxWithSessionID(ctx, sessID)
// 	result, err := c.getPrompt(ctx, "test-prompt", map[string]string{
// 		"test-arg": "test-value",
// 	}, "123")
// 	if err != nil {
// 		t.Fatalf("getPrompt failed: %v", err)
// 	}
//
// 	// Verify result
// 	if result.Description != "Test Prompt" {
// 		t.Errorf("expected description 'Test Prompt', got %s", result.Description)
// 	}
// 	if len(result.Messages) != 1 {
// 		t.Errorf("expected 1 message, got %d", len(result.Messages))
// 	}
// 	if result.Messages[0].Role != PromptRoleAssistant {
// 		t.Errorf("expected message role %s, got %s", PromptRoleAssistant, result.Messages[0].Role)
// 	}
// 	if result.Messages[0].Content.Type != ContentTypeText {
// 		t.Errorf("expected content type %s, got %s", ContentTypeText, result.Messages[0].Content.Type)
// 	}
// 	if result.Messages[0].Content.Text != "Test response message" {
// 		t.Errorf("expected message text 'Test response message', got %s", result.Messages[0].Content.Text)
// 	}
// }
//
// func TestClientCompletesPrompt(t *testing.T) {
// 	writer := &mockWriter{}
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second), WithClientReadTimeout(time.Second))
//
// 	ctx := context.Background()
// 	sessID := "test-session"
// 	c.startSession(ctx, writer, sessID)
//
// 	// Start goroutine to handle mock response
// 	go func() {
// 		// Loop until we get a request
// 		var msg JSONRPCMessage
// 		for {
// 			wbs := writer.getWritten()
// 			if len(wbs) == 0 {
// 				continue
// 			}
// 			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
// 			if err != nil {
// 				t.Errorf("failed to decode request: %v", err)
// 				return
// 			}
// 			break
// 		}
//
// 		// Verify request
// 		if msg.Method != MethodCompletionComplete {
// 			t.Errorf("expected method %s, got %s", MethodCompletionComplete, msg.Method)
// 		}
//
// 		// Verify params
// 		var params CompletionCompleteParams
// 		if err := json.Unmarshal(msg.Params, &params); err != nil {
// 			t.Errorf("failed to decode params: %v", err)
// 			return
// 		}
// 		if params.Ref.Type != CompletionRefPrompt {
// 			t.Errorf("expected ref type ref/prompt, got %s", params.Ref.Type)
// 		}
// 		if params.Ref.Name != "test-prompt" {
// 			t.Errorf("expected prompt name test-prompt, got %s", params.Ref.Name)
// 		}
// 		if params.Argument.Name != "test-arg" {
// 			t.Errorf("expected argument name test-arg, got %s", params.Argument.Name)
// 		}
// 		if params.Argument.Value != "test-" {
// 			t.Errorf("expected argument value test-, got %s", params.Argument.Value)
// 		}
//
// 		// Send mock response
// 		mockResponse := CompletionResult{
// 			Completion: struct {
// 				Values  []string `json:"values"`
// 				HasMore bool     `json:"hasMore"`
// 			}{
// 				Values:  []string{"test-value1", "test-value2"},
// 				HasMore: true,
// 			},
// 		}
// 		mockBs, err := json.Marshal(mockResponse)
// 		if err != nil {
// 			t.Errorf("failed to marshal mock response: %v", err)
// 		}
//
// 		responseMsg := JSONRPCMessage{
// 			JSONRPC: JSONRPCVersion,
// 			ID:      msg.ID,
// 			Result:  mockBs,
// 		}
//
// 		responseBs, _ := json.Marshal(responseMsg)
// 		err = c.handleMsg(bytes.NewReader(responseBs), sessID)
// 		if err != nil {
// 			t.Errorf("handleMsg failed: %v", err)
// 		}
// 	}()
//
// 	// Call completesPrompt
// 	ctx = ctxWithSessionID(ctx, sessID)
// 	result, err := c.completesPrompt(ctx, "test-prompt", CompletionArgument{
// 		Name:  "test-arg",
// 		Value: "test-",
// 	})
// 	if err != nil {
// 		t.Fatalf("completesPrompt failed: %v", err)
// 	}
//
// 	// Verify result
// 	if len(result.Completion.Values) != 2 {
// 		t.Errorf("expected 2 completion values, got %d", len(result.Completion.Values))
// 	}
// 	if result.Completion.Values[0] != "test-value1" {
// 		t.Errorf("expected first completion value test-value1, got %s", result.Completion.Values[0])
// 	}
// 	if result.Completion.Values[1] != "test-value2" {
// 		t.Errorf("expected second completion value test-value2, got %s", result.Completion.Values[1])
// 	}
// 	if !result.Completion.HasMore {
// 		t.Error("expected HasMore to be true")
// 	}
// }
//
// func TestClientListPrompts(t *testing.T) {
// 	writer := &mockWriter{}
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second), WithClientReadTimeout(time.Second))
//
// 	ctx := context.Background()
// 	sessID := "test-session"
// 	c.startSession(ctx, writer, sessID)
//
// 	// Start goroutine to handle mock response
// 	go func() {
// 		// Loop until we get a request
// 		var msg JSONRPCMessage
// 		for {
// 			wbs := writer.getWritten()
// 			if len(wbs) == 0 {
// 				continue
// 			}
// 			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
// 			if err != nil {
// 				t.Errorf("failed to decode request: %v", err)
// 				return
// 			}
// 			break
// 		}
//
// 		// Verify request
// 		if msg.Method != MethodPromptsList {
// 			t.Errorf("expected method %s, got %s", MethodPromptsList, msg.Method)
// 		}
//
// 		// Send mock response
// 		mockResponse := PromptList{
// 			Prompts: []Prompt{
// 				{
// 					Name:        "test-prompt-1",
// 					Description: "Test Prompt 1",
// 					Arguments: []PromptArgument{
// 						{Name: "arg1", Description: "Argument 1", Required: true},
// 					},
// 				},
// 				{
// 					Name:        "test-prompt-2",
// 					Description: "Test Prompt 2",
// 					Arguments: []PromptArgument{
// 						{Name: "arg2", Description: "Argument 2", Required: false},
// 					},
// 				},
// 			},
// 			NextCursor: "next-page",
// 		}
// 		mockBs, err := json.Marshal(mockResponse)
// 		if err != nil {
// 			t.Errorf("failed to marshal mock response: %v", err)
// 		}
//
// 		responseMsg := JSONRPCMessage{
// 			JSONRPC: JSONRPCVersion,
// 			ID:      msg.ID,
// 			Result:  mockBs,
// 		}
//
// 		responseBs, _ := json.Marshal(responseMsg)
// 		err = c.handleMsg(bytes.NewReader(responseBs), sessID)
// 		if err != nil {
// 			t.Errorf("handleMsg failed: %v", err)
// 		}
// 	}()
//
// 	// Call listPrompts
// 	ctx = ctxWithSessionID(ctx, sessID)
// 	result, err := c.listPrompts(ctx, "", "test-token")
// 	if err != nil {
// 		t.Fatalf("listPrompts failed: %v", err)
// 	}
//
// 	// Verify result
// 	if len(result.Prompts) != 2 {
// 		t.Errorf("expected 2 prompts, got %d", len(result.Prompts))
// 	}
// 	if result.Prompts[0].Name != "test-prompt-1" || result.Prompts[0].Description != "Test Prompt 1" {
// 		t.Errorf("unexpected prompt[0]: %+v", result.Prompts[0])
// 	}
// 	if result.Prompts[1].Name != "test-prompt-2" || result.Prompts[1].Description != "Test Prompt 2" {
// 		t.Errorf("unexpected prompt[1]: %+v", result.Prompts[1])
// 	}
// 	if result.NextCursor != "next-page" {
// 		t.Errorf("expected next cursor 'next-page', got %s", result.NextCursor)
// 	}
// }
//
// func TestClientListResources(t *testing.T) {
// 	writer := &mockWriter{}
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second), WithClientReadTimeout(time.Second))
//
// 	ctx := context.Background()
// 	sessID := "test-session"
// 	c.startSession(ctx, writer, sessID)
//
// 	// Start goroutine to handle mock response
// 	go func() {
// 		// Loop until we get a request
// 		var msg JSONRPCMessage
// 		for {
// 			wbs := writer.getWritten()
// 			if len(wbs) == 0 {
// 				continue
// 			}
// 			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
// 			if err != nil {
// 				t.Errorf("failed to decode request: %v", err)
// 				return
// 			}
// 			break
// 		}
//
// 		// Verify request
// 		if msg.Method != MethodResourcesList {
// 			t.Errorf("expected method %s, got %s", MethodResourcesList, msg.Method)
// 		}
//
// 		// Verify params
// 		var params ResourcesListParams
// 		if err := json.Unmarshal(msg.Params, &params); err != nil {
// 			t.Errorf("failed to decode params: %v", err)
// 			return
// 		}
// 		if params.Cursor != "test-cursor" {
// 			t.Errorf("expected cursor test-cursor, got %s", params.Cursor)
// 		}
// 		if params.Meta.ProgressToken != "123" {
// 			t.Errorf("expected progress token 123, got %s", params.Meta.ProgressToken)
// 		}
//
// 		// Send mock response
// 		mockResponse := ResourceList{
// 			Resources: []Resource{
// 				{
// 					URI:         "test://resource1",
// 					Name:        "Test Resource 1",
// 					Description: "First test resource",
// 					MimeType:    "text/plain",
// 					Text:        "Resource 1 content",
// 				},
// 				{
// 					URI:         "test://resource2",
// 					Name:        "Test Resource 2",
// 					Description: "Second test resource",
// 					MimeType:    "application/json",
// 					Text:        "{\"key\": \"value\"}",
// 				},
// 			},
// 			NextCursor: "next-cursor",
// 		}
// 		mockBs, err := json.Marshal(mockResponse)
// 		if err != nil {
// 			t.Errorf("failed to marshal mock response: %v", err)
// 		}
//
// 		responseMsg := JSONRPCMessage{
// 			JSONRPC: JSONRPCVersion,
// 			ID:      msg.ID,
// 			Result:  mockBs,
// 		}
//
// 		responseBs, _ := json.Marshal(responseMsg)
// 		err = c.handleMsg(bytes.NewReader(responseBs), sessID)
// 		if err != nil {
// 			t.Errorf("handleMsg failed: %v", err)
// 		}
// 	}()
//
// 	// Call listResources
// 	ctx = ctxWithSessionID(ctx, sessID)
// 	result, err := c.listResources(ctx, "test-cursor", "123")
// 	if err != nil {
// 		t.Fatalf("listResources failed: %v", err)
// 	}
//
// 	// Verify result
// 	if len(result.Resources) != 2 {
// 		t.Errorf("expected 2 resources, got %d", len(result.Resources))
// 	}
// 	if result.Resources[0].URI != "test://resource1" || result.Resources[0].Name != "Test Resource 1" {
// 		t.Errorf("unexpected resource[0]: %+v", result.Resources[0])
// 	}
// 	if result.Resources[1].URI != "test://resource2" || result.Resources[1].Name != "Test Resource 2" {
// 		t.Errorf("unexpected resource[1]: %+v", result.Resources[1])
// 	}
// 	if result.NextCursor != "next-cursor" {
// 		t.Errorf("expected next cursor 'next-cursor', got %s", result.NextCursor)
// 	}
// }
//
// func TestClientReadResource(t *testing.T) {
// 	writer := &mockWriter{}
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second), WithClientReadTimeout(time.Second))
//
// 	ctx := context.Background()
// 	sessID := "test-session"
// 	c.startSession(ctx, writer, sessID)
//
// 	// Start goroutine to handle mock response
// 	go func() {
// 		// Loop until we get a request
// 		var msg JSONRPCMessage
// 		for {
// 			wbs := writer.getWritten()
// 			if len(wbs) == 0 {
// 				continue
// 			}
// 			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
// 			if err != nil {
// 				t.Errorf("failed to decode request: %v", err)
// 				return
// 			}
// 			break
// 		}
//
// 		// Verify request
// 		if msg.Method != MethodResourcesRead {
// 			t.Errorf("expected method %s, got %s", MethodResourcesRead, msg.Method)
// 		}
//
// 		// Verify params
// 		var params ResourcesReadParams
// 		if err := json.Unmarshal(msg.Params, &params); err != nil {
// 			t.Errorf("failed to decode params: %v", err)
// 			return
// 		}
// 		if params.URI != "test://resource" {
// 			t.Errorf("expected URI test://resource, got %s", params.URI)
// 		}
// 		if params.Meta.ProgressToken != "123" {
// 			t.Errorf("expected progress token 123, got %s", params.Meta.ProgressToken)
// 		}
//
// 		// Send mock response
// 		mockResponse := Resource{
// 			URI:         "test://resource",
// 			Name:        "Test Resource",
// 			Description: "A test resource",
// 			MimeType:    "text/plain",
// 			Text:        "This is the resource content",
// 		}
// 		mockBs, err := json.Marshal(mockResponse)
// 		if err != nil {
// 			t.Errorf("failed to marshal mock response: %v", err)
// 		}
//
// 		responseMsg := JSONRPCMessage{
// 			JSONRPC: JSONRPCVersion,
// 			ID:      msg.ID,
// 			Result:  mockBs,
// 		}
//
// 		responseBs, _ := json.Marshal(responseMsg)
// 		err = c.handleMsg(bytes.NewReader(responseBs), sessID)
// 		if err != nil {
// 			t.Errorf("handleMsg failed: %v", err)
// 		}
// 	}()
//
// 	// Call readResource
// 	ctx = ctxWithSessionID(ctx, sessID)
// 	result, err := c.readResource(ctx, "test://resource", "123")
// 	if err != nil {
// 		t.Fatalf("readResource failed: %v", err)
// 	}
//
// 	// Verify result
// 	if result.URI != "test://resource" {
// 		t.Errorf("expected URI test://resource, got %s", result.URI)
// 	}
// 	if result.Name != "Test Resource" {
// 		t.Errorf("expected name 'Test Resource', got %s", result.Name)
// 	}
// 	if result.Description != "A test resource" {
// 		t.Errorf("expected description 'A test resource', got %s", result.Description)
// 	}
// 	if result.MimeType != "text/plain" {
// 		t.Errorf("expected MIME type text/plain, got %s", result.MimeType)
// 	}
// 	if result.Text != "This is the resource content" {
// 		t.Errorf("expected text content 'This is the resource content', got %s", result.Text)
// 	}
// }
//
// func TestClientListResourceTemplates(t *testing.T) {
// 	writer := &mockWriter{}
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second), WithClientReadTimeout(time.Second))
//
// 	ctx := context.Background()
// 	sessID := "test-session"
// 	c.startSession(ctx, writer, sessID)
//
// 	// Start goroutine to handle mock response
// 	go func() {
// 		// Loop until we get a request
// 		var msg JSONRPCMessage
// 		for {
// 			wbs := writer.getWritten()
// 			if len(wbs) == 0 {
// 				continue
// 			}
// 			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
// 			if err != nil {
// 				t.Errorf("failed to decode request: %v", err)
// 				return
// 			}
// 			break
// 		}
//
// 		// Verify request
// 		if msg.Method != MethodResourcesTemplatesList {
// 			t.Errorf("expected method %s, got %s", MethodResourcesTemplatesList, msg.Method)
// 		}
//
// 		// Verify params
// 		var params ResourcesTemplatesListParams
// 		if err := json.Unmarshal(msg.Params, &params); err != nil {
// 			t.Errorf("failed to decode params: %v", err)
// 			return
// 		}
// 		if params.Meta.ProgressToken != "123" {
// 			t.Errorf("expected progress token 123, got %s", params.Meta.ProgressToken)
// 		}
//
// 		// Send mock response
// 		mockResponse := []ResourceTemplate{
// 			{
// 				URITemplate: "test://resource/{name}",
// 				Name:        "Test Template 1",
// 				Description: "First test template",
// 				MimeType:    "text/plain",
// 			},
// 			{
// 				URITemplate: "test://resource/{id}",
// 				Name:        "Test Template 2",
// 				Description: "Second test template",
// 				MimeType:    "application/json",
// 			},
// 		}
// 		mockBs, err := json.Marshal(mockResponse)
// 		if err != nil {
// 			t.Errorf("failed to marshal mock response: %v", err)
// 		}
//
// 		responseMsg := JSONRPCMessage{
// 			JSONRPC: JSONRPCVersion,
// 			ID:      msg.ID,
// 			Result:  mockBs,
// 		}
//
// 		responseBs, _ := json.Marshal(responseMsg)
// 		err = c.handleMsg(bytes.NewReader(responseBs), sessID)
// 		if err != nil {
// 			t.Errorf("handleMsg failed: %v", err)
// 		}
// 	}()
//
// 	// Call listResourceTemplates
// 	ctx = ctxWithSessionID(ctx, sessID)
// 	result, err := c.listResourceTemplates(ctx, "123")
// 	if err != nil {
// 		t.Fatalf("listResourceTemplates failed: %v", err)
// 	}
//
// 	// Verify result
// 	if len(result) != 2 {
// 		t.Errorf("expected 2 templates, got %d", len(result))
// 	}
// 	if result[0].URITemplate != "test://resource/{name}" || result[0].Name != "Test Template 1" {
// 		t.Errorf("unexpected template[0]: %+v", result[0])
// 	}
// 	if result[1].URITemplate != "test://resource/{id}" || result[1].Name != "Test Template 2" {
// 		t.Errorf("unexpected template[1]: %+v", result[1])
// 	}
// }
//
// func TestClientCompletesResourceTemplate(t *testing.T) {
// 	writer := &mockWriter{}
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second), WithClientReadTimeout(time.Second))
//
// 	ctx := context.Background()
// 	sessID := "test-session"
// 	c.startSession(ctx, writer, sessID)
//
// 	// Start goroutine to handle mock response
// 	go func() {
// 		// Loop until we get a request
// 		var msg JSONRPCMessage
// 		for {
// 			wbs := writer.getWritten()
// 			if len(wbs) == 0 {
// 				continue
// 			}
// 			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
// 			if err != nil {
// 				t.Errorf("failed to decode request: %v", err)
// 				return
// 			}
// 			break
// 		}
//
// 		// Verify request
// 		if msg.Method != MethodCompletionComplete {
// 			t.Errorf("expected method %s, got %s", MethodCompletionComplete, msg.Method)
// 		}
//
// 		// Verify params
// 		var params CompletionCompleteParams
// 		if err := json.Unmarshal(msg.Params, &params); err != nil {
// 			t.Errorf("failed to decode params: %v", err)
// 			return
// 		}
// 		if params.Ref.Type != CompletionRefResource {
// 			t.Errorf("expected ref type ref/resource, got %s", params.Ref.Type)
// 		}
// 		if params.Ref.URI != "test://resource/{name}" {
// 			t.Errorf("expected URI test://resource/{name}, got %s", params.Ref.URI)
// 		}
// 		if params.Argument.Name != "name" {
// 			t.Errorf("expected argument name 'name', got %s", params.Argument.Name)
// 		}
// 		if params.Argument.Value != "test" {
// 			t.Errorf("expected argument value 'test', got %s", params.Argument.Value)
// 		}
//
// 		// Send mock response
// 		mockResponse := CompletionResult{
// 			Completion: struct {
// 				Values  []string `json:"values"`
// 				HasMore bool     `json:"hasMore"`
// 			}{
// 				Values:  []string{"test-resource1", "test-resource2"},
// 				HasMore: true,
// 			},
// 		}
// 		mockBs, err := json.Marshal(mockResponse)
// 		if err != nil {
// 			t.Errorf("failed to marshal mock response: %v", err)
// 		}
//
// 		responseMsg := JSONRPCMessage{
// 			JSONRPC: JSONRPCVersion,
// 			ID:      msg.ID,
// 			Result:  mockBs,
// 		}
//
// 		responseBs, _ := json.Marshal(responseMsg)
// 		err = c.handleMsg(bytes.NewReader(responseBs), sessID)
// 		if err != nil {
// 			t.Errorf("handleMsg failed: %v", err)
// 		}
// 	}()
//
// 	// Call completesResourceTemplate
// 	ctx = ctxWithSessionID(ctx, sessID)
// 	result, err := c.completesResourceTemplate(ctx, "test://resource/{name}", CompletionArgument{
// 		Name:  "name",
// 		Value: "test",
// 	})
// 	if err != nil {
// 		t.Fatalf("completesResourceTemplate failed: %v", err)
// 	}
//
// 	// Verify result
// 	if len(result.Completion.Values) != 2 {
// 		t.Errorf("expected 2 completion values, got %d", len(result.Completion.Values))
// 	}
// 	if result.Completion.Values[0] != "test-resource1" {
// 		t.Errorf("expected first completion value test-resource1, got %s", result.Completion.Values[0])
// 	}
// 	if result.Completion.Values[1] != "test-resource2" {
// 		t.Errorf("expected second completion value test-resource2, got %s", result.Completion.Values[1])
// 	}
// 	if !result.Completion.HasMore {
// 		t.Error("expected HasMore to be true")
// 	}
// }
//
// func TestClientSubscribeResource(t *testing.T) {
// 	writer := &mockWriter{}
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second), WithClientReadTimeout(time.Second))
//
// 	ctx := context.Background()
// 	sessID := "test-session"
// 	c.startSession(ctx, writer, sessID)
//
// 	// Start goroutine to handle mock response
// 	go func() {
// 		// Loop until we get a request
// 		var msg JSONRPCMessage
// 		for {
// 			wbs := writer.getWritten()
// 			if len(wbs) == 0 {
// 				continue
// 			}
// 			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
// 			if err != nil {
// 				t.Errorf("failed to decode request: %v", err)
// 				return
// 			}
// 			break
// 		}
//
// 		// Verify request
// 		if msg.Method != MethodResourcesSubscribe {
// 			t.Errorf("expected method %s, got %s", MethodResourcesSubscribe, msg.Method)
// 		}
//
// 		// Verify params
// 		var params ResourcesSubscribeParams
// 		if err := json.Unmarshal(msg.Params, &params); err != nil {
// 			t.Errorf("failed to decode params: %v", err)
// 			return
// 		}
// 		if params.URI != "test://resource" {
// 			t.Errorf("expected URI test://resource, got %s", params.URI)
// 		}
//
// 		// Send mock response
// 		responseMsg := JSONRPCMessage{
// 			JSONRPC: JSONRPCVersion,
// 			ID:      msg.ID,
// 			Result:  json.RawMessage("null"), // Subscribe returns void/null
// 		}
//
// 		responseBs, _ := json.Marshal(responseMsg)
// 		err := c.handleMsg(bytes.NewReader(responseBs), sessID)
// 		if err != nil {
// 			t.Errorf("handleMsg failed: %v", err)
// 		}
// 	}()
//
// 	// Call subscribeResource
// 	ctx = ctxWithSessionID(ctx, sessID)
// 	err := c.subscribeResource(ctx, "test://resource")
// 	if err != nil {
// 		t.Fatalf("subscribeResource failed: %v", err)
// 	}
// }
//
// func TestClientListTools(t *testing.T) {
// 	writer := &mockWriter{}
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second), WithClientReadTimeout(time.Second))
//
// 	ctx := context.Background()
// 	sessID := "test-session"
// 	c.startSession(ctx, writer, sessID)
//
// 	// Start goroutine to handle mock response
// 	go func() {
// 		// Loop until we get a request
// 		var msg JSONRPCMessage
// 		for {
// 			wbs := writer.getWritten()
// 			if len(wbs) == 0 {
// 				continue
// 			}
// 			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
// 			if err != nil {
// 				t.Errorf("failed to decode request: %v", err)
// 				return
// 			}
// 			break
// 		}
//
// 		// Verify request
// 		if msg.Method != MethodToolsList {
// 			t.Errorf("expected method %s, got %s", MethodToolsList, msg.Method)
// 		}
//
// 		// Verify params
// 		var params ToolsListParams
// 		if err := json.Unmarshal(msg.Params, &params); err != nil {
// 			t.Errorf("failed to decode params: %v", err)
// 			return
// 		}
// 		if params.Cursor != "test-cursor" {
// 			t.Errorf("expected cursor test-cursor, got %s", params.Cursor)
// 		}
// 		if params.Meta.ProgressToken != "123" {
// 			t.Errorf("expected progress token 123, got %s", params.Meta.ProgressToken)
// 		}
//
// 		// Send mock response
// 		mockResponse := ToolList{
// 			Tools: []Tool{
// 				{
// 					Name:        "test-tool-1",
// 					Description: "First test tool",
// 					InputSchema: jsonschema.Must(`{
//             "type": "object",
//             "properties": {
//               "param1": { "type": "string" }
//             }
//           }`),
// 				},
// 				{
// 					Name:        "test-tool-2",
// 					Description: "Second test tool",
// 					InputSchema: jsonschema.Must(`{
//             "type": "object",
//             "properties": {
//               "param2": { "type": "number" }
//             }
//           }`),
// 				},
// 			},
// 			NextCursor: "next-cursor",
// 		}
// 		mockBs, err := json.Marshal(mockResponse)
// 		if err != nil {
// 			t.Errorf("failed to marshal mock response: %v", err)
// 		}
//
// 		responseMsg := JSONRPCMessage{
// 			JSONRPC: JSONRPCVersion,
// 			ID:      msg.ID,
// 			Result:  mockBs,
// 		}
//
// 		responseBs, _ := json.Marshal(responseMsg)
// 		err = c.handleMsg(bytes.NewReader(responseBs), sessID)
// 		if err != nil {
// 			t.Errorf("handleMsg failed: %v", err)
// 		}
// 	}()
//
// 	// Call listTools
// 	ctx = ctxWithSessionID(ctx, sessID)
// 	result, err := c.listTools(ctx, "test-cursor", "123")
// 	if err != nil {
// 		t.Fatalf("listTools failed: %v", err)
// 	}
//
// 	// Verify result
// 	if len(result.Tools) != 2 {
// 		t.Errorf("expected 2 tools, got %d", len(result.Tools))
// 	}
// 	if result.Tools[0].Name != "test-tool-1" || result.Tools[0].Description != "First test tool" {
// 		t.Errorf("unexpected tool[0]: %+v", result.Tools[0])
// 	}
// 	if result.Tools[1].Name != "test-tool-2" || result.Tools[1].Description != "Second test tool" {
// 		t.Errorf("unexpected tool[1]: %+v", result.Tools[1])
// 	}
// 	if result.NextCursor != "next-cursor" {
// 		t.Errorf("expected next cursor 'next-cursor', got %s", result.NextCursor)
// 	}
// }
//
// func TestClientInitialize(t *testing.T) {
// 	writer := &mockWriter{}
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second), WithClientReadTimeout(time.Second))
//
// 	ctx := context.Background()
// 	sessID := "test-session"
// 	c.startSession(ctx, writer, sessID)
//
// 	// Start goroutine to handle mock response
// 	go func() {
// 		// Loop until we get a request
// 		var msg JSONRPCMessage
// 		for {
// 			wbs := writer.getWritten()
// 			if len(wbs) == 0 {
// 				continue
// 			}
// 			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
// 			if err != nil {
// 				t.Errorf("failed to decode request: %v", err)
// 				return
// 			}
// 			break
// 		}
//
// 		// Verify request
// 		if msg.Method != methodInitialize {
// 			t.Errorf("expected method %s, got %s", methodInitialize, msg.Method)
// 		}
//
// 		// Verify params
// 		var params initializeParams
// 		if err := json.Unmarshal(msg.Params, &params); err != nil {
// 			t.Errorf("failed to decode params: %v", err)
// 			return
// 		}
//
// 		if params.ProtocolVersion != protocolVersion {
// 			t.Errorf("expected protocol version %s, got %s", protocolVersion, params.ProtocolVersion)
// 		}
//
// 		if params.ClientInfo.Name != "test-client" {
// 			t.Errorf("expected client name test-client, got %s", params.ClientInfo.Name)
// 		}
//
// 		// Send mock response
// 		mockResponse := initializeResult{
// 			ProtocolVersion: protocolVersion,
// 			Capabilities: ServerCapabilities{
// 				Prompts: &PromptsCapability{
// 					ListChanged: true,
// 				},
// 				Resources: &ResourcesCapability{
// 					Subscribe:   true,
// 					ListChanged: true,
// 				},
// 				Tools: &ToolsCapability{
// 					ListChanged: true,
// 				},
// 				Logging: &LoggingCapability{},
// 			},
// 			ServerInfo: Info{
// 				Name:    "test-server",
// 				Version: "1.0",
// 			},
// 		}
// 		mockBs, err := json.Marshal(mockResponse)
// 		if err != nil {
// 			t.Errorf("failed to marshal mock response: %v", err)
// 		}
//
// 		responseMsg := JSONRPCMessage{
// 			JSONRPC: JSONRPCVersion,
// 			ID:      msg.ID,
// 			Result:  mockBs,
// 		}
//
// 		responseBs, _ := json.Marshal(responseMsg)
// 		err = c.handleMsg(bytes.NewReader(responseBs), sessID)
// 		if err != nil {
// 			t.Errorf("handleMsg failed: %v", err)
// 		}
// 	}()
//
// 	ctx = ctxWithSessionID(ctx, sessID)
// 	err := c.initialize(ctx)
// 	if err != nil {
// 		t.Fatalf("initialize failed: %v", err)
// 	}
// }
//
// func TestClientCallTool(t *testing.T) {
// 	writer := &mockWriter{}
// 	cli := &mockClient{}
// 	c := newClient(cli, WithClientWriteTimeout(time.Second), WithClientReadTimeout(time.Second))
//
// 	ctx := context.Background()
// 	sessID := "test-session"
// 	c.startSession(ctx, writer, sessID)
//
// 	// Start goroutine to handle mock response
// 	go func() {
// 		// Loop until we get a request
// 		var msg JSONRPCMessage
// 		for {
// 			wbs := writer.getWritten()
// 			if len(wbs) == 0 {
// 				continue
// 			}
// 			err := json.NewDecoder(bytes.NewReader(wbs)).Decode(&msg)
// 			if err != nil {
// 				t.Errorf("failed to decode request: %v", err)
// 				return
// 			}
// 			break
// 		}
//
// 		// Verify request
// 		if msg.Method != MethodToolsCall {
// 			t.Errorf("expected method %s, got %s", MethodToolsCall, msg.Method)
// 		}
//
// 		// Verify params
// 		var params ToolsCallParams
// 		if err := json.Unmarshal(msg.Params, &params); err != nil {
// 			t.Errorf("failed to decode params: %v", err)
// 			return
// 		}
// 		if params.Name != "test-tool" {
// 			t.Errorf("expected tool name test-tool, got %s", params.Name)
// 		}
// 		if params.Meta.ProgressToken != "123" {
// 			t.Errorf("expected progress token 123, got %s", params.Meta.ProgressToken)
// 		}
// 		testArg, ok := params.Arguments["test-arg"]
// 		if !ok {
// 			t.Error("test-arg not found in arguments")
// 			return
// 		}
// 		if testArg != "test-value" {
// 			t.Errorf("expected argument value test-value, got %v", testArg)
// 		}
//
// 		// Send mock response
// 		mockResponse := ToolResult{
// 			Content: []Content{
// 				{
// 					Type: ContentTypeText,
// 					Text: "Tool execution result",
// 				},
// 				{
// 					Type:     ContentTypeImage,
// 					Data:     "base64data",
// 					MimeType: "image/png",
// 				},
// 			},
// 			IsError: false,
// 		}
// 		mockBs, err := json.Marshal(mockResponse)
// 		if err != nil {
// 			t.Errorf("failed to marshal mock response: %v", err)
// 		}
//
// 		responseMsg := JSONRPCMessage{
// 			JSONRPC: JSONRPCVersion,
// 			ID:      msg.ID,
// 			Result:  mockBs,
// 		}
//
// 		responseBs, _ := json.Marshal(responseMsg)
// 		err = c.handleMsg(bytes.NewReader(responseBs), sessID)
// 		if err != nil {
// 			t.Errorf("handleMsg failed: %v", err)
// 		}
// 	}()
//
// 	// Call callTool
// 	ctx = ctxWithSessionID(ctx, sessID)
// 	result, err := c.callTool(ctx, "test-tool", map[string]any{
// 		"test-arg": "test-value",
// 	}, "123")
// 	if err != nil {
// 		t.Fatalf("callTool failed: %v", err)
// 	}
//
// 	// Verify result
// 	if len(result.Content) != 2 {
// 		t.Errorf("expected 2 content items, got %d", len(result.Content))
// 	}
// 	if result.Content[0].Type != ContentTypeText || result.Content[0].Text != "Tool execution result" {
// 		t.Errorf("unexpected content[0]: %+v", result.Content[0])
// 	}
// 	if result.Content[1].Type != ContentTypeImage ||
// 		result.Content[1].Data != "base64data" ||
// 		result.Content[1].MimeType != "image/png" {
// 		t.Errorf("unexpected content[1]: %+v", result.Content[1])
// 	}
// 	if result.IsError {
// 		t.Error("expected IsError to be false")
// 	}
// }
