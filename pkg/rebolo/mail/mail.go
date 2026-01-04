package mail

import (
	"bytes"
	"fmt"
	"io"
	"net/smtp"
	"strings"
	"sync"
)

// Message represents an email message
type Message struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	HTMLBody    string
	Headers     map[string]string
	Attachments []Attachment
	mu          sync.Mutex
}

// Attachment represents an email attachment
type Attachment struct {
	Name        string
	ContentType string
	Data        []byte
	Embedded    bool
}

// NewMessage creates a new email message
func NewMessage() *Message {
	return &Message{
		Headers: make(map[string]string),
	}
}

// SetFrom sets the sender email address
func (m *Message) SetFrom(from string) *Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.From = from
	return m
}

// AddTo adds a recipient email address
func (m *Message) AddTo(to string) *Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.To = append(m.To, to)
	return m
}

// AddCc adds a CC recipient email address
func (m *Message) AddCc(cc string) *Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Cc = append(m.Cc, cc)
	return m
}

// AddBcc adds a BCC recipient email address
func (m *Message) AddBcc(bcc string) *Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Bcc = append(m.Bcc, bcc)
	return m
}

// SetSubject sets the email subject
func (m *Message) SetSubject(subject string) *Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Subject = subject
	return m
}

// SetBody sets the plain text body
func (m *Message) SetBody(body string) *Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Body = body
	return m
}

// SetHTMLBody sets the HTML body
func (m *Message) SetHTMLBody(html string) *Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.HTMLBody = html
	return m
}

// AddAttachment adds an attachment to the message
func (m *Message) AddAttachment(name, contentType string, data []byte) *Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Attachments = append(m.Attachments, Attachment{
		Name:        name,
		ContentType: contentType,
		Data:        data,
		Embedded:    false,
	})
	return m
}

// SetHeader sets a custom header
func (m *Message) SetHeader(key, value string) *Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Headers[key] = value
	return m
}

// Sender interface for sending emails
type Sender interface {
	Send(*Message) error
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Auth     smtp.Auth
}

// SMTPSender implements Sender using SMTP
type SMTPSender struct {
	config SMTPConfig
}

// NewSMTPSender creates a new SMTP sender
func NewSMTPSender(host string, port int, username, password string) *SMTPSender {
	auth := smtp.PlainAuth("", username, password, host)
	return &SMTPSender{
		config: SMTPConfig{
			Host:     host,
			Port:     port,
			Username: username,
			Password: password,
			Auth:     auth,
		},
	}
}

// Send sends an email message via SMTP
func (s *SMTPSender) Send(msg *Message) error {
	if msg.From == "" {
		return fmt.Errorf("from address is required")
	}
	if len(msg.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Build email body
	body, err := s.buildBody(msg)
	if err != nil {
		return err
	}

	// Build headers
	headers := s.buildHeaders(msg)

	// Combine headers and body
	email := append(headers, body...)

	// Send email
	recipients := append(msg.To, append(msg.Cc, msg.Bcc...)...)
	return smtp.SendMail(addr, s.config.Auth, msg.From, recipients, email)
}

func (s *SMTPSender) buildHeaders(msg *Message) []byte {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("From: %s\r\n", msg.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))

	if len(msg.Cc) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.Cc, ", ")))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")

	// Add custom headers
	for key, value := range msg.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	return buf.Bytes()
}

func (s *SMTPSender) buildBody(msg *Message) ([]byte, error) {
	var buf bytes.Buffer

	// Simple multipart if we have both text and HTML
	if msg.HTMLBody != "" && msg.Body != "" {
		boundary := "----=_Part_0_" + generateBoundary()

		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary))
		buf.WriteString("\r\n")

		// Text part
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(msg.Body)
		buf.WriteString("\r\n")

		// HTML part
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(msg.HTMLBody)
		buf.WriteString("\r\n")

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if msg.HTMLBody != "" {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(msg.HTMLBody)
	} else {
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(msg.Body)
	}

	buf.WriteString("\r\n")

	return buf.Bytes(), nil
}

func generateBoundary() string {
	return fmt.Sprintf("%d", len("boundary"))
}

// ReadAttachment reads an attachment from an io.Reader
func ReadAttachment(name, contentType string, r io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
