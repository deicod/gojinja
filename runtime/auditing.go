package runtime

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditLevel represents the level of audit logging
type AuditLevel int

const (
	AuditLevelOff AuditLevel = iota
	AuditLevelDebug
	AuditLevelInfo
	AuditLevelWarning
	AuditLevelError
)

// String returns the string representation of the audit level
func (al AuditLevel) String() string {
	switch al {
	case AuditLevelOff:
		return "off"
	case AuditLevelDebug:
		return "debug"
	case AuditLevelInfo:
		return "info"
	case AuditLevelWarning:
		return "warning"
	case AuditLevelError:
		return "error"
	default:
		return "unknown"
	}
}

// AuditEventType represents different types of audit events
type AuditEventType int

const (
	AuditEventSecurityViolation AuditEventType = iota
	AuditEventTemplateAccess
	AuditEventFilterAccess
	AuditEventFunctionAccess
	AuditEventAttributeAccess
	AuditEventMethodCall
	AuditEventExecutionStart
	AuditEventExecutionEnd
	AuditEventExecutionTimeout
	AuditEventMemoryLimitExceeded
	AuditEventOutputLimitExceeded
	AuditEventRecursionLimitExceeded
	AuditEventInputValidation
	AuditEventPolicyViolation
	AuditEventSystemEvent
)

// String returns the string representation of the audit event type
func (aet AuditEventType) String() string {
	switch aet {
	case AuditEventSecurityViolation:
		return "security_violation"
	case AuditEventTemplateAccess:
		return "template_access"
	case AuditEventFilterAccess:
		return "filter_access"
	case AuditEventFunctionAccess:
		return "function_access"
	case AuditEventAttributeAccess:
		return "attribute_access"
	case AuditEventMethodCall:
		return "method_call"
	case AuditEventExecutionStart:
		return "execution_start"
	case AuditEventExecutionEnd:
		return "execution_end"
	case AuditEventExecutionTimeout:
		return "execution_timeout"
	case AuditEventMemoryLimitExceeded:
		return "memory_limit_exceeded"
	case AuditEventOutputLimitExceeded:
		return "output_limit_exceeded"
	case AuditEventRecursionLimitExceeded:
		return "recursion_limit_exceeded"
	case AuditEventInputValidation:
		return "input_validation"
	case AuditEventPolicyViolation:
		return "policy_violation"
	case AuditEventSystemEvent:
		return "system_event"
	default:
		return "unknown"
	}
}

// AuditEvent represents a single audit event
type AuditEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       AuditLevel             `json:"level"`
	Type        AuditEventType         `json:"type"`
	Message     string                 `json:"message"`
	Template    string                 `json:"template,omitempty"`
	Context     string                 `json:"context,omitempty"`
	Resource    string                 `json:"resource,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Policy      string                 `json:"policy,omitempty"`
	Violation   *SecurityViolation     `json:"violation,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Success     bool                   `json:"success,omitempty"`
	ErrorMessage string                `json:"error_message,omitempty"`
}

// AuditLogger interface for different audit logging backends
type AuditLogger interface {
	LogEvent(event *AuditEvent) error
	Close() error
}

// FileAuditLogger writes audit events to a file
type FileAuditLogger struct {
	filePath   string
	file       *os.File
	mu         sync.Mutex
	maxSize    int64
	maxBackups int
}

// NewFileAuditLogger creates a new file audit logger
func NewFileAuditLogger(filePath string, maxSize int64, maxBackups int) (*FileAuditLogger, error) {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	// Open file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	return &FileAuditLogger{
		filePath:   filePath,
		file:       file,
		maxSize:    maxSize,
		maxBackups: maxBackups,
	}, nil
}

// LogEvent writes an audit event to the file
func (fal *FileAuditLogger) LogEvent(event *AuditEvent) error {
	fal.mu.Lock()
	defer fal.mu.Unlock()

	// Check file size and rotate if necessary
	if fal.maxSize > 0 {
		stat, err := fal.file.Stat()
		if err == nil && stat.Size() > fal.maxSize {
			if err := fal.rotateFile(); err != nil {
				return fmt.Errorf("failed to rotate audit log file: %w", err)
			}
		}
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Write to file
	_, err = fal.file.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	// Sync to disk
	err = fal.file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync audit log file: %w", err)
	}

	return nil
}

