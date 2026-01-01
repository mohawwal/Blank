package processor

import (
	"fmt"
	"log"
	"whatsapp-bot-app/cmd/api/services"
	"whatsapp-bot-app/internal/messaging"
	"whatsapp-bot-app/internal/models"

	"gorm.io/gorm"
)

// MessageProcessor handles all message processing logic
type MessageProcessor struct {
	DB              *gorm.DB
	AIService       *services.AIService
	PaystackService *services.PaystackService
	MessagingService messaging.MessagingService
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(db *gorm.DB, aiService *services.AIService, paystackService *services.PaystackService, messagingService messaging.MessagingService) *MessageProcessor {
	return &MessageProcessor{
		DB:              db,
		AIService:       aiService,
		PaystackService: paystackService,
		MessagingService: messagingService,
	}
}

// ProcessMessage handles incoming messages from any platform
func (p *MessageProcessor) ProcessMessage(userPhone, userName, messageBody string) error {
	var user models.UserModel
	result := p.DB.Where("phone_number = ?", userPhone).First(&user)

	// 1. NEW USER - Create user and send welcome message
	if result.Error != nil {
		return p.handleNewUser(userPhone, userName)
	}

	// 2. EXISTING USER - Check payment status
	if user.Status != "active" && user.AuthorizationCode == "" {
		return p.handleUnpaidUser(userPhone, user)
	}

	// 3. VERIFIED USER - Process their message with AI
	return p.handleVerifiedUser(userPhone, user, messageBody)
}

// handleNewUser creates a new user and sends welcome message
func (p *MessageProcessor) handleNewUser(userPhone, userName string) error {
	log.Printf("New user detected: %s", userPhone)

	if userName == "" {
		userName = "there"
	}

	newUser := models.UserModel{
		PhoneNumber:    userPhone,
		UserName:       userName,
		Status:         "new",
		OnboardingStep: "awaiting_payment",
	}

	if err := p.DB.Create(&newUser).Error; err != nil {
		log.Printf("ERROR creating user: %v", err)
		return fmt.Errorf("failed to create user: %w", err)
	}
	log.Printf("User created successfully: %s", userPhone)

	// Initialize Paystack transaction (₦100 = 10000 kobo)
	paystackResponse, err := p.PaystackService.InitializeTransaction(userPhone, 10000)
	if err != nil {
		log.Printf("ERROR initializing Paystack transaction: %v", err)
		fallbackMsg := fmt.Sprintf("Hello %s! \n\nWelcome to *Blank AI Bot*! We're experiencing technical difficulties. Please try again later.", userName)
		return p.MessagingService.SendMessage(userPhone, fallbackMsg)
	}

	log.Printf("Paystack transaction initialized: %s", paystackResponse.Data.AuthorizationURL)

	// Save transaction reference and set status to pending
	newUser.OnboardingTxnReference = paystackResponse.Data.Reference
	newUser.OnboardingTxnStatus = "pending"
	if err := p.DB.Save(&newUser).Error; err != nil {
		log.Printf("ERROR saving transaction reference: %v", err)
	}

	// Generate AI welcome message
	aiWelcomeMsg, err := p.AIService.ProcessUserMessage("Generate a warm welcome message for a new user. Keep it brief and mention we help with airtime, bills, and subscriptions.", userName, true)
	if err != nil {
		log.Printf("ERROR generating AI welcome message: %v", err)
		aiWelcomeMsg = fmt.Sprintf("Hello %s! \n\nWelcome to *Blank AI Bot*! 🎉\n\nI can help you with airtime, data, bills, and subscriptions.", userName)
	}

	aiWelcomeMsg = aiWelcomeMsg + "\n\nTo get started, please complete your wallet setup by clicking the button below."

	// Send welcome message with payment link
	templateData := messaging.TemplateData{
		TemplateName: "HXd7de8b83f27b0f8d42581a083f26bb09",
		Variables: map[string]string{
			"1": aiWelcomeMsg,
			"2": paystackResponse.Data.AccessCode,
		},
	}

	if err := p.MessagingService.SendTemplateMessage(userPhone, templateData); err != nil {
		log.Printf("ERROR sending welcome template: %v", err)
		// Fallback to regular message
		fallbackMsg := aiWelcomeMsg + "\n\n" + paystackResponse.Data.AuthorizationURL
		return p.MessagingService.SendMessage(userPhone, fallbackMsg)
	}

	log.Printf("Successfully sent AI welcome message with payment link to %s", userPhone)
	return nil
}

// handleUnpaidUser handles users who haven't completed payment
func (p *MessageProcessor) handleUnpaidUser(userPhone string, user models.UserModel) error {
	log.Printf("User %s has not completed payment. Resending payment link.", userPhone)

	// Initialize new Paystack transaction
	paystackResponse, err := p.PaystackService.InitializeTransaction(userPhone, 10000)
	if err != nil {
		log.Printf("ERROR initializing Paystack transaction: %v", err)
		reminderMsg := "Please complete your wallet setup to continue using the bot. We're having trouble generating your payment link. Please try again later."
		return p.MessagingService.SendMessage(userPhone, reminderMsg)
	}

	// Update user with new transaction reference
	user.OnboardingTxnReference = paystackResponse.Data.Reference
	user.OnboardingTxnStatus = "pending"
	if err := p.DB.Save(&user).Error; err != nil {
		log.Printf("ERROR updating transaction reference: %v", err)
	}

	// Send payment reminder
	reminderMsg := "👋 Welcome back!\n\n*Please complete your wallet setup to start using the bot:*\n\n" + paystackResponse.Data.AuthorizationURL
	return p.MessagingService.SendMessage(userPhone, reminderMsg)
}

// handleVerifiedUser processes messages from verified users
func (p *MessageProcessor) handleVerifiedUser(userPhone string, user models.UserModel, messageBody string) error {
	log.Printf("Processing message for verified user: %s", user.UserName)

	response, err := p.AIService.ProcessUserMessage(messageBody, user.UserName, false)
	if err != nil {
		log.Printf("Error processing message with AI: %v", err)
		errorMsg := "Sorry, I encountered an error processing your message. Please try again."
		return p.MessagingService.SendMessage(userPhone, errorMsg)
	}

	// Send AI response back to the user
	if err := p.MessagingService.SendMessage(userPhone, response); err != nil {
		log.Printf("ERROR sending response to user: %v", err)
		return err
	}

	log.Printf("Successfully sent response to %s", userPhone)
	return nil
}
