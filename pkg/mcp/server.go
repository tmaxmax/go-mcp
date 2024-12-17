package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Server represents the main MCP server interface that users will implement.
type Server interface {
	Info() Info
	RequiredClientCapabilities() ClientCapabilities
}

// LoggingHandler is an interface for logging.
type LoggingHandler interface {
	LogStream() <-chan LogParams
}

// ServerOption represents the options for the server.
type ServerOption func(*server)

// WithPromptServer sets the prompt server for the server.
func WithPromptServer(srv PromptServer) ServerOption {
	return func(s *server) {
		s.promptServer = srv
	}
}

// WithPromptListUpdater sets the prompt list watcher for the server.
func WithPromptListUpdater(updater PromptListUpdater) ServerOption {
	return func(s *server) {
		s.promptListUpdater = updater
	}
}

// WithResourceServer sets the resource server for the server.
func WithResourceServer(srv ResourceServer) ServerOption {
	return func(s *server) {
		s.resourceServer = srv
	}
}

// WithResourceListUpdater sets the resource list watcher for the server.
func WithResourceListUpdater(updater ResourceListUpdater) ServerOption {
	return func(s *server) {
		s.resourceListUpdater = updater
	}
}

// WithResourceSubscribedUpdater sets the resource subscribe watcher for the server.
func WithResourceSubscribedUpdater(updater ResourceSubscribedUpdater) ServerOption {
	return func(s *server) {
		s.resourceSubscribedUpdater = updater
	}
}

// WithToolServer sets the tool server for the server.
func WithToolServer(srv ToolServer) ServerOption {
	return func(s *server) {
		s.toolServer = srv
	}
}

// WithToolListUpdater sets the tool list watcher for the server.
func WithToolListUpdater(updater ToolListUpdater) ServerOption {
	return func(s *server) {
		s.toolListUpdater = updater
	}
}

// WithRootsListWatcher sets the roots list watcher for the server.
func WithRootsListWatcher(watcher RootsListWatcher) ServerOption {
	return func(s *server) {
		s.rootsListWatcher = watcher
	}
}

// WithLogHandler sets the log handler for the server.
func WithLogHandler(handler LogHandler) ServerOption {
	return func(s *server) {
		s.logHandler = handler
	}
}

// WithProgressReporter sets the progress reporter for the server.
func WithProgressReporter(reporter ProgressReporter) ServerOption {
	return func(s *server) {
		s.progressReporter = reporter
	}
}

// WithWriteTimeout sets the write timeout for the server.
func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(s *server) {
		s.writeTimeout = timeout
	}
}

// WithReadTimeout sets the read timeout for the server.
func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(s *server) {
		s.readTimeout = timeout
	}
}

// WithPingInterval sets the ping interval for the server.
func WithPingInterval(interval time.Duration) ServerOption {
	return func(s *server) {
		s.pingInterval = interval
	}
}

var (
	defaultWriteTimeout = 30 * time.Second
	defaultReadTimeout  = 30 * time.Second
	defaultPingInterval = 30 * time.Second

	errInvalidJSON     = errors.New("invalid json")
	errSessionNotFound = errors.New("session not found")
)

type server struct {
	capabilities               ServerCapabilities
	serverInfo                 Info
	requiredClientCapabilities ClientCapabilities

	sessions   *sync.Map // map[sessionID]*serverSession
	progresses *sync.Map // map[progressToken]sessionID

	promptServer      PromptServer
	promptListUpdater PromptListUpdater

	resourceServer            ResourceServer
	resourceListUpdater       ResourceListUpdater
	resourceSubscribedUpdater ResourceSubscribedUpdater

	toolServer      ToolServer
	toolListUpdater ToolListUpdater

	rootsListWatcher RootsListWatcher

	logHandler       LogHandler
	progressReporter ProgressReporter

	writeTimeout time.Duration
	readTimeout  time.Duration
	pingInterval time.Duration

	sessionStopChan chan string
	closeChan       chan struct{}
}

func newServer(srv Server, options ...ServerOption) server {
	s := server{
		serverInfo:                 srv.Info(),
		requiredClientCapabilities: srv.RequiredClientCapabilities(),
		sessions:                   new(sync.Map),
		progresses:                 new(sync.Map),
		sessionStopChan:            make(chan string),
		closeChan:                  make(chan struct{}),
	}
	for _, opt := range options {
		opt(&s)
	}

	if s.writeTimeout == 0 {
		s.writeTimeout = defaultWriteTimeout
	}
	if s.readTimeout == 0 {
		s.readTimeout = defaultReadTimeout
	}
	if s.pingInterval == 0 {
		s.pingInterval = defaultPingInterval
	}

	s.capabilities = ServerCapabilities{}

	if s.promptServer != nil {
		s.capabilities.Prompts = &PromptsCapability{}
		if s.promptListUpdater != nil {
			s.capabilities.Prompts.ListChanged = true
		}
	}
	if s.resourceServer != nil {
		s.capabilities.Resources = &ResourcesCapability{}
		if s.resourceListUpdater != nil {
			s.capabilities.Resources.ListChanged = true
		}
		if s.resourceSubscribedUpdater != nil {
			s.capabilities.Resources.Subscribe = true
		}
	}
	if s.toolServer != nil {
		s.capabilities.Tools = &ToolsCapability{}
		if s.toolListUpdater != nil {
			s.capabilities.Tools.ListChanged = true
		}
	}
	if s.logHandler != nil {
		s.capabilities.Logging = &LoggingCapability{}
	}

	return s
}

