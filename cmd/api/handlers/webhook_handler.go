package handlers

import (
	"fmt"
	"log"
	"net/http"
	// "strconv"
	"strings"
	"whatsapp-bot-app/internal/models"

	"github.com/labstack/echo/v4"
)

func (h *Handler) WhatsAppWebhook(c echo.Context) error {
	from := c.FormValue("From")
	body := c.FormValue("Body")

	if from == "" || body == "" {
		return c.String(http.StatusBadRequest, "Missing From or Body parameter")
	}

	userPhone := strings.Replace(from, "whatsapp:", "", 1)
	cleanBody := strings.TrimSpace(body)

	log.Printf("Received message from %s: %s", userPhone, cleanBody)

	var user models.UserModel
	result := h.DB.Where("phone_number = ?", userPhone).First(&user)

	// 1. NEW USER - Create user and ask for name
	if result.Error != nil {
		newUser := models.UserModel{
			PhoneNumber:    userPhone,
			Status:         "new",
			OnboardingStep: "awaiting_name",
		}
		if err := h.DB.Create(&newUser).Error; err != nil {
			log.Printf("Error creating user: %v", err)
			return c.String(http.StatusInternalServerError, "Error creating user")
		}

		response, err := h.ClaudeService.ProcessUserMessage(cleanBody, "", true)
		if err != nil {
			log.Printf("Claude API error: %v", err)
			response = "Welcome! To get started, please tell me your full name."
		}

		if err := h.TwilioService.SendMessage(from, response); err != nil {
			log.Printf("Error sending message: %v", err)
		}

		return c.String(http.StatusOK, "OK")
	}

	// 2. AWAITING NAME - Collect name and setup Paystack card authorization
	if user.Status == "new" && user.OnboardingStep == "awaiting_name" {
		if len(cleanBody) < 3 {
			h.TwilioService.SendMessage(from, "Please enter a valid full name (at least 3 characters).")
			return c.String(http.StatusOK, "OK")
		}

		// Generate email from phone number
		email := fmt.Sprintf("%s@whatsappbot.com", userPhone)

		// Split name for First/Last
		names := strings.Fields(cleanBody)
		firstName := names[0]
		lastName := ""
		if len(names) > 1 {
			lastName = strings.Join(names[1:], " ")
		} else {
			lastName = "User"
		}

		h.TwilioService.SendMessage(from, "Creating your account... Please wait.")

		// Create Paystack Customer
		custResp, err := h.PaystackService.CreateCustomer(email, firstName, lastName, userPhone)
		if err != nil {
			log.Printf("Create Customer Error: %v", err)
			h.TwilioService.SendMessage(from, "Sorry, I couldn't create your account right now. Please try again later.")
			return c.String(http.StatusOK, "OK")
		}

		// Initialize transaction for card authorization (₦100 = 10000 kobo)
		callbackURL := "https://unbeaming-insensately-konner.ngrok-free.dev/webhook/paystack"
		txnResp, err := h.PaystackService.InitializeTransaction(email, 100, callbackURL)
		if err != nil {
			log.Printf("Initialize Transaction Error: %v", err)
			h.TwilioService.SendMessage(from, "I created your profile but failed to generate payment link. Please try again later.")
			return c.String(http.StatusOK, "OK")
		}

		// Save user details (NOT active yet!)
		user.FullName = cleanBody
		user.Email = email
		user.PaystackCustomerCode = custResp.Data.CustomerCode
		user.OnboardingStep = "awaiting_payment"
		user.Status = "new" // Still NEW, not active!

		if err := h.DB.Save(&user).Error; err != nil {
			log.Printf("Error updating user: %v", err)
			return c.String(http.StatusInternalServerError, "Error updating user")
		}

		// Send payment link
		msg := fmt.Sprintf("Account Created!\n\n"+
			"👋 Hi %s!\n\n"+
			"To enable instant purchases, please link your card securely.\n\n"+
			"💳 One-time verification: ₦100\n"+
			"🔒 Secure Paystack checkout\n\n"+
			"Click here to authorize your card:\n%s\n\n"+
			"After payment, you can instantly buy airtime & data!",
			firstName, txnResp.Data.AuthorizationURL)

		if err := h.TwilioService.SendMessage(from, msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}

		return c.String(http.StatusOK, "OK")
	}

	// 3. AWAITING PAYMENT - User sent message while waiting for payment
	if user.Status == "new" && user.OnboardingStep == "awaiting_payment" {
		msg := "⏳ Still waiting for your card authorization.\n\n" +
			"Please complete the ₦100 payment to link your card.\n\n" +
			"Once done, you can instantly buy airtime & data!"
		h.TwilioService.SendMessage(from, msg)
		return c.String(http.StatusOK, "OK")
	}

	// 4. ACTIVE USER - Process commands
	if user.Status == "active" {
		// Check for confirmation keywords
		// if user.PendingTransactionID != nil && (cleanBody == "YES" || cleanBody == "yes" || cleanBody == "Yes" || cleanBody == "confirm" || cleanBody == "CONFIRM") {
		// 	// User confirmed the purchase
		// 	var txn models.TransactionModel
		// 	if err := h.DB.First(&txn, *user.PendingTransactionID).Error; err == nil {
		// 		if txn.Status == "pending" {
		// 			h.TwilioService.SendMessage(from, "⏳ Processing your purchase... Please wait.")

		// 			// Clear pending transaction
		// 			user.PendingTransactionID = nil
		// 			h.DB.Save(&user)

		// 			// Process in background
		// 			go h.processPurchase(&txn, &user)
		// 			return c.String(http.StatusOK, "OK")
		// 		}
		// 	}
		// }

		// Use Claude AI to parse the request
		response, err := h.ClaudeService.ProcessUserMessage(cleanBody, user.FullName, false)
		if err != nil {
			log.Printf("Claude API error: %v", err)
			response = "Sorry, I'm having trouble processing your request right now. Please try again."
			h.TwilioService.SendMessage(from, response)
			return c.String(http.StatusOK, "OK")
		}

		// Check if response contains purchase intent (simplified parsing)
		// In production, Claude should return structured data
		// if strings.Contains(strings.ToLower(cleanBody), "buy") || strings.Contains(strings.ToLower(cleanBody), "send") || strings.Contains(strings.ToLower(cleanBody), "recharge") {
		// 	// Parse purchase details (simplified - in production use Claude's structured output)
		// 	network, amount, phoneNumber := h.parsePurchaseRequest(cleanBody, userPhone)

		// 	if network != "" && amount > 0 {
		// 		// Create pending transaction
		// 		txn := models.TransactionModel{
		// 			UserID:      user.ID,
		// 			Type:        "airtime",
		// 			Network:     network,
		// 			Amount:      amount,
		// 			PhoneNumber: phoneNumber,
		// 			Status:      "pending",
		// 		}
		// 		h.DB.Create(&txn)

		// 		// Update user's pending transaction
		// 		user.PendingTransactionID = &txn.ID
		// 		h.DB.Save(&user)

		// 		// Send confirmation request
		// 		confirmationLink := fmt.Sprintf("https://unbeaming-insensately-konner.ngrok-free.dev/purchase/confirm/%d", txn.ID)
		// 		msg := fmt.Sprintf("📶 %s Airtime\n"+
		// 			"💰 ₦%d\n"+
		// 			"📱 %s\n\n"+
		// 			"💳 Card: **** %s\n\n"+
		// 			"Click to confirm:\n%s\n\n"+
		// 			"Or reply *YES* to confirm",
		// 			strings.ToUpper(network), amount, phoneNumber, user.CardLast4, confirmationLink)

		// 		h.TwilioService.SendMessage(from, msg)
		// 		return c.String(http.StatusOK, "OK")
		// 	}
		// }

		// Send Claude's response
		if err := h.TwilioService.SendMessage(from, response); err != nil {
			log.Printf("Error sending message: %v", err)
		}

		return c.String(http.StatusOK, "OK")
	}

	// 5. FALLBACK - Other statuses
	h.TwilioService.SendMessage(from, "Your account is pending activation. Please contact support.")
	return c.String(http.StatusOK, "OK")
}

