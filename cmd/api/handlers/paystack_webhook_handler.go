package handlers

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"whatsapp-bot-app/internal/models"

	"github.com/labstack/echo/v4"
)

// PaystackWebhookEvent represents the webhook payload from Paystack
type PaystackWebhookEvent struct {
	Event string `json:"event"`
	Data  struct {
		Reference    string `json:"reference"`
		Status       string `json:"status"`
		Amount       int    `json:"amount"`
		Customer     struct {
			CustomerCode string `json:"customer_code"`
			Email        string `json:"email"`
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
			Reusable          bool   `json:"reusable"`
		} `json:"authorization"`
		Metadata struct {
			WhatsApp string `json:"whatsapp"`
			Purpose  string `json:"purpose"`
		} `json:"metadata"`
	} `json:"data"`
}

// PaystackWebhook handles incoming webhooks from Paystack
func (h *Handler) PaystackWebhook(c echo.Context) error {
	// Verify webhook signature
	signature := c.Request().Header.Get("x-paystack-signature")
	if signature == "" {
		log.Println("Missing Paystack signature")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing signature"})
	}

	// Read the request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to read body"})
	}

	// Verify the signature
	secretKey := os.Getenv("PAYSTACK_SECRET_KEY")
	if !verifyPaystackSignature(body, signature, secretKey) {
		log.Println("Invalid Paystack signature")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid signature"})
	}

	// Parse the webhook event
	var event PaystackWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error parsing webhook: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	log.Printf("Received Paystack webhook: %s for reference: %s", event.Event, event.Data.Reference)

	// Handle the event
	switch event.Event {
	case "charge.success":
		return h.handleChargeSuccess(c, &event)
	default:
		log.Printf("Unhandled Paystack event: %s", event.Event)
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "received"})
}

// handleChargeSuccess processes successful payment webhooks
func (h *Handler) handleChargeSuccess(c echo.Context, event *PaystackWebhookEvent) error {
	// Only process wallet setup transactions
	if event.Data.Metadata.Purpose != "wallet_setup" {
		log.Printf("Ignoring non-wallet-setup transaction: %s", event.Data.Reference)
		return c.JSON(http.StatusOK, map[string]string{"status": "ignored"})
	}

	phoneNumber := event.Data.Metadata.WhatsApp
	if phoneNumber == "" {
		log.Println("Missing WhatsApp number in metadata")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing phone number"})
	}

	// Find user by phone number
	var user models.UserModel
	if err := h.DB.Where("phone_number = ?", phoneNumber).First(&user).Error; err != nil {
		log.Printf("User not found for phone: %s, error: %v", phoneNumber, err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	// Update user with Paystack authorization details
	user.PaystackCustomerCode = event.Data.Customer.CustomerCode
	user.AuthorizationCode = event.Data.Authorization.AuthorizationCode
	user.CardLast4 = event.Data.Authorization.Last4
	user.CardType = event.Data.Authorization.CardType
	user.CardBank = event.Data.Authorization.Bank
	user.CardExpMonth = event.Data.Authorization.ExpMonth
	user.CardExpYear = event.Data.Authorization.ExpYear
	user.Status = "active" // Update user status to active
	user.OnboardingStep = "completed"
	user.OnboardingTxnStatus = "success" // Mark transaction as successful

	if err := h.DB.Save(&user).Error; err != nil {
		log.Printf("Failed to update user: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save user"})
	}

	log.Printf("Successfully saved payment authorization for user: %s", phoneNumber)

	// Send confirmation message to user via WhatsApp using MessageProcessor
	confirmationMsg := "Payment successful! Your wallet has been set up. You can now make purchases using your saved card."
	if err := h.MessageProcessor.MessagingService.SendMessage("whatsapp:"+phoneNumber, confirmationMsg); err != nil {
		log.Printf("Failed to send confirmation message: %v", err)
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// verifyPaystackSignature validates the webhook signature
func verifyPaystackSignature(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