func (s server) start() {
	go s.listenStopSession()
	if s.promptListUpdater != nil {
		go s.listenPromptsList()
	}
	if s.resourceListUpdater != nil {
		go s.listenResourcesList()
	}
	if s.resourceSubscribedUpdater != nil {
		go s.listenResourcesSubscribe()
	}
	if s.toolListUpdater != nil {
		go s.listenToolsList()
	}

	if s.logHandler != nil {
		go s.listenLog()
	}
	if s.progressReporter != nil {
		go s.listenProgress()
	}
}

func (s server) listenStopSession() {
	for {
		var id string
		select {
		case <-s.closeChan:
			return
		case id = <-s.sessionStopChan:
		}
		s.sessions.Delete(id)
	}
}

func (s server) listenPromptsList() {
	lists := s.promptListUpdater.PromptListUpdates()

	for {
		select {
		case <-s.closeChan:
			return
		case <-lists:
		}

		s.sessions.Range(func(_, value any) bool {
			sess, _ := value.(*serverSession)
			sess.promptsListChan <- struct{}{}
			return true
		})
	}
}

func (s server) listenResourcesList() {
	lists := s.resourceListUpdater.ResourceListUpdates()

	for {
		select {
		case <-s.closeChan:
			return
		case <-lists:
		}

		s.sessions.Range(func(_, value any) bool {
			sess, _ := value.(*serverSession)
			sess.resourcesListChan <- struct{}{}
			return true
		})
	}
}

func (s server) listenResourcesSubscribe() {
	subscribes := s.resourceSubscribedUpdater.ResourceSubscriberUpdates()
	var uri string

	for {
		select {
		case <-s.closeChan:
			return
		case uri = <-subscribes:
		}

		s.sessions.Range(func(_, value any) bool {
			sess, _ := value.(*serverSession)
			sess.resourcesSubscribeChan <- uri
			return true
		})
	}
}

func (s server) listenToolsList() {
	lists := s.toolListUpdater.WatchToolList()

	for {
		select {
		case <-s.closeChan:
			return
		case <-lists:
		}

		s.sessions.Range(func(_, value any) bool {
			sess, _ := value.(*serverSession)
			sess.toolsListChan <- struct{}{}
			return true
		})
	}
}

func (s server) listenLog() {
	logs := s.logHandler.LogStreams()
	var params LogParams

	for {
		select {
		case <-s.closeChan:
			return
		case params = <-logs:
		}

		s.sessions.Range(func(_, value any) bool {
			sess, _ := value.(*serverSession)
			sess.logChan <- params
			return true
		})
	}
}

func (s server) listenProgress() {
	progresses := s.progressReporter.ProgressReports()
	var params ProgressParams

	for {
		select {
		case <-s.closeChan:
			return
		case params = <-progresses:
		}

		sessID, ok := s.progresses.Load(params.ProgressToken)
		if !ok {
			continue
		}
		ss, ok := s.sessions.Load(sessID)
		if !ok {
			continue
		}
		sess, _ := ss.(*serverSession)
		sess.progressChan <- params
	}
}

func (s server) startSession(ctx context.Context, w io.Writer) string {
	sCtx, sCancel := context.WithCancel(ctx)

	sessID := uuid.New().String()
	sess := &serverSession{
		id:                     sessID,
		ctx:                    sCtx,
		cancel:                 sCancel,
		writter:                w,
		writeTimeout:           s.writeTimeout,
		readTimeout:            s.readTimeout,
		pingInterval:           s.pingInterval,
		stopChan:               s.sessionStopChan,
		promptsListChan:        make(chan struct{}),
		resourcesListChan:      make(chan struct{}),
		resourcesSubscribeChan: make(chan string),
		toolsListChan:          make(chan struct{}),
		logChan:                make(chan LogParams),
		progressChan:           make(chan ProgressParams),
	}

	s.sessions.Store(sessID, sess)
	go sess.listen()

	return sessID
}

func (s server) handleMsg(r io.Reader, sessionID string) error {
	msg, err := readMessage(r)
	if err != nil {
		return errInvalidJSON
	}

	ss, ok := s.sessions.Load(sessionID)
	if !ok {
		return errSessionNotFound
	}
	sess, _ := ss.(*serverSession)

	// Handle basic protocol messages
	if err := s.handleBasicMessages(sess, msg); err != nil {
		return err
	}

	// Handle prompt-related messages
	if err := s.handlePromptMessages(sess, msg); err != nil {
		return err
	}

	// Handle resource-related messages
	if err := s.handleResourceMessages(sess, msg); err != nil {
		return err
	}

	// Handle tool-related messages
	if err := s.handleToolMessages(sess, msg); err != nil {
		return err
	}

	// Handle completion messages
	if err := s.handleCompletionMessages(sess, msg); err != nil {
		return err
	}

	// Handle notification messages
	if err := s.handleNotificationMessages(sess, msg); err != nil {
		return err
	}

	// Handle result messages
	if err := s.handleResultMessages(sess, msg); err != nil {
		return err
	}

	return nil
}

