package notifier

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/config"
	"go.uber.org/zap"
)

type Notifier struct {
	logger         *zap.Logger
	smtpHost       string
	smtpPort       string
	emailSender    string
	emailPassword  string
	emailReceivers []string // Changed to slice
}

func NewNotifier(logger *zap.Logger, cfg *config.Config) *Notifier {
	// Parse comma-separated receivers
	receivers := strings.Split(cfg.EmailReceivers, ",")
	for i := range receivers {
		receivers[i] = strings.TrimSpace(receivers[i])
	}

	return &Notifier{
		logger:         logger,
		smtpHost:       cfg.SMTPHost,
		smtpPort:       cfg.SMTPPort,
		emailSender:    cfg.EmailSender,
		emailPassword:  cfg.EmailPassword,
		emailReceivers: receivers,
	}
}

func (n *Notifier) Send(subject, body string, attachmentPaths []string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Build headers
	headers := textproto.MIMEHeader{}
	headers.Set("From", n.emailSender)
	headers.Set("To", strings.Join(n.emailReceivers, ", "))
	headers.Set("Subject", subject)
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", "multipart/mixed; boundary="+writer.Boundary())

	for k, v := range headers {
		fmt.Fprintf(&buf, "%s: %s\r\n", k, v[0])
	}
	buf.WriteString("\r\n")

	// Body
	bodyPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type": {"text/plain; charset=utf-8"},
	})
	if err != nil {
		return fmt.Errorf("failed to create body part: %w", err)
	}
	
	if _, err := bodyPart.Write([]byte(body)); err != nil {
		return fmt.Errorf("failed to write body: %w", err)
	}

	// Attachments
	for _, path := range attachmentPaths {
		if err := attachFile(writer, path); err != nil {
			return fmt.Errorf("failed to attach file %s: %w", path, err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Use configured SMTP settings
	auth := smtp.PlainAuth("", n.emailSender, n.emailPassword, n.smtpHost)
	addr := fmt.Sprintf("%s:%s", n.smtpHost, n.smtpPort)
	
	if err := smtp.SendMail(addr, auth, n.emailSender, n.emailReceivers, buf.Bytes()); err != nil {
		n.logger.Error("Failed to send email", zap.Error(err))
		return fmt.Errorf("failed to send email: %w", err)
	}

	n.logger.Info("Email sent successfully")
	return nil
}

func attachFile(writer *multipart.Writer, filePath string) error {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Type", "text/csv")
	partHeader.Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(filePath)))
	partHeader.Set("Content-Transfer-Encoding", "base64")

	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %w", err)
	}

	encoder := base64.NewEncoder(base64.StdEncoding, part)
	if _, err := encoder.Write(fileData); err != nil {
		encoder.Close()
		return fmt.Errorf("failed to write encoded data: %w", err)
	}
	
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close encoder: %w", err)
	}

	return nil
}