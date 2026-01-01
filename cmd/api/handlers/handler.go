package handlers

import (
	"whatsapp-bot-app/cmd/api/services"
	"whatsapp-bot-app/internal/processor"

	"gorm.io/gorm"
)

type Handler struct {
	DB               *gorm.DB
	ClaudeService    *services.AIService
	PaystackService  *services.PaystackService
	MessageProcessor *processor.MessageProcessor
}
