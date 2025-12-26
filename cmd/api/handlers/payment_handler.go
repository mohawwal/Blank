package handlers

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"whatsapp-bot-app/internal/models"

	"github.com/labstack/echo/v4"
)

type PaystackWebhookEvent struct {
	Event string `json:"event"`
	Data  struct {
		Status    string `json:"status"`
		Reference string `json:"reference"`
		Amount    int    `json:"amount"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			Bin               string `json:"bin"`
			Last4             string `json:"last4"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			CardType          string `json:"card_type"`
			Bank              string `json:"bank"`
		} `json:"authorization"`
		Customer struct {
			Email        string `json:"email"`
			CustomerCode string `json:"customer_code"`
		} `json:"customer"`
	} `json:"data"`
}

func (h *Handler) PaystackWebhook(c echo.Context) error {
	// 1. Verify Signature
	secret := os.Getenv("PAYSTACK_SECRET_KEY")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Read body failed: %v", err)
		return c.String(http.StatusBadRequest, "Read body failed")
	}

	sig := c.Request().Header.Get("x-paystack-signature")
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if sig != expectedSig {
		log.Printf("Invalid signature. Expected: %s, Got: %s", expectedSig, sig)
		return c.String(http.StatusUnauthorized, "Invalid signature")
	}

	// 2. Parse Event
	var event PaystackWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Parse error: %v", err)
		return c.String(http.StatusBadRequest, "Parse error")
	}

	log.Printf("Received Paystack event: %s", event.Event)

	// 3. Handle charge.success event
	if event.Event == "charge.success" && event.Data.Status == "success" {
		customerCode := event.Data.Customer.CustomerCode

		var user models.UserModel
		if err := h.DB.Where("paystack_customer_code = ?", customerCode).First(&user).Error; err != nil {
			log.Printf("User not found for customer code: %s", customerCode)
			return c.String(http.StatusOK, "User not found")
		}

		log.Printf("Payment received for %s (%s). Amount: ₦%.2f", user.FullName, user.PhoneNumber, float64(event.Data.Amount)/100)

		// Save card authorization details
		user.AuthorizationCode = event.Data.Authorization.AuthorizationCode
		user.CardLast4 = event.Data.Authorization.Last4
		user.CardType = event.Data.Authorization.CardType
		user.CardBank = event.Data.Authorization.Bank
		user.CardExpMonth = event.Data.Authorization.ExpMonth
		user.CardExpYear = event.Data.Authorization.ExpYear
		user.Status = "active"
		user.OnboardingStep = "completed"

		if err := h.DB.Save(&user).Error; err != nil {
			log.Printf("Error saving user: %v", err)
			return c.String(http.StatusInternalServerError, "Error saving user")
		}

		// Notify User
		msg := fmt.Sprintf("✅ Card Linked Successfully!\n\n"+
			"💳 Card: **** **** **** %s\n"+
			"🏦 Bank: %s\n"+
			"✨ Your account is now active!\n\n"+
			"You can now buy airtime & data instantly.\n\n"+
			"Try:\n"+
			"- \"Buy Glo ₦200\"\n"+
			"- \"Send MTN data 2GB to 0813xxxx\"",
			user.CardLast4, user.CardBank)

		if err := h.TwilioService.SendMessage("whatsapp:"+user.PhoneNumber, msg); err != nil {
			log.Printf("Error sending activation message: %v", err)
		}
	}

	// Paystack expects 200 OK
	return c.String(http.StatusOK, "Event received")
}
