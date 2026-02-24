package smtp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/xyroscar/common-lib/pkg/config"
	"go.uber.org/zap"
)

var (
	ErrNoRecp    error = errors.New("no recipients specified")
	ErrEmptySub  error = errors.New("subject cannot be empty")
	ErrEmptyBody error = errors.New("body cannot be empty")
)

type Email struct {
	From    string
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

func sanitizeHeader(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", "")
}

func Send(email *Email) error {
	if len(email.To) == 0 {
		return ErrNoRecp
	}
	if email.Subject == "" {
		return ErrEmptySub
	}
	if email.Body == "" {
		return ErrEmptyBody
	}

	c, err := config.GetSmtpConfig()
	if err != nil {
		zap.L().Error("Error loading smtp config", zap.Error(err))
		return err
	}

	from := email.From
	if from == "" {
		from = c.Upstream.From
	}

	zap.L().Debug("Sending email to upstream SMTP")

	mime := "MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n"
	if email.IsHTML {
		mime = "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	}

	message := fmt.Sprintf("From: %s\r\n", sanitizeHeader(from))
	message += fmt.Sprintf("To: %s\r\n", sanitizeHeader(strings.Join(email.To, ", ")))
	message += fmt.Sprintf("Subject: %s\r\n", sanitizeHeader(email.Subject))
	message += mime
	message += email.Body

	auth := smtp.PlainAuth("", c.Upstream.Username, c.Upstream.Password, c.Upstream.Host)

	err = sendWithStartTLS(c.Upstream.Host, c.Upstream.Port, auth, from, email.To, []byte(message))
	if err != nil {
		zap.L().Error("Failed to send email to upstream SMTP", zap.Error(err))
		return fmt.Errorf("failed to send email: %w", err)
	}

	zap.L().Info("Email sent to upstream SMTP successfully")

	return nil
}

func sendWithStartTLS(host, port string, auth smtp.Auth, from string, to []string, msg []byte) error {
	addr := fmt.Sprintf("%s:%s", host, port)

	// Using a custom dialer so that I can control the timeout. Default smtp client uses context.Background which might timeout.
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to add recipient %s: %w", addr, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send DATA command: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}
