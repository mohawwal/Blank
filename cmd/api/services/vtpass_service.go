package services

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"os"
// )

// type VTPassService struct {
// 	APIKey     string
// 	PublicKey  string
// 	SecretKey  string
// 	SandboxURL string
// 	LiveURL    string
// }

// func NewVTPassService() *VTPassService {
// 	return &VTPassService{
// 		APIKey:     os.Getenv("VT_PASS_API_KEY"),
// 		PublicKey:  os.Getenv("VT_PASS_PUBLIC_KEY"),
// 		SecretKey:  os.Getenv("VT_PASS_SECRET_KEY"),
// 		SandboxURL: os.Getenv("VT_PASS_SANDBOX_API_URL"),
// 		LiveURL:    os.Getenv("VT_PASS_LIVE_API_URL"),
// 	}
// }

// // BuyAirtimeResponse represents VTPass airtime purchase response
// type BuyAirtimeResponse struct {
// 	Code    string `json:"code"`
// 	Message string `json:"response_description"`
// 	Data    struct {
// 		TransactionID string `json:"transactionId"`
// 		Status        string `json:"status"`
// 		ProductName   string `json:"product_name"`
// 		Amount        string `json:"amount"`
// 		PhoneNumber   string `json:"phone"`
// 	} `json:"content"`
// }

// // BuyAirtime purchases airtime via VTPass
// func (s *VTPassService) BuyAirtime(network string, amount int, phoneNumber string, requestID string) (*BuyAirtimeResponse, error) {
// 	url := s.SandboxURL + "pay"

// 	// Map network names to VTPass service IDs
// 	serviceID := map[string]string{
// 		"mtn":     "mtn",
// 		"glo":     "glo",
// 		"airtel":  "airtel",
// 		"9mobile": "etisalat",
// 	}[network]

// 	if serviceID == "" {
// 		return nil, fmt.Errorf("invalid network: %s", network)
// 	}

// 	payload := map[string]interface{}{
// 		"request_id":   requestID,
// 		"serviceID":    serviceID,
// 		"amount":       amount,
// 		"phone":        phoneNumber,
// 	}

// 	jsonPayload, err := json.Marshal(payload)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("api-key", s.APIKey)
// 	req.Header.Set("public-key", s.PublicKey)

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var result BuyAirtimeResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return nil, err
// 	}

// 	if result.Code != "000" {
// 		return nil, fmt.Errorf("VTPass error: %s", result.Message)
// 	}

// 	return &result, nil
// }

// // BuyDataResponse represents VTPass data purchase response
// type BuyDataResponse struct {
// 	Code    string `json:"code"`
// 	Message string `json:"response_description"`
// 	Data    struct {
// 		TransactionID string `json:"transactionId"`
// 		Status        string `json:"status"`
// 		ProductName   string `json:"product_name"`
// 		Amount        string `json:"amount"`
// 		PhoneNumber   string `json:"phone"`
// 	} `json:"content"`
// }

// // BuyData purchases data bundle via VTPass
// func (s *VTPassService) BuyData(network string, variationCode string, phoneNumber string, requestID string) (*BuyDataResponse, error) {
// 	url := s.SandboxURL + "pay"

// 	// Map network names to VTPass service IDs
// 	serviceID := map[string]string{
// 		"mtn":     "mtn-data",
// 		"glo":     "glo-data",
// 		"airtel":  "airtel-data",
// 		"9mobile": "etisalat-data",
// 	}[network]

// 	if serviceID == "" {
// 		return nil, fmt.Errorf("invalid network: %s", network)
// 	}

// 	payload := map[string]interface{}{
// 		"request_id":     requestID,
// 		"serviceID":      serviceID,
// 		"billersCode":    phoneNumber,
// 		"variation_code": variationCode,
// 		"phone":          phoneNumber,
// 	}

// 	jsonPayload, err := json.Marshal(payload)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("api-key", s.APIKey)
// 	req.Header.Set("public-key", s.PublicKey)

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var result BuyDataResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return nil, err
// 	}

// 	if result.Code != "000" {
// 		return nil, fmt.Errorf("VTPass error: %s", result.Message)
// 	}

// 	return &result, nil
// }

// // GetDataPlans returns available data plans for a network
// type DataPlan struct {
// 	VariationCode string `json:"variation_code"`
// 	Name          string `json:"name"`
// 	Amount        int    `json:"variation_amount"`
// 	FixedPrice    string `json:"fixedPrice"`
// }

// type GetDataPlansResponse struct {
// 	Code    string     `json:"code"`
// 	Message string     `json:"response_description"`
// 	Data    []DataPlan `json:"content"`
// }

// func (s *VTPassService) GetDataPlans(network string) (*GetDataPlansResponse, error) {
// 	serviceID := map[string]string{
// 		"mtn":     "mtn-data",
// 		"glo":     "glo-data",
// 		"airtel":  "airtel-data",
// 		"9mobile": "etisalat-data",
// 	}[network]

// 	if serviceID == "" {
// 		return nil, fmt.Errorf("invalid network: %s", network)
// 	}

// 	url := fmt.Sprintf("%sservice-variations?serviceID=%s", s.SandboxURL, serviceID)

// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("api-key", s.APIKey)
// 	req.Header.Set("public-key", s.PublicKey)

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var result GetDataPlansResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return nil, err
// 	}

// 	return &result, nil
// }
