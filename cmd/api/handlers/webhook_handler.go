package handlers

import (
	// "fmt"
	"log"
	"net/http"
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
	ProfileName := c.FormValue("ProfileName")

	log.Printf("Received message from %s: %s", userPhone, cleanBody)

	// Process message in background to avoid Twilio timeout
	go h.processWhatsAppMessage(from, userPhone, ProfileName, cleanBody)

	// return c.String(http.StatusOK, "message received")
	return c.NoContent(http.StatusOK)
}

// processWhatsAppMessage handles the actual message processing asynchronously
func (h *Handler) processWhatsAppMessage(from, userPhone, ProfileName, cleanBody string) {
	log.Printf("Processing message - From: %s, Phone: %s, Name: %s", from, userPhone, ProfileName)

	var user models.UserModel
	result := h.DB.Where("phone_number = ?", userPhone).First(&user)

	// 1. NEW USER - Create user and send welcome message
	if result.Error != nil {
		log.Printf("New user detected: %s", userPhone)
		userName := ProfileName
		if userName == "" {
			userName = "there"
		}

		newUser := models.UserModel{
			PhoneNumber:    userPhone,
			UserName:       userName,
			Status:         "new",
			OnboardingStep: "awaiting_start",
		}
		if err := h.DB.Create(&newUser).Error; err != nil {
			log.Printf("ERROR creating user: %v", err)
			return
		}
		log.Printf("User created successfully: %s", userPhone)

		welcomeMsg, err := h.ClaudeService.ProcessUserMessage("Generate a welcome message", userName, true)
		if err != nil {
			log.Printf("API error for welcome message: %v", err)
			// Fallback message when Claude API fails
			welcomeMsg = "Hello 😁" + userName + "! \n\nWelcome to *Blank AI Bot*!\n\nI can help you with:\n• Airtime & Data (MTN, Glo, Airtel, 9mobile)\n• Electricity bills (PHCN)\n• TV subscriptions (DStv, GOtv, Startimes)\n• Internet bills\n• Water bills...\n\n"
		}

		log.Printf("Sending welcome message to: %s", from)
		if err := h.TwilioService.SendMessage(from, welcomeMsg); err != nil {
			log.Printf("ERROR sending welcome message: %v", err)
			return
		}
		log.Printf("Welcome message sent successfully to: %s", from)

		return
	}

	// 2. EXISTING USER - Process their message with AI
	log.Printf("Processing message for existing user: %s", user.UserName)

	response, err := h.ClaudeService.ProcessUserMessage(cleanBody, user.UserName, false)
	if err != nil {
		log.Printf("Error processing message with Claude: %v", err)
		// Send fallback message on error
		if sendErr := h.TwilioService.SendMessage(from, "Sorry, I encountered an error processing your message. Please try again."); sendErr != nil {
			log.Printf("Error sending error message: %v", sendErr)
		}
		return
	}

	// Send Claude's response back to the user
	if err := h.TwilioService.SendMessage(from, response); err != nil {
		log.Printf("ERROR sending response to user: %v", err)
		return
	}

	log.Printf("Successfully sent response to %s", userPhone)
}
