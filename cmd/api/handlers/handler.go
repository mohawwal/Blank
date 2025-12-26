package handlers

import (
	"whatsapp-bot-app/cmd/api/services"

	"gorm.io/gorm"
)

type Handler struct {
	DB              *gorm.DB
	TwilioService   *services.TwilioService
	PaystackService *services.PaystackService
	ClaudeService   *services.ClaudeService
	// VTPassService   *services.VTPassService
}
