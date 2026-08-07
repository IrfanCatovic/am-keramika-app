package mailer

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Mailer sends plain-text emails. Implementations must be safe for concurrent use.
type Mailer interface {
	Send(to, subject, body string) error
	Configured() bool
}

// NoopMailer never sends; used when SMTP is not configured.
type NoopMailer struct{}

func (NoopMailer) Send(to, subject, body string) error { return nil }
func (NoopMailer) Configured() bool                    { return false }

// RecordingMailer stores sent messages for tests.
type RecordingMailer struct {
	mu      sync.Mutex
	Sent    []SentMessage
	FailErr error
}

type SentMessage struct {
	To      string
	Subject string
	Body    string
}

func (m *RecordingMailer) Configured() bool { return true }

func (m *RecordingMailer) Send(to, subject, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FailErr != nil {
		return m.FailErr
	}
	m.Sent = append(m.Sent, SentMessage{To: to, Subject: subject, Body: body})
	return nil
}

func (m *RecordingMailer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Sent = nil
	m.FailErr = nil
}

func (m *RecordingMailer) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Sent)
}

func (m *RecordingMailer) Last() (SentMessage, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Sent) == 0 {
		return SentMessage{}, false
	}
	return m.Sent[len(m.Sent)-1], true
}

func (m *RecordingMailer) SetFailErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FailErr = err
}

// SMTPMailer sends via net/smtp using env configuration.
type SMTPMailer struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func (m *SMTPMailer) Configured() bool {
	return m != nil && m.host != "" && m.port > 0 && m.from != ""
}

func (m *SMTPMailer) Send(to, subject, body string) error {
	if !m.Configured() {
		return fmt.Errorf("smtp nije konfigurisan")
	}
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	msg := strings.Join([]string{
		"From: " + m.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}
	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
}

// NewFromEnv builds a Mailer from SMTP_* env vars. Missing config → NoopMailer.
func NewFromEnv() Mailer {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	portStr := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	if host == "" || from == "" || portStr == "" {
		return NoopMailer{}
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		return NoopMailer{}
	}
	return &SMTPMailer{
		host:     host,
		port:     port,
		username: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     from,
	}
}

func OrderNotificationEmail() string {
	return strings.TrimSpace(os.Getenv("ORDER_NOTIFICATION_EMAIL"))
}

func FrontendAppURL() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_APP_URL")), "/")
}