// Close closes the audit log file
func (fal *FileAuditLogger) Close() error {
	fal.mu.Lock()
	defer fal.mu.Unlock()

	if fal.file != nil {
		return fal.file.Close()
	}
	return nil
}

// rotateFile rotates the audit log file
func (fal *FileAuditLogger) rotateFile() error {
	if err := fal.file.Close(); err != nil {
		return err
	}

	// Move existing files
	for i := fal.maxBackups - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", fal.filePath, i)
		newPath := fmt.Sprintf("%s.%d", fal.filePath, i+1)
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	// Move current file
	backupPath := fmt.Sprintf("%s.1", fal.filePath)
	os.Rename(fal.filePath, backupPath)

	// Create new file
	file, err := os.OpenFile(fal.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	fal.file = file
	return nil
}

// ConsoleAuditLogger writes audit events to the console
type ConsoleAuditLogger struct {
	minLevel AuditLevel
	mu       sync.Mutex
}

// NewConsoleAuditLogger creates a new console audit logger
func NewConsoleAuditLogger(minLevel AuditLevel) *ConsoleAuditLogger {
	return &ConsoleAuditLogger{
		minLevel: minLevel,
	}
}

// LogEvent writes an audit event to the console
func (cal *ConsoleAuditLogger) LogEvent(event *AuditEvent) error {
	if event.Level < cal.minLevel {
		return nil
	}

	cal.mu.Lock()
	defer cal.mu.Unlock()

	// Format event
	timestamp := event.Timestamp.Format("2006-01-02 15:04:05")
	level := event.Level.String()
	eventType := event.Type.String()

	message := fmt.Sprintf("[%s] %s %s: %s", timestamp, level, eventType, event.Message)

	// Add additional context
	if event.Template != "" {
		message += fmt.Sprintf(" (template: %s)", event.Template)
	}
	if event.Resource != "" {
		message += fmt.Sprintf(" (resource: %s)", event.Resource)
	}
	if event.ErrorMessage != "" {
		message += fmt.Sprintf(" (error: %s)", event.ErrorMessage)
	}

	// Write to appropriate log output
	switch event.Level {
	case AuditLevelError:
		log.Println("ERROR:", message)
	case AuditLevelWarning:
		log.Println("WARNING:", message)
	case AuditLevelInfo:
		log.Println("INFO:", message)
	case AuditLevelDebug:
		log.Println("DEBUG:", message)
	}

	return nil
}

// Close is a no-op for console logger
func (cal *ConsoleAuditLogger) Close() error {
	return nil
}

// MemoryAuditLogger stores audit events in memory
type MemoryAuditLogger struct {
	events    []*AuditEvent
	maxEvents int
	mu        sync.RWMutex
}

// NewMemoryAuditLogger creates a new memory audit logger
func NewMemoryAuditLogger(maxEvents int) *MemoryAuditLogger {
	return &MemoryAuditLogger{
		events:    make([]*AuditEvent, 0),
		maxEvents: maxEvents,
	}
}

// LogEvent stores an audit event in memory
func (mal *MemoryAuditLogger) LogEvent(event *AuditEvent) error {
	mal.mu.Lock()
	defer mal.mu.Unlock()

	// Add event
	mal.events = append(mal.events, event)

	// Trim if necessary
	if mal.maxEvents > 0 && len(mal.events) > mal.maxEvents {
		mal.events = mal.events[len(mal.events)-mal.maxEvents:]
	}

	return nil
}

// GetEvents returns all stored audit events
func (mal *MemoryAuditLogger) GetEvents() []*AuditEvent {
	mal.mu.RLock()
	defer mal.mu.RUnlock()

	events := make([]*AuditEvent, len(mal.events))
	copy(events, mal.events)

	return events
}

// GetEventsByType returns events filtered by type
func (mal *MemoryAuditLogger) GetEventsByType(eventType AuditEventType) []*AuditEvent {
	mal.mu.RLock()
	defer mal.mu.RUnlock()

	var filtered []*AuditEvent
	for _, event := range mal.events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// GetEventsByLevel returns events filtered by level
func (mal *MemoryAuditLogger) GetEventsByLevel(level AuditLevel) []*AuditEvent {
	mal.mu.RLock()
	defer mal.mu.RUnlock()

	var filtered []*AuditEvent
	for _, event := range mal.events {
		if event.Level >= level {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// Clear clears all stored events
func (mal *MemoryAuditLogger) Clear() {
	mal.mu.Lock()
	defer mal.mu.Unlock()

	mal.events = make([]*AuditEvent, 0)
}

// Close is a no-op for memory logger
func (mal *MemoryAuditLogger) Close() error {
	return nil
}

// MultiAuditLogger writes to multiple audit loggers
type MultiAuditLogger struct {
	loggers []AuditLogger
	mu      sync.RWMutex
}

// NewMultiAuditLogger creates a new multi audit logger
func NewMultiAuditLogger(loggers ...AuditLogger) *MultiAuditLogger {
	return &MultiAuditLogger{
		loggers: loggers,
	}
}

// AddLogger adds an audit logger
func (mal *MultiAuditLogger) AddLogger(logger AuditLogger) {
	mal.mu.Lock()
	defer mal.mu.Unlock()

	mal.loggers = append(mal.loggers, logger)
}

// RemoveLogger removes an audit logger
func (mal *MultiAuditLogger) RemoveLogger(logger AuditLogger) {
	mal.mu.Lock()
	defer mal.mu.Unlock()

	for i, l := range mal.loggers {
		if l == logger {
			mal.loggers = append(mal.loggers[:i], mal.loggers[i+1:]...)
			break
		}
	}
}

// LogEvent writes an audit event to all loggers
func (mal *MultiAuditLogger) LogEvent(event *AuditEvent) error {
	mal.mu.RLock()
	loggers := make([]AuditLogger, len(mal.loggers))
	copy(loggers, mal.loggers)
	mal.mu.RUnlock()

	var errors []error
	for _, logger := range loggers {
		if err := logger.LogEvent(event); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple logger errors: %v", errors)
	}

	return nil
}

// Close closes all loggers
func (mal *MultiAuditLogger) Close() error {
	mal.mu.RLock()
	loggers := make([]AuditLogger, len(mal.loggers))
	copy(loggers, mal.loggers)
	mal.mu.RUnlock()

	var errors []error
	for _, logger := range loggers {
		if err := logger.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple logger close errors: %v", errors)
	}

	return nil
}

// AuditManager manages audit logging for the template engine
type AuditManager struct {
	logger    AuditLogger
	minLevel  AuditLevel
	enabled   bool
	mu        sync.RWMutex
	eventID   uint64
}

// NewAuditManager creates a new audit manager
func NewAuditManager(logger AuditLogger) *AuditManager {
	return &AuditManager{
		logger:   logger,
		minLevel: AuditLevelInfo,
		enabled:  true,
	}
}

// SetLogger sets the audit logger
func (am *AuditManager) SetLogger(logger AuditLogger) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.logger != nil {
		am.logger.Close()
	}

	am.logger = logger
}

// SetMinLevel sets the minimum audit level
func (am *AuditManager) SetMinLevel(level AuditLevel) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.minLevel = level
}

// SetEnabled enables or disables audit logging
func (am *AuditManager) SetEnabled(enabled bool) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.enabled = enabled
}

// LogEvent logs an audit event
func (am *AuditManager) LogEvent(event *AuditEvent) {
	am.mu.RLock()
	enabled := am.enabled
	minLevel := am.minLevel
	logger := am.logger
	am.mu.RUnlock()

	if !enabled || event.Level < minLevel || logger == nil {
		return
	}

	// Generate event ID
	am.mu.Lock()
	am.eventID++
	event.ID = fmt.Sprintf("audit_%d_%d", time.Now().UnixNano(), am.eventID)
	am.mu.Unlock()

	// Log event
	if err := logger.LogEvent(event); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}
}

// LogSecurityViolation logs a security violation
func (am *AuditManager) LogSecurityViolation(violation *SecurityViolation, template, context string) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		Level:     AuditLevelError,
		Type:      AuditEventSecurityViolation,
		Message:   violation.Description,
		Template:  template,
		Context:   context,
		Resource:  violation.Context,
		Violation: violation,
		Success:   !violation.Blocked,
	}

	am.LogEvent(event)
}

