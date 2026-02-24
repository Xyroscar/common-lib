package sender

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/xyroscar/common-lib/pkg/config"
	"go.uber.org/zap"
)

type Email struct {
	From    string
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

func Send(email *Email) error {
	c, err := config.GetSmtpConfig()
	if err != nil {
		zap.L().Error("Error loading smtp config", zap.Error(err))
		return err
	}

	from := email.From
	if from == "" {
		from = c.Upstream.From
	}

	zap.L().Debug("Sending email to upstream SMTP",
		zap.String("from", from),
		zap.Strings("to", email.To),
		zap.String("subject", email.Subject),
		zap.String("upstream_host", c.Upstream.Host),
		zap.String("upstream_port", c.Upstream.Port),
	)

	mime := "MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n"
	if email.IsHTML {
		mime = "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	}

	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", "))
	message += fmt.Sprintf("Subject: %s\r\n", email.Subject)
	message += mime
	message += email.Body

	auth := smtp.PlainAuth("", c.Upstream.Username, c.Upstream.Password, c.Upstream.Host)

	err = sendWithStartTLS(c.Upstream.Host, c.Upstream.Port, auth, from, email.To, []byte(message))
	if err != nil {
		zap.L().Error("Failed to send email to upstream SMTP",
			zap.Error(err),
			zap.String("from", from),
			zap.Strings("to", email.To),
			zap.String("upstream_host", c.Upstream.Host),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	zap.L().Info("Email sent to upstream SMTP successfully",
		zap.String("from", from),
		zap.Strings("to", email.To),
		zap.String("subject", email.Subject),
	)

	return nil
}

func sendWithStartTLS(host, port string, auth smtp.Auth, from string, to []string, msg []byte) error {
	addr := fmt.Sprintf("%s:%s", host, port)
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		ServerName: host,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return err
	}

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
