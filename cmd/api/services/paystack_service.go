package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type PaystackService struct {
	SecretKey string
}

func NewPaystackService() *PaystackService {
	return &PaystackService{
		SecretKey: os.Getenv("PAYSTACK_SECRET_KEY"),
	}
}

type CreateCustomerResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		CustomerCode string `json:"customer_code"`
		Id           int    `json:"id"`
	} `json:"data"`
}

type CreateDVANResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Bank struct {
			Name string `json:"name"`
			Id   int    `json:"id"`
			Slug string `json:"slug"`
		} `json:"bank"`
		AccountNumber string `json:"account_number"`
		AccountName   string `json:"account_name"`
	} `json:"data"`
}

func (s *PaystackService) CreateCustomer(email string, firstName string, lastName string, phone string) (*CreateCustomerResponse, error) {
	url := "https://api.paystack.co/customer"

	payload := map[string]interface{}{
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"phone":      phone,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CreateCustomerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, fmt.Errorf("create customer failed: %s", result.Message)
	}

	return &result, nil
}

func (s *PaystackService) CreateDedicatedAccount(customerCode string) (*CreateDVANResponse, error) {
	url := "https://api.paystack.co/dedicated_account"

	payload := map[string]interface{}{
		"customer":       customerCode,
		"preferred_bank": "wema-bank", // Usually recommended for DVAN
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CreateDVANResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, fmt.Errorf("create dedicated account failed: %s", result.Message)
	}

	return &result, nil
}

// InitializeTransaction creates a payment link for card authorization
type InitializeTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

func (s *PaystackService) InitializeTransaction(email string, amount int, callbackURL string) (*InitializeTransactionResponse, error) {
	url := "https://api.paystack.co/transaction/initialize"

	payload := map[string]interface{}{
		"email":        email,
		"amount":       amount * 100, // Convert to kobo
		"callback_url": callbackURL,
		"metadata": map[string]string{
			"purpose": "card_authorization",
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result InitializeTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, fmt.Errorf("initialize transaction failed: %s", result.Message)
	}

	return &result, nil
}

// VerifyTransaction confirms a payment was successful
type VerifyTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status        string `json:"status"`
		Reference     string `json:"reference"`
		Amount        int    `json:"amount"`
		Customer      struct {
			CustomerCode string `json:"customer_code"`
		} `json:"customer"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			Bin               string `json:"bin"`
			Last4             string `json:"last4"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			CardType          string `json:"card_type"`
			Bank              string `json:"bank"`
			Brand             string `json:"brand"`
		} `json:"authorization"`
	} `json:"data"`
}

func (s *PaystackService) VerifyTransaction(reference string) (*VerifyTransactionResponse, error) {
	url := fmt.Sprintf("https://api.paystack.co/transaction/verify/%s", reference)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.SecretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result VerifyTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, fmt.Errorf("verify transaction failed: %s", result.Message)
	}

	return &result, nil
}

// ChargeAuthorization charges a previously authorized card
type ChargeAuthorizationResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status    string `json:"status"`
		Reference string `json:"reference"`
		Amount    int    `json:"amount"`
	} `json:"data"`
}

func (s *PaystackService) ChargeAuthorization(authorizationCode string, email string, amount int) (*ChargeAuthorizationResponse, error) {
	url := "https://api.paystack.co/transaction/charge_authorization"

	payload := map[string]interface{}{
		"authorization_code": authorizationCode,
		"email":              email,
		"amount":             amount * 100, // Convert to kobo
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ChargeAuthorizationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, fmt.Errorf("charge authorization failed: %s", result.Message)
	}

	return &result, nil
}
