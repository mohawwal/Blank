package main

import (
	"whatsapp-bot-app/cmd/api/handlers"
)

func (app *Application) routes(handler handlers.Handler) {
	app.server.GET("/", handler.HealthCheck)
	app.server.POST("/webhook/whatsapp", handler.WhatsAppWebhook)
	app.server.POST("/webhook/paystack", handler.PaystackWebhook)
	app.server.GET("/pay", handler.PaymentRedirect)
	// app.server.GET("/purchase/confirm/:txn_id", handler.ConfirmPurchase)
}