// LogTemplateAccess logs template access
func (am *AuditManager) LogTemplateAccess(templateName, userID, sessionID string, allowed bool) {
	level := AuditLevelInfo
	if !allowed {
		level = AuditLevelWarning
	}

	event := &AuditEvent{
		Timestamp: time.Now(),
		Level:     level,
		Type:      AuditEventTemplateAccess,
		Message:   fmt.Sprintf("Template access: %s", templateName),
		Template:  templateName,
		UserID:    userID,
		SessionID: sessionID,
		Resource:  templateName,
		Success:   allowed,
	}

	am.LogEvent(event)
}

// LogExecutionStart logs the start of template execution
func (am *AuditManager) LogExecutionStart(templateName, userID, sessionID, policy string, vars map[string]interface{}) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		Level:     AuditLevelInfo,
		Type:      AuditEventExecutionStart,
		Message:   fmt.Sprintf("Template execution started: %s", templateName),
		Template:  templateName,
		UserID:    userID,
		SessionID: sessionID,
		Policy:    policy,
		Success:   true,
		Metadata: map[string]interface{}{
			"variables_count": len(vars),
		},
	}

	am.LogEvent(event)
}

// LogExecutionEnd logs the end of template execution
func (am *AuditManager) LogExecutionEnd(templateName, userID, sessionID string, duration time.Duration, success bool, errorMsg string) {
	level := AuditLevelInfo
	if !success {
		level = AuditLevelError
	}

	event := &AuditEvent{
		Timestamp:    time.Now(),
		Level:        level,
		Type:         AuditEventExecutionEnd,
		Message:      fmt.Sprintf("Template execution ended: %s", templateName),
		Template:     templateName,
		UserID:       userID,
		SessionID:    sessionID,
		Duration:     duration,
		Success:      success,
		ErrorMessage: errorMsg,
		Metadata: map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
		},
	}

	am.LogEvent(event)
}

