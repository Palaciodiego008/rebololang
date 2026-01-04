package session

import (
	"fmt"
	"html/template"
)

// FlashMessage represents a flash message with a type and content
type FlashMessage struct {
	Type    string // success, error, warning, info
	Message string
}

// Flash provides convenient methods for flash messages
type Flash struct {
	session *Session
}

// NewFlash creates a new Flash instance
func NewFlash(session *Session) *Flash {
	return &Flash{session: session}
}

// Add adds a flash message with the specified type
func (f *Flash) Add(msgType, message string) {
	f.session.AddFlash(FlashMessage{
		Type:    msgType,
		Message: message,
	})
}

// Success adds a success flash message
func (f *Flash) Success(message string) {
	f.Add("success", message)
}

// Error adds an error flash message
func (f *Flash) Error(message string) {
	f.Add("error", message)
}

// Warning adds a warning flash message
func (f *Flash) Warning(message string) {
	f.Add("warning", message)
}

// Info adds an info flash message
func (f *Flash) Info(message string) {
	f.Add("info", message)
}

// Get retrieves all flash messages and clears them
func (f *Flash) Get() []FlashMessage {
	flashes := f.session.Flashes()
	var messages []FlashMessage

	for _, flash := range flashes {
		if msg, ok := flash.(FlashMessage); ok {
			messages = append(messages, msg)
		}
	}

	return messages
}

// GetByType retrieves flash messages of a specific type
func (f *Flash) GetByType(msgType string) []string {
	messages := f.Get()
	var result []string

	for _, msg := range messages {
		if msg.Type == msgType {
			result = append(result, msg.Message)
		}
	}

	return result
}

// HTML returns HTML for all flash messages
func (f *Flash) HTML() template.HTML {
	messages := f.Get()
	if len(messages) == 0 {
		return ""
	}

	html := ""
	for _, msg := range messages {
		alertClass := getAlertClass(msg.Type)
		html += fmt.Sprintf(`<div class="alert alert-%s" role="alert">%s</div>`, alertClass, msg.Message)
	}

	return template.HTML(html)
}

// getAlertClass returns Bootstrap-compatible alert classes
func getAlertClass(msgType string) string {
	switch msgType {
	case "success":
		return "success"
	case "error":
		return "danger"
	case "warning":
		return "warning"
	case "info":
		return "info"
	default:
		return "secondary"
	}
}
