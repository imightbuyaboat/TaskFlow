package email

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strconv"

	"github.com/imightbuyaboat/TaskFlow/pkg/task"
	"gopkg.in/gomail.v2"
)

type MailDialer struct {
	dialer       *gomail.Dialer
	from         string
	baseFilePath string
}

func NewMailDialer() (*MailDialer, error) {
	from := os.Getenv("MAIL_USERNAME")
	_, err := mail.ParseAddress(from)
	if err != nil {
		return nil, fmt.Errorf("incorrect mail address: %v", err)
	}

	baseFilePath := os.Getenv("BASE_FILE_PATH")

	host := os.Getenv("MAIL_HOST")
	portStr := os.Getenv("MAIL_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("incorrect format of port: %v", err)
	}

	username := os.Getenv("MAIL_USERNAME")
	password := os.Getenv("MAIL_PASSWORD")

	if host == "" || username == "" || password == "" {
		return nil, fmt.Errorf("env vars are empty")
	}

	d := gomail.NewDialer(host, port, username, password)

	return &MailDialer{
		dialer:       d,
		from:         from,
		baseFilePath: baseFilePath,
	}, nil
}

func (md *MailDialer) ExecuteTask(rawPayload interface{}) error {
	data, err := json.Marshal(rawPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal rawPayload: %v", err)
	}

	var payload task.SendEmailPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload to SendEmailPayload: %v", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", md.from)
	m.SetHeader("To", payload.To)

	if payload.Subject != "" {
		m.SetHeader("Subject", payload.Subject)
	}

	if payload.Body != "" {
		m.SetBody("text/html", payload.Body)
	}

	if payload.AttachedFiles != nil {
		for _, fileName := range payload.AttachedFiles {
			filePath := filepath.Join(md.baseFilePath, fileName)
			m.Attach(filePath)
		}
	}

	if err := md.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send mail: %v", err)
	}

	return nil
}