// parsePurchaseRequest is a simplified parser for purchase requests
// In production, you should use Claude AI to return structured JSON
// func (h *Handler) parsePurchaseRequest(message string, defaultPhone string) (network string, amount int, phoneNumber string) {
// 	message = strings.ToLower(message)

// 	// Parse network
// 	if strings.Contains(message, "mtn") {
// 		network = "mtn"
// 	} else if strings.Contains(message, "glo") {
// 		network = "glo"
// 	} else if strings.Contains(message, "airtel") {
// 		network = "airtel"
// 	} else if strings.Contains(message, "9mobile") || strings.Contains(message, "etisalat") {
// 		network = "9mobile"
// 	}

// 	// Parse amount (look for numbers between 50 and 10000)
// 	words := strings.Fields(message)
// 	for _, word := range words {
// 		// Remove currency symbols
// 		word = strings.ReplaceAll(word, "₦", "")
// 		word = strings.ReplaceAll(word, "ngn", "")
// 		if num, err := strconv.Atoi(word); err == nil {
// 			if num >= 50 && num <= 10000 {
// 				amount = num
// 				break
// 			}
// 		}
// 	}

// 	// Parse phone number (look for 11-digit numbers starting with 0)
// 	for _, word := range words {
// 		if len(word) == 11 && word[0] == '0' {
// 			if _, err := strconv.Atoi(word); err == nil {
// 				phoneNumber = word
// 				break
// 			}
// 		}
// 	}

// 	// Default to user's own phone if not specified
// 	if phoneNumber == "" {
// 		phoneNumber = defaultPhone
// 	}

// 	return network, amount, phoneNumber
// }
