package task

import (
	"encoding/json"
	"fmt"
	"net/mail"
)

var validatePayloadsFunctions = map[string]func(map[string]interface{}) error{
	"send_email":     validateSendEmailPayload,
	"process_image":  validateImageProcessingPayload,
	"download_files": validateFileDownloadingPayload,
}

type SendEmailPayload struct {
	To            string   `json:"to"`
	Subject       string   `json:"subject"`
	Body          string   `json:"body"`
	AttachedFiles []string `json:"attached_files"`
}

type ImageProcessingPayload struct {
	Path       string  `json:"path"`
	Blur       float64 `json:"blur"`
	Sharpen    float64 `json:"sharpen"`
	Gamma      float64 `json:"gamma"`
	Contrast   float64 `json:"contrast"`
	Brightness float64 `json:"brightness"`
	Saturation float64 `json:"saturation"`
	Grayscale  bool    `json:"grayscale"`
	Invert     bool    `json:"invert"`
}

type FileDownloadingPayload struct {
	URLs []string `json:"urls"`
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

func validateImageProcessingPayload(payload map[string]interface{}) error {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload into json: %v", err)
	}

	var ipp ImageProcessingPayload
	err = json.Unmarshal(jsonBytes, &ipp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json into payload: %v", err)
	}

	if ipp.Path == "" {
		return fmt.Errorf("missing image path")
	}

	if ipp.Blur < 0 || ipp.Sharpen < 0 || ipp.Gamma < 0 {
		return fmt.Errorf("blur, sharpen, gamma must be positive")
	}

	if ipp.Contrast < -100 || ipp.Contrast > 100 ||
		ipp.Brightness < -100 || ipp.Brightness > 100 ||
		ipp.Saturation < -100 || ipp.Saturation > 100 {
		return fmt.Errorf("contrast, brightness, saturation must be in the range (-100, 100)")
	}

	return nil
}

func validateFileDownloadingPayload(payload map[string]interface{}) error {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload into json: %v", err)
	}

	var fdp FileDownloadingPayload
	err = json.Unmarshal(jsonBytes, &fdp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json into payload: %v", err)
	}

	if fdp.URLs == nil {
		return fmt.Errorf("missing URLs")
	}

	if len(fdp.URLs) == 0 {
		return fmt.Errorf("missing URLs")
	}

	return nil
}
