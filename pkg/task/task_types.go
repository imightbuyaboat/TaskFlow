package task

import (
	"encoding/json"
	"fmt"
	"net/mail"
)

var validatePayloadsFunctions = map[string]func(map[string]interface{}) error{
	"send_email": validateSendEmailPayload,
}

type SendEmailPayload struct {
	To            string   `json:"to"`
	Subject       string   `json:"subject"`
	Body          string   `json:"body"`
	AttachedFiles []string `json:"attached_files"`
}

func ValidateType(typeOfTask string) bool {
	_, ok := validatePayloadsFunctions[typeOfTask]
	return ok
}

func ValidatePayload(typeOfTask string, payload map[string]interface{}) error {
	return validatePayloadsFunctions[typeOfTask](payload)
}

func validateSendEmailPayload(payload map[string]interface{}) error {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload into json: %v", err)
	}

	var semp SendEmailPayload
	err = json.Unmarshal(jsonBytes, &semp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json into payload: %v", err)
	}

	if semp.To == "" {
		return fmt.Errorf("field 'to' cant be empty")
	}

	_, err = mail.ParseAddress(semp.To)
	if err != nil {
		return fmt.Errorf("incorrect address in filed 'To'")
	}

	if semp.Subject == "" && semp.Body == "" && semp.AttachedFiles == nil {
		return fmt.Errorf("fields 'subject', 'body', 'attached_files' cant be empty at the same time")
	}

	return nil
}
