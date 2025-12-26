package services

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type TwilioService struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

func NewTwilioService() *TwilioService {
	return &TwilioService{
		AccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		FromNumber: os.Getenv("TWILIO_WHATSAPP_NUMBER"),
	}
}

func (s *TwilioService) SendMessage(to string, body string) error {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.AccountSID)

	data := url.Values{}
	data.Set("From", s.FromNumber)
	data.Set("To", to)
	data.Set("Body", body)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(s.AccountSID, s.AuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("failed to send message, status: %s", resp.Status)
}
