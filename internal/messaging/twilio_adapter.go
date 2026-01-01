package messaging

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// TwilioAdapter implements MessagingService for Twilio WhatsApp
type TwilioAdapter struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

// NewTwilioAdapter creates a new Twilio messaging adapter
func NewTwilioAdapter() *TwilioAdapter {
	return &TwilioAdapter{
		AccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		FromNumber: os.Getenv("TWILIO_WHATSAPP_NUMBER"),
	}
}

// SendMessage sends a regular WhatsApp message via Twilio
func (t *TwilioAdapter) SendMessage(to string, body string) error {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.AccountSID)

	data := url.Values{}
	data.Set("From", t.FromNumber)
	data.Set("To", to)
	data.Set("Body", body)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(t.AccountSID, t.AuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Read error response body for debugging
	bodyBytes := make([]byte, 1024)
	n, _ := resp.Body.Read(bodyBytes)
	errorBody := string(bodyBytes[:n])

	return fmt.Errorf("failed to send message, status: %s, body: %s", resp.Status, errorBody)
}

// SendTemplateMessage sends a Twilio Content Template message
func (t *TwilioAdapter) SendTemplateMessage(to string, templateData TemplateData) error {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.AccountSID)

	data := url.Values{}
	data.Set("From", t.FromNumber)
	data.Set("To", to)
	data.Set("ContentSid", templateData.TemplateName)

	// Add ContentVariables if provided
	if len(templateData.Variables) > 0 {
		varsJSON, err := json.Marshal(templateData.Variables)
		if err != nil {
			return fmt.Errorf("failed to marshal variables: %w", err)
		}
		data.Set("ContentVariables", string(varsJSON))
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(t.AccountSID, t.AuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes := make([]byte, 2048)
	n, _ := resp.Body.Read(bodyBytes)
	responseBody := string(bodyBytes[:n])

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("Twilio template response (SUCCESS): %s\n", responseBody)
		return nil
	}

	return fmt.Errorf("failed to send template message, status: %s, body: %s", resp.Status, responseBody)
}
