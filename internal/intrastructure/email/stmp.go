package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/Vasenti/stori_challenge/internal/config"
)

type SMTPSender struct {
	config *config.Config
}

func NewSMTPSender(cfg *config.Config) *SMTPSender {
	return &SMTPSender{config: cfg}
}

func (s *SMTPSender) Send(to string, subject string, htmlBody string) error {

	addr := net.JoinHostPort(s.config.SMTPHost, strconv.Itoa(s.config.SMTPPort))
	toList := []string{to}

	headers := map[string]string{
		"From":         s.config.SMTPFrom,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}
	var sb strings.Builder
	for k, v := range headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	sb.WriteString("\r\n")
	sb.WriteString(htmlBody)

	msg := []byte(sb.String())

	// MailHog: sin auth
	if s.config.SMTPUsername == "" && s.config.SMTPPassword == "" {
		return smtp.SendMail(addr, nil, s.config.SMTPFrom, toList, msg)
	}

	// Con AUTH (+ STARTTLS si el servidor lo soporta)
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
	tlsconfig := &tls.Config{ServerName: s.config.SMTPHost, InsecureSkipVerify: true}

	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(tlsconfig); err != nil {
			return err
		}
	}
	if ok, _ := c.Extension("AUTH"); ok {
		if err := c.Auth(auth); err != nil {
			return err
		}
	}

	if err := c.Mail(s.config.SMTPFrom); err != nil {
		return err
	}
	for _, rcpt := range toList {
		if err := c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	return w.Close()
	return nil
}