// LogResourceAccess logs access to a resource (filter, function, attribute, method)
func (am *AuditManager) LogResourceAccess(resourceType, resourceName, templateName, context string, allowed bool) {
	level := AuditLevelDebug
	if !allowed {
		level = AuditLevelWarning
	}

	var eventType AuditEventType
	switch resourceType {
	case "filter":
		eventType = AuditEventFilterAccess
	case "function":
		eventType = AuditEventFunctionAccess
	case "attribute":
		eventType = AuditEventAttributeAccess
	case "method":
		eventType = AuditEventMethodCall
	default:
		eventType = AuditEventSystemEvent
	}

	event := &AuditEvent{
		Timestamp: time.Now(),
		Level:     level,
		Type:      eventType,
		Message:   fmt.Sprintf("%s access: %s", resourceType, resourceName),
		Template:  templateName,
		Context:   context,
		Resource:  resourceName,
		Success:   allowed,
	}

	am.LogEvent(event)
}

// Close closes the audit manager
func (am *AuditManager) Close() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.logger != nil {
		return am.logger.Close()
	}

	return nil
}

// Global audit manager instance
var globalAuditManager = NewAuditManager(NewConsoleAuditLogger(AuditLevelInfo))

// GetGlobalAuditManager returns the global audit manager
func GetGlobalAuditManager() *AuditManager {
	return globalAuditManager
}

// ConfigureAuditLogging configures the global audit logging
func ConfigureAuditLogging(logger AuditLogger, minLevel AuditLevel) {
	globalAuditManager.SetLogger(logger)
	globalAuditManager.SetMinLevel(minLevel)
}

// EnableAuditLogging enables audit logging
func EnableAuditLogging() {
	globalAuditManager.SetEnabled(true)
}

// DisableAuditLogging disables audit logging
func DisableAuditLogging() {
	globalAuditManager.SetEnabled(false)
}