func (s server) handleBasicMessages(sess *serverSession, msg jsonRPCMessage) error {
	switch msg.Method {
	case methodPing:
		return sess.handlePing(msg.ID)
	case methodInitialize:
		var params initializeParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return errInvalidJSON
		}
		return sess.handleInitialize(msg.ID, params, s.capabilities,
			s.requiredClientCapabilities, s.serverInfo)
	}
	return nil
}

func (s server) handlePromptMessages(sess *serverSession, msg jsonRPCMessage) error {
	if s.promptServer == nil {
		return nil
	}

	switch msg.Method {
	case methodPromptsList:
		var params promptsListParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return errInvalidJSON
		}
		return sess.handlePromptsList(msg.ID, params, s.promptServer)
	case methodPromptsGet:
		var params promptsGetParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return errInvalidJSON
		}
		return sess.handlePromptsGet(msg.ID, params, s.promptServer)
	}
	return nil
}

func (s server) handleResourceMessages(sess *serverSession, msg jsonRPCMessage) error {
	if s.resourceServer == nil {
		return nil
	}

	switch msg.Method {
	case methodResourcesList:
		var params resourcesListParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return errInvalidJSON
		}
		return sess.handleResourcesList(msg.ID, params, s.resourceServer)
	case methodResourcesRead:
		var params resourcesReadParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return errInvalidJSON
		}
		return sess.handleResourcesRead(msg.ID, params, s.resourceServer)
	case methodResourcesTemplatesList:
		var params resourcesTemplatesListParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return errInvalidJSON
		}
		return sess.handleResourcesListTemplates(msg.ID, params, s.resourceServer)
	case methodResourcesSubscribe:
		var params resourcesSubscribeParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return errInvalidJSON
		}
		return sess.handleResourcesSubscribe(msg.ID, params, s.resourceServer)
	}
	return nil
}

func (s server) handleToolMessages(sess *serverSession, msg jsonRPCMessage) error {
	if s.toolServer == nil {
		return nil
	}

	switch msg.Method {
	case methodToolsList:
		var params toolsListParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return errInvalidJSON
		}
		return sess.handleToolsList(msg.ID, params, s.toolServer)
	case methodToolsCall:
		var params toolsCallParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return errInvalidJSON
		}
		return sess.handleToolsCall(msg.ID, params, s.toolServer)
	}
	return nil
}

func (s server) handleCompletionMessages(sess *serverSession, msg jsonRPCMessage) error {
	if msg.Method != methodCompletionComplete {
		return nil
	}

	var params completionCompleteParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return errInvalidJSON
	}

	switch params.Ref.Type {
	case "ref/prompt":
		return sess.handleCompletePrompt(msg.ID, params.Ref.Name, params.Argument, s.promptServer)
	case "ref/resource":
		return sess.handleCompleteResource(msg.ID, params.Ref.Name, params.Argument, s.resourceServer)
	}
	return nil
}

func (s server) handleNotificationMessages(sess *serverSession, msg jsonRPCMessage) error {
	switch msg.Method {
	case methodNotificationsInitialized:
		sess.handleNotificationsInitialized()
	case methodNotificationsCancelled:
		var params notificationsCancelledParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return errInvalidJSON
		}
		sess.handleNotificationsCancelled(params)
	case methodNotificationsRootsListChanged:
		if s.rootsListWatcher != nil {
			s.rootsListWatcher.OnRootsListChanged()
		}
	}

	return nil
}

func (s server) handleResultMessages(sess *serverSession, msg jsonRPCMessage) error {
	if msg.Method != "" {
		return nil
	}

	return sess.handleResult(msg)
}

func (s server) rootsList(ctx context.Context) (RootList, error) {
	ss, ok := s.sessions.Load(sessionIDFromContext(ctx))
	if !ok {
		return RootList{}, errSessionNotFound
	}
	sess, _ := ss.(*serverSession)
	return sess.rootsList()
}

func (s server) createSampleMessage(ctx context.Context, params SamplingParams) (SamplingResult, error) {
	ss, ok := s.sessions.Load(sessionIDFromContext(ctx))
	if !ok {
		return SamplingResult{}, errSessionNotFound
	}
	sess, _ := ss.(*serverSession)
	return sess.createSampleMessage(params)
}

func (s server) stop() {
	s.sessions.Range(func(_, value any) bool {
		sess, _ := value.(*serverSession)
		sess.cancel()
		return true
	})
	close(s.closeChan)
}
