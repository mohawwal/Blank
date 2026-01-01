package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type PaystackService struct {
	SecretKey string
	BaseURL   string
}

func NewPaystackService() *PaystackService {
	return &PaystackService{
		SecretKey: os.Getenv("PAYSTACK_SECRET_KEY"),
		BaseURL:   "https://api.paystack.co",
	}
}

// InitializeTransactionRequest represents the request payload for transaction initialization
type InitializeTransactionRequest struct {
	Email    string                 `json:"email"`
	Amount   int                    `json:"amount"` // Amount in kobo (₦1 = 100 kobo)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	CallbackURL string              `json:"callback_url,omitempty"`
}

// InitializeTransactionResponse represents Paystack's response
type InitializeTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

// InitializeTransaction creates a new transaction and returns the payment URL
func (s *PaystackService) InitializeTransaction(phoneNumber string, amount int) (*InitializeTransactionResponse, error) {
	// Create email from phone number (Paystack requires email)
	email := fmt.Sprintf("user_%s@blankaibot.com", phoneNumber)

	payload := InitializeTransactionRequest{
		Email:  email,
		Amount: amount, // Amount in kobo
		Metadata: map[string]interface{}{
			"whatsapp": phoneNumber,
			"purpose":  "wallet_setup",
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", s.BaseURL+"/transaction/initialize", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paystack error: %s, body: %s", resp.Status, string(body))
	}

	var response InitializeTransactionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Status {
		return nil, fmt.Errorf("paystack returned status false: %s", response.Message)
	}

	return &response, nil
}

// VerifyTransaction verifies a transaction by reference
func (s *PaystackService) VerifyTransaction(reference string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", s.BaseURL+"/transaction/verify/"+reference, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.SecretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}
