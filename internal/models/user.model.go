package models

import "gorm.io/gorm"
type UserModel struct {
	gorm.Model
	PhoneNumber     string         `gorm:"uniqueIndex;not null" json:"phone_number"` // whatsapp:+2348012345678
	FullName        string         `gorm:"size:100" json:"full_name"`                 // User's full name
	Email           string         `gorm:"uniqueIndex;size:100" json:"email"`        // User's email
	UserName		string         `gorm:"size:50" json:"user_name"`                 // User's display name

	// Onboarding tracking
	Status          string         `gorm:"size:20;default:'new'" json:"status"`      // new, onboarding, active, suspended
	OnboardingStep  string         `gorm:"size:50" json:"onboarding_step"`           // awaiting_name, awaiting_email, awaiting_bank, etc.
	
	// Paystack Integration (Card Tokenization)
	PaystackCustomerCode  string   `gorm:"size:100" json:"paystack_customer_code"`   // Paystack customer ID
	AuthorizationCode     string   `gorm:"size:100" json:"authorization_code"`       // Authorization code for charging
	CardLast4             string   `gorm:"size:4" json:"card_last4"`                 // Last 4 digits of card
	CardType              string   `gorm:"size:20" json:"card_type"`                 // visa, mastercard, verve
	CardBank              string   `gorm:"size:100" json:"card_bank"`                // Issuing bank
	CardExpMonth          string   `gorm:"size:2" json:"card_exp_month"`             // Card expiry month
	CardExpYear           string   `gorm:"size:4" json:"card_exp_year"`              // Card expiry year
	
	// Current conversation state
	PendingTransactionID  *uint    `json:"pending_transaction_id"`                   // Transaction waiting for confirmation
	ConversationContext   string   `gorm:"type:text" json:"conversation_context"`    // Store conversation history (JSON)
}

func (receiver UserModel) TableName() string {
	return "users"
}